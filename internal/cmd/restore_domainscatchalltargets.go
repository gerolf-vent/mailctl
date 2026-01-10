package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var RestoreDomainCatchallTargetsCmd = &cobra.Command{
	Use:     "catchall-targets <domain> <target-email> [<target-email>...]",
	Aliases: []string{"catchall-target", "catchalls", "catchall"},
	Short:   "Restores soft-deleted domain catch-all targets",
	Long:    "Restores soft-deleted domain catch-all targets.",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		argDomain := args[0]
		argEmails := ParseEmailArgs(args[1:])
		if len(argEmails) != len(args)-1 {
			return fmt.Errorf("invalid email arguments")
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.DomainsCatchallTargets(tx).Restore(argDomain, item)
			},
			ItemString:     func(item utils.EmailAddress) string { return "@" + argDomain + " -> " + item.String() },
			FailureMessage: "failed to restore catchall target",
			SuccessMessage: "Successfully restored catchall target",
		}

		runner.Run()
		return nil
	},
}
