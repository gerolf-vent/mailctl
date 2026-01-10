package cmd

import (
	"database/sql"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RestoreTransportsCmd = &cobra.Command{
	Use:     "transports <name> [<name>...]",
	Aliases: []string{"transport"},
	Short:   "Restores soft-deleted transports",
	Long:    "Restores soft-deleted transports.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := db.TxForEachRunner[string]{
			Items: args,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Transports(tx).Restore(item)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to restore transport",
			SuccessMessage: "Successfully restored transport",
		}

		runner.Run()
		return nil
	},
}
