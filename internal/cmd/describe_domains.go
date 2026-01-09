package cmd

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

func describeDomains(r sq.BaseRunner, fqdn string) (bool, error) {
	options := db.DomainsListOptions{
		ByFQDN:     fqdn,
		IncludeAll: true,
	}

	domains, err := db.Domains(r).List(options)
	if err != nil {
		return false, err
	} else if len(domains) == 0 {
		return false, nil
	}

	domain := domains[0]

	// Determine status
	var statusStr string
	if domain.DeletedAt != nil {
		statusStr = utils.RedStyle.Bold(true).Render("Deleted")
	} else if domain.Enabled {
		statusStr = utils.GreenStyle.Bold(true).Render("Operational")
	} else if !domain.Enabled {
		statusStr = utils.YellowStyle.Bold(true).Render("Disabled")
	} else {
		statusStr = utils.BlueStyle.Bold(true).Render("Unknown")
	}

	// Properties
	propT := table.New().
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

	propT.Rows([][]string{
		{"FQDN:", fqdn},
		{"Type:", strings.ToUpper(domain.Type[0:1]) + domain.Type[1:]},
		{"Enabled:", utils.MaybeEnabledStyle.Render(domain.Enabled)},
	}...)
	switch domain.Type {
	case "canonical":
		propT.Row("Target Domain:", utils.MaybeEmptyStyle.Render(domain.TargetDomainFQDN))
	case "alias":
		// Alias domains have no extra properties
	default:
		propT.Row("Transport:", utils.MaybeIDSuffixStyle.Render(domain.Transport, domain.TransportName))
	}

	// Reference counts
	var referencesCount map[string]int64 = make(map[string]int64)
	switch domain.Type {
	case "managed":
		var count int64
		err = sq.
			Select("COUNT(*)").
			From("mailboxes").
			Join("domains_managed AS d ON mailboxes.domain_id = d.id").
			Where(sq.Eq{"d.fqdn": fqdn}).
			PlaceholderFormat(sq.Dollar).
			RunWith(r).
			QueryRow().
			Scan(&count)
		if err != nil {
			utils.PrintErrorWithMessage("failed to count related mailboxes", err)
			return true, err
		}
		referencesCount["Mailboxes"] = count

		err = sq.
			Select("COUNT(*)").
			From("aliases AS a").
			Join("domains_managed AS d ON a.domain_id = d.id").
			Where(sq.Eq{"d.fqdn": fqdn}).
			PlaceholderFormat(sq.Dollar).
			RunWith(r).
			QueryRow().
			Scan(&count)
		if err != nil {
			utils.PrintErrorWithMessage("failed to count related aliases", err)
			return true, err
		}
		referencesCount["Aliases"] = count
	case "relayed":
		var count int64
		err = sq.
			Select("COUNT(*)").
			From("recipients_relayed AS rr").
			Join("domains_relayed AS d ON rr.domain_id = d.id").
			Where(sq.Eq{"d.fqdn": fqdn}).
			PlaceholderFormat(sq.Dollar).
			RunWith(r).
			QueryRow().
			Scan(&count)
		if err != nil {
			utils.PrintErrorWithMessage("failed to count related relayed recipients", err)
			return true, err
		}
		referencesCount["Recipients"] = count

		err = sq.
			Select("COUNT(*)").
			From("aliases").
			Join("domains_relayed AS d ON aliases.domain_id = d.id").
			Where(sq.Eq{"d.fqdn": fqdn}).
			PlaceholderFormat(sq.Dollar).
			RunWith(r).
			QueryRow().
			Scan(&count)
		if err != nil {
			utils.PrintErrorWithMessage("failed to count related aliases", err)
			return true, err
		}
		referencesCount["Aliases"] = count
	case "alias":
		var count int64
		err = sq.
			Select("COUNT(*)").
			From("aliases").
			Join("domains_alias AS d ON aliases.domain_id = d.id").
			Where(sq.Eq{"d.fqdn": fqdn}).
			PlaceholderFormat(sq.Dollar).
			RunWith(r).
			QueryRow().
			Scan(&count)
		if err != nil {
			utils.PrintErrorWithMessage("failed to count related aliases", err)
			return true, err
		}
		referencesCount["Aliases"] = count
	}
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
	for objType, count := range referencesCount {
		referencesT = referencesT.Row([]string{objType + ":", fmt.Sprintf("%d", count)}...)
	}

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
		result, err := db.PostfixVirtualMailboxDomains(r, fqdn)
		funcsT.Row("postfix.virtual_mailbox_domains", utils.TestFunctionResultStyle.Render(result, err))

		result, err = db.PostfixVirtualAliasDomains(r, fqdn)
		funcsT.Row("postfix.virtual_alias_domains", utils.TestFunctionResultStyle.Render(result, err))

		result, err = db.PostfixRelayDomains(r, fqdn)
		funcsT.Row("postfix.relay_domains", utils.TestFunctionResultStyle.Render(result, err))
	}

	// Output final table
	headerStyle := lipgloss.NewStyle().Bold(true)
	t := table.New().
		BorderStyle(utils.BlackStyle).
		BorderRow(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return utils.TableRowStyle
		})
	t.Row(headerStyle.Render("Domain"))
	t.Row(headerStyle.Render("Status: ") + statusStr)
	t.Row(headerStyle.Render("Properties") + "\n\n" + propT.Render())
	t.Row(headerStyle.Render("Meta") + "\n\n" + RenderMetaSection(domain.CreatedAt, domain.UpdatedAt, domain.DeletedAt))
	if len(referencesCount) > 0 {
		t.Row(headerStyle.Render("References") + "\n\n" + referencesT.Render())
	}
	t.Row(headerStyle.Render("Functions") + "\n\n" + funcsT.Render())
	fmt.Println(t.Render())
	return true, nil
}

