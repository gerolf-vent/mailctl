package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// AliasesVariant captures alias config and ID.
type AliasesVariant struct {
	ID        int
	DomainID  int
	Name      string
	Enabled   bool
	DeletedAt sql.NullTime
}

func (b *Builder) seedAliases() error {
	enabledOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	var domainIDs []int
	var activeDomainIDs []int
	for _, d := range b.f.DomainsManaged {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		domainIDs = append(domainIDs, d.ID)
		if d.Enabled {
			activeDomainIDs = append(activeDomainIDs, d.ID)
		}
	}
	for _, d := range b.f.DomainsRelayed {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		domainIDs = append(domainIDs, d.ID)
		if d.Enabled {
			activeDomainIDs = append(activeDomainIDs, d.ID)
		}
	}
	for _, d := range b.f.DomainsAlias {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		domainIDs = append(domainIDs, d.ID)
		if d.Enabled {
			activeDomainIDs = append(activeDomainIDs, d.ID)
		}
	}

	aliasSeq := 0
	var variants []AliasesVariant
	q := sq.Insert("aliases").Columns("domain_id", "name", "enabled", "deleted_at")

	for _, domainID := range domainIDs {
		for _, enabled := range enabledOptions {
			for _, deleted := range deletedOptions {
				aliasSeq++
				name := fmt.Sprintf("alias_%d", aliasSeq)

				q = q.Values(domainID, name, enabled, b.nullTime(deleted))
				variants = append(variants, AliasesVariant{
					DomainID:  domainID,
					Name:      name,
					Enabled:   enabled,
					DeletedAt: b.nullTime(deleted),
				})
			}
		}
	}

	for _, domainID := range activeDomainIDs {
		// Add some always-enabled aliases for target testing
		for range 3 {
			aliasSeq++
			name := fmt.Sprintf("alias_%d", aliasSeq)

			q = q.Values(domainID, name, true, nil)
			variants = append(variants, AliasesVariant{
				DomainID: domainID,
				Name:     name,
				Enabled:  true,
			})
		}
	}

	ids, err := b.insertIDs(q)
	if err != nil {
		return err
	}

	for i, id := range ids {
		variants[i].ID = id
		b.f.Aliases[id] = variants[i]
	}

	return nil
}
