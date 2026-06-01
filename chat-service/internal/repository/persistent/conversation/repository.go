package conversation

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/chat-service/database"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	"gorm.io/gorm"
)

type conversationQueryRepository struct {
	readDB  *gorm.DB
	writeDB *gorm.DB
}

func InitConversationRepository(readDb *database.GormReadDb, writeDb *database.GormWriteDb) interfaces.ConversationRepositoryI {
	return &conversationQueryRepository{
		readDB:  *readDb,
		writeDB: *writeDb,
	}
}

func (r *conversationQueryRepository) GetDirectConversationByKey(ctx context.Context, directKey string) (*entities.Conversation, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.Conversation
	err := run.Model(&entities.Conversation{}).
		Where("direct_key = ? AND deleted_at IS NULL", directKey).
		Take(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}

func (r *conversationQueryRepository) GetConversationByID(ctx context.Context, id uint64) (*entities.Conversation, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.Conversation
	err := run.Model(&entities.Conversation{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&res).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}

func (r *conversationQueryRepository) CreateConversation(ctx context.Context, entity *entities.Conversation) *errHandler.ErrorBuilder {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	if err := run.Model(&entities.Conversation{}).Create(entity).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

func (r *conversationQueryRepository) ListInbox(ctx context.Context, userID uint64) ([]models.InboxConversationResponse, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)

	type row struct {
		ConversationID      uint64
		ConversationType    string
		DirectKey           *string
		Name                *string
		CreatedBy           uint64
		LastMessageID       *uint64
		LastMessageAt       *time.Time
		ConversationMeta    map[string]any
		ConversationCreated time.Time
		ConversationUpdated time.Time
		MemberID            uint64
		MemberRole          string
		MemberStatus        string
		MemberJoined        time.Time
		MemberLeftAt        *time.Time
		LastReadMessageID   *uint64
		LastReadAt          *time.Time
		MutedUntil          *time.Time
		MessageID           *uint64
		MessageSenderID     *uint64
		MessageType         *string
		MessageBody         *string
		MessageReplyTo      *uint64
		MessageClientID     *string
		MessageMeta         map[string]any
		MessageCreatedAt    *time.Time
		MessageUpdatedAt    *time.Time
		MessageEditedAt     *time.Time
		MessageDeletedAt    *time.Time
		UnreadCount         int64
	}

	rows := []row{}
	err := run.Raw(`
		SELECT
			c.id AS conversation_id,
			c.type AS conversation_type,
			c.direct_key,
			c.name,
			c.created_by,
			c.last_message_id,
			c.last_message_at,
			c.metadata AS conversation_meta,
			c.created_at AS conversation_created,
			c.updated_at AS conversation_updated,
			cm.id AS member_id,
			cm.role AS member_role,
			cm.status AS member_status,
			cm.joined_at AS member_joined,
			cm.left_at,
			cm.last_read_message_id,
			cm.last_read_at,
			cm.muted_until,
			m.id AS message_id,
			m.sender_id AS message_sender_id,
			m.type AS message_type,
			m.body AS message_body,
			m.reply_to_message_id AS message_reply_to,
			m.client_message_id AS message_client_id,
			m.metadata AS message_meta,
			m.created_at AS message_created_at,
			m.updated_at AS message_updated_at,
			m.edited_at AS message_edited_at,
			m.deleted_at AS message_deleted_at,
			(
				SELECT COUNT(*)
				FROM "chat-service".messages um
				WHERE um.conversation_id = c.id
				  AND um.deleted_at IS NULL
				  AND um.sender_id <> cm.user_id
				  AND um.created_at > COALESCE(cm.last_read_at, TO_TIMESTAMP(0))
			) AS unread_count
		FROM "chat-service".conversation_members cm
		JOIN "chat-service".conversations c
			ON c.id = cm.conversation_id
		LEFT JOIN "chat-service".messages m
			ON m.id = c.last_message_id
		WHERE cm.user_id = ?
		  AND cm.status = 'active'
		  AND c.deleted_at IS NULL
		ORDER BY c.last_message_at DESC NULLS LAST, c.updated_at DESC
	`, userID).Scan(&rows).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	result := make([]models.InboxConversationResponse, 0, len(rows))
	for _, item := range rows {
		conversation := models.ConversationResponse{
			ID:            item.ConversationID,
			Type:          item.ConversationType,
			DirectKey:     item.DirectKey,
			Name:          item.Name,
			CreatedBy:     item.CreatedBy,
			LastMessageID: item.LastMessageID,
			LastMessageAt: item.LastMessageAt,
			Metadata:      item.ConversationMeta,
			CreatedAt:     item.ConversationCreated,
			UpdatedAt:     item.ConversationUpdated,
		}
		member := models.ConversationMemberResponse{
			ID:                item.MemberID,
			ConversationID:    item.ConversationID,
			UserID:            userID,
			Role:              item.MemberRole,
			Status:            item.MemberStatus,
			JoinedAt:          item.MemberJoined,
			LeftAt:            item.MemberLeftAt,
			LastReadMessageID: item.LastReadMessageID,
			LastReadAt:        item.LastReadAt,
			MutedUntil:        item.MutedUntil,
		}
		var lastMessage *models.MessageResponse
		if item.MessageID != nil {
			lastMessage = &models.MessageResponse{
				ID:               *item.MessageID,
				ConversationID:   item.ConversationID,
				SenderID:         derefUint64(item.MessageSenderID),
				Type:             derefString(item.MessageType),
				Body:             derefString(item.MessageBody),
				ReplyToMessageID: item.MessageReplyTo,
				ClientMessageID:  item.MessageClientID,
				Metadata:         item.MessageMeta,
				CreatedAt:        derefTime(item.MessageCreatedAt),
				UpdatedAt:        derefTime(item.MessageUpdatedAt),
				EditedAt:         item.MessageEditedAt,
				DeletedAt:        item.MessageDeletedAt,
			}
		}
		result = append(result, models.InboxConversationResponse{
			Conversation: conversation,
			LastMessage:  lastMessage,
			Member:       member,
			UnreadCount:  item.UnreadCount,
		})
	}

	return result, nil
}

func derefTime(v *time.Time) time.Time {
	if v == nil {
		return time.Time{}
	}
	return *v
}

func derefUint64(v *uint64) uint64 {
	if v == nil {
		return 0
	}
	return *v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
