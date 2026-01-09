package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DeleteAliasesCmd = &cobra.Command{
	Use:     "aliases <email> [email...]",
	Aliases: []string{"alias"},
	Short:   "Deletes aliases",
	Long:    "Deletes aliases. By default performs a soft delete. Use --permanent for hard delete.\nEmails must be in the format \"name@example.com\".",
	Args:    cobra.MinimumNArgs(1),
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

		options := db.DeleteOptions{
			Permanent: flagPermanent,
			Force:     flagForce,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.Aliases(tx).Delete(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to delete alias",
			SuccessMessage: "Successfully deleted alias",
		}

		if flagPermanent {
			runner.SuccessMessage += " permanently"
		}

		runner.Run()
		return nil
	},
}
