package test

import (
	"database/sql"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestDovecotPassdbRemotes(t *testing.T) {
	for _, r := range fixtures.Remotes {
		// Row should exist if remote is not soft-deleted
		expectRow := !r.DeletedAt.Valid

		// Determine expected row
		var expectedRow passdbRow
		if expectRow {
			switch {
			case !r.Enabled:
				expectedRow = passdbRow{Password: sql.NullString{}, NoLogin: sql.NullBool{Bool: true, Valid: true}, Reason: sql.NullString{String: "Remote is disabled.", Valid: true}}
			case !r.Password.Valid:
				expectedRow = passdbRow{Password: sql.NullString{}, NoLogin: sql.NullBool{Bool: true, Valid: true}, Reason: sql.NullString{String: "No password set.", Valid: true}}
			default:
				expectedRow = passdbRow{Password: r.Password, NoLogin: sql.NullBool{}, Reason: sql.NullString{}}
			}
		}

		assertDovecotPassdbRemote(t, r.Name, expectRow, expectedRow)
	}
}

func assertDovecotPassdbRemote(t *testing.T, name string, expectRow bool, expectedRow passdbRow) {
	t.Helper()

	var got passdbRow

	err := sq.
		Select("password", "nologin", "reason").
		Suffix("FROM dovecot.passdb_remotes(?)", name).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got.Password, &got.NoLogin, &got.Reason)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if expectRow {
				t.Fatalf("expected row for %s, got no rows", name)
			} else {
				return // expected no row, got no row
			}
		}

		t.Fatalf("query: %v", err)
	}

	if !expectRow {
		t.Fatalf("expected no rows for %s, got %+v", name, got)
	} else if !got.Equals(expectedRow) {
		t.Fatalf("unexpected row for %s: got %+v want %+v", name, got, expectedRow)
	}
}
