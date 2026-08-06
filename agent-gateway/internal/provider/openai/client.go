package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vucongthanh92/courier/agent-gateway/config"
	"github.com/vucongthanh92/courier/agent-gateway/internal/domain/models"
)

const defaultBaseURL = "https://api.openai.com/v1"

type Client struct {
	cfg        config.OpenAIConfig
	httpClient *http.Client
	baseURL    string
}

func NewClient(cfg config.OpenAIConfig) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		baseURL: defaultBaseURL,
	}
}

func (c *Client) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if strings.TrimSpace(c.cfg.APIKey) == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}
	reqBody := map[string]any{
		"model": c.cfg.EmbeddingModel,
		"input": text,
	}
	var respBody embeddingResponse
	if err := c.post(ctx, "/embeddings", reqBody, &respBody); err != nil {
		return nil, err
	}
	if len(respBody.Data) == 0 {
		return nil, fmt.Errorf("OpenAI embeddings response is empty")
	}
	return respBody.Data[0].Embedding, nil
}

func (c *Client) GenerateAnswer(ctx context.Context, req models.GenerateAnswerRequest) (*models.GenerateAnswerResponse, error) {
	if strings.TrimSpace(c.cfg.APIKey) == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	body := map[string]any{
		"model":        c.cfg.Model,
		"instructions": req.ContextPackage.SystemInstructions,
		"input":        buildInput(req.ContextPackage),
	}
	if req.WebSearch {
		body["tools"] = []map[string]any{
			{
				"type":                "web_search",
				"search_context_size": "medium",
			},
		}
	}

	var resp responseCreateResponse
	if err := c.post(ctx, "/responses", body, &resp); err != nil {
		return nil, err
	}
	text := strings.TrimSpace(resp.OutputText)
	if text == "" {
		text = strings.TrimSpace(resp.extractOutputText())
	}
	return &models.GenerateAnswerResponse{
		Text:       text,
		Model:      resp.Model,
		ResponseID: resp.ID,
		Usage:      resp.Usage,
	}, nil
}

func buildInput(ctxPackage models.ContextPackage) string {
	var builder strings.Builder
	if ctxPackage.ConversationSummary != "" {
		builder.WriteString("Conversation summary:\n")
		builder.WriteString(ctxPackage.ConversationSummary)
		builder.WriteString("\n\n")
	}
	if len(ctxPackage.RelevantMemories) > 0 {
		builder.WriteString("Relevant memories:\n")
		for _, memory := range ctxPackage.RelevantMemories {
			builder.WriteString("- ")
			builder.WriteString(memory.Role)
			builder.WriteString(": ")
			builder.WriteString(memory.Body)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}
	if len(ctxPackage.RecentMessages) > 0 {
		builder.WriteString("Recent messages:\n")
		for _, message := range ctxPackage.RecentMessages {
			builder.WriteString("- ")
			builder.WriteString(message.Role)
			builder.WriteString(": ")
			builder.WriteString(message.Body)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("Current user message:\n")
	builder.WriteString(ctxPackage.CurrentMessage)
	return builder.String()
}

func (c *Client) post(ctx context.Context, path string, reqBody any, respBody any) error {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("OpenAI API request failed: %s: %s", resp.Status, string(data))
	}
	if err := json.Unmarshal(data, respBody); err != nil {
		return err
	}
	return nil
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

type responseCreateResponse struct {
	ID         string         `json:"id"`
	Model      string         `json:"model"`
	OutputText string         `json:"output_text"`
	Output     []responseItem `json:"output"`
	Usage      map[string]any `json:"usage"`
}

type responseItem struct {
	Type    string            `json:"type"`
	Content []responseContent `json:"content"`
}

type responseContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (r responseCreateResponse) extractOutputText() string {
	var builder strings.Builder
	for _, item := range r.Output {
		for _, content := range item.Content {
			if content.Text == "" {
				continue
			}
			if builder.Len() > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(content.Text)
		}
	}
	return builder.String()
}

func TimeoutContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, timeout)
}
