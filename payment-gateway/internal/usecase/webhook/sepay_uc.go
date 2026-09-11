package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/helper/constants"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
	"github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/sepay"
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

	eventID, e := utils.NewSnowflakeID()
	if e != nil {
		return result, errHandler.InitErrorBuilder(ctx).ValidateError(e)
	}

	// Create a new provider event for the webhook and check if it already exists
	event := &entities.ProviderEvent{
		ID:              eventID,
		Provider:        constants.SePayProvider,
		ProviderEventID: strconv.FormatInt(req.ID, 10),
		Payload:         raw,
		SignatureValid:  true,
		Status:          constants.SepayEventStatusReceived,
	}

	// Process the webhook within a transaction to ensure atomicity
	err := u.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {

		// Create a new provider event for the webhook and check if it already exists
		created, commonErr := u.events.CreateIfAbsent(txCtx, event)
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
		if intent.Status == constants.TopupIntentStatusSucceeded {
			return u.ProcessIgnore(txCtx, eventID, constants.TOPUP_ALREADY_SUCCEEDED_CODE, &result)
		}

		// Validate the top-up intent's status and expiration before proceeding with the wallet crediting process
		if intent.Status != constants.TopupIntentStatusPending || !intent.ExpiresAt.After(time.Now().UTC()) {
			return u.ProcessIgnore(txCtx, eventID, "topup_not_payable", &result)
		}

		// Check if the transfer amount in the webhook matches the expected amount in the top-up intent
		if intent.AmountMinor != req.TransferAmount {
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

		journal := &entities.LedgerJournal{
			ID:             journalID,
			ReferenceType:  constants.SePayRefTypeBankTransaction,
			ReferenceID:    strconv.FormatInt(req.ID, 10),
			SourceType:     "external_provider",
			SourceProvider: &source,
			Status:         "posted",
			Narrative:      "SePay wallet top-up " + *req.Code,
		}

		if commonErr = u.ledgerCmd.CreateJournalEntries(txCtx, journal, []entities.LedgerEntry{{ID: debitID, JournalID: journalID, AccountID: clearing.ID, Side: "debit", AmountMinor: req.TransferAmount, Currency: "VND"}, {ID: creditID, JournalID: journalID, AccountID: walletAccount.ID, Side: "credit", AmountMinor: req.TransferAmount, Currency: "VND"}}); commonErr != nil {
			return commonErr
		}

		paidAt, parseErr := time.ParseInLocation("2006-01-02 15:04:05", req.TransactionDate, time.FixedZone("Asia/Ho_Chi_Minh", 7*3600))
		if parseErr != nil {
			return errHandler.InitErrorBuilder(txCtx).SetStatus(http.StatusBadRequest).SetLogError(parseErr).SetError(models.ErrorDTO{Code: "invalid_provider_date", Message: "invalid SePay transaction date"})
		}
		metadata, _ := json.Marshal(req)
		if commonErr = u.ledgerCmd.CreateProviderTransaction(txCtx, &entities.ProviderTransaction{ID: providerTxnID, Provider: "sepay", ProviderTransactionID: strconv.FormatInt(req.ID, 10), TopUpIntentID: intent.ID, AmountMinor: req.TransferAmount, Currency: "VND", PaidAt: &paidAt, ReceivingAccountKey: &req.AccountNumber, SourceMetadata: metadata}); commonErr != nil {
			return commonErr
		}
		if commonErr = u.topupCmd.MarkSucceeded(txCtx, intent.ID, req.AccountNumber); commonErr != nil {
			return commonErr
		}
		if commonErr = u.wallets.CreditAvailable(txCtx, intent.WalletID, req.TransferAmount); commonErr != nil {
			return commonErr
		}
		eventPayload, _ := json.Marshal(map[string]any{"user_id": intent.UserID, "wallet_id": intent.WalletID, "topup_intent_id": intent.ID, "amount_minor": req.TransferAmount, "currency": "VND", "provider": "sepay"})
		if commonErr = u.outbox.Create(txCtx, &entities.OutboxEvent{ID: outboxID, AggregateType: "wallet", AggregateID: strconv.FormatUint(intent.WalletID, 10), EventType: "payment.wallet_credited.v1", Payload: eventPayload}); commonErr != nil {
			return commonErr
		}
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
	*result = models.WebhookIgnored
	return nil
}
