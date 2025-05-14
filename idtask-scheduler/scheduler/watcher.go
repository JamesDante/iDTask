package main

import (
	"context"
	"log"
	"strings"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type WorkerWatcher struct {
	Client    *clientv3.Client
	Prefix    string
	OnAdd     func(addr string)
	OnDelete  func(addr string)
	CancelCtx context.CancelFunc
}

func NewWorkerWatcher(cli *clientv3.Client, prefix string) (*WorkerWatcher, error) {

	// cli, err := clientv3.New(clientv3.Config{
	// 	Endpoints:   []string{configs.Config.EtcdAddress},
	// 	DialTimeout: 5 * time.Second,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	return &WorkerWatcher{
		Client: cli,
		Prefix: prefix,
	}, nil
}

func (w *WorkerWatcher) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.CancelCtx = cancel

	go func() {
		log.Println("[watcher] start watching workers...")
		rch := w.Client.Watch(ctx, w.Prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
		for resp := range rch {
			for _, ev := range resp.Events {
				addr := strings.TrimPrefix(string(ev.Kv.Key), w.Prefix)
				switch ev.Type {
				case mvccpb.PUT:
					log.Printf("[watcher] Worker added: %s", addr)
					if w.OnAdd != nil {
						w.OnAdd(string(ev.Kv.Value))
					}
				case mvccpb.DELETE:
					log.Printf("[watcher] Worker removed: %s", addr)
					if w.OnDelete != nil {
						w.OnDelete(string(ev.PrevKv.Value))
					}
				}
			}
		}
	}()
}

func (w *WorkerWatcher) Stop() {
	if w.CancelCtx != nil {
		w.CancelCtx()
	}
	w.Client.Close()
}
