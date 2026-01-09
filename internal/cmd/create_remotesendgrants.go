package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var CreateRemoteSendGrantsCmd = &cobra.Command{
	Use:   "send-grant <remote> <email|domain>",
	Short: "Create a new send grant",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		argRemoteName := args[0]
		argEmails := ParseEmailOrWildcardArgs(args[1:])
		if len(argEmails) != len(args[1:]) {
			return fmt.Errorf("invalid email or domain argument")
		}
		argEmail := argEmails[0]

		options := db.RemotesSendGrantsCreateOptions{}

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.RemotesSendGrants(tx).Create(argRemoteName, argEmail, options)
			},
			FailureMessage: "failed to create send grant",
			SuccessMessage: "Successfully created send grant",
		}

		runner.Run()
		return nil
	},
}
