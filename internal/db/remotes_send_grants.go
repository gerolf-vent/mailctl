package db

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

type RemoteSendGrant struct {
	ID            int        `json:"-"`
	RemoteID      int        `json:"-"`
	RemoteName    string     `json:"remote_name"`
	DomainID      int        `json:"-"`
	DomainFQDN    string     `json:"domain_fqdn"`
	DomainEnabled bool       `json:"domain_enabled"`
	Name          string     `json:"name"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type RemotesSendGrantsCreateOptions struct {
}

type RemotesSendGrantsListOptions struct {
	FilterRemoteNames []string
	MatchEmail        *utils.EmailAddressOrWildcard
	IncludeDeleted    bool
	IncludeAll        bool
}

type RemotesSendGrantsRepository interface {
	List(options RemotesSendGrantsListOptions) ([]RemoteSendGrant, error)
	Create(remoteName string, email utils.EmailAddressOrWildcard, options RemotesSendGrantsCreateOptions) error
	Delete(remoteName string, email utils.EmailAddressOrWildcard, options DeleteOptions) error
	Restore(remoteName string, email utils.EmailAddressOrWildcard) error
}

type remotesSendGrantsRepository struct {
	r sq.BaseRunner
}

func RemotesSendGrants(r sq.BaseRunner) RemotesSendGrantsRepository {
	return &remotesSendGrantsRepository{
		r: r,
	}
}

func (r *remotesSendGrantsRepository) List(options RemotesSendGrantsListOptions) ([]RemoteSendGrant, error) {
	q := sq.
		Select(
			"rsg.ID",
			"rsg.remote_id",
			"r.name",
			"rsg.domain_id",
			"d.fqdn AS domain_fqdn",
			"d.enabled AS domain_enabled",
			"rsg.name",
			"rsg.created_at",
			"rsg.updated_at",
			"rsg.deleted_at",
		).
		From("remotes_send_grants rsg").
		Join("remotes r ON rsg.remote_id = r.ID").
		Join("domains d ON rsg.domain_id = d.ID").
		OrderBy("r.name", "d.fqdn", "rsg.name")

	if len(options.FilterRemoteNames) > 0 {
		q = q.Where(sq.Eq{"r.name": options.FilterRemoteNames})
	}

	if options.MatchEmail != nil {
		if options.MatchEmail.IsWildcard() {
			q = q.Where(sq.Eq{"d.fqdn": options.MatchEmail.DomainFQDN})
		} else {
			q = q.Where(sq.Or{
				sq.Eq{"d.fqdn": options.MatchEmail.DomainFQDN, "rsg.name": *options.MatchEmail.LocalPart},
				sq.Eq{"d.fqdn": options.MatchEmail.DomainFQDN, "rsg.name": nil},
			})
		}
	}

	if options.IncludeDeleted {
		q = q.Where(sq.NotEq{"rsg.deleted_at": nil})
	} else if !options.IncludeAll {
		q = q.Where(sq.Eq{"rsg.deleted_at": nil})
		q = q.Where(sq.Eq{"r.deleted_at": nil})
		q = q.Where(sq.Eq{"d.deleted_at": nil})
	}

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(r.r).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RemoteSendGrant
	for rows.Next() {
		var sg RemoteSendGrant
		var deletedAt sql.NullTime
		if err := rows.Scan(&sg.ID, &sg.RemoteID, &sg.RemoteName, &sg.DomainID, &sg.DomainFQDN, &sg.Name, &sg.CreatedAt, &sg.UpdatedAt, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			sg.DeletedAt = &t
		}
		out = append(out, sg)
	}

	return out, nil
}

func (r *remotesSendGrantsRepository) Create(remoteName string, email utils.EmailAddressOrWildcard, options RemotesSendGrantsCreateOptions) error {
	q := sq.
		Insert("remotes_send_grants").
		Columns(
			"remote_id",
			"domain_id",
			"name",
		).
		Values(
			sq.Expr("(?)", sq.
				Select("ID").
				From("remotes").
				Where(sq.Eq{
					"name":       remoteName,
					"deleted_at": nil,
				}),
			),
			sq.Expr("(?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{
					"fqdn":       email.DomainFQDN,
					"deleted_at": nil,
				}),
			),
			email.LocalPart,
		)

	return Exec(r.r, q, 1)
}

func (r *remotesSendGrantsRepository) Delete(remoteName string, email utils.EmailAddressOrWildcard, options DeleteOptions) error {
	var q sq.Sqlizer
	if options.Permanent {
		// Hard delete
		q = sq.
			Delete("remotes_send_grants").
			Where(sq.Expr("remote_id = (?)", sq.
				Select("ID").
				From("remotes").
				Where(sq.Eq{
					"name":       remoteName,
					"deleted_at": nil,
				}),
			)).
			Where(sq.Expr("domain_id = (?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{
					"fqdn":       email.DomainFQDN,
					"deleted_at": nil,
				}),
			)).
			Where(sq.Eq{
				"name": email.LocalPart,
			})
	} else {
		// Soft delete
		uq := sq.
			Update("remotes_send_grants").
			Set("deleted_at", sq.Expr("NOW()")).
			Where(sq.Expr("remote_id = (?)", sq.
				Select("ID").
				From("remotes").
				Where(sq.Eq{
					"name":       remoteName,
					"deleted_at": nil,
				}),
			)).
			Where(sq.Expr("domain_id = (?)", sq.
				Select("ID").
				From("domains").
				Where(sq.Eq{
					"fqdn":       email.DomainFQDN,
					"deleted_at": nil,
				}),
			)).
			Where(sq.Eq{
				"name": email.LocalPart,
			})

		if !options.Force {
			uq = uq.Where(sq.Eq{
				"deleted_at": nil,
			})
		}

		q = uq
	}

	return Exec(r.r, q, 1)
}

func (r *remotesSendGrantsRepository) Restore(remoteName string, email utils.EmailAddressOrWildcard) error {
	q := sq.
		Update("remotes_send_grants").
		Set("deleted_at", nil).
		Where(sq.Expr("remote_id = (?)", sq.
			Select("ID").
			From("remotes").
			Where(sq.Eq{
				"name":       remoteName,
				"deleted_at": nil,
			}),
		)).
		Where(sq.Expr("domain_id = (?)", sq.
			Select("ID").
			From("domains").
			Where(sq.Eq{
				"fqdn":       email.DomainFQDN,
				"deleted_at": nil,
			}),
		)).
		Where(sq.Eq{
			"name": email.LocalPart,
		})

	return Exec(r.r, q, 1)
}
