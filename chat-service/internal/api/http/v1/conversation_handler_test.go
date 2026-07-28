package v1

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type conversationServiceStub struct {
	listRequest *models.ListConversationsRequest
}

func (s *conversationServiceStub) CreateConversation(context.Context, *models.CreateConversationRequest) (*models.CreateConversationResponse, *errHandler.ErrorBuilder) {
	return nil, nil
}

func (s *conversationServiceStub) ListConversations(_ context.Context, request *models.ListConversationsRequest) (*models.ListConversationsResponse, *errHandler.ErrorBuilder) {
	s.listRequest = request
	now := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	return &models.ListConversationsResponse{
		Conversations: []models.ConversationListResponse{
			{ID: 10, Type: "direct", CreatedAt: now, UpdatedAt: now},
		},
		Pagination: models.ListConversationsPaginationResponse{Limit: request.Limit},
	}, nil
}

func TestConversationHandlerListConversations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &conversationServiceStub{}
	handler := InitConversationHandler(service)
	router := gin.New()
	router.GET(
		"/api/v1/conversations",
		func(c *gin.Context) {
			c.Set("authClaims", jwt.MapClaims{"sub": "20"})
			c.Next()
		},
		handler.ListConversations,
	)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/conversations?limit=25&before_last_message_at=2026-07-27T10:00:00Z&before_conversation_id=10",
		nil,
	)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if service.listRequest == nil ||
		service.listRequest.RequesterID != 20 ||
		service.listRequest.Limit != 25 ||
		service.listRequest.BeforeConversationID == nil ||
		*service.listRequest.BeforeConversationID != 10 ||
		service.listRequest.BeforeLastMessageAt == nil {
		t.Fatalf("handler did not bind list request correctly: %#v", service.listRequest)
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"success":true`)) ||
		!bytes.Contains(response.Body.Bytes(), []byte(`"conversations"`)) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}
