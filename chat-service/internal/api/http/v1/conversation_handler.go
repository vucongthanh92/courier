package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	httpcommon "github.com/vucongthanh92/courier/chat-service/helper/http_common"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type ConversationHandler struct {
	conversationService interfaces.ConversationServiceI
}

func InitConversationHandler(
	conversationService interfaces.ConversationServiceI,
) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
	}
}

// API Create Conversation godoc
// @Tags Conversation
// @Summary create new conversation
// @Accept json
// @Produce json
// @Param params body models.CreateConversationRequest true "CreateConversationRequest"
// @Router /api/v1/conversation/create [post]
// @Success	200 {object} models.CreateConversationResponse
func (h *ConversationHandler) CreateConversation(c *gin.Context) {

	var req models.CreateConversationRequest
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

	ctx := utils.SetHeaderByKey(c, "headers")

	// Extract the authenticated user's ID from the context and set it as the creator ID in the request
	if v, ok := c.Get("authClaims"); ok {
		if claims, isExist := v.(jwt.MapClaims); isExist {
			req.CreatorID = utils.ParseUserID(claims["sub"])
		} else {
			resErr := errHandler.InitErrorBuilder(c).
				SetLogError(errors.New(constants.InvalidValue)).
				SetStatus(http.StatusBadRequest).
				SetError(models.ErrorDTO{
					Code:    "invalid_auth_claims",
					Message: "Invalid auth claims in context",
				})
			resErr.ExposeHttpError(c)
			return
		}
	}

	// Set the creator ID in the request
	resp, errBuilder := h.conversationService.CreateConversation(ctx, &req)
	if errBuilder != nil {
		errBuilder.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(resp))
}

// API List Conversations godoc
// @Tags Conversation
// @Summary list conversations for authenticated user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit, default 20 and max 100"
// @Param before_last_message_at query string false "Cursor activity timestamp"
// @Param before_conversation_id query int false "Cursor conversation ID"
// @Router /api/v1/conversations [get]
// @Success 200 {object} httpcommon.SuccessResponse[models.ListConversationsResponse]
// @Failure 400 {object} httpcommon.SuccessResponse[any]
// @Failure 401 {object} httpcommon.SuccessResponse[any]
func (h *ConversationHandler) ListConversations(c *gin.Context) {

	var req models.ListConversationsRequest
	if err := httpcommon.GetQueryParamsHTTP(c, &req); err != nil {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    constants.INVALID_FORMAT,
				Message: "Invalid Request",
			}).ExposeHttpError(c)
		return
	}

	//
	claimsValue, ok := c.Get("authClaims")
	claims, claimsOK := claimsValue.(jwt.MapClaims)
	if !ok || !claimsOK {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{
				Code:    "unauthorized",
				Message: "missing authenticated user",
			}).ExposeHttpError(c)
		return
	}

	//
	req.RequesterID = utils.ParseUserID(claims["sub"])
	ctx := utils.SetHeaderByKey(c, "headers")
	resp, errBuilder := h.conversationService.ListConversations(ctx, &req)
	if errBuilder != nil {
		errBuilder.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(*resp))
}
