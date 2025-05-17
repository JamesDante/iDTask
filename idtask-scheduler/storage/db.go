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
		status TEXT,
		retries  INT,
		max_retry INT,
		priority INT DEFAULT 0,
		expire_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS task_logs (
		id SERIAL PRIMARY KEY,
		task_id TEXT NOT NULL,
		result TEXT,
		executed_by TEXT NOT NULL,
		executed_at TIMESTAMP DEFAULT now()
	);
	`

	db.MustExec(schema)
}

func UpdateTasks(taskID, status string) {
	_, err := db.Exec(`UPDATE tasks SET status = $1 WHERE id = $2;`, status, taskID)
	if err != nil {
		log.Printf("⚠️ Failed to update task execution: %v\n", err)
	}
}

func CreateTaskLogs(taskID, executedBy, result string) {
	_, err := db.Exec(`
        INSERT INTO task_logs (task_id, executed_by, result)
        VALUES ($1, $2, $3)
    `, taskID, executedBy, result)
	if err != nil {
		log.Printf("⚠️ Failed to log task execution: %v\n", err)
	}
}
