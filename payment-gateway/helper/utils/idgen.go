package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

const paymentCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	idEpoch   int64 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	idCounter uint64
)

type SnowflakeConfig struct {
	MachineID      uint16
	CustomEpoch    time.Time
	CheckMachineID func(uint16) (bool, error)
}

func InitSnowflake(cfg SnowflakeConfig) {
	if cfg.CustomEpoch.IsZero() {
		return
	}
	idEpoch = cfg.CustomEpoch.UnixMilli()
}

func NewSnowflakeID() (uint64, error) {
	now := time.Now().UnixMilli()
	if now < idEpoch {
		return 0, errors.New("invalid snowflake epoch")
	}
	seq := atomic.AddUint64(&idCounter, 1) & 0xFFF
	return uint64(now-idEpoch)<<12 | seq, nil
}

// NewSePayPaymentCode generates a Test mode-compatible payment code in the
// format COUR[A-Za-z0-9]{6,8}. The current format uses eight suffix characters.
func NewSePayPaymentCode(prefix string, suffixLength int) (string, error) {
	if prefix == "" || suffixLength < 6 || suffixLength > 8 {
		return "", errors.New("invalid SePay payment-code format")
	}

	bytes := make([]byte, suffixLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate SePay payment code: %w", err)
	}
	for index, value := range bytes {
		bytes[index] = paymentCodeAlphabet[int(value)%len(paymentCodeAlphabet)]
	}

	return prefix + string(bytes), nil
}
