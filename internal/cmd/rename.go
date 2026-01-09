package cmd

import "github.com/spf13/cobra"

var RenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename mail system objects",
}

func init() {
	// Add subcommands
	RenameCmd.AddCommand(RenameDomainsCmd)
	RenameCmd.AddCommand(RenameMailboxesCmd)
	RenameCmd.AddCommand(RenameAliasesCmd)
	RenameCmd.AddCommand(RenameRecipientsRelayedCmd)
	RenameCmd.AddCommand(RenameTransportsCmd)
	RenameCmd.AddCommand(RenameRemotesCmd)
}
