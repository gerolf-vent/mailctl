package test

import (
	"database/sql"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestPostfixVirtualAliasDomains(t *testing.T) {
	for _, d := range fixtures.DomainsAlias {
		// Row should exist if domain is enabled and not soft-deleted
		expectRow := d.Enabled && !d.DeletedAt.Valid

		assertPostfixVirtualAliasDomain(t, d.FQDN, expectRow)
	}
}

func assertPostfixVirtualAliasDomain(t *testing.T, fqdn string, expectRow bool) {
	t.Helper()

	var got string

	err := sq.
		Select("result").
		Suffix("FROM postfix.virtual_alias_domains(?)", fqdn).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if expectRow {
				t.Fatalf("expected row for %s, got no rows", fqdn)
			} else {
				return // expected no row, got none
			}
		}

		t.Fatalf("query %s: %v", fqdn, err)
	}

	if !expectRow {
		t.Fatalf("expected no rows for %s, got %+v", fqdn, got)
	} else if got != "OK" {
		t.Fatalf("unexpected response for %s: got %q", fqdn, got)
	}
}
