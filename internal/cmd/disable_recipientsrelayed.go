package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DisableRecipientsRelayedCmd = &cobra.Command{
	Use:     "recipients-relayed <email> [<email>...]",
	Aliases: []string{"recipient-relayed", "relayed-recipients", "relayed-recipient", "relayed"},
	Short:   "Disables relayed recipients",
	Long:    "Disables relayed recipients.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		enabled := false
		options := db.RecipientsRelayedPatchOptions{
			Enabled: &enabled,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.RecipientsRelayed(tx).Patch(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to disable relayed recipient",
			SuccessMessage: "Successfully disabled relayed recipient",
		}

		runner.Run()
		return nil
	},
}
