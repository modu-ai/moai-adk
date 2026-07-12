package curator

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestRecallContract_DigestLayerSummaryOnly verifies the digest layer is the
// summary-only layer of the 2-layer recall contract (AC-HEV2-021 / REQ-HEV2-015).
func TestRecallContract_DigestLayerSummaryOnly(t *testing.T) {
	t.Parallel()

	// The digest layer is summary-only; the ledger layer carries full evidence.
	if !DigestLayer.SummaryOnly() {
		t.Errorf("DigestLayer.SummaryOnly() = false, want true (digest is the summary-only layer)")
	}
	if LedgerLayer.SummaryOnly() {
		t.Errorf("LedgerLayer.SummaryOnly() = true, want false (ledger carries full evidence)")
	}

	// Layer name rendering (incl. the default branch for an unknown layer value).
	if got := DigestLayer.String(); got != "digest" {
		t.Errorf("DigestLayer.String() = %q, want %q", got, "digest")
	}
	if got := LedgerLayer.String(); got != "ledger" {
		t.Errorf("LedgerLayer.String() = %q, want %q", got, "ledger")
	}
	if got := RecallLayer(99).String(); got != "unknown" {
		t.Errorf("RecallLayer(99).String() = %q, want %q", got, "unknown")
	}

	// The contract names the digest layer as its summary-only surface.
	if (RecallContract{}).Digest() != DigestLayer {
		t.Errorf("RecallContract.Digest() did not return DigestLayer")
	}

	// A digest entry with a ledger_key links to the ledger layer; an empty
	// ledger_key marks a provisional (early-tier) entry (evidence-or-null).
	linked := DigestEntry{Summary: "prefer specific-path git add over -A", LedgerKey: "k-42"}
	if linked.Provisional() {
		t.Errorf("DigestEntry with LedgerKey should not be provisional")
	}
	provisional := DigestEntry{Summary: "single early observation", LedgerKey: ""}
	if !provisional.Provisional() {
		t.Errorf("DigestEntry with empty LedgerKey should be provisional")
	}
}

// fakeLedgerSearcher is a minimal LedgerSearcher used to exercise the
// ledger-layer search interface without the (EVOLVE-005) live implementation.
type fakeLedgerSearcher struct {
	byKey map[string][]string
}

func (f fakeLedgerSearcher) SearchByKey(ledgerKey string) []string {
	if ledgerKey == "" {
		return nil
	}
	return f.byKey[ledgerKey]
}

// TestRecallContract_LedgerLayerSearchInterface verifies the ledger layer is a
// search interface over the evidence surfaces (AC-HEV2-022 / REQ-HEV2-016). The
// digest holds summaries; the ledger is SEARCHED by ledger_key on demand rather
// than the evidence being duplicated into the always-loaded digest.
func TestRecallContract_LedgerLayerSearchInterface(t *testing.T) {
	t.Parallel()

	// The contract's searchable layer is the ledger layer, not the digest.
	if (RecallContract{}).SearchableLayer() != LedgerLayer {
		t.Errorf("RecallContract.SearchableLayer() did not return LedgerLayer")
	}
	if (RecallContract{}).Ledger() != LedgerLayer {
		t.Errorf("RecallContract.Ledger() did not return LedgerLayer")
	}

	// A LedgerSearcher resolves a digest summary's ledger_key to its evidence.
	var searcher LedgerSearcher = fakeLedgerSearcher{byKey: map[string][]string{
		"k-42": {".moai/state/routing-ledger.jsonl:117", "lineage:entry-9"},
	}}
	got := searcher.SearchByKey("k-42")
	if len(got) != 2 {
		t.Fatalf("SearchByKey(k-42) returned %d evidence refs, want 2", len(got))
	}

	// A provisional (empty-key) digest entry has no ledger evidence to search.
	if refs := searcher.SearchByKey(""); refs != nil {
		t.Errorf("SearchByKey(\"\") = %v, want nil (a provisional entry has no ledger evidence)", refs)
	}
}

// TestRecallContract_NoWriteFullEvidencePath verifies the principle
// "remember everything ✗, search when needed ○" (AC-HEV2-024 / REQ-HEV2-018):
// the digest is summary-only, and there is deliberately NO code path that writes
// a full evidence trail into the always-loaded digest.
func TestRecallContract_NoWriteFullEvidencePath(t *testing.T) {
	t.Parallel()

	// Principle: the always-loaded digest layer never carries full evidence.
	if !DigestLayer.SummaryOnly() {
		t.Errorf("principle violated: DigestLayer must be summary-only")
	}

	// Structural guard: no function named WriteFullEvidenceToDigest may exist in
	// the curator package. A call/definition carries an opening paren
	// (WriteFullEvidenceToDigest(...)); the godoc mentions the identifier only in
	// prose (no paren), so this scan does not flag the documentation.
	forbidden := regexp.MustCompile(`WriteFullEvidenceToDigest\s*\(`)
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading curator package dir: %v", err)
	}
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, readErr := os.ReadFile(filepath.Clean(name))
		if readErr != nil {
			t.Fatalf("reading %s: %v", name, readErr)
		}
		scanned++
		if forbidden.Match(data) {
			t.Errorf("%s defines/calls a WriteFullEvidenceToDigest code path — the digest "+
				"is summary-only; full evidence lives in the ledger layer (REQ-HEV2-018)", name)
		}
	}
	if scanned == 0 {
		t.Fatal("no non-test .go files scanned in curator package — guard did not run")
	}
}
