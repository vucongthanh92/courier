package entities

import "time"

type Supply struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"type:varchar(255);not null"`
	Contact   string `gorm:"type:varchar(255)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Products  []Product `gorm:"foreignKey:SupplyID"`
}
