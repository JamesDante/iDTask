package models

import (
	"database/sql"
	"time"
)

type Task struct {
	ID          string         `db:"id" json:"id"`
	Type        string         `db:"type" json:"type"`
	Payload     string         `db:"payload" json:"payload"`
	Retries     sql.NullInt64  `db:"retries" json:"retries"`
	Status      string         `db:"status" json:"status"`
	MaxRetry    sql.NullInt64  `db:"max_retry" json:"max_retry"`
	CreatedAt   *time.Time     `db:"created_at" json:"created_at"`
	ExpireAt    *time.Time     `db:"expire_at" json:"expire_at"`
	Priority    sql.NullInt64  `db:"priority" json:"priority"`
	ExecutedBy  sql.NullString `db:"executed_by" json:"executed_by"`
	ExecutedAt  *time.Time     `db:"executed_at" json:"executed_at"`
	ScheduledAt *time.Time     `db:"scheduled_at" json:"scheduled_at"`
}

type TaskLogs struct {
	ID         sql.NullInt64 `db:"id" json:"id"`
	TaskID     string        `db:"task_id" json:"task_id"`
	ExecutedBy string        `db:"executed_by" json:"executed_by"`
	Result     string        `db:"result" json:"result"`
	ExecutedAt *time.Time    `db:"executed_at" json:"executed_at"`
}

type AIPredictionResponse struct {
	Priority          sql.NullInt64 `json:"priority"`
	EstimatedTime     float64       `json:"estimated_time"`
	RecommendedWorker string        `json:"recommended_worker"`
}

type AIPredictionRequest struct {
	TaskID   string            `json:"task_id"`
	Metadata map[string]string `json:"metadata"`
}

type SchedulerStatus struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	IsLeader  string    `json:"isLeader"`
	HeartBeat time.Time `json:"heart_beat"`
}

type WorkerStatus struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	HeartBeat time.Time `json:"heart_beat"`
}

type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Total  int         `json:"total"`
	Error  string      `json:"error,omitempty"`
}

type APIListResponse struct {
	Status   string      `json:"status"`
	ListData interface{} `json:"list_data,omitempty"`
	Total    int         `json:"total"`
}

type APIListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
