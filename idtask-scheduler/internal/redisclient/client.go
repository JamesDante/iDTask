package redisclient

import (
	"context"
	"log"
	"sync"

	"github.com/JamesDante/idtask-scheduler/configs"

	"github.com/go-redis/redis/v8"
)

var (
	client *redis.Client
	once   sync.Once
)

// Init initializes the Redis client using configs.Config.
func Init() {
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr: configs.Config.RedisAddress,
		})

		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Fatalf("❌ Redis connection failed: %v", err)
		}

		log.Println("✅ Redis connected:", configs.Config.RedisAddress)
	})
}

// GetClient returns the initialized Redis client.
func GetClient() *redis.Client {
	if client == nil {
		log.Fatal("Redis client not initialized. Call redisclient.Init() first.")
	}
	return client
}
