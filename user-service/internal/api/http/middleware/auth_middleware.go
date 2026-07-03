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
	cacheRepo "github.com/vucongthanh92/courier/user-service/internal/repository/external/redis"
)

// JWTMiddleware verifies JWT (RS256), checks denylist, and injects claims as "authClaims".
// keyResolver loads public key by kid using cache-first strategy.
func JWTMiddleware(deny interfaces.TokenDenylistI, keyResolver func(context.Context, string) (any, *errHandler.ErrorBuilder)) gin.HandlerFunc {
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

// LoadPubKeys fetches active JWK and warms cache for the active kid.
func LoadPubKeys(ctx context.Context, jwkRepo interfaces.JWKQueryRepoI, cache cacheRepo.JWKCacheRepo) (map[string]any, *errHandler.ErrorBuilder) {
	jwk, err := jwkRepo.GetActiveKey(ctx)
	if err != nil {
		return nil, err
	}
	pub, errBuilder := ResolvePublicKey(ctx, jwkRepo, cache, jwk.Kid)
	if errBuilder != nil {
		return nil, errBuilder
	}
	return map[string]any{jwk.Kid: pub}, nil
}

func ResolvePublicKey(ctx context.Context, jwkRepo interfaces.JWKQueryRepoI, cache cacheRepo.JWKCacheRepo, kid string) (any, *errHandler.ErrorBuilder) {
	if kid == "" {
		active, err := jwkRepo.GetActiveKey(ctx)
		if err != nil {
			return nil, err
		}
		kid = active.Kid
	}

	if cache != nil {
		if cached, err := cache.GetByKid(ctx, kid); err == nil && cached != nil && cached.PublicPEM != "" {
			pub, errJWT := jwt.ParseRSAPublicKeyFromPEM([]byte(cached.PublicPEM))
			if errJWT == nil {
				return pub, nil
			}
		}
	}

	jwk, err := jwkRepo.GetKeyByKid(ctx, kid)
	if err != nil {
		return nil, err
	}

	pub, errJWT := jwt.ParseRSAPublicKeyFromPEM([]byte(jwk.PublicPEM))
	if errJWT != nil {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusInternalServerError).
			SetError(models.ErrorDTO{Code: "invalid_public_key", Message: "Invalid public key"})
	}
	if cache != nil {
		_ = cache.SetByKid(ctx, models.JWKCacheEntry{Kid: jwk.Kid, PublicPEM: jwk.PublicPEM, Alg: jwk.Alg}, 15*time.Minute)
	}
	return pub, nil
}
