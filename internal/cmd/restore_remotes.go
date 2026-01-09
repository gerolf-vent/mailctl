package cmd

import (
	"database/sql"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RestoreRemotesCmd = &cobra.Command{
	Use:   "remotes <name> [<name>...]",
	Short: "Restore one or more soft-deleted remotes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := db.TxForEachRunner[string]{
			Items: args,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Remotes(tx).Restore(item)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to restore remote",
			SuccessMessage: "Successfully restored remote",
		}

		runner.Run()
		return nil
	},
}
