package db

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Transport struct {
	Name      string     `json:"name"`
	Method    string     `json:"method"`
	Host      string     `json:"host"`
	Port      *uint16    `json:"port,omitempty"`
	MXLookup  bool       `json:"mx_lookup"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type TransportsCreateOptions struct {
	Method   string
	Host     string
	Port     sql.NullInt32
	MxLookup bool
}

type TransportsPatchOptions struct {
	Method   *string
	Host     *string
	Port     *sql.NullInt32
	MxLookup *bool
}

type TransportsListOptions struct {
	ByName         string
	IncludeDeleted bool
	IncludeAll     bool
}

type TransportsRepository interface {
	List(options TransportsListOptions) ([]Transport, error)
	Create(name string, options TransportsCreateOptions) error
	Patch(name string, options TransportsPatchOptions) error
	Rename(oldName, newName string) error
	Delete(name string, options DeleteOptions) error
	Restore(name string) error
}

type transportsRepository struct {
	r sq.BaseRunner
}

func Transports(r sq.BaseRunner) TransportsRepository {
	return &transportsRepository{
		r: r,
	}
}

func (r *transportsRepository) List(options TransportsListOptions) ([]Transport, error) {
	q := sq.
		Select(
			"name",
			"method",
			"host",
			"port",
			"mx_lookup",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		From("transports")

	if options.IncludeDeleted {
		q = q.Where(sq.NotEq{"deleted_at": nil}).OrderBy("deleted_at")
	} else if !options.IncludeAll {
		q = q.Where(sq.Eq{"deleted_at": nil})
	}

	if !options.IncludeDeleted {
		q = q.OrderBy("name", "method", "host", "port")
	}

	if options.ByName != "" {
		q = q.Where(sq.Eq{"name": options.ByName}).Limit(1)
	}

	rows, err := q.
		PlaceholderFormat(sq.Dollar).
		RunWith(r.r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Transport
	for rows.Next() {
		var t Transport
		var port sql.NullInt64
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&t.Name,
			&t.Method,
			&t.Host,
			&port,
			&t.MXLookup,
			&t.CreatedAt,
			&t.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if port.Valid {
			v := uint16(port.Int64)
			t.Port = &v
		}
		if deletedAt.Valid {
			dat := deletedAt.Time
			t.DeletedAt = &dat
		}
		results = append(results, t)
	}

	return results, nil
}

func (r *transportsRepository) Create(name string, options TransportsCreateOptions) error {
	q := sq.
		Insert("transports").
		Columns(
			"name",
			"method",
			"host",
			"port",
			"mx_lookup",
		).
		Values(
			name,
			options.Method,
			options.Host,
			options.Port,
			options.MxLookup,
		)

	return Exec(r.r, q, 1)
}

func (r *transportsRepository) Patch(name string, options TransportsPatchOptions) error {
	q := sq.Update("transports").
		Where(sq.Eq{
			"name":       name,
			"deleted_at": nil,
		})

	if options.Method != nil {
		q = q.Set("method", *options.Method)
	}
	if options.Host != nil {
		q = q.Set("host", *options.Host)
	}
	if options.Port != nil {
		q = q.Set("port", *options.Port)
	}
	if options.MxLookup != nil {
		q = q.Set("mx_lookup", *options.MxLookup)
	}

	return Exec(r.r, q, 1)
}

func (r *transportsRepository) Rename(oldName, newName string) error {
	q := sq.
		Update("transports").
		Set("name", newName).
		Where(sq.Eq{
			"name":       oldName,
			"deleted_at": nil,
		})

	return Exec(r.r, q, 1)
}

func (r *transportsRepository) Delete(name string, options DeleteOptions) error {
	var q sq.Sqlizer
	if options.Permanent {
		// Hard delete
		q = sq.
			Delete("transports").
			Where(sq.Eq{
				"name": name,
			})
	} else {
		// Soft delete
		uq := sq.
			Update("transports").
			Set("deleted_at", sq.Expr("NOW()")).
			Where(sq.Eq{
				"name": name,
			})

		if !options.Force {
			uq = uq.Where(sq.Eq{"deleted_at": nil})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *transportsRepository) Restore(name string) error {
	q := sq.
		Update("transports").
		Set("deleted_at", nil).
		Where(sq.Eq{
			"name": name,
		})

	return Exec(r.r, q, 1)
}
