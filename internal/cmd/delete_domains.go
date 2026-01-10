package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var DeleteDomainsCmd = &cobra.Command{
	Use:     "domains [flags] <fqdn> [<fqdn>...]",
	Aliases: []string{"domain"},
	Short:   "Deletes domains",
	Long:    "Deletes domains. By default performs a soft delete. Use --permanent for hard delete.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPermanent, _ := cmd.Flags().GetBool("permanent")
		flagForce, _ := cmd.Flags().GetBool("force")

		if flagPermanent && flagForce {
			return fmt.Errorf("cannot use --permanent and --force flags together")
		}

		argDomains := ParseDomainFQDNArgs(args)
		if len(argDomains) != len(args) {
			return fmt.Errorf("invalid domain arguments")
		}

		options := db.DeleteOptions{
			Permanent: flagPermanent,
			Force:     flagForce,
		}

		runner := db.TxForEachRunner[string]{
			Items: argDomains,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Domains(tx).Delete(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to delete domain",
			SuccessMessage: "Successfully deleted domain",
		}

		if flagPermanent {
			runner.SuccessMessage += " permanently"
		}

		runner.Run()
		return nil
	},
}
