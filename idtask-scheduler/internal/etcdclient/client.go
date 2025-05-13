package etcdclient

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JamesDante/idtask-scheduler/configs"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var cli *clientv3.Client

func Init() {
	var err error
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{configs.Config.EtcdAddress},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to etcd: %v", err)
	}

	log.Println("✅ etcd connected:", configs.Config.EtcdAddress)
}

func GetClient() *clientv3.Client {
	if cli == nil {
		log.Fatal("etcd client not initialized. Call etcdclient.Init() first.")
	}
	return cli
}

func RegisterWithTTL(ctx context.Context, key, value string, ttl int64) (clientv3.LeaseID, error) {
	// create lease
	leaseResp, err := cli.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}

	_, err = cli.Put(ctx, key, value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return 0, err
	}

	ch, err := cli.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return 0, err
	}

	go func() {
		for ka := range ch {
			if ka == nil {
				log.Printf("Lease keepalive channel closed for key: %s", key)
				return
			}
		}
	}()

	log.Printf("Key %s registered with TTL %d seconds", key, ttl)
	return leaseResp.ID, nil
}

func Get(prefix string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}

	return result, nil
}

func Set(key string, val string, leaseRespID clientv3.LeaseID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := cli.Put(ctx, key, val, clientv3.WithLease(leaseRespID))
	return err
}

func Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := cli.Delete(ctx, key)
	return err
}

func DeleteWithPrefix(prefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := cli.Delete(ctx, prefix, clientv3.WithPrefix())
	return err
}

func Update(ctx context.Context, key string, val string, leaseRespID clientv3.LeaseID) error {

	resp, err := cli.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return fmt.Errorf("key %s not found", key)
	}

	_, err = cli.Put(ctx, key, val, clientv3.WithLease(leaseRespID))
	return err
}
