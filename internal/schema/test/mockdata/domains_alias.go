package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// DomainsAliasVariant captures alias domain config and ID.
type DomainsAliasVariant struct {
	ID        int
	FQDN      string
	Enabled   bool
	DeletedAt sql.NullTime
}

func (b *Builder) seedDomainsAlias() error {
	enabledOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	aliasSeq := 0
	var variants []DomainsAliasVariant
	stmt := sq.Insert("domains_alias").Columns("fqdn", "enabled", "deleted_at")

	for _, enabled := range enabledOptions {
		for _, deleted := range deletedOptions {
			aliasSeq++
			fqdn := fmt.Sprintf("a%d.test", aliasSeq)

			stmt = stmt.Values(fqdn, enabled, b.nullTime(deleted))
			variants = append(variants, DomainsAliasVariant{
				FQDN:      fqdn,
				Enabled:   enabled,
				DeletedAt: b.nullTime(deleted),
			})
		}
	}

	ids, err := b.insertIDs(stmt)
	if err != nil {
		return err
	}

	for i, id := range ids {
		variants[i].ID = id
		b.f.DomainsAlias[id] = variants[i]
	}

	return nil
}
