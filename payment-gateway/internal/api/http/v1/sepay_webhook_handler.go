package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
	"github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/sepay"
)

type SePayWebhookHandler struct {
	provider *sepay.Provider
	usecase  interfaces.SePayWebhookServiceI
}

func InitSePayWebhookHandler(provider *sepay.Provider, usecase interfaces.SePayWebhookServiceI) *SePayWebhookHandler {
	return &SePayWebhookHandler{provider: provider, usecase: usecase}
}

// HealthCheck confirms that the public webhook route is reachable. It is kept
// signature-free so it can be tested through the ngrok public domain.
func (h *SePayWebhookHandler) HealthCheck(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Receive verifies the raw SePay payload before dispatching it to the wallet-credit usecase.
func (h *SePayWebhookHandler) Receive(c *gin.Context) {
	rawBody, err := c.GetRawData()
	if err != nil || len(rawBody) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid webhook body"})
		return
	}
	if err := h.provider.VerifyBankWebhook(rawBody, c.GetHeader("X-SePay-Signature"), c.GetHeader("X-SePay-Timestamp"), time.Now().UTC()); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
		return
	}
	var payload models.SePayBankWebhook
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid webhook JSON"})
		return
	}
	if _, commonErr := h.usecase.ProcessBankWebhook(c.Request.Context(), payload, rawBody); commonErr != nil {
		commonErr.ExposeHttpError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
