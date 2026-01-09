package cmd

import "github.com/spf13/cobra"

var RestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore soft-deleted mail system objects",
}

func init() {
	// Add subcommands
	RestoreCmd.AddCommand(RestoreDomainsCmd)
	RestoreCmd.AddCommand(RestoreMailboxesCmd)
	RestoreCmd.AddCommand(RestoreAliasesCmd)
	RestoreCmd.AddCommand(RestoreAliasTargetsCmd)
	RestoreCmd.AddCommand(RestoreDomainCatchallTargetsCmd)
	RestoreCmd.AddCommand(RestoreRecipientsRelayedCmd)
	RestoreCmd.AddCommand(RestoreTransportsCmd)
	RestoreCmd.AddCommand(RestoreRemotesCmd)
	RestoreCmd.AddCommand(RestoreRemoteSendGrantsCmd)
}
