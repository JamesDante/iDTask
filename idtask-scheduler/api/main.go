package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JamesDante/idtask-scheduler/configs"
	"github.com/JamesDante/idtask-scheduler/internal/etcdclient"
	"github.com/JamesDante/idtask-scheduler/internal/redisclient"
	"github.com/JamesDante/idtask-scheduler/models"
	"github.com/JamesDante/idtask-scheduler/storage"
	"github.com/jmoiron/sqlx"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	rdb *redis.Client
	//etcd *clientv3.Client
	ctx = context.Background()
	db  *sqlx.DB
)

func main() {
	// Connect Postgres
	storage.Init()
	db = storage.GetDB()

	// Connect Redis
	redisclient.Init()
	rdb = redisclient.GetClient()

	etcdclient.Init()
	//etcd = etcdclient.GetClient()

	// Register HTTP handler
	http.HandleFunc("/tasks", withCORS(handleTaskSubmit))
	http.HandleFunc("/delayedtasks", withCORS(handleDelayedTaskSubmit))
	http.HandleFunc("/scheduler/status", withCORS(getSchedulerStatus))

	log.Printf("Server started at %s", configs.Config.WebApiPort)
	http.ListenAndServe(configs.Config.WebApiPort, nil)
}

func handleTaskSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	payloadBytes, _ := json.Marshal(t.Payload)
	t.Payload = string(payloadBytes)

	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()
	t.ExpireAt = time.Now().AddDate(0, 0, 1)

	// Save to DB
	result := db.QueryRowx(
		"INSERT INTO tasks(id, type, payload, status, expire_at) VALUES($1, $2, $3, $4, $5) RETURNING created_at",
		t.ID, t.Type, t.Payload, t.Status, t.ExpireAt,
	)

	// log.Printf("%s", t.CreatedAt.GoString())

	err := result.Scan(&t.CreatedAt)
	if err != nil {
		log.Fatalf("Failed to insert task: %v", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// Push to Redis queue
	jobBytes, err := json.Marshal(t)
	if err != nil {
		log.Printf("Failed to marshal job: %v", err)
		return
	}
	rdb.RPush(ctx, "task-queue", jobBytes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func getSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	kvMap, err := etcdclient.Get("scheduler/status")
	if err != nil {
		http.Error(w, "Failed to get status from etcd", http.StatusInternalServerError)
		return
	}

	statuses := make([]models.SchedulerStatus, 0, len(kvMap))
	for _, val := range kvMap {
		var s models.SchedulerStatus
		if err := json.Unmarshal([]byte(val), &s); err == nil {
			statuses = append(statuses, s)
		}
	}

	json.NewEncoder(w).Encode(statuses)
}

func handleDelayedTaskSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_ = enqueueDelayedTask(t, 5*time.Second)
}

func enqueueDelayedTask(task models.Task, delay time.Duration) error {
	jobBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return rdb.ZAdd(ctx, "delayed_tasks", &redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: jobBytes,
	}).Err()
}

// middleware
func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h(w, r)
	}
}
