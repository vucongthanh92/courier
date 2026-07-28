package conversation

import (
	"context"
	"testing"
	"time"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type conversationQueryStub struct {
	listRequest       models.ListConversationsRequest
	listConversations []models.ConversationListResponse
}

func (s *conversationQueryStub) GetDirectConversationByKey(context.Context, string) (*entities.Conversation, *errHandler.ErrorBuilder) {
	return nil, nil
}

func (s *conversationQueryStub) GetConversationByID(context.Context, uint64) (*entities.Conversation, *errHandler.ErrorBuilder) {
	return nil, nil
}

func (s *conversationQueryStub) ListConversationsByMember(_ context.Context, req *models.ListConversationsRequest) ([]models.ConversationListResponse, *errHandler.ErrorBuilder) {
	s.listRequest = *req
	return s.listConversations, nil
}

type conversationCommandStub struct{}

func (s conversationCommandStub) CreateConversation(context.Context, *entities.Conversation) *errHandler.ErrorBuilder {
	return nil
}

type memberCommandStub struct{}

func (s memberCommandStub) CreateMembers(context.Context, []entities.ConversationMember) *errHandler.ErrorBuilder {
	return nil
}

func (s memberCommandStub) UpdateReadState(context.Context, uint64, uint64, uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder) {
	return nil, nil
}

type memberQueryStub struct{}

func (s memberQueryStub) ListConversationMembers(context.Context, uint64) ([]entities.ConversationMember, *errHandler.ErrorBuilder) {
	return nil, nil
}

func (s memberQueryStub) GetConversationMember(context.Context, uint64, uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder) {
	return nil, nil
}

type userGrpcStub struct{}

func (s userGrpcStub) GetPublicKey(context.Context, string) (string, string, string, *errHandler.ErrorBuilder) {
	return "", "", "", nil
}

func (s userGrpcStub) CheckUsersStatus(context.Context, []uint64) ([]uint64, bool, *errHandler.ErrorBuilder) {
	return nil, true, nil
}

func TestListConversationsSuccess(t *testing.T) {
	now := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	query := &conversationQueryStub{
		listConversations: []models.ConversationListResponse{
			{ID: 12, Type: "group", LastMessageAt: &now, CreatedAt: now},
			{ID: 11, Type: "direct", CreatedAt: now.Add(-time.Hour)},
			{ID: 10, Type: "group", CreatedAt: now.Add(-2 * time.Hour)},
		},
	}
	service := InitConversationUsecase(query, conversationCommandStub{}, memberCommandStub{}, memberQueryStub{}, userGrpcStub{}, nil)

	response, resultErr := service.ListConversations(context.Background(), &models.ListConversationsRequest{
		RequesterID: 20,
		Limit:       2,
	})

	if resultErr != nil {
		t.Fatalf("ListConversations() error = %#v", resultErr)
	}
	if len(response.Conversations) != 2 || !response.Pagination.HasMore {
		t.Fatalf("unexpected response: %#v", response)
	}
	if query.listRequest.Limit != 3 {
		t.Fatalf("query limit = %d, want limit+1", query.listRequest.Limit)
	}
	if response.Pagination.NextBeforeConversationID == nil || *response.Pagination.NextBeforeConversationID != 11 {
		t.Fatalf("next conversation cursor = %#v, want 11", response.Pagination.NextBeforeConversationID)
	}
	if response.Pagination.NextBeforeLastMessageAt == nil || !response.Pagination.NextBeforeLastMessageAt.Equal(now.Add(-time.Hour)) {
		t.Fatalf("next time cursor = %#v", response.Pagination.NextBeforeLastMessageAt)
	}
}

func TestListConversationsValidation(t *testing.T) {
	service := InitConversationUsecase(&conversationQueryStub{}, conversationCommandStub{}, memberCommandStub{}, memberQueryStub{}, userGrpcStub{}, nil)

	_, resultErr := service.ListConversations(context.Background(), &models.ListConversationsRequest{})

	if resultErr == nil || resultErr.Status != 400 || len(resultErr.Errors) == 0 || resultErr.Errors[0].Code != "unauthorized" {
		t.Fatalf("error = %#v, want unauthorized", resultErr)
	}
}
