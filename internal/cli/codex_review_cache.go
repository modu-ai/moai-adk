package cli

// codex_review_cache.go — the tree-keyed verdict reuse cache of the Claude
// turn-end review gate (SPEC-GATE-BOTTLENECK-001 REQ-GBN-001): a fresh
// receipt for the resolved scope's key is the same reviewer judging the same
// code, so the gate reuses its verdict instead of re-running the RPC. The
// store, the key shape, and the freshness predicate are the verify package's
// — the same receipt the Codex Stop chain's member 6 already consumes, so
// both harnesses share one cache (REQ-GBN-004 keeps every cache path
// fail-open: the cache is an accelerator, never an authority).

import (
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/modu-ai/moai-adk/internal/verify"
)

// codexReviewCacheTTL bounds the wall-clock freshness of a reused review
// verdict on top of the key comparison. A review verdict is a function of the
// code state and the reviewer version — both in the key — so the bound is
// deliberately looser than the verify-run DefaultTTL (whose checks are
// time-sensitive): the common failure it guards is a reviewer behavior change,
// which the tool_version field already catches.
const codexReviewCacheTTL = 24 * time.Hour

// codexReviewCacheSkips counts the Stops a cache hit saved an RPC on — the
// "skipped-turn count" REQ-GBN-001 records. Diagnostic; never gates.
var codexReviewCacheSkips atomic.Int64

// recordCodexReviewReceipt stores the gate's disposition of this tree state
// so the next Stop over the same key reuses it. stateErr != nil (the key
// itself could not be measured) records nothing — an unkeyed verdict cannot
// be reused safely. A tree-scope fail whose findings are all runtime-config
// drift stores as a pass: the receipt carries the GATE's disposition, and the
// gate allows that shape in step 7-pre (mirrors produceCodexReviewReceipt).
// Every failure here is fail-open (stderr note, verdict unaffected).
func recordCodexReviewReceipt(scope reviewScope, state verify.ReceiptState, stateErr error, out ReviewOutput) {
	if stateErr != nil {
		return
	}
	verdict := codexReviewVerdictPass
	if isBlockVerdict(out.Verdict) {
		if _, ok := runtimeConfigOnlyFindings(out.Findings, scope.Dir); ok && scope.Class == reviewScopeTree {
			// The gate allows this shape in step 7-pre; the receipt carries
			// that disposition (mirrors produceCodexReviewReceipt).
			verdict = codexReviewVerdictPass
		} else {
			verdict = codexReviewVerdictFail
		}
	} else if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(out.Verdict)), codexReviewVerdictPass) {
		verdict = codexReviewVerdictInconclusive
	}
	r := verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		Verdict:      verdict,
		RecordedAt:   time.Now(),
	}
	if err := verify.RecordReceipt(scope.Dir, r); err != nil {
		_, _ = os.Stderr.WriteString("codex review gate: verdict receipt not recorded (" + err.Error() + ") — the next Stop re-reviews\n")
	}
}
