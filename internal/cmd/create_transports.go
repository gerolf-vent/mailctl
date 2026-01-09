package cmd

import (
	"database/sql"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var CreateTransportsCmd = &cobra.Command{
	Use:   "transport <name> [name...]",
	Short: "Create a new transport",
	Long:  "Creates a new mail transport configuration.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		flagMethod, _ := cmd.Flags().GetString("method")
		flagHost, _ := cmd.Flags().GetString("host")
		flagPort, _ := cmd.Flags().GetUint16("port")
		flagMxLookup, _ := cmd.Flags().GetBool("mx-lookup")

		options := db.TransportsCreateOptions{
			Method:   flagMethod,
			Host:     flagHost,
			MxLookup: flagMxLookup,
		}
		if flagPort > 0 {
			options.Port = sql.NullInt32{Int32: int32(flagPort), Valid: true}
		}

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.Transports(tx).Create(name, options)
			},
			ItemString:     name,
			FailureMessage: "failed to create transport",
			SuccessMessage: "Successfully created transport",
		}

		runner.Run()
		return nil
	},
}

func init() {
	CreateTransportsCmd.Flags().StringP("method", "m", "", "Transport method (required, e.g. 'lmtp', 'smtp', or 'relay')")
	CreateTransportsCmd.MarkFlagRequired("method")
	CreateTransportsCmd.Flags().String("host", "", "Remote server hostname (required)")
	CreateTransportsCmd.MarkFlagRequired("host")
	CreateTransportsCmd.Flags().Uint16("port", 0, "Transport port")
	CreateTransportsCmd.Flags().Bool("mx-lookup", false, "Enable MX lookup")
}
