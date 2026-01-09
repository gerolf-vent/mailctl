package db

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

func queryAliasesTargetsIdAndTable(r sq.BaseRunner, aliasEmail, targetEmail utils.EmailAddress) (int, string, error) {
	aliasIdQ := sq.
		Select("a.ID").
		From("aliases a").
		Join("domains d ON a.domain_id = d.ID").
		Where(sq.Eq{
			"d.fqdn": aliasEmail.DomainFQDN,
			"a.name": aliasEmail.LocalPart,
		}).
		Limit(1)

	var targetId int

	// Check recursive targets table
	err := sq.
		Select("atr.ID").
		From("aliases_targets_recursive atr").
		Join("domains d ON a.domain_id = d.ID").
		Where(sq.Expr("atr.alias_id = (?)", aliasIdQ)).
		Where(sq.Eq{
			"d.fqdn":   targetEmail.DomainFQDN,
			"atr.name": targetEmail.LocalPart,
		}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&targetId)
	if err == nil {
		return targetId, "aliases_targets_recursive", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, "", err
	}

	// Check foreign targets table
	err = sq.
		Select("atf.ID").
		From("aliases_targets_foreign atf").
		Where(sq.Expr("atf.alias_id =  (?)", aliasIdQ)).
		Where(sq.Eq{
			"atf.fqdn": targetEmail.DomainFQDN,
			"atf.name": targetEmail.LocalPart,
		}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&targetId)
	if err == nil {
		return targetId, "aliases_targets_foreign", nil
	}

	return 0, "", err
}
