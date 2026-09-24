package sepay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vucongthanh92/courier/payment-gateway/config"
)

func TestVerifyBankWebhook(t *testing.T) {
	now := time.Unix(1787213762, 0)
	body := []byte(`{"id":26671,"transferAmount":1000}`)
	secret := "webhook-secret"
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("1787213762."))
	_, _ = mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	provider := New(config.SePayConfig{WebhookSecret: secret, TimestampToleranceSec: 300})
	require.NoError(t, provider.VerifyBankWebhook(body, signature, "1787213762", now))
	require.Error(t, provider.VerifyBankWebhook(body, "sha256=wrong", "1787213762", now))
}
