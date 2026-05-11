package command

import (
	"os"

	"github.com/vucongthanh92/go-base-utils/command/constants"
	migrationV2 "github.com/vucongthanh92/go-base-utils/command/migration_v2"
	"github.com/vucongthanh92/go-base-utils/command/start"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func UseCommands(commands ...*cobra.Command) {
	var rootCmd = &cobra.Command{}

	pflag.String(constants.ConfigFlagName, "", "--config=<config-path>")
	pflag.Int(constants.ForceFlagName, 0, "--force=<version>")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return
	}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func WithStartCommand(startFunc func(), cfg interface{}, dbConfigKeys ...string) *cobra.Command {
	return start.WithStartCommand(startFunc, cfg, dbConfigKeys...)
}

func WithMigrationCommand(dbConfigKeys ...string) *cobra.Command {
	return migrationV2.MigrationCommand(dbConfigKeys...)
}
