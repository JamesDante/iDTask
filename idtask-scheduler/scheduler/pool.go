package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/JamesDante/idtask-scheduler/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type WorkerPool struct {
	mu      sync.RWMutex
	workers []string
	index   int
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		workers: []string{},
		index:   0,
	}
}

func (wp *WorkerPool) InitFromEtcd(cli *clientv3.Client, prefix string) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	wp.workers = make([]string, 0, len(resp.Kvs))
	seen := make(map[string]bool)
	for _, kv := range resp.Kvs {
		var status models.WorkerStatus
		if err := json.Unmarshal(kv.Value, &status); err != nil {
			continue
		}

		if !seen[status.ID] {
			wp.workers = append(wp.workers, status.ID)
			seen[status.ID] = true
		}
	}
	wp.index = 0
	return nil
}

func (wp *WorkerPool) StartAutoRefresh(etcd *clientv3.Client, prefix string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("[pool] Auto-refreshing worker pool from etcd...")
		err := wp.InitFromEtcd(etcd, prefix)
		if err != nil {
			log.Printf("[pool] Failed to refresh from etcd: %v", err)
		}
	}
}

func (wp *WorkerPool) Exists(id string) bool {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	for _, w := range wp.workers {
		if w == id {
			return true
		}
	}
	return false
}

func (wp *WorkerPool) Add(worker string) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	for _, w := range wp.workers {
		if w == worker {
			return
		}
	}
	wp.workers = append(wp.workers, worker)
}

func (wp *WorkerPool) Remove(worker string) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	newWorkers := []string{}
	for _, w := range wp.workers {
		if w != worker {
			newWorkers = append(newWorkers, w)
		}
	}
	wp.workers = newWorkers
	if wp.index >= len(wp.workers) {
		wp.index = 0
	}
}

func (wp *WorkerPool) Next() (string, error) {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if len(wp.workers) == 0 {
		return "", errors.New("no available workers")
	}

	worker := wp.workers[wp.index]
	wp.index = (wp.index + 1) % len(wp.workers)
	return worker, nil
}
