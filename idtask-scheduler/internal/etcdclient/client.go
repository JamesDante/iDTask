package etcdclient

import (
	"context"
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

func Set(key, val string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := cli.Put(ctx, key, val)
	return err
}
