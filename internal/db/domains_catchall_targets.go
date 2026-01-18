package db

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

type DomainCatchallTarget struct {
	DomainFQDN                string     `json:"domain"`
	TargetEmail               string     `json:"targetEmail"`
	ForwardingToTargetEnabled bool       `json:"forwardingEnabled"`
	FallbackOnly              bool       `json:"fallbackOnly"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
	DeletedAt                 *time.Time `json:"deletedAt,omitempty"`

	DomainEnabled bool `json:"-"`
}

type DomainsCatchallTargetsCreateOptions struct {
	ForwardEnabled bool
	FallbackOnly   bool
}

type DomainsCatchallTargetsPatchOptions struct {
	ForwardingToTargetEnabled *bool
	FallbackOnly              *bool
}

type DomainsCatchallTargetsListOptions struct {
	FilterDomains  []string
	IncludeDeleted bool
	IncludeAll     bool
}

type DomainsCatchallTargetsRepository interface {
	List(options DomainsCatchallTargetsListOptions) ([]DomainCatchallTarget, error)
	Create(srcDomainFQDN string, targetEmail utils.EmailAddress, options DomainsCatchallTargetsCreateOptions) error
	Patch(srcDomainFQDN string, targetEmail utils.EmailAddress, options DomainsCatchallTargetsPatchOptions) error
	Delete(srcDomainFQDN string, targetEmail utils.EmailAddress, options DeleteOptions) error
	Restore(srcDomainFQDN string, targetEmail utils.EmailAddress) error
}

type domainsCatchallTargetsRepository struct {
	r sq.BaseRunner
}

func DomainsCatchallTargets(r sq.BaseRunner) DomainsCatchallTargetsRepository {
	return &domainsCatchallTargetsRepository{
		r: r,
	}
}

func (r *domainsCatchallTargetsRepository) List(options DomainsCatchallTargetsListOptions) ([]DomainCatchallTarget, error) {
	q := sq.Select(
		"d.fqdn",
		"rd.fqdn",
		"r.name",
		"dct.forwarding_to_target_enabled",
		"dct.fallback_only",
		"dct.created_at",
		"dct.updated_at",
		"dct.deleted_at",
		"d.enabled as domain_enabled",
	).
		From("domains_catchall_targets dct").
		Join("domains d ON dct.domain_id = d.ID").
		Join("recipients r ON dct.recipient_id = r.ID").
		Join("domains rd ON r.domain_id = rd.ID").
		OrderBy("d.fqdn", "rd.fqdn", "r.name")

	if len(options.FilterDomains) > 0 {
		q = q.Where(sq.Eq{"d.fqdn": options.FilterDomains})
	}

	if options.IncludeDeleted {
		q = q.Where(sq.NotEq{"dct.deleted_at": nil})
	} else if !options.IncludeAll {
		q = q.Where(sq.Eq{"dct.deleted_at": nil})
	}

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(r.r).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DomainCatchallTarget
	for rows.Next() {
		var td DomainCatchallTarget
		var targetName string
		var targetDomain string
		var deletedAt sql.NullTime
		var domainEnabled sql.NullBool

		err := rows.Scan(
			&td.DomainFQDN,
			&targetDomain,
			&targetName,
			&td.ForwardingToTargetEnabled,
			&td.FallbackOnly,
			&td.CreatedAt,
			&td.UpdatedAt,
			&deletedAt,
			&domainEnabled,
		)
		if err != nil {
			return nil, err
		}

		td.TargetEmail = fmt.Sprintf("%s@%s", targetName, targetDomain)

		if deletedAt.Valid {
			td.DeletedAt = &deletedAt.Time
		}
		if domainEnabled.Valid {
			td.DomainEnabled = domainEnabled.Bool
		}

		out = append(out, td)
	}

	return out, nil
}

func (r *domainsCatchallTargetsRepository) Create(srcDomainFQDN string, targetEmail utils.EmailAddress, options DomainsCatchallTargetsCreateOptions) (err error) {
	q := sq.Insert("domains_catchall_targets").
		Columns(
			"domain_id",
			"recipient_id",
			"forwarding_to_target_enabled",
			"fallback_only",
		).
		Values(
			sq.Expr("(?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{"fqdn": srcDomainFQDN, "deleted_at": nil}).
				Limit(1),
			),
			sq.Expr("(?)", sq.
				Select("r.ID").
				From("recipients r").
				Where(sq.Expr("domain_id = (?)", sq.
					Select("ID").
					From("domains").
					Where(sq.Eq{"fqdn": targetEmail.DomainFQDN, "deleted_at": nil}).
					Limit(1),
				)).
				Where(sq.Eq{"r.name": targetEmail.LocalPart, "r.deleted_at": nil}).
				Limit(1),
			),
			options.ForwardEnabled,
			options.FallbackOnly,
		)

	return Exec(r.r, q, 1)
}

func (r *domainsCatchallTargetsRepository) Patch(srcDomainFQDN string, targetEmail utils.EmailAddress, options DomainsCatchallTargetsPatchOptions) (err error) {
	q := sq.
		Update("domains_catchall_targets").
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains").
			Where(sq.Eq{
				"fqdn":       srcDomainFQDN,
				"deleted_at": nil,
			},
			).Limit(1),
		)).
		Where(sq.Expr("recipient_id = (?)", sq.
			Select("r.ID").
			From("recipients r").
			Where(sq.Expr("domain_id = (?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{
					"fqdn":       targetEmail.DomainFQDN,
					"deleted_at": nil,
				}).Limit(1),
			)).
			Where(sq.Eq{
				"r.name":       targetEmail.LocalPart,
				"r.deleted_at": nil,
			}).Limit(1),
		))

	if options.ForwardingToTargetEnabled != nil {
		q = q.Set("forwarding_to_target_enabled", *options.ForwardingToTargetEnabled)
	}
	if options.FallbackOnly != nil {
		q = q.Set("fallback_only", *options.FallbackOnly)
	}

	return Exec(r.r, q, 1)
}

func (r *domainsCatchallTargetsRepository) Delete(srcDomainFQDN string, targetEmail utils.EmailAddress, options DeleteOptions) error {
	domainIdExpr := sq.Expr("domain_id = (?)", sq.
		Select("ID").
		From("domains").
		Where(sq.Eq{
			"fqdn": srcDomainFQDN,
		}).
		Limit(1),
	)

	targetIdExpr := sq.Expr("recipient_id = (?)", sq.
		Select("r.ID").
		From("recipients r").
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains").
			Where(sq.Eq{
				"fqdn": targetEmail.DomainFQDN,
			}).
			Limit(1),
		)).
		Where(sq.Eq{
			"r.name": targetEmail.LocalPart,
		}).
		Limit(1),
	)

	var q sq.Sqlizer
	if options.Permanent {
		q = sq.
			Delete("domains_catchall_targets").
			Where(domainIdExpr).
			Where(targetIdExpr)
	} else {
		uq := sq.
			Update("domains_catchall_targets").
			Set("deleted_at", sq.Expr("NOW()")).
			Where(domainIdExpr).
			Where(targetIdExpr)

		if !options.Force {
			uq = uq.Where(sq.Eq{"deleted_at": nil})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *domainsCatchallTargetsRepository) Restore(srcDomainFQDN string, targetEmail utils.EmailAddress) (err error) {
	q := sq.
		Update("domains_catchall_targets").
		Set("deleted_at", nil).
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains").
			Where(sq.Eq{
				"fqdn":       srcDomainFQDN,
				"deleted_at": nil, // Source domain must exist and not be deleted
			}).
			Limit(1),
		)).
		Where(sq.Expr("recipient_id = (?)", sq.
			Select("r.ID").
			From("recipients r").
			Where(sq.Expr("domain_id = (?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{
					"fqdn":       targetEmail.DomainFQDN,
					"deleted_at": nil,
				}).
				Limit(1),
			)).
			Where(sq.Eq{
				"r.name": targetEmail.LocalPart,
			}).
			Limit(1),
		))

	return Exec(r.r, q, 1)
}
