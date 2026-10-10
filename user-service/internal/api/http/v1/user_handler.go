package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	httpcommon "github.com/vucongthanh92/courier/user-service/helper/http_common"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/helper/constants"
	baseUtils "github.com/vucongthanh92/go-base-utils/helper/utils"
)

type UserHandler struct {
	userService interfaces.UserServiceI
}

func InitUserHandler(userService interfaces.UserServiceI) *UserHandler {
	return &UserHandler{userService: userService}
}

// API Search Users godoc
// @Tags User
// @Summary search verified users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param search_key query string true "Search text for display name, phone number, or email"
// @Param limit query int false "Limit, default 10 and max 20"
// @Router /api/v1/user/search [get]
// @Success 200 {object} httpcommon.SuccessResponse[[]models.SearchUserResponse]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	req := models.SearchUsersRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "invalid_query", Field: "search_key", Message: "search_key is required"}).
			ExposeHttpError(c)
		return
	}

	claimsValue, ok := c.Get("authClaims")
	if !ok {
		errHandler.InitErrorBuilder(c).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{Code: "unauthorized", Message: "Authentication is required"}).
			ExposeHttpError(c)
		return
	}
	claims := claimsValue.(jwt.MapClaims)
	req.ExcludeUserID = baseUtils.ParseUserID(claims["sub"])

	res, resErr := h.userService.SearchUsers(c.Request.Context(), req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}
