package aiclient

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/JamesDante/idtask-scheduler/internal/aiclient/predict"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/JamesDante/idtask-scheduler/configs"
)

var (
	once   sync.Once
	client *AIClient
)

type AIClient struct {
	conn   *grpc.ClientConn
	client pb.AIPredictorClient
}

func NewAIClient(addr string) (*AIClient, error) {

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AI gRPC service: %w", err)
	}
	client := pb.NewAIPredictorClient(conn)
	return &AIClient{conn: conn, client: client}, nil
}

func Init() {
	once.Do(func() {
		client, _ = NewAIClient(configs.Config.AIPredictURL)
	})
}

func GetClient() *AIClient {
	if client == nil {
		log.Fatal("ai client not initialized. Call aiclient.Init() first.")
	}
	return client
}

func (c *AIClient) Predict(taskID string, metadata map[string]string) (*pb.PredictResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := &pb.PredictRequest{
		TaskId:   taskID,
		Metadata: metadata,
	}

	resp, err := c.client.Predict(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC Predict call failed: %w", err)
	}
	return resp, nil
}

func (c *AIClient) Close() {
	c.conn.Close()
}

func TestConnection() {
	Init()
	client := GetClient()
	defer client.Close()

	// try Predict
	resp, err := client.Predict("test-task", map[string]string{"key": "value"})
	if err != nil {
		log.Fatalf("❌ Predict failed: %v", err)
	}
	log.Printf("✅ Predict succeeded: Priority=%d, Worker=%s", resp.Priority, resp.RecommendedWorker)
}

// func NewAIClient(baseUrl string) *AIClient {
// 	return &AIClient{
// 		baseURL: baseUrl,
// 		client: &http.Client{
// 			Timeout: 5 * time.Second,
// 		},
// 	}
// }

// Predict is safe for concurrent use.
// func (c *AIClient) Predict(request models.AIPredictionRequest) (*models.AIPredictionResponse, error) {

// 	jsonData, err := json.Marshal(request)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal AI request: %w", err)
// 	}

// 	resp, err := http.Post(configs.Config.AIPredictURL, "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to call AI service: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return nil, fmt.Errorf("AI service error: %s", string(body))
// 	}

// 	var aiResp models.AIPredictionResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
// 		return nil, fmt.Errorf("failed to decode AI response: %w", err)
// 	}

// 	return &aiResp, nil
// }
