package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RenameRecipientsRelayedCmd = &cobra.Command{
	Use:     "recipient-relayed <email> <new-email>",
	Aliases: []string{"relayed-recipient", "relayed"},
	Short:   "Rename a relayed recipient",
	Long:    "Rename a relayed recipient.\nEmails must be in the format \"name@example.com\".",
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
