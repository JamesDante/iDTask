package models

type AIPredictionRequest struct {
	TaskID   string            `json:"task_id"`
	Metadata map[string]string `json:"metadata"`
}
