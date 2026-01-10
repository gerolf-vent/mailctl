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

func listMailboxes(options db.MailboxesListOptions) ([]db.Mailbox, error) {
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

	mailboxes, err := db.Mailboxes(dbConn).List(options)
	if err != nil {
		utils.PrintErrorWithMessage("failed to list mailboxes", err)
		return nil, err
	}
	return mailboxes, nil
}

var ListMailboxesCmd = &cobra.Command{
	Use:     "mailboxes [flags] [<domain>...]",
	Aliases: []string{"mailbox"},
	Short:   "List mailboxes",
	Long:    "List mailboxes.\nIf domains are provided, only mailboxes for these domains are listed.",
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

		options := db.MailboxesListOptions{
			FilterDomains:  filterDomains,
			IncludeDeleted: flagDeleted,
			IncludeAll:     flagAll,
		}

		mailboxes, err := listMailboxes(options)
		if err != nil {
			return nil
		}

		if flagJSON {
			out, err := json.Marshal(mailboxes)
			if err != nil {
				utils.PrintErrorWithMessage("Failed to marshal mailboxes to JSON", err)
				return nil
			}
			fmt.Println(string(out))
			return nil
		}

		headers := []string{"Domain", "Name", "Login", "Receive", "Send", "Pwd", "Quota", "Transport"}
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
				case 2, 3, 4, 5: // Login, Receive, Send, Password
					return cellStyle.Align(lipgloss.Center)
				case 6: // Quota
					return cellStyle.Align(lipgloss.Right)
				default:
					return cellStyle.Align(lipgloss.Left)
				}
			}).
			Headers(headers...)

		for _, m := range mailboxes {
			row := []string{
				m.DomainFQDN,
				m.Name,
				utils.MaybeEnabledTableStyle.Render(m.LoginEnabled, m.DomainEnabled),
				utils.MaybeEnabledTableStyle.Render(m.ReceivingEnabled, m.DomainEnabled),
				utils.MaybeEnabledTableStyle.Render(m.SendingEnabled, m.DomainEnabled),
				utils.MaybePasswordStyle.Render(m.PasswordSet),
				utils.MaybeQuotaStyle.Render(m.StorageQuota, 1024*1024),
				utils.MaybeIDSuffixStyle.Render(m.Transport, m.TransportName),
			}

			if flagVerbose {
				row = append(row,
					utils.MaybeTimeStyle.Render(m.CreatedAt),
					utils.MaybeTimeStyle.Render(m.UpdatedAt),
				)
			}
			if flagDeleted || flagAll {
				row = append(row,
					utils.MaybeTimeStyle.Render(m.DeletedAt),
				)
			}

			t.Row(row...)
		}

		fmt.Println(t.Render())
		return nil
	},
}
