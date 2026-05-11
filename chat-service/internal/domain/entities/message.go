package entities

import "time"

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

func (m Message) IsText() bool {
	return m.Type == MessageTypeText
}
