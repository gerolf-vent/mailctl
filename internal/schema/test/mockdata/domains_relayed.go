package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// DomainsRelayedVariant captures relayed domain config and ID.
type DomainsRelayedVariant struct {
	ID          int
	FQDN        string
	TransportID int
	Enabled     bool
	DeletedAt   sql.NullTime
}

func (b *Builder) seedDomainsRelayed() error {
	enabledOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	relayedSeq := 0
	var variants []DomainsRelayedVariant
	stmt := sq.Insert("domains_relayed").Columns("fqdn", "transport_id", "enabled", "deleted_at")

	for _, transport := range b.f.Transports {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if transport.DeletedAt.Valid {
			continue
		}
		for _, enabled := range enabledOptions {
			for _, deleted := range deletedOptions {
				relayedSeq++
				fqdn := fmt.Sprintf("r%d.test", relayedSeq)

				stmt = stmt.Values(fqdn, transport.ID, enabled, b.nullTime(deleted))
				variants = append(variants, DomainsRelayedVariant{
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
		b.f.DomainsRelayed[id] = variants[i]
	}

	return nil
}
