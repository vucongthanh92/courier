package topup

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
)

type topUpUsecase struct {
	txn         *transaction.ManagerTxn
	wallet      interfaces.WalletServiceI
	topUpCmd    interfaces.TopUpCommandRepoI
	idempotency interfaces.IdempotencyServiceI
	resolver    interfaces.PaymentGatewayResolverI
}

func InitTopUpUsecase(
	txn *transaction.ManagerTxn,
	wallet interfaces.WalletServiceI,
	topUpCmd interfaces.TopUpCommandRepoI,
	idempotency interfaces.IdempotencyServiceI,
	resolver interfaces.PaymentGatewayResolverI,
) interfaces.TopUpServiceI {
	return &topUpUsecase{
		txn:         txn,
		wallet:      wallet,
		topUpCmd:    topUpCmd,
		idempotency: idempotency,
		resolver:    resolver,
	}
}

// CreateTopUp creates a pending SePay top-up intent and returns the signed checkout form.
func (u *topUpUsecase) CreateTopUp(ctx context.Context, req models.CreateTopUpRequest) (
	response models.CheckoutInstruction, replayed bool, resErr *errHandler.ErrorBuilder) {

	// Run the top-up creation process within a transaction to ensure atomicity
	if err := u.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {

		// Claim an idempotency key to ensure that the same request is not processed multiple times
		claim, txnErr := u.idempotency.Claim(txCtx, models.IdempotencyClaimRequest{
			Scope:     constants.IdempotencyScope,
			UserID:    req.UserID,
			Key:       req.IdempotencyKey,
			Request:   req,
			ExpiresIn: constants.Time_Cache_1_day,
		})
		if txnErr != nil {
			return txnErr
		}

		// If the idempotency record was not created, it means that the same request has already been processed
		if !claim.Created {
			// If the idempotency record exists and has a response, return the stored response to avoid duplicate processing
			if err := json.Unmarshal(claim.ResponseBody, &response); err != nil {
				return errHandler.InitErrorBuilder(txCtx).ValidateError(err)
			}

			replayed = true
			return nil
		}

		gateway, txnErr := u.resolver.Resolve(txCtx, req.ProviderName)
		if txnErr != nil {
			return txnErr
		}

		// Get or create the user's wallet for the top-up operation
		wallet, txnErr := u.wallet.GetOrCreateWallet(txCtx, req.UserID)
		if txnErr != nil {
			return txnErr
		}

		// Check if the wallet is active before proceeding with the top-up
		if wallet.Status != constants.WalletStatusActive {
			return errHandler.InitErrorBuilder(txCtx).SetStatus(http.StatusUnprocessableEntity).
				SetError(models.ErrorDTO{
					Code:    constants.WALLET_INACTIVE_CODE,
					Message: constants.ErrWalletInactive,
				})
		}

		// Generate a unique invoice number for the top-up operation and create a checkout with the payment gateway
		topupID, idErr := utils.NewSnowflakeID()
		if idErr != nil {
			return errHandler.InitErrorBuilder(txCtx).ValidateError(idErr)
		}

		// Create a unique invoice number for the top-up operation and create a checkout with the payment gateway
		invoice := fmt.Sprintf("CRTOP_%d", topupID)
		paymentCode, paymentCodeErr := utils.NewSePayPaymentCode(
			constants.SePayPaymentCodePrefix,
			constants.SePayPaymentCodeLength,
		)
		if paymentCodeErr != nil {
			return errHandler.InitErrorBuilder(txCtx).ValidateError(paymentCodeErr)
		}

		// Call the payment gateway to create a top-up checkout instruction with the generated invoice number and payment code
		checkout, gatewayErr := gateway.CreateTopUp(txCtx, interfaces.CreateTopUpInput{
			InvoiceNumber: invoice,
			AmountMinor:   req.AmountMinor,
			Currency:      constants.CurrencyVND,
			Method:        req.Method,
			CustomerID:    fmt.Sprint(req.UserID),
			Description:   fmt.Sprintf("%s %s", constants.CourierWalletTopUpDescription, paymentCode),
		})
		if gatewayErr != nil {
			return errHandler.InitErrorBuilder(txCtx).SetStatus(http.StatusUnprocessableEntity).
				SetLogError(gatewayErr).SetError(models.ErrorDTO{
				Code:    constants.PROVIDER_CHECKOUT_FAILED_CODE,
				Message: constants.ErrProviderSepayCheckoutFailed,
			})
		}

		// Create a new top-up intent in the database with the generated invoice number and checkout details
		intent := &entities.TopUpIntent{
			ID:                    topupID,
			UserID:                req.UserID,
			WalletID:              wallet.ID,
			AmountMinor:           req.AmountMinor,
			Currency:              constants.CurrencyVND,
			Provider:              gateway.Name(),
			Method:                req.Method,
			Status:                constants.TopupIntentStatusPending,
			ProviderInvoiceNumber: invoice,
			PaymentCode:           &paymentCode,
			ExpiresAt:             time.Now().UTC().Add(constants.Time_Cache_15_minutes),
			Metadata:              []byte(`{}`),
		}
		if txnErr = u.topUpCmd.Create(txCtx, intent); txnErr != nil {
			return txnErr
		}

		// Prepare the response with the checkout instruction and top-up details to be returned to the client
		response = models.CheckoutInstruction{
			TopUpID: fmt.Sprint(topupID), InvoiceNumber: invoice, PaymentCode: paymentCode,
			ExpiresAt:      intent.ExpiresAt.Format(time.RFC3339),
			CheckoutAction: checkout.Action, CheckoutFields: checkout.Fields,
		}

		return u.idempotency.Complete(txCtx, claim.Record.ID, http.StatusCreated, response)

	}); err != nil {
		var transactionErr *transaction.Error
		if errors.As(err, &transactionErr) && transactionErr.Builder != nil {
			return response, false, transactionErr.Builder
		}
		return response, false, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return response, replayed, nil
}
