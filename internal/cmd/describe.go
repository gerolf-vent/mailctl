package cmd

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var DescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe mail system objects",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbConn, err := db.Connect()
		if err != nil {
			utils.PrintErrorWithMessage("failed to connect to database", err)
			return nil
		}
		defer func() {
			if err := dbConn.Close(); err != nil {
				utils.PrintErrorWithMessage("failed to close database connection", err)
			}
		}()

		if strings.Contains(args[0], "@") {
			// The arg might be a wildcard or specific email address

			emailOrWildcard, err := utils.ParseEmailAddressOrWildcard(args[0])
			if err != nil {
				utils.PrintError(fmt.Errorf("failed to parse email: %w", err))
				return nil
			}

			if emailOrWildcard.IsWildcard() {
				// A wildcard email address can only be a catchall address
				if described, err := describeDomainscatchalltargets(dbConn, emailOrWildcard.DomainFQDN); described || err != nil {
					if err != nil {
						utils.PrintError(err)
					}
					return nil
				}
			} else {
				// A specific email address can be an alias, a mailbox or a relayed recipient

				email := utils.EmailAddress{
					DomainFQDN: emailOrWildcard.DomainFQDN,
					LocalPart:  *emailOrWildcard.LocalPart,
				}

				// Try aliases first
				if described, err := describeAliases(dbConn, email); described || err != nil {
					if err != nil {
						utils.PrintError(err)
					}
					return nil
				}

				// Try canonical addresses next
				if described, err := DescribeCanonicalAddress(dbConn, email); described || err != nil {
					if err != nil {
						utils.PrintError(err)
					}
					return nil
				}

				// Try mailboxes next
				if described, err := describeMailboxes(dbConn, email); described || err != nil {
					if err != nil {
						utils.PrintError(err)
					}
					return nil
				}

				// Try relayed recipients last
				if described, err := describeRecipientsrelayed(dbConn, email); described || err != nil {
					if err != nil {
						utils.PrintError(err)
					}
					return nil
				}
			}

			// Fallback to unknown email
			describeUnknownEmail(dbConn, emailOrWildcard)
			return nil
		} else {
			// A plain string might be a fqdn (for domains) or a a name of a transport or remote

			// Try domains first
			if described, err := describeDomains(dbConn, args[0]); described || err != nil {
				if err != nil {
					utils.PrintError(err)
				}
				return nil
			}

			// Try transports next
			if described, err := describeTransports(dbConn, args[0]); described || err != nil {
				if err != nil {
					utils.PrintError(err)
				}
				return nil
			}

			// Try remotes last
			if described, err := describeRemotes(dbConn, args[0]); described || err != nil {
				if err != nil {
					utils.PrintError(err)
				}
				return nil
			}

			// Fallback to unknown
			describeUnknown(dbConn, args[0])
			return nil
		}
	},
}

func describeUnknownEmail(r sq.BaseRunner, emailOrWildcard utils.EmailAddressOrWildcard) {
	title := "Unknown Address"

	// Functions
	funcsT := table.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			cellStyle := utils.TableRowStyle
			if col == 0 {
				return cellStyle.PaddingLeft(0).PaddingRight(3)
			}
			return cellStyle
		}).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false)

	if !emailOrWildcard.IsWildcard() {
		email := utils.EmailAddress{
			DomainFQDN: emailOrWildcard.DomainFQDN,
			LocalPart:  *emailOrWildcard.LocalPart,
		}

		{
			result, err := db.PostfixVirtualMailboxMaps(r, email)
			funcsT.Row("postfix.virtual_mailbox_maps", utils.TestFunctionResultStyle.Render(result, err))
		}
		{
			result, err := db.PostfixRelayRecipientMaps(r, email)
			funcsT.Row("postfix.relay_recipient_maps", utils.TestFunctionResultStyle.Render(result, err))
		}
		{
			result, err := db.PostfixVirtualAliasMaps(r, email, 50)
			funcsT.Row("postfix.virtual_alias_maps", utils.TestFunctionResultListStyle.Render(result, err))
		}
		{
			result, err := db.PostfixSMTPDSenderLoginMapsMailboxes(r, email, 50)
			funcsT.Row("postfix.smtpd_sender_login_maps_mailboxes", utils.TestFunctionResultListStyle.Render(result, err))
		}
		{
			result, err := db.PostfixSMTPDSenderLoginMapsRemotes(r, email)
			funcsT.Row("postfix.smtpd_sender_login_maps_remotes", utils.TestFunctionResultListStyle.Render(result, err))
		}
		{
			result, err := db.PostfixTransportMaps(r, email)
			funcsT.Row("postfix.transport_maps", utils.TestFunctionResultStyle.Render(result, err))
		}
	}

	// Output final table
	headerStyle := lipgloss.NewStyle().Bold(true)
	t := table.New().
		BorderStyle(utils.BlackStyle).
		BorderRow(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return utils.TableRowStyle
		})
	t.Row(headerStyle.Render(title))
	t.Row(headerStyle.Render("Status: ") + utils.BlueStyle.Bold(true).Render("Not Found"))
	t.Row(headerStyle.Render("Address: ") + utils.MaybeWildcardNameStyle.Render(emailOrWildcard.LocalPart) + "@" + emailOrWildcard.DomainFQDN)
	if !emailOrWildcard.IsWildcard() {
		t.Row(headerStyle.Render("Functions") + "\n\n" + funcsT.Render())
	}
	fmt.Println(t.Render())
}

func describeUnknown(r sq.BaseRunner, arg string) {
	title := "Unknown Object"

	// Functions
	funcsT := table.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			cellStyle := utils.TableRowStyle
			if col == 0 {
				return cellStyle.PaddingLeft(0).PaddingRight(3)
			}
			return cellStyle
		}).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false)

	{
		result, err := db.PostfixVirtualMailboxDomains(r, arg)
		funcsT.Row("postfix.virtual_mailbox_domains", utils.TestFunctionResultStyle.Render(result, err))
	}
	{
		result, err := db.PostfixRelayDomains(r, arg)
		funcsT.Row("postfix.relay_domains", utils.TestFunctionResultStyle.Render(result, err))
	}
	{
		result, err := db.PostfixVirtualAliasDomains(r, arg)
		funcsT.Row("postfix.virtual_alias_domains", utils.TestFunctionResultStyle.Render(result, err))
	}

	// Output final table
	headerStyle := lipgloss.NewStyle().Bold(true)
	t := table.New().
		BorderStyle(utils.BlackStyle).
		BorderRow(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return utils.TableRowStyle
		})
	t.Row(headerStyle.Render(title))
	t.Row(headerStyle.Render("Status: ") + utils.BlueStyle.Bold(true).Render("Not Found"))
	t.Row(headerStyle.Render("Argument: ") + arg)
	t = t.Row(headerStyle.Render("Functions") + "\n\n" + funcsT.Render())
	fmt.Println(t.Render())
}
