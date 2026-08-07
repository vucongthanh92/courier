package gateway

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vucongthanh92/courier/agent-gateway/config"
	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
	"github.com/vucongthanh92/courier/agent-gateway/helper/utils"
	"github.com/vucongthanh92/courier/agent-gateway/internal/domain/models"
	"github.com/vucongthanh92/courier/agent-gateway/internal/safety"
)

type QdrantMemoryStore interface {
	Ready(ctx context.Context) error
	EnsureCollection(ctx context.Context) error
	UpsertMemory(ctx context.Context, point models.MemoryPoint, vector []float32) error
	SearchMemories(ctx context.Context, vector []float32, limit int, conversationID uint64) ([]models.MemoryPoint, error)
}

type AIProvider interface {
	CreateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateAnswer(ctx context.Context, req models.GenerateAnswerRequest) (*models.GenerateAnswerResponse, error)
}

type Service struct {
	cfg       config.AppConfig
	memory    QdrantMemoryStore
	provider  AIProvider
	guardrail *safety.Guardrail
}

func NewService(cfg config.AppConfig, memory QdrantMemoryStore, provider AIProvider) *Service {
	return &Service{
		cfg:       cfg,
		memory:    memory,
		provider:  provider,
		guardrail: safety.NewGuardrail(cfg.Safety),
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

func (s *Service) EvaluateSafety(text string) models.SafetyEvaluationResult {
	return s.guardrail.Evaluate(text)
}

func (s *Service) BuildBlockedAssistantResponse(req models.AssistantRequestedPayload, safetyResult models.SafetyEvaluationResult) models.AssistantRespondedPayload {
	body := safetyResult.UserMessage
	if body == "" {
		body = "Mình không thể hỗ trợ yêu cầu này vì nó thuộc nhóm nội dung nhạy cảm hoặc bị hạn chế."
	}
	return models.AssistantRespondedPayload{
		ConversationID:      req.ConversationID,
		TriggeringMessageID: req.TriggeringMessageID,
		Body:                body,
		MessageParts: BuildAssistantResponseParts(body, s.cfg.Memory.MaxMessageRunes, map[string]any{
			"status": "blocked",
		}),
		CorrelationID: req.CorrelationID,
		Metadata: map[string]any{
			"status":          "blocked",
			"safety_decision": safetyResult.Decision,
			"safety_category": safetyResult.Category,
			"safety_reasons":  safetyResult.Reasons,
		},
	}
}

func (s *Service) ProcessAssistantRequest(ctx context.Context, req models.AssistantRequestedPayload) (models.AssistantRespondedPayload, error) {
	if req.CorrelationID == "" {
		req.CorrelationID = utils.NewCorrelationID()
	}
	log.Printf("processing assistant request: conversation_id=%d triggering_message_id=%d correlation_id=%s web_search=%t", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, s.cfg.Safety.WebSearchEnabled)

	safetyResult := s.EvaluateSafety(req.Body)
	if safetyResult.Decision == constants.SafetyDecisionBlock {
		log.Printf("assistant request blocked by guardrail: conversation_id=%d triggering_message_id=%d correlation_id=%s category=%s reasons=%v", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, safetyResult.Category, safetyResult.Reasons)
		return s.BuildBlockedAssistantResponse(req, safetyResult), nil
	}
	log.Printf("assistant request allowed by guardrail: conversation_id=%d triggering_message_id=%d correlation_id=%s", req.ConversationID, req.TriggeringMessageID, req.CorrelationID)

	providerCtx, cancel := utils.TimeoutContext(ctx, s.cfg.OpenAI.RequestTimeout)
	defer cancel()

	vector, err := s.provider.CreateEmbedding(providerCtx, req.Body)
	if err != nil {
		log.Printf("create user embedding failed: conversation_id=%d triggering_message_id=%d correlation_id=%s error=%v", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, err)
		return s.BuildErrorAssistantResponse(req, err), nil
	}
	log.Printf("created user embedding: conversation_id=%d triggering_message_id=%d correlation_id=%s vector_size=%d", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, len(vector))

	userMemory := models.MemoryPoint{
		ID:             utils.StableUUID(req.CorrelationID + ":" + constants.MemoryRoleUser),
		ConversationID: req.ConversationID,
		ChatMessageID:  req.TriggeringMessageID,
		Role:           constants.MemoryRoleUser,
		Body:           req.Body,
		Source:         "chat-service",
		CorrelationID:  req.CorrelationID,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.memory.UpsertMemory(ctx, userMemory, vector); err != nil {
		log.Printf("upsert user memory failed: conversation_id=%d triggering_message_id=%d correlation_id=%s error=%v", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, err)
	} else {
		log.Printf("upserted user memory: conversation_id=%d triggering_message_id=%d correlation_id=%s memory_id=%s", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, userMemory.ID)
	}

	memories, searchErr := s.memory.SearchMemories(ctx, vector, s.cfg.OpenAI.MaxMemoryChunks, req.ConversationID)
	if searchErr != nil {
		log.Printf("search memories failed: conversation_id=%d triggering_message_id=%d correlation_id=%s error=%v", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, searchErr)
	} else {
		log.Printf("searched memories: conversation_id=%d triggering_message_id=%d correlation_id=%s count=%d", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, len(memories))
	}
	answer, err := s.provider.GenerateAnswer(providerCtx, models.GenerateAnswerRequest{
		WebSearch: s.cfg.Safety.WebSearchEnabled,
		ContextPackage: models.ContextPackage{
			SystemInstructions: constants.AssistantSystemInstructions,
			RelevantMemories:   memories,
			CurrentMessage:     req.Body,
		},
	})
	if err != nil {
		log.Printf("generate assistant answer failed: conversation_id=%d triggering_message_id=%d correlation_id=%s error=%v", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, err)
		return s.BuildErrorAssistantResponse(req, err), nil
	}
	log.Printf("generated assistant answer: conversation_id=%d triggering_message_id=%d correlation_id=%s model=%s response_id=%s runes=%d", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, answer.Model, answer.ResponseID, len([]rune(answer.Text)))

	assistantMemory := models.MemoryPoint{
		ID:             utils.StableUUID(req.CorrelationID + ":" + constants.MemoryRoleAssistant),
		ConversationID: req.ConversationID,
		Role:           constants.MemoryRoleAssistant,
		Body:           answer.Text,
		Source:         constants.ServiceNameAgentGateway,
		CorrelationID:  req.CorrelationID,
		Metadata: map[string]any{
			"openai_response_id": answer.ResponseID,
			"model":              answer.Model,
		},
		CreatedAt: time.Now().UTC(),
	}
	if answerVector, embeddingErr := s.provider.CreateEmbedding(providerCtx, answer.Text); embeddingErr == nil {
		if err := s.memory.UpsertMemory(ctx, assistantMemory, answerVector); err != nil {
			log.Printf("upsert assistant memory failed: conversation_id=%d triggering_message_id=%d correlation_id=%s error=%v", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, err)
		} else {
			log.Printf("upserted assistant memory: conversation_id=%d triggering_message_id=%d correlation_id=%s memory_id=%s vector_size=%d", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, assistantMemory.ID, len(answerVector))
		}
	} else {
		log.Printf("create assistant embedding failed: conversation_id=%d triggering_message_id=%d correlation_id=%s error=%v", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, embeddingErr)
	}

	parts := BuildAssistantResponseParts(answer.Text, s.cfg.Memory.MaxMessageRunes, map[string]any{
		"status":             "completed",
		"model":              answer.Model,
		"openai_response_id": answer.ResponseID,
		"web_search_enabled": s.cfg.Safety.WebSearchEnabled,
	})
	return models.AssistantRespondedPayload{
		ConversationID:      req.ConversationID,
		TriggeringMessageID: req.TriggeringMessageID,
		Body:                parts[0].Body,
		MessageParts:        parts,
		CorrelationID:       req.CorrelationID,
		Metadata: map[string]any{
			"status":             "completed",
			"model":              answer.Model,
			"openai_response_id": answer.ResponseID,
			"usage":              answer.Usage,
		},
	}, nil
}

func (s *Service) BuildErrorAssistantResponse(req models.AssistantRequestedPayload, cause error) models.AssistantRespondedPayload {
	log.Printf("building failed assistant response: conversation_id=%d triggering_message_id=%d correlation_id=%s error=%v", req.ConversationID, req.TriggeringMessageID, req.CorrelationID, cause)
	body := "Mình đang gặp lỗi khi xử lý yêu cầu này. Bạn vui lòng thử lại sau."
	parts := BuildAssistantResponseParts(body, s.cfg.Memory.MaxMessageRunes, map[string]any{
		"status": "failed",
	})
	return models.AssistantRespondedPayload{
		ConversationID:      req.ConversationID,
		TriggeringMessageID: req.TriggeringMessageID,
		Body:                parts[0].Body,
		MessageParts:        parts,
		CorrelationID:       req.CorrelationID,
		Metadata: map[string]any{
			"status": "failed",
			"error":  cause.Error(),
		},
	}
}
