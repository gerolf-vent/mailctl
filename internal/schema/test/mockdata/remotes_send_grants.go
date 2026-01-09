package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// RemotesSendGrantsVariant captures remote send grant config and ID.
type RemotesSendGrantsVariant struct {
	ID        int
	RemoteID  int
	DomainID  int
	Name      string
	DeletedAt sql.NullTime
}

func (b *Builder) seedRemotesSendGrants() error {
	namePatterns := []string{"%", "sales%", "user"}
	deletedOptions := []bool{false, true}

	var domainIDs []int
	for _, d := range b.f.DomainsManaged {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		domainIDs = append(domainIDs, d.ID)
	}
	for _, d := range b.f.DomainsRelayed {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		domainIDs = append(domainIDs, d.ID)
	}
	for _, d := range b.f.DomainsAlias {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if d.DeletedAt.Valid {
			continue
		}
		domainIDs = append(domainIDs, d.ID)
	}

	var variants []RemotesSendGrantsVariant
	stmt := sq.Insert("remotes_send_grants").Columns("remote_id", "domain_id", "name", "deleted_at")

	for _, remote := range b.f.Remotes {
		for _, domainID := range domainIDs {
			for _, pattern := range namePatterns {
				for _, deleted := range deletedOptions {
					name := fmt.Sprintf("%s_%d_%t", pattern, remote.ID, deleted)

					stmt = stmt.Values(remote.ID, domainID, name, b.nullTime(deleted))
					variants = append(variants, RemotesSendGrantsVariant{
						RemoteID:  remote.ID,
						DomainID:  domainID,
						Name:      name,
						DeletedAt: b.nullTime(deleted),
					})
				}
			}
		}
	}

	ids, err := b.insertIDs(stmt)
	if err != nil {
		return err
	}

	for i, id := range ids {
		variants[i].ID = id
		b.f.RemotesSendGrants[id] = variants[i]
	}

	return nil
}
