package main

import (
	"context"
	"encoding/json"
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

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

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

func executeTask(t Task) {
	log.Printf("[Worker] Executing Task #%d: Type=%s, Payload=%s", t.ID, t.Type, t.Payload)
	// 模拟执行耗时
	time.Sleep(1 * time.Second)
	log.Printf("[Worker] Task #%d completed", t.ID)
}
