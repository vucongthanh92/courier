package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	cacheRepo "github.com/vucongthanh92/courier/user-service/internal/repository/external/redis"

	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	utilsModels "github.com/vucongthanh92/go-base-utils/models"
)

// LoadPubKeys fetches active JWK and warms cache for the active kid.
func LoadPubKeys(ctx context.Context, jwkRepo interfaces.JWKQueryRepoI, cache cacheRepo.JWKCacheRepo) (map[string]any, *utilsError.ErrorBuilder) {
	jwk, err := jwkRepo.GetActiveKey(ctx)
	if err != nil {
		return nil, utilsError.InitErrorBuilder(ctx).SetStatus(http.StatusInternalServerError).
			SetError(utilsModels.ErrorDTO{
				Code:    "system_error",
				Message: "Failed to get active key",
			})
	}

	pub, errBuilder := ResolvePublicKey(ctx, jwkRepo, cache, jwk.Kid)
	if errBuilder != nil {
		return nil, errBuilder
	}

	return map[string]any{jwk.Kid: pub}, nil
}

// ResolvePublicKey retrieves the public key for a given kid, using cache if available.
// If kid is empty, it fetches the active key. Returns an error if the key cannot be resolved.
func ResolvePublicKey(ctx context.Context, jwkRepo interfaces.JWKQueryRepoI, cache cacheRepo.JWKCacheRepo, kid string) (any, *utilsError.ErrorBuilder) {
	if kid == "" {
		active, err := jwkRepo.GetActiveKey(ctx)
		if err != nil {
			return nil, utilsError.InitErrorBuilder(ctx).SetStatus(http.StatusInternalServerError).
				SetError(utilsModels.ErrorDTO{
					Code:    "system_error",
					Message: "Failed to get active key",
				})
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
		return nil, utilsError.InitErrorBuilder(ctx).SetStatus(http.StatusInternalServerError).
			SetError(utilsModels.ErrorDTO{
				Code:    "system_error",
				Message: "Failed to get key by kid",
			})
	}

	pub, errJWT := jwt.ParseRSAPublicKeyFromPEM([]byte(jwk.PublicPEM))
	if errJWT != nil {
		return nil, utilsError.InitErrorBuilder(ctx).
			SetStatus(http.StatusInternalServerError).
			SetError(utilsModels.ErrorDTO{
				Code:    "invalid_public_key",
				Message: "Invalid public key",
			})
	}

	if cache != nil {
		_ = cache.SetByKid(ctx, models.JWKCacheEntry{Kid: jwk.Kid, PublicPEM: jwk.PublicPEM, Alg: jwk.Alg}, 15*time.Minute)
	}

	return pub, nil
}
