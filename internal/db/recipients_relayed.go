package db

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

type RecipientRelayed struct {
	DomainFQDN    string     `json:"domainFQDN"`
	DomainEnabled bool       `json:"domainEnabled"`
	Name          string     `json:"name"`
	Enabled       bool       `json:"enabled"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`
}

type RecipientsRelayedCreateOptions struct {
	Enabled bool
}

type RecipientsRelayedPatchOptions struct {
	Enabled *bool
}

type RecipientsRelayedListOptions struct {
	FilterDomains  []string
	ByEmail        *utils.EmailAddress
	IncludeDeleted bool
	IncludeAll     bool
}

type RecipientsRelayedRepository interface {
	List(options RecipientsRelayedListOptions) ([]RecipientRelayed, error)
	Create(email utils.EmailAddress, options RecipientsRelayedCreateOptions) error
	Patch(email utils.EmailAddress, options RecipientsRelayedPatchOptions) error
	Rename(oldEmail, newEmail utils.EmailAddress) error
	Delete(email utils.EmailAddress, options DeleteOptions) error
	Restore(email utils.EmailAddress) error
}

type recipientsRelayedRepository struct {
	r sq.BaseRunner
}

func RecipientsRelayed(r sq.BaseRunner) RecipientsRelayedRepository {
	return &recipientsRelayedRepository{
		r: r,
	}
}

func (r *recipientsRelayedRepository) List(options RecipientsRelayedListOptions) ([]RecipientRelayed, error) {
	q := sq.
		Select(
			"d.fqdn",
			"d.enabled AS domain_enabled",
			"r.name",
			"r.enabled",
			"r.created_at",
			"r.updated_at",
			"r.deleted_at",
		).
		From("recipients_relayed r").
		Join("domains_relayed d ON r.domain_id = d.ID")

	if !options.IncludeDeleted && !options.IncludeAll {
		q = q.Where(sq.Eq{"r.deleted_at": nil, "d.deleted_at": nil})
	}

	if options.IncludeDeleted {
		q = q.Where(sq.NotEq{"r.deleted_at": nil}).OrderBy("r.deleted_at")
	} else {
		q = q.OrderBy("d.fqdn", "r.name")
	}

	if len(options.FilterDomains) > 0 {
		q = q.Where(sq.Eq{"d.fqdn": options.FilterDomains})
	}

	if options.ByEmail != nil {
		q = q.Where(sq.Eq{
			"d.fqdn": options.ByEmail.DomainFQDN,
			"r.name": options.ByEmail.LocalPart,
		}).Limit(1)
	}

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(r.r).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RecipientRelayed
	for rows.Next() {
		var rr RecipientRelayed
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&rr.DomainFQDN,
			&rr.DomainEnabled,
			&rr.Name,
			&rr.Enabled,
			&rr.CreatedAt,
			&rr.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			rr.DeletedAt = &deletedAt.Time
		}
		out = append(out, rr)
	}

	return out, nil
}

func (r *recipientsRelayedRepository) Create(email utils.EmailAddress, options RecipientsRelayedCreateOptions) error {
	q := sq.
		Insert("recipients_relayed").
		Columns(
			"domain_id",
			"name",
			"enabled",
		).
		Values(
			sq.Expr("(?)", sq.
				Select("ID").
				From("domains_relayed").
				Where(sq.Eq{
					"fqdn":       email.DomainFQDN,
					"deleted_at": nil,
				}).
				Limit(1),
			),
			email.LocalPart,
			options.Enabled,
		)

	return Exec(r.r, q, 1)
}

func (r *recipientsRelayedRepository) Patch(email utils.EmailAddress, options RecipientsRelayedPatchOptions) error {
	q := sq.Update("recipients_relayed").
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains_relayed").
			Where(sq.Eq{
				"fqdn":       email.DomainFQDN,
				"deleted_at": nil,
			}).
			Limit(1),
		)).
		Where(sq.Eq{
			"name":       email.LocalPart,
			"deleted_at": nil,
		})

	if options.Enabled != nil {
		q = q.Set("enabled", *options.Enabled)
	}

	return Exec(r.r, q, 1)
}

func (r *recipientsRelayedRepository) Rename(oldEmail, newEmail utils.EmailAddress) error {
	q := sq.Update("recipients_relayed").
		Set("name", newEmail.LocalPart).
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains_relayed").
			Where(sq.Eq{
				"fqdn":       newEmail.DomainFQDN,
				"deleted_at": nil,
			}).
			Limit(1),
		)).
		Where(sq.Eq{
			"name":       oldEmail.LocalPart,
			"deleted_at": nil,
		})

	if oldEmail.DomainFQDN != newEmail.DomainFQDN {
		q = q.Set("domain_id", sq.Expr("(?)", sq.
			Select("ID").
			From("domains_relayed").
			Where(sq.Eq{
				"fqdn":       newEmail.DomainFQDN,
				"deleted_at": nil,
			}).
			Limit(1),
		))
	}

	return Exec(r.r, q, 1)
}

func (r *recipientsRelayedRepository) Delete(email utils.EmailAddress, options DeleteOptions) error {
	var q sq.Sqlizer
	if options.Permanent {
		// Hard delete
		q = sq.
			Delete("recipients_relayed").
			Where(sq.Expr("domain_id = (?)", sq.
				Select("ID").
				From("domains_relayed").
				Where(sq.Eq{
					"fqdn": email.DomainFQDN,
				}).
				Limit(1),
			)).
			Where(sq.Eq{
				"name": email.LocalPart,
			})
	} else {
		// Soft delete
		uq := sq.Update("recipients_relayed").
			Set("deleted_at", sq.Expr("NOW()")).
			Where(sq.Expr("domain_id = (?)", sq.
				Select("ID").
				From("domains_relayed").
				Where(sq.Eq{
					"fqdn": email.DomainFQDN,
				}).
				Limit(1),
			)).
			Where(sq.Eq{
				"name": email.LocalPart,
			})

		if !options.Force {
			uq = uq.Where(sq.Eq{"deleted_at": nil})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *recipientsRelayedRepository) Restore(email utils.EmailAddress) error {
	qb := sq.
		Update("recipients_relayed").
		Set("deleted_at", nil).
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains_relayed").
			Where(sq.Eq{
				"fqdn":       email.DomainFQDN,
				"deleted_at": nil, // Only allow to restore if domain is not deleted
			}).
			Limit(1),
		)).
		Where(sq.Eq{
			"name": email.LocalPart,
		})

	return Exec(r.r, qb, 1)
}
