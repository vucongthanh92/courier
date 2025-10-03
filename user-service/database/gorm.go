package database

import (
	"context"
	"time"

	"github.com/vucongthanh92/go-base-project/config"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/go-sql-driver/mysql"
	"github.com/vucongthanh92/go-base-utils/database"
	"github.com/vucongthanh92/go-base-utils/logger"
)

type GormReadDb *gorm.DB
type GormWriteDb *gorm.DB

func GetConnectByGorm(cfg *config.DatabaseConfig) (GormReadDb, GormWriteDb) {

	// readDB
	readConn, err := database.GormConnectDB(cfg.ReadDbCfg.DbType, cfg.ReadDbCfg.ConnectionString)
	if err != nil {
		logger.Error("connect error", zap.Error(err))
		return nil, nil
	}

	readDb, err := readConn.DB()
	if err != nil {
		logger.Error("connect error", zap.Error(err))
		return nil, nil
	}

	err = readDb.PingContext(context.Background())
	if err != nil {
		logger.Error("Read database", zap.Error(err))
	}

	readDb.SetMaxIdleConns(cfg.ReadDbCfg.MaxIdleConns)
	readDb.SetMaxOpenConns(cfg.ReadDbCfg.MaxOpenConns)
	readDb.SetConnMaxLifetime(time.Duration(cfg.ReadDbCfg.ConnMaxLifetime) * time.Minute)

	// writeDB
	writeConn, err := database.GormConnectDB(cfg.WriteDbCfg.DbType, cfg.WriteDbCfg.ConnectionString)
	if err != nil {
		logger.Error("connect error", zap.Error(err))
		return nil, nil
	}

	writeDb, err := writeConn.DB()
	if err != nil {
		logger.Error("connect error", zap.Error(err))
		return nil, nil
	}

	err = writeDb.PingContext(context.Background())
	if err != nil {
		logger.Error("Write database", zap.Error(err))
	}

	writeDb.SetMaxIdleConns(cfg.WriteDbCfg.MaxIdleConns)
	writeDb.SetMaxOpenConns(cfg.WriteDbCfg.MaxOpenConns)
	writeDb.SetConnMaxLifetime(time.Duration(cfg.WriteDbCfg.ConnMaxLifetime) * time.Minute)

	// return
	logger.Info("Connected to read & write database!")
	return readConn, writeConn
}
