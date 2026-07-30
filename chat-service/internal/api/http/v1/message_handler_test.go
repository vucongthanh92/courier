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

type messageServiceStub struct {
	createRequest *models.SendMessageRequest
	listRequest   *models.ListMessagesRequest
	created       bool
}

func (s *messageServiceStub) CreateMessage(_ context.Context, request *models.SendMessageRequest) (*models.MessageResponse, bool, *errHandler.ErrorBuilder) {
	s.createRequest = request
	return &models.MessageResponse{
		ID:             100,
		ConversationID: request.ConversationID,
		SenderID:       request.SenderID,
		Type:           request.Type,
		Body:           request.Body,
		Metadata:       map[string]any{},
	}, s.created, nil
}

func (s *messageServiceStub) ListMessages(_ context.Context, request *models.ListMessagesRequest) (*models.ListMessagesResponse, *errHandler.ErrorBuilder) {
	s.listRequest = request
	now := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	nextID := uint64(99)
	return &models.ListMessagesResponse{
		ConversationID: request.ConversationID,
		Messages: []models.MessageResponse{
			{
				ID:             100,
				ConversationID: request.ConversationID,
				SenderID:       request.RequesterID,
				Type:           "text",
				Body:           "hello",
				Metadata:       map[string]any{},
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
		Pagination: models.MessagePaginationResponse{
			Limit:               request.Limit,
			NextBeforeMessageID: &nextID,
			HasMore:             true,
		},
	}, nil
}

func TestMessageHandlerCreateMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &messageServiceStub{created: true}
	handler := InitMessageHandler(service)
	router := gin.New()
	router.POST(
		"/api/v1/conversations/:id/messages/create",
		func(c *gin.Context) {
			c.Set("authClaims", jwt.MapClaims{"sub": "20"})
			c.Next()
		},
		handler.CreateMessage,
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/conversations/10/messages/create",
		bytes.NewBufferString(`{"type":"text","body":"hello","metadata":{}}`),
	)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}
	if service.createRequest == nil || service.createRequest.ConversationID != 10 || service.createRequest.SenderID != 20 {
		t.Fatalf("handler did not source identities from path/JWT: %#v", service.createRequest)
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"success":true`)) ||
		!bytes.Contains(response.Body.Bytes(), []byte(`"conversation_id":"10"`)) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestMessageHandlerListMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &messageServiceStub{}
	handler := InitMessageHandler(service)
	router := gin.New()
	router.GET(
		"/api/v1/conversations/:id/messages",
		func(c *gin.Context) {
			c.Set("authClaims", jwt.MapClaims{"sub": "20"})
			c.Next()
		},
		handler.ListMessages,
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/conversations/10/messages?limit=25&before_message_id=100",
		nil,
	)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if service.listRequest == nil ||
		service.listRequest.ConversationID != 10 ||
		service.listRequest.RequesterID != 20 ||
		service.listRequest.Limit != 25 ||
		service.listRequest.BeforeMessageID == nil ||
		*service.listRequest.BeforeMessageID != 100 {
		t.Fatalf("handler did not bind list request correctly: %#v", service.listRequest)
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"success":true`)) ||
		!bytes.Contains(response.Body.Bytes(), []byte(`"has_more":true`)) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}
