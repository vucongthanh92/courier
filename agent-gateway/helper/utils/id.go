package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func NewCorrelationID() string {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(random)
}

func StableUUID(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	sum[6] = (sum[6] & 0x0f) | 0x50
	sum[8] = (sum[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		sum[0:4],
		sum[4:6],
		sum[6:8],
		sum[8:10],
		sum[10:16],
	)
}
