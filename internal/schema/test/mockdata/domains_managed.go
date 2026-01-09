package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// DomainsManagedVariant captures managed domain config and ID.
type DomainsManagedVariant struct {
	ID          int
	FQDN        string
	TransportID int
	Enabled     bool
	DeletedAt   sql.NullTime
}

func (b *Builder) seedDomainsManaged() error {
	enabledOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	managedSeq := 0
	var variants []DomainsManagedVariant
	stmt := sq.Insert("domains_managed").Columns("fqdn", "transport_id", "enabled", "deleted_at")

	for _, transport := range b.f.Transports {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if transport.DeletedAt.Valid {
			continue
		}
		for _, enabled := range enabledOptions {
			for _, deleted := range deletedOptions {
				managedSeq++
				fqdn := fmt.Sprintf("m%d.test", managedSeq)

				stmt = stmt.Values(fqdn, transport.ID, enabled, b.nullTime(deleted))
				variants = append(variants, DomainsManagedVariant{
					FQDN:        fqdn,
					TransportID: transport.ID,
					Enabled:     enabled,
					DeletedAt:   b.nullTime(deleted),
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
		b.f.DomainsManaged[id] = variants[i]
	}

	return nil
}
