package sepay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/config"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
)

type Provider struct {
	cfg config.SePayConfig
}

func New(cfg config.SePayConfig) *Provider {
	return &Provider{cfg: cfg}
}

func (p *Provider) Name() string {
	return "sepay"
}

// VerifyBankWebhook verifies the SePay format: sha256=HMAC_HEX(timestamp.rawBody).
func (p *Provider) VerifyBankWebhook(rawBody []byte, signature, timestamp string, now time.Time) error {
	if p.cfg.WebhookSecret == "" {
		return fmt.Errorf("sepay webhook secret is not configured")
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || ts <= 0 {
		return fmt.Errorf("invalid sepay webhook timestamp")
	}
	if delta := now.Unix() - ts; delta > p.cfg.TimestampToleranceSec || delta < -p.cfg.TimestampToleranceSec {
		return fmt.Errorf("sepay webhook timestamp is outside the allowed window")
	}
	mac := hmac.New(sha256.New, []byte(p.cfg.WebhookSecret))
	_, _ = mac.Write([]byte(timestamp + "."))
	_, _ = mac.Write(rawBody)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("invalid sepay webhook signature")
	}
	return nil
}

func (p *Provider) IsReceivingAccount(accountNumber string) bool {
	for _, allowed := range p.cfg.ReceivingAccountNumbers {
		if allowed != "" && allowed == accountNumber {
			return true
		}
	}
	return false
}

func (p *Provider) CreateTopUp(_ context.Context, in interfaces.CreateTopUpInput) (interfaces.CheckoutInstruction, error) {
	if p.cfg.MerchantID == "" || p.cfg.SecretKey == "" {
		return interfaces.CheckoutInstruction{}, fmt.Errorf("sepay sandbox merchant credentials are not configured")
	}
	method, err := sepayMethod(in.Method)
	if err != nil {
		return interfaces.CheckoutInstruction{}, err
	}
	fields := map[string]string{
		"order_amount": strconv.FormatInt(in.AmountMinor, 10), "merchant": p.cfg.MerchantID,
		"currency": in.Currency, "operation": "PURCHASE", "order_description": in.Description,
		"order_invoice_number": in.InvoiceNumber, "customer_id": in.CustomerID, "payment_method": method,
		"success_url": p.cfg.SuccessURL, "error_url": p.cfg.ErrorURL, "cancel_url": p.cfg.CancelURL,
	}
	fields["signature"] = sign(fields, p.cfg.SecretKey)
	return interfaces.CheckoutInstruction{Action: p.cfg.CheckoutURL, Fields: fields}, nil
}

func sepayMethod(method string) (string, error) {
	switch method {
	case "bank_transfer":
		return "BANK_TRANSFER", nil
	case "napas_bank_transfer":
		return "NAPAS_BANK_TRANSFER", nil
	case "card":
		return "CARD", nil
	default:
		return "", fmt.Errorf("unsupported sepay method %q", method)
	}
}

func sign(fields map[string]string, secret string) string {
	ordered := []string{"order_amount", "merchant", "currency", "operation", "order_description", "order_invoice_number", "customer_id", "payment_method", "success_url", "error_url", "cancel_url"}
	parts := make([]string, 0, len(ordered))
	for _, key := range ordered {
		if value, ok := fields[key]; ok {
			parts = append(parts, key+"="+value)
		}
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.Join(parts, ",")))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
