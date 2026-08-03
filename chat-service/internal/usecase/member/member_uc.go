package member

import (
	"context"
	"net/http"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	userGrpc "github.com/vucongthanh92/courier/chat-service/internal/repository/external/user_grpc"
)

type MemberUseCaseImpl struct {
	memberQueryRepo  interfaces.MemberQueryRepoI
	userProfileCache interfaces.UserProfileCacheI
	userGrpcClient   userGrpc.UserGrpcClient
}

func InitMemberUsecase(
	memberQueryRepo interfaces.MemberQueryRepoI,
	userProfileCache interfaces.UserProfileCacheI,
	userGrpcClient userGrpc.UserGrpcClient,
) interfaces.MemberServiceI {
	return &MemberUseCaseImpl{
		memberQueryRepo:  memberQueryRepo,
		userProfileCache: userProfileCache,
		userGrpcClient:   userGrpcClient,
	}
}

// ListConversationMembers retrieves the list of members for a specific conversation, along with their user profiles.
func (s *MemberUseCaseImpl) ListConversationMembers(ctx context.Context, req *models.ListConversationMembersRequest) (
	*models.ListConversationMembersResponse, *errHandler.ErrorBuilder) {

	// Validate the request parameters
	if messageCode, messageErr := req.ValidateRequest(); messageCode != "" {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    messageCode,
				Message: messageErr,
			})
	}

	// Check if the requester is an active member of the conversation
	requesterMember, queryErr := s.memberQueryRepo.GetConversationMember(ctx, req.ConversationID, req.RequesterID)
	if queryErr != nil {
		return nil, queryErr
	}

	// If the requester is not an active member, return a forbidden error
	if requesterMember == nil || requesterMember.Status != constants.ConversationMemberStatusActive {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(http.StatusForbidden).SetError(models.ErrorDTO{
			Code:    "forbidden",
			Message: "requester is not an active member of this conversation",
		})
	}

	// Retrieve the list of conversation members from the repository
	memberEntities, queryErr := s.memberQueryRepo.ListConversationMembers(ctx, req.ConversationID)
	if queryErr != nil {
		return nil, queryErr
	}

	// Extract unique user IDs from the member entities to fetch their profiles
	userIDs := make([]uint64, 0, len(memberEntities))
	seenUserIDs := make(map[uint64]struct{}, len(memberEntities))
	for _, member := range memberEntities {
		if _, ok := seenUserIDs[member.UserID]; ok {
			continue
		}
		seenUserIDs[member.UserID] = struct{}{}
		userIDs = append(userIDs, member.UserID)
	}

	// Initialize a map to hold user profiles, attempting to retrieve them from the cache first
	profilesByID := make(map[uint64]models.UserProfileSummaryResponse, len(userIDs))
	if s.userProfileCache != nil {
		cachedProfiles, cacheErr := s.userProfileCache.GetMany(ctx, userIDs)
		if cacheErr == nil {
			for userID, profile := range cachedProfiles {
				profilesByID[userID] = profile
			}
		}
	}

	// Identify user IDs for which profiles are missing from the cache
	missingUserIDs := make([]uint64, 0)
	for _, userID := range userIDs {
		if _, ok := profilesByID[userID]; !ok {
			missingUserIDs = append(missingUserIDs, userID)
		}
	}

	// If there are missing profiles, fetch them from the user gRPC service and update the cache
	if len(missingUserIDs) > 0 {
		profiles, grpcErr := s.userGrpcClient.BatchGetUserProfiles(ctx, missingUserIDs)
		if grpcErr != nil {
			return nil, grpcErr
		}
		for _, profile := range profiles {
			profilesByID[profile.UserID] = profile
		}
		if s.userProfileCache != nil {
			_ = s.userProfileCache.SetMany(ctx, profiles, constants.Time_Cache_5_minutes)
		}
	}

	// Map the member entities to response objects, including their profiles
	members := make([]models.ConversationMemberResponse, 0, len(memberEntities))
	for _, memberEntity := range memberEntities {
		var member models.ConversationMemberResponse
		member.MappeDTO(memberEntity)
		if profile, ok := profilesByID[member.UserID]; ok {
			member.Profile = profile
		} else {
			member.Profile = models.UserProfileSummaryResponse{
				UserID: member.UserID,
				Status: member.Status,
			}
		}
		members = append(members, member)
	}

	// Return the final response containing the conversation ID and the list of members with their profiles
	return &models.ListConversationMembersResponse{
		ConversationID: req.ConversationID,
		Members:        members,
	}, nil
}