func DescribeCanonicalAddress(r sq.BaseRunner, email utils.EmailAddress) (bool, error) {
	options := db.DomainsListOptions{
		ByFQDN:     email.DomainFQDN,
		IncludeAll: true,
	}

	domains, err := db.Domains(r).List(options)
	if err != nil {
		return false, err
	} else if len(domains) == 0 {
		return false, nil
	}

	domain := domains[0]

	if domain.Type != "canonical" {
		return false, nil
	}

	// Determine status
	var statusStr string
	if domain.DeletedAt != nil {
		statusStr = utils.RedStyle.Bold(true).Render("Deleted")
	} else if domain.Enabled && domain.TargetDomainEnabled {
		statusStr = utils.GreenStyle.Bold(true).Render("Operational")
	} else {
		statusStr = utils.YellowStyle.Bold(true).Render("Disabled")
	}

	// Properties
	propT := table.New().
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

	propT.Rows([][]string{
		{"Address:", email.String()},
		{"Type:", "Canonical Address"},
		{"Enabled:", utils.MaybeEnabledStyle.Render(domain.Enabled)},
		{"Target Domain:", utils.MaybeEmptyStyle.Render(domain.TargetDomainFQDN)},
	}...)

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
		result, err := db.PostfixCanonicalMaps(r, email)
		funcsT.Row("postfix.canonical_maps", utils.TestFunctionResultStyle.Render(result, err))
	}

	// Output final table
	headerStyle := lipgloss.NewStyle().Bold(true)
	t := table.New().
		BorderStyle(utils.BlackStyle).
		BorderRow(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return utils.TableRowStyle
		})

	t.Row(headerStyle.Render("Canonical Address"))
	t.Row(headerStyle.Render("Status: ") + statusStr)
	t.Row(headerStyle.Render("Properties") + "\n\n" + propT.Render())
	t.Row(headerStyle.Render("Functions") + "\n\n" + funcsT.Render())
	fmt.Println(t.Render())
	return true, nil
}
