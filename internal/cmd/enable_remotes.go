package cmd

import (
	"database/sql"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var EnableRemotesCmd = &cobra.Command{
	Use:     "remotes <name> [<name>...]",
	Aliases: []string{"remote"},
	Short:   "Enables remotes",
	Long:    "Enables remotes.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		enabled := true
		options := db.RemotesPatchOptions{
			Enabled: &enabled,
		}

		runner := db.TxForEachRunner[string]{
			Items: args,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Remotes(tx).Patch(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to enable remote",
			SuccessMessage: "Successfully enabled remote",
		}

		runner.Run()
		return nil
	},
}
