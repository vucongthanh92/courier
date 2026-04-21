package database

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/chat-service/config"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/go-sql-driver/mysql"
	baseDatabase "github.com/vucongthanh92/go-base-utils/database"
	"github.com/vucongthanh92/go-base-utils/logger"
)

type GormReadDb *gorm.DB
type GormWriteDb *gorm.DB

func GetConnectByGorm(cfg *config.DatabaseConfig) (GormReadDb, GormWriteDb) {
	readConn, err := baseDatabase.GormConnectDB(cfg.ReadDbCfg.DbType, cfg.ReadDbCfg.ConnectionString)
	if err != nil {
		logger.Error("connect error", zap.Error(err))
		return nil, nil
	}

	readDb, err := readConn.DB()
	if err != nil {
		logger.Error("connect error", zap.Error(err))
		return nil, nil
	}

	if err = readDb.PingContext(context.Background()); err != nil {
		logger.Error("read database", zap.Error(err))
	}

	readDb.SetMaxIdleConns(cfg.ReadDbCfg.MaxIdleConns)
	readDb.SetMaxOpenConns(cfg.ReadDbCfg.MaxOpenConns)
	readDb.SetConnMaxLifetime(time.Duration(cfg.ReadDbCfg.ConnMaxLifetime) * time.Minute)

	writeConn, err := baseDatabase.GormConnectDB(cfg.WriteDbCfg.DbType, cfg.WriteDbCfg.ConnectionString)
	if err != nil {
		logger.Error("connect error", zap.Error(err))
		return nil, nil
	}

	writeDb, err := writeConn.DB()
	if err != nil {
		logger.Error("connect error", zap.Error(err))
		return nil, nil
	}

	if err = writeDb.PingContext(context.Background()); err != nil {
		logger.Error("write database", zap.Error(err))
	}

	writeDb.SetMaxIdleConns(cfg.WriteDbCfg.MaxIdleConns)
	writeDb.SetMaxOpenConns(cfg.WriteDbCfg.MaxOpenConns)
	writeDb.SetConnMaxLifetime(time.Duration(cfg.WriteDbCfg.ConnMaxLifetime) * time.Minute)

	logger.Info("Connected to read & write database!")
	return readConn, writeConn
}
