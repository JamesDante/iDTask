package configs

import "time"

var LockTTL = 10 * time.Second

var Config = struct {
	AIPredictURL          string
	EtcdAddress           string
	RedisAddress          string
	PostgresConnectString string
	WebApiPort            string
}{
	AIPredictURL:          "localhost:50051",
	EtcdAddress:           "localhost:2379",
	RedisAddress:          "localhost:6379",
	PostgresConnectString: "host=localhost port=5432 user=postgres password=postgres dbname=tasks sslmode=disable",
	WebApiPort:            ":8080",
}
