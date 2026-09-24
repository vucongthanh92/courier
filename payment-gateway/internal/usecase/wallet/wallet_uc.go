package wallet

import (
	"context"
	"net/http"

	"github.com/vucongthanh92/courier/payment-gateway/helper/constants"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
)

type walletUsecase struct {
	walletQuery interfaces.WalletQueryRepoI
	walletCmd   interfaces.WalletCommandRepoI
}

func InitWalletUsecase(
	walletQuery interfaces.WalletQueryRepoI,
	walletCmd interfaces.WalletCommandRepoI,
) interfaces.WalletServiceI {
	return &walletUsecase{
		walletQuery: walletQuery,
		walletCmd:   walletCmd,
	}
}

// GetOrCreateWallet returns the user's VND wallet and creates its balance and
// liability ledger account together when this is the user's first wallet flow.
func (u *walletUsecase) GetOrCreateWallet(ctx context.Context, userID uint64,
) (*entities.Wallet, *errHandler.ErrorBuilder) {

	// Attempt to retrieve the user's existing wallet by user ID and currency
	wallet, commonErr := u.walletQuery.GetByUserAndCurrency(ctx, userID, "VND")
	if commonErr == nil {
		return wallet, nil
	}
	if commonErr.Status != http.StatusNotFound {
		return nil, commonErr
	}

	walletID, err := utils.NewSnowflakeID()
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	accountID, err := utils.NewSnowflakeID()
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	// Create a new wallet entity with the generated wallet ID, user ID,
	// and default values for currency, wallet type, status, risk level, and metadata
	wallet = &entities.Wallet{
		ID:         walletID,
		UserID:     userID,
		Currency:   constants.CurrencyVND,
		WalletType: constants.WalletTypePersonal,
		Status:     constants.WalletStatusActive,
		RiskLevel:  constants.RiskLevelNormal,
		Metadata:   []byte(`{}`),
	}

	// Create a new wallet balance for the user with the specified wallet ID and currency
	balance := &entities.WalletBalance{
		WalletID: walletID,
		Currency: constants.CurrencyVND,
	}

	// Initialize the liability ledger account for the wallet with the appropriate details
	account := entities.LedgerAccount{}
	account.Initialize(
		accountID,
		constants.LedgerAccountTypeLiability,
		constants.CurrencyVND,
		&walletID,
		constants.NormalSideCredit,
	)

	// Create the wallet, balance, and liability ledger account in a single transaction to ensure atomicity
	// This prevents partial creation of wallet components in case of errors
	if commonErr = u.walletCmd.CreateWallet(ctx, wallet, balance, &account); commonErr != nil {
		return nil, commonErr
	}

	return wallet, nil
}
