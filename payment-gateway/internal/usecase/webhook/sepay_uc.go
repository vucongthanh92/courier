package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/helper/constants"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
	"github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/providers/sepay"
)

type sePayWebhookUsecase struct {
	txn         *transaction.ManagerTxn
	provider    *sepay.Provider
	events      interfaces.ProviderEventCommandRepoI
	topups      interfaces.TopUpQueryRepoI
	topupCmd    interfaces.TopUpCommandRepoI
	wallets     interfaces.WalletCommandRepoI
	ledgerQuery interfaces.LedgerQueryRepoI
	ledgerCmd   interfaces.LedgerCommandRepoI
	outbox      interfaces.OutboxCommandRepoI
	auditLog    interfaces.AuditLogServiceI
}

func InitSePayWebhookUsecase(
	txn *transaction.ManagerTxn,
	provider *sepay.Provider,
	events interfaces.ProviderEventCommandRepoI,
	topups interfaces.TopUpQueryRepoI,
	topupCmd interfaces.TopUpCommandRepoI,
	wallets interfaces.WalletCommandRepoI,
	ledgerQuery interfaces.LedgerQueryRepoI,
	ledgerCmd interfaces.LedgerCommandRepoI,
	outbox interfaces.OutboxCommandRepoI,
	auditLog interfaces.AuditLogServiceI,
) interfaces.SePayWebhookServiceI {
	return &sePayWebhookUsecase{
		txn:         txn,
		provider:    provider,
		events:      events,
		topups:      topups,
		topupCmd:    topupCmd,
		wallets:     wallets,
		ledgerQuery: ledgerQuery,
		ledgerCmd:   ledgerCmd,
		outbox:      outbox,
		auditLog:    auditLog,
	}
}

