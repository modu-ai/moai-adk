package curator

import "strings"

// --- 2-layer Recall contract (REQ-HEV2-015..018) ---
//
// The Curator's memory obeys a 2-layer recall contract, formalized here as types
// + godoc. This file declares the contract ONLY; the CONSUMPTION wiring (a live
// search implementation over the routing ledger + lineage surfaces) is deferred
// to EVOLVE-005.
//
// Layer 1 — the DIGEST layer (REQ-HEV2-015): the always-loaded
// MOAI:LEARNED-WORKFLOW block. It carries SUMMARY-ONLY distilled workflow
// knowledge — one short bullet per learned pattern, never the full evidence
// trail. Its size is bounded (the digest-budget + bullet-cap enforced in
// budget.go) precisely because it is loaded on every turn.
//
// Layer 2 — the LEDGER layer (REQ-HEV2-016): the searchable evidence surfaces
// (the EVOLVE-001 routing ledger + the lineage audit trail). It is NOT
// always-loaded; it is SEARCHED on demand when a digest summary's backing
// evidence is needed.
//
// Cross-layer linkage (REQ-HEV2-017): each digest bullet carries a ledger_key
// (rendered as a trailing "<!-- key: ... -->" HTML comment by the writer) that
// links a summary in the digest layer to its evidence entry in the ledger layer.
// A provisional (early-tier) bullet may carry an empty ledger_key
// (evidence-or-null); a later-tier promotion replaces it with a real key.
//
// Principle (REQ-HEV2-018): "remember everything ✗, search when needed ○". The
// digest does not attempt to remember everything; it remembers summaries and
// defers to a search over the ledger layer when full evidence is needed.
// Consequently there is deliberately NO WriteFullEvidenceToDigest code path —
// writing a full evidence trail into the always-loaded digest is the exact
// anti-pattern this contract forbids.

// RecallLayer identifies one of the two layers of the recall contract.
type RecallLayer int

const (
	// DigestLayer is the always-loaded, summary-only layer (the
	// MOAI:LEARNED-WORKFLOW block) — REQ-HEV2-015.
	DigestLayer RecallLayer = iota
	// LedgerLayer is the searched-on-demand evidence layer (the routing ledger
	// + lineage surfaces) — REQ-HEV2-016.
	LedgerLayer
)

// String renders the layer name.
func (l RecallLayer) String() string {
	switch l {
	case DigestLayer:
		return "digest"
	case LedgerLayer:
		return "ledger"
	default:
		return "unknown"
	}
}

// SummaryOnly reports whether the layer is summary-only. Only the digest layer
// is summary-only by contract (REQ-HEV2-015 / REQ-HEV2-018); the ledger layer
// carries the full evidence trail.
func (l RecallLayer) SummaryOnly() bool {
	return l == DigestLayer
}

// DigestEntry is a single summary-only entry in the digest layer. It carries a
// distilled Summary and a LedgerKey linking to the ledger-layer evidence
// (REQ-HEV2-017). It deliberately has NO field for a full evidence body — the
// contract stores summaries in the digest and defers full evidence to the ledger
// layer.
type DigestEntry struct {
	// Summary is the distilled generic workflow knowledge (one short line).
	Summary string
	// LedgerKey links this summary to its ledger-layer evidence entry
	// (REQ-HEV2-017). Empty for a provisional (early-tier) entry
	// (evidence-or-null).
	LedgerKey string
}

// Provisional reports whether the entry has no ledger linkage yet — an
// early-tier observation whose evidence key is not assigned (REQ-HEV2-017
// evidence-or-null).
func (e DigestEntry) Provisional() bool {
	return strings.TrimSpace(e.LedgerKey) == ""
}

// LedgerSearcher is the ledger-layer search interface (REQ-HEV2-016). The digest
// layer holds summaries; when the backing evidence for a summary is needed, the
// ledger layer is SEARCHED by ledger_key rather than the evidence being
// duplicated into the always-loaded digest. The concrete implementation (over
// the EVOLVE-001 routing ledger + lineage surfaces) is EVOLVE-005 scope; this
// contract declares the interface only.
type LedgerSearcher interface {
	// SearchByKey returns the ledger-layer evidence references linked to the
	// given ledger_key. An empty key yields no results (a provisional digest
	// entry has no ledger evidence to search).
	SearchByKey(ledgerKey string) []string
}

// RecallContract is the type-level statement of the 2-layer contract. It names
// the summary-only digest layer and the searchable ledger layer, and exposes the
// searchable layer for callers that need to resolve a digest summary to its
// evidence.
type RecallContract struct{}

// Digest returns the summary-only layer (REQ-HEV2-015).
func (RecallContract) Digest() RecallLayer { return DigestLayer }

// Ledger returns the searchable evidence layer (REQ-HEV2-016).
func (RecallContract) Ledger() RecallLayer { return LedgerLayer }

// SearchableLayer returns the layer that is searched on demand (the ledger
// layer). The digest layer is always-loaded and never searched — keeping search
// off the always-loaded surface is the whole point of the contract
// (REQ-HEV2-018).
func (RecallContract) SearchableLayer() RecallLayer { return LedgerLayer }
