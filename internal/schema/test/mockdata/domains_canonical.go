package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// DomainsCanonicalVariant captures canonical domain config and ID.
type DomainsCanonicalVariant struct {
	ID           int
	FQDN         string
	TargetDomain int
	Enabled      bool
	DeletedAt    sql.NullTime
}

func (b *Builder) seedDomainsCanonical() error {
	enabledOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	canonicalSeq := 0

	// Targets can be any recipient-capable domain: managed, relayed, alias
	var targetDomains []int
	for _, d := range b.f.DomainsManaged {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		targetDomains = append(targetDomains, d.ID)
	}
	for _, d := range b.f.DomainsRelayed {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		targetDomains = append(targetDomains, d.ID)
	}
	for _, d := range b.f.DomainsAlias {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		targetDomains = append(targetDomains, d.ID)
	}

	var variants []DomainsCanonicalVariant
	stmt := sq.Insert("domains_canonical").Columns("fqdn", "target_domain_id", "enabled", "deleted_at")

	for _, targetID := range targetDomains {
		for _, enabled := range enabledOptions {
			for _, deleted := range deletedOptions {
				canonicalSeq++
				fqdn := fmt.Sprintf("c%d.test", canonicalSeq)

				stmt = stmt.Values(fqdn, targetID, enabled, b.nullTime(deleted))
				variants = append(variants, DomainsCanonicalVariant{
					FQDN:         fqdn,
					TargetDomain: targetID,
					Enabled:      enabled,
					DeletedAt:    b.nullTime(deleted),
				})
			}
		}
	}

	ids, err := b.insertIDs(stmt)
	if err != nil {
		return err
	}

	for i, id := range ids {
		variants[i].ID = id
		b.f.DomainsCanonical[id] = variants[i]
	}

	return nil
}
