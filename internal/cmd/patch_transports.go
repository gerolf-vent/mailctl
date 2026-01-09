package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var PatchTransportsCmd = &cobra.Command{
	Use:     "transports <name> [flags]",
	Aliases: []string{"transport"},
	Short:   "Update an existing transport",
	Long:    "Updates properties of an existing transport.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagMethod, _ := cmd.Flags().GetString("method")
		flagHost, _ := cmd.Flags().GetString("host")
		flagPort, _ := cmd.Flags().GetUint16("port")
		flagMxLookup, _ := cmd.Flags().GetBool("mx-lookup")

		// Check if at least one flag was changed
		if !cmd.Flags().Changed("method") && !cmd.Flags().Changed("host") && !cmd.Flags().Changed("port") && !cmd.Flags().Changed("mx-lookup") {
			return fmt.Errorf("no changes specified")
		}

		options := db.TransportsPatchOptions{}
		if cmd.Flags().Changed("method") {
			options.Method = &flagMethod
		}
		if cmd.Flags().Changed("host") {
			options.Host = &flagHost
		}
		if cmd.Flags().Changed("port") {
			if flagPort > 0 {
				options.Port = &sql.NullInt32{Int32: int32(flagPort), Valid: true}
			} else {
				options.Port = &sql.NullInt32{Valid: false}
			}
		}
		if cmd.Flags().Changed("mx-lookup") {
			options.MxLookup = &flagMxLookup
		}

		runner := db.TxForEachRunner[string]{
			Items: args,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Transports(tx).Patch(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to patch transport",
			SuccessMessage: "Successfully patched transport",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchTransportsCmd.Flags().StringP("method", "m", "", "New transport method (lmtp, smtp, relay)")
	PatchTransportsCmd.Flags().StringP("host", "h", "", "New transport host")
	PatchTransportsCmd.Flags().Uint16("port", 0, "New transport port")
	PatchTransportsCmd.Flags().Bool("mx-lookup", false, "Enable/disable MX lookup")
}
