package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var CreateDomainsCmd = &cobra.Command{
	Use:     "domains <fqdn> [fqdn...]",
	Aliases: []string{"domain"},
	Short:   "Create a new domain",
	Long:    "Creates a new domain of the specified type.\nFQDNs must be valid domain names.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDisabled, _ := cmd.Flags().GetBool("disabled")
		flagType, _ := cmd.Flags().GetString("type")
		flagTransport, _ := cmd.Flags().GetString("transport")
		flagTargetDomain, _ := cmd.Flags().GetString("target-domain")

		domainType := strings.ToLower(flagType)
		switch domainType {
		case "managed", "relayed":
			if flagTransport == "" {
				return fmt.Errorf("transport name is required for %s domains", domainType)
			}
		case "canonical":
			if flagTargetDomain == "" {
				return fmt.Errorf("target domain FQDN is required for canonical domains")
			}
		case "alias":
			// No specific flags required
		default:
			return fmt.Errorf("invalid domain type: %s (must be 'managed', 'relayed', 'alias' or 'canonical')", flagType)
		}

		// Parse and validate all domain FQDNs
		argDomains := ParseDomainFQDNArgs(args)
		if len(argDomains) != len(args) {
			return fmt.Errorf("invalid domain arguments")
		}

		options := db.DomainsCreateOptions{
			DomainType:       domainType,
			TransportName:    flagTransport,
			TargetDomainFQDN: flagTargetDomain,
			Enabled:          !flagDisabled,
		}

		runner := db.TxForEachRunner[string]{
			Items: argDomains,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Domains(tx).Create(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to create domain",
			SuccessMessage: "Successfully created domain",
		}

		runner.Run()
		return nil
	},
}

func init() {
	CreateDomainsCmd.Flags().String("transport", "", "Transport name (required for managed/relayed domains)")
	CreateDomainsCmd.Flags().String("target-domain", "", "Target domain FQDN (required for canonical domains)")
	CreateDomainsCmd.Flags().StringP("type", "t", "managed", "Domain type: 'managed', 'relayed', 'alias' or 'canonical' (default: \"managed\")")
	CreateDomainsCmd.Flags().BoolP("disabled", "d", false, "Create in disabled state")
}
