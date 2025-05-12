package main

import (
	"context"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type LeaderElector struct {
	Client     *clientv3.Client
	Session    *concurrency.Session
	Election   *concurrency.Election
	Key        string
	ID         string
	isLeader   bool
	mu         sync.RWMutex
	OnElected  func()
	OnResigned func()
}

func NewLeaderElector(cli *clientv3.Client, electionKey string, id string, ttl time.Duration) (*LeaderElector, error) {

	// cli, err := clientv3.New(clientv3.Config{
	// 	Endpoints:   []string{configs.Config.EtcdAddress},
	// 	DialTimeout: 5 * time.Second,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	session, err := concurrency.NewSession(cli, concurrency.WithTTL(int(ttl.Seconds())))
	if err != nil {
		return nil, err
	}

	e := concurrency.NewElection(session, electionKey)

	return &LeaderElector{
		Client:   cli,
		Session:  session,
		Election: e,
		Key:      electionKey,
		ID:       id,
	}, nil
}

func (le *LeaderElector) CampaignLoop(ctx context.Context) {
	for {
		log.Println("[election] Starting leader campaign...")
		err := le.Election.Campaign(ctx, le.ID)
		if err != nil {
			log.Println("[election] Campaign error:", err)
			time.Sleep(3 * time.Second)
			continue
		}

		le.setLeader(true)
		log.Println("[election] I am the leader")

		if le.OnElected != nil {
			le.OnElected()
		}

		// wait session expire
		select {
		case <-le.Session.Done():
			log.Println("[election] Leadership lost due to session end")
		case <-ctx.Done():
			log.Println("[election] Leadership loop context cancelled")
		}

		le.setLeader(false)
		if le.OnResigned != nil {
			le.OnResigned()
		}
	}
}

func (le *LeaderElector) setLeader(val bool) {
	le.mu.Lock()
	defer le.mu.Unlock()
	le.isLeader = val
}

func (le *LeaderElector) IsLeader() bool {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return le.isLeader
}
