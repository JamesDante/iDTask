package main

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type WorkerRegistry struct {
	Client   *clientv3.Client
	LeaseID  clientv3.LeaseID
	Key      string
	StopChan chan struct{}
}

func NewWorkerRegistry(endpoints []string) (*WorkerRegistry, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &WorkerRegistry{Client: cli, StopChan: make(chan struct{})}, nil
}

func (r *WorkerRegistry) Register(address string, ttl time.Duration) error {
	leaseResp, err := r.Client.Grant(context.Background(), int64(ttl.Seconds()))
	if err != nil {
		return fmt.Errorf("grant lease failed: %w", err)
	}
	r.LeaseID = leaseResp.ID

	key := fmt.Sprintf("/workers/%s", address)
	_, err = r.Client.Put(context.Background(), key, address, clientv3.WithLease(r.LeaseID))
	if err != nil {
		return fmt.Errorf("put with lease failed: %w", err)
	}
	r.Key = key

	// KeepAlive goroutine
	go func() {
		ch, err := r.Client.KeepAlive(context.Background(), r.LeaseID)
		if err != nil {
			log.Printf("keepalive failed: %v", err)
			return
		}
		for {
			select {
			case <-r.StopChan:
				log.Println("stop keepalive")
				return
			case _, ok := <-ch:
				if !ok {
					log.Println("keepalive channel closed")
					return
				}
			}
		}
	}()
	log.Printf("Worker registered with key: %s", key)
	return nil
}

func (r *WorkerRegistry) Unregister() {
	close(r.StopChan)
	r.Client.Delete(context.Background(), r.Key)
	r.Client.Close()
	log.Printf("Worker unregistered: %s", r.Key)
}
