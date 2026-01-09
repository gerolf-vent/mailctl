package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var PatchRemotesCmd = &cobra.Command{
	Use:   "remote <hostname> [flags]",
	Short: "Update an existing remote",
	Long:  "Updates properties of an existing remote.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPassword, _ := cmd.Flags().GetBool("password")
		flagPasswordStdin, _ := cmd.Flags().GetBool("password-stdin")
		flagPasswordMethod, _ := cmd.Flags().GetString("password-method")
		flagPasswordNo, _ := cmd.Flags().GetBool("no-password")

		if (flagPassword || flagPasswordStdin) && flagPasswordNo {
			return fmt.Errorf("cannot use --no-password with --password or --password-stdin")
		}

		if flagPassword && flagPasswordStdin {
			return fmt.Errorf("cannot use both --password and --password-stdin")
		}

		if !flagPassword && !flagPasswordStdin && !flagPasswordNo && !cmd.Flags().Changed("enabled") {
			return fmt.Errorf("no changes specified. Use --password, --no-password, or --enabled flags")
		}

		name := args[0]

		options := db.RemotesPatchOptions{}

		if cmd.Flags().Changed("enabled") {
			flagEnabled, _ := cmd.Flags().GetBool("enabled")
			options.Enabled = &flagEnabled
		}

		if flagPassword || flagPasswordStdin {
			passwordHash, err := ReadPasswordHashed(flagPasswordMethod, flagPasswordStdin)
			if err != nil {
				utils.PrintErrorWithMessage("failed to read password", err)
				return nil
			}
			options.PasswordHash = &sql.NullString{
				String: passwordHash,
				Valid:  true,
			}
		}
		if flagPasswordNo {
			options.PasswordHash = &sql.NullString{
				Valid: false,
			}
		}

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.Remotes(tx).Patch(name, options)
			},
			ItemString:     name,
			FailureMessage: "failed to patch remote",
			SuccessMessage: "Successfully patched remote",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchRemotesCmd.Flags().Bool("enabled", false, "Enable or disable the remote")
	PatchRemotesCmd.Flags().String("username", "", "New SMTP username")
	PatchRemotesCmd.Flags().Bool("password", false, "Update password interactively (prompts)")
	PatchRemotesCmd.Flags().String("password-method", "ARGON2ID", "Password hashing method")
	PatchRemotesCmd.Flags().Bool("password-stdin", false, "Read new password from stdin")
	PatchRemotesCmd.Flags().Int("port", 0, "New SMTP port")
	PatchRemotesCmd.Flags().Bool("no-password", false, "Remove password")
}
