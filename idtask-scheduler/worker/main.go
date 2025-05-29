package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/JamesDante/idtask-scheduler/configs"
	"github.com/JamesDante/idtask-scheduler/internal/redisclient"
	"github.com/JamesDante/idtask-scheduler/models"
	"github.com/JamesDante/idtask-scheduler/monitor"
	"github.com/JamesDante/idtask-scheduler/storage"
	"github.com/google/uuid"

	"github.com/go-redis/redis/v8"

	jsoniter "github.com/json-iterator/go"
)

var (
	//db       *sqlx.DB
	rdb          *redis.Client
	ctx          = context.Background()
	workerId     string
	failureCount int
	unHealth     bool
	json         = jsoniter.ConfigFastest
	taskPool     = sync.Pool{
		New: func() any {
			return new(models.Task)
		},
	}
)

const maxFailures = 3

func main() {

	configs.InitConfig()

	redisclient.Init()
	rdb = redisclient.GetClient()

	storage.Init()

	//configs.InitConfig()

	//TODO: initialize monitoring
	//monitor.InitWorkerMetrics()

	unHealth = false

	workerId = generateWorkerID()
	registry, _ := NewWorkerRegistry([]string{configs.Config.EtcdAddress})
	err := registry.Register(workerId, configs.LockTTL)
	if err != nil {
		log.Fatal(err)
	}
	defer registry.Unregister()

	go consumeTasks(registry)

	go startWorkerHeartbeat(registry, workerId)

	select {}
}

func consumeTasks(registry *WorkerRegistry) {
	log.Println("Worker started. Waiting for tasks...")
	for {
		//start := time.Now()
		res, err := rdb.BLPop(ctx, 0*time.Second, workerId).Result()
		if err != nil {
			log.Printf("Redis error: %v", err)
			continue
		}

		if len(res) < 2 {
			continue
		}

		log.Printf("Raw task from Redis: %s\n", res[1])

		t := taskPool.Get().(*models.Task)
		*t = models.Task{}

		defer taskPool.Put(t)

		//var t models.Task
		rawTask := res[1]
		err = json.Unmarshal([]byte(res[1]), t)
		if err != nil {
			log.Printf("Invalid task JSON: %v", err)
			//monitor.WorkerTasksFailed().Inc()
			continue
		}

		processTask(registry, *t, rawTask)
		//monitor.WorkerTasksExecuted().Inc()
		//monitor.WorkerTaskExecDuration().Observe(time.Since(start).Seconds())
	}
}

func generateWorkerID() string {
	host, _ := os.Hostname()
	return fmt.Sprintf("worker-%s-%s", host, uuid.New().String()[:6])
}

func processTask(registry *WorkerRegistry, task models.Task, rawTask string) error {
	// key：task-executed:<task-id>
	key := fmt.Sprintf("task-executed:%s", task.ID)

	// SetNX: set when key not exists
	success, err := rdb.SetNX(ctx, key, 1, 24*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("Redis error: %w", err)
	}

	if !success {
		log.Printf("⚠️ Task already executed: %s, skipping\n", task.ID)
		rdb.LRem(ctx, "processing-queue", 1, rawTask)
		return nil
	}

	log.Printf("✅ Executing task %s\n", task.ID)
	err = executeTask(task, rawTask)

	if err != nil {
		rdb.Del(ctx, key)
		rdb.LRem(ctx, "processing-queue", 1, rawTask)
		storage.UpdateTasks(task.ID, "Failed")
		storage.CreateTaskLogs(task.ID, workerId, "Task Failed")

		failureCount++
		if failureCount >= maxFailures {
			log.Printf("❌ Worker %s marked as failed after %d consecutive failures", workerId, failureCount)

			unHealth = true
		}

		monitor.WorkerTasksFailed().Inc()
		return fmt.Errorf("task failed: %w", err)
	}

	unHealth = false
	failureCount = 0

	log.Printf("🎉 Task %s executed successfully\n", task.ID)
	return nil
}

func executeTask(t models.Task, rawTask string) error {
	log.Printf("[Worker] Executing Task #%s: Type=%s, Payload=%s", t.ID, t.Type, t.Payload)
	// TODO
	time.Sleep(1 * time.Second)
	log.Printf("[Worker] Task #%s completed", t.ID)

	rdb.LRem(ctx, "processing-queue", 1, rawTask)

	storage.UpdateTasks(t.ID, "Completed")
	storage.CreateTaskLogs(t.ID, workerId, "Task completed")
	//updateTaskExecution(db, t.ID, "Completed")
	//logTaskExecution(db, t.ID, workerId, "Task completed")

	return nil
}

func startWorkerHeartbeat(registry *WorkerRegistry, workerId string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		statusStr := "ok"
		if unHealth {
			statusStr = "failed"
		}

		status := models.WorkerStatus{
			ID:        workerId,
			Status:    statusStr,
			HeartBeat: time.Now(),
		}
		data, _ := json.Marshal(status)
		if err := registry.Update(workerId, string(data)); err != nil {
			log.Printf("Failed to refresh heartbeat for worker %s: %v", workerId, err)
		}
	}
}
