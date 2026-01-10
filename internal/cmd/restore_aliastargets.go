package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var RestoreAliasTargetsCmd = &cobra.Command{
	Use:     "alias-targets <alias-email> <target-email> [<target-email>...]",
	Aliases: []string{"alias-target", "targets", "target"},
	Short:   "Restores soft-deleted alias targets",
	Long:    "Restore soft-deleted targets to an alias.",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		argAliasEmail := argEmails[0]
		argTargetEmails := argEmails[1:]

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argTargetEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.AliasesTargets(tx).Restore(argAliasEmail, item)
			},
			ItemString:     func(item utils.EmailAddress) string { return argAliasEmail.String() + " -> " + item.String() },
			FailureMessage: "failed to restore alias target",
			SuccessMessage: "Successfully restore alias target",
		}

		runner.Run()
		return nil
	},
}
