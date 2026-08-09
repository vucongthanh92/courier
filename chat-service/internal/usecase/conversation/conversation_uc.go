package conversation

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	userGrpc "github.com/vucongthanh92/courier/chat-service/internal/repository/external/user_grpc"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

type ConversationUseCaseImpl struct {
	conversationReadRepo interfaces.ConversationQueryRepoI
	conversationCmdRepo  interfaces.ConversationCommandRepoI
	memberCmdRepo        interfaces.MemberCmdRepoI
	memberQueryRepo      interfaces.MemberQueryRepoI
	userGrpcClient       userGrpc.UserGrpcClient
	chatEventPublisher   interfaces.ChatEventPublisherI
	txn                  *transaction.ManagerTxn
}

func InitConversationUsecase(
	conversationReadRepo interfaces.ConversationQueryRepoI,
	conversationCmdRepo interfaces.ConversationCommandRepoI,
	memberCmdRepo interfaces.MemberCmdRepoI,
	memberQueryRepo interfaces.MemberQueryRepoI,
	userGrpcClient userGrpc.UserGrpcClient,
	txn *transaction.ManagerTxn,
	chatEventPublishers ...interfaces.ChatEventPublisherI,
) interfaces.ConversationServiceI {
	var chatEventPublisher interfaces.ChatEventPublisherI
	if len(chatEventPublishers) > 0 {
		chatEventPublisher = chatEventPublishers[0]
	}
	return &ConversationUseCaseImpl{
		conversationReadRepo: conversationReadRepo,
		conversationCmdRepo:  conversationCmdRepo,
		memberCmdRepo:        memberCmdRepo,
		memberQueryRepo:      memberQueryRepo,
		userGrpcClient:       userGrpcClient,
		chatEventPublisher:   chatEventPublisher,
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

	// Ensure the creator ID is provided and valid
	if req.CreatorID == 0 {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(401).SetError(models.ErrorDTO{
			Code:    "unauthorized",
			Message: "missing authenticated user",
		})
	}
	creatorID := req.CreatorID

	memberIDs := append([]uint64{creatorID}, req.MemberIDs()...)
	sortedMemberIDs, err := utils.NormalizeMemberIDs(memberIDs)
	if err != nil {
		return nil, errHandler.InitErrorBuilder(context.Background()).SetStatus(400).
			SetError(models.ErrorDTO{
				Code:    "invalid_member_ids",
				Message: err.Error(),
			})
	}
	if len(sortedMemberIDs) == 1 {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_member_ids",
			Message: "conversation must include at least one other user",
			Field:   "member_user_ids",
		})
	}

	// Validate the conversation type and member count based on the request
	err = req.ValidateConversationType(sortedMemberIDs)
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_conversation_type",
			Message: err.Error(),
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

	directKey := utils.GenerateConversationDirectKey(sortedMemberIDs)
	if req.Type == constants.ConversationTypeDirect && directKey == "" {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_direct_key",
			Message: "failed to build conversation key",
		})
	}
	conversationName, nameErr := s.resolveConversationName(ctx, req.Name, sortedMemberIDs)
	if nameErr != nil {
		return nil, nameErr
	}

	var (
		resp               models.CreateConversationResponse
		conversationEntity = entities.Conversation{}
	)

	// Initialize the conversation entity with the provided type, direct key, name, and creator ID.
	conversationEntity.InitConversationEntity(req.Type, &directKey, conversationName, creatorID)

	if req.Type == constants.ConversationTypeDirect {
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

	// Publish a conversation created event if a chat event publisher is available, allowing other services to react to the new conversation.
	if s.chatEventPublisher != nil {
		if err := s.chatEventPublisher.PublishConversationCreated(ctx, models.ConversationCreatedPayload{
			ConversationID:   resp.ID,
			ConversationType: resp.Type,
			CreatedBy:        creatorID,
			MemberUserIDs:    sortedMemberIDs,
		}); err != nil {
			logger.Error("publish conversation created event failed", zap.Error(err), zap.Uint64("conversation_id", resp.ID))
		}
	}

	return &resp, nil
}

// resolveConversationName determines the name of the conversation based on a custom name provided by the user or by concatenating the display names of the members.
func (s *ConversationUseCaseImpl) resolveConversationName(ctx context.Context, customName *string, memberIDs []uint64) (string, *errHandler.ErrorBuilder) {
	if customName != nil {
		name := strings.TrimSpace(*customName)
		if name != "" {
			return name, nil
		}
	}

	profiles, grpcErr := s.userGrpcClient.BatchGetUserProfiles(ctx, memberIDs)
	if grpcErr != nil {
		return "", grpcErr
	}
	profileNames := make(map[uint64]string, len(profiles))
	for _, profile := range profiles {
		name := strings.TrimSpace(profile.DisplayName)
		if name != "" {
			profileNames[profile.UserID] = name
		}
	}

	names := make([]string, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		if name := profileNames[memberID]; name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "", nil
	}
	return strings.Join(names, ", "), nil
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
