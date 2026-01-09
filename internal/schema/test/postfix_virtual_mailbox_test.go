package test

import (
	"database/sql"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestPostfixVirtualMailbox(t *testing.T) {
	for _, m := range fixtures.Mailboxes {
		d, ok := fixtures.DomainsManaged[m.DomainID]
		if !ok {
			t.Fatalf("managed domain %d not found", m.DomainID)
		}

		// Row should exist if mailbox and domain are enabled and not soft-deleted
		expectRow := d.Enabled && !d.DeletedAt.Valid && m.ReceivingEnabled && !m.DeletedAt.Valid

		assertPostfixVirtualMailbox(t, d.FQDN, m.Name, expectRow)
	}
}

func assertPostfixVirtualMailbox(t *testing.T, fqdn, name string, expectRow bool) {
	t.Helper()

	var got string

	err := sq.
		Select("result").
		Suffix("FROM postfix.virtual_mailbox_maps(?, ?)", fqdn, name).
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
