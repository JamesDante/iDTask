package configs

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var LockTTL = 10 * time.Second

type ConfigStruct struct {
	AIPredictURL          string
	EtcdAddress           string
	RedisAddress          string
	PostgresConnectString string
	WebApiPort            string
}

var Config ConfigStruct

func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, using fallback values")
	}

	Config = ConfigStruct{
		AIPredictURL:          getEnv("AI_PREDICT_URL", "localhost:50051"),
		EtcdAddress:           getEnv("ETCD_ADDRESS", "localhost:2379"),
		RedisAddress:          getEnv("REDIS_ADDRESS", "localhost:6379"),
		PostgresConnectString: getEnv("PG_CONN_STRING", "host=localhost port=5432 user=postgres password=postgres dbname=tasks sslmode=disable"),
		WebApiPort:            getEnv("WEB_API_PORT", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	// path, _ := os.Getwd()
	// log.Println("📁 Working directory:", path)

	fmt.Println("fallback for", key, "is", fallback)
	return fallback
}
