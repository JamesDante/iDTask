package storage

import (
	"log"

	"github.com/JamesDante/idtask-scheduler/configs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

func Init() {
	var err error
	db, err = sqlx.Connect("postgres", configs.Config.PostgresConnectString)
	if err != nil {
		log.Fatalf("Postgres error: %v", err)
	}

	createTables()
}

func GetDB() *sqlx.DB {
	return db
}

func createTables() {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		type TEXT,
		payload TEXT,
		retries  INT,
		max_retry INT,
		priority INT DEFAULT 0,
		expire_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS task_logs (
		id SERIAL PRIMARY KEY,
		task_id TEXT NOT NULL,
		status TEXT NOT NULL,
		result TEXT,
		executed_at TIMESTAMP DEFAULT now()
	);
	`

	db.MustExec(schema)
}
