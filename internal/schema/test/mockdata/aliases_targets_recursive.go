package mockdata

import (
	"database/sql"
	"fmt"
	"sort"

	sq "github.com/Masterminds/squirrel"
)

// AliasesTargetsRecursiveVariant captures recursive alias target config and ID.
type AliasesTargetsRecursiveVariant struct {
	ID          int
	AliasID     int
	RecipientID int
	Forwarding  bool
	Sending     bool
	DeletedAt   sql.NullTime
}

func (b *Builder) seedAliasesTargetsRecursive() error {
	const maxDepth = 50

	combos := []struct {
		forwarding bool
		sending    bool
		deleted    bool
	}{
		{forwarding: false, sending: false, deleted: false},
		{forwarding: false, sending: true, deleted: false},
		{forwarding: true, sending: false, deleted: false},
		{forwarding: true, sending: true, deleted: false},
		{forwarding: false, sending: false, deleted: true},
		{forwarding: false, sending: true, deleted: true},
		{forwarding: true, sending: false, deleted: true},
		{forwarding: true, sending: true, deleted: true},
	}

	recipientIDs := make([]int, 0, len(b.f.Mailboxes)+len(b.f.RecipientsRelayed))
	for _, m := range b.f.Mailboxes {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if m.DeletedAt.Valid {
			continue
		}
		recipientIDs = append(recipientIDs, m.ID)
	}
	for _, r := range b.f.RecipientsRelayed {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if r.DeletedAt.Valid {
			continue
		}
		recipientIDs = append(recipientIDs, r.ID)
	}

	if len(recipientIDs) == 0 {
		return fmt.Errorf("no recipients available for alias targets")
	}

	// Sort aliases for deterministic behavior
	var aliases []AliasesVariant
	for _, a := range b.f.Aliases {
		// Skip soft-deleted parents to satisfy hook_check_foreign_key_soft_delete
		if a.DeletedAt.Valid {
			continue
		}
		aliases = append(aliases, a)
	}
	sort.Slice(aliases, func(i, j int) bool {
		return aliases[i].ID < aliases[j].ID
	})

	// Identify active and inactive aliases
	var activeAliases []int
	var inactiveAliases []int
	for _, a := range aliases {
		dActive := false
		if d, ok := b.f.DomainsManaged[a.DomainID]; ok {
			dActive = d.Enabled && !d.DeletedAt.Valid
		} else if d, ok := b.f.DomainsRelayed[a.DomainID]; ok {
			dActive = d.Enabled && !d.DeletedAt.Valid
		} else if d, ok := b.f.DomainsAlias[a.DomainID]; ok {
			dActive = d.Enabled && !d.DeletedAt.Valid
		}

		if a.Enabled && !a.DeletedAt.Valid && dActive {
			activeAliases = append(activeAliases, a.ID)
		} else {
			inactiveAliases = append(inactiveAliases, a.ID)
		}
	}

	// Validate we have enough aliases to build chains
	if len(activeAliases) < maxDepth {
		return fmt.Errorf("not enough active aliases to build chains: have %d, want %d", len(activeAliases), maxDepth)
	}
	if len(inactiveAliases) == 0 {
		return fmt.Errorf("not enough active aliases to build chains: have %d, want %d", len(inactiveAliases), maxDepth/2)
	}

	// AliasID -> TargetAliasID (Active -> Active)
	goodChain := make(map[int]int)
	// AliasID -> TargetAliasID (Active -> Active/Inactive)
	brokenChain := make(map[int]int)

	// Build Good Chain: A_50 -> A_49 -> ... -> A_1
	// We use the first `maxDepth` active aliases.
	// A_1 (index 0) will not have a recursive target in this map (it relies on direct mailbox target).
	// A_2 (index 1) -> A_1 (index 0)
	for i := 1; i < maxDepth; i++ {
		source := activeAliases[i]
		target := activeAliases[i-1]
		goodChain[source] = target
	}

	// Build Broken Chain: B_25 -> A_25 -> ... -> B_2 -> A_2 -> B_1 -> A_1
	for i := 1; i < maxDepth; i++ {
		chainLen := maxDepth / 2
		if i%2 == 1 {
			// Odd steps: B_n -> A_n
			n := chainLen - (i-1)/2
			source := inactiveAliases[(n-1)%len(inactiveAliases)]
			target := activeAliases[n-1]
			brokenChain[source] = target
		} else {
			// Even steps: A_n -> B_{n-1}
			n := chainLen - i/2 + 1
			source := activeAliases[n-1]
			target := inactiveAliases[(n-2+len(inactiveAliases))%len(inactiveAliases)]
			brokenChain[source] = target
		}
	}

	pairUsed := make(map[string]bool)
	pairIdx := 0

	var variants []AliasesTargetsRecursiveVariant
	q := sq.
		Insert("aliases_targets_recursive").
		Columns("alias_id", "recipient_id", "forwarding_to_target_enabled", "sending_from_target_enabled", "deleted_at")

	for _, alias := range aliases {
		// --- A. Add Direct Targets (Mailbox/Recipient) ---

		// We add these for ALL aliases to ensure base validity and flag testing.
		for _, combo := range combos {
			recipientID := recipientIDs[pairIdx%len(recipientIDs)]
			pairKey := fmt.Sprintf("%d_%d", alias.ID, recipientID)
			attempts := 0
			for pairUsed[pairKey] && attempts < len(recipientIDs) {
				pairIdx++
				recipientID = recipientIDs[pairIdx%len(recipientIDs)]
				pairKey = fmt.Sprintf("%d_%d", alias.ID, recipientID)
				attempts++
			}
			pairUsed[pairKey] = true
			pairIdx++

			if recipientID == 0 {
				return fmt.Errorf("recipientID is 0")
			}

			q = q.Values(alias.ID, recipientID, combo.forwarding, combo.sending, b.nullTime(combo.deleted))
			variants = append(variants, AliasesTargetsRecursiveVariant{
				AliasID:     alias.ID,
				RecipientID: recipientID,
				Forwarding:  combo.forwarding,
				Sending:     combo.sending,
				DeletedAt:   b.nullTime(combo.deleted),
			})
		}

		// --- B. Add Recursive Targets ---

		// Check if this alias is part of our constructed chains
		var targetID int
		var isChainLink bool

		if tID, ok := goodChain[alias.ID]; ok {
			targetID = tID
			isChainLink = true
		} else if tID, ok := brokenChain[alias.ID]; ok {
			targetID = tID
			isChainLink = true
		}

		// Track used recursive targets for this alias to avoid PK violation
		usedRecursiveTargets := make(map[int]bool)

		if isChainLink {
			q = q.Values(alias.ID, targetID, true, true, nil)
			variants = append(variants, AliasesTargetsRecursiveVariant{
				AliasID:     alias.ID,
				RecipientID: targetID,
				Forwarding:  true,
				Sending:     true,
				DeletedAt:   sql.NullTime{Valid: false},
			})
			usedRecursiveTargets[targetID] = true
		}
	}

	ids, err := b.insertIDs(q)
	if err != nil {
		return err
	}

	for i, id := range ids {
		variants[i].ID = id
		b.f.AliasesTargetsRecursive[id] = variants[i]
	}

	return nil
}
