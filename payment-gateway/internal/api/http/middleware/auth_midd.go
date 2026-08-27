package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/payment-gateway/helper/constants"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
	cacheRepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/redis"
	user_grpc "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/user_grpc"
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

		claims, errBuilder := VerifyToken(c.Request.Context(), tokenStr, deny, keyResolver)
		if errBuilder != nil {
			errBuilder.ExposeHttpError(c)
			c.Abort()
			return
		}

		c.Set("authClaims", claims)
		c.Next()
	}
}

func VerifyToken(ctx context.Context, tokenStr string, deny interfaces.TokenDenylistI, keyResolver func(context.Context, string) (interface{}, *errHandler.ErrorBuilder)) (jwt.MapClaims, *errHandler.ErrorBuilder) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		kid, _ := t.Header["kid"].(string)
		key, resErr := keyResolver(ctx, kid)
		if resErr != nil || key == nil {
			return nil, jwt.ErrSignatureInvalid
		}
		return key, nil
	})
	if err != nil || !token.Valid {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{Code: "invalid_token", Field: "authorization token", Message: "Token is invalid"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{Code: "token_expired", Field: "authorization token", Message: "Token expired"})
	}

	if jti, _ := claims["jti"].(string); jti != "" {
		blocked, derr := deny.IsBlocked(ctx, jti)
		if derr != nil {
			return nil, errHandler.InitErrorBuilder(ctx).
				SetStatus(http.StatusUnauthorized).
				SetError(models.ErrorDTO{Code: "system_error", Field: "authorization token", Message: "Denylist check failed"})
		}
		if blocked {
			return nil, errHandler.InitErrorBuilder(ctx).
				SetStatus(http.StatusUnauthorized).
				SetError(models.ErrorDTO{Code: "token_revoked", Field: "authorization token", Message: "Token revoked"})
		}
	}

	return claims, nil
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
		return nil, errHandler.InitErrorBuilder(context.Background()).
			SetStatus(http.StatusInternalServerError).
			SetError(models.ErrorDTO{Code: "invalid_public_key", Message: "Invalid public key"})
	}
	return pub, nil
}

// ResolvePublicKey fetches the public key by kid, first checking the cache, then falling back to the JWK client.
func ResolvePublicKey(ctx context.Context, cache cacheRepo.JWKCacheRepo, client user_grpc.UserGrpcClient, kid string) (any, *errHandler.ErrorBuilder) {
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
		_ = cache.SetByKid(ctx, cacheRepo.JWKCacheEntry{Kid: respKid, PublicPEM: respPublicPEM, Alg: respAlg}, constants.Time_Cache_15_minutes)
	}
	return pub, nil
}
