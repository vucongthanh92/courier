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

type WalletHandler struct {
	usecase interfaces.WalletServiceI
}

func InitWalletHandler(usecase interfaces.WalletServiceI) *WalletHandler {
	return &WalletHandler{
		usecase: usecase,
	}
}

// GetBalance godoc
// @Tags Wallet
// @Summary Gets the authenticated user's Courier wallet balance.
// @Produce json
// @Router /api/v1/wallet/balance [get]
// @Success 200 {object} models.WalletBalanceResponse
func (h *WalletHandler) GetBalance(c *gin.Context) {
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

	response, commonErr := h.usecase.GetBalance(c.Request.Context(), utils.ParseUserID(claims["sub"]))
	if commonErr != nil {
		commonErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(response))
}
