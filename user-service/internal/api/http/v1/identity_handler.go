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

type IdentityHandler struct {
	identityService interfaces.IdentityServiceI
}

func InitIdentityHandler(
	identityService interfaces.IdentityServiceI,
) *IdentityHandler {
	return &IdentityHandler{
		identityService: identityService,
	}
}

// API get identities list godoc
// @Tags Identity
// @Summary search products with filter and return pagination
// @Accept json
// @Produce json
// @Param  params body models.CreateCategoryReq true "CreateCategoryReq"
// @Router 	/api/v1/products [get]
// @Success	200
func (h *IdentityHandler) CreateIdentity(c *gin.Context) {
	req := models.CreateIdentityParams{}

	err := httpcommon.ValidatorParams(req)
	if err != nil {
		resErr := errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetArrayError(err)
		resErr.ExposeHttpError(c)
		return
	}

	res, resErr := h.identityService.CreateIdentity(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}
