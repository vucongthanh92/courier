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
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type memberServiceStub struct {
	listMembersRequest *models.ListConversationMembersRequest
}

func (s *memberServiceStub) ListConversationMembers(_ context.Context, request *models.ListConversationMembersRequest) (*models.ListConversationMembersResponse, *errHandler.ErrorBuilder) {
	s.listMembersRequest = request
	return &models.ListConversationMembersResponse{ConversationID: request.ConversationID}, nil
}

func TestMemberHandlerListConversationMembers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &memberServiceStub{}
	handler := InitMemberHandler(service)
	router := gin.New()
	router.GET(
		"/api/v1/conversation/:id/members",
		func(c *gin.Context) {
			c.Set("authClaims", jwt.MapClaims{"sub": "20"})
			c.Next()
		},
		handler.ListConversationMembers,
	)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/conversation/99/members", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if service.listMembersRequest == nil ||
		service.listMembersRequest.RequesterID != 20 ||
		service.listMembersRequest.ConversationID != 99 {
		t.Fatalf("handler did not bind list members request correctly: %#v", service.listMembersRequest)
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"conversation_id":"99"`)) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}
