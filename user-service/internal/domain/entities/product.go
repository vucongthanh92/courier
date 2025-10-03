package entities

import "time"

type Product struct {
	ID         uint    `gorm:"primaryKey"`
	Name       string  `gorm:"type:varchar(255);not null"`
	CategoryID uint    `gorm:"not null"`
	SupplyID   uint    `gorm:"not null"`
	Price      float64 `gorm:"type:decimal(10,2);not null"`
	Stock      int     `gorm:"not null;default:0"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Category   Category `gorm:"foreignKey:CategoryID"`
	Supply     Supply   `gorm:"foreignKey:SupplyID"`
}
