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
	"fmt"
	"os"
	"path/filepath"
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
// gate allows that shape in step 7-pre (mirrors produceCodexReviewReceipt,
// including the fail exit code 1 — a stored pass-shaped exit on a fail verdict
// would read as success evidence to the receipt's other consumers). On a fail
// the summary and findings are preserved to the local detail file, so a cached
// block can still say WHAT to fix (the receipt store carries no free text).
// Every failure here is fail-open (stderr note, verdict unaffected).
func recordCodexReviewReceipt(scope reviewScope, state verify.ReceiptState, stateErr error, out ReviewOutput) {
	if stateErr != nil {
		return
	}
	verdict, exit := codexReviewVerdictPass, 0
	if isBlockVerdict(out.Verdict) {
		if _, ok := runtimeConfigOnlyFindings(out.Findings, scope.Dir); ok && scope.Class == reviewScopeTree {
			// The gate allows this shape in step 7-pre; the receipt carries
			// that disposition (mirrors produceCodexReviewReceipt).
			verdict = codexReviewVerdictPass
		} else {
			verdict, exit = codexReviewVerdictFail, 1
		}
	} else if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(out.Verdict)), codexReviewVerdictPass) {
		// Inconclusive must not read as success evidence either: HasLocalPass
		// keys on ExitCode == 0, so an inconclusive receipt stored with the
		// zero value invents a local pass for the escalation detector's
		// contradiction limb. Exit 2 (the system-error class) keeps the
		// gate's fail-open allow while the RECORD reads "no verdict".
		verdict, exit = codexReviewVerdictInconclusive, 2
	}
	r := verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		ExitCode:     exit,
		Verdict:      verdict,
		RecordedAt:   time.Now(),
	}
	if err := verify.RecordReceipt(scope.Dir, r); err != nil {
		_, _ = os.Stderr.WriteString("codex review gate: verdict receipt not recorded (" + err.Error() + ") — the next Stop re-reviews\n")
		return
	}
	if verdict == codexReviewVerdictFail {
		codexReviewPreserveFindings(scope.Dir, state, out)
	}
}

// codexReviewFindingsPath is the deterministic local path a fail verdict's
// summary and findings are preserved at, keyed by the same tree key the
// receipt binds (HEAD + porcelain digest). Under .moai/state/ the file is
// runtime-managed state: gitignored, ignored by the gate's own change
// detector, and gone when the key moves — a stale detail never outlives its
// verdict. The receipt store carries no free text, so this file is the only
// carrier for the "what to fix" a cached block must still report.
func codexReviewFindingsPath(dir, head, digest string) string {
	return filepath.Join(dir, ".moai", "state", "verify", "codex-review",
		fmt.Sprintf("%s-%s.md", head, digest))
}

// codexReviewPreserveFindings writes the fail's summary and findings to the
// detail file. Fail-open: a write failure loses only the cached block's
// detail — the verdict and the receipt are already stored, and the next
// live review regenerates everything.
func codexReviewPreserveFindings(dir string, state verify.ReceiptState, out ReviewOutput) {
	var b strings.Builder
	fmt.Fprintf(&b, "# codex review fail — %s\n\n%s\n\n## Findings\n\n",
		strings.TrimSpace(out.Verdict), strings.TrimSpace(out.Summary))
	for _, f := range out.Findings {
		fmt.Fprintf(&b, "- [%s] %s — %s (%s:%d)\n",
			strings.TrimSpace(f.Severity), strings.TrimSpace(f.Title),
			strings.TrimSpace(f.Body), strings.TrimSpace(f.File), f.Line)
	}
	path := codexReviewFindingsPath(dir, state.Head, state.TreeDigest)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		_, _ = os.Stderr.WriteString("codex review gate: fail detail not preserved (" + err.Error() + ")\n")
	}
}

// codexReviewCachedDetail reads the preserved fail detail back for the cached
// block message. Returns "" when unreadable — the block then keeps the bare
// cached-verdict form rather than inventing detail (fail-open, REQ-GBN-004).
func codexReviewCachedDetail(dir string, r verify.Receipt) string {
	b, err := os.ReadFile(codexReviewFindingsPath(dir, r.Head, r.TreeDigest))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
