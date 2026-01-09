package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DisableAliasTargetsCmd = &cobra.Command{
	Use:     "alias-target <alias-email> <target-email>",
	Aliases: []string{"target"},
	Short:   "Disable properties on an alias target",
	Long:    `Disable forwarding and/or sending on an alias target. Use flags to select which property to disable.`,
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagForward, _ := cmd.Flags().GetBool("forward")
		flagSend, _ := cmd.Flags().GetBool("send")

		if !flagForward && !flagSend {
			return fmt.Errorf("at least one of --forward or --send flags must be specified")
		}

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		argAliasEmail := argEmails[0]
		argTargetEmails := argEmails[1:]

		disabled := false
		options := db.AliasesTargetsPatchOptions{}
		if flagForward {
			options.ForwardingToTargetEnabled = &disabled
		}
		if flagSend {
			options.SendingFromTargetEnabled = &disabled
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argTargetEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.AliasesTargets(tx).Patch(argAliasEmail, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return argAliasEmail.String() + " -> " + item.String() },
			FailureMessage: "failed to disable alias target",
			SuccessMessage: "Successfully disable alias target",
		}

		runner.Run()
		return nil
	},
}

func init() {
	DisableAliasTargetsCmd.Flags().BoolP("forward", "f", false, "Disable forwarding to target")
	DisableAliasTargetsCmd.Flags().BoolP("send", "s", false, "Disable sending from target")
}
