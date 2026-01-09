package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// AliasesTargetsForeignVariant captures foreign alias target config and ID.
type AliasesTargetsForeignVariant struct {
	ID         int
	AliasID    int
	FQDN       string
	Name       string
	Forwarding bool
	DeletedAt  sql.NullTime
}

func (b *Builder) seedAliasesTargetsForeign() error {
	forwardingOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	foreignSeq := 0
	var variants []AliasesTargetsForeignVariant

	q := sq.
		Insert("aliases_targets_foreign").
		Columns("alias_id", "fqdn", "name", "forwarding_to_target_enabled", "deleted_at")

	for _, alias := range b.f.Aliases {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if alias.DeletedAt.Valid {
			continue
		}
		for _, forwarding := range forwardingOptions {
			for _, deleted := range deletedOptions {
				for range 5 { // Create 5 variants for each combination
					foreignSeq++
					fqdn := fmt.Sprintf("foreign-%d.example", foreignSeq)
					name := fmt.Sprintf("f%d", foreignSeq)

					q = q.Values(alias.ID, fqdn, name, forwarding, b.nullTime(deleted))
					variants = append(variants, AliasesTargetsForeignVariant{
						AliasID:    alias.ID,
						FQDN:       fqdn,
						Name:       name,
						Forwarding: forwarding,
						DeletedAt:  b.nullTime(deleted),
					})
				}
			}
		}
	}

	ids, err := b.insertIDs(q)
	if err != nil {
		return err
	}

	for i, id := range ids {
		variants[i].ID = id
		b.f.AliasesTargetsForeign[id] = variants[i]
	}

	return nil
}
