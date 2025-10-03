package entities

import "time"

type Category struct {
	ID          uint64     `gorm:"column:id;primaryKey" json:"id"`
	Name        string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description string     `gorm:"column:description;type:text" json:"description"`
	Status      bool       `gorm:"column:status;type:tinyint" json:"status"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
	Products    []Product  `gorm:"foreignKey:category_id" json:"products,omitempty"`
}
