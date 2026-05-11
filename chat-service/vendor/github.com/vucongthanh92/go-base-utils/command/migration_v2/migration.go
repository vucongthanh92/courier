package migrationv2

import (
	"github.com/spf13/cobra"
)

type databaseConfig struct {
	DbType            string
	ConnectionString  string
	MigrationFilePath string
	Database          string
	DisableMigration  bool
}

var migrateCmd = &cobra.Command{}
var step int

func MigrationCommand(dbConfigKeys ...string) *cobra.Command {
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "migrate cmd is used for database migration",
		Long:  `migrate cmd is used for database migration: migrate < up | down >`,
	}

	migrateCmd.PersistentFlags().IntVar(&step, "step", 0, "migration step")
	migrateCmd.AddCommand(initMigrationUpCmd(dbConfigKeys...))
	migrateCmd.AddCommand(initMigrationDownCmd(dbConfigKeys...))

	return migrateCmd
}
