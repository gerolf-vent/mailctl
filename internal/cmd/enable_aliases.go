package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var EnableAliasesCmd = &cobra.Command{
	Use:     "aliases <email> [<email>...]",
	Aliases: []string{"alias"},
	Short:   "Enables aliases",
	Long:    "Enables aliases.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		enabled := true
		options := db.AliasesPatchOptions{
			Enabled: &enabled,
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.Aliases(tx).Patch(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to enable alias",
			SuccessMessage: "Successfully enabled alias",
		}

		runner.Run()
		return nil
	},
}
