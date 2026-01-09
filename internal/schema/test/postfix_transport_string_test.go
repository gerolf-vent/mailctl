package test

import (
	"database/sql"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestPostfixTransportString(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		host     string
		port     sql.NullInt32
		mx       bool
		expected string
	}{
		{name: "no_mx_no_port", method: "smtp", host: "mx.example", port: sql.NullInt32{}, mx: false, expected: "smtp:[mx.example]"},
		{name: "no_mx_port", method: "smtp", host: "mx.example", port: sql.NullInt32{Int32: 25, Valid: true}, mx: false, expected: "smtp:[mx.example]:25"},
		{name: "mx_port", method: "smtp", host: "mx.example", port: sql.NullInt32{Int32: 2525, Valid: true}, mx: true, expected: "smtp:mx.example:2525"},
		{name: "mx_no_port", method: "lmtp", host: "lmtp.example", port: sql.NullInt32{}, mx: true, expected: "lmtp:lmtp.example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlStr, args, err := sq.
				Expr("SELECT postfix.transport_string($1, $2, $3, $4)", tt.method, tt.host, tt.port, tt.mx).
				ToSql()
			if err != nil {
				t.Fatalf("build query: %v", err)
			}

			var got string
			if err := testDB.QueryRow(sqlStr, args...).Scan(&got); err != nil {
				t.Fatalf("query: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("unexpected transport string: got %q want %q", got, tt.expected)
			}
		})
	}
}
