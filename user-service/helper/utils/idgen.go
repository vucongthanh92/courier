package utils

import (
	"errors"
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

var (
	sf   *sonyflake.Sonyflake
	once sync.Once
)

type SnowflakeConfig struct {
	MachineID      uint16                     // 0..1023 sonyflake
	CustomEpoch    time.Time                  // ex: time.Date(2020,1,1,0,0,0,0,time.UTC)
	CheckMachineID func(uint16) (bool, error) // optional
}

func InitSnowflake(cfg SnowflakeConfig) {
	once.Do(func() {
		st := sonyflake.Settings{
			StartTime: cfg.CustomEpoch,
			MachineID: func() (uint16, error) {
				if cfg.CheckMachineID != nil {
					_, err := cfg.CheckMachineID(cfg.MachineID)
					if err != nil {
						return 0, err
					}
				}
				return cfg.MachineID, nil
			},
		}

		sf = sonyflake.NewSonyflake(st)
		if sf == nil {
			panic("sonyflake not created")
		}
	})
}

func NewSnowflakeID() (uint64, error) {
	if sf == nil {
		return 0, errors.New("idgen not initialized")
	}

	id, err := sf.NextID()
	if err != nil {
		return 0, err
	}

	return id, nil
}
