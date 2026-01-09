package db

import (
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/gerolf-vent/mailctl/internal/utils/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

type Mailbox struct {
	DomainFQDN       string     `json:"domainFQDN"`
	DomainEnabled    bool       `json:"domainEnabled"`
	Name             string     `json:"name"`
	LoginEnabled     bool       `json:"loginEnabled"`
	ReceivingEnabled bool       `json:"receivingEnabled"`
	SendingEnabled   bool       `json:"sendingEnabled"`
	PasswordSet      bool       `json:"passwordHashSet"`
	StorageQuota     *int32     `json:"storageQuota,omitempty"`
	Transport        *string    `json:"transport,omitempty"`
	TransportName    *string    `json:"transportName,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"deletedAt,omitempty"`
}

type MailboxesCreateOptions struct {
	PasswordHash     sql.NullString
	Quota            sql.NullInt32
	TransportName    sql.NullString
	LoginEnabled     bool
	ReceivingEnabled bool
	SendingEnabled   bool
}

type MailboxesPatchOptions struct {
	PasswordHash  *sql.NullString
	Quota         *sql.NullInt32
	TransportName *sql.NullString
	Login         *bool
	Receiving     *bool
	Sending       *bool
}

type MailboxesListOptions struct {
	FilterDomains  []string
	ByEmail        *utils.EmailAddress
	IncludeDeleted bool
	IncludeAll     bool
}

type MailboxesRepository interface {
	List(options MailboxesListOptions) ([]Mailbox, error)
	Authenticate(email utils.EmailAddress, givenPassword string) (matches bool, err error)
	Create(address utils.EmailAddress, options MailboxesCreateOptions) error
	Patch(email utils.EmailAddress, options MailboxesPatchOptions) error
	Rename(oldEmail utils.EmailAddress, newEmail utils.EmailAddress) error
	Delete(email utils.EmailAddress, options DeleteOptions) error
	Restore(email utils.EmailAddress) error
}

type mailboxesRepository struct {
	r sq.BaseRunner
}

func Mailboxes(r sq.BaseRunner) MailboxesRepository {
	return &mailboxesRepository{
		r: r,
	}
}

func (r *mailboxesRepository) List(options MailboxesListOptions) ([]Mailbox, error) {
	q := sq.
		Select(
			"d.fqdn",
			"d.enabled AS domain_enabled",
			"m.name",
			"m.login_enabled",
			"m.receiving_enabled",
			"m.sending_enabled",
			"m.password_hash IS NOT NULL AS auth_data_hash_set",
			"m.storage_quota",
			"CASE WHEN t.ID IS NOT NULL THEN postfix.transport_string(t.method, t.host, t.port, t.mx_lookup) ELSE NULL END AS transport",
			"t.name AS transport_name",
			"m.created_at",
			"m.updated_at",
			"m.deleted_at",
		).
		From("mailboxes m").
		Join("domains_managed d ON m.domain_id = d.ID").
		LeftJoin("transports t ON m.transport_id = t.ID")

	if !options.IncludeDeleted && !options.IncludeAll {
		q = q.Where(sq.Eq{
			"m.deleted_at": nil,
			"d.deleted_at": nil,
		})
	}

	if options.IncludeDeleted {
		q = q.Where(sq.NotEq{"m.deleted_at": nil}).OrderBy("m.deleted_at")
	} else {
		q = q.OrderBy("d.fqdn", "m.name")
	}

	if len(options.FilterDomains) > 0 {
		q = q.Where(sq.Eq{"d.fqdn": options.FilterDomains})
	}

	if options.ByEmail != nil {
		q = q.Where(sq.Eq{
			"d.fqdn": options.ByEmail.DomainFQDN,
			"m.name": options.ByEmail.LocalPart,
		}).Limit(1)
	}

	rows, err := q.
		PlaceholderFormat(sq.Dollar).
		RunWith(r.r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Mailbox
	for rows.Next() {
		var m Mailbox
		var storageQuota sql.NullInt32
		var transport sql.NullString
		var transportName sql.NullString
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&m.DomainFQDN,
			&m.DomainEnabled,
			&m.Name,
			&m.LoginEnabled,
			&m.ReceivingEnabled,
			&m.SendingEnabled,
			&m.PasswordSet,
			&storageQuota,
			&transport,
			&transportName,
			&m.CreatedAt,
			&m.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}

		if storageQuota.Valid {
			m.StorageQuota = &storageQuota.Int32
		}
		if transport.Valid {
			m.Transport = &transport.String
		}
		if transportName.Valid {
			m.TransportName = &transportName.String
		}
		if deletedAt.Valid {
			m.DeletedAt = &deletedAt.Time
		}

		out = append(out, m)
	}

	return out, nil
}

func (r *mailboxesRepository) Authenticate(email utils.EmailAddress, givenPassword string) (matches bool, err error) {
	var passwordHash sql.NullString

	err = sq.
		Select("password_hash").
		From("mailboxes").
		Where("domain_id = (?)", sq.
			Select("ID").
			From("domains_managed").
			Where(sq.Eq{
				"fqdn": email.DomainFQDN,
			}),
		).
		Where(sq.Eq{
			"name": email.LocalPart,
		}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.r).
		QueryRow().
		Scan(&passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return
	}

	if !passwordHash.Valid {
		return false, nil
	}

	if passwordHash.String[:10] == "$argon2id$" {
		err = argon2.CompareHashAndPassword([]byte(passwordHash.String), []byte(givenPassword))
		if errors.Is(err, argon2.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return err == nil, err
	} else if passwordHash.String[:2] == "$2" {
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash.String), []byte(givenPassword))
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return err == nil, err
	}

	return false, errors.New("unsupported password hash type")
}

