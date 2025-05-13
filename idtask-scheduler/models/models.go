package models

import "time"

type Task struct {
	ID        string    `db:"id" json:"id"`
	Type      string    `db:"type" json:"type"`
	Payload   string    `db:"payload" json:"payload"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Retries   int       `db:"retries" json:"retries"`
	MaxRetry  int       `db:"maxRetry" json:"max_retry"`
	ExpireAt  time.Time `db:"expire_at" json:"expire_at"`
	Priority  int       `db:"priority" json:"priority"`
}

type TaskLogs struct {
	ID         int       `db:"id" json:"id"`
	TaskID     string    `db:"task_id" json:"task_id"`
	Status     string    `db:"status" json:"status"`
	Result     string    `db:"result" json:"result"`
	ExecutedAt time.Time `db:"created_at" json:"created_at"`
}

type AIPredictionResponse struct {
	Priority          int     `json:"priority"`
	EstimatedTime     float64 `json:"estimated_time"`
	RecommendedWorker string  `json:"recommended_worker"`
}

type AIPredictionRequest struct {
	TaskID   string            `json:"task_id"`
	Metadata map[string]string `json:"metadata"`
}

type SchedulerStatus struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	IsLeader string `json:"isLeader"`
}
