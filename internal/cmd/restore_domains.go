package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RestoreDomainsCmd = &cobra.Command{
	Use:     "domains <fqdn> [<fqdn>...]",
	Aliases: []string{"domain"},
	Short:   "Restores soft-deleted domains",
	Long:    "Restores soft-deleted domains.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		argDomains := ParseDomainFQDNArgs(args)
		if len(argDomains) != len(args) {
			return fmt.Errorf("invalid domain arguments")
		}

		runner := db.TxForEachRunner[string]{
			Items: argDomains,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Domains(tx).Restore(item)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to restore domain",
			SuccessMessage: "Successfully restored domain",
		}

		runner.Run()
		return nil
	},
}
