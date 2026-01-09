package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DisableMailboxesCmd = &cobra.Command{
	Use:     "mailboxes <email> [email...]",
	Aliases: []string{"mailbox"},
	Short:   "Disables mailboxes",
	Long:    "Disables login, receiving, and/or sending for mailboxes.\nEmails must be in the format \"name@example.com\".\nIf no specific flags are provided, all three (login, receiving, sending) will be disabled.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagLogin, _ := cmd.Flags().GetBool("login")
		flagReceiving, _ := cmd.Flags().GetBool("receiving")
		flagSending, _ := cmd.Flags().GetBool("sending")

		if !flagLogin && !flagReceiving && !flagSending {
			return fmt.Errorf("at least one of --login, --receiving, or --sending must be true when specified")
		}

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		disabled := false
		options := db.MailboxesPatchOptions{}
		if flagLogin {
			options.Login = &disabled
		}
		if flagReceiving {
			options.Receiving = &disabled
		}
		if flagSending {
			options.Sending = &disabled
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.Mailboxes(tx).Patch(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to patch mailbox",
			SuccessMessage: "Successfully patched mailbox",
		}

		runner.Run()
		return nil
	},
}

func init() {
	DisableMailboxesCmd.Flags().BoolP("login", "l", false, "Disable login only")
	DisableMailboxesCmd.Flags().BoolP("receiving", "r", false, "Disable receiving only")
	DisableMailboxesCmd.Flags().BoolP("sending", "s", false, "Disable sending only")
}
