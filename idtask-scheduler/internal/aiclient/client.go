package aiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/JamesDante/idtask-scheduler/configs"
	"github.com/JamesDante/idtask-scheduler/models"
)

var (
	once   sync.Once
	client *AIClient
)

type AIClient struct {
	baseURL string
	client  *http.Client
}

func NewAIClient(baseUrl string) *AIClient {
	return &AIClient{
		baseURL: baseUrl,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func Init() {
	once.Do(func() {
		client = NewAIClient(configs.Config.RedisAddress)
	})
}

func GetClient() *AIClient {
	if client == nil {
		log.Fatal("ai client not initialized. Call aiclient.Init() first.")
	}
	return client
}

func (c *AIClient) Predict(request models.AIPredictionRequest) (*models.AIPredictionResponse, error) {

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI request: %w", err)
	}

	resp, err := http.Post(configs.Config.AIPredictURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service error: %s", string(body))
	}

	var aiResp models.AIPredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	return &aiResp, nil
}
