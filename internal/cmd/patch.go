package cmd

import "github.com/spf13/cobra"

var PatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Patch mail system objects",
}

func init() {
	// Add subcommands
	PatchCmd.AddCommand(PatchDomainsCmd)
	PatchCmd.AddCommand(PatchMailboxesCmd)
	PatchCmd.AddCommand(PatchAliasesCmd)
	PatchCmd.AddCommand(PatchAliasTargetsCmd)
	PatchCmd.AddCommand(PatchDomainCatchallTargetsCmd)
	PatchCmd.AddCommand(PatchRecipientsRelayedCmd)
	PatchCmd.AddCommand(PatchTransportsCmd)
	PatchCmd.AddCommand(PatchRemotesCmd)
}
