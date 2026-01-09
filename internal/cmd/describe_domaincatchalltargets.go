package cmd

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

// Describe prints a detailed view for a catchall address (*@domain).
// Returns (true, nil) when the catchall was found and printed, (false, nil) when
// the catchall was not found, or (true, err) on error.
func describeDomainscatchalltargets(r sq.BaseRunner, domain string) (bool, error) {
	options := db.DomainsCatchallTargetsListOptions{
		FilterDomains: []string{domain},
		IncludeAll:    true,
	}

	targets, err := db.DomainsCatchallTargets(r).List(options)
	if err != nil {
		return false, err
	}
	if len(targets) == 0 {
		return false, nil
	}

	// Since there is no single "catchall" object, we just show the targets.
	// But we can infer the domain status from the first target (it has DomainEnabled).
	firstTarget := targets[0]

	title := "Catch-All Address"

	var statusStr string
	if firstTarget.DomainEnabled {
		statusStr = utils.GreenStyle.Bold(true).Render("Operational")
	} else {
		statusStr = utils.YellowStyle.Bold(true).Render("Disabled")
	}

	// Targets Table
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
		Row([]string{"Targets", "Fwd.", "Fallback"}...).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false)

	for _, target := range targets {
		addrStr := target.TargetEmail
		if target.DeletedAt != nil {
			addrStr += " " + utils.RedStyle.Render("(deleted)")
		}

		targetsT.Row(
			addrStr,
			utils.MaybeEnabledStyle.Render(target.ForwardingToTargetEnabled),
			utils.MaybeEnabledStyle.Render(target.FallbackOnly),
		)
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
	t.Row(lipgloss.NewStyle().Bold(true).Render("Address: ") + utils.MaybeWildcardNameStyle.Render(nil) + "@" + domain)
	t.Row(targetsT.Render())
	fmt.Println(t.Render())
	return true, nil
}
