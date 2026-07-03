package models

type JWKCacheEntry struct {
	Kid       string `json:"kid"`
	PublicPEM string `json:"public_pem"`
	Alg       string `json:"alg,omitempty"`
}
