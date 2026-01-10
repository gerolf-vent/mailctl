package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

func listTransports(options db.TransportsListOptions) ([]db.Transport, error) {
	dbConn, err := db.Connect()
	if err != nil {
		utils.PrintErrorWithMessage("Failed to connect to database", err)
		return nil, err
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			utils.PrintErrorWithMessage("failed to close database connection", err)
		}
	}()

	transports, err := db.Transports(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("Failed to get transports", err)
		return nil, err
	}
	return transports, nil
}

var ListTransportsCmd = &cobra.Command{
	Use:   "transports [flags]",
	Short: "List transports",
	Long:  "List all transports. Use --deleted to show only deleted transports, --all to show both active and deleted.",
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDeleted, _ := cmd.Flags().GetBool("deleted")
		flagAll, _ := cmd.Flags().GetBool("all")
		flagJSON, _ := cmd.Flags().GetBool("json")
		flagVerbose, _ := cmd.Flags().GetBool("verbose")

		transports, err := listTransports(db.TransportsListOptions{IncludeDeleted: flagDeleted, IncludeAll: flagAll})
		if err != nil {
			return nil
		}

		if flagJSON {
			encoder := json.NewEncoder(os.Stdout)
			if err := encoder.Encode(transports); err != nil {
				utils.PrintErrorWithMessage("Failed to encode JSON", err)
				return nil
			}
			return nil
		}

		// Create lipgloss table
		headers := []string{"Name", "Method", "Host", "Port", "MX Lookup"}
		if flagVerbose {
			headers = append(headers, "Created", "Last Updated")
		}
		if flagDeleted || flagAll {
			headers = append(headers, "Deleted")
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(utils.BlackStyle).
			Headers(headers...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return utils.TableHeaderStyle
				}
				cellStyle := utils.TableRowStyle
				switch col {
				case 3: // Port
					return cellStyle.Align(lipgloss.Right)
				case 4: // MX Lookup
					return cellStyle.Align(lipgloss.Center)
				}
				return cellStyle
			})

		mxLookupStyle := utils.MaybeEnabledTableStyle
		mxLookupStyle.TrueStyle = utils.BlueStyle.Bold(true)
		mxLookupStyle.FalseStyle = utils.BlackStyle.Bold(true)

		for _, transport := range transports {
			row := []string{
				transport.Name,
				transport.Method,
				transport.Host,
				utils.MaybeEmptyStyle.Render(transport.Port),
				mxLookupStyle.Render(transport.MXLookup),
			}

			if flagVerbose {
				row = append(row,
					utils.MaybeTimeStyle.Render(transport.CreatedAt),
					utils.MaybeTimeStyle.Render(transport.UpdatedAt),
				)
			}

			if flagAll || flagDeleted {
				row = append(row,
					utils.MaybeTimeStyle.Render(transport.DeletedAt),
				)
			}

			t.Row(row...)
		}

		fmt.Println(t.Render())
		return nil
	},
}
