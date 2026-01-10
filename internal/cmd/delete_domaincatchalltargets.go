package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DeleteDomainCatchallTargetsCmd = &cobra.Command{
	Use:     "catchall-targets [flags] <domain> <target-email> [<target-email>...]",
	Aliases: []string{"catchall-target", "catchalls", "catchall"},
	Short:   "Deletes catch-all targets from a domain",
	Long:    "Deletes catch-all targets from a domain. By default performs a soft delete. Use --permanent for hard delete.",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPermanent, _ := cmd.Flags().GetBool("permanent")
		flagForce, _ := cmd.Flags().GetBool("force")

		if flagPermanent && flagForce {
			return fmt.Errorf("cannot use --permanent and --force flags together")
		}

		argDomain := args[0]
		argEmails := ParseEmailArgs(args[1:])
		if len(argEmails) != len(args)-1 {
			return fmt.Errorf("invalid email arguments")
		}

		options := db.DeleteOptions{
			Permanent: flagPermanent,
			Force:     flagForce,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.DomainsCatchallTargets(tx).Delete(argDomain, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return "@" + argDomain + " -> " + item.String() },
			FailureMessage: "failed to delete catchall target",
			SuccessMessage: "Successfully deleted catchall target",
		}

		if flagPermanent {
			runner.SuccessMessage += " permanently"
		}

		runner.Run()
		return nil
	},
}
