package db

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Remote struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Enabled     bool       `json:"enabled"`
	PasswordSet bool       `json:"passwordSet"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type RemotesCreateOptions struct {
	PasswordHash sql.NullString
	Enabled      bool
}

type RemotesPatchOptions struct {
	PasswordHash *sql.NullString
	Enabled      *bool
}

type RemotesListOptions struct {
	ByName         string
	IncludeDeleted bool
	IncludeAll     bool
}

type RemotesRepository interface {
	List(options RemotesListOptions) ([]Remote, error)
	Create(name string, options RemotesCreateOptions) error
	Patch(name string, options RemotesPatchOptions) error
	Rename(oldName, newName string) error
	Delete(name string, options DeleteOptions) error
	Restore(name string) error
}

type remotesRepository struct {
	r sq.BaseRunner
}

func Remotes(r sq.BaseRunner) RemotesRepository {
	return &remotesRepository{
		r: r,
	}
}

func (r *remotesRepository) List(options RemotesListOptions) ([]Remote, error) {
	q := sq.
		Select(
			"id",
			"name",
			"enabled",
			"password_hash IS NOT NULL AS password_set",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		From("remotes")

	if options.ByName != "" {
		q = q.Where(sq.Eq{"name": options.ByName}).Limit(1)
	}

	if !options.IncludeDeleted && !options.IncludeAll {
		q = q.Where(sq.Eq{"deleted_at": nil})
	}

	if options.IncludeDeleted {
		q = q.Where(sq.NotEq{"deleted_at": nil}).OrderBy("deleted_at")
	} else {
		q = q.OrderBy("name")
	}

	rows, err := q.
		PlaceholderFormat(sq.Dollar).
		RunWith(r.r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Remote
	for rows.Next() {
		var rr Remote
		var deletedAt sql.NullTime
		if err := rows.Scan(&rr.ID, &rr.Name, &rr.Enabled, &rr.PasswordSet, &rr.CreatedAt, &rr.UpdatedAt, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			rr.DeletedAt = &t
		}
		out = append(out, rr)
	}

	return out, nil
}

func (r *remotesRepository) Create(name string, options RemotesCreateOptions) error {
	q := sq.
		Insert("remotes").
		Columns(
			"name",
			"password_hash",
			"enabled",
		).
		Values(
			name,
			options.PasswordHash,
			options.Enabled,
		)

	return Exec(r.r, q, 1)
}

func (r *remotesRepository) Patch(name string, options RemotesPatchOptions) error {
	q := sq.Update("remotes").
		Where(sq.Eq{
			"name":       name,
			"deleted_at": nil,
		})

	if options.PasswordHash != nil {
		q = q.Set("password_hash", *options.PasswordHash)
	}
	if options.Enabled != nil {
		q = q.Set("enabled", *options.Enabled)
	}

	return Exec(r.r, q, 1)
}

func (r *remotesRepository) Rename(oldName, newName string) error {
	q := sq.Update("remotes").
		Set("name", newName).
		Where(sq.Eq{
			"name":       oldName,
			"deleted_at": nil,
		})

	return Exec(r.r, q, 1)
}

func (r *remotesRepository) Delete(name string, options DeleteOptions) error {
	var q sq.Sqlizer
	if options.Permanent {
		// Hard delete
		q = sq.
			Delete("remotes").
			Where(sq.Expr("ID = (?)", sq.
				Select("id").
				From("remotes").
				Where(sq.Eq{
					"name": name,
				}),
			))
	} else {
		// Soft delete
		uq := sq.Update("remotes").
			Set("deleted_at", sq.Expr("NOW()")).
			Where(sq.Expr("ID = (?)", sq.
				Select("id").
				From("remotes").
				Where(sq.Eq{
					"name": name,
				}),
			))

		if !options.Force {
			uq = uq.Where(sq.Eq{"deleted_at": nil})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *remotesRepository) Restore(name string) error {
	q := sq.Update("remotes").
		Set("deleted_at", nil).
		Where(sq.Eq{
			"name": name,
		})

	return Exec(r.r, q, 1)
}
