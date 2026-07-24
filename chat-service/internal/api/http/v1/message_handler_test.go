package v1

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type messageServiceStub struct {
	request *models.SendMessageRequest
	created bool
}

func (s *messageServiceStub) CreateMessage(_ context.Context, request *models.SendMessageRequest) (*entities.Message, bool, *errHandler.ErrorBuilder) {
	s.request = request
	return &entities.Message{
		ID:             100,
		ConversationID: request.ConversationID,
		SenderID:       request.SenderID,
		Type:           request.Type,
		Body:           request.Body,
		Metadata:       map[string]any{},
	}, s.created, nil
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
	if service.request == nil || service.request.ConversationID != 10 || service.request.SenderID != 20 {
		t.Fatalf("handler did not source identities from path/JWT: %#v", service.request)
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"success":true`)) ||
		!bytes.Contains(response.Body.Bytes(), []byte(`"conversation_id":10`)) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}
