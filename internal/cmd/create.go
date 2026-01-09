package cmd

import "github.com/spf13/cobra"

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create mail system objects",
}

func init() {
	// Add subcommands
	CreateCmd.AddCommand(CreateDomainsCmd)
	CreateCmd.AddCommand(CreateMailboxesCmd)
	CreateCmd.AddCommand(CreateAliasesCmd)
	CreateCmd.AddCommand(CreateAliasTargetsCmd)
	CreateCmd.AddCommand(CreateDomainCatchallTargetsCmd)
	CreateCmd.AddCommand(CreateRecipientsRelayedCmd)
	CreateCmd.AddCommand(CreateTransportsCmd)
	CreateCmd.AddCommand(CreateRemotesCmd)
	CreateCmd.AddCommand(CreateRemoteSendGrantsCmd)
}
