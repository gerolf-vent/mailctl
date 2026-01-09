package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var PatchRecipientsRelayedCmd = &cobra.Command{
	Use:     "recipients-relayed <email> [flags]",
	Aliases: []string{"recipient-relayed", "relayed-recipient", "relayed-recipients", "relayed"},
	Short:   "Update an existing relayed recipient",
	Long:    "Updates properties of an existing relayed recipient. Emails must be in the format \"name@example.com\".",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagEnabled, _ := cmd.Flags().GetBool("enabled")

		if !cmd.Flags().Changed("enabled") {
			return fmt.Errorf("no changes specified")
		}

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		options := db.RecipientsRelayedPatchOptions{}
		if cmd.Flags().Changed("enabled") {
			options.Enabled = &flagEnabled
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.RecipientsRelayed(tx).Patch(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to patch relayed recipient",
			SuccessMessage: "Successfully patched relayed recipient",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchRecipientsRelayedCmd.Flags().BoolP("enabled", "e", false, "Enable or disable the relayed recipient")
	PatchRecipientsRelayedCmd.Flags().String("email", "", "New email address")
}
