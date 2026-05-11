package utils

import (
	"errors"
	"sync/atomic"
	"time"
)

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
