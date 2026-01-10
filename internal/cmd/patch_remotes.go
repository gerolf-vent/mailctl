package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var PatchRemotesCmd = &cobra.Command{
	Use:   "remotes [flags] <hostname> [<hostname>...]",
	Short: "Updates existing remotes",
	Long:  "Updates specified properties for existing remotes.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagEnabled, _ := cmd.Flags().GetBool("enabled")
		flagPassword, _ := cmd.Flags().GetBool("password")
		flagPasswordStdin, _ := cmd.Flags().GetBool("password-stdin")
		flagPasswordMethod, _ := cmd.Flags().GetString("password-method")
		flagPasswordHashOptions, _ := cmd.Flags().GetString("password-hash-options")
		flagPasswordNo, _ := cmd.Flags().GetBool("no-password")

		if (flagPassword || flagPasswordStdin) && flagPasswordNo {
			return fmt.Errorf("cannot use --no-password with --password or --password-stdin")
		}

		if flagPassword && flagPasswordStdin {
			return fmt.Errorf("cannot use both --password and --password-stdin")
		}

		if len(args) > 1 && (flagPassword || flagPasswordStdin) {
			return fmt.Errorf("cannot set password while updating multiple remotes")
		}

		if !flagPassword && !flagPasswordStdin && !flagPasswordNo && !cmd.Flags().Changed("enabled") {
			return fmt.Errorf("no changes specified. Use --password, --no-password, or --enabled flags")
		}

		options := db.RemotesPatchOptions{}

		if cmd.Flags().Changed("enabled") {
			options.Enabled = &flagEnabled
		}

		if flagPassword || flagPasswordStdin {
			passwordHash, err := ReadPasswordHashed(flagPasswordMethod, flagPasswordHashOptions, flagPasswordStdin)
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

		runner := db.TxForEachRunner[string]{
			Items: args,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Remotes(tx).Patch(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to patch remote",
			SuccessMessage: "Successfully patched remote",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchRemotesCmd.Flags().BoolP("enabled", "e", false, "Enable or disable the remote")
	PatchRemotesCmd.Flags().BoolP("password", "p", false, "Update password interactively (prompts)")
	PatchRemotesCmd.Flags().String("password-method", "bcrypt", "Password hashing method (default: \"bcrypt\", options: \"bcrypt\" or \"argon2id\")")
	PatchRemotesCmd.Flags().String("password-hash-options", "", "Password hash options (bcrypt: <cost>; argon2id: m=<number>,t=<number>,p=<number>)")
	PatchRemotesCmd.Flags().Bool("password-stdin", false, "Read new password from stdin")
	PatchRemotesCmd.Flags().Bool("no-password", false, "Remove password")
}
