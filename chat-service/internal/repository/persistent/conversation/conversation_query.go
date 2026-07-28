package conversation

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/database"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	"gorm.io/gorm"
)

type repoQueryConversation struct {
	readDB *gorm.DB
}

func InitConversationQueryRepo(readDb *database.GormReadDb, writeDb *database.GormWriteDb) interfaces.ConversationQueryRepoI {
	return &repoQueryConversation{
		readDB: *readDb,
	}
}

func (r *repoQueryConversation) GetDirectConversationByKey(ctx context.Context, directKey string) (*entities.Conversation, *errHandler.ErrorBuilder) {
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

func (r *repoQueryConversation) GetConversationByID(ctx context.Context, id uint64) (*entities.Conversation, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.Conversation
	err := run.Model(&entities.Conversation{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}

// func ListConversationsByMember
func (r *repoQueryConversation) ListConversationsByMember(ctx context.Context, req *models.ListConversationsRequest) (
	[]models.ConversationListResponse, *errHandler.ErrorBuilder) {

	run := transaction.RunnerFromCtx(ctx, r.readDB)
	query := run.Table(`"chat-service".conversations AS c`).
		Select(`
			c.id,
			c.type,
			c.direct_key,
			c.name,
			c.created_by,
			c.last_message_id,
			c.last_message_at,
			c.metadata,
			c.created_at,
			c.updated_at,
			m.id AS message_id,
			m.conversation_id AS message_conversation_id,
			m.sender_id AS message_sender_id,
			m.type AS message_type,
			m.body AS message_body,
			m.reply_to_message_id AS message_reply_to_message_id,
			m.client_message_id AS message_client_message_id,
			m.metadata AS message_metadata,
			m.created_at AS message_created_at,
			m.updated_at AS message_updated_at,
			m.edited_at AS message_edited_at
		`).
		Joins(`JOIN "chat-service".conversation_members AS cm ON cm.conversation_id = c.id`).
		Joins(`LEFT JOIN "chat-service".messages AS m ON m.id = c.last_message_id AND m.deleted_at IS NULL`).
		Where("cm.user_id = ? AND cm.status = ? AND c.deleted_at IS NULL", req.RequesterID, "active").
		Order("COALESCE(c.last_message_at, c.created_at) DESC, c.id DESC")

	if req.BeforeLastMessageAt != nil && req.BeforeConversationID != nil {
		query = query.Where(
			"(COALESCE(c.last_message_at, c.created_at), c.id) < (?, ?)",
			*req.BeforeLastMessageAt,
			*req.BeforeConversationID,
		)
	}

	var rows []models.ConversationWithLastMessage
	if err := query.Limit(req.Limit).Scan(&rows).Error; err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	res := make([]models.ConversationListResponse, 0, len(rows))
	for i := range rows {
		var item models.ConversationListResponse
		item.MapConversationFromDB(rows[i])
		res = append(res, item)
	}

	return res, nil
}
