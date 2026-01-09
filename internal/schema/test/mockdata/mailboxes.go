package mockdata

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// MailboxesVariant captures mailbox config and ID.
type MailboxesVariant struct {
	ID               int
	DomainID         int
	Name             string
	TransportID      sql.NullInt64
	PasswordHash     sql.NullString
	StorageQuota     sql.NullInt32
	LoginEnabled     bool
	ReceivingEnabled bool
	SendingEnabled   bool
	DeletedAt        sql.NullTime
}

func (b *Builder) seedMailboxes() error {
	passOptions := []sql.NullString{{}, {String: "bcrypt:mbx", Valid: true}}
	quotaOptions := []sql.NullInt32{{}, {Int32: 512, Valid: true}}
	loginOptions := []bool{false, true}
	recvOptions := []bool{false, true}
	sendOptions := []bool{false, true}
	deletedOptions := []bool{false, true}

	var transportID int
	transportOptions := []sql.NullInt64{}
	for _, t := range b.f.Transports {
		if t.DeletedAt.Valid {
			continue
		}
		transportID = t.ID
		transportOptions = append(transportOptions, sql.NullInt64{Int64: int64(t.ID), Valid: true})
		break // use only the first transport to avoid combinatorial explosion
	}
	if len(transportOptions) == 0 {
		transportOptions = append(transportOptions, sql.NullInt64{}) // allow null transport if none exist
	}

	mailboxSeq := 0
	var variants []MailboxesVariant
	q := sq.
		Insert("mailboxes").
		Columns("domain_id", "name", "transport_id", "password_hash", "storage_quota", "login_enabled", "receiving_enabled", "sending_enabled", "deleted_at")

	for _, domain := range b.f.DomainsManaged {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if domain.DeletedAt.Valid {
			continue
		}
		if domain.TransportID != transportID {
			continue
		}
		for _, pass := range passOptions {
			for _, quota := range quotaOptions {
				for _, login := range loginOptions {
					for _, recv := range recvOptions {
						for _, send := range sendOptions {
							for _, del := range deletedOptions {
								for _, transport := range transportOptions {
									mailboxSeq++
									name := fmt.Sprintf("mbx_%d", mailboxSeq)
									pwd := pass
									if pass.Valid {
										pwd.String = fmt.Sprintf("%s_%d", pass.String, mailboxSeq)
									}

									q = q.Values(domain.ID, name, transport, pwd, quota, login, recv, send, b.nullTime(del))
									variants = append(variants, MailboxesVariant{
										DomainID:         domain.ID,
										Name:             name,
										TransportID:      transport,
										PasswordHash:     pwd,
										StorageQuota:     quota,
										LoginEnabled:     login,
										ReceivingEnabled: recv,
										SendingEnabled:   send,
										DeletedAt:        b.nullTime(del),
									})
								}
							}
						}
					}
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
		b.f.Mailboxes[id] = variants[i]
	}

	return nil
}
