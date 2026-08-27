package topup

import (
	"context"
	"fmt"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
	topuprepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/topup"
	walletrepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/wallet"
)

type Usecase struct {
	wallets *walletrepo.Repository
	intents *topuprepo.Repository
	gateway interfaces.PaymentGateway
}

func New(wallets *walletrepo.Repository, intents *topuprepo.Repository, gateway interfaces.PaymentGateway) *Usecase {
	return &Usecase{wallets: wallets, intents: intents, gateway: gateway}
}

func (u *Usecase) Create(ctx context.Context, userID uint64, req models.CreateTopUpRequest) (models.CheckoutInstruction, error) {
	wallet, err := u.wallets.GetOrCreate(userID)
	if err != nil {
		return models.CheckoutInstruction{}, fmt.Errorf("get wallet: %w", err)
	}
	if wallet.Status != "active" {
		return models.CheckoutInstruction{}, fmt.Errorf("wallet is not active")
	}
	id, err := utils.NewSnowflakeID()
	if err != nil {
		return models.CheckoutInstruction{}, err
	}
	invoice := fmt.Sprintf("CRTOP_%d", id)
	intent := &entities.TopUpIntent{ID: id, UserID: userID, WalletID: wallet.ID, AmountMinor: req.AmountMinor, Currency: "VND", Provider: u.gateway.Name(), Method: req.Method, Status: "pending", ProviderInvoiceNumber: invoice, PaymentCode: &invoice, ExpiresAt: time.Now().UTC().Add(15 * time.Minute), Metadata: []byte(`{}`)}
	checkout, err := u.gateway.CreateTopUp(ctx, interfaces.CreateTopUpInput{InvoiceNumber: invoice, AmountMinor: req.AmountMinor, Currency: "VND", Method: req.Method, CustomerID: fmt.Sprint(userID), Description: "Courier wallet top-up"})
	if err != nil {
		return models.CheckoutInstruction{}, err
	}
	if err := u.intents.Create(intent); err != nil {
		return models.CheckoutInstruction{}, fmt.Errorf("create topup intent: %w", err)
	}
	return models.CheckoutInstruction{TopUpID: fmt.Sprint(id), InvoiceNumber: invoice, ExpiresAt: intent.ExpiresAt.Format(time.RFC3339), CheckoutAction: checkout.Action, CheckoutFields: checkout.Fields}, nil
}
