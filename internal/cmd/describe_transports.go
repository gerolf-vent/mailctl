package cmd

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

// Describe prints a detailed view for a single transport by name.
// Returns (true, nil) when the transport was found and printed, (false, nil) when
// the transport was not found, or (true, err) on error.
func describeTransports(r sq.BaseRunner, name string) (bool, error) {
	options := db.TransportsListOptions{
		ByName:     name,
		IncludeAll: true,
	}

	transports, err := db.Transports(r).List(options)
	if err != nil {
		return false, err
	}
	if len(transports) == 0 {
		return false, nil // No transport found with that name
	}

	transport := transports[0]

	// Determine status
	var statusStr string
	if transport.DeletedAt != nil {
		statusStr = utils.RedStyle.Bold(true).Render("Deleted")
	} else {
		statusStr = utils.GreenStyle.Bold(true).Render("Ready")
	}

	// MX Lookup style
	mxLookupStyle := utils.MaybeEnabledStyle
	mxLookupStyle.TrueStyle = utils.BlueStyle
	mxLookupStyle.FalseStyle = utils.BlackStyle

	// Properties
	propT := table.New().
		Rows([][]string{
			{"Name:", transport.Name},
			{"Method:", transport.Method},
			{"Host:", transport.Host},
			{"Port:", utils.MaybeEmptyStyle.Render(transport.Port)},
			{"MX Lookup:", mxLookupStyle.Render(transport.MXLookup)},
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

	// Reference counts
	referencesT := table.New().
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

	// Count domains using this transport
	var domainCount int64
	err = sq.
		Select("COUNT(*)").
		From("domains").
		Join("transports t ON domains.transport_id = t.id").
		Where(sq.Eq{"t.name": name}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&domainCount)
	if err != nil {
		utils.PrintErrorWithMessage("failed to count related domains", err)
		return true, err
	}
	referencesT = referencesT.Row([]string{"Domains:", fmt.Sprintf("%d", domainCount)}...)

	// Output final table
	headerStyle := lipgloss.NewStyle().Bold(true)
	t := table.New().
		BorderStyle(utils.BlackStyle).
		BorderRow(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return utils.TableRowStyle
		})
	t.Row(headerStyle.Render("Transport"))
	t.Row(headerStyle.Render("Status: ") + statusStr)
	t.Row(headerStyle.Render("Properties") + "\n\n" + propT.Render())
	t.Row(headerStyle.Render("Meta") + "\n\n" + RenderMetaSection(transport.CreatedAt, transport.UpdatedAt, transport.DeletedAt))
	t.Row(headerStyle.Render("References") + "\n\n" + referencesT.Render())
	fmt.Println(t.Render())
	return true, nil
}
