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

type AuthHandler struct {
	authService interfaces.AuthServiceI
}

func InitAuthHandler(authService interfaces.AuthServiceI) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// API Signup godoc
// @Tags Auth
// @Summary create new user
// @Accept json
// @Produce json
// @Param params body models.SignupRequest true "SignupRequest"
// @Router /api/v1/auth/sign-up [post]
// @Success	200 {object} entities.User
func (h *AuthHandler) Signup(c *gin.Context) {

	// Parse request body
	req := models.SignupRequest{}
	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	// Validate request body
	err := httpcommon.ValidatorParams(req)
	if err != nil {
		resErr := errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetArrayError(err)
		resErr.ExposeHttpError(c)
		return
	}

	// Call usecase
	res, resErr := h.authService.Signup(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	// Return response
	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// API VerifyEmail godoc
// @Tags Auth
// @Summary verify user email
// @Accept json
// @Produce json
// @Param params body models.VerifyEmailRequest true "VerifyEmailRequest"
// @Router /api/v1/auth/verify-email [post]
// @Success	200 {object} models.VerifyEmailResponse
func (h *AuthHandler) VerifyEmail(c *gin.Context) {

	req := models.VerifyEmailRequest{}
	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	if err := httpcommon.ValidatorParams(req); err != nil {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetArrayError(err).
			ExposeHttpError(c)
		return
	}

	// Call usecase
	res, resErr := h.authService.VerifyEmail(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	// Return response
	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// API ResendVerifyEmail godoc
// @Tags Auth
// @Summary resend verification email
// @Accept json
// @Produce json
// @Param params body models.ResendVerifyEmailRequest true "ResendVerifyEmailRequest"
// @Router /api/v1/auth/verify-email/resend [post]
// @Success	200 {object} models.ResendVerifyEmailResponse
func (h *AuthHandler) ResendVerifyEmail(c *gin.Context) {

	req := models.ResendVerifyEmailRequest{}
	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	// Validate request body
	if err := httpcommon.ValidatorParams(req); err != nil {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetArrayError(err).
			ExposeHttpError(c)
		return
	}

	// Call usecase
	res, resErr := h.authService.ResendVerifyEmail(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	// Return response
	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	req := models.LoginRequest{}
	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}
	if err := httpcommon.ValidatorParams(req); err != nil { /* build 400 similar to others */
		return
	}
	res, resErr := h.authService.Login(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}
	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// @Router /api/v1/auth/token/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {}

// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {}
