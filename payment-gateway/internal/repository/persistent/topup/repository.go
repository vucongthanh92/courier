package topup

import (
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func New(db *gorm.DB) *Repository                               { return &Repository{db: db} }
func (r *Repository) Create(intent *entities.TopUpIntent) error { return r.db.Create(intent).Error }
