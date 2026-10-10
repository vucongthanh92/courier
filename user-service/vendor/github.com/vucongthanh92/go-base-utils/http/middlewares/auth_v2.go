package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	httpError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	"github.com/vucongthanh92/go-base-utils/models"
)

type TokenDenylistI interface {
	Block(ctx context.Context, jti string, ttl time.Duration) error
	IsBlocked(ctx context.Context, jti string) (bool, error)
}

// JWTMiddleware verifies JWT (RS256), checks denylist, and injects claims as "authClaims".
// keyResolver loads public key by kid using cache-first strategy.
func JWTMiddleware(deny TokenDenylistI, keyResolver func(context.Context, string) (any, *httpError.ErrorBuilder)) gin.HandlerFunc {

	return func(c *gin.Context) {
		tokenStr := extractBearer(c.GetHeader("Authorization"))
		if tokenStr == "" {
			unauthorized(c, "token_missing", "Authorization header missing")
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			kid, _ := t.Header["kid"].(string)
			k, resErr := keyResolver(c.Request.Context(), kid)
			if resErr != nil || k == nil {
				return nil, jwt.ErrSignatureInvalid
			}
			return k, nil
		})
		if err != nil || !token.Valid {
			unauthorized(c, "invalid_token", "Token is invalid")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !claims.VerifyExpiresAt(time.Now().Unix(), true) {
			unauthorized(c, "token_expired", "Token expired")
			return
		}

		// denylist check
		if jti, _ := claims["jti"].(string); jti != "" {
			blocked, derr := deny.IsBlocked(c, jti)
			if derr != nil {
				unauthorized(c, "system_error", "Denylist check failed")
				return
			}
			if blocked {
				unauthorized(c, "token_revoked", "Token revoked")
				return
			}
		}

		c.Set("authClaims", claims)
		c.Next()
	}
}

func extractBearer(h string) string {
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func unauthorized(c *gin.Context, code, msg string) {
	resErr := httpError.InitErrorBuilder(c).
		SetStatus(http.StatusUnauthorized).
		SetError(models.ErrorDTO{
			Code:    code,
			Field:   "authorization token",
			Message: msg,
		})
	resErr.ExposeHttpError(c)
	c.Abort()
}
