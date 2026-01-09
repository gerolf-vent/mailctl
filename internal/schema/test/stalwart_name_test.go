package test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

type stalwartNameRow struct {
	Name        sql.NullString
	Type        sql.NullString
	Email       sql.NullString
	Secret      sql.NullString
	Description sql.NullString
	Quota       sql.NullInt64
}

func (r *stalwartNameRow) Equals(o stalwartNameRow) bool {
	if r.Name.Valid != o.Name.Valid || (r.Name.Valid && r.Name.String != o.Name.String) {
		return false
	}
	if r.Type.Valid != o.Type.Valid || (r.Type.Valid && r.Type.String != o.Type.String) {
		return false
	}
	if r.Email.Valid != o.Email.Valid || (r.Email.Valid && r.Email.String != o.Email.String) {
		return false
	}
	if r.Secret.Valid != o.Secret.Valid || (r.Secret.Valid && r.Secret.String != o.Secret.String) {
		return false
	}
	if r.Description.Valid != o.Description.Valid || (r.Description.Valid && r.Description.String != o.Description.String) {
		return false
	}
	if r.Quota.Valid != o.Quota.Valid || (r.Quota.Valid && r.Quota.Int64 != o.Quota.Int64) {
		return false
	}
	return true
}

func TestStalwartName(t *testing.T) {
	for _, m := range fixtures.Mailboxes {
		d, ok := fixtures.DomainsManaged[m.DomainID]
		if !ok {
			t.Fatalf("managed domain %d not found", m.DomainID)
		}

		full := fmt.Sprintf("%s@%s", m.Name, d.FQDN)

		// Row should exist only if domain is enabled and not soft-deleted
		// and mailbox receiving is enabled and not soft-deleted
		expectRow := d.Enabled && !d.DeletedAt.Valid && m.ReceivingEnabled && !m.DeletedAt.Valid

		var expected stalwartNameRow
		if expectRow {
			expected.Name = sql.NullString{String: full, Valid: true}
			expected.Type = sql.NullString{String: "individual", Valid: true}
			expected.Email = sql.NullString{String: full, Valid: true}
			expected.Secret = m.PasswordHash
			expected.Description = sql.NullString{String: "", Valid: true}
			if m.StorageQuota.Valid {
				expected.Quota = sql.NullInt64{Int64: int64(m.StorageQuota.Int32) * 1024 * 1024, Valid: true}
			}
		}

		assertStalwartName(t, full, expectRow, expected)
	}
}

func assertStalwartName(t *testing.T, name string, expectRow bool, expected stalwartNameRow) {
	t.Helper()

	var got stalwartNameRow

	err := sq.
		Select("name", "type", "email", "secret", "description", "quota").
		Suffix("FROM stalwart.name(?)", name).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got.Name, &got.Type, &got.Email, &got.Secret, &got.Description, &got.Quota)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if expectRow {
				t.Fatalf("expected row for %s, got no rows", name)
			}
			return // expected no row, got no row
		}
		t.Fatalf("query: %v", err)
	}

	if !expectRow {
		t.Fatalf("expected no rows for %s, got %+v", name, got)
	} else if !got.Equals(expected) {
		t.Fatalf("unexpected row for %s: got %+v want %+v", name, got, expected)
	}
}
