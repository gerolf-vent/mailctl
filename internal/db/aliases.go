package db

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

type Alias struct {
	DomainFQDN    string     `json:"domainFQDN"`
	DomainEnabled bool       `json:"domainEnabled"`
	Name          *string    `json:"name"`
	Enabled       bool       `json:"enabled"`
	TargetCount   int        `json:"targetCount"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`
}

type AliasesListOptions struct {
	FilterDomains  []string
	ByEmail        *utils.EmailAddress
	IncludeDeleted bool
	IncludeAll     bool
	Verbose        bool
}

type AliasesCreateOptions struct {
	Disabled bool
}

type AliasesPatchOptions struct {
	Enabled *bool
}

type AliasesRepository interface {
	List(options AliasesListOptions) ([]Alias, error)
	Create(email utils.EmailAddress, options AliasesCreateOptions) error
	Patch(email utils.EmailAddress, options AliasesPatchOptions) error
	Rename(oldEmail utils.EmailAddress, newEmail utils.EmailAddress) error
	Delete(email utils.EmailAddress, options DeleteOptions) error
	Restore(email utils.EmailAddress) error
}

type aliasesRepository struct {
	r sq.BaseRunner
}

func Aliases(r sq.BaseRunner) AliasesRepository {
	return &aliasesRepository{
		r: r,
	}
}

func (r *aliasesRepository) List(options AliasesListOptions) ([]Alias, error) {
	q := sq.
		Select(
			"d.fqdn",
			"d.enabled AS domain_enabled",
			"a.name",
			"a.enabled",
			"COUNT(at.ID) as target_count",
			"a.created_at",
			"a.updated_at",
			"a.deleted_at",
		).
		From("aliases a").
		Join("domains d ON a.domain_id = d.ID").
		LeftJoin("aliases_targets_recursive at ON a.ID = at.alias_id").
		GroupBy("d.fqdn", "a.name", "a.enabled", "d.enabled", "a.created_at", "a.updated_at", "a.deleted_at")

	if !options.IncludeDeleted && !options.IncludeAll {
		q = q.
			Where(sq.Eq{
				"at.deleted_at": nil,
				"a.deleted_at":  nil,
				"d.deleted_at":  nil,
			})
	}

	if options.IncludeDeleted {
		q = q.
			Where(sq.NotEq{"a.deleted_at": nil}).
			OrderBy("a.deleted_at")
	} else {
		q = q.OrderBy("d.fqdn", "a.name")
	}

	if len(options.FilterDomains) > 0 {
		q = q.Where(sq.Eq{"d.fqdn": options.FilterDomains})
	}

	if options.ByEmail != nil {
		q = q.Where(sq.Eq{
			"d.fqdn": options.ByEmail.DomainFQDN,
			"a.name": options.ByEmail.LocalPart,
		}).Limit(1)
	}

	// Query entries from database
	rows, err := q.
		PlaceholderFormat(sq.Dollar).
		RunWith(r.r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Read rows
	var out []Alias
	for rows.Next() {
		var a Alias
		var name sql.NullString
		var deletedAt sql.NullTime

		err = rows.Scan(
			&a.DomainFQDN,
			&a.DomainEnabled,
			&name,
			&a.Enabled,
			&a.TargetCount,
			&a.CreatedAt,
			&a.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		if name.Valid {
			a.Name = &name.String
		}

		if deletedAt.Valid {
			a.DeletedAt = &deletedAt.Time
		}

		out = append(out, a)
	}

	return out, nil
}

func (r *aliasesRepository) Create(email utils.EmailAddress, options AliasesCreateOptions) error {
	q := sq.
		Insert("aliases").
		Columns(
			"domain_id",
			"name",
			"enabled",
		).
		Values(
			sq.Expr("(?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{
					"fqdn":       email.DomainFQDN,
					"deleted_at": nil,
				}).
				Limit(1),
			),
			email.LocalPart,
			!options.Disabled,
		)

	return Exec(r.r, q, 1)
}

func (r *aliasesRepository) Patch(email utils.EmailAddress, options AliasesPatchOptions) error {
	q := sq.
		Update("aliases").
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains").
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

func (r *aliasesRepository) Rename(oldEmail utils.EmailAddress, newEmail utils.EmailAddress) error {
	q := sq.
		Update("aliases").
		Set("name", newEmail.LocalPart).
		Set("domain_id", sq.Expr("(?)", sq.
			Select("ID").
			From("domains").
			Where(sq.Eq{
				"fqdn":       newEmail.DomainFQDN,
				"deleted_at": nil,
			}).
			Limit(1),
		)).
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains").
			Where(sq.Eq{
				"fqdn":       oldEmail.DomainFQDN,
				"deleted_at": nil,
			}).
			Limit(1),
		)).
		Where(sq.Eq{
			"name":       oldEmail.LocalPart,
			"deleted_at": nil,
		})

	return Exec(r.r, q, 1)
}

func (r *aliasesRepository) Delete(email utils.EmailAddress, options DeleteOptions) error {
	var q sq.Sqlizer
	if options.Permanent {
		// Hard delete
		q = sq.
			Delete("aliases").
			Where(sq.Eq{
				"name": email.LocalPart,
			}).
			Where(sq.Expr("domain_id = (?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{
					"fqdn": email.DomainFQDN,
				}).
				Limit(1),
			))
	} else {
		// Soft delete
		uq := sq.
			Update("aliases").
			Set("deleted_at", sq.Expr("NOW()")).
			Where(sq.Eq{
				"name": email.LocalPart,
			}).
			Where(sq.Expr("domain_id = (?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{
					"fqdn": email.DomainFQDN,
				}).
				Limit(1),
			))

		// Override deleted_at if force option is set
		if !options.Force {
			uq = uq.Where(sq.Eq{"deleted_at": nil})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *aliasesRepository) Restore(email utils.EmailAddress) error {
	q := sq.
		Update("aliases").
		Set("deleted_at", nil).
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains").
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
