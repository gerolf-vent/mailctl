package test

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestPostfixSenderLoginsMailboxes(t *testing.T) {
	const maxDepth = 45

	t.Run("DirectMailboxes", func(t *testing.T) {
		// Direct mailbox lookups (no aliases involved)
		for _, m := range fixtures.Mailboxes {
			d, ok := fixtures.DomainsManaged[m.DomainID]
			if !ok {
				t.Fatalf("managed domain %d not found", m.DomainID)
			}

			// Collect all mailboxes, which are enabled and not soft-deleted and belong
			// to a domain which is enabled and not soft-deleted too
			var expectedResults []string
			if d.Enabled && !d.DeletedAt.Valid && m.SendingEnabled && !m.DeletedAt.Valid {
				expectedResults = append(expectedResults, fmt.Sprintf("%s@%s", m.Name, d.FQDN))
			}

			assertPostfixSenderLoginsMailboxes(t, d.FQDN, m.Name, maxDepth, expectedResults)
		}
	})

	t.Run("Aliases", func(t *testing.T) {
		// Alias-based lookups (recursing through alias targets)
		maxSeenDepth := 0
		for _, a := range fixtures.Aliases {
			dFQDN, _, dEnabled, dDeletedAt, ok := lookupDomain(a.DomainID)
			if !ok {
				t.Fatalf("domain %d not found for alias %d", a.DomainID, a.ID)
			}

			// Determine all recursive mailboxes which are allowed to send from this alias,
			// but only if the alias and it's domain are enabled and not soft-deleted
			var expectedResults []string
			var depth int
			if dEnabled && !dDeletedAt.Valid && a.Enabled && !a.DeletedAt.Valid {
				expectedResults, depth = buildExpectedPostfixSenderLoginsMailboxes(a.ID, maxDepth)
				maxSeenDepth = max(maxSeenDepth, depth)
			}

			assertPostfixSenderLoginsMailboxes(t, dFQDN, a.Name, maxDepth, expectedResults)
		}
		t.Logf("Max seen depth: %d", maxSeenDepth)
	})
}

func assertPostfixSenderLoginsMailboxes(t *testing.T, fqdn, name string, maxDepth int, expectedResults []string) {
	t.Helper()

	rows, err := sq.
		Select("result").
		Suffix("FROM postfix.smtpd_sender_login_maps_mailboxes(?, ?, ?)", fqdn, name, maxDepth).
		PlaceholderFormat(sq.Dollar).
		RunWith(testDB).
		Query()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("query %s@%s: %v", name, fqdn, err)
	}

	var got []string
	if rows != nil {
		for rows.Next() {
			var r string
			if err := rows.Scan(&r); err != nil {
				t.Fatalf("scan: %v", err)
			}
			got = append(got, r)
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("close rows: %v", err)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows: %v", err)
		}
	}

	sort.Strings(got)
	sort.Strings(expectedResults)

	if len(got) != len(expectedResults) {
		t.Fatalf("unexpected row count for %s@%s: got %d want %d (rows=%v)", name, fqdn, len(got), len(expectedResults), got)
	}

	for i := range got {
		if got[i] != expectedResults[i] {
			t.Fatalf("unexpected row %d for %s@%s: got %q want %q", i, name, fqdn, got[i], expectedResults[i])
		}
	}
}

func buildExpectedPostfixSenderLoginsMailboxes(recipientID int, maxDepth int) ([]string, int) {
	recipients := make(map[string]struct{})

	type queueItem struct {
		RecipientID int
		Depth       int
	}

	queue := []queueItem{
		{
			RecipientID: recipientID,
			Depth:       0,
		},
	}
	seen := make(map[int]struct{})
	maxSeenDepth := 0

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:] // dequeue

		if item.Depth > maxDepth {
			continue // exceeded max depth
		}
		if _, ok := seen[item.RecipientID]; ok {
			continue // already seen
		}

		// Mark recipient as seen
		seen[item.RecipientID] = struct{}{}
		maxSeenDepth = max(maxSeenDepth, item.Depth)

		// Check if it's an recursive alias
		if a, ok := fixtures.Aliases[item.RecipientID]; ok {
			_, _, dEnabled, dDeletedAt, ok := lookupDomain(a.DomainID)
			if !ok {
				continue // domain not found
			}
			if !dEnabled || dDeletedAt.Valid || !a.Enabled || a.DeletedAt.Valid {
				continue // skip disabled or soft-deleted aliases or domains
			}
			// Enqueue all targets of this alias which allow sending and are not soft-deleted
			for _, r := range fixtures.AliasesTargetsRecursive {
				if r.AliasID == item.RecipientID && r.Sending && !r.DeletedAt.Valid {
					queue = append(queue, queueItem{
						RecipientID: r.RecipientID,
						Depth:       item.Depth + 1,
					})
				}
			}
			continue
		}

		// Check if it's a mailbox
		if m, ok := fixtures.Mailboxes[item.RecipientID]; ok {
			d, ok := fixtures.DomainsManaged[m.DomainID]
			if !ok {
				continue // domain not found
			}
			if !d.Enabled || d.DeletedAt.Valid || !m.SendingEnabled || m.DeletedAt.Valid {
				continue // skip disabled or soft-deleted domains or mailboxes
			}
			recipients[fmt.Sprintf("%s@%s", m.Name, d.FQDN)] = struct{}{}
		}

		// Relayed recipients are not checked, as they cannot send mail
	}

	expectedResults := make([]string, 0, len(recipients))
	for r := range recipients {
		expectedResults = append(expectedResults, r)
	}

	sort.Strings(expectedResults)

	return expectedResults, maxSeenDepth
}
