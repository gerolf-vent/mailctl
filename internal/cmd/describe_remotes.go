package cmd

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

// Describe prints a detailed view for a single remote by name.
// Returns (true, nil) when the remote was found and printed, (false, nil) when
// the remote was not found, or (true, err) on error.
func describeRemotes(r sq.BaseRunner, name string) (bool, error) {
	options := db.RemotesListOptions{
		ByName:     name,
		IncludeAll: true,
	}

	remotes, err := db.Remotes(r).List(options)
	if err != nil {
		return false, err
	}
	if len(remotes) == 0 {
		return false, nil // No remote found with that name
	}

	remote := remotes[0]

	// Determine status
	var statusStr string
	if remote.DeletedAt != nil {
		statusStr = utils.RedStyle.Bold(true).Render("Deleted")
	} else if remote.Enabled {
		statusStr = utils.GreenStyle.Bold(true).Render("Operational")
	} else {
		statusStr = utils.YellowStyle.Bold(true).Render("Disabled")
	}

	// Properties
	propT := table.New().
		Rows([][]string{
			{"Name:", remote.Name},
			{"Enabled:", utils.MaybeEnabledStyle.Render(remote.Enabled)},
			{"Password:", utils.MaybePasswordStyle.Render(remote.PasswordSet)},
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

	// Send Grants
	hasSendGrants := false
	sendGrantsT := table.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			cellStyle := utils.TableRowStyle
			if row == 0 {
				cellStyle = cellStyle.Bold(true).PaddingBottom(1)
			}
			if col == 0 {
				cellStyle = cellStyle.PaddingLeft(0).PaddingRight(3)
			}
			return cellStyle
		}).
		Row([]string{"Send Grants", ""}...).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false)

	{
		sendGrantOptions := db.RemotesSendGrantsListOptions{
			FilterRemoteNames: []string{remote.Name},
			IncludeAll:        true,
		}
		sendGrants, err := db.RemotesSendGrants(r).List(sendGrantOptions)
		if err != nil {
			utils.PrintErrorWithMessage("failed to query send grants", err)
			return true, err
		}

		hasSendGrants = len(sendGrants) > 0

		for _, sg := range sendGrants {
			grantStr := utils.SQLLikeStyle.Render(sg.Name) + "@" + sg.DomainFQDN
			if sg.DeletedAt != nil {
				grantStr = grantStr + " " + utils.RedStyle.Render("(deleted)")
			}

			sendGrantsT = sendGrantsT.Row([]string{grantStr, ""}...)
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
	t.Row(headerStyle.Render("Remote"))
	t.Row(headerStyle.Render("Status: ") + statusStr)
	t.Row(headerStyle.Render("Properties") + "\n\n" + propT.Render())
	t.Row(headerStyle.Render("Meta") + "\n\n" + RenderMetaSection(remote.CreatedAt, remote.UpdatedAt, remote.DeletedAt))
	if hasSendGrants {
		t.Row(sendGrantsT.Render())
	} else {
		t.Row(headerStyle.Render("Send Grants") + "\n\n" + utils.BlackStyle.Render("No send grants configured."))
	}
	fmt.Println(t.Render())
	return true, nil
}
