package conversation

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	userGrpc "github.com/vucongthanh92/courier/chat-service/internal/repository/external/user_grpc"
)

type ConversationUseCaseImpl struct {
	conversationReadRepo interfaces.ConversationQueryRepoI
	conversationCmdRepo  interfaces.ConversationCommandRepoI
	memberCmdRepo        interfaces.MemberCmdRepoI
	memberQueryRepo      interfaces.MemberQueryRepoI
	userGrpcClient       userGrpc.UserGrpcClient
	txn                  *transaction.ManagerTxn
}

func InitConversationUsecase(
	conversationReadRepo interfaces.ConversationQueryRepoI,
	conversationCmdRepo interfaces.ConversationCommandRepoI,
	memberCmdRepo interfaces.MemberCmdRepoI,
	memberQueryRepo interfaces.MemberQueryRepoI,
	userGrpcClient userGrpc.UserGrpcClient,
	txn *transaction.ManagerTxn,
) interfaces.ConversationServiceI {
	return &ConversationUseCaseImpl{
		conversationReadRepo: conversationReadRepo,
		conversationCmdRepo:  conversationCmdRepo,
		memberCmdRepo:        memberCmdRepo,
		memberQueryRepo:      memberQueryRepo,
		userGrpcClient:       userGrpcClient,
		txn:                  txn,
	}
}

// func CreateConversation handles the creation of a new conversation,
// ensuring that the request is valid and that the conversation is created atomically within a transaction.
// It checks for existing direct conversations, validates member IDs, and constructs the response with the newly created conversation and its members.
func (s *ConversationUseCaseImpl) CreateConversation(ctx context.Context, req *models.CreateConversationRequest) (
	*models.CreateConversationResponse, *errHandler.ErrorBuilder) {

	// ctx, span := tracing.StartSpanFromContext(ctx, "CreateConversation")
	// defer span.End()

	// Normalize and validate member IDs
	sortedMemberIDs, err := utils.NormalizeMemberIDs(req.MemberUserIDs)
	if err != nil {
		return nil, errHandler.InitErrorBuilder(context.Background()).SetStatus(400).
			SetError(models.ErrorDTO{
				Code:    "invalid_member_ids",
				Message: err.Error(),
			})
	}

	// Ensure the creator ID is provided and valid
	if req.CreatorID == 0 {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(401).SetError(models.ErrorDTO{
			Code:    "unauthorized",
			Message: "missing authenticated user",
		})
	}
	creatorID := req.CreatorID

	// Validate the conversation type and member count based on the request
	err = req.ValidateConversationType(sortedMemberIDs)
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_conversation_type",
			Message: err.Error(),
		})
	}

	// Ensure the creator is included in the member IDs
	if isExist := utils.Contains(sortedMemberIDs, creatorID); !isExist {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_members",
			Message: "conversation creator must be included in member_user_ids",
		})
	}

	if invalidIDs, allVerified, grpcErr := s.userGrpcClient.CheckUsersStatus(ctx, sortedMemberIDs); grpcErr != nil {
		return nil, grpcErr
	} else if !allVerified {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_user_status",
			Message: fmt.Sprintf("one or more users are not verified: %v", invalidIDs),
			Field:   "member_user_ids",
		})
	}

	// Generate a unique key for direct conversations based on member IDs
	directKey := utils.GenerateConversationDirectKey(sortedMemberIDs)
	if directKey == "" {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_direct_key",
			Message: "failed to build conversation key",
		})
	}

	var (
		resp               models.CreateConversationResponse
		conversationEntity = entities.Conversation{}
	)

	// Initialize the conversation entity with the provided type, direct key, name, and creator ID.
	conversationEntity.InitConversationEntity(req.Type, &directKey, *req.Name, creatorID)

	// Check if a direct conversation with the same key already exists to prevent duplicates
	isExisted, txnErr := s.conversationReadRepo.GetDirectConversationByKey(ctx, directKey)
	if txnErr != nil {
		return nil, txnErr
	}

	if isExisted != nil {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "conversation_exists",
			Message: "conversation already exists",
		})
	}

	// Execute the conversation creation logic within a transaction to ensure atomicity
	if err := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {

		// Create the conversation and its members in the database within the transaction
		if txnErr := s.conversationCmdRepo.CreateConversation(txCtx, &conversationEntity); txnErr != nil {
			return txnErr
		}

		memberEntities := make([]entities.ConversationMember, 0, len(sortedMemberIDs))
		// Build member entities after the conversation ID is available so FK checks pass.
		for _, memberID := range sortedMemberIDs {
			var memberEntity entities.ConversationMember
			memberEntity.InitMemberEntity(conversationEntity.ID, creatorID, memberID)
			memberEntities = append(memberEntities, memberEntity)
		}

		if txnErr := s.memberCmdRepo.CreateMembers(txCtx, memberEntities); txnErr != nil {
			return txnErr
		}

		resp.FromEntity(&conversationEntity, memberEntities)

		return nil

	}); err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(500).SetLogError(err)
	}

	return &resp, nil
}

