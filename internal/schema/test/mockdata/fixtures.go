package mockdata

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// Fixture collects all inserted rows keyed by a simple integer variant index.
type Fixture struct {
	Transports              map[int]TransportsVariant
	Remotes                 map[int]RemotesVariant
	DomainsManaged          map[int]DomainsManagedVariant
	DomainsRelayed          map[int]DomainsRelayedVariant
	DomainsAlias            map[int]DomainsAliasVariant
	DomainsCanonical        map[int]DomainsCanonicalVariant
	Mailboxes               map[int]MailboxesVariant
	RecipientsRelayed       map[int]RecipientsRelayedVariant
	Aliases                 map[int]AliasesVariant
	AliasesTargetsRecursive map[int]AliasesTargetsRecursiveVariant
	AliasesTargetsForeign   map[int]AliasesTargetsForeignVariant
	DomainsCatchallTargets  map[int]DomainsCatchallTargetsVariant
	RemotesSendGrants       map[int]RemotesSendGrantsVariant
}

// Builder carries shared state during seeding.
type Builder struct {
	ctx context.Context
	tx  *sql.Tx
	now time.Time
	f   *Fixture
}

// SeedAll inserts every property variant multiplied by all foreign key variants (bounded recursion) and returns the resulting IDs.
func SeedAll(ctx context.Context, tx *sql.Tx) (*Fixture, error) {
	b := &Builder{
		ctx: ctx,
		tx:  tx,
		now: time.Now().UTC(),
		f: &Fixture{
			Transports:              map[int]TransportsVariant{},
			Remotes:                 map[int]RemotesVariant{},
			DomainsManaged:          map[int]DomainsManagedVariant{},
			DomainsRelayed:          map[int]DomainsRelayedVariant{},
			DomainsAlias:            map[int]DomainsAliasVariant{},
			DomainsCanonical:        map[int]DomainsCanonicalVariant{},
			Mailboxes:               map[int]MailboxesVariant{},
			RecipientsRelayed:       map[int]RecipientsRelayedVariant{},
			Aliases:                 map[int]AliasesVariant{},
			AliasesTargetsRecursive: map[int]AliasesTargetsRecursiveVariant{},
			AliasesTargetsForeign:   map[int]AliasesTargetsForeignVariant{},
			DomainsCatchallTargets:  map[int]DomainsCatchallTargetsVariant{},
			RemotesSendGrants:       map[int]RemotesSendGrantsVariant{},
		},
	}

	steps := []func() error{
		b.seedTransports,
		b.seedRemotes,
		b.seedDomainsManaged,
		b.seedDomainsRelayed,
		b.seedDomainsAlias,
		b.seedDomainsCanonical,
		b.seedMailboxes,
		b.seedRelayedRecipients,
		b.seedAliases,
		b.seedAliasesTargetsRecursive,
		b.seedAliasesTargetsForeign,
		b.seedDomainsCatchallTargets,
		b.seedRemotesSendGrants,
	}

	for _, step := range steps {
		if err := step(); err != nil {
			return nil, err
		}
	}

	return b.f, nil
}

// insertIDs runs a squirrel insert with RETURNING ID and returns all IDs.
func (b *Builder) insertIDs(q sq.InsertBuilder) ([]int, error) {
	rows, err := q.
		PlaceholderFormat(sq.Dollar).
		Suffix("RETURNING ID").
		RunWith(b.tx).
		QueryContext(b.ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (b *Builder) nullTime(deleted bool) sql.NullTime {
	if !deleted {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: b.now.Add(-time.Hour), Valid: true}
}
