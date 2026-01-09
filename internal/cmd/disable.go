package cmd

import "github.com/spf13/cobra"

var DisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable mail system objects",
}

func init() {
	// Add subcommands
	DisableCmd.AddCommand(DisableDomainsCmd)
	DisableCmd.AddCommand(DisableMailboxesCmd)
	DisableCmd.AddCommand(DisableAliasesCmd)
	DisableCmd.AddCommand(DisableAliasTargetsCmd)
	DisableCmd.AddCommand(DisableDomainCatchallTargetsCmd)
	DisableCmd.AddCommand(DisableRecipientsRelayed)
	DisableCmd.AddCommand(DisableRemotesCmd)
}
