package test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestPostfixTransportsMailboxes(t *testing.T) {
	for _, m := range fixtures.Mailboxes {
		d, ok := fixtures.DomainsManaged[m.DomainID]
		if !ok {
			t.Fatalf("managed domain %d not found", m.DomainID)
		}

		tpID := d.TransportID
		if m.TransportID.Valid {
			tpID = int(m.TransportID.Int64)
		}

		tp, ok := fixtures.Transports[tpID]
		if !ok {
			t.Fatalf("transport %d not found", tpID)
		}

		// Row should exist if mailbox, domain and transport are enabled and not soft-deleted
		expectRow := d.Enabled && !d.DeletedAt.Valid && m.ReceivingEnabled && !m.DeletedAt.Valid && !tp.DeletedAt.Valid

		expected := buildTransportString(tp.Method, tp.Host, tp.Port, tp.MxLookup)

		assertPostfixTransports(t, d.FQDN, m.Name, expectRow, expected)
	}
}

func TestPostfixTransportsRecipientsRelayed(t *testing.T) {
	for _, r := range fixtures.RecipientsRelayed {
		d, ok := fixtures.DomainsRelayed[r.DomainID]
		if !ok {
			t.Fatalf("relayed domain %d not found", r.DomainID)
		}

		tp, ok := fixtures.Transports[d.TransportID]
		if !ok {
			t.Fatalf("transport %d not found", d.TransportID)
		}

		// Row should exist if domain, recipient and transport are enabled and not soft-deleted
		expectRow := d.Enabled && !d.DeletedAt.Valid && r.Enabled && !r.DeletedAt.Valid && !tp.DeletedAt.Valid

		expected := buildTransportString(tp.Method, tp.Host, tp.Port, tp.MxLookup)

		assertPostfixTransports(t, d.FQDN, r.Name, expectRow, expected)
	}
}

func assertPostfixTransports(t *testing.T, fqdn, name string, expectRow bool, expected string) {
	t.Helper()

	var got string

	err := sq.
		Select("result").
		Suffix("FROM postfix.transport_maps(?, ?)", fqdn, name).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if expectRow {
				t.Fatalf("expected row for %s@%s, got no rows", name, fqdn)
			} else {
				return // expected no row, got no row
			}
		}

		t.Fatalf("query: %v", err)
	}

	if !expectRow {
		t.Fatalf("expected no rows for %s@%s, got %+v", name, fqdn, got)
	} else if got != expected {
		t.Fatalf("unexpected row for %s@%s: got %q want %q", name, fqdn, got, expected)
	}
}

func buildTransportString(method, host string, port sql.NullInt32, mx bool) string {
	switch {
	case !mx && !port.Valid:
		return fmt.Sprintf("%s:[%s]", method, host)
	case !mx && port.Valid:
		return fmt.Sprintf("%s:[%s]:%d", method, host, port.Int32)
	case mx && port.Valid:
		return fmt.Sprintf("%s:%s:%d", method, host, port.Int32)
	default:
		return fmt.Sprintf("%s:%s", method, host)
	}
}
