package main

import (
	"errors"
	"sync"
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

func (wp *WorkerPool) Add(worker string) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	// 避免重复添加
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
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if len(wp.workers) == 0 {
		return "", errors.New("no available workers")
	}

	worker := wp.workers[wp.index]
	wp.index = (wp.index + 1) % len(wp.workers)
	return worker, nil
}
