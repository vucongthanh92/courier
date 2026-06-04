package migrationv2

import (
	"github.com/spf13/cobra"
)

func initMigrationDownCmd(dbConfigKeys ...string) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "migrate down command",
		Long:  "migrate down command",
		Run: func(cmd *cobra.Command, args []string) {
			dbConfigs := GetDbConfigs(dbConfigKeys...)
			MigrateDatabase(dbConfigs, "down")
		},
	}
}
