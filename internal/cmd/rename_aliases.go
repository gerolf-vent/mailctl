package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RenameAliasesCmd = &cobra.Command{
	Use:     "alias <email> <new-email>",
	Aliases: []string{"aliases"},
	Short:   "Rename an alias",
	Long:    "Rename an alias.\nEmail must be in the format \"name@example.com\".",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}
		argOldEmail := argEmails[0]
		argNewEmail := argEmails[1]

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.Aliases(tx).Rename(argOldEmail, argNewEmail)
			},
			ItemString:     fmt.Sprintf("%s -> %s", argOldEmail.String(), argNewEmail.String()),
			FailureMessage: "failed to rename alias",
			SuccessMessage: "Successfully renamed alias",
		}

		runner.Run()
		return nil
	},
}
