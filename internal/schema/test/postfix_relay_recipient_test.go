package test

import (
	"database/sql"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestPostfixRelayRecipient(t *testing.T) {
	for _, r := range fixtures.RecipientsRelayed {
		d, ok := fixtures.DomainsRelayed[r.DomainID]
		if !ok {
			t.Fatalf("relayed domain %d not found", r.DomainID)
		}

		// Row should exist if domain and recipient are enabled and not soft-deleted
		expectRow := d.Enabled && !d.DeletedAt.Valid && r.Enabled && !r.DeletedAt.Valid

		assertPostfixRelayRecipient(t, d.FQDN, r.Name, expectRow)
	}
}

func assertPostfixRelayRecipient(t *testing.T, fqdn, name string, expectRow bool) {
	t.Helper()

	var got string

	err := sq.
		Select("result").
		Suffix("FROM postfix.relay_recipient_maps(?, ?)", fqdn, name).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if expectRow {
				t.Fatalf("expected row for %s@%s, got no rows", name, fqdn)
			} else {
				return // expected no row, got none
			}
		}

		t.Fatalf("query %s@%s: %v", name, fqdn, err)
	}

	if !expectRow {
		t.Fatalf("expected no rows for %s@%s, got %+v", name, fqdn, got)
	} else if got != "OK" {
		t.Fatalf("unexpected response for %s@%s: got %q", name, fqdn, got)
	}
}
