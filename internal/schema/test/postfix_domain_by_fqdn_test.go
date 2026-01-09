package test

import (
	"database/sql"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

type domainByFQDNRow struct {
	ID        int
	Type      string
	Enabled   bool
	DeletedAt sql.NullTime
}

func (r *domainByFQDNRow) Equals(other domainByFQDNRow) bool {
	return r.ID == other.ID &&
		r.Type == other.Type &&
		r.Enabled == other.Enabled &&
		r.DeletedAt.Valid == other.DeletedAt.Valid
	// Not testing the timestamp value itself, because soft-deletion of a
	// parent could alter it
}

func TestPostfixDomainByFQDN(t *testing.T) {
	for _, d := range fixtures.DomainsManaged {
		assertPostfixDomainByFQDN(t, d.FQDN, domainByFQDNRow{
			ID:        d.ID,
			Type:      "managed",
			Enabled:   d.Enabled,
			DeletedAt: d.DeletedAt,
		})
	}
	for _, d := range fixtures.DomainsRelayed {
		assertPostfixDomainByFQDN(t, d.FQDN, domainByFQDNRow{
			ID:        d.ID,
			Type:      "relayed",
			Enabled:   d.Enabled,
			DeletedAt: d.DeletedAt,
		})
	}
	for _, d := range fixtures.DomainsAlias {
		assertPostfixDomainByFQDN(t, d.FQDN, domainByFQDNRow{
			ID:        d.ID,
			Type:      "alias",
			Enabled:   d.Enabled,
			DeletedAt: d.DeletedAt,
		})
	}
	for _, d := range fixtures.DomainsCanonical {
		assertPostfixDomainByFQDN(t, d.FQDN, domainByFQDNRow{
			ID:        d.ID,
			Type:      "canonical",
			Enabled:   d.Enabled,
			DeletedAt: d.DeletedAt,
		})
	}
}

func assertPostfixDomainByFQDN(t *testing.T, fqdn string, expectedRow domainByFQDNRow) {
	t.Helper()

	var got domainByFQDNRow

	err := sq.
		Select("ID", "type", "enabled", "deleted_at").
		Suffix("FROM postfix.domain_by_fqdn(?)", fqdn).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		QueryRow().
		Scan(&got.ID, &got.Type, &got.Enabled, &got.DeletedAt)
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	if !got.Equals(expectedRow) {
		t.Fatalf("domain mismatch for %s: got %+v want %+v", fqdn, got, expectedRow)
	}
}
