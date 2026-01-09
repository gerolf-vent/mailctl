package cmd

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

// Describe prints a detailed view for a single mailbox (name@domain).
// Returns (true, nil) when the mailbox was found and printed, (false, nil)
// when the mailbox was not found, or (true, err) on error.
func describeMailboxes(r sq.BaseRunner, email utils.EmailAddress) (bool, error) {
	options := db.MailboxesListOptions{
		ByEmail:    &email,
		IncludeAll: true,
	}

	mailboxes, err := db.Mailboxes(r).List(options)
	if err != nil {
		return false, err
	}
	if len(mailboxes) == 0 {
		return false, nil
	}

	mailbox := mailboxes[0]

	// Determine status
	var statusStr string
	if mailbox.DeletedAt != nil {
		statusStr = utils.RedStyle.Bold(true).Render("Deleted")
	} else if mailbox.DomainEnabled && (mailbox.LoginEnabled || mailbox.ReceivingEnabled || mailbox.SendingEnabled) {
		statusStr = utils.GreenStyle.Bold(true).Render("Operational")
	} else {
		statusStr = utils.YellowStyle.Bold(true).Render("Disabled")
	}

	title := "Mailbox"

	// Properties
	propT := table.New().
		Rows([][]string{
			{"Address:", email.String()},
			{"Login:", utils.MaybeEnabledStyle.Render(mailbox.LoginEnabled, mailbox.DomainEnabled)},
			{"Receiving:", utils.MaybeEnabledStyle.Render(mailbox.ReceivingEnabled, mailbox.DomainEnabled)},
			{"Sending:", utils.MaybeEnabledStyle.Render(mailbox.SendingEnabled, mailbox.DomainEnabled)},
			{"Password:", utils.MaybePasswordStyle.Render(mailbox.PasswordSet)},
			{"Storage Quota:", utils.MaybeQuotaStyle.Render(mailbox.StorageQuota, 1024*1024)},
			{"Transport:", utils.MaybeIDSuffixStyle.Render(mailbox.Transport, mailbox.TransportName)},
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
		result, err := db.PostfixVirtualMailboxMaps(r, email)
		funcsT.Row("postfix.virtual_mailbox_maps", utils.TestFunctionResultStyle.Render(result, err))

		results, err := db.PostfixSMTPDSenderLoginMapsMailboxes(r, email, 100)
		funcsT.Row("postfix.smtpd_sender_login_maps_mailboxes", utils.TestFunctionResultListStyle.Render(results, err))

		results, err = db.PostfixSMTPDSenderLoginMapsRemotes(r, email)
		funcsT.Row("postfix.smtpd_sender_login_maps_remotes", utils.TestFunctionResultListStyle.Render(results, err))

		result, err = db.PostfixTransportMaps(r, email)
		funcsT.Row("postfix.transport_maps", utils.TestFunctionResultStyle.Render(result, err))
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
	t.Row(headerStyle.Render("Status: ") + statusStr)
	t = t.Row(headerStyle.Render("Properties") + "\n\n" + propT.Render())
	t = t.Row(headerStyle.Render("Meta") + "\n\n" + RenderMetaSection(mailbox.CreatedAt, mailbox.UpdatedAt, mailbox.DeletedAt))
	t = t.Row(headerStyle.Render("Functions") + "\n\n" + funcsT.Render())
	fmt.Println(t.Render())
	return true, nil
}
