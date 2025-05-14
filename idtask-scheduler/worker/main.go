package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/JamesDante/idtask-scheduler/configs"
	"github.com/JamesDante/idtask-scheduler/internal/redisclient"
	"github.com/JamesDante/idtask-scheduler/models"
	"github.com/JamesDante/idtask-scheduler/storage"
	"github.com/google/uuid"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
)

var (
	db       *sqlx.DB
	rdb      *redis.Client
	ctx      = context.Background()
	workerId string
)

func main() {

	redisclient.Init()
	rdb = redisclient.GetClient()

	storage.Init()
	db = storage.GetDB()

	initWorkerMetrics()

	workerId = generateWorkerID()
	registry, _ := NewWorkerRegistry([]string{configs.Config.EtcdAddress})
	err := registry.Register(workerId, configs.LockTTL)
	if err != nil {
		log.Fatal(err)
	}
	defer registry.Unregister()

	go consumeTasks()
	go pollDelayedTasks()

	select {}
}

func consumeTasks() {
	log.Println("Worker started. Waiting for tasks...")
	for {
		start := time.Now()
		res, err := rdb.BLPop(ctx, 0*time.Second, workerId).Result()
		if err != nil {
			log.Printf("Redis error: %v", err)
			tasksFailed.Inc()
			continue
		}

		if len(res) < 2 {
			continue
		}

		log.Printf("Raw task from Redis: %s\n", res[1])

		var t models.Task
		err = json.Unmarshal([]byte(res[1]), &t)
		if err != nil {
			log.Printf("Invalid task JSON: %v", err)
			tasksFailed.Inc()
			continue
		}

		processTask(t)
		tasksExecuted.Inc()
		taskExecDuration.Observe(time.Since(start).Seconds())
	}
}

func generateWorkerID() string {
	return fmt.Sprintf("worker-%s", uuid.New().String()[:8])
}

func pollDelayedTasks() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		now := time.Now().Unix()
		tasks, err := rdb.ZRangeByScore(ctx, "delayed_tasks", &redis.ZRangeBy{
			Min:   "-inf",
			Max:   fmt.Sprintf("%d", now),
			Count: 10,
		}).Result()
		if err != nil {
			log.Println("Polling error:", err)
			continue
		}
		for _, taskStr := range tasks {
			var task models.Task
			_ = json.Unmarshal([]byte(taskStr), &task)
			rdb.ZRem(ctx, "delayed_tasks", taskStr)
			rdb.LPush(ctx, "task_queue", taskStr)
		}
	}
}

func processTask(task models.Task) error {
	// key：task-executed:<task-id>
	key := fmt.Sprintf("task-executed:%s", task.ID)

	// SetNX: set when key not exists
	success, err := rdb.SetNX(ctx, key, 1, 24*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("Redis error: %w", err)
	}

	if !success {
		log.Printf("⚠️ Task already executed: %s, skipping\n", task.ID)
		return nil
	}

	log.Printf("✅ Executing task %s\n", task.ID)
	err = executeTask(task)

	if err != nil {
		rdb.Del(ctx, key)
		return fmt.Errorf("task failed: %w", err)
	}

	log.Printf("🎉 Task %s executed successfully\n", task.ID)
	return nil
}

func executeTask(t models.Task) error {
	log.Printf("[Worker] Executing Task #%s: Type=%s, Payload=%s", t.ID, t.Type, t.Payload)
	// TODO
	time.Sleep(1 * time.Second)
	log.Printf("[Worker] Task #%s completed", t.ID)

	logTaskExecution(db, t.ID, "success", "Task completed")

	return nil
}

func logTaskExecution(db *sqlx.DB, taskID, status, result string) {
	_, err := db.Exec(`
        INSERT INTO task_logs (task_id, status, result)
        VALUES ($1, $2, $3)
    `, taskID, status, result)
	if err != nil {
		log.Printf("⚠️ Failed to log task execution: %v\n", err)
	}
}
