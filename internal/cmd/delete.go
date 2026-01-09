package cmd

import "github.com/spf13/cobra"

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete mail system objects",
}

func init() {
	// Add common flags
	DeleteCmd.PersistentFlags().BoolP("permanent", "p", false, "Perform permanent deletion instead of soft delete")
	DeleteCmd.PersistentFlags().BoolP("force", "f", false, "Delete even if already soft-deleted")

	// Add subcommands
	DeleteCmd.AddCommand(DeleteDomainsCmd)
	DeleteCmd.AddCommand(DeleteMailboxesCmd)
	DeleteCmd.AddCommand(DeleteAliasesCmd)
	DeleteCmd.AddCommand(DeleteAliasTargetsCmd)
	DeleteCmd.AddCommand(DeleteDomainCatchallTargetsCmd)
	DeleteCmd.AddCommand(DeleteRecipientsRelayedCmd)
	DeleteCmd.AddCommand(DeleteTransportsCmd)
	DeleteCmd.AddCommand(DeleteRemotesCmd)
	DeleteCmd.AddCommand(DeleteRemoteSendGrantsCmd)
}
