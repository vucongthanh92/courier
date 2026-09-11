package database

import (
	"context"
	"fmt"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormReadDb *gorm.DB
type GormWriteDb *gorm.DB

func Open(connectionString string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return db, nil
}

// GetConnectByGorm keeps the same read/write database contract as the other
// Courier services. Local development may point both connections at one DB.
func GetConnectByGorm(cfg *config.DatabaseConfig) (GormReadDb, GormWriteDb) {
	readDB, err := openAndConfigure(cfg.ReadDbCfg.ConnectionString, cfg.ReadDbCfg.MaxIdleConns, cfg.ReadDbCfg.MaxOpenConns, cfg.ReadDbCfg.ConnMaxLifetime)
	if err != nil {
		return nil, nil
	}
	writeDB, err := openAndConfigure(cfg.WriteDbCfg.ConnectionString, cfg.WriteDbCfg.MaxIdleConns, cfg.WriteDbCfg.MaxOpenConns, cfg.WriteDbCfg.ConnMaxLifetime)
	if err != nil {
		return nil, nil
	}
	return readDB, writeDB
}

func openAndConfigure(connectionString string, maxIdle, maxOpen, lifetimeMinutes int) (*gorm.DB, error) {
	db, err := Open(connectionString)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err = sqlDB.PingContext(context.Background()); err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(lifetimeMinutes) * time.Minute)
	return db, nil
}
