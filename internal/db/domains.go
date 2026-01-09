package db

import (
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Domain struct {
	FQDN                string     `json:"fqdn"`
	Type                string     `json:"type"`
	Enabled             bool       `json:"enabled"`
	Transport           *string    `json:"transport,omitempty"`
	TransportName       *string    `json:"transportName,omitempty"`
	TargetDomainFQDN    *string    `json:"targetDomainFQDN,omitempty"`
	TargetDomainEnabled bool       `json:"targetDomainEnabled"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	DeletedAt           *time.Time `json:"deletedAt,omitempty"`
}

type DomainsCreateOptions struct {
	DomainType       string
	TransportName    string // for managed and relayed
	TargetDomainFQDN string // for canonical
	Enabled          bool
}

type DomainsPatchOptions struct {
	Enabled          *bool
	TransportName    *string // for managed and relayed
	TargetDomainFQDN *string // for canonical
}

type DomainsListOptions struct {
	ByFQDN         string
	IncludeDeleted bool
	IncludeAll     bool
}

type DomainsRepository interface {
	List(options DomainsListOptions) ([]Domain, error)
	Create(fqdn string, options DomainsCreateOptions) error
	Patch(fqdn string, options DomainsPatchOptions) error
	Rename(oldFQDN, newFQDN string) error
	Delete(fqdn string, options DeleteOptions) error
	Restore(fqdn string) error
}

type domainsRepository struct {
	r sq.BaseRunner
}

func Domains(r sq.BaseRunner) DomainsRepository {
	return &domainsRepository{
		r: r,
	}
}

func (r *domainsRepository) List(options DomainsListOptions) ([]Domain, error) {
	q := sq.
		Select(
			"d.fqdn",
			"d.type",
			"d.enabled",
			"postfix.transport_string(t.method, t.host, t.port, t.mx_lookup) AS transport_spec",
			"t.name AS transport_name",
			"td.fqdn AS target_domain_fqdn",
			"COALESCE(td.enabled, false) AS target_domain_enabled",
			"d.created_at",
			"d.updated_at",
			"d.deleted_at",
		).
		From("domains d").
		LeftJoin("transports t ON d.transport_id = t.ID").
		LeftJoin("domains td ON d.target_domain_id = td.ID")

	if !options.IncludeDeleted && !options.IncludeAll {
		q = q.Where(sq.Eq{"d.deleted_at": nil})
	}

	if options.IncludeDeleted {
		q = q.Where(sq.NotEq{"d.deleted_at": nil}).OrderBy("d.deleted_at")
	} else {
		q = q.OrderBy("d.type", "d.fqdn")
	}

	if options.ByFQDN != "" {
		q = q.Where(sq.Eq{
			"d.fqdn": options.ByFQDN,
		}).Limit(1)
	}

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(r.r).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Domain
	for rows.Next() {
		var d Domain
		var transport sql.NullString
		var transportName sql.NullString
		var targetDomainFQDN sql.NullString
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&d.FQDN,
			&d.Type,
			&d.Enabled,
			&transport,
			&transportName,
			&targetDomainFQDN,
			&d.TargetDomainEnabled,
			&d.CreatedAt,
			&d.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}

		if transport.Valid {
			s := transport.String
			d.Transport = &s
		}
		if transportName.Valid {
			s := transportName.String
			d.TransportName = &s
		}
		if targetDomainFQDN.Valid {
			s := targetDomainFQDN.String
			d.TargetDomainFQDN = &s
		}
		if deletedAt.Valid {
			d.DeletedAt = &deletedAt.Time
		}

		out = append(out, d)
	}

	return out, nil
}

func (r *domainsRepository) Create(fqdn string, options DomainsCreateOptions) error {
	tableName := "domains_" + options.DomainType

	switch options.DomainType {
	case "managed", "relayed":
		q := sq.
			Insert(tableName).
			Columns(
				"fqdn",
				"transport_id",
				"enabled",
			).
			Values(
				fqdn,
				sq.Expr("(?)", sq.
					Select("ID").
					From("transports").
					Where(sq.Eq{"name": options.TransportName, "deleted_at": nil}).
					Limit(1),
				),
				options.Enabled,
			)
		return Exec(r.r, q, 1)
	case "canonical":
		q := sq.
			Insert(tableName).
			Columns(
				"fqdn",
				"target_domain_id",
				"enabled",
			).
			Values(
				fqdn,
				sq.Expr("(?)", sq.
					Select("ID").
					From("domains").
					Where(sq.Eq{"fqdn": options.TargetDomainFQDN, "deleted_at": nil}).
					Limit(1),
				),
				options.Enabled,
			)
		return Exec(r.r, q, 1)
	case "alias":
		q := sq.
			Insert(tableName).
			Columns(
				"fqdn",
				"enabled",
			).
			Values(
				fqdn,
				options.Enabled,
			)
		return Exec(r.r, q, 1)
	default:
		return errors.New("unsupported domain type")
	}
}

func (r *domainsRepository) Patch(fqdn string, options DomainsPatchOptions) error {
	domainId, tableName, err := queryDomainIdAndTable(r.r, fqdn)
	if err != nil {
		return err
	}

	switch tableName {
	case "domains_managed", "domains_relayed":
		if options.TargetDomainFQDN != nil {
			return errors.New("only domains of type 'canonical' can have a target domain")
		}
	case "domains_alias":
		if options.TransportName != nil {
			return errors.New("domains of type 'alias' cannot have a transport")
		}
		if options.TargetDomainFQDN != nil {
			return errors.New("domains of type 'alias' cannot have a target domain")
		}
	case "domains_canonical":
		if options.TransportName != nil {
			return errors.New("domains of type 'canonical' cannot have a transport")
		}
	}

	q := sq.
		Update(tableName).
		Where(sq.Eq{
			"ID":         domainId,
			"deleted_at": nil,
		})

	if options.Enabled != nil {
		q = q.Set("enabled", *options.Enabled)
	}
	if options.TransportName != nil {
		q = q.Set("transport_id", sq.Expr("(?)", sq.
			Select("ID").
			From("transports").
			Where(sq.Eq{
				"name":       *options.TransportName,
				"deleted_at": nil,
			}).
			Limit(1)),
		)
	}
	if options.TargetDomainFQDN != nil {
		q = q.Set("target_domain_id", sq.Expr("(?)", sq.
			Select("ID").
			From("domains").
			Where(sq.Eq{
				"fqdn":       *options.TargetDomainFQDN,
				"deleted_at": nil,
			}).
			Limit(1)),
		)
	}

	return Exec(r.r, q, 1)
}

func (r *domainsRepository) Rename(oldFQDN, newFQDN string) error {
	domainId, tableName, err := queryDomainIdAndTable(r.r, oldFQDN)
	if err != nil {
		return err
	}

	q := sq.Update(tableName).
		Set("fqdn", newFQDN).
		Where(sq.Eq{
			"ID":         domainId,
			"deleted_at": nil,
		})

	return Exec(r.r, q, 1)
}

func (r *domainsRepository) Delete(fqdn string, options DeleteOptions) error {
	domainId, tableName, err := queryDomainIdAndTable(r.r, fqdn)
	if err != nil {
		return err
	}

	var q sq.Sqlizer
	if options.Permanent {
		q = sq.
			Delete(tableName).
			Where(sq.Eq{
				"ID": domainId,
			}).
			PlaceholderFormat(sq.Dollar)
	} else {
		uq := sq.
			Update(tableName).
			Set("deleted_at", sq.Expr("NOW()")).
			Where(sq.Eq{
				"ID": domainId,
			})

		if !options.Force {
			uq = uq.Where(sq.Eq{"deleted_at": nil})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *domainsRepository) Restore(fqdn string) error {
	domainId, tableName, err := queryDomainIdAndTable(r.r, fqdn)
	if err != nil {
		return err
	}

	qb := sq.Update(tableName).
		Set("deleted_at", nil).
		Where(sq.Eq{
			"ID": domainId,
		})

	return Exec(r.r, qb, 1)
}
