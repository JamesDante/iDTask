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
)

func main() {
	redisclient.Init()
	rdb = redisclient.GetClient()

	aiclient.Init()
	aic = aiclient.GetClient()

	aiclient.TestConnection()

	ctx := context.Background()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{configs.Config.EtcdAddress},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	instanceID := generateInstanceID()
	le, err := NewLeaderElector(cli, "/scheduler/leader", instanceID, configs.LockTTL)
	if err != nil {
		log.Fatal(err)
	}
	//release resources
	defer le.Client.Close()
	defer le.Session.Close()

	le.OnElected = func() {
		log.Println("Elected leader, starting scheduler")

		watcher, _ := NewWorkerWatcher(cli, "/workers/")
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
