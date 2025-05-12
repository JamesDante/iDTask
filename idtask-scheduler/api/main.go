package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JamesDante/idtask-scheduler/configs"
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

	// Register HTTP handler
	http.HandleFunc("/tasks", handleTaskSubmit)
	http.HandleFunc("/delayedtasks", handleDelayedTaskSubmit)
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

	// Save to DB
	result := db.QueryRowx(
		"INSERT INTO tasks(id, type, payload) VALUES($1, $2, $3) RETURNING created_at",
		t.ID, t.Type, t.Payload,
	)

	if err := result.Scan(&t.CreatedAt); err != nil {
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
