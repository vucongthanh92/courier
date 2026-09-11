package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vucongthanh92/courier/payment-gateway/helper/constants"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	httpcommon "github.com/vucongthanh92/courier/payment-gateway/helper/http_common"
	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

type TopUpHandler struct {
	usecase interfaces.TopUpServiceI
}

func InitTopUpHandler(usecase interfaces.TopUpServiceI) *TopUpHandler {
	return &TopUpHandler{
		usecase: usecase,
	}
}

// API Create TopUp godoc
// @Tags TopUp
// @Summary creates a pending SePay top-up intent and returns the signed checkout form.
// @Accept json
// @Produce json
// @Param params body models.CreateTopUpRequest true "CreateTopUpRequest"
// @Router /api/v1/wallet/top-up [post]
// @Success	200 {object} models.CreateTopUpResponse
func (h *TopUpHandler) CreateTopUp(c *gin.Context) {
	var request models.CreateTopUpRequest

	// Extract the authenticated user's ID from the context
	claims := utils.ClaimsPayloadJWT(c)
	if claims == nil {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{
				Code:    "unauthorized",
				Message: "missing authenticated user",
			}).ExposeHttpError(c)
		return
	}
	request.UserID = utils.ParseUserID(claims["sub"])

	// Check for the presence of the Idempotency-Key header
	if c.GetHeader("Idempotency-Key") == "" {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    "bad_request",
				Message: "Idempotency-Key is required",
			}).ExposeHttpError(c)
		return
	}
	request.IdempotencyKey = c.GetHeader("Idempotency-Key")

	// Parse and validate the request body
	if err := c.ShouldBindJSON(&request); err != nil || request.AmountMinor <= 0 {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    "bad_request",
				Message: "amount_minor must be positive",
			}).ExposeHttpError(c)
		return
	}

	// Call the use case to create the top-up intent
	res, replayed, commonErr := h.usecase.CreateTopUp(c.Request.Context(), request)
	if commonErr != nil {
		commonErr.ExposeHttpError(c)
		return
	}

	// Return the response with appropriate status code
	status := http.StatusCreated
	if replayed {
		status = http.StatusOK
	}
	c.JSON(status, httpcommon.NewSuccessResponse(res))
}
