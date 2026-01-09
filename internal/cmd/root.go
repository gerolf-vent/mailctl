package cmd

import (
	"os"

	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "mailctl",
	Short:         "Mail system management CLI",
	Long:          `mailctl is a command-line interface for managing the mail database system including domains, mailboxes, aliases, transports, and more.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	// Add main commands
	rootCmd.AddCommand(ListCmd)
	rootCmd.AddCommand(CreateCmd)
	rootCmd.AddCommand(SchemaCmd)
}

func Execute() {
	cmd, err := rootCmd.ExecuteC()
	if err != nil {
		utils.PrintError(err)
		if cmd == nil {
			_ = rootCmd.Usage()
		} else {
			_ = cmd.Usage()
		}
		os.Exit(1)
	}
}
