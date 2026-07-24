package entities

import (
	"time"

	"github.com/vucongthanh92/courier/chat-service/helper/utils"
)

type Message struct {
	ID               uint64         `gorm:"column:id;primaryKey;type:bigint;check:id>0" json:"id"`
	ConversationID   uint64         `gorm:"column:conversation_id;type:bigint;not null;index" json:"conversation_id"`
	SenderID         uint64         `gorm:"column:sender_id;type:bigint;not null;index" json:"sender_id"`
	Type             string         `gorm:"column:type;type:\"chat-service\".message_type_enum;not null;default:'text'" json:"type"`
	Body             string         `gorm:"column:body;type:text;not null" json:"body"`
	ReplyToMessageID *uint64        `gorm:"column:reply_to_message_id;type:bigint" json:"reply_to_message_id,omitempty"`
	ClientMessageID  *string        `gorm:"column:client_message_id;type:varchar(64)" json:"client_message_id,omitempty"`
	Metadata         map[string]any `gorm:"column:metadata;type:jsonb;serializer:json;not null;default:'{}'" json:"metadata"`
	CreatedAt        time.Time      `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;type:timestamptz;autoUpdateTime" json:"updated_at"`
	EditedAt         *time.Time     `gorm:"column:edited_at;type:timestamptz" json:"edited_at,omitempty"`
	DeletedAt        *time.Time     `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at,omitempty"`
}

func (Message) TableName() string {
	return `"chat-service".messages`
}

func (m *Message) InitMessageEntity(conversationID, senderID uint64, messageType, body string, clientMessageID *string, replyToMessageID *uint64, metadata map[string]any) error {
	newMessageID, err := utils.NewSnowflakeID()
	if err != nil {
		return err
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	m.ID = newMessageID
	m.ConversationID = conversationID
	m.SenderID = senderID
	m.Type = messageType
	m.Body = body
	m.ClientMessageID = clientMessageID
	m.ReplyToMessageID = replyToMessageID
	m.Metadata = metadata

	return nil
}
