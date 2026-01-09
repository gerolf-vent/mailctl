package cmd

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

// Describe prints a detailed view for a single relayed recipient (name@domain).
// Returns (true, nil) when the recipient was found and printed, (false, nil)
// when the recipient was not found, or (true, err) on error.
func describeRecipientsrelayed(r sq.BaseRunner, email utils.EmailAddress) (bool, error) {
	options := db.RecipientsRelayedListOptions{
		ByEmail:    &email,
		IncludeAll: true,
	}

	recipients, err := db.RecipientsRelayed(r).List(options)
	if err != nil {
		return false, err
	}
	if len(recipients) == 0 {
		return false, nil
	}

	recipient := recipients[0]

	title := "Relayed Recipient"

	var statusStr string
	if recipient.DeletedAt != nil {
		statusStr = utils.RedStyle.Bold(true).Render("Deleted")
	} else if recipient.DomainEnabled && recipient.Enabled {
		statusStr = utils.GreenStyle.Bold(true).Render("Operational")
	} else {
		statusStr = utils.YellowStyle.Bold(true).Render("Disabled")
	}

	// Properties
	propT := table.New().
		Rows([][]string{
			{"Address:", email.String()},
			{"Enabled:", utils.MaybeEnabledStyle.Render(recipient.Enabled, recipient.DomainEnabled)},
		}...).
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
		result, err := db.PostfixRelayRecipientMaps(r, email)
		funcsT.Row("postfix.relay_recipient_maps", utils.TestFunctionResultStyle.Render(result, err))
	}
	{
		result, err := db.PostfixSMTPDSenderLoginMapsMailboxes(r, email, 100)
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

	// Output final table
	t := table.New().
		BorderStyle(utils.BlackStyle).
		BorderRow(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return utils.TableRowStyle
		})
	t.Row(lipgloss.NewStyle().Bold(true).Render(title))
	t.Row(lipgloss.NewStyle().Bold(true).Render("Status: ") + statusStr)
	t = t.Row(propT.Render())
	t = t.Row(lipgloss.NewStyle().Bold(true).Render("Meta") + "\n\n" + RenderMetaSection(recipient.CreatedAt, recipient.UpdatedAt, recipient.DeletedAt))
	t = t.Row(lipgloss.NewStyle().Bold(true).Render("Functions") + "\n\n" + funcsT.Render())
	fmt.Println(t.Render())
	return true, nil
}
