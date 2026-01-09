package test

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestPostfixVirtualAlias(t *testing.T) {
	const maxDepth = 45

	t.Run("ExplicitAliases", func(t *testing.T) {
		maxSeenDepth := 0

		for _, a := range fixtures.Aliases {
			dFQDN, _, dEnabled, dDeletedAt, ok := lookupDomain(a.DomainID)
			if !ok {
				t.Fatalf("domain %d not found for alias %d", a.DomainID, a.ID)
			}

			var expectedResults []string
			if dEnabled && !dDeletedAt.Valid {
				var depth int
				if a.Enabled && !a.DeletedAt.Valid {
					expectedResults, depth = buildExpectedPostfixVirtualAlias(a.ID, a.DomainID, maxDepth)
				} else {
					expectedResults, depth = buildExpectedVirtualAliasForDomain(a.DomainID, maxDepth)
				}
				maxSeenDepth = max(maxSeenDepth, depth)
			}

			assertPostfixVirtualAlias(t, dFQDN, a.Name, maxDepth, expectedResults)
		}

		t.Logf("Max seen depth: %d", maxSeenDepth)
	})

	t.Run("CatchAll", func(t *testing.T) {
		const name = "nonexistent"
		var maxSeenDepth int

		for _, d := range fixtures.DomainsManaged {
			var expectedResults []string
			var depth int
			if d.Enabled && !d.DeletedAt.Valid {
				expectedResults, depth = buildExpectedVirtualAliasForDomain(d.ID, maxDepth)
				maxSeenDepth = max(maxSeenDepth, depth)
			}

			assertPostfixVirtualAlias(t, d.FQDN, name, maxDepth, expectedResults)
		}
		t.Logf("Max seen depth (managed domains): %d", maxSeenDepth)

		for _, d := range fixtures.DomainsRelayed {
			var expectedResults []string
			var depth int
			if d.Enabled && !d.DeletedAt.Valid {
				expectedResults, depth = buildExpectedVirtualAliasForDomain(d.ID, maxDepth)
				maxSeenDepth = max(maxSeenDepth, depth)
			}

			assertPostfixVirtualAlias(t, d.FQDN, name, maxDepth, expectedResults)
		}
		t.Logf("Max seen depth (relayed domains): %d", maxSeenDepth)

		for _, d := range fixtures.DomainsAlias {
			var expectedResults []string
			var depth int
			if d.Enabled && !d.DeletedAt.Valid {
				expectedResults, depth = buildExpectedVirtualAliasForDomain(d.ID, maxDepth)
				maxSeenDepth = max(maxSeenDepth, depth)
			}

			assertPostfixVirtualAlias(t, d.FQDN, name, maxDepth, expectedResults)
		}
		t.Logf("Max seen depth (alias domains): %d", maxSeenDepth)
	})
}

func assertPostfixVirtualAlias(t *testing.T, fqdn, name string, maxDepth int, expectedResults []string) {
	t.Helper()

	rows, err := sq.
		Select("result").
		Suffix("FROM postfix.virtual_alias_maps(?, ?, ?)", fqdn, name, maxDepth).
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
		t.Fatalf("unexpected row count for %s@%s: got %d want %d (got=%v, want=%v)", name, fqdn, len(got), len(expectedResults), got, expectedResults)
	}

	for i := range got {
		if got[i] != expectedResults[i] {
			t.Fatalf("unexpected row %d for %s@%s: got %q want %q", i, name, fqdn, got[i], expectedResults[i])
		}
	}
}

func buildExpectedPostfixVirtualAlias(recipientID int, domainID int, maxDepth int) ([]string, int) {
	// First get all normal virtual alias results for the recipient
	normalResults, normalMaxSeenDepth := buildExpectedVirtualAliasForRecipients([]int{recipientID}, maxDepth)

	// Collect all catch-all virtual alias results for the recipient, which will be always included
	catchAllAlwaysResults, catchAllAlwaysMaxSeenDepth := buildExpectedVirtualAliasForCatchAll(domainID, false, maxDepth)

	// Also include fallback-only catch-all results if no normal results exist
	var catchAllFallbackOnlyResults []string
	var catchAllFallbackOnlyMaxSeenDepth int
	if len(normalResults) == 0 {
		catchAllFallbackOnlyResults, catchAllFallbackOnlyMaxSeenDepth = buildExpectedVirtualAliasForCatchAll(domainID, true, maxDepth)
	}

	// Merge all result sets
	recipients := make(map[string]struct{}, len(normalResults)+len(catchAllAlwaysResults)+len(catchAllFallbackOnlyResults))
	for _, r := range normalResults {
		recipients[r] = struct{}{}
	}
	for _, r := range catchAllAlwaysResults {
		recipients[r] = struct{}{}
	}
	for _, r := range catchAllFallbackOnlyResults {
		recipients[r] = struct{}{}
	}

	// Convert to slice
	expectedResults := make([]string, 0, len(recipients))
	for r := range recipients {
		expectedResults = append(expectedResults, r)
	}

	sort.Strings(expectedResults)

	return expectedResults, max(normalMaxSeenDepth, max(catchAllAlwaysMaxSeenDepth, catchAllFallbackOnlyMaxSeenDepth))
}

