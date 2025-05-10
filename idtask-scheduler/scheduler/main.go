package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/JamesDante/idtask-scheduler/configs"
	"github.com/JamesDante/idtask-scheduler/internal/aiclient"
	"github.com/JamesDante/idtask-scheduler/internal/redisclient"
	"github.com/JamesDante/idtask-scheduler/models"

	"github.com/go-redis/redis/v8"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	rdb *redis.Client
	aic *aiclient.AIClient
)

func main() {

	redisclient.Init()
	rdb = redisclient.GetClient()

	aiclient.Init()
	aic = aiclient.GetClient()

	log.Println("Connecting to etcd...")

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{configs.Config.EtcdAddress}, // etcd address
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
	defer cli.Close()
	log.Println("Connected to etcd.")

	session, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	election := concurrency.NewElection(session, "/scheduler-leader")

	ctx := context.Background()
	instanceID := fmt.Sprintf("instance-%d", time.Now().UnixNano())

	// 发起竞选，阻塞直到成为 Leader
	if err := election.Campaign(ctx, instanceID); err != nil {
		log.Fatalf("Campaign error: %v", err)
	}

	log.Printf("[Leader] I am the leader now: %s", instanceID)

	for {
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

		aiPrediction, err := aic.Predict(models.AIPredictionRequest{
			TaskID:   task.ID,
			Metadata: meta,
		})
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

	// 退出时可以 resign 释放领导权
	// _ = election.Resign(context.Background())
}

func chooseWorker(prediction *models.AIPredictionResponse) string {
	if prediction.EstimatedTime > 2 {
		return "worker-2"
	}
	return "worker-1"
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
