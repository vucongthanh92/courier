package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strings"
)

// GetHeaderFromKey retrieves the header value from the context for the specified key.
func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func VerifyPKCE(challenge string, verifier string) bool {
	sum := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(sum[:])
	return expected == challenge
}

// EncodeAuthorizeRequest encodes the SSO authorize request parameters into a URL-encoded query string.
func EncodeAuthorizeRequest(clientID, redirectURI, responseType, scope, state, nonce, codeChallenge, codeChallengeMethod string) string {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("response_type", responseType)
	values.Set("scope", scope)
	values.Set("state", state)
	values.Set("nonce", nonce)
	values.Set("code_challenge", codeChallenge)
	values.Set("code_challenge_method", codeChallengeMethod)
	return values.Encode()
}

// OptionalString returns a pointer to the string if it's not empty, otherwise returns nil.
func OptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// NormalizeScope trims whitespace from the scope string and returns a default scope if it's empty.
func RemoveWhiteSpace(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return "empty string"
	}
	return str
}

func CodeTTLSeconds(codeTTLSeconds int) int {
	if codeTTLSeconds <= 0 {
		return 120
	}
	return codeTTLSeconds
}

func CheckIdTokenTTLMinutes(min int) int {
	if min <= 0 {
		return 30
	}
	return min
}

func CheckSessionTTLMinutes(min int) int {
	if min <= 0 {
		return 30 * 24 * 60
	}
	return min
}

// CheckRefreshTokenTTLMinutes checks the refresh token TTL in minutes and returns a default value if it's less than or equal to 0.
// If the provided value is greater than 0, it returns the provided value.
func LoginRedirect(loginURL, clientID, redirectURI, responseType, scope, state, nonce, codeChallenge, codeChallengeMethod string) string {
	parsedURL, _ := url.Parse(loginURL)
	query := parsedURL.Query()
	query.Set("sso_authorize", EncodeAuthorizeRequest(
		clientID,
		redirectURI,
		responseType,
		scope,
		state,
		nonce,
		codeChallenge,
		codeChallengeMethod,
	))
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}
