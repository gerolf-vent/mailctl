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

func listGrants(options db.RemotesSendGrantsListOptions) ([]db.RemoteSendGrant, error) {
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

	grants, err := db.RemotesSendGrants(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("failed to list send grants", err)
		return nil, err
	}
	return grants, nil
}

var ListRemoteSendGrantsCmd = &cobra.Command{
	Use:   "send-grants [name...]",
	Short: "List send grants for remotes",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDeleted, _ := cmd.Flags().GetBool("deleted")
		flagAll, _ := cmd.Flags().GetBool("all")
		flagJSON, _ := cmd.Flags().GetBool("json")
		flagVerbose, _ := cmd.Flags().GetBool("verbose")

		if flagDeleted && flagAll {
			return fmt.Errorf("cannot use --deleted and --all flags together")
		}

		var filterNames []string
		if len(args) > 0 {
			filterNames = args
		}

		grants, err := listGrants(db.RemotesSendGrantsListOptions{
			FilterRemoteNames: filterNames,
			IncludeDeleted:    flagDeleted,
			IncludeAll:        flagAll,
		})
		if err != nil {
			return nil
		}

		if flagJSON {
			out, err := json.Marshal(grants)
			if err != nil {
				utils.PrintErrorWithMessage("Failed to marshal send grants to JSON", err)
				return nil
			}
			fmt.Println(string(out))
			return nil
		}

		if len(grants) == 0 {
			fmt.Println("No send grants found")
			return nil
		}

		headers := []string{"Remote", "Domain", "Name"}
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
				default:
					return cellStyle.Align(lipgloss.Left)
				}
			}).
			Headers(headers...)

		for _, g := range grants {
			row := []string{
				g.RemoteName,
				g.DomainFQDN,
				utils.SQLLikeStyle.Render(g.Name),
			}

			if flagVerbose {
				row = append(row,
					utils.MaybeTimeStyle.Render(g.CreatedAt),
					utils.MaybeTimeStyle.Render(g.UpdatedAt),
				)
			}
			if flagDeleted || flagAll {
				row = append(row,
					utils.MaybeTimeStyle.Render(g.DeletedAt),
				)
			}

			t.Row(row...)
		}

		fmt.Println(t.Render())
		return nil
	},
}
