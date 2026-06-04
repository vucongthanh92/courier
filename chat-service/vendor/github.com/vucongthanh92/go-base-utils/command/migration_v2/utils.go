package migrationv2

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/vucongthanh92/go-base-utils/logger"

	"github.com/vucongthanh92/go-base-utils/command/constants"
	"github.com/vucongthanh92/go-base-utils/config"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pressly/goose/v3"
)

func MigrateDatabase(dbConfigs []databaseConfig, cmd string) {
	for _, cfg := range dbConfigs {
		if cfg.DisableMigration {
			logger.Info("Skip Migrate database on local", zap.String("db", cfg.Database))
			return
		}
		execute(cfg, cmd)
	}
}

func execute(cfg databaseConfig, cmd string) {
	fmt.Println("Running migrate command")

	driver, err := getDbDriver(cfg)
	if err != nil {
		logger.Error("Get db driver", zap.Error(err))
		return
	}

	if err := goose.Run(cmd, driver, cfg.MigrationFilePath); err != nil {
		logger.Error("goose run cmd error", zap.Error(err))
	}
}

func getDbDriver(cfg databaseConfig) (*sql.DB, error) {
	db, err := goose.OpenDBWithDriver(cfg.DbType, cfg.ConnectionString)
	if err != nil {
		logger.Error("goose: failed to open DB", zap.Error(err))
	}

	return db, err
}

func GetDbConfigs(dbConfigKeys ...string) []databaseConfig {
	var (
		cfg        interface{}
		configPath = viper.GetString(constants.ConfigFlagName)
		dbConfigs  = make([]databaseConfig, 0)
	)

	// Priority config from env
	if env := os.Getenv(config.AppEnv); env != "" {
		configPath = fmt.Sprintf("./config/%s/config.yaml", env)
	}

	config.LoadConfig(configPath, &cfg)
	data := cfg.(map[string]interface{})

	if len(dbConfigKeys) == 0 {
		return dbConfigs
	}

	for _, key := range dbConfigKeys {
		keys := strings.Split(key, ".")
		var cfgData map[string]interface{}

		for i, k := range keys {
			if i == 0 {
				cfgData = data[strings.ToLower(k)].(map[string]interface{})
			} else {
				cfgData = cfgData[strings.ToLower(k)].(map[string]interface{})
			}
		}

		dbName, _ := cfgData["database"].(string)

		dbConfig := databaseConfig{
			DbType:            cfgData["dbtype"].(string),
			ConnectionString:  cfgData["connectionstring"].(string),
			MigrationFilePath: cfgData["migrationfilepath"].(string),
			Database:          dbName,
			DisableMigration:  cfgData["disablemigration"].(bool),
		}

		dbConfigs = append(dbConfigs, dbConfig)
	}

	return dbConfigs
}
