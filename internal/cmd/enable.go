package cmd

import "github.com/spf13/cobra"

var EnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable mail system objects",
}

func init() {
	// Add enable subcommands
	EnableCmd.AddCommand(EnableDomainsCmd)
	EnableCmd.AddCommand(EnableMailboxesCmd)
	EnableCmd.AddCommand(EnableAliasesCmd)
	EnableCmd.AddCommand(EnableAliasTargetsCmd)
	EnableCmd.AddCommand(EnableDomainCatchallTargetsCmd)
	EnableCmd.AddCommand(EnableRecipientsRelayedCmd)
	EnableCmd.AddCommand(EnableRemotesCmd)
}
