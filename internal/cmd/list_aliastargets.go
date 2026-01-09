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

func listAliasTargets(options db.AliasesTargetsListOptions) ([]db.AliasTarget, error) {
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

	targets, err := db.AliasesTargets(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("failed to list alias targets", err)
		return nil, err
	}
	return targets, nil
}

var ListAliasTargetsCmd = &cobra.Command{
	Use:     "alias-targets [alias-email...]",
	Aliases: []string{"alias-target", "targets", "target"},
	Short:   "List all targets for an alias",
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDeleted, _ := cmd.Flags().GetBool("deleted")
		flagAll, _ := cmd.Flags().GetBool("all")
		flagJSON, _ := cmd.Flags().GetBool("json")
		flagVerbose, _ := cmd.Flags().GetBool("verbose")

		var filterEmails []utils.EmailAddress
		if len(args) > 0 {
			filterEmails = ParseEmailArgs(args)
			if len(filterEmails) != len(args) {
				return fmt.Errorf("invalid email arguments")
			}
		}

		targets, err := listAliasTargets(db.AliasesTargetsListOptions{
			FilterAliasEmails: filterEmails,
			IncludeDeleted:    flagDeleted,
			IncludeAll:        flagAll,
		})
		if err != nil {
			return nil
		}

		if len(targets) == 0 {
			fmt.Println("No targets found")
			return nil
		}

		// Output as JSON
		if flagJSON {
			output, err := json.MarshalIndent(targets, "", "  ")
			if err != nil {
				utils.PrintErrorWithMessage("Failed to marshal targets", err)
				return nil
			}
			fmt.Println(string(output))
			return nil
		}

		// Display Table
		headers := []string{"Alias", "Target", "Foreign", "Forward", "Send"}
		if flagVerbose {
			headers = append(headers, "Created", "Updated")
		}
		if flagAll || flagDeleted {
			headers = append(headers, "Deleted")
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(utils.BlackStyle).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return utils.TableHeaderStyle
				}
				cellStyle := utils.TableRowStyle
				switch col {
				case 2, 3, 4:
					cellStyle = cellStyle.Align(lipgloss.Center)
				}
				return cellStyle
			}).
			Headers(headers...)

		for _, target := range targets {
			row := []string{
				target.AliasEmail,
				target.TargetEmail,
				utils.MaybeEnabledTableStyle.Render(target.IsForeign),
				utils.MaybeEnabledTableStyle.Render(target.ForwardingToTargetEnabled, target.AliasEnabled, target.DomainEnabled),
				utils.MaybeEnabledTableStyle.Render(target.SendingFromTargetEnabled, target.AliasEnabled, target.DomainEnabled),
			}
			if flagVerbose {
				row = append(row,
					utils.MaybeTimeStyle.Render(target.CreatedAt),
					utils.MaybeTimeStyle.Render(target.UpdatedAt),
				)
			}
			if flagAll || flagDeleted {
				row = append(row,
					utils.MaybeTimeStyle.Render(target.DeletedAt),
				)
			}

			t.Row(row...)
		}

		fmt.Println(t)
		return nil
	},
}
