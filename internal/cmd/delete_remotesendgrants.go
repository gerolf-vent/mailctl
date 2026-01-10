package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DeleteRemoteSendGrantsCmd = &cobra.Command{
	Use:     "send-grants <remote-name> <email> [<email>...]",
	Aliases: []string{"send-grant"},
	Short:   "Deletes send grants from a remote",
	Long:    "Deletes send grants from a remote. By default performs a soft delete. Use --permanent --force for hard delete.",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPermanent, _ := cmd.Flags().GetBool("permanent")
		flagForce, _ := cmd.Flags().GetBool("force")

		if flagPermanent && flagForce {
			return fmt.Errorf("cannot use --permanent and --force flags together")
		}

		argRemoteName := args[0]
		argEmails := ParseEmailOrWildcardArgs(args[1:])
		if len(argEmails) != len(args[1:]) {
			return fmt.Errorf("invalid email argument")
		}

		options := db.DeleteOptions{
			Permanent: flagPermanent,
			Force:     flagForce,
		}

		runner := db.TxForEachRunner[utils.EmailAddressOrWildcard]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddressOrWildcard) error {
				return db.RemotesSendGrants(tx).Delete(argRemoteName, item, options)
			},
			ItemString:     func(item utils.EmailAddressOrWildcard) string { return argRemoteName + " -> " + item.String() },
			FailureMessage: "failed to delete send grant",
			SuccessMessage: "Successfully deleted send grant",
		}

		runner.Run()
		return nil
	},
}
