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

func listRemotes(options db.RemotesListOptions) ([]db.Remote, error) {
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

	remotes, err := db.Remotes(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("failed to list remotes", err)
		return nil, err
	}
	return remotes, nil
}

var ListRemotesCmd = &cobra.Command{
	Use:   "remotes [flags]",
	Short: "List remotes",
	Long:  "List remotes.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDeleted, _ := cmd.Flags().GetBool("deleted")
		flagAll, _ := cmd.Flags().GetBool("all")
		flagJSON, _ := cmd.Flags().GetBool("json")
		flagVerbose, _ := cmd.Flags().GetBool("verbose")

		if flagDeleted && flagAll {
			return fmt.Errorf("cannot use --deleted and --all flags together")
		}

		options := db.RemotesListOptions{
			IncludeDeleted: flagDeleted,
			IncludeAll:     flagAll,
		}

		remotes, err := listRemotes(options)
		if err != nil {
			return nil
		}

		if flagJSON {
			out, err := json.Marshal(remotes)
			if err != nil {
				utils.PrintErrorWithMessage("Failed to marshal remotes to JSON", err)
				return nil
			}
			fmt.Println(string(out))
			return nil
		}

		if len(remotes) == 0 {
			fmt.Println("No remotes found")
			return nil
		}

		headers := []string{"Name", "Enabled", "Pwd"}
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
				case 1, 2: // Enabled and Pwd columns
					return cellStyle.Align(lipgloss.Center)
				default:
					return cellStyle.Align(lipgloss.Left)
				}
			}).
			Headers(headers...)

		for _, r := range remotes {
			row := []string{
				r.Name,
				utils.MaybeEnabledTableStyle.Render(r.Enabled),
				utils.MaybePasswordStyle.Render(r.PasswordSet),
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