func buildExpectedVirtualAliasForRecipients(recipientIDs []int, maxDepth int) ([]string, int) {
	recipients := make(map[string]struct{})

	type queueItem struct {
		RecipientID int
		Depth       int
	}

	queue := []queueItem{}
	seen := make(map[int]struct{})
	maxSeenDepth := 0
	for _, rid := range recipientIDs {
		queue = append(queue, queueItem{
			RecipientID: rid,
			Depth:       0,
		})
	}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.Depth > maxDepth {
			continue // exceeded max depth
		}
		if _, ok := seen[item.RecipientID]; ok {
			continue // already seen
		}

		// Mark as seen
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
			// Enqueue all targets of this alias which allow forwarding and are not soft-deleted
			for _, r := range fixtures.AliasesTargetsRecursive {
				if r.AliasID == item.RecipientID && r.Forwarding && !r.DeletedAt.Valid {
					queue = append(queue, queueItem{
						RecipientID: r.RecipientID,
						Depth:       item.Depth + 1,
					})
				}
			}
			// Collect all foreign alias targets
			for _, f := range fixtures.AliasesTargetsForeign {
				if f.AliasID == item.RecipientID && f.Forwarding && !f.DeletedAt.Valid {
					recipients[fmt.Sprintf("%s@%s", f.Name, f.FQDN)] = struct{}{}
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
			if !d.Enabled || d.DeletedAt.Valid || !m.ReceivingEnabled || m.DeletedAt.Valid {
				continue // skip disabled or soft-deleted domains or mailboxes
			}
			recipients[fmt.Sprintf("%s@%s", m.Name, d.FQDN)] = struct{}{}
			continue
		}

		// Check if it's a relayed recipient
		if r, ok := fixtures.RecipientsRelayed[item.RecipientID]; ok {
			d, ok := fixtures.DomainsRelayed[r.DomainID]
			if !ok {
				continue // domain not found
			}
			if !d.Enabled || d.DeletedAt.Valid || !r.Enabled || r.DeletedAt.Valid {
				continue // skip disabled or soft-deleted domains or relayed recipients
			}
			recipients[fmt.Sprintf("%s@%s", r.Name, d.FQDN)] = struct{}{}
		}
	}

	// Convert to slice
	expectedResults := make([]string, 0, len(recipients))
	for r := range recipients {
		expectedResults = append(expectedResults, r)
	}

	sort.Strings(expectedResults)

	return expectedResults, maxSeenDepth
}

func buildExpectedVirtualAliasForDomain(domainID int, maxDepth int) ([]string, int) {
	// Determine all catch-all targets for the domain
	var recipientIDs []int
	for _, dct := range fixtures.DomainsCatchallTargets {
		// Only consider enabled and not soft-deleted catch-all targets
		if dct.DomainID == domainID && dct.Forwarding && !dct.DeletedAt.Valid {
			recipientIDs = append(recipientIDs, dct.Recipient)
		}
	}

	// The max depth is reduced by 1, because the catch-all are resolved in the SQL function in level 1
	return buildExpectedVirtualAliasForRecipients(recipientIDs, maxDepth-1)
}

func buildExpectedVirtualAliasForCatchAll(domainID int, fallbackOnly bool, maxDepth int) ([]string, int) {
	// Determine all catch-all targets for the domain
	var recipientIDs []int
	for _, dct := range fixtures.DomainsCatchallTargets {
		// Only consider enabled and not soft-deleted catch-all targets
		if dct.DomainID == domainID && dct.Forwarding && !dct.DeletedAt.Valid && dct.FallbackOnly == fallbackOnly {
			recipientIDs = append(recipientIDs, dct.Recipient)
		}
	}

	return buildExpectedVirtualAliasForRecipients(recipientIDs, maxDepth)
}
