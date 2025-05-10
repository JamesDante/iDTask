package models

type AIPredictionResponse struct {
	Priority          int     `json:"priority"`
	EstimatedTime     float64 `json:"estimated_time"`
	RecommendedWorker string  `json:"recommended_worker"`
}
