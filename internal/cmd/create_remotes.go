package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var CreateRemotesCmd = &cobra.Command{
	Use:     "remotes <hostname> [hostname...]",
	Aliases: []string{"remote"},
	Short:   "Create a new remote",
	Long:    "Creates a new remote SMTP relay server configuration.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPassword, _ := cmd.Flags().GetBool("password")
		flagPasswordStdin, _ := cmd.Flags().GetBool("password-stdin")
		flagPasswordMethod, _ := cmd.Flags().GetString("password-method")
		flagDisabled, _ := cmd.Flags().GetBool("disabled")

		if flagPassword && flagPasswordStdin {
			return fmt.Errorf("cannot use both --password and --password-stdin")
		}

		if len(args) > 1 && (flagPassword || flagPasswordStdin) {
			return fmt.Errorf("cannot set password while creating multiple mailboxes")
		}

		options := db.RemotesCreateOptions{
			Enabled: !flagDisabled,
		}

		if flagPassword || flagPasswordStdin {
			passwordHash, err := ReadPasswordHashed(flagPasswordMethod, flagPasswordStdin)
			if err != nil {
				utils.PrintErrorWithMessage("failed to read password", err)
				return nil
			}
			options.PasswordHash = sql.NullString{
				String: passwordHash,
				Valid:  true,
			}
		}

		runner := db.TxForEachRunner[string]{
			Items: args,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Remotes(tx).Create(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to create remote",
			SuccessMessage: "Successfully created remote",
		}

		runner.Run()
		return nil
	},
}

func init() {
	CreateRemotesCmd.Flags().String("username", "", "SMTP username (required)")
	CreateRemotesCmd.MarkFlagRequired("username")
	CreateRemotesCmd.Flags().Bool("password", false, "Set password interactively (prompts)")
	CreateRemotesCmd.Flags().String("password-method", "ARGON2ID", "Password hashing method (default: \"ARGON2ID\")")
	CreateRemotesCmd.Flags().Bool("password-stdin", false, "Read password from stdin")
	CreateRemotesCmd.Flags().Int("port", 25, "SMTP port (default: 25)")
	CreateRemotesCmd.Flags().Bool("enabled", true, "Enable or disable the remote (default: true)")
}
