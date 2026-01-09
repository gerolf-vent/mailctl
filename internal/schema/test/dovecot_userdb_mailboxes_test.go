package test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestDovecotUserdbMailboxes(t *testing.T) {
	for _, m := range fixtures.Mailboxes {
		d, ok := fixtures.DomainsManaged[m.DomainID]
		if !ok {
			t.Fatalf("managed domain %d not found", m.DomainID)
		}

		// Row should exist if mailbox and domain are not soft-deleted
		expectRow := !d.DeletedAt.Valid && !m.DeletedAt.Valid

		// Determine expectedRow row
		var expectedRow userdbRow
		if expectRow {
			if m.StorageQuota.Valid {
				expectedRow.QuotaStorageSize.Valid = true
				expectedRow.QuotaStorageSize.String = fmt.Sprintf("%dM", m.StorageQuota.Int32)
			} else {
				expectedRow.QuotaStorageSize.Valid = true
				expectedRow.QuotaStorageSize.String = "0"
			}
		}

		assertUserdbMailbox(t, d.FQDN, m.Name, expectRow, expectedRow)
	}
}

func assertUserdbMailbox(t *testing.T, fqdn, name string, expectRow bool, expected userdbRow) {
	t.Helper()

	var got userdbRow

	err := sq.
		Select("quota_storage_size").
		Suffix("FROM dovecot.userdb_mailboxes(?, ?)", fqdn, name).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got.QuotaStorageSize)

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
	} else if !got.Equals(expected) {
		t.Fatalf("unexpected row for %s@%s: got %+v want %+v", name, fqdn, got, expected)
	}
}
