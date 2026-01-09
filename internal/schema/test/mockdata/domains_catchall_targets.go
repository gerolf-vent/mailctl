package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// DomainsCatchallTargetsVariant captures catch-all target config and ID.
type DomainsCatchallTargetsVariant struct {
	ID           int
	DomainID     int
	Recipient    int
	Forwarding   bool
	FallbackOnly bool
	DeletedAt    sql.NullTime
}

func (b *Builder) seedDomainsCatchallTargets() error {
	forwardingOptions := []bool{false, true}
	fallbackOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	catchallSeq := 0

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

	recipientIDs := make([]int, 0, len(b.f.Mailboxes)+len(b.f.Aliases)+len(b.f.RecipientsRelayed))
	for _, m := range b.f.Mailboxes {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if m.DeletedAt.Valid {
			continue
		}
		recipientIDs = append(recipientIDs, m.ID)
	}
	for _, a := range b.f.Aliases {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if a.DeletedAt.Valid {
			continue
		}
		recipientIDs = append(recipientIDs, a.ID)
	}
	for _, r := range b.f.RecipientsRelayed {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if r.DeletedAt.Valid {
			continue
		}
		recipientIDs = append(recipientIDs, r.ID)
	}

	if len(recipientIDs) == 0 {
		return fmt.Errorf("no recipients available for catch-all targets")
	}

	used := make(map[string]bool)
	idx := 0

	var variants []DomainsCatchallTargetsVariant
	stmt := sq.Insert("domains_catchall_targets").Columns("domain_id", "recipient_id", "forwarding_to_target_enabled", "fallback_only", "deleted_at")

	for _, domainID := range domainIDs {
		for _, forwarding := range forwardingOptions {
			for _, fallback := range fallbackOptions {
				for _, deleted := range deletedOptions {
					// ensure unique domain/recipient pair
					recipientID := recipientIDs[idx%len(recipientIDs)]
					pair := fmt.Sprintf("%d_%d", domainID, recipientID)
					attempts := 0
					for used[pair] && attempts < len(recipientIDs) {
						idx++
						recipientID = recipientIDs[idx%len(recipientIDs)]
						pair = fmt.Sprintf("%d_%d", domainID, recipientID)
						attempts++
					}
					used[pair] = true
					idx++

					catchallSeq++

					stmt = stmt.Values(domainID, recipientID, forwarding, fallback, b.nullTime(deleted))
					variants = append(variants, DomainsCatchallTargetsVariant{
						DomainID:     domainID,
						Recipient:    recipientID,
						Forwarding:   forwarding,
						FallbackOnly: fallback,
						DeletedAt:    b.nullTime(deleted),
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
		b.f.DomainsCatchallTargets[id] = variants[i]
	}

	return nil
}
