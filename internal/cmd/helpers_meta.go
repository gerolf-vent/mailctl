package cmd

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

// RenderMetaSection returns a standardized table showing created/updated/deleted timestamps.
func RenderMetaSection(createdAt time.Time, updatedAt time.Time, deletedAt *time.Time) string {
	metaT := table.New().
		Rows([][]string{
			{"Created:", utils.MaybeTimeStyle.Render(createdAt)},
			{"Updated:", utils.MaybeTimeStyle.Render(updatedAt)},
			{"Deleted:", utils.MaybeTimeStyle.Render(deletedAt)},
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

	return metaT.Render()
}
