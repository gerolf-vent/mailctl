package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var EnableDomainsCmd = &cobra.Command{
	Use:     "domains <fqdn> [fqdn...]",
	Aliases: []string{"domain"},
	Short:   "Enables domains",
	Long:    "Enables domains.\nFQDNs must be valid domain names.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		argDomains := ParseDomainFQDNArgs(args)
		if len(argDomains) != len(args) {
			return fmt.Errorf("invalid domain arguments")
		}

		enabled := true
		options := db.DomainsPatchOptions{
			Enabled: &enabled,
		}

		runner := db.TxForEachRunner[string]{
			Items: argDomains,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Domains(tx).Patch(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to enable domain",
			SuccessMessage: "Successfully enabled domain",
		}

		runner.Run()
		return nil
	},
}
