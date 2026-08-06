package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func NewCorrelationID() string {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(random)
}
