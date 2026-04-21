package entities

import "time"

type ConversationMember struct {
	ID                uint64     `gorm:"column:id;primaryKey;type:bigint;check:id>0" json:"id"`
	ConversationID    uint64     `gorm:"column:conversation_id;type:bigint;not null;index" json:"conversation_id"`
	UserID            uint64     `gorm:"column:user_id;type:bigint;not null;index" json:"user_id"`
	Role              string     `gorm:"column:role;type:\"chat-service\".conversation_member_role_enum;not null;default:'member'" json:"role"`
	Status            string     `gorm:"column:status;type:\"chat-service\".conversation_member_status_enum;not null;default:'active'" json:"status"`
	JoinedAt          time.Time  `gorm:"column:joined_at;type:timestamptz;not null;default:now()" json:"joined_at"`
	LeftAt            *time.Time `gorm:"column:left_at;type:timestamptz" json:"left_at,omitempty"`
	LastReadMessageID *uint64    `gorm:"column:last_read_message_id;type:bigint" json:"last_read_message_id,omitempty"`
	LastReadAt        *time.Time `gorm:"column:last_read_at;type:timestamptz" json:"last_read_at,omitempty"`
	MutedUntil        *time.Time `gorm:"column:muted_until;type:timestamptz" json:"muted_until,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;type:timestamptz;autoUpdateTime" json:"updated_at"`
}

func (ConversationMember) TableName() string {
	return `"chat-service".conversation_members`
}

func (m ConversationMember) IsActive() bool {
	return m.Status == ConversationMemberStatusActive
}

func (m ConversationMember) IsOwner() bool {
	return m.Role == ConversationMemberRoleOwner
}
