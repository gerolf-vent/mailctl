package test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestPostfixCanonical(t *testing.T) {
	// Use simple name part for testing
	const name = "user"

	for _, d := range fixtures.DomainsCanonical {
		tFQDN, _, tEnabled, tDeletedAt, ok := lookupDomain(d.TargetDomain)
		if !ok {
			t.Fatalf("domain %d not found for canonical target", d.TargetDomain)
		}

		// Row should exist if canonincal domain and target domain are enabled and not soft-deleted
		expectRow := d.Enabled && !d.DeletedAt.Valid && tEnabled && !tDeletedAt.Valid

		// Determine expected mail address
		expectedResult := ""
		if expectRow {
			expectedResult = fmt.Sprintf("%s@%s", name, tFQDN)
		}

		assertPostfixCanonical(t, d.FQDN, name, expectRow, expectedResult)
	}
}

func assertPostfixCanonical(t *testing.T, fqdn, name string, expectRow bool, expectedResult string) {
	t.Helper()

	var got string

	err := sq.
		Select("result").
		Suffix("FROM postfix.canonical_maps(?, ?)", fqdn, name).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if expectRow {
				t.Fatalf("expected row for %s@%s, got no rows", name, fqdn)
			}
			return // expected no row, got no row
		}

		t.Fatalf("query %s@%s: %v", name, fqdn, err)
	}

	if !expectRow {
		t.Fatalf("expected no rows for %s@%s, got %+v", name, fqdn, got)
	} else if got != expectedResult {
		t.Fatalf("unexpected row for %s@%s: got %q want %q", name, fqdn, got, expectedResult)
	}
}
