package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var EnableMailboxesCmd = &cobra.Command{
	Use:     "mailboxes <email> [email...]",
	Aliases: []string{"mailbox"},
	Short:   "Enables mailboxes",
	Long:    "Enables login, receiving, and/or sending for mailboxes.\nEmails must be in the format \"name@example.com\".\nIf no specific flags are provided, all three (login, receiving, sending) will be enabled.",
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

		enabled := true
		options := db.MailboxesPatchOptions{}
		if flagLogin {
			options.Login = &enabled
		}
		if flagReceiving {
			options.Receiving = &enabled
		}
		if flagSending {
			options.Sending = &enabled
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.Mailboxes(tx).Patch(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to enable mailbox",
			SuccessMessage: "Successfully enabled mailbox",
		}

		runner.Run()
		return nil
	},
}

func init() {
	EnableMailboxesCmd.Flags().BoolP("login", "l", false, "Enable login only")
	EnableMailboxesCmd.Flags().BoolP("receiving", "r", false, "Enable receiving only")
	EnableMailboxesCmd.Flags().BoolP("sending", "s", false, "Enable sending only")
}
