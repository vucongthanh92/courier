package wallet

import (
	"fmt"

	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) GetOrCreate(userID uint64) (*entities.Wallet, error) {
	var result entities.Wallet
	err := r.db.Where("user_id = ? AND currency = ?", userID, "VND").First(&result).Error
	if err == nil {
		return &result, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	walletID, err := utils.NewSnowflakeID()
	if err != nil {
		return nil, err
	}
	accountID, err := utils.NewSnowflakeID()
	if err != nil {
		return nil, err
	}
	result = entities.Wallet{ID: walletID, UserID: userID, Currency: "VND", Status: "active"}
	return &result, r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		if err := tx.Create(&entities.WalletBalance{WalletID: walletID, Currency: "VND"}).Error; err != nil {
			return err
		}
		return tx.Create(&entities.LedgerAccount{ID: accountID, AccountCode: fmt.Sprintf("liability:wallet:%d:vnd", walletID), AccountType: "liability", Currency: "VND", WalletID: &walletID, NormalSide: "credit", IsActive: true}).Error
	})
}
