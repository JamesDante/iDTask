package models

import "time"

type TaskLogs struct {
	ID        int       `db:"id" json:"id"`
	TaskID    string    `db:"task_id" json:"task_id"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
