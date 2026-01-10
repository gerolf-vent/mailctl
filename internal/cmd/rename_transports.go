package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RenameTransportsCmd = &cobra.Command{
	Use:     "transport <old-name> <new-name>",
	Aliases: []string{"transports"},
	Short:   "Renames a transport",
	Long:    "Renames a transport by changing its name.",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := args[0]
		newName := args[1]

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.Transports(tx).Rename(oldName, newName)
			},
			ItemString:     fmt.Sprintf("%s -> %s", oldName, newName),
			FailureMessage: "failed to rename transport",
			SuccessMessage: "Successfully renamed transport",
		}

		runner.Run()
		return nil
	},
}
