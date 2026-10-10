package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	baseUtils "github.com/vucongthanh92/go-base-utils/helper/utils"
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
func (s *jwtSigner) SignAccessToken(user entities.User, now time.Time, ttl time.Duration) (string, *utilsError.ErrorBuilder) {
	return s.SignAccessTokenForClient(user, now, ttl, "", "", "user")
}

func (s *jwtSigner) SignAccessTokenForClient(
	user entities.User,
	now time.Time,
	ttl time.Duration,
	clientID string,
	sessionID string,
	scope string,
) (string, *utilsError.ErrorBuilder) {
	jti, _ := baseUtils.NewSnowflakeID()
	if scope == "" {
		scope = "user"
	}

	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", user.ID),
		"email": user.Email,
		"scope": scope,
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
		"jti":   fmt.Sprintf("%d", jti),
		"iss":   s.issuer,
	}
	if clientID != "" {
		claims["aud"] = clientID
		claims["client_id"] = clientID
	}
	if sessionID != "" {
		claims["sid"] = sessionID
	}

	// Sign the token with the RSA private key
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if s.kid != "" {
		token.Header["kid"] = s.kid
	}

	// Sign the token and return the signed string
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", utilsError.InitErrorBuilder(nil).SetLogError(err).SetStatus(500)
	}

	return signed, nil
}

func (s *jwtSigner) SignIDToken(
	user entities.User,
	now time.Time,
	ttl time.Duration,
	audience string,
	nonce string,
	authTime time.Time,
	sessionID string,
) (string, *utilsError.ErrorBuilder) {
	jti, _ := baseUtils.NewSnowflakeID()
	claims := jwt.MapClaims{
		"iss":                s.issuer,
		"sub":                fmt.Sprintf("%d", user.ID),
		"aud":                audience,
		"email":              user.Email,
		"email_verified":     user.EmailVerified,
		"name":               user.DisplayName,
		"preferred_username": user.DisplayName,
		"picture":            user.AvatarURL,
		"auth_time":          authTime.Unix(),
		"iat":                now.Unix(),
		"exp":                now.Add(ttl).Unix(),
		"jti":                fmt.Sprintf("%d", jti),
	}
	if nonce != "" {
		claims["nonce"] = nonce
	}
	if sessionID != "" {
		claims["sid"] = sessionID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if s.kid != "" {
		token.Header["kid"] = s.kid
	}

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", utilsError.InitErrorBuilder(nil).SetLogError(err).SetStatus(500)
	}

	return signed, nil
}
