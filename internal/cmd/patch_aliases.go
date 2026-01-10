package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var PatchAliasesCmd = &cobra.Command{
	Use:     "aliases [flags] <email> [<email>...]",
	Aliases: []string{"alias"},
	Short:   "Updates an existing alias",
	Long:    "Updates specified properties of an existing alias.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagEnabled, _ := cmd.Flags().GetBool("enabled")

		if !cmd.Flags().Changed("enabled") {
			return fmt.Errorf("no changes specified")
		}

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		options := db.AliasesPatchOptions{}

		if cmd.Flags().Changed("enabled") {
			options.Enabled = &flagEnabled
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.Aliases(tx).Patch(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to patch alias",
			SuccessMessage: "Successfully patched alias",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchAliasesCmd.Flags().BoolP("enabled", "e", false, "Enable or disable the alias")
}
