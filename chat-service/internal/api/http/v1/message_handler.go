package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	httpcommon "github.com/vucongthanh92/courier/chat-service/helper/http_common"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type MessageHandler struct {
	messageService interfaces.MessageServiceI
}

func InitMessageHandler(messageService interfaces.MessageServiceI) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// CreateMessage godoc
// @Tags Message
// @Summary create a text message in a conversation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Param params body models.SendMessageRequest true "SendMessageRequest"
// @Router /api/v1/conversations/{id}/messages/create [post]
// @Success 200 {object} httpcommon.SuccessResponse[models.MessageResponse]
// @Success 201 {object} httpcommon.SuccessResponse[models.MessageResponse]
// @Failure 400 {object} httpcommon.SuccessResponse[any]
// @Failure 401 {object} httpcommon.SuccessResponse[any]
// @Failure 403 {object} httpcommon.SuccessResponse[any]
// @Failure 404 {object} httpcommon.SuccessResponse[any]
// @Failure 409 {object} httpcommon.SuccessResponse[any]
// @Failure 429 {object} httpcommon.SuccessResponse[any]
func (h *MessageHandler) CreateMessage(c *gin.Context) {

	// Parse and validate the conversation ID from the URL parameter
	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || conversationID == 0 {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    "invalid_conversation_id",
				Message: "conversation id must be a positive integer",
			}).ExposeHttpError(c)
		return
	}

	// Parse and validate the request body
	var req models.SendMessageRequest
	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    constants.INVALID_FORMAT,
				Message: "Invalid Request",
			}).ExposeHttpError(c)
		return
	}

	// Validate the request body using the ValidatorParams function
	// If there are validation errors, return a 400 Bad Request response with the validation errors
	if validationErrors := httpcommon.ValidatorParams(req); validationErrors != nil {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetArrayError(validationErrors).
			ExposeHttpError(c)
		return
	}

	// Set the authenticated user's ID as the sender ID in the request
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

	// Set the authenticated user's ID as the sender ID in the request
	req.ConversationID = conversationID
	req.SenderID = utils.ParseUserID(claims["sub"])
	ctx := utils.SetHeaderByKey(c, "headers")
	response, created, serviceErr := h.messageService.CreateMessage(ctx, &req)
	if serviceErr != nil {
		serviceErr.ExposeHttpError(c)
		return
	}

	// Determine the appropriate HTTP status code based on whether the message was created or not
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}

	c.JSON(status, httpcommon.NewSuccessResponse(*response))
}
