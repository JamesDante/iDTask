package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Task struct {
	ID        int       `db:"id" json:"id"`
	Type      string    `db:"type" json:"type"`
	Payload   string    `db:"payload" json:"payload"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Retries   int       `db:"retries" json:"retries"`
	MaxRetry  int       `db:"maxRetry" json:"max_retry"`
}

var (
	db  *sqlx.DB
	rdb *redis.Client
	ctx = context.Background()
)

func main() {
	// Connect Postgres
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=tasks sslmode=disable"
	var err error
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Postgres error: %v", err)
	}

	// Create table if not exists
	db.MustExec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			type TEXT,
			payload TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)

	// Connect Redis
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Register HTTP handler
	http.HandleFunc("/tasks", handleTaskSubmit)
	http.HandleFunc("/delayedtasks", handleDelayedTaskSubmit)
	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

func handleTaskSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	payloadBytes, _ := json.Marshal(t.Payload)
	t.Payload = string(payloadBytes)

	// Save to DB
	result := db.QueryRowx(
		"INSERT INTO tasks(type, payload) VALUES($1, $2) RETURNING id, created_at",
		t.Type, t.Payload,
	)
	if err := result.Scan(&t.ID, &t.CreatedAt); err != nil {
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

	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_ = enqueueDelayedTask(t, 5*time.Second)
}

func enqueueDelayedTask(task Task, delay time.Duration) error {

	jobBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return rdb.ZAdd(ctx, "delayed_tasks", &redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: jobBytes,
	}).Err()
}
