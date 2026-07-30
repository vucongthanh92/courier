package v1

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/api/http/middleware"
	"github.com/vucongthanh92/courier/chat-service/internal/api/ws"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type RealtimeHandler struct {
	hub         *ws.Hub
	upgrader    websocket.Upgrader
	tokenDeny   interfaces.TokenDenylistI
	keyResolver func(context.Context, string) (interface{}, *errHandler.ErrorBuilder)
}

func InitRealtimeHandler(
	hub *ws.Hub,
	allowOrigins []string,
	tokenDeny interfaces.TokenDenylistI,
	keyResolver func(context.Context, string) (interface{}, *errHandler.ErrorBuilder),
) *RealtimeHandler {
	return &RealtimeHandler{
		hub:         hub,
		upgrader:    ws.Upgrader(allowOrigins),
		tokenDeny:   tokenDeny,
		keyResolver: keyResolver,
	}
}

func (h *RealtimeHandler) Handle(c *gin.Context) {
	token := c.Query("access_token")
	if token == "" {
		token = c.Query("token")
	}
	if token == "" {
		errHandler.InitErrorBuilder(c).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{Code: "token_missing", Message: "access_token query parameter missing"}).
			ExposeHttpError(c)
		return
	}

	claims, authErr := middleware.VerifyToken(c.Request.Context(), token, h.tokenDeny, h.keyResolver)
	if authErr != nil {
		authErr.ExposeHttpError(c)
		return
	}
	userID := utils.ParseUserID(claims["sub"])
	if userID == 0 {
		errHandler.InitErrorBuilder(c).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{Code: "unauthorized", Message: "missing authenticated user"}).
			ExposeHttpError(c)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.Register(userID, conn)
}
