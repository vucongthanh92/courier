package entities

import (
	"time"

	"github.com/vucongthanh92/courier/chat-service/helper/utils"
)

type Conversation struct {
	ID            uint64         `gorm:"column:id;primaryKey;type:bigint;check:id>0" json:"id"`
	Type          string         `gorm:"column:type;type:\"chat-service\".conversation_type_enum;not null;index" json:"type"`
	DirectKey     *string        `gorm:"column:direct_key;type:text" json:"direct_key,omitempty"`
	Name          *string        `gorm:"column:name;type:varchar(255)" json:"name,omitempty"`
	CreatedBy     uint64         `gorm:"column:created_by;type:bigint;not null;index" json:"created_by"`
	LastMessageID *uint64        `gorm:"column:last_message_id;type:bigint" json:"last_message_id,omitempty"`
	LastMessageAt *time.Time     `gorm:"column:last_message_at;type:timestamptz;index" json:"last_message_at,omitempty"`
	Metadata      map[string]any `gorm:"column:metadata;type:jsonb;serializer:json;not null;default:'{}'" json:"metadata"`
	CreatedAt     time.Time      `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt     *time.Time     `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at,omitempty"`
}

func (Conversation) TableName() string {
	return `"chat-service".conversations`
}

// InitConversationEntity initializes a Conversation entity with the provided parameters,
// generating a unique ID and setting default values for other fields.
func (s *Conversation) InitConversationEntity(converType string, directKey *string, converName string, creatorID uint64) {
	s.ID, _ = utils.NewSnowflakeID()
	s.Type = converType
	s.DirectKey = directKey
	s.CreatedBy = creatorID
	s.Metadata = map[string]any{}

	if converName == "" {
		converName = utils.RandString(10)
	}
	s.Name = &converName
}
