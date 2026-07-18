package conversation

import (
	"context"
	"strings"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type ConversationUseCaseImpl struct {
	conversationReadRepo interfaces.ConversationQueryRepoI
	conversationCmdRepo  interfaces.ConversationCommandRepoI
	memberCmdRepo        interfaces.MemberCmdRepoI
	memberQueryRepo      interfaces.MemberQueryRepoI
	txn                  *transaction.ManagerTxn
}

func InitConversationUsecase(
	conversationReadRepo interfaces.ConversationQueryRepoI,
	conversationCmdRepo interfaces.ConversationCommandRepoI,
	memberCmdRepo interfaces.MemberCmdRepoI,
	memberQueryRepo interfaces.MemberQueryRepoI,
	txn *transaction.ManagerTxn,
) interfaces.ConversationServiceI {
	return &ConversationUseCaseImpl{
		conversationReadRepo: conversationReadRepo,
		conversationCmdRepo:  conversationCmdRepo,
		memberCmdRepo:        memberCmdRepo,
		memberQueryRepo:      memberQueryRepo,
		txn:                  txn,
	}
}

// func CreateConversation handles the creation of a new conversation,
// ensuring that the request is valid and that the conversation is created atomically within a transaction.
// It checks for existing direct conversations, validates member IDs, and constructs the response with the newly created conversation and its members.
func (s *ConversationUseCaseImpl) CreateConversation(ctx context.Context, req *models.CreateConversationRequest) (
	*models.CreateConversationResponse, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "CreateConversation")
	defer span.End()

	// Normalize and validate member IDs
	sortedMemberIDs, err := utils.NormalizeMemberIDs(req.MemberUserIDs)
	if err != nil {
		return nil, errHandler.InitErrorBuilder(context.Background()).SetStatus(400).
			SetError(models.ErrorDTO{
				Code:    "invalid_member_ids",
				Message: err.Error(),
			})
	}

	if req.Type == "direct" && len(sortedMemberIDs) != 2 {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_direct_members",
			Message: "direct conversation must have exactly 2 members",
		})
	}
	if req.Type == "group" && len(sortedMemberIDs) < 2 {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_group_members",
			Message: "group conversation must have at least 2 members",
		})
	}
	if req.Type == "group" && (req.Name == nil || strings.TrimSpace(*req.Name) == "") {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_group_name",
			Message: "group conversation name is required",
		})
	}

	creatorID := req.CreatorID
	if creatorID == 0 {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(401).SetError(models.ErrorDTO{
			Code:    "unauthorized",
			Message: "missing authenticated user",
		})
	}

	// Ensure the creator is included in the member IDs
	if isExist := utils.Contains(sortedMemberIDs, creatorID); !isExist {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "invalid_members",
			Message: "conversation creator must be included in member_user_ids",
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
		conversationEntity *entities.Conversation
		memberEntities     []entities.ConversationMember
	)

	// Initialize member entities for the conversation,
	// setting the appropriate role for each member based on whether they are the creator or not.
	for _, memberID := range sortedMemberIDs {
		var memberEntity entities.ConversationMember
		memberEntity.InitMemberEntity(0, creatorID, memberID)
		memberEntities = append(memberEntities, memberEntity)
	}

	// Initialize the conversation entity with the provided type, direct key, name, and creator ID.
	conversationEntity.InitConversationEntity(req.Type, &directKey, *req.Name, creatorID)

	// Execute the conversation creation logic within a transaction to ensure atomicity
	if err := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {

		// Check if a direct conversation with the same key already exists to prevent duplicates
		isExisted, txnErr := s.conversationReadRepo.GetDirectConversationByKey(txCtx, directKey)
		if txnErr != nil {
			return txnErr
		}

		if isExisted != nil {
			return errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
				Code:    "conversation_exists",
				Message: "conversation already exists",
			})
		}

		// Create the conversation and its members in the database within the transaction
		if txnErr = s.conversationCmdRepo.CreateConversation(txCtx, conversationEntity); txnErr != nil {
			return txnErr
		}

		// Update the conversation ID for each member entity to associate them with the newly created conversation
		if txnErr := s.memberCmdRepo.CreateMembers(txCtx, memberEntities); txnErr != nil {
			return txnErr
		}

		return nil

	}); err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).SetStatus(500).SetLogError(err)
	}

	// Construct and return the response payload with the newly created conversation and its members.
	resp.FromEntity(conversationEntity, memberEntities)
	return &resp, nil
}
