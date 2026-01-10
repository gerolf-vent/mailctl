package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var CreateRemotesCmd = &cobra.Command{
	Use:     "remotes <hostname> [<hostname>...]",
	Aliases: []string{"remote"},
	Short:   "Creates new remotes",
	Long:    "Creates new remotes.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPassword, _ := cmd.Flags().GetBool("password")
		flagPasswordStdin, _ := cmd.Flags().GetBool("password-stdin")
		flagPasswordMethod, _ := cmd.Flags().GetString("password-method")
		flagPasswordHashOptions, _ := cmd.Flags().GetString("password-hash-options")
		flagDisabled, _ := cmd.Flags().GetBool("disabled")

		if flagPassword && flagPasswordStdin {
			return fmt.Errorf("cannot use both --password and --password-stdin")
		}

		if len(args) > 1 && (flagPassword || flagPasswordStdin) {
			return fmt.Errorf("cannot set password while creating multiple remotes")
		}

		options := db.RemotesCreateOptions{
			Enabled: !flagDisabled,
		}

		if flagPassword || flagPasswordStdin {
			passwordHash, err := ReadPasswordHashed(flagPasswordMethod, flagPasswordHashOptions, flagPasswordStdin)
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
	CreateRemotesCmd.Flags().BoolP("password", "p", false, "Set password interactively (prompts)")
	CreateRemotesCmd.Flags().String("password-method", "bcrypt", "Password hashing method (default: \"bcrypt\", options: \"bcrypt\" or \"argon2id\")")
	CreateRemotesCmd.Flags().String("password-hash-options", "", "Password hash options (bcrypt: <cost>; argon2id: m=<number>,t=<number>,p=<number>)")
	CreateRemotesCmd.Flags().Bool("password-stdin", false, "Read password from stdin")
	CreateRemotesCmd.Flags().BoolP("disabled", "d", false, "Create the remote in disabled state")
}
