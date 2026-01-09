package test

import (
	"database/sql"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestDovecotPassdbMailboxes(t *testing.T) {
	for _, m := range fixtures.Mailboxes {
		d, ok := fixtures.DomainsManaged[m.DomainID]
		if !ok {
			t.Fatalf("managed domain %d not found", m.DomainID)
		}

		// Row should exist if mailbox and domain are not soft-deleted
		expectRow := !d.DeletedAt.Valid && !m.DeletedAt.Valid

		// Determine expected row
		var expectedRow passdbRow
		if expectRow {
			switch {
			case !m.LoginEnabled || !d.Enabled:
				// Login is disabled, either by mailbox or domain
				expectedRow = passdbRow{Password: sql.NullString{}, NoLogin: sql.NullBool{Bool: true, Valid: true}, Reason: sql.NullString{String: "Login is disabled.", Valid: true}}
			case !m.PasswordHash.Valid:
				// No password set for mailbox
				expectedRow = passdbRow{Password: sql.NullString{}, NoLogin: sql.NullBool{Bool: true, Valid: true}, Reason: sql.NullString{String: "No password set.", Valid: true}}
			default:
				// Valid mailbox with password
				expectedRow = passdbRow{Password: m.PasswordHash, NoLogin: sql.NullBool{}, Reason: sql.NullString{}}
			}
		}

		assertDovecotPassdbMailbox(t, d.FQDN, m.Name, expectRow, expectedRow)
	}
}

func assertDovecotPassdbMailbox(t *testing.T, fqdn, name string, expectRow bool, expectedRow passdbRow) {
	t.Helper()

	var got passdbRow

	err := sq.
		Select("password", "nologin", "reason").
		Suffix("FROM dovecot.passdb_mailboxes(?, ?)", fqdn, name).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got.Password, &got.NoLogin, &got.Reason)

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
	} else if !got.Equals(expectedRow) {
		t.Fatalf("unexpected row for %s@%s: got %+v want %+v", name, fqdn, got, expectedRow)
	}
}
