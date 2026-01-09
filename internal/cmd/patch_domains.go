package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/spf13/cobra"
)

var PatchDomainsCmd = &cobra.Command{
	Use:     "domains <fqdn> [fqdn...]",
	Aliases: []string{"domain"},
	Short:   "Update an existing domain",
	Long:    "Updates properties of an existing domain.\nFQDNs must be valid domain names.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagEnabled, _ := cmd.Flags().GetBool("enabled")
		flagTransport, _ := cmd.Flags().GetString("transport")
		flagTargetDomain, _ := cmd.Flags().GetString("target-domain")

		// Check if at least one flag was changed
		if !cmd.Flags().Changed("enabled") && !cmd.Flags().Changed("transport") && !cmd.Flags().Changed("target-domain") {
			return fmt.Errorf("no changes specified")
		}

		argDomains := ParseDomainFQDNArgs(args)
		if len(argDomains) != len(args) {
			return fmt.Errorf("invalid domain arguments")
		}

		var options db.DomainsPatchOptions
		if cmd.Flags().Changed("enabled") {
			options.Enabled = &flagEnabled
		}
		if cmd.Flags().Changed("transport") {
			options.TransportName = &flagTransport
		}
		if cmd.Flags().Changed("target-domain") {
			options.TargetDomainFQDN = &flagTargetDomain
		}

		runner := db.TxForEachRunner[string]{
			Items: argDomains,
			Exec: func(tx *sql.Tx, item string) error {
				return db.Domains(tx).Patch(item, options)
			},
			ItemString:     func(item string) string { return item },
			FailureMessage: "failed to patch domain",
			SuccessMessage: "Successfully patched domain",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchDomainsCmd.Flags().BoolP("enabled", "e", false, "Enable or disable the domain")
	PatchDomainsCmd.Flags().String("transport", "", "New transport name (only for managed/relayed domains)")
	PatchDomainsCmd.Flags().String("target-domain", "", "New target domain FQDN (only for canonical domains)")
}
