package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JamesDante/idtask-scheduler/configs"
	"github.com/JamesDante/idtask-scheduler/internal/aiclient"
	pb "github.com/JamesDante/idtask-scheduler/internal/aiclient/predict"
	"github.com/JamesDante/idtask-scheduler/internal/etcdclient"
	"github.com/JamesDante/idtask-scheduler/internal/redisclient"
	"github.com/JamesDante/idtask-scheduler/models"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/go-redis/redis/v8"
)

var (
	rdb  *redis.Client
	aic  *aiclient.AIClient
	pool *WorkerPool
	etcd *clientv3.Client
)

func main() {

	redisclient.Init()
	rdb = redisclient.GetClient()

	aiclient.Init()
	aic = aiclient.GetClient()

	etcdclient.Init()
	etcd = etcdclient.GetClient()

	ctx := context.Background()

	instanceID := generateInstanceID()
	le, err := NewLeaderElector(etcd, "/scheduler/leader", instanceID, configs.LockTTL)
	if err != nil {
		log.Fatal(err)
	}

	status := models.SchedulerStatus{
		ID:       fmt.Sprintf("scheduler-%s", instanceID),
		Status:   "running",
		IsLeader: "No",
	}

	key := fmt.Sprintf("scheduler/status/%s", status.ID)
	data, _ := json.Marshal(status)
	leaseID, err := etcdclient.RegisterWithTTL(ctx, key, string(data), 10) // TTL 10 秒，可调
	if err != nil {
		log.Fatal(err)
	}

	// err = etcdclient.Set(key, string(data))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// //Remove key when application exit
	// defer func() {
	// 	err := etcdclient.Delete(key)
	// 	if err != nil {
	// 		log.Printf("Failed to delete scheduler status from etcd: %v", err)
	// 	} else {
	// 		log.Printf("Deleted scheduler status from etcd: %s", key)
	// 	}
	// }()

	// stopChan := make(chan os.Signal, 1)
	// signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// go func() {
	// 	<-stopChan
	// 	log.Println("Received termination signal, cleaning up...")

	// 	err := etcdclient.Delete(key)
	// 	if err != nil {
	// 		log.Printf("Failed to delete scheduler status from etcd: %v", err)
	// 	} else {
	// 		log.Printf("Deleted scheduler status from etcd: %s", key)
	// 	}

	// 	os.Exit(0)
	// }()

	//release resources
	defer le.Client.Close()
	defer le.Session.Close()

	le.OnElected = func() {
		log.Println("Elected leader, starting scheduler")

		status = models.SchedulerStatus{
			ID:       fmt.Sprintf("scheduler-%s", instanceID),
			Status:   "running",
			IsLeader: "Yes",
		}

		data, _ := json.Marshal(status)
		err = etcdclient.Update(ctx, key, string(data), leaseID) // TTL 10 秒，可调
		if err != nil {
			log.Fatal(err)
		}

		watcher, _ := NewWorkerWatcher("/workers/")
		pool := NewWorkerPool()

		watcher.OnAdd = func(addr string) {
			log.Println("add worker:", addr)
			pool.Add(addr)
		}

		watcher.OnDelete = func(addr string) {
			log.Println("remove worker:", addr)
			pool.Remove(addr)
		}

		watcher.Start()
		defer watcher.Stop()

		schedulingWork(le)
	}

	le.OnResigned = func() {
		status = models.SchedulerStatus{
			ID:       fmt.Sprintf("scheduler-%s", instanceID),
			Status:   "running",
			IsLeader: "No",
		}

		data, _ := json.Marshal(status)
		err = etcdclient.Update(ctx, key, string(data), leaseID) // TTL 10 秒，可调
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Resigned from leadership, stopping scheduler")
	}

	le.CampaignLoop(ctx)
}

func schedulingWork(le *LeaderElector) {
	go func() {
		for le.IsLeader() {
			log.Println("[Leader] Doing scheduling work...")

			res, err := rdb.BLPop(context.Background(), 0*time.Second, "task-queue").Result()
			if err != nil {
				log.Println("Error fetching task:", err)
				continue
			}

			if len(res) < 2 {
				continue
			}

			task := parseTask(res[1])

			if task == nil {
				continue
			}

			meta := map[string]string{
				"TaskId":   task.ID,
				"TaskType": task.Type,
				"Priority": fmt.Sprintf("%d", task.Priority),
			}

			aiPrediction, err := aic.Predict(task.ID, meta)
			if err != nil {
				log.Println("Error AI Predict task:", err)
				continue
			}

			if time.Now().After(task.ExpireAt) {
				log.Printf("Task %s is expired, skipping\n", task.ID)
				continue
			}

			workerNode := chooseWorker(aiPrediction)

			err = rdb.RPush(context.Background(), workerNode, task.ID).Err()
			if err != nil {
				log.Println("Failed to push task to worker:", err)
			} else {
				log.Printf("Task %s scheduled to worker %s\n", task.ID, workerNode)
			}
		}
	}()
}

func generateInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%s", hostname, uuid.New().String()[:8])
}

func chooseWorker(prediction *pb.PredictResponse) string {
	// find worker recommended by AI
	for _, w := range pool.workers {
		if w == prediction.RecommendedWorker {
			log.Printf("AI recommended worker selected: %s", w)
			return w
		}
	}

	worker, err := pool.Next()
	if err != nil {
		log.Println("No available worker, fallback failed")
		return ""
	}

	log.Printf("Fallback to round-robin worker: %s", worker)
	return worker
}

func parseTask(taskstr string) *models.Task {
	var t models.Task
	err := json.Unmarshal([]byte(taskstr), &t)
	if err != nil {
		log.Printf("Invalid task JSON: %v", err)
		return nil
	}
	return &t
}
