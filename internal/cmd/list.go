package cmd

import "github.com/spf13/cobra"

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List mail system objects",
}

func init() {
	// Add common flags
	ListCmd.PersistentFlags().BoolP("deleted", "d", false, "Show soft-deleted entries")
	ListCmd.PersistentFlags().BoolP("all", "a", false, "Show all entries, including soft-deleted ones")
	ListCmd.PersistentFlags().BoolP("json", "j", false, "Output in JSON format")
	ListCmd.PersistentFlags().BoolP("verbose", "v", false, "Show created/updated timestamps where applicable")

	// Add subcommands
	ListCmd.AddCommand(ListDomainsCmd)
	ListCmd.AddCommand(ListMailboxesCmd)
	ListCmd.AddCommand(ListAliasesCmd)
	ListCmd.AddCommand(ListAliasTargetsCmd)
	ListCmd.AddCommand(ListDomainCatchallTargetsCmd)
	ListCmd.AddCommand(ListRecipientsRelayedCmd)
	ListCmd.AddCommand(ListTransportsCmd)
	ListCmd.AddCommand(ListRemotesCmd)
	ListCmd.AddCommand(ListRemoteSendGrantsCmd)
}
