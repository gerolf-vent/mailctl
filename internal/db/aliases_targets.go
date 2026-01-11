package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

type AliasTarget struct {
	AliasEmail                string     `json:"aliasEmail"`
	TargetEmail               string     `json:"targetEmail"`
	IsForeign                 bool       `json:"isForeign"`
	ForwardingToTargetEnabled bool       `json:"forwardingEnabled"`
	SendingFromTargetEnabled  bool       `json:"sendingEnabled"` // Only for recursive
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
	DeletedAt                 *time.Time `json:"deletedAt,omitempty"`

	AliasEnabled  bool  `json:"-"`
	DomainEnabled *bool `json:"-"`
}

type AliasesTargetsCreateOptions struct {
	ForwardEnabled bool
	SendEnabled    bool
}

type AliasesTargetsPatchOptions struct {
	ForwardingToTargetEnabled *bool
	SendingFromTargetEnabled  *bool
}

type AliasesTargetsListOptions struct {
	FilterAliasEmails []utils.EmailAddress
	IncludeDeleted    bool
	IncludeAll        bool
}

type AliasesTargetsRepository interface {
	List(options AliasesTargetsListOptions) ([]AliasTarget, error)
	Create(aliasEmail utils.EmailAddress, targetEmail utils.EmailAddress, options AliasesTargetsCreateOptions) error
	Patch(aliasEmail utils.EmailAddress, targetEmail utils.EmailAddress, options AliasesTargetsPatchOptions) error
	Delete(aliasEmail utils.EmailAddress, targetEmail utils.EmailAddress, options DeleteOptions) error
	Restore(aliasEmail utils.EmailAddress, targetEmail utils.EmailAddress) error
}

type aliasesTargetsRepository struct {
	r sq.BaseRunner
}

func AliasesTargets(r sq.BaseRunner) AliasesTargetsRepository {
	return &aliasesTargetsRepository{
		r: r,
	}
}

func (r *aliasesTargetsRepository) List(options AliasesTargetsListOptions) ([]AliasTarget, error) {
	q := sq.
		Select(
			"ad.fqdn",
			"a.name",
			"at.fqdn",
			"at.name",
			"(at.type = 'foreign') as is_foreign",
			"at.forwarding_to_target_enabled",
			"at.sending_from_target_enabled",
			"at.created_at",
			"at.updated_at",
			"at.deleted_at",
			"td.enabled as domain_enabled",
			"a.enabled as alias_enabled",
		).
		From("aliases_targets at").
		Join("aliases a ON at.alias_id = a.ID").
		Join("domains ad ON a.domain_id = ad.ID").
		LeftJoin("domains td ON td.fqdn = at.fqdn").
		OrderBy("ad.fqdn", "a.name", "at.fqdn", "at.name")

	if len(options.FilterAliasEmails) > 0 {
		or := sq.Or{}
		for _, email := range options.FilterAliasEmails {
			or = append(or, sq.Eq{
				"a.name":  email.LocalPart,
				"ad.fqdn": email.DomainFQDN,
			})
		}
		q = q.Where(or)
	}

	if options.IncludeDeleted {
		q = q.Where(sq.NotEq{"at.deleted_at": nil}).OrderBy("deleted_at DESC")
	} else if !options.IncludeAll {
		q = q.Where(sq.Eq{"at.deleted_at": nil})
	}

	if !options.IncludeDeleted {
		q = q.OrderBy("ad.fqdn", "a.name", "at.fqdn", "at.name")
	}

	rows, err := q.
		PlaceholderFormat(sq.Dollar).
		RunWith(r.r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AliasTarget
	for rows.Next() {
		var at AliasTarget
		var aliasName string
		var aliasDomain string
		var targetName string
		var targetDomain string
		var sendingEnabled sql.NullBool
		var deletedAt sql.NullTime
		var domainEnabled sql.NullBool

		err := rows.Scan(
			&aliasDomain,
			&aliasName,
			&targetDomain,
			&targetName,
			&at.IsForeign,
			&at.ForwardingToTargetEnabled,
			&sendingEnabled,
			&at.CreatedAt,
			&at.UpdatedAt,
			&deletedAt,
			&domainEnabled,
			&at.AliasEnabled,
		)
		if err != nil {
			return nil, err
		}

		at.AliasEmail = fmt.Sprintf("%s@%s", aliasName, aliasDomain)
		at.TargetEmail = fmt.Sprintf("%s@%s", targetName, targetDomain)

		if sendingEnabled.Valid {
			at.SendingFromTargetEnabled = sendingEnabled.Bool
		}

		if deletedAt.Valid {
			at.DeletedAt = &deletedAt.Time
		}

		if domainEnabled.Valid {
			at.DomainEnabled = &domainEnabled.Bool
		}

		out = append(out, at)
	}

	return out, nil
}

func (r *aliasesTargetsRepository) Create(aliasEmail utils.EmailAddress, targetEmail utils.EmailAddress, options AliasesTargetsCreateOptions) (err error) {
	// Common alias ID query
	aliasIdQ := sq.
		Select("a.ID").
		From("aliases a").
		Join("domains d ON a.domain_id = d.ID").
		Where(sq.Eq{
			"a.name":       aliasEmail.LocalPart,
			"d.fqdn":       aliasEmail.DomainFQDN,
			"a.deleted_at": nil,
			"d.deleted_at": nil,
		}).
		Limit(1)

	// Check if target domain exists in the db
	var doesTargetDomainExist bool
	doesTargetDomainExist, err = doesDomainExist(r.r, targetEmail.DomainFQDN)
	if err != nil {
		return
	}

	var q sq.Sqlizer
	if doesTargetDomainExist {
		q = sq.
			Insert("aliases_targets_recursive").
			Columns(
				"alias_id",
				"recipient_id",
				"forwarding_to_target_enabled",
				"sending_from_target_enabled",
			).
			Values(
				sq.Expr("(?)", aliasIdQ),
				sq.Expr("(?)", sq.
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
						"r.name":       targetEmail.LocalPart,
						"r.deleted_at": nil,
					}).
					Limit(1),
				),
				options.ForwardEnabled,
				options.SendEnabled,
			)
	} else {
		if options.SendEnabled == true {
			err = errors.New("sending from foreign targets is not supported")
			return
		}

		q = sq.
			Insert("aliases_targets_foreign").
			Columns(
				"alias_id",
				"fqdn",
				"name",
				"forwarding_to_target_enabled",
			).
			Values(
				sq.Expr("(?)", aliasIdQ),
				targetEmail.DomainFQDN,
				targetEmail.LocalPart,
				options.ForwardEnabled,
			)
	}

	return Exec(r.r, q, 1)
}

