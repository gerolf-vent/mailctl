package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DeleteAliasTargetsCmd = &cobra.Command{
	Use:     "alias-targets [flags] <alias-email> <target-email> [<target-email>...]",
	Aliases: []string{"alias-target", "targets", "target"},
	Short:   "Deletes alias targets from an alias",
	Long:    `Deletes alias targets from an alias. By default performs a soft delete. Use --permanent for hard delete.`,
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPermanent, _ := cmd.Flags().GetBool("permanent")
		flagForce, _ := cmd.Flags().GetBool("force")

		if flagPermanent && flagForce {
			return fmt.Errorf("cannot use --permanent and --force flags together")
		}

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		argAliasEmail := argEmails[0]
		argTargetEmails := argEmails[1:]

		options := db.DeleteOptions{
			Permanent: flagPermanent,
			Force:     flagForce,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argTargetEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.AliasesTargets(tx).Delete(argAliasEmail, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return argAliasEmail.String() + " -> " + item.String() },
			FailureMessage: "failed to delete alias target",
			SuccessMessage: "Successfully deleted alias target",
		}

		if flagPermanent {
			runner.SuccessMessage += " permanently"
		}

		runner.Run()
		return nil
	},
}
