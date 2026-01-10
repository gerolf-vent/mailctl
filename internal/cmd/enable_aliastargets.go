package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var EnableAliasTargetsCmd = &cobra.Command{
	Use:     "alias-targets [flags] <alias-email> <target-email> [<target-email>...]",
	Aliases: []string{"alias-target", "targets", "target"},
	Short:   "Enables features on alias targets of an alias",
	Long:    `Enables forwarding and/or sending on an alias target. Use flags to select which property to enable.`,
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

		enabled := true
		options := db.AliasesTargetsPatchOptions{}
		if flagForward {
			options.ForwardingToTargetEnabled = &enabled
		}
		if flagSend {
			options.SendingFromTargetEnabled = &enabled
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argTargetEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.AliasesTargets(tx).Patch(argAliasEmail, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return argAliasEmail.String() + " -> " + item.String() },
			FailureMessage: "failed to enable alias target",
			SuccessMessage: "Successfully enable alias target",
		}

		runner.Run()
		return nil
	},
}

func init() {
	EnableAliasTargetsCmd.Flags().BoolP("forward", "f", false, "Enable forwarding to target")
	EnableAliasTargetsCmd.Flags().BoolP("send", "s", false, "Enable sending from target")
}
