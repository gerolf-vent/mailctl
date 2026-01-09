package cmd

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

// Describe prints a detailed view for a single alias (name@domain).
// Returns (true, nil) when the alias was found and printed, (false, nil) when
// the alias was not found, or (true, err) on error.
func describeAliases(r sq.BaseRunner, email utils.EmailAddress) (bool, error) {
	options := db.AliasesListOptions{
		ByEmail:    &email,
		IncludeAll: true,
	}

	aliases, err := db.Aliases(r).List(options)
	if err != nil {
		return false, nil
	}
	if len(aliases) == 0 {
		return false, nil // No alias found with that email
	}

	alias := aliases[0]

	// Determine status
	var statusStr string
	if alias.DeletedAt != nil {
		statusStr = utils.RedStyle.Bold(true).Render("Deleted")
	} else if alias.Enabled && alias.DomainEnabled {
		statusStr = utils.GreenStyle.Bold(true).Render("Operational")
	} else {
		statusStr = utils.YellowStyle.Bold(true).Render("Disabled")
	}

	// Properties
	propT := table.New().
		Rows([][]string{
			{"Address:", email.String()},
			{"Enabled:", utils.MaybeEnabledStyle.Render(alias.Enabled, alias.DomainEnabled)},
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

	// Targets
	hasTargets := false
	targetsT := table.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			cellStyle := utils.TableRowStyle
			if row == 0 {
				cellStyle = cellStyle.Bold(true).PaddingBottom(1)
			}
			if col == 0 {
				cellStyle = cellStyle.PaddingLeft(0).PaddingRight(3)
			}
			if col > 0 {
				cellStyle = cellStyle.Align(lipgloss.Center)
			}
			return cellStyle
		}).
		Row([]string{"Targets", "Frgn.", "Fwd.", "Snd."}...).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false)

	{
		targetOptions := db.AliasesTargetsListOptions{
			FilterAliasEmails: []utils.EmailAddress{email},
			IncludeAll:        true,
		}
		targets, err := db.AliasesTargets(r).List(targetOptions)
		if err != nil {
			utils.PrintErrorWithMessage("failed to query alias targets", err)
			return true, err
		}

		hasTargets = len(targets) > 0

		for _, target := range targets {
			addrStr := target.TargetEmail
			if target.DeletedAt != nil {
				addrStr = addrStr + " " + utils.RedStyle.Render("(deleted)")
			}

			foreignStyle := utils.MaybeEnabledStyle
			foreignStyle.TrueStyle = utils.BlueStyle
			foreignStyle.FalseStyle = utils.BlackStyle

			targetsT = targetsT.Row([]string{
				addrStr,
				foreignStyle.Render(target.IsForeign),
				utils.MaybeEmptyStyle.Render(target.ForwardingToTargetEnabled),
				utils.MaybeEmptyStyle.Render(target.SendingFromTargetEnabled),
			}...)
		}
	}

	// Functions table
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
		result, err := db.PostfixVirtualAliasMaps(r, email, 1000)
		funcsT.Row("postfix.virtual_alias_maps", utils.TestFunctionResultStyle.Render(result, err))

		result, err = db.PostfixSMTPDSenderLoginMapsMailboxes(r, email, 1000)
		funcsT.Row("postfix.smtpd_sender_login_maps_mailboxes", utils.TestFunctionResultStyle.Render(result, err))

		result, err = db.PostfixSMTPDSenderLoginMapsRemotes(r, email)
		funcsT.Row("postfix.smtpd_sender_login_maps_remotes", utils.TestFunctionResultStyle.Render(result, err))
	}

	// Output final table
	headerStyle := lipgloss.NewStyle().Bold(true)
	t := table.New().
		BorderStyle(utils.BlackStyle).
		BorderRow(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return utils.TableRowStyle
		})
	t.Row(headerStyle.Render("Alias"))
	t.Row(headerStyle.Render("Status: ") + statusStr)
	t.Row(headerStyle.Render("Properties") + "\n\n" + propT.Render())
	t.Row(headerStyle.Render("Meta") + "\n\n" + RenderMetaSection(alias.CreatedAt, alias.UpdatedAt, alias.DeletedAt))
	if hasTargets {
		t.Row(targetsT.Render())
	} else {
		t.Row(headerStyle.Render("Targets") + "\n\n" + utils.BlackStyle.Render("No targets configured."))
	}
	t.Row(headerStyle.Render("Functions") + "\n\n" + funcsT.Render())
	fmt.Println(t.Render())
	return true, nil
}
