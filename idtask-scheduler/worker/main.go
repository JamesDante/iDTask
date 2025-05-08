package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Task struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx = context.Background()
)

func main() {
	go worker()
	go pollDelayedTasks()
}

func worker() {
	log.Println("Worker started. Waiting for tasks...")
	for {
		res, err := rdb.BLPop(ctx, 0*time.Second, "task-queue").Result()
		if err != nil {
			log.Printf("Redis error: %v", err)
			continue
		}

		if len(res) < 2 {
			continue
		}

		var t Task
		err = json.Unmarshal([]byte(res[1]), &t)
		if err != nil {
			log.Printf("Invalid task JSON: %v", err)
			continue
		}

		executeTask(t)
	}
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
			var task Task
			_ = json.Unmarshal([]byte(taskStr), &task)
			rdb.ZRem(ctx, "delayed_tasks", taskStr)
			rdb.LPush(ctx, "task_queue", taskStr)
		}
	}
}

func executeTask(t Task) {
	log.Printf("[Worker] Executing Task #%d: Type=%s, Payload=%s", t.ID, t.Type, t.Payload)
	// 模拟执行耗时
	time.Sleep(1 * time.Second)
	log.Printf("[Worker] Task #%d completed", t.ID)
}
