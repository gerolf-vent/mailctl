package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var PatchDomainCatchallTargetsCmd = &cobra.Command{
	Use:     "catchall-targets [flags] <domain> <target> [<target>...]",
	Aliases: []string{"catchall-target", "catchalls", "catchall"},
	Short:   "Updates existing catchall targets",
	Long:    "Updates specified properties for existing catchall targets.",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("forward") && !cmd.Flags().Changed("fallback-only") {
			return fmt.Errorf("at least one of --forward or --fallback-only flags must be specified")
		}

		argDomain := args[0]
		argEmails := ParseEmailArgs(args[1:])
		if len(argEmails) != len(args)-1 {
			return fmt.Errorf("invalid email arguments")
		}

		options := db.DomainsCatchallTargetsPatchOptions{}
		if cmd.Flags().Changed("forward") {
			flagForward, _ := cmd.Flags().GetBool("forward")
			options.ForwardingToTargetEnabled = &flagForward
		}
		if cmd.Flags().Changed("fallback-only") {
			flagOnly, _ := cmd.Flags().GetBool("fallback-only")
			options.FallbackOnly = &flagOnly
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.DomainsCatchallTargets(tx).Patch(argDomain, item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return "@" + argDomain + " -> " + item.String() },
			FailureMessage: "failed to patch catchall target",
			SuccessMessage: "Successfully patched catchall target",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchDomainCatchallTargetsCmd.Flags().BoolP("forward", "f", false, "Enable or disable forwarding to the catchall target")
	PatchDomainCatchallTargetsCmd.Flags().BoolP("fallback-only", "b", false, "Set fallback-only mode")
}
