package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/user-service/config"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/go-base-utils/logger"
)

type jwtSigner struct {
	privateKey *rsa.PrivateKey
	kid        string
	issuer     string
	log        logger.Logger
}

func InitJWTSigner(cfg *config.JWTConfig, log logger.Logger) interfaces.JWTSignerI {
	key, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.PrivateKey))
	return &jwtSigner{
		privateKey: key,
		kid:        cfg.Kid,
		issuer:     cfg.Issuer,
		log:        log,
	}
}

// SignAccessToken implements interfaces.JWTSignerI
func (s *jwtSigner) SignAccessToken(user entities.User, now time.Time, ttl time.Duration) (string, *errHandler.ErrorBuilder) {
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", user.ID),
		"email": user.Email,
		"scope": "user",
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
		"iss":   s.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if s.kid != "" {
		token.Header["kid"] = s.kid
	}

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", errHandler.InitErrorBuilder(nil).SetLogError(err).SetStatus(500)
	}

	return signed, nil
}
