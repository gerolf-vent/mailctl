package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var CreateAliasTargetsCmd = &cobra.Command{
	Use:     "alias-targets [flags] <alias-email> <target-email> [<target-email>...]",
	Aliases: []string{"alias-target", "targets", "target"},
	Short:   "Creates new alias targets for an alias",
	Long:    "Creates new alias targets for an alias.",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagForward, _ := cmd.Flags().GetBool("forward")
		flagSend, _ := cmd.Flags().GetBool("send")

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		argAliasEmail := argEmails[0]
		argTargetEmails := argEmails[1:]

		options := db.AliasesTargetsCreateOptions{
			ForwardEnabled: flagForward,
		}
		if cmd.Flags().Changed("send") {
			options.SendEnabled = &flagSend
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argTargetEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.AliasesTargets(tx).Create(argAliasEmail, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return argAliasEmail.String() + " -> " + item.String() },
			FailureMessage: "failed to create alias target",
			SuccessMessage: "Successfully created alias target",
		}

		runner.Run()
		return nil
	},
}

func init() {
	CreateAliasTargetsCmd.Flags().BoolP("forward", "f", false, "Enable forwarding to target (default: false)")
	CreateAliasTargetsCmd.Flags().BoolP("send", "s", false, "Enable sending from target (default: false)")
}
