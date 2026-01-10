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

func listRelayedRecipients(options db.RecipientsRelayedListOptions) ([]db.RecipientRelayed, error) {
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

	recipients, err := db.RecipientsRelayed(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("Failed to get relayed recipients", err)
		return nil, err
	}
	return recipients, nil
}

var ListRecipientsRelayedCmd = &cobra.Command{
	Use:     "relayed-recipients [flags] [<domain>...]",
	Aliases: []string{"recipient-relayed", "relayed-recipients", "relayed-recipient", "relayed"},
	Short:   "List relayed recipients",
	Long:    "List relayed recipients.\nIf domains are provided, only recipients for these domains are listed.",
	Args:    cobra.ArbitraryArgs,
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

		options := db.RecipientsRelayedListOptions{
			FilterDomains:  filterDomains,
			IncludeDeleted: flagDeleted,
			IncludeAll:     flagAll,
		}

		recipients, err := listRelayedRecipients(options)
		if err != nil {
			return nil
		}

		if flagJSON {
			out, err := json.Marshal(recipients)
			if err != nil {
				utils.PrintErrorWithMessage("Failed to marshal recipients to JSON", err)
				return nil
			}
			fmt.Println(string(out))
			return nil
		}

		if len(recipients) == 0 {
			fmt.Println("No relayed recipients found")
			return nil
		}

		headers := []string{"Domain", "Name", "Enabled"}
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

		for _, r := range recipients {
			row := []string{
				r.DomainFQDN,
				r.Name,
				utils.MaybeEnabledTableStyle.Render(r.Enabled, r.DomainEnabled),
			}
			if flagVerbose {
				row = append(row,
					utils.MaybeTimeStyle.Render(r.CreatedAt),
					utils.MaybeTimeStyle.Render(r.UpdatedAt),
				)
			}
			if flagDeleted || flagAll {
				row = append(row,
					utils.MaybeTimeStyle.Render(r.DeletedAt),
				)
			}

			t.Row(row...)
		}

		fmt.Println(t.Render())
		return nil
	},
}
