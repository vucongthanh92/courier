package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vucongthanh92/courier/user-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	httpcommon "github.com/vucongthanh92/courier/user-service/helper/http_common"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type UserHandler struct {
	userService interfaces.UserServiceI
}

func InitUserHandler(userService interfaces.UserServiceI) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// API CreateUser godoc
// @Tags User
// @Summary create new user
// @Accept json
// @Produce json
// @Param params body models.CreateUserRequest true "CreateUserRequest"
// @Router /api/v1/user [post]
// @Success	200 {object} entities.User
func (h *UserHandler) CreateUser(c *gin.Context) {
	req := models.CreateUserRequest{}

	err := httpcommon.ValidatorParams(req)
	if err != nil {
		resErr := errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetArrayError(err)
		resErr.ExposeHttpError(c)
		return
	}

	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	res, resErr := h.userService.CreateUser(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}
