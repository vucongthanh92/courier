package member

import (
	"context"
	"testing"
	"time"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type memberQueryStub struct {
	members map[uint64]entities.ConversationMember
	list    []entities.ConversationMember
}

func (s memberQueryStub) ListConversationMembers(context.Context, uint64) ([]entities.ConversationMember, *errHandler.ErrorBuilder) {
	return s.list, nil
}

func (s memberQueryStub) GetConversationMember(_ context.Context, _ uint64, userID uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder) {
	if s.members == nil {
		return nil, nil
	}
	member, ok := s.members[userID]
	if !ok {
		return nil, nil
	}
	return &member, nil
}

type userGrpcStub struct {
	requestedProfiles []uint64
	profiles          []models.UserProfileSummaryResponse
}

func (s *userGrpcStub) GetPublicKey(context.Context, string) (string, string, string, *errHandler.ErrorBuilder) {
	return "", "", "", nil
}

func (s *userGrpcStub) CheckUsersStatus(context.Context, []uint64) ([]uint64, bool, *errHandler.ErrorBuilder) {
	return nil, true, nil
}

func (s *userGrpcStub) BatchGetUserProfiles(_ context.Context, userIDs []uint64) ([]models.UserProfileSummaryResponse, *errHandler.ErrorBuilder) {
	s.requestedProfiles = userIDs
	return s.profiles, nil
}

type userProfileCacheStub struct {
	cached map[uint64]models.UserProfileSummaryResponse
	set    []models.UserProfileSummaryResponse
}

func (s *userProfileCacheStub) GetMany(context.Context, []uint64) (map[uint64]models.UserProfileSummaryResponse, error) {
	if s.cached == nil {
		return map[uint64]models.UserProfileSummaryResponse{}, nil
	}
	return s.cached, nil
}

func (s *userProfileCacheStub) SetMany(_ context.Context, profiles []models.UserProfileSummaryResponse, _ time.Duration) error {
	s.set = profiles
	return nil
}

func TestListConversationMembersResolvesProfiles(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	memberQuery := memberQueryStub{
		members: map[uint64]entities.ConversationMember{
			20: {ID: 1, ConversationID: 99, UserID: 20, Role: constants.ConversationMemberRoleMember, Status: constants.ConversationMemberStatusActive, JoinedAt: now},
		},
		list: []entities.ConversationMember{
			{ID: 1, ConversationID: 99, UserID: 20, Role: constants.ConversationMemberRoleMember, Status: constants.ConversationMemberStatusActive, JoinedAt: now},
			{ID: 2, ConversationID: 99, UserID: 21, Role: constants.ConversationMemberRoleAdmin, Status: constants.ConversationMemberStatusActive, JoinedAt: now},
		},
	}
	cache := &userProfileCacheStub{}
	grpcClient := &userGrpcStub{
		profiles: []models.UserProfileSummaryResponse{
			{UserID: 20, DisplayName: "Thanh Vu", AvatarURL: "https://example.com/t.png", Status: "verified"},
			{UserID: 21, DisplayName: "Lan Tran", Status: "verified"},
		},
	}
	service := InitMemberUsecase(memberQuery, cache, grpcClient)

	response, resultErr := service.ListConversationMembers(context.Background(), &models.ListConversationMembersRequest{
		ConversationID: 99,
		RequesterID:    20,
	})

	if resultErr != nil {
		t.Fatalf("ListConversationMembers() error = %#v", resultErr)
	}
	if response.ConversationID != 99 || len(response.Members) != 2 {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response.Members[0].Profile.DisplayName != "Thanh Vu" || response.Members[1].Profile.DisplayName != "Lan Tran" {
		t.Fatalf("profiles were not mapped into members: %#v", response.Members)
	}
	if len(grpcClient.requestedProfiles) != 2 {
		t.Fatalf("requested profiles = %v, want 2 user ids", grpcClient.requestedProfiles)
	}
	if len(cache.set) != 2 {
		t.Fatalf("cached profiles = %d, want 2", len(cache.set))
	}
}
