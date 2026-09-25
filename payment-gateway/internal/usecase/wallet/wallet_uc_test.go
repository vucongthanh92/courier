package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"gorm.io/gorm"
)

type walletQueryRepositoryStub struct {
	snapshot *entities.WalletBalanceSnapshot
	err      *errHandler.ErrorBuilder
}

func (s *walletQueryRepositoryStub) GetByUserAndCurrency(context.Context, uint64, string) (*entities.Wallet, *errHandler.ErrorBuilder) {
	return nil, s.err
}

func (s *walletQueryRepositoryStub) GetBalanceByUserAndCurrency(context.Context, uint64, string) (*entities.WalletBalanceSnapshot, *errHandler.ErrorBuilder) {
	return s.snapshot, s.err
}

type walletCommandRepositoryStub struct{}

func (walletCommandRepositoryStub) CreateWallet(context.Context, *entities.Wallet, *entities.WalletBalance, *entities.LedgerAccount) *errHandler.ErrorBuilder {
	return nil
}

func (walletCommandRepositoryStub) CreditAvailable(context.Context, uint64, int64) *errHandler.ErrorBuilder {
	return nil
}

func TestGetBalanceReturnsWalletBalance(t *testing.T) {
	updatedAt := time.Date(2026, time.September, 24, 10, 0, 0, 0, time.UTC)
	usecase := InitWalletUsecase(&walletQueryRepositoryStub{
		snapshot: &entities.WalletBalanceSnapshot{
			WalletID:       101,
			Currency:       "VND",
			Status:         "active",
			AvailableMinor: 1000,
			PendingMinor:   200,
			HeldMinor:      300,
			UpdatedAt:      updatedAt,
		},
	}, walletCommandRepositoryStub{})

	response, commonErr := usecase.GetBalance(context.Background(), 42)

	require.Nil(t, commonErr)
	require.NotNil(t, response.WalletID)
	require.Equal(t, "101", *response.WalletID)
	require.Equal(t, "VND", response.Currency)
	require.Equal(t, "active", response.Status)
	require.EqualValues(t, 1000, response.AvailableMinor)
	require.EqualValues(t, 200, response.PendingMinor)
	require.EqualValues(t, 300, response.HeldMinor)
	require.Equal(t, updatedAt, *response.UpdatedAt)
}

func TestGetBalanceReturnsZeroForUserWithoutWallet(t *testing.T) {
	ctx := context.Background()
	usecase := InitWalletUsecase(&walletQueryRepositoryStub{
		err: errHandler.InitErrorBuilder(ctx).ValidateError(gorm.ErrRecordNotFound),
	}, walletCommandRepositoryStub{})

	response, commonErr := usecase.GetBalance(ctx, 42)

	require.Nil(t, commonErr)
	require.Nil(t, response.WalletID)
	require.Equal(t, "VND", response.Currency)
	require.Equal(t, "not_created", response.Status)
	require.Zero(t, response.AvailableMinor)
	require.Zero(t, response.PendingMinor)
	require.Zero(t, response.HeldMinor)
	require.Nil(t, response.UpdatedAt)
}
