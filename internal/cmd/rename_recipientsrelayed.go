package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RenameRecipientsRelayedCmd = &cobra.Command{
	Use:     "recipient-relayed <email> <new-email>",
	Aliases: []string{"recipients-relayed", "relayed-recipients", "relayed-recipient", "relayed"},
	Short:   "Renames a relayed recipient",
	Long:    "Renames a relayed recipient. This also supports changing the domain without loosing any of the relations.",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}
		oldEmail := argEmails[0]
		newEmail := argEmails[1]

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.RecipientsRelayed(tx).Rename(oldEmail, newEmail)
			},
			ItemString:     fmt.Sprintf("%s -> %s", oldEmail.String(), newEmail.String()),
			FailureMessage: "failed to rename relayed recipient",
			SuccessMessage: "Successfully renamed relayed recipient",
		}

		runner.Run()
		return nil
	},
}
