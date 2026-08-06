package gateway

import (
	"context"
	"fmt"

	"github.com/vucongthanh92/courier/agent-gateway/config"
	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
	"github.com/vucongthanh92/courier/agent-gateway/helper/utils"
)

type QdrantMemoryStore interface {
	Ready(ctx context.Context) error
	EnsureCollection(ctx context.Context) error
}

type Service struct {
	cfg    config.AppConfig
	memory QdrantMemoryStore
}

func NewService(cfg config.AppConfig, memory QdrantMemoryStore) *Service {
	return &Service{
		cfg:    cfg,
		memory: memory,
	}
}

func (s *Service) Bootstrap(ctx context.Context) error {
	qdrantCtx, cancel := utils.TimeoutContext(ctx, s.cfg.Qdrant.Timeout)
	defer cancel()

	if err := s.memory.Ready(qdrantCtx); err != nil {
		return fmt.Errorf("qdrant is not ready: %w", err)
	}
	if err := s.memory.EnsureCollection(qdrantCtx); err != nil {
		return err
	}
	return nil
}

func (s *Service) Health(ctx context.Context) error {
	qdrantCtx, cancel := utils.TimeoutContext(ctx, s.cfg.Qdrant.Timeout)
	defer cancel()
	return s.memory.Ready(qdrantCtx)
}

func (s *Service) SystemInstructions() string {
	return constants.AssistantSystemInstructions
}
