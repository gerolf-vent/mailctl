package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var RestoreAliasesCmd = &cobra.Command{
	Use:     "aliases <email> [email...]",
	Aliases: []string{"alias"},
	Short:   "Restores soft-deleted aliases",
	Long:    "Restores soft-deleted aliases.\nEmails must be in the format \"name@example.com\".",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.Aliases(tx).Restore(item)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to restore alias",
			SuccessMessage: "Successfully restored alias",
		}

		runner.Run()
		return nil
	},
}
