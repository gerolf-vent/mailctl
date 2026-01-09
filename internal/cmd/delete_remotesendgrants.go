package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var DeleteRemoteSendGrantsCmd = &cobra.Command{
	Use:   "send-grant <remote> <email|domain>",
	Short: "Delete a send grant",
	Long:  "Delete a send grant from a remote. By default performs a soft delete. Use --permanent --force for hard delete.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPermanent, _ := cmd.Flags().GetBool("permanent")
		flagForce, _ := cmd.Flags().GetBool("force")

		if flagPermanent && flagForce {
			return fmt.Errorf("cannot use --permanent and --force flags together")
		}

		argRemoteName := args[0]
		argEmails := ParseEmailOrWildcardArgs(args[1:])
		if len(argEmails) != len(args[1:]) {
			return fmt.Errorf("invalid email or domain argument")
		}
		argEmail := argEmails[0]

		options := db.DeleteOptions{
			Permanent: flagPermanent,
			Force:     flagForce,
		}

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.RemotesSendGrants(tx).Delete(argRemoteName, argEmail, options)
			},
			FailureMessage: "failed to delete send grant",
			SuccessMessage: "Successfully deleted send grant",
		}

		runner.Run()
		return nil
	},
}
