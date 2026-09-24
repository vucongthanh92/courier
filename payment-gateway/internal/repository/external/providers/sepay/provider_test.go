package sepay

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vucongthanh92/courier/payment-gateway/config"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
)

func TestCreateTopUpBuildsSandboxCheckoutForm(t *testing.T) {
	provider := New(config.SePayConfig{
		MerchantID:  "sandbox-merchant",
		SecretKey:   "sandbox-secret",
		CheckoutURL: "https://pay-sandbox.sepay.vn/v1/checkout/init",
		SuccessURL:  "https://merchant.test/success",
		ErrorURL:    "https://merchant.test/error",
		CancelURL:   "https://merchant.test/cancel",
	})

	checkout, err := provider.CreateTopUp(context.Background(), interfaces.CreateTopUpInput{
		InvoiceNumber: "CRTOP_123", AmountMinor: 100000, Currency: "VND",
		Method: "bank_transfer", CustomerID: "42", Description: "Courier wallet top-up",
	})

	require.NoError(t, err)
	require.Equal(t, "https://pay-sandbox.sepay.vn/v1/checkout/init", checkout.Action)
	require.Equal(t, "BANK_TRANSFER", checkout.Fields["payment_method"])
	require.Equal(t, "100000", checkout.Fields["order_amount"])
	require.NotEmpty(t, checkout.Fields["signature"])
	require.Equal(t, checkout.Fields["signature"], sign(checkout.Fields, "sandbox-secret"))
}

func TestCreateTopUpRejectsUnsupportedMethod(t *testing.T) {
	provider := New(config.SePayConfig{MerchantID: "merchant", SecretKey: "secret"})
	_, err := provider.CreateTopUp(context.Background(), interfaces.CreateTopUpInput{Method: "wallet"})
	require.ErrorContains(t, err, "unsupported sepay method")
}
