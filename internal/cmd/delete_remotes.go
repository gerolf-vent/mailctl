package cmd

import (
	"database/sql"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var DeleteRemotesCmd = &cobra.Command{
	Use:   "remotes <name> [<name>...]",
	Short: "Delete one or more remotes (soft delete by default)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPermanent, _ := cmd.Flags().GetBool("permanent")
		flagForce, _ := cmd.Flags().GetBool("force")

		options := db.DeleteOptions{
			Permanent: flagPermanent,
			Force:     flagForce,
		}

		runner := db.TxForEachRunner[string]{
			Items: args,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Remotes(tx).Delete(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to delete remote",
			SuccessMessage: "Successfully deleted remote",
		}

		runner.Run()
		return nil
	},
}
