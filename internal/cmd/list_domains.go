package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

func listDomains(options db.DomainsListOptions) ([]db.Domain, error) {
	dbConn, err := db.Connect()
	if err != nil {
		utils.PrintErrorWithMessage("failed to connect to database", err)
		return nil, err
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			utils.PrintErrorWithMessage("failed to close database connection", err)
		}
	}()

	domains, err := db.Domains(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("failed to get domains", err)
		return nil, err
	}
	return domains, nil
}

var ListDomainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "List domains",
	Long:  "List domains.",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDeleted, _ := cmd.Flags().GetBool("deleted")
		flagAll, _ := cmd.Flags().GetBool("all")
		flagJSON, _ := cmd.Flags().GetBool("json")
		flagVerbose, _ := cmd.Flags().GetBool("verbose")

		if flagDeleted && flagAll {
			return fmt.Errorf("cannot use --deleted and --all flags together")
		}

		options := db.DomainsListOptions{
			IncludeDeleted: flagDeleted,
			IncludeAll:     flagAll,
		}

		domains, err := listDomains(options)
		if err != nil {
			return nil
		}

		if flagJSON {
			jsonData, err := json.Marshal(domains)
			if err != nil {
				utils.PrintErrorWithMessage("failed to marshal domains to JSON", err)
				return nil
			}
			fmt.Println(string(jsonData))
			return nil
		}

		headers := []string{"FQDN", "Type", "Enabled", "Transport / Target Domain"}
		if flagVerbose {
			headers = append(headers, "Created", "Last Updated")
		}
		if flagDeleted || flagAll {
			headers = append(headers, "Deleted")
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(utils.BlackStyle).
			StyleFunc(func(row, col int) lipgloss.Style {
				cellStyle := utils.TableRowStyle
				if row == table.HeaderRow {
					cellStyle = utils.TableHeaderStyle
				}
				switch col {
				case 2: // Enabled
					return cellStyle.Align(lipgloss.Center)
				default:
					return cellStyle.Align(lipgloss.Left)
				}
			}).
			Headers(headers...)

		for _, domain := range domains {
			row := []string{
				domain.FQDN,
				domain.Type,
			}
			switch domain.Type {
			case "canonical":
				row = append(row,
					utils.MaybeEnabledTableStyle.Render(domain.Enabled, domain.TargetDomainEnabled),
					utils.MaybeEmptyStyle.Render(domain.TargetDomainFQDN),
				)
			case "alias":
				row = append(row,
					utils.MaybeEnabledTableStyle.Render(domain.Enabled),
					utils.BlackStyle.Render("-"),
				)
			default:
				row = append(row,
					utils.MaybeEnabledTableStyle.Render(domain.Enabled),
					utils.MaybeIDSuffixStyle.Render(domain.Transport, domain.TransportName),
				)
			}
			if flagVerbose {
				row = append(row,
					utils.MaybeTimeStyle.Render(domain.CreatedAt),
					utils.MaybeTimeStyle.Render(domain.UpdatedAt),
				)
			}
			if flagDeleted || flagAll {
				row = append(row,
					utils.MaybeTimeStyle.Render(domain.DeletedAt),
				)
			}

			t.Row(row...)
		}

		fmt.Println(t.Render())
		return nil
	},
}
