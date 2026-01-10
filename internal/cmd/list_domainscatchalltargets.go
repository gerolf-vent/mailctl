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

func listCatchallTargets(options db.DomainsCatchallTargetsListOptions) ([]db.DomainCatchallTarget, error) {
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

	targets, err := db.DomainsCatchallTargets(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("failed to list domain catchall targets", err)
		return nil, err
	}
	return targets, nil
}

var ListDomainCatchallTargetsCmd = &cobra.Command{
	Use:     "catchall-targets [flags] [<domain>...]",
	Aliases: []string{"catchall-target", "catchalls", "catchall"},
	Short:   "List catch-all targets.",
	Long:    "List catch-all targets. If domains are provided, only catch-all targets for these domains are listed.",
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDeleted, _ := cmd.Flags().GetBool("deleted")
		flagAll, _ := cmd.Flags().GetBool("all")
		flagJSON, _ := cmd.Flags().GetBool("json")
		flagVerbose, _ := cmd.Flags().GetBool("verbose")

		argDomains := args

		options := db.DomainsCatchallTargetsListOptions{
			FilterDomains:  argDomains,
			IncludeDeleted: flagDeleted,
			IncludeAll:     flagAll,
		}

		targets, err := listCatchallTargets(options)
		if err != nil {
			return nil
		}

		if len(targets) == 0 {
			fmt.Println("No catchall targets found for this domain")
			return nil
		}

		if flagJSON {
			out, err := json.MarshalIndent(targets, "", "  ")
			if err != nil {
				utils.PrintErrorWithMessage("Failed to marshal targets", err)
				return nil
			}
			fmt.Println(string(out))
			return nil
		}

		headers := []string{"Domain", "Target", "Forward", "Fallback Only"}
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
				if col == 2 || col == 3 {
					cellStyle = cellStyle.Align(lipgloss.Center)
				}
				return cellStyle
			}).
			Headers(headers...)

		for _, tgt := range targets {
			row := []string{
				tgt.DomainFQDN,
				tgt.TargetEmail,
				utils.MaybeEnabledTableStyle.Render(tgt.ForwardingToTargetEnabled),
				utils.MaybeEnabledTableStyle.Render(tgt.FallbackOnly),
			}
			if flagVerbose {
				row = append(row, utils.MaybeTimeStyle.Render(tgt.CreatedAt), utils.MaybeTimeStyle.Render(tgt.UpdatedAt))
			}
			if flagAll || flagDeleted {
				row = append(row, utils.MaybeTimeStyle.Render(tgt.DeletedAt))
			}
			t.Row(row...)
		}

		fmt.Println(t)
		return nil
	},
}
