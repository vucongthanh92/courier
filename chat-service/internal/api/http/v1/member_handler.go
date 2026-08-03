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

type MemberHandler struct {
	memberService interfaces.MemberServiceI
}

func InitMemberHandler(memberService interfaces.MemberServiceI) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}

// API List Conversation Members godoc
// @Tags Conversation
// @Summary list members for a conversation with resolved user profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Router /api/v1/conversation/{id}/members [get]
// @Success 200 {object} httpcommon.SuccessResponse[models.ListConversationMembersResponse]
// @Failure 400 {object} httpcommon.SuccessResponse[any]
// @Failure 401 {object} httpcommon.SuccessResponse[any]
// @Failure 403 {object} httpcommon.SuccessResponse[any]
func (h *MemberHandler) ListConversationMembers(c *gin.Context) {
	var req models.ListConversationMembersRequest
	if err := c.ShouldBindUri(&req); err != nil {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    constants.INVALID_FORMAT,
				Message: "Invalid Request",
			}).ExposeHttpError(c)
		return
	}

	// Retrieve the authenticated user's claims from the context
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

	// Set the requester ID in the request object based on the authenticated user's claims
	req.RequesterID = utils.ParseUserID(claims["sub"])
	ctx := utils.SetHeaderByKey(c, "headers")
	resp, errBuilder := h.memberService.ListConversationMembers(ctx, &req)
	if errBuilder != nil {
		errBuilder.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(*resp))
}
