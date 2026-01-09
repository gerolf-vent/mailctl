package test

import (
	"database/sql"
	"errors"
	"sort"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/schema/test/mockdata"
)

func TestPostfixSenderLoginsRemotes(t *testing.T) {
	// Group send grants by domain because wildcards might match multiple names
	// inside the same domain
	grantsPerDomain := make(map[int][]mockdata.RemotesSendGrantsVariant)
	for _, grant := range fixtures.RemotesSendGrants {
		if testing.Short() && len(grantsPerDomain) > 10 {
			break // limit number of domains to speed up tests
		}
		grantsPerDomain[grant.DomainID] = append(grantsPerDomain[grant.DomainID], grant)
	}

	// Process each domain separately
	for domainID, grants := range grantsPerDomain {
		dFQDN, _, dEnabled, dDeletedAt, ok := lookupDomain(domainID)
		if !ok {
			t.Fatalf("domain %d not found", domainID)
		}

		names := make(map[string]struct{}) // Use a set to avoid duplicates
		names["simple"] = struct{}{}       // ensure at least one simple name is tested

		// Generate test names based on wildcards in the grants
		for _, grant := range grants {
			// If the grant has no wildcards, just use it as-is
			if !strings.Contains(grant.Name, "%") && !strings.Contains(grant.Name, "_") {
				names[grant.Name] = struct{}{}
				continue
			}

			// Generate test names by replacing wildcards with sample characters
			for _, wc := range []string{"", "user", "t1234+!#"} {
				for _, sc := range []string{"a", "3", "_"} {
					name := strings.ReplaceAll(grant.Name, "%", wc)
					name = strings.ReplaceAll(name, "_", sc)
					names[name] = struct{}{}
				}
			}
		}

		// Now test each generated name
		for name := range names {
			// Determine expected remotes for this name, but only if the domain is enabled
			// and not soft-deleted
			var expectedResults []string
			if dEnabled && !dDeletedAt.Valid {
				for _, grant := range grants {
					remote, ok := fixtures.Remotes[grant.RemoteID]
					if !ok {
						t.Fatalf("remote %d not found", grant.RemoteID)
					}

					if remote.Enabled && !remote.DeletedAt.Valid && !grant.DeletedAt.Valid {
						// Test if the name matches the grant pattern
						if SQLPatternToRegex(grant.Name).MatchString(name) {
							expectedResults = append(expectedResults, remote.Name)
						}
					}
				}
			}

			assertPostfixSenderLoginsRemotes(t, dFQDN, name, expectedResults)
		}
	}
}

func assertPostfixSenderLoginsRemotes(t *testing.T, fqdn, name string, expectedResults []string) {
	t.Helper()

	rows, err := sq.
		Select("result").
		Suffix("FROM postfix.smtpd_sender_login_maps_remotes(?, ?)", fqdn, name).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		Query()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("query: %v", err)
	}

	defer rows.Close()

	var got []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}

	if len(got) != len(expectedResults) {
		t.Fatalf("unexpected number of rows for %s@%s: got %d want %d", name, fqdn, len(got), len(expectedResults))
	}

	sort.Strings(got)
	sort.Strings(expectedResults)

	for i := range got {
		if got[i] != expectedResults[i] {
			t.Fatalf("unexpected row %d for %s@%s: got %q want %q", i, name, fqdn, got[i], expectedResults[i])
		}
	}
}
