package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
	topupuc "github.com/vucongthanh92/courier/payment-gateway/internal/usecase/topup"
)

type TopUpHandler struct{ usecase *topupuc.Usecase }

func InitTopUpHandler(usecase *topupuc.Usecase) *TopUpHandler { return &TopUpHandler{usecase: usecase} }

// Create creates a pending SePay top-up intent and returns the signed checkout form.
func (h *TopUpHandler) Create(c *gin.Context) {
	userID, err := authenticatedUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "missing or invalid authenticated user"})
		return
	}
	if c.GetHeader("Idempotency-Key") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Idempotency-Key is required"})
		return
	}
	var request models.CreateTopUpRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.AmountMinor <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "amount_minor must be positive"})
		return
	}
	instruction, err := h.usecase.Create(c.Request.Context(), userID, request)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": instruction})
}

func authenticatedUserID(c *gin.Context) (uint64, error) {
	value, ok := c.Get("authClaims")
	if !ok {
		return 0, strconv.ErrSyntax
	}
	claims, ok := value.(jwt.MapClaims)
	if !ok {
		return 0, strconv.ErrSyntax
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseUint(sub, 10, 64)
}
