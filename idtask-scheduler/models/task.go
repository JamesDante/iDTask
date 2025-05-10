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
