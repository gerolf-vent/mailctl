package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var CreateAliasesCmd = &cobra.Command{
	Use:     "aliases <email> <target> [target...]",
	Aliases: []string{"alias"},
	Short:   "Create a new alias",
	Long:    "Creates a new alias with the specified target addresses.\nEmails must be in the format \"name@example.com\".",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDisabled, _ := cmd.Flags().GetBool("disabled")

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		options := db.AliasesCreateOptions{
			Disabled: flagDisabled,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.Aliases(tx).Create(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to create alias",
			SuccessMessage: "Successfully created alias",
		}

		runner.Run()
		return nil
	},
}

func init() {
	CreateAliasesCmd.Flags().BoolP("disabled", "d", false, "Create the alias in disabled state")
}
