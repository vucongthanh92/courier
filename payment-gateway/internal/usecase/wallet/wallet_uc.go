package wallet

import (
	"context"
	"fmt"
	"net/http"

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
func (u *walletUsecase) GetOrCreateWallet(
	ctx context.Context,
	userID uint64,
) (*entities.Wallet, *errHandler.ErrorBuilder) {
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

	wallet = &entities.Wallet{
		ID: walletID, UserID: userID, Currency: "VND", WalletType: "personal",
		Status: "active", RiskLevel: "normal", Metadata: []byte(`{}`),
	}
	balance := &entities.WalletBalance{WalletID: walletID, Currency: "VND"}
	account := &entities.LedgerAccount{
		ID: accountID, AccountCode: fmt.Sprintf("liability:wallet:%d:vnd", walletID),
		AccountType: "liability", Currency: "VND", WalletID: &walletID,
		NormalSide: "credit", IsActive: true,
	}
	if commonErr = u.walletCmd.CreateWallet(ctx, wallet, balance, account); commonErr != nil {
		return nil, commonErr
	}

	return wallet, nil
}
