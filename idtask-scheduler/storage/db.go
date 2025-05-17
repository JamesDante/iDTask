package storage

import (
	"log"

	"github.com/JamesDante/idtask-scheduler/configs"
	"github.com/JamesDante/idtask-scheduler/models"
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

func CreateTask(t *models.Task) *sqlx.Row {
	result := db.QueryRowx(
		"INSERT INTO tasks(id, type, payload, status, expire_at) VALUES($1, $2, $3, $4, $5) RETURNING created_at",
		t.ID, t.Type, t.Payload, t.Status, t.ExpireAt,
	)

	return result
}

func GetTasksCount() int {
	var total int
	_ = db.Get(&total, "SELECT COUNT(*) FROM tasks")

	return total
}

func GetTasks(req *models.APIListRequest) ([]models.Task, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	offset := (req.Page - 1) * req.PageSize

	tasks := []models.Task{}
	err := db.Select(&tasks, `
		SELECT 
		  t.id,
		  t.type,
		  t.payload,
		  t.status,
		  t.retries,
		  t.max_retry,
		  t.priority,
		  t.expire_at,
		  t.created_at,
		  l.executed_by,
		  l.executed_at
		FROM tasks t
		LEFT JOIN LATERAL (
		  SELECT * FROM task_logs l
		  WHERE l.task_id = t.id
		  ORDER BY l.executed_at DESC
		  LIMIT 1
		) l ON true
		ORDER BY t.created_at DESC LIMIT $1 OFFSET $2;`, req.PageSize, offset)

	if err != nil {
		log.Printf("Failed to query tasks: %v", err)
		return tasks, err
	}

	return tasks, nil
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