func (r *mailboxesRepository) Create(address utils.EmailAddress, options MailboxesCreateOptions) (err error) {
	var transportId any = nil
	if options.TransportName.Valid {
		transportId = sq.
			Select("ID").
			From("transports").
			Where(sq.Eq{
				"name":       options.TransportName.String,
				"deleted_at": nil,
			})
	}

	q := sq.
		Insert("mailboxes").
		Columns(
			"domain_id",
			"name",
			"password_hash",
			"storage_quota",
			"transport_id",
			"login_enabled",
			"receiving_enabled",
			"sending_enabled",
		).
		Values(
			sq.Expr("(?)", sq.
				Select("ID").
				From("domains_managed").
				Where(sq.Eq{
					"fqdn":       address.DomainFQDN,
					"deleted_at": nil,
				}).
				Limit(1)),
			address.LocalPart,
			options.PasswordHash,
			options.Quota,
			transportId,
			options.LoginEnabled,
			options.ReceivingEnabled,
			options.SendingEnabled,
		).
		PlaceholderFormat(sq.Dollar)

	return Exec(r.r, q, 1)
}

func (r *mailboxesRepository) Patch(email utils.EmailAddress, options MailboxesPatchOptions) error {
	q := sq.
		Update("mailboxes").
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains").
			Where(sq.Eq{
				"fqdn":       email.DomainFQDN,
				"deleted_at": nil,
			}),
		)).
		Where(sq.Eq{
			"name":       email.LocalPart,
			"deleted_at": nil,
		})

	if options.PasswordHash != nil {
		q = q.Set("password_hash", *options.PasswordHash)
	}
	if options.Quota != nil {
		q = q.Set("storage_quota", *options.Quota)
	}
	if options.TransportName != nil {
		if options.TransportName.Valid {
			q = q.Set("transport_id", sq.Expr("(?)", sq.
				Select("ID").
				From("transports").
				Where(sq.Eq{
					"name":       *options.TransportName,
					"deleted_at": nil,
				}).
				Limit(1),
			))
		} else {
			q = q.Set("transport_id", nil)
		}
	}
	if options.Login != nil {
		q = q.Set("login_enabled", *options.Login)
	}
	if options.Receiving != nil {
		q = q.Set("receiving_enabled", *options.Receiving)
	}
	if options.Sending != nil {
		q = q.Set("sending_enabled", *options.Sending)
	}

	return Exec(r.r, q, 1)
}

func (r *mailboxesRepository) Rename(oldEmail utils.EmailAddress, newEmail utils.EmailAddress) error {
	q := sq.
		Update("mailboxes").
		Set("name", newEmail.LocalPart).
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains_managed").
			Where(sq.Eq{
				"fqdn":       oldEmail.DomainFQDN,
				"deleted_at": nil,
			}).
			Limit(1)),
		).
		Where(sq.Eq{
			"name":       oldEmail.LocalPart,
			"deleted_at": nil,
		})

	if oldEmail.DomainFQDN != newEmail.DomainFQDN {
		q.Set("domain_id", sq.Expr("(?)", sq.
			Select("ID").
			From("domains_managed").
			Where(sq.Eq{
				"fqdn":       newEmail.DomainFQDN,
				"deleted_at": nil,
			}).
			Limit(1)),
		)
	}

	return Exec(r.r, q, 1)
}

func (r *mailboxesRepository) Delete(email utils.EmailAddress, options DeleteOptions) error {
	var q sq.Sqlizer
	if options.Permanent {
		// Hard delete
		q = sq.
			Delete("mailboxes").
			Where(sq.Eq{"name": email.LocalPart}).
			Where(sq.Expr("domain_id = (?)", sq.Select("ID").From("domains_managed").Where(sq.Eq{"fqdn": email.DomainFQDN}).Limit(1)))
	} else {
		// Soft delete
		uq := sq.
			Update("mailboxes").
			Set("deleted_at", sq.Expr("NOW()")).
			Where(sq.Eq{"name": email.LocalPart}).
			Where(sq.Expr("domain_id = (?)", sq.Select("ID").From("domains_managed").Where(sq.Eq{"fqdn": email.DomainFQDN}).Limit(1)))

		if !options.Force {
			uq = uq.Where(sq.Eq{"deleted_at": nil})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *mailboxesRepository) Restore(email utils.EmailAddress) error {
	q := sq.
		Update("mailboxes").
		Set("deleted_at", nil).
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains_managed").
			Where(sq.Eq{
				"fqdn":       email.DomainFQDN,
				"deleted_at": nil, // Only allow restoring if domain is not deleted too
			}).
			Limit(1),
		)).
		Where(sq.Eq{
			"name": email.LocalPart,
		})

	return Exec(r.r, q, 1)
}
