package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// RecipientsRelayedVariant captures relayed recipient config and ID.
type RecipientsRelayedVariant struct {
	ID        int
	DomainID  int
	Name      string
	Enabled   bool
	DeletedAt sql.NullTime
}

func (b *Builder) seedRelayedRecipients() error {
	enabledOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	relayedSeq := 0
	var variants []RecipientsRelayedVariant
	stmt := sq.Insert("recipients_relayed").Columns("domain_id", "name", "enabled", "deleted_at")

	for _, domain := range b.f.DomainsRelayed {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if domain.DeletedAt.Valid {
			continue
		}
		for _, enabled := range enabledOptions {
			for _, deleted := range deletedOptions {
				relayedSeq++
				name := fmt.Sprintf("relay_%d", relayedSeq)

				stmt = stmt.Values(domain.ID, name, enabled, b.nullTime(deleted))
				variants = append(variants, RecipientsRelayedVariant{
					DomainID:  domain.ID,
					Name:      name,
					Enabled:   enabled,
					DeletedAt: b.nullTime(deleted),
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
		b.f.RecipientsRelayed[id] = variants[i]
	}

	return nil
}
