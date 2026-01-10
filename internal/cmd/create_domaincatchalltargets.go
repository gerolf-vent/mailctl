package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var CreateDomainCatchallTargetsCmd = &cobra.Command{
	Use:     "catchall-targets [flags] <domain> <target-email> [<target-email>...]",
	Aliases: []string{"catchall-target", "catchalls", "catchall"},
	Short:   "Creates new catch-all targets for a domain",
	Long:    "Creates new catch-all targets for a domain.",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagForward, _ := cmd.Flags().GetBool("forward")
		flagFallbackOnly, _ := cmd.Flags().GetBool("fallback-only")

		argDomain := args[0]
		argEmails := ParseEmailArgs(args[1:])
		if len(argEmails) != len(args)-1 {
			return fmt.Errorf("invalid email arguments")
		}

		options := db.DomainsCatchallTargetsCreateOptions{
			ForwardEnabled: flagForward,
			FallbackOnly:   flagFallbackOnly,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.DomainsCatchallTargets(tx).Create(argDomain, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return "@" + argDomain + " -> " + item.String() },
			FailureMessage: "failed to create catchall target",
			SuccessMessage: "Successfully created catchall target",
		}

		runner.Run()
		return nil
	},
}

func init() {
	CreateDomainCatchallTargetsCmd.Flags().BoolP("forward", "f", true, "Enable forwarding to target (default: true)")
	CreateDomainCatchallTargetsCmd.Flags().BoolP("fallback-only", "b", true, "Only forward if no other recipient matched (default: true)")
}
