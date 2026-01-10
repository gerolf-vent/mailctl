package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DisableDomainCatchallTargetsCmd = &cobra.Command{
	Use:     "catchall-targets <domain> <target-email> [<target-email>...]",
	Aliases: []string{"catchall-target", "catchalls", "catchall"},
	Short:   "Disables forwarding on catch-all targets of a domain",
	Long:    "Disables forwarding on catch-all targets of a domain.",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		argDomain := args[0]
		argEmails := ParseEmailArgs(args[1:])
		if len(argEmails) != len(args)-1 {
			return fmt.Errorf("invalid email arguments")
		}

		enabled := false
		options := db.DomainsCatchallTargetsPatchOptions{
			ForwardingToTargetEnabled: &enabled,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.DomainsCatchallTargets(tx).Patch(argDomain, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return "@" + argDomain + " -> " + item.String() },
			FailureMessage: "failed to disable catchall target",
			SuccessMessage: "Successfully disable catchall target",
		}

		runner.Run()
		return nil
	},
}
