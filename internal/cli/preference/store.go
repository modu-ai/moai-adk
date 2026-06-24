package preference

import (
	"errors"
)

// Tier identifies which storage layer an entry was found in during a Get.
// It is returned by Store.Get so callers can observe the cascade behavior
// (AC-ADM-004).
type Tier int

const (
	// TierNone signals "not found" — Get returned ok=false.
	TierNone Tier = iota
	// TierCore: the always-loaded user-profile layer (core.yaml, ≤4KB). A core
	// hit MUST short-circuit the cascade (recall/archival not accessed).
	TierCore
	// TierRecall: the recent-session layer (recall.jsonl). Hit only on a core
	// miss.
	TierRecall
	// TierArchival: the full-search layer (archival/). Hit only on core+recall
	// miss.
	TierArchival
)

// String renders the Tier for diagnostics / logging.
func (t Tier) String() string {
	switch t {
	case TierCore:
		return "core"
	case TierRecall:
		return "recall"
	case TierArchival:
		return "archival"
	default:
		return "none"
	}
}

// ErrNotFound is returned indirectly (via ok=false) by Store.Get when no entry
// matches the (domain, decision_key) pair across all tiers. It is exported for
// callers that prefer errors.Is over the ok boolean.
var ErrNotFound = errors.New("preference: entry not found")

// Store is the preference memory layer contract (REQ-ADM-001, REQ-ADM-004).
//
// All implementations MUST be safe for sequential use within a single process.
// Cross-process atomicity is provided by the file-store implementation via
// temp-file-then-rename per recall.jsonl write; concurrent writers from
// separate processes follow last-writer-wins (REQ-ADM-009 advisory, accepted
// race per design.md §C cohabitation).
type Store interface {
	// Upsert stores entry under (domain, decisionKey), REPLACING any prior
	// entry for the same key rather than appending (REQ-ADM-001). The entry
	// is validated first (AC-ADM-003); an invalid entry returns
	// ErrInvalidEntry and is not persisted. The write is atomic w.r.t.
	// SIGKILL (AC-ADM-001 edge case).
	Upsert(domain, decisionKey string, entry Entry) error

	// Get retrieves the entry for (domain, decisionKey) via the 3-tier
	// cascade: core → recall → archival (REQ-ADM-004). On a core hit, recall
	// and archival MUST NOT be accessed. Returns (entry, true, tier, nil) on
	// a hit; (zero, false, TierNone, nil) on a miss.
	Get(domain, decisionKey string) (Entry, bool, Tier, error)

	// Query returns every entry whose Domain field matches, across all tiers.
	// The cascade order in the returned slice is core-first, then recall, then
	// archival, but callers SHOULD treat Query as a namespace-scoped scan,
	// not an ordered lookup.
	Query(domain string) ([]Entry, error)
}
