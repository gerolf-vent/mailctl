package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var PatchAliasTargetsCmd = &cobra.Command{
	Use:     "alias-target <email> [flags]",
	Aliases: []string{"target"},
	Short:   "Update an existing alias target",
	Long:    "Updates properties of an existing alias target.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("forward") && !cmd.Flags().Changed("send") {
			return fmt.Errorf("at least one of --forward or --send flags must be specified")
		}

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		argAliasEmail := argEmails[0]
		argTargetEmails := argEmails[1:]

		options := db.AliasesTargetsPatchOptions{}
		if cmd.Flags().Changed("forward") {
			flagForward, _ := cmd.Flags().GetBool("forward")
			options.ForwardingToTargetEnabled = &flagForward
		}
		if cmd.Flags().Changed("send") {
			flagSend, _ := cmd.Flags().GetBool("send")
			options.SendingFromTargetEnabled = &flagSend
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argTargetEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.AliasesTargets(tx).Patch(argAliasEmail, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return argAliasEmail.String() + " -> " + item.String() },
			FailureMessage: "failed to patch alias target",
			SuccessMessage: "Successfully patch alias target",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchAliasTargetsCmd.Flags().BoolP("forward", "f", false, "Enable/disable forwarding to target")
	PatchAliasTargetsCmd.Flags().BoolP("send", "s", false, "Enable/disable sending from target")
}
