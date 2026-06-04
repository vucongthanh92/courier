package migrationv2

import (
	"github.com/spf13/cobra"
)

func initMigrationUpCmd(dbConfigKeys ...string) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "migrate up command",
		Long:  "migrate up command",
		Run: func(cmd *cobra.Command, args []string) {
			dbConfigs := GetDbConfigs(dbConfigKeys...)
			MigrateDatabase(dbConfigs, "up")
		},
	}
}
