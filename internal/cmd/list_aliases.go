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

func listAliases(options db.AliasesListOptions) ([]db.Alias, error) {
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

	result, err := db.Aliases(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("failed to list aliases", err)
		return nil, err
	}
	return result, nil
}

var ListAliasesCmd = &cobra.Command{
	Use:   "aliases [domain...]",
	Short: "List aliases",
	Long:  "List aliases.\nIf a domains are provided, only aliases for these domains are listed.",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDeleted, _ := cmd.Flags().GetBool("deleted")
		flagAll, _ := cmd.Flags().GetBool("all")
		flagJSON, _ := cmd.Flags().GetBool("json")
		flagVerbose, _ := cmd.Flags().GetBool("verbose")

		if flagDeleted && flagAll {
			return fmt.Errorf("cannot use --deleted and --all flags together")
		}

		filterDomains := ParseDomainFQDNArgs(args)
		if len(filterDomains) != len(args) {
			return fmt.Errorf("invalid domain arguments")
		}

		aliases, err := listAliases(db.AliasesListOptions{
			FilterDomains:  filterDomains,
			IncludeDeleted: flagDeleted,
			IncludeAll:     flagAll,
		})
		if err != nil {
			return nil
		}

		if flagJSON {
			json, err := json.Marshal(aliases)
			if err != nil {
				utils.PrintErrorWithMessage("Failed to marshal aliases to JSON", err)
			}
			fmt.Println(string(json))
			return nil
		}

		headers := []string{"Domain", "Name", "Enabled", "Targets"}
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
				case 3: // Targets
					return cellStyle.Align(lipgloss.Right)
				default:
					return cellStyle.Align(lipgloss.Left)
				}
			}).
			Headers(headers...)

		for _, a := range aliases {
			row := []string{
				a.DomainFQDN,
				utils.MaybeWildcardNameStyle.Render(a.Name),
				utils.MaybeEnabledTableStyle.Render(a.Enabled, a.DomainEnabled),
				utils.MaybeZeroStyle.Render(a.TargetCount),
			}
			if flagVerbose {
				row = append(row,
					utils.MaybeTimeStyle.Render(a.CreatedAt),
					utils.MaybeTimeStyle.Render(a.UpdatedAt),
				)
			}
			if flagDeleted || flagAll {
				row = append(row,
					utils.MaybeTimeStyle.Render(a.DeletedAt),
				)
			}

			t.Row(row...)
		}

		fmt.Println(t.Render())
		return nil
	},
}
