package cmd

import (
	"github.com/spf13/cobra"
)

var SchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage database schema",
	Long:  `Manage database schema setup and upgrades`,
}

func init() {
	// Add subcommands
	SchemaCmd.AddCommand(SchemaStatusCmd)
	SchemaCmd.AddCommand(SchemaUpgradeCmd)
	SchemaCmd.AddCommand(SchemaPurgeCmd)
	SchemaCmd.AddCommand(SchemaEnsureUserCmd)
	SchemaCmd.AddCommand(SchemaDropUserCmd)
}
