package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

// JWTMiddleware verifies JWT (RS256), checks denylist, and injects claims as "authClaims".
// pubKeys is a map kid -> public key; if only one key, kid may be empty.
func JWTMiddleware(deny interfaces.TokenDenylistI, pubKeys map[string]interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearer(c.GetHeader("Authorization"))
		if tokenStr == "" {
			unauthorized(c, "token_missing", "Authorization header missing")
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			kid, _ := t.Header["kid"].(string)
			if k, ok := pubKeys[kid]; ok {
				return k, nil
			}
			// fallback: only one key without kid
			if len(pubKeys) == 1 {
				for _, k := range pubKeys {
					return k, nil
				}
			}
			return nil, jwt.ErrSignatureInvalid
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
	resErr := errHandler.InitErrorBuilder(c).
		SetStatus(http.StatusUnauthorized).
		SetError(models.ErrorDTO{
			Code:    code,
			Field:   "authorization token",
			Message: msg,
		})
	resErr.ExposeHttpError(c)
	c.Abort()
}

// LoadPubKeys fetches active JWK and returns map[kid]publicKey for middleware use.
func LoadPubKeys(ctx context.Context, jwkRepo interfaces.JWKQueryRepoI) (map[string]interface{}, *errHandler.ErrorBuilder) {
	jwk, err := jwkRepo.GetActiveKey(ctx)
	if err != nil {
		return nil, err
	}

	pub, errJWT := jwt.ParseRSAPublicKeyFromPEM([]byte(jwk.PublicPEM))
	if errJWT != nil {
		return nil, errHandler.InitErrorBuilder(nil).
			SetStatus(http.StatusInternalServerError).
			SetError(models.ErrorDTO{Code: "invalid_public_key", Message: "Invalid public key"})
	}

	m := make(map[string]interface{})

	if jwk.Kid != "" {
		m[jwk.Kid] = pub
	} else {
		m[""] = pub
	}

	return m, nil
}