func (r *aliasesTargetsRepository) Patch(aliasEmail utils.EmailAddress, targetEmail utils.EmailAddress, options AliasesTargetsPatchOptions) (err error) {
	// Get type of target
	var targetId int
	var targetTable string
	targetId, targetTable, err = queryAliasesTargetsIdAndTable(r.r, aliasEmail, targetEmail)
	if err != nil {
		err = fmt.Errorf("failed to query target ID and table: %w", err)
		return
	}

	q := sq.
		Update(targetTable).
		Where(sq.Eq{
			"ID": targetId,
		})

	if options.ForwardingToTargetEnabled != nil {
		q = q.Set("forwarding_to_target_enabled", *options.ForwardingToTargetEnabled)
	}

	// Sending from target is only supported for recursive targets
	switch targetTable {
	case "aliases_targets_recursive":
		if options.SendingFromTargetEnabled != nil {
			q = q.Set("sending_from_target_enabled", *options.SendingFromTargetEnabled)
		}
	case "aliases_targets_foreign":
		if options.SendingFromTargetEnabled != nil {
			err = errors.New("enabling/disabling sending from foreign targets is not supported")
			return
		}
	}

	return Exec(r.r, q, 1)
}

func (r *aliasesTargetsRepository) Delete(aliasEmail utils.EmailAddress, targetEmail utils.EmailAddress, options DeleteOptions) (err error) {
	// Get target ID and type
	var targetId int
	var targetTable string
	targetId, targetTable, err = queryAliasesTargetsIdAndTable(r.r, aliasEmail, targetEmail)
	if err != nil {
		err = fmt.Errorf("failed to query target ID and table: %w", err)
		return
	}

	var q sq.Sqlizer
	if options.Permanent {
		// Hard delete
		q = sq.
			Delete(targetTable).
			Where(sq.Eq{
				"ID": targetId,
			})
	} else {
		// Soft delete
		uq := sq.
			Update(targetTable).
			Set("deleted_at", sq.Expr("NOW()")).
			Where(sq.Eq{
				"ID": targetId,
			})

		// Override deleted_at if force option is set
		if !options.Force {
			uq = uq.Where(sq.Eq{"deleted_at": nil})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *aliasesTargetsRepository) Restore(aliasEmail utils.EmailAddress, targetEmail utils.EmailAddress) (err error) {
	// Get target ID and type
	var targetId int
	var targetTable string
	targetId, targetTable, err = queryAliasesTargetsIdAndTable(r.r, aliasEmail, targetEmail)
	if err != nil {
		err = fmt.Errorf("failed to query target ID and table: %w", err)
		return
	}

	q := sq.
		Update(targetTable).
		Set("deleted_at", nil).
		Where(sq.Eq{
			"ID": targetId,
		})

	return Exec(r.r, q, 1)
}
