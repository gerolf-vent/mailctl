package db

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
)

func doesDomainExist(r sq.BaseRunner, domainFQDN string) (bool, error) {
	var exists int
	err := sq.
		Select("1").
		From("domains").
		Where(sq.Eq{
			"fqdn":       domainFQDN,
			"deleted_at": nil,
		}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func queryDomainIdAndTable(r sq.BaseRunner, fqdn string) (int, string, error) {
	var domainId int

	// Check managed domains
	err := sq.
		Select("ID").
		From("domains_managed").
		Where(sq.Eq{"fqdn": fqdn}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&domainId)
	if err == nil {
		return domainId, "domains_managed", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, "", err
	}

	// Check relayed domains
	err = sq.
		Select("ID").
		From("domains_relayed").
		Where(sq.Eq{"fqdn": fqdn}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&domainId)
	if err == nil {
		return domainId, "domains_relayed", nil
	}

	// Check alias domains
	err = sq.
		Select("ID").
		From("domains_alias").
		Where(sq.Eq{"fqdn": fqdn}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&domainId)
	if err == nil {
		return domainId, "domains_alias", nil
	}

	// Check canonical domains
	err = sq.
		Select("ID").
		From("domains_canonical").
		Where(sq.Eq{"fqdn": fqdn}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&domainId)
	if err == nil {
		return domainId, "domains_canonical", nil
	}

	return 0, "", err
}
