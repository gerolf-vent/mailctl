package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RenameMailboxesCmd = &cobra.Command{
	Use:     "mailbox <email> <new-email>",
	Aliases: []string{"mailboxes"},
	Short:   "Renames a mailbox",
	Long:    "Renames a mailbox. This also supports changing the domain without loosing any of the aliases, or other relations.",
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
				return db.Mailboxes(tx).Rename(oldEmail, newEmail)
			},
			ItemString:     fmt.Sprintf("%s -> %s", oldEmail, newEmail),
			FailureMessage: "failed to rename mailbox",
			SuccessMessage: "Successfully renamed mailbox",
		}

		runner.Run()
		return nil
	},
}
