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
