package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	httpcommon "github.com/vucongthanh92/courier/user-service/helper/http_common"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type IdentityHandler struct {
	identityService interfaces.IdentityUseCaseI
}

func InitIdentityHandler(
	identityService interfaces.IdentityUseCaseI,
) *IdentityHandler {
	return &IdentityHandler{
		identityService: identityService,
	}
}

// @Tags Auth
// @Summary OAuth callback (GitHub code flow)
// @Param provider path string true "github"
// @Param code query string true "code from GitHub"
// @Param state query string false "csrf state"
// @Success 200 {object} models.OAuthLoginResponse
// @Router /api/v1/auth/3rd/{provider}/callback [get]
func (h *IdentityHandler) OAuthCallback(c *gin.Context) {

	// Validate request body
	req := models.OAuthCallbackRequest{}
	if err := httpcommon.GetQueryParamsHTTP(c, &req); err != nil {
		errHandler.InitErrorBuilder(c).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "invalid_request", Message: err.Error()}).
			ExposeHttpError(c)
		return
	}

	// Get provider from path param and set to request
	if req.Code == "" {
		req.Code = c.Query("code")
		req.State = c.Query("state")
		req.RedirectURI = c.Query("redirect_uri")
	}

	// Get provider from path param and set to request
	req.Provider = c.Param("provider")
	if req.Code == "" {
		errHandler.InitErrorBuilder(c).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "missing_code", Message: "code required"}).
			ExposeHttpError(c)
		return
	}

	// Call usecase
	ctx := utils.SetHeaderByKey(c, "headers")
	res, resErr := h.identityService.OAuthCallback(ctx, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	// Return response
	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// @Tags Auth
// @Summary OAuth login/signup via provider
// @Param provider path string true "google or github"
// @Param params body models.OAuthLoginRequest true "OAuth token"
// @Success 200 {object} models.OAuthLoginResponse
// @Router /api/v1/auth/3rd/{provider} [post]
func (h *IdentityHandler) OAuthLogin(c *gin.Context) {

	// Validate request body
	req := models.OAuthLoginRequest{}
	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	// Get provider from path param and set to request
	req.Provider = c.Param("provider")
	if err := httpcommon.ValidatorParams(req); err != nil {
		errHandler.InitErrorBuilder(c).
			SetStatus(http.StatusBadRequest).
			SetArrayError(err).
			ExposeHttpError(c)
		return
	}

	// Call usecase
	ctx := utils.SetHeaderByKey(c, "headers")
	res, resErr := h.identityService.OAuthLogin(ctx, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	// Return response
	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}
