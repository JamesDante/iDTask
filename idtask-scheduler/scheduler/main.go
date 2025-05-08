package main

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func main() {

	log.Println("Connecting to etcd...")

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"}, // etcd address
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

	// 执行调度任务
	for {
		log.Println("[Leader] Doing scheduling work...")
		time.Sleep(5 * time.Second)
	}

	// 退出时可以 resign 释放领导权
	// _ = election.Resign(context.Background())
}
