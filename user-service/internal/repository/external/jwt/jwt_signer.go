package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
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

func InitJWTSigner(jwk entities.JWKKey, log logger.Logger) (interfaces.JWTSignerI, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(jwk.PrivatePEM))
	if err != nil {
		return nil, fmt.Errorf("parse rsa key: %w", err)
	}
	return &jwtSigner{
		privateKey: key,
		kid:        jwk.Kid,
		issuer:     "user-service",
		log:        log,
	}, nil
}

// SignAccessToken implements interfaces.JWTSignerI
func (s *jwtSigner) SignAccessToken(user entities.User, now time.Time, ttl time.Duration) (string, *errHandler.ErrorBuilder) {
	jti := utils.RandString(16)
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", user.ID),
		"email": user.Email,
		"scope": "user",
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
		"jti":   jti,
		"iss":   s.issuer,
	}

	// Sign the token with the RSA private key
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if s.kid != "" {
		token.Header["kid"] = s.kid
	}

	// Sign the token and return the signed string
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", errHandler.InitErrorBuilder(nil).SetLogError(err).SetStatus(500)
	}

	return signed, nil
}
