package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var RestoreRemoteSendGrantsCmd = &cobra.Command{
	Use:   "send-grant <remote-name> <email> [<email>...]",
	Short: "Restores soft-deleted send grants for a remote",
	Long:  "Restores soft-deleted send grants for a remote.",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		argRemoteName := args[0]
		argEmails := ParseEmailOrWildcardArgs(args[1:])
		if len(argEmails) != len(args[1:]) {
			return fmt.Errorf("invalid email or domain argument")
		}

		runner := db.TxForEachRunner[utils.EmailAddressOrWildcard]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddressOrWildcard) error {
				return db.RemotesSendGrants(tx).Restore(argRemoteName, item)
			},
			ItemString:     func(item utils.EmailAddressOrWildcard) string { return argRemoteName + " -> " + item.String() },
			FailureMessage: "failed to restore send grant",
			SuccessMessage: "Successfully restored send grant",
		}

		runner.Run()
		return nil
	},
}