// func ListConversations retrieves a list of conversations for a user, applying pagination and returning relevant metadata about the result set.
// It validates the request, fetches conversations from the repository, and constructs a response that includes pagination information for client-side handling.
func (s *ConversationUseCaseImpl) ListConversations(ctx context.Context, req *models.ListConversationsRequest) (
	*models.ListConversationsResponse, *errHandler.ErrorBuilder) {

	if messageCode, messageErr := req.ValidateRequest(); messageCode != "" {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    messageCode,
				Message: messageErr,
			})
	}

	limit := req.Limit
	queryReq := *req
	queryReq.Limit = limit + 1
	conversations, queryErr := s.conversationReadRepo.ListConversationsByMember(ctx, &queryReq)
	if queryErr != nil {
		return nil, queryErr
	}

	hasMore := len(conversations) > limit
	if hasMore {
		conversations = conversations[:limit]
	}

	var (
		nextBeforeLastMessageAt  *time.Time
		nextBeforeConversationID *uint64
	)

	if hasMore && len(conversations) > 0 {
		last := conversations[len(conversations)-1]
		cursorTime := last.CreatedAt
		if last.LastMessageAt != nil {
			cursorTime = *last.LastMessageAt
		}
		nextBeforeLastMessageAt = &cursorTime
		nextConversationID := last.ID
		nextBeforeConversationID = &nextConversationID
	}

	return &models.ListConversationsResponse{
		Conversations: conversations,
		Pagination: models.ListConversationsPaginationResponse{
			Limit:                    limit,
			HasMore:                  hasMore,
			NextBeforeLastMessageAt:  nextBeforeLastMessageAt,
			NextBeforeConversationID: nextBeforeConversationID,
		},
	}, nil
}

// func EnsureSystemConversations ensures that system conversations exist for a given user, creating them if necessary.
func (s *ConversationUseCaseImpl) EnsureSystemConversations(ctx context.Context, req *models.EnsureSystemConversationsRequest) (
	*models.EnsureSystemConversationsResponse, *errHandler.ErrorBuilder) {

	if messageCode, messageErr := req.ValidateRequest(); messageCode != "" {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(http.StatusBadRequest).SetError(models.ErrorDTO{
			Code:    messageCode,
			Message: messageErr,
		})
	}

	resp := &models.EnsureSystemConversationsResponse{
		UserID:        req.UserID,
		Conversations: make([]models.CreateConversationResponse, 0, len(models.SystemConversationNames())),
	}

	if err := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {
		insertedEvent, txnErr := s.conversationCmdRepo.CreateProcessedEvent(txCtx, req.EventID, req.EventType)
		if txnErr != nil {
			return txnErr
		}
		if !insertedEvent {
			resp.Processed = false
			return nil
		}

		resp.Processed = true
		for _, name := range models.SystemConversationNames() {
			conversationEntity, txnErr := s.conversationReadRepo.GetSystemConversation(txCtx, req.UserID, name)
			if txnErr != nil {
				return txnErr
			}

			var memberEntities []entities.ConversationMember
			if conversationEntity == nil {
				conversationEntity = &entities.Conversation{}
				directKey := utils.GenerateSystemConversationDirectKey(req.UserID, name)
				conversationEntity.InitConversationEntity(constants.ConversationTypeSystem, &directKey, name, req.UserID)
				if txnErr := s.conversationCmdRepo.CreateConversation(txCtx, conversationEntity); txnErr != nil {
					if !txnErr.IsUniqueViolation() {
						return txnErr
					}
					conversationEntity, txnErr = s.conversationReadRepo.GetSystemConversation(txCtx, req.UserID, name)
					if txnErr != nil {
						return txnErr
					}
				}

				var memberEntity entities.ConversationMember
				memberEntity.InitMemberEntity(conversationEntity.ID, req.UserID, req.UserID)
				memberEntities = append(memberEntities, memberEntity)
				if txnErr := s.memberCmdRepo.CreateMembers(txCtx, memberEntities); txnErr != nil && !txnErr.IsUniqueViolation() {
					return txnErr
				}
			}

			if len(memberEntities) == 0 {
				member, txnErr := s.memberQueryRepo.GetConversationMember(txCtx, conversationEntity.ID, req.UserID)
				if txnErr != nil {
					return txnErr
				}
				if member != nil {
					memberEntities = append(memberEntities, *member)
				}
			}

			var conversationResp models.CreateConversationResponse
			conversationResp.FromEntity(conversationEntity, memberEntities)
			resp.Conversations = append(resp.Conversations, conversationResp)
		}
		return nil
	}); err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(http.StatusInternalServerError).SetLogError(err)
	}

	return resp, nil
}
