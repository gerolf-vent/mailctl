package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var CreateRecipientsRelayedCmd = &cobra.Command{
	Use:     "recipients-relayed <email> [email...]",
	Aliases: []string{"recipient-relayed", "relayed-recipient", "relayed-recipients", "relayed"},
	Short:   "Create a new relayed recipient",
	Long:    "Creates a new relayed recipient.\nEmails must be in the format \"name@example.com\".",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDisabled, _ := cmd.Flags().GetBool("disabled")

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		options := db.RecipientsRelayedCreateOptions{
			Enabled: !flagDisabled,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.RecipientsRelayed(tx).Create(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to create relayed recipient",
			SuccessMessage: "Successfully created relayed recipient",
		}

		runner.Run()
		return nil
	},
}

func init() {
	CreateRecipientsRelayedCmd.Flags().BoolP("disabled", "d", false, "Create the relayed recipient in disabled state")
}
