package cmd

import (
	"database/sql"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var DeleteTransportsCmd = &cobra.Command{
	Use:     "transports <name> [name...]",
	Aliases: []string{"transport"},
	Short:   "Delete a transport",
	Long:    "Delete a transport. By default performs a soft delete.",
	Args:    cobra.MinimumNArgs(1),
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
				return db.Transports(tx).Delete(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to delete transport",
			SuccessMessage: "Successfully deleted transport",
		}

		if flagPermanent {
			runner.SuccessMessage += " permanently"
		}

		runner.Run()
		return nil
	},
}
