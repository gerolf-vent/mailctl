package test

import (
	"database/sql"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

// Helper assertions used by the split stalwart tests.
func assertSingleStringColumn(t *testing.T, suffix, param string, expectRow bool, expected string) {
	t.Helper()

	var got sql.NullString
	err := sq.
		Select("*").
		Suffix("FROM "+suffix, param).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if expectRow {
				t.Fatalf("expected row for %s, got no rows", param)
			}
			return
		}
		t.Fatalf("query: %v", err)
	}

	if !expectRow {
		t.Fatalf("expected no rows for %s, got %v", param, got)
	}

	if !got.Valid || got.String != expected {
		t.Fatalf("unexpected value for %s: got %+v want %q", param, got, expected)
	}
}

func assertSecretColumn(t *testing.T, param string, expectRow bool, expected sql.NullString) {
	t.Helper()

	var got sql.NullString
	err := sq.
		Select("secret").
		Suffix("FROM stalwart.secrets(?)", param).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if expectRow {
				t.Fatalf("expected row for %s, got no rows", param)
			}
			return
		}
		t.Fatalf("query: %v", err)
	}

	if !expectRow {
		t.Fatalf("expected no rows for %s, got %+v", param, got)
	}

	if got.Valid != expected.Valid || (got.Valid && got.String != expected.String) {
		t.Fatalf("unexpected secret for %s: got %+v want %+v", param, got, expected)
	}
}
