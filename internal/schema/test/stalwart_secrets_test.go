package test

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestStalwartSecrets(t *testing.T) {
	for _, m := range fixtures.Mailboxes {
		d, ok := fixtures.DomainsManaged[m.DomainID]
		if !ok {
			t.Fatalf("managed domain %d not found", m.DomainID)
		}

		full := fmt.Sprintf("%s@%s", m.Name, d.FQDN)

		expectRow := d.Enabled && !d.DeletedAt.Valid && m.ReceivingEnabled && !m.DeletedAt.Valid

		var expectedSecret sql.NullString
		if expectRow {
			expectedSecret = m.PasswordHash
		}
		assertSecretColumn(t, full, expectRow, expectedSecret)
	}
}
