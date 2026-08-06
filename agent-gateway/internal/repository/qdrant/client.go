package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vucongthanh92/courier/agent-gateway/config"
	"github.com/vucongthanh92/courier/agent-gateway/internal/domain/models"
)

type Client struct {
	cfg        config.QdrantConfig
	httpClient *http.Client
}

func NewClient(cfg config.QdrantConfig) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *Client) Ready(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/readyz"), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("qdrant ready check failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) EnsureCollection(ctx context.Context) error {
	body := map[string]any{
		"vectors": map[string]any{
			"size":     c.cfg.VectorSize,
			"distance": c.cfg.Distance,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.url("/collections/"+c.cfg.CollectionName),
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("ensure qdrant collection failed: %s: %s", resp.Status, string(data))
	}
	return nil
}

func (c *Client) UpsertMemory(ctx context.Context, point models.MemoryPoint, vector []float32) error {
	if len(vector) == 0 {
		return fmt.Errorf("memory vector is required")
	}

	body := map[string]any{
		"points": []map[string]any{
			{
				"id":      point.ID,
				"vector":  vector,
				"payload": point,
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.url("/collections/"+c.cfg.CollectionName+"/points?wait=true"),
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("upsert qdrant memory failed: %s: %s", resp.Status, string(data))
	}
	return nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.cfg.APIKey != "" {
		req.Header.Set("api-key", c.cfg.APIKey)
	}
	return c.httpClient.Do(req)
}

func (c *Client) url(path string) string {
	base := strings.TrimRight(c.cfg.URL, "/")
	if strings.HasPrefix(path, "/") {
		return base + path
	}
	return base + "/" + path
}
