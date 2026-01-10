package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RenameRemotesCmd = &cobra.Command{
	Use:   "remote <old-name> <new-name>",
	Short: "Rename a remote",
	Long:  "Rename a remote by changing its name.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := args[0]
		newName := args[1]

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.Remotes(tx).Rename(oldName, newName)
			},
			ItemString:     fmt.Sprintf("%s -> %s", oldName, newName),
			FailureMessage: "failed to rename remote",
			SuccessMessage: "Successfully renamed remote",
		}

		runner.Run()
		return nil
	},
}