// ProcessBankWebhook processes a webhook from SePay for a bank transfer.
// It checks the validity of the webhook, updates the top-up intent, credits the user's wallet,
// and creates necessary ledger entries and outbox events.
func (u *sePayWebhookUsecase) ProcessBankWebhook(ctx context.Context, req models.SePayBankWebhook, raw []byte) (
	result models.WebhookResult, resErr *errHandler.ErrorBuilder) {

	// Validate the webhook signature to ensure it is from SePay
	if req.ID <= 0 || req.TransferAmount <= 0 || !u.provider.IsReceivingAccount(req.AccountNumber) {
		return models.WebhookIgnored, nil
	}

	// Generate a unique event ID for the webhook processing
	eventID, e := utils.NewSnowflakeID()
	if e != nil {
		return result, errHandler.InitErrorBuilder(ctx).ValidateError(e)
	}

	// Create a new provider event for the webhook and check if it already exists
	event := entities.ProviderEvent{}
	event.Initialize(eventID, constants.SePayProvider, req.ID, raw)

	// Process the webhook within a transaction to ensure atomicity
	err := u.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {

		// Create a new provider event for the webhook and check if it already exists
		created, commonErr := u.events.CreateIfAbsent(txCtx, &event)
		if commonErr != nil {
			return commonErr
		}
		if !created {
			result = models.WebhookDuplicate
			return nil
		}

		// Check if the webhook is eligible for processing based on its transfer type and code
		if req.TransferType != constants.TransferTypeIn || req.Code == nil || *req.Code == "" {
			return u.ProcessIgnore(txCtx, eventID, constants.TRANSACTION_NOT_ELIGIBLE_CODE, &result)
		}

		// Retrieve the top-up intent associated with the payment code and validate its status
		intent, commonErr := u.topups.GetForUpdateByPaymentCode(txCtx, constants.SePayProvider, *req.Code)
		if commonErr != nil {
			if commonErr.Status == http.StatusNotFound {
				return u.ProcessIgnore(txCtx, eventID, constants.TOPUP_INTENT_NOT_FOUND_CODE, &result)
			}
			return commonErr
		}

		// Check if the top-up intent has already been marked as succeeded or is not eligible for processing
		if intent.CheckIsSucceeded() {
			return u.ProcessIgnore(txCtx, eventID, constants.TOPUP_ALREADY_SUCCEEDED_CODE, &result)
		}

		// Validate the top-up intent's status and expiration before proceeding with the wallet crediting process
		if intent.CheckStatusAndExpires() {
			return u.ProcessIgnore(txCtx, eventID, "topup_not_payable", &result)
		}

		// Check if the transfer amount in the webhook matches the expected amount in the top-up intent
		if intent.CheckAmountMinor(req.TransferAmount) {
			return u.ProcessIgnore(txCtx, eventID, constants.AMOUNT_MISMATCH_CODE, &result)
		}

		// Retrieve the wallet's liability account and the clearing account for ledger entries
		walletAccount, commonErr := u.ledgerQuery.GetWalletLiabilityAccount(txCtx, intent.WalletID)
		if commonErr != nil {
			return commonErr
		}

		// Retrieve the clearing account for ledger entries
		clearing, commonErr := u.ledgerQuery.GetAccountByCode(txCtx, "asset:sepay:clearing:vnd")
		if commonErr != nil {
			return commonErr
		}

		// Generate unique IDs for the journal and ledger entries to ensure proper tracking of the top-up transaction
		journalID, e := utils.NewSnowflakeID()
		if e != nil {
			return errHandler.InitErrorBuilder(txCtx).ValidateError(e)
		}

		// Create ledger journal entries for the top-up transaction, including debit and credit entries
		debitID, _ := utils.NewSnowflakeID()
		creditID, _ := utils.NewSnowflakeID()
		providerTxnID, _ := utils.NewSnowflakeID()
		outboxID, _ := utils.NewSnowflakeID()
		source := constants.SePayProvider

		journal := &entities.LedgerJournal{}
		journal.Initialize(
			journalID,
			constants.SePayRefTypeBankTransaction,
			req.ID,
			constants.SourceTypeExternalProvider,
			&source,
			constants.LedgerJournalStatusPosted,
			req.Code,
		)

		// Create ledger entries for the debit and credit sides of the top-up transaction
		ledgerEntries := make([]entities.LedgerEntry, 2)
		ledgerEntries[0].Initialize(debitID, journalID, clearing.ID, constants.NormalSideDebit, req.TransferAmount, constants.CurrencyVND)
		ledgerEntries[1].Initialize(creditID, journalID, walletAccount.ID, constants.NormalSideCredit, req.TransferAmount, constants.CurrencyVND)

		// Create the journal and ledger entries in the database to record the top-up transaction
		if commonErr = u.ledgerCmd.CreateJournalEntries(txCtx, journal, ledgerEntries); commonErr != nil {
			return commonErr
		}

		// Parse the transaction date from the webhook and create a provider transaction record in the database
		paidAt, parseErr := time.ParseInLocation(
			constants.DateTimeFormatYearMonthDay,
			req.TransactionDate,
			time.FixedZone(constants.GlobalTimeZoneHoChiMinh, 7*3600),
		)
		if parseErr != nil {
			return errHandler.InitErrorBuilder(txCtx).
				SetStatus(http.StatusBadRequest).SetLogError(parseErr).
				SetError(models.ErrorDTO{
					Code:    "invalid_provider_date",
					Message: "invalid SePay transaction date"},
				)
		}

		// Create a provider transaction record in the database to track the top-up transaction with SePay
		metadata, _ := json.Marshal(req)
		providerTxn := &entities.ProviderTransaction{}
		providerTxn.Initialize(
			providerTxnID,
			constants.SePayProvider,
			req.ID,
			intent.ID,
			req.TransferAmount,
			constants.CurrencyVND,
			&paidAt,
			&req.AccountNumber,
			metadata,
		)

		// Create the provider transaction record in the database to track the top-up transaction with SePay
		if commonErr = u.ledgerCmd.CreateProviderTransaction(txCtx, providerTxn); commonErr != nil {
			return commonErr
		}

		// Mark the top-up intent as succeeded, credit the user's wallet, and create an outbox event for further processing
		if commonErr = u.topupCmd.MarkSucceeded(txCtx, intent.ID, req.AccountNumber); commonErr != nil {
			return commonErr
		}

		// Credit the user's wallet with the transfer amount from the top-up intent
		if commonErr = u.wallets.CreditAvailable(txCtx, intent.WalletID, req.TransferAmount); commonErr != nil {
			return commonErr
		}

		// Create an outbox event to notify other services of the wallet crediting operation
		eventPayload, _ := json.Marshal(map[string]any{
			"user_id":         intent.UserID,
			"wallet_id":       intent.WalletID,
			"topup_intent_id": intent.ID,
			"amount_minor":    req.TransferAmount,
			"currency":        constants.CurrencyVND,
			"provider":        constants.SePayProvider,
		})

		outbox := &entities.OutboxEvent{}
		outbox.Initialize(
			outboxID,
			constants.OutboxAggregateTypeWallet,
			intent.WalletID,
			constants.OutboxEventTypeCredit,
			eventPayload,
		)

		// Create the outbox event in the database to notify other services of the wallet crediting operation
		if commonErr = u.outbox.Create(txCtx, outbox); commonErr != nil {
			return commonErr
		}

		auditMetadata, marshalErr := json.Marshal(map[string]any{
			"amount_minor":            req.TransferAmount,
			"currency":                constants.CurrencyVND,
			"journal_id":              journalID,
			"payment_code":            req.Code,
			"provider":                constants.SePayProvider,
			"provider_event_id":       req.ID,
			"provider_transaction_id": providerTxn.ID,
			"reference_code":          req.ReferenceCode,
		})
		if marshalErr != nil {
			return errHandler.InitErrorBuilder(txCtx).ValidateError(marshalErr)
		}

		providerActorID := constants.SePayProvider
		if commonErr = u.auditLog.Write(txCtx, models.AuditLogInput{
			ActorType:    "provider",
			ActorID:      &providerActorID,
			Action:       "wallet.topup.succeeded",
			ResourceType: "topup_intent",
			ResourceID:   fmt.Sprint(intent.ID),
			Metadata:     auditMetadata,
		}); commonErr != nil {
			return commonErr
		}

		// Mark the provider event as processed to prevent duplicate processing of the same webhook
		if commonErr = u.events.MarkProcessed(txCtx, eventID); commonErr != nil {
			return commonErr
		}

		result = models.WebhookCredited
		return nil
	})

	if err != nil {
		var transactionErr *transaction.Error
		if errors.As(err, &transactionErr) && transactionErr.Builder != nil {
			return result, transactionErr.Builder
		}
		return result, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return result, nil
}

// ProcessIgnore marks the provider event as ignored and sets the result to WebhookIgnored.
func (u *sePayWebhookUsecase) ProcessIgnore(ctx context.Context, eventID uint64, code string, result *models.WebhookResult) *errHandler.ErrorBuilder {
	if err := u.events.MarkIgnored(ctx, eventID, code); err != nil {
		return err
	}

	metadata, marshalErr := json.Marshal(map[string]string{
		"error_code": code,
		"provider":   constants.SePayProvider,
	})
	if marshalErr != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(marshalErr)
	}

	providerActorID := constants.SePayProvider
	if err := u.auditLog.Write(ctx, models.AuditLogInput{
		ActorType:    "provider",
		ActorID:      &providerActorID,
		Action:       "wallet.topup.ignored",
		ResourceType: "provider_event",
		ResourceID:   fmt.Sprint(eventID),
		Metadata:     metadata,
	}); err != nil {
		return err
	}

	*result = models.WebhookIgnored
	return nil
}
