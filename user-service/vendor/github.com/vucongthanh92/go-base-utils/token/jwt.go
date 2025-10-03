package token

import (
	"github.com/golang-jwt/jwt"
)

type S14eClaims jwt.MapClaims

var (
	S14EAudience = "s14e.aud"
)

func NewClaims(kid string) S14eClaims {
	return map[string]interface{}{
		"aud": S14EAudience,
		"kid": kid,
	}
}

// ParseToken without verify signed and parse jwtToken to Claims
func ParseTokenUnverify(tokenString string) (jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}
