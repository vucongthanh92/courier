package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
	"github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/sepay"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Result string

const (
	ResultCredited  Result = "credited"
	ResultDuplicate Result = "duplicate"
	ResultIgnored   Result = "ignored"
)

type Usecase struct {
	db       *gorm.DB
	provider *sepay.Provider
}

func New(db *gorm.DB, provider *sepay.Provider) *Usecase {
	return &Usecase{db: db, provider: provider}
}

// ProcessBankWebhook persists the provider evidence and credits a matching intent
// atomically. A valid but non-matching transaction is deliberately ignored and
// acknowledged so SePay does not retry it forever.
func (u *Usecase) ProcessBankWebhook(ctx context.Context, payload models.SePayBankWebhook, rawBody []byte) (Result, error) {
	if payload.ID <= 0 || payload.TransferAmount <= 0 || !u.provider.IsReceivingAccount(payload.AccountNumber) {
		return ResultIgnored, nil
	}
	providerEventID := strconv.FormatInt(payload.ID, 10)
	var result Result
	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entities.ProviderEvent
		err := tx.Where("provider = ? AND provider_event_id = ?", "sepay", providerEventID).First(&existing).Error
		if err == nil {
			result = ResultDuplicate
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		eventID, err := utils.NewSnowflakeID()
		if err != nil {
			return err
		}
		event := &entities.ProviderEvent{ID: eventID, Provider: "sepay", ProviderEventID: providerEventID, Payload: rawBody, SignatureValid: true, Status: "received"}
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		if payload.TransferType != "in" || payload.Code == nil || *payload.Code == "" {
			return u.ignoreEvent(tx, event.ID, "transaction_not_eligible", &result)
		}

		var intent entities.TopUpIntent
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider = ? AND payment_code = ?", "sepay", *payload.Code).First(&intent).Error
		if err == gorm.ErrRecordNotFound {
			return u.ignoreEvent(tx, event.ID, "topup_intent_not_found", &result)
		}
		if err != nil {
			return err
		}
		if intent.Status == "succeeded" {
			return u.ignoreEvent(tx, event.ID, "topup_already_succeeded", &result)
		}
		if intent.Status != "pending" || !intent.ExpiresAt.After(time.Now().UTC()) {
			return u.ignoreEvent(tx, event.ID, "topup_not_payable", &result)
		}
		if intent.AmountMinor != payload.TransferAmount {
			return u.ignoreEvent(tx, event.ID, "amount_mismatch", &result)
		}

		var balance entities.WalletBalance
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("wallet_id = ?", intent.WalletID).First(&balance).Error; err != nil {
			return err
		}
		walletAccount, err := walletLedgerAccount(tx, intent.WalletID)
		if err != nil {
			return err
		}
		clearingAccount, err := sepayClearingAccount(tx)
		if err != nil {
			return err
		}
		journalID, debitID, creditID, providerTxnID, outboxID, err := fiveIDs()
		if err != nil {
			return err
		}
		paidAt, err := time.ParseInLocation("2006-01-02 15:04:05", payload.TransactionDate, time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60))
		if err != nil {
			return fmt.Errorf("parse sepay transactionDate: %w", err)
		}
		receivingAccount := payload.AccountNumber
		if err := tx.Create(&entities.LedgerJournal{ID: journalID, ReferenceType: "sepay_bank_transaction", ReferenceID: providerEventID, Status: "posted", Narrative: "SePay wallet top-up " + *payload.Code}).Error; err != nil {
			return err
		}
		entries := []entities.LedgerEntry{{ID: debitID, JournalID: journalID, AccountID: clearingAccount.ID, Side: "debit", AmountMinor: payload.TransferAmount, Currency: "VND"}, {ID: creditID, JournalID: journalID, AccountID: walletAccount.ID, Side: "credit", AmountMinor: payload.TransferAmount, Currency: "VND"}}
		if err := tx.Create(&entries).Error; err != nil {
			return err
		}
		metadata, _ := json.Marshal(payload)
		if err := tx.Create(&entities.ProviderTransaction{ID: providerTxnID, Provider: "sepay", ProviderTransactionID: providerEventID, TopUpIntentID: intent.ID, AmountMinor: payload.TransferAmount, Currency: "VND", PaidAt: &paidAt, ReceivingAccountKey: &receivingAccount, SourceMetadata: metadata}).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := tx.Model(&entities.TopUpIntent{}).Where("id = ?", intent.ID).Updates(map[string]any{"status": "succeeded", "succeeded_at": now, "receiving_account_key": receivingAccount, "updated_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&entities.WalletBalance{}).Where("wallet_id = ?", intent.WalletID).Updates(map[string]any{"available_minor": balance.AvailableMinor + payload.TransferAmount, "version": balance.Version + 1, "updated_at": now}).Error; err != nil {
			return err
		}
		eventPayload, _ := json.Marshal(map[string]any{"user_id": intent.UserID, "wallet_id": intent.WalletID, "topup_intent_id": intent.ID, "amount_minor": payload.TransferAmount, "currency": "VND", "provider": "sepay"})
		if err := tx.Create(&entities.OutboxEvent{ID: outboxID, AggregateType: "wallet", AggregateID: strconv.FormatUint(intent.WalletID, 10), EventType: "payment.wallet_credited.v1", Payload: eventPayload}).Error; err != nil {
			return err
		}
		if err := tx.Model(&entities.ProviderEvent{}).Where("id = ?", event.ID).Updates(map[string]any{"status": "processed", "processed_at": now}).Error; err != nil {
			return err
		}
		result = ResultCredited
		return nil
	})
	return result, err
}

func (u *Usecase) ignoreEvent(tx *gorm.DB, eventID uint64, code string, result *Result) error {
	now := time.Now().UTC()
	if err := tx.Model(&entities.ProviderEvent{}).Where("id = ?", eventID).Updates(map[string]any{"status": "ignored", "error_code": code, "processed_at": now}).Error; err != nil {
		return err
	}
	*result = ResultIgnored
	return nil
}

func walletLedgerAccount(tx *gorm.DB, walletID uint64) (*entities.LedgerAccount, error) {
	var account entities.LedgerAccount
	err := tx.Where("wallet_id = ? AND account_type = ?", walletID, "liability").First(&account).Error
	return &account, err
}

func sepayClearingAccount(tx *gorm.DB) (*entities.LedgerAccount, error) {
	const code = "asset:sepay:clearing:vnd"
	var account entities.LedgerAccount
	err := tx.Where("account_code = ?", code).First(&account).Error
	if err == nil {
		return &account, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	id, err := utils.NewSnowflakeID()
	if err != nil {
		return nil, err
	}
	account = entities.LedgerAccount{ID: id, AccountCode: code, AccountType: "asset", Currency: "VND", NormalSide: "debit", IsActive: true}
	if err := tx.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func fiveIDs() (uint64, uint64, uint64, uint64, uint64, error) {
	ids := make([]uint64, 5)
	for i := range ids {
		id, err := utils.NewSnowflakeID()
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}
		ids[i] = id
	}
	return ids[0], ids[1], ids[2], ids[3], ids[4], nil
}
