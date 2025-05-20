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
	"github.com/JamesDante/idtask-scheduler/monitor"
	"github.com/JamesDante/idtask-scheduler/storage"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

func main() {
	// Connect Postgres
	storage.Init()

	// Connect Redis
	redisclient.Init()
	rdb = redisclient.GetClient()

	etcdclient.Init()

	monitor.InitApiMetrics()

	// Register HTTP handler
	http.HandleFunc("/tasks", withCORS(handleTaskSubmit))
	http.HandleFunc("/tasks/list", withCORS(handleTaskList))
	http.HandleFunc("/delayedtasks", withCORS(handleDelayedTaskSubmit))
	http.HandleFunc("/scheduler/status", withCORS(getSchedulerStatus))
	http.HandleFunc("/worker/status", withCORS(getWorkerStatus))

	log.Printf("Server started at %s", configs.Config.WebApiPort)
	http.ListenAndServe(configs.Config.WebApiPort, nil)
}

func handleTaskList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		//http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		writeJSON(w, http.StatusMethodNotAllowed, nil, "Only POST allowed")
		return
	}

	var req models.APIListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, nil, "Invalid JSON")
		return
	}

	tasks, err := storage.GetTasks(&req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, nil, "Failed to fetch tasks")
		return
	}

	tasksCount := storage.GetTasksCount()

	resp := models.APIListResponse{
		Status:   "OK",
		ListData: tasks,
		Total:    tasksCount,
	}

	//resp.ListData = tasks

	writeJSON(w, http.StatusOK, resp, "")
}

func handleTaskSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		//http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		writeJSON(w, http.StatusMethodNotAllowed, nil, "Only POST allowed")
		return
	}

	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		//http.Error(w, "Invalid JSON", http.StatusBadRequest)
		writeJSON(w, http.StatusBadRequest, nil, "Invalid JSON")
		return
	}

	payloadBytes, _ := json.Marshal(t.Payload)
	t.Payload = string(payloadBytes)

	t.ID = uuid.New().String()
	createdAt := time.Now()
	expireAt := time.Now().AddDate(0, 0, 1)

	t.CreatedAt = &createdAt
	t.ExpireAt = &expireAt

	// Save to DB
	result := storage.CreateTask(&t)

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

	monitor.ApiRequestsTotal().Inc()
	//w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(t)
	writeJSON(w, http.StatusOK, t, "")
}

func getWorkerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		//http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		writeJSON(w, http.StatusMethodNotAllowed, nil, "Only POST allowed")
		return
	}

	kvMap, err := etcdclient.Get("/workers")
	if err != nil {
		//http.Error(w, "Failed to get status from etcd", http.StatusInternalServerError)
		writeJSON(w, http.StatusInternalServerError, nil, "Failed to get status from etcd")
		return
	}

	statuses := make([]models.WorkerStatus, 0, len(kvMap))
	for _, val := range kvMap {
		var s models.WorkerStatus
		if err := json.Unmarshal([]byte(val), &s); err == nil {
			statuses = append(statuses, s)
		} else {
			log.Printf("❌ Failed to parse worker status: %v\nRaw: %s", err, val)
		}
	}

	//json.NewEncoder(w).Encode(statuses)
	writeJSON(w, http.StatusOK, statuses, "")
}

func getSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		//http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		writeJSON(w, http.StatusMethodNotAllowed, nil, "Only POST allowed")
		return
	}

	kvMap, err := etcdclient.Get("scheduler/status")
	if err != nil {
		//http.Error(w, "Failed to get status from etcd", http.StatusInternalServerError)
		writeJSON(w, http.StatusInternalServerError, nil, "Failed to get status from etcd")
		return
	}

	statuses := make([]models.SchedulerStatus, 0, len(kvMap))
	for _, val := range kvMap {
		var s models.SchedulerStatus
		if err := json.Unmarshal([]byte(val), &s); err == nil {
			statuses = append(statuses, s)
		}
	}

	//json.NewEncoder(w).Encode(statuses)
	writeJSON(w, http.StatusOK, statuses, "")
}

func handleDelayedTaskSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		//http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		writeJSON(w, http.StatusMethodNotAllowed, nil, "Only POST allowed")
		return
	}

	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		//http.Error(w, "Invalid JSON", http.StatusBadRequest)
		writeJSON(w, http.StatusBadRequest, nil, "Invalid JSON")
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

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := models.APIResponse{
		Status: http.StatusText(statusCode),
	}

	if errMsg != "" {
		resp.Error = errMsg
	} else {
		resp.Data = data
	}

	json.NewEncoder(w).Encode(resp)
}
