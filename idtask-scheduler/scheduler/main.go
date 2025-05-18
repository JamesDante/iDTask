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
	"github.com/JamesDante/idtask-scheduler/storage"
	"github.com/JamesDante/idtask-scheduler/utils"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/go-redis/redis/v8"
)

var (
	rdb     *redis.Client
	aic     *aiclient.AIClient
	pool    *WorkerPool
	etcd    *clientv3.Client
	watcher *WorkerWatcher
	ctx     = context.Background()
)

func main() {

	redisclient.Init()
	rdb = redisclient.GetClient()

	aiclient.Init()
	aic = aiclient.GetClient()

	etcdclient.Init()
	etcd = etcdclient.GetClient()

	//ctx := context.Background()

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
	leaseID, err := etcdclient.RegisterWithTTL(ctx, key, string(data), 10) // TTL 10 sec
	if err != nil {
		log.Fatal(err)
	}

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
		err = etcdclient.Update(ctx, key, string(data), leaseID) // TTL 10 sec.
		if err != nil {
			log.Fatal(err)
		}

		watcher, _ = NewWorkerWatcher(etcd, "/workers/")
		pool = NewWorkerPool()
		pool.InitFromEtcd(etcd, "/workers/")

		watcher.OnAdd = func(worker models.WorkerStatus) {
			le.mu.Lock()
			defer le.mu.Unlock()
			log.Println("add worker:", worker.ID)
			pool.Add(worker.ID)
		}

		watcher.OnDelete = func(worker models.WorkerStatus) {
			le.mu.Lock()
			defer le.mu.Unlock()
			log.Println("remove worker:", worker.ID)
			pool.Remove(worker.ID)
		}

		go pool.StartAutoRefresh(etcd, "/workers/", 10*time.Second)

		watcher.Start()
		//defer watcher.Stop()

		schedulingWork(le)
		go startProcessingQueueWatcher()
	}

	le.OnResigned = func() {
		status = models.SchedulerStatus{
			ID:       fmt.Sprintf("scheduler-%s", instanceID),
			Status:   "running",
			IsLeader: "No",
		}

		if watcher != nil {
			watcher.Stop()
			log.Println("Worker watcher stopped.")
		}

		data, _ := json.Marshal(status)
		err = etcdclient.Update(ctx, key, string(data), leaseID) // TTL 10 sec.
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

			res, err := rdb.BRPopLPush(ctx, "task-queue", "processing-queue", 0).Result()
			if err != nil {
				log.Println("Error fetching task:", err)
				continue
			}

			log.Printf("[Scheduler] Task popped: raw=%v", res)

			if len(res) < 2 {
				continue
			}

			task := parseTask(res)

			if task == nil {
				rdb.LRem(ctx, "processing-queue", 1, res)
				continue
			}

			meta := map[string]string{
				"TaskId":   task.ID,
				"TaskType": task.Type,
				"Priority": utils.FormatNullInt(task.Priority),
			}

			aiPrediction, err := aic.Predict(task.ID, meta)
			if err != nil {
				log.Println("Error AI Predict task:", err)
				continue
			}

			if task.ExpireAt != nil && time.Now().After(*task.ExpireAt) {
				log.Printf("Task %s is expired, skipping\n", task.ID)
				rdb.LRem(ctx, "processing-queue", 1, res)
				storage.UpdateTasks(task.ID, "Expired")
				continue
			}

			workerNode := chooseWorker(aiPrediction)
			// if workerNode == "" {
			// 	jobBytes, err := json.Marshal(task)
			// 	if err != nil {
			// 		log.Printf("Failed to marshal job: %v", err)
			// 		return
			// 	}
			// 	rdb.RPush(ctx, "task-queue", jobBytes)
			// }

			taskBytes, err := json.Marshal(task)
			if err != nil {
				log.Printf("Failed to marshal task %s: %v", task.ID, err)
				continue
			}

			if !pool.Exists(workerNode) {
				log.Printf("Worker %s not registered or online. Requeue task.", workerNode)
				rdb.RPush(ctx, "task-queue", taskBytes)
				continue
			}

			err = rdb.RPush(ctx, workerNode, taskBytes).Err()
			if err != nil {
				log.Println("Failed to push task to worker:", err)
			} else {
				//rdb.LRem(ctx, "processing-queue", 1, res)
				log.Printf("Task %s scheduled to worker %s\n", task.ID, workerNode)
			}
		}
	}()
}

func startProcessingQueueWatcher() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("[recovery] Checking stuck tasks in processing-queue...")

		tasks, err := rdb.LRange(context.Background(), "processing-queue", 0, -1).Result()
		if err != nil {
			log.Printf("[recovery] Failed to read processing-queue: %v", err)
			continue
		}

		for _, taskStr := range tasks {
			var task models.Task
			if err := json.Unmarshal([]byte(taskStr), &task); err != nil {
				log.Printf("[recovery] Invalid task JSON, removing: %v", err)
				rdb.LRem(context.Background(), "processing-queue", 1, taskStr)
				continue
			}

			if task.CreatedAt != nil && time.Since(*task.CreatedAt) > 30*time.Second {
				log.Printf("[recovery] Task %s expired in processing queue, requeueing", task.ID)

				rdb.LPush(context.Background(), "task-queue", taskStr)
				rdb.LRem(context.Background(), "processing-queue", 1, taskStr)
			}
		}
	}
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
