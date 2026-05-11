package start

import (
	"fmt"
	"os"

	"github.com/vucongthanh92/go-base-utils/command/constants"
	migrationV2 "github.com/vucongthanh92/go-base-utils/command/migration_v2"
	"github.com/vucongthanh92/go-base-utils/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func WithStartCommand(startFunc func(), cfg interface{}, dbConfigKeys ...string) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start the server",
		Run: func(cmd *cobra.Command, args []string) {
			var configPath string
			// Priority config from env
			if environ := os.Getenv(config.AppEnv); environ != "" {
				configPath = fmt.Sprintf("./config/%s/config.yaml", environ)
			} else {
				configPath = viper.GetString(constants.ConfigFlagName)
			}
			config.LoadConfig(configPath, cfg)
			if len(dbConfigKeys) != 0 {
				dbConfigs := migrationV2.GetDbConfigs(dbConfigKeys...)
				migrationV2.MigrateDatabase(dbConfigs, "up")
			}

			startFunc()
		},
	}
}
