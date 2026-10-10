package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	httpcommon "github.com/vucongthanh92/courier/user-service/helper/http_common"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/helper/constants"
	baseUtils "github.com/vucongthanh92/go-base-utils/helper/utils"
)

type SsoHandler struct {
	ssoService interfaces.SsoServiceI
	cookieName string
}

func InitSsoHandler(ssoService interfaces.SsoServiceI) *SsoHandler {
	return &SsoHandler{
		ssoService: ssoService,
		cookieName: constants.DefaultSsoCookieName,
	}
}

// Authorize godoc
// @Summary Authorize SSO request
// @Description Authorize SSO request
// @Tags SSO
// @Accept json
// @Produce json
// @Param client_id query string true "Client ID"
// @Param redirect_uri query string true "Redirect URI"
// @Param response_type query string true "Response Type"
// @Param scope query string false "Scope"
// @Param state query string false "State"
// @Success 302 {string} string "Redirect to the provided redirect_uri with authorization code"
// @Failure 400 {object} httpcommon.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} httpcommon.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpcommon.ErrorResponse "Forbidden"
// @Failure 500 {object} httpcommon.ErrorResponse "Internal server error"
// @Router /api/v1/sso/authorize [get]
func (h *SsoHandler) Authorize(c *gin.Context) {
	req := models.SsoAuthorizeRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpcommon.NewErrorResponse(err.Error(), "invalid_request", ""))
		return
	}

	ctx := baseUtils.SetHeaderByKey(c, "headers")
	res, resErr := h.ssoService.Authorize(ctx, req, h.sessionCookie(c))
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.Redirect(http.StatusFound, res.RedirectURI)
}

// Login godoc
// @Summary Login via SSO
// @Description Login via SSO
// @Tags SSO
// @Accept json
// @Produce json
// @Param loginRequest body models.SsoLoginRequest true "Login Request"
// @Success 200 {object} httpcommon.SuccessResponse{data=models.SsoLoginResponse} "Successful login"
// @Failure 400 {object} httpcommon.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} httpcommon.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpcommon.ErrorResponse "Forbidden"
// @Failure 500 {object} httpcommon.ErrorResponse "Internal server error"
// @Router /api/v1/sso/login [post]
func (h *SsoHandler) Login(c *gin.Context) {
	c.Header("Access-Control-Allow-Credentials", "true")
	req := models.SsoLoginRequest{}
	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	ctx := baseUtils.SetHeaderByKey(c, "headers")
	res, sessionToken, expiresAt, resErr := h.ssoService.Login(ctx, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	h.setSessionCookie(c, sessionToken, expiresAt)
	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// Token godoc
// @Summary Exchange authorization code for access token
// @Description Exchange authorization code for access token
// @Tags SSO
// @Accept json
// @Produce json
// @Param tokenRequest body models.SsoTokenRequest true "Token Request"
// @Success 200 {object} httpcommon.SuccessResponse{data=models.SsoTokenResponse} "Successful token exchange"
// @Failure 400 {object} httpcommon.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} httpcommon.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpcommon.ErrorResponse "Forbidden"
// @Failure 500 {object} httpcommon.ErrorResponse "Internal server error"
// @Router /api/v1/sso/token [post]
func (h *SsoHandler) Token(c *gin.Context) {
	req := models.SsoTokenRequest{}
	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	ctx := baseUtils.SetHeaderByKey(c, "headers")
	res, resErr := h.ssoService.Token(ctx, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// Session godoc
// @Summary Get current SSO session
// @Description Get current SSO session
// @Tags SSO
// @Accept json
// @Produce json
// @Success 200 {object} httpcommon.SuccessResponse{data=models.SsoSessionResponse} "Successful session retrieval"
// @Failure 400 {object} httpcommon.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} httpcommon.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpcommon.ErrorResponse "Forbidden"
// @Failure 500 {object} httpcommon.ErrorResponse "Internal server error"
// @Router /api/v1/sso/session [get]
func (h *SsoHandler) Session(c *gin.Context) {
	ctx := baseUtils.SetHeaderByKey(c, "headers")
	res, resErr := h.ssoService.Session(ctx, h.sessionCookie(c))
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// Logout godoc
// @Summary Logout from SSO
// @Description Logout from SSO
// @Tags SSO
// @Accept json
// @Produce json
// @Success 200 {object} httpcommon.SuccessResponse "Successful logout"
// @Failure 400 {object} httpcommon.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} httpcommon.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpcommon.ErrorResponse "Forbidden"
// @Failure 500 {object} httpcommon.ErrorResponse "Internal server error"
// @Router /api/v1/sso/logout [post]
func (h *SsoHandler) Logout(c *gin.Context) {
	ctx := baseUtils.SetHeaderByKey(c, "headers")
	res, resErr := h.ssoService.Logout(ctx, h.sessionCookie(c))
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	h.clearSessionCookie(c)
	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// sessionCookie retrieves the SSO session cookie from the request context.
func (h *SsoHandler) sessionCookie(c *gin.Context) string {
	value, err := c.Cookie(h.cookieName)
	if err != nil {
		return ""
	}
	return value
}

// setSessionCookie sets the SSO session cookie in the response context with the specified value and expiration time.
func (h *SsoHandler) setSessionCookie(c *gin.Context, value string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cookieName, value, maxAge, "/", "", false, true)
}

// clearSessionCookie clears the SSO session cookie from the response context.
func (h *SsoHandler) clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cookieName, "", -1, "/", "", false, true)
}
