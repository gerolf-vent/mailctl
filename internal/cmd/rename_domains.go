package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var RenameDomainsCmd = &cobra.Command{
	Use:     "domain <old-fqdn> <new-fqdn>",
	Aliases: []string{"domains"},
	Short:   "Rename a domain",
	Long:    "Rename a domain by changing its FQDN.\nFQDNs must be valid domain names.",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		argDomains := ParseDomainFQDNArgs(args)
		if len(argDomains) != len(args) {
			return fmt.Errorf("invalid domain arguments")
		}

		oldDomain := argDomains[0]
		newDomain := argDomains[1]

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.Domains(tx).Rename(oldDomain, newDomain)
			},
			ItemString:     fmt.Sprintf("%s -> %s", oldDomain, newDomain),
			FailureMessage: "failed to rename domain",
			SuccessMessage: "Successfully renamed domain",
		}

		runner.Run()
		return nil
	},
}
