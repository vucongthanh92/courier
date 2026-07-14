package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	jwkclient "github.com/vucongthanh92/courier/chat-service/internal/repository/external/jwkclient"
	cacheRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/external/redis"
)

// JWTMiddleware verifies JWT (RS256), checks denylist, and injects claims as "authClaims".
// keyResolver fetches public key by kid using cache-first strategy.
func JWTMiddleware(deny interfaces.TokenDenylistI, keyResolver func(context.Context, string) (interface{}, *errHandler.ErrorBuilder)) gin.HandlerFunc {
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
			key, resErr := keyResolver(c.Request.Context(), kid)
			if resErr != nil || key == nil {
				return nil, jwt.ErrSignatureInvalid
			}
			return key, nil
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

// extractBearer extracts the token from "Authorization: Bearer
func extractBearer(h string) string {
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// unauthorized responds with 401 Unauthorized and a JSON error message.
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

// decodePublicKeyPEM decodes a PEM-encoded RSA public key.
func decodePublicKeyPEM(publicPEM string) (interface{}, *errHandler.ErrorBuilder) {
	pub, errJWT := jwt.ParseRSAPublicKeyFromPEM([]byte(publicPEM))
	if errJWT != nil {
		return nil, errHandler.InitErrorBuilder(nil).
			SetStatus(http.StatusInternalServerError).
			SetError(models.ErrorDTO{Code: "invalid_public_key", Message: "Invalid public key"})
	}
	return pub, nil
}

// ResolvePublicKey fetches the public key by kid, first checking the cache, then falling back to the JWK client.
func ResolvePublicKey(ctx context.Context, cache cacheRepo.JWKCacheRepo, client jwkclient.Client, kid string) (interface{}, *errHandler.ErrorBuilder) {
	if kid != "" && cache != nil {
		if cached, err := cache.GetByKid(ctx, kid); err == nil && cached != nil && cached.PublicPEM != "" {
			return decodePublicKeyPEM(cached.PublicPEM)
		}
	}

	respKid, respPublicPEM, respAlg, errBuilder := client.GetPublicKey(ctx, kid)
	if errBuilder != nil {
		return nil, errBuilder
	}
	pub, errBuilder := decodePublicKeyPEM(respPublicPEM)
	if errBuilder != nil {
		return nil, errBuilder
	}
	if cache != nil && respKid != "" {
		_ = cache.SetByKid(ctx, cacheRepo.JWKCacheEntry{Kid: respKid, PublicPEM: respPublicPEM, Alg: respAlg}, 15*time.Minute)
	}
	return pub, nil
}
