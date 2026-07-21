package conversation

import (
	"context"
	"fmt"

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
