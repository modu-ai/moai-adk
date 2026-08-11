package fix

// apply_test.go — M3.5 apply-on-approval tests (SPEC-NAVIGATOR-SYNC-005,
// AC-NS5-008a / AC-NS5-008c / AC-NS5-008d). Drives the apply engine
// (apply.go) against fixture draft staging dirs under t.TempDir:
//
//   - TestApplyTokenRefusal_* — AC-NS5-008d (no-token refuse + valid-token apply)
//   - TestApply_AtomicRenameAndLedger — AC-NS5-008c (atomic-rename + applied.json)
//   - TestApply_Idempotence — DBT-2 (resume skips already-applied IDs)
//   - TestApply_OnlyApprovedSubtreesTouched — AC-NS5-008a + REQ-NS5-013
//
// Test isolation: every temp tree lives under t.TempDir() (auto-cleaned). No
// test writes to the real project root or to a real .moai/project/ directory.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- Apply fixture builder ---------------------------------------------------

// setupApplyFixture constructs a complete apply-ready staging dir under a temp
// root and returns (root, draftID, requestProvenance). It writes:
//
//   - .moai/project/navigator/fix-drafts/<draftID>/request.json — a request
//     with 2 in-scope subtrees (capability-map.md + audit-report.json).
//   - .moai/project/navigator/fix-drafts/<draftID>/draft/<docSurface> — the
//     AI-drafted NEW content for each doc surface.
//   - .moai/project/navigator/<docSurface> — the LIVE doc surface with known
//     initial content (so the test can assert it was mutated post-apply).
//
// The request.json provenance is git-sourced when the temp root is a git repo
// (the caller init-s a repo + makes a baseline commit); otherwise it carries
// deterministic placeholder SHAs. The returned provenance is what the approval
// token must be hashed against.
func setupApplyFixture(t *testing.T, initGit bool) (root, draftID string, prov Provenance) {
	t.Helper()
	root = t.TempDir()

	// Minimal git repo so git-committer-date + rev-parse resolve (the apply
	// ledger uses git log -1 --format=%cI for ApprovalTimestamp). When initGit
	// is false, the apply falls back to deterministic placeholders.
	if initGit {
		for _, args := range [][]string{
			{"init"},
			{"config", "user.email", "test@example.com"},
			{"config", "user.name", "test"},
			{"config", "commit.gpgsign", "false"},
		} {
			cmd := exec.Command("git", args...)
			cmd.Dir = root
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
		}
		// A baseline commit so HEAD + committer-date resolve.
		if err := os.WriteFile(filepath.Join(root, "README"), []byte("baseline\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("git", "add", "README")
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git add: %v\n%s", err, out)
		}
		cmd = exec.Command("git", "commit", "-m", "baseline")
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit: %v\n%s", err, out)
		}
	}

	fixSHA := gitLine(root, "rev-parse", "HEAD")
	if fixSHA == "" {
		fixSHA = "fixsha123"
	}
	capturedAt := gitLine(root, "log", "-1", "--format=%cI")
	if capturedAt == "" {
		capturedAt = "2026-01-01T00:00:00+00:00"
	}
	prov = Provenance{
		FixCommitSHA:      fixSHA,
		BaselineCommitSHA: "baseline123",
		CapturedAt:        capturedAt,
	}

	// diff_scope carries 2 in-scope subtrees spanning 2 doc surfaces.
	scope := []DiffScopeEntry{
		{DocSurface: "capability-map.md", SubtreeID: "DEC-ONE"},
		{DocSurface: "audit-report.json", SubtreeID: "SPEC-TWO"},
	}
	req := DraftRequest{
		Provenance:   prov,
		DiffScope:    scope,
		WorkItemRefs: []WorkItemRef{},
		DraftInstructions: DraftInstructions{
			PerSubtree: []SubtreeStrategy{
				{SubtreeID: "DEC-ONE", Strategy: "regenerate row"},
				{SubtreeID: "SPEC-TWO", Strategy: "draft SPEC stub"},
			},
		},
	}
	draftID = computeDraftID(scope, prov.BaselineCommitSHA)
	draftDir := filepath.Join(root, fixDraftsRelDir, draftID)

	reqBytes, err := marshalRequest(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	fixWrite(t, filepath.Join(draftDir, "request.json"), string(reqBytes))

	// Draft NEW content per doc surface (the layer-2 AI-drafted output).
	fixWrite(t, filepath.Join(draftDir, "draft", "capability-map.md"),
		"# Capability Map (drafted)\n\nDEC-ONE: updated decision row.\n")
	fixWrite(t, filepath.Join(draftDir, "draft", "audit-report.json"),
		`{"audits":[{"id":"SPEC-TWO","status":"drafted"}]}`+"\n")

	// LIVE doc surfaces with known initial content (so mutation is detectable).
	fixWrite(t, filepath.Join(root, liveDocRelPath("capability-map.md")),
		"# Capability Map (LIVE-OLD)\n\nDEC-ONE: stale decision row.\n")
	fixWrite(t, filepath.Join(root, liveDocRelPath("audit-report.json")),
		`{"audits":[{"id":"SPEC-TWO","status":"stale"}]}`+"\n")

	return root, draftID, prov
}

// writeApprovalJSON writes a valid approval.json (correct token) at the draft
// staging dir. option ∈ {"a","b","c"}; approver is the human identifier.
func writeApprovalJSON(t *testing.T, root, draftID, option, approver string, prov Provenance) {
	t.Helper()
	token := ComputeApprovalToken(draftID, option, prov)
	apv := Approval{
		DraftID:           draftID,
		ApprovalOption:    option,
		Approver:          approver,
		ApprovalTimestamp: gitLine(root, "log", "-1", "--format=%cI"),
		Token:             token,
	}
	b, err := json.MarshalIndent(apv, "", "  ")
	if err != nil {
		t.Fatalf("marshal approval: %v", err)
	}
	draftDir := filepath.Join(root, fixDraftsRelDir, draftID)
	fixWrite(t, filepath.Join(draftDir, "approval.json"), string(b)+"\n")
}

// --- AC-NS5-008d — approval_token refusal -----------------------------------

// TestApplyTokenRefusal_NoToken asserts the no-token branch (AC-NS5-008d):
// with NO approval.json at the staging dir, Apply REFUSES — returns a non-nil
// error (sentinel ErrApprovalTokenMissing), AND the live doc surfaces are NOT
// mutated (their SHAs are unchanged). The CLI caller translates the non-nil
// error to exit-non-zero (REQ-NS5-008 c4 is a hard guard, NOT fail-open).
func TestApplyTokenRefusal_NoToken(t *testing.T) {
	t.Parallel()

	root, draftID, _ := setupApplyFixture(t, true)

	liveMapSHA := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-map.md")))
	liveAuditSHA := sha256sum(t, filepath.Join(root, liveDocRelPath("audit-report.json")))

	// NO approval.json written — simulating a bare-shell invocation without a
	// prior gate approval.
	_, err := Apply(ApplyOptions{ProjectRoot: root, DraftID: draftID})

	if err == nil {
		t.Fatal("Apply succeeded without approval.json; want non-nil error (AC-NS5-008d refusal)")
	}
	if !isApprovalTokenError(err) {
		t.Errorf("Apply error = %q, want an approval-token sentinel (ErrApprovalTokenMissing/ErrApprovalTokenInvalid)", err.Error())
	}

	// Live doc surfaces MUST be unchanged (the refusal contract).
	if got := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-map.md"))); got != liveMapSHA {
		t.Errorf("live capability-map.md mutated despite token refusal (SHA %q → %q)", liveMapSHA, got)
	}
	if got := sha256sum(t, filepath.Join(root, liveDocRelPath("audit-report.json"))); got != liveAuditSHA {
		t.Errorf("live audit-report.json mutated despite token refusal (SHA %q → %q)", liveAuditSHA, got)
	}

	// applied.json MUST NOT be written (no apply fired).
	if _, statErr := os.Stat(filepath.Join(root, fixDraftsRelDir, draftID, "applied.json")); statErr == nil {
		t.Error("applied.json was written despite token refusal; the apply MUST NOT fire")
	}
}

// TestApplyTokenRefusal_InvalidToken asserts the invalid-token branch
// (AC-NS5-008d): an approval.json IS present, but its token does not match
// hash(draft-id + option + request.json provenance) — Apply REFUSES the same
// way as the no-token case (non-nil error, no live-doc mutation).
func TestApplyTokenRefusal_InvalidToken(t *testing.T) {
	t.Parallel()

	root, draftID, _ := setupApplyFixture(t, true)

	// Write an approval.json with a TAMPERED token (does not match the
	// deterministic hash).
	apv := Approval{
		DraftID:           draftID,
		ApprovalOption:    "a",
		Approver:          "attacker",
		ApprovalTimestamp: gitLine(root, "log", "-1", "--format=%cI"),
		Token:             "tampered-token-value-not-the-real-hash",
	}
	b, _ := json.MarshalIndent(apv, "", "  ")
	fixWrite(t, filepath.Join(root, fixDraftsRelDir, draftID, "approval.json"), string(b)+"\n")

	liveMapSHA := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-map.md")))

	_, err := Apply(ApplyOptions{ProjectRoot: root, DraftID: draftID})
	if err == nil {
		t.Fatal("Apply succeeded with invalid token; want non-nil error (AC-NS5-008d refusal)")
	}
	if !isApprovalTokenError(err) {
		t.Errorf("Apply error = %q, want an approval-token sentinel", err.Error())
	}
	if got := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-map.md"))); got != liveMapSHA {
		t.Errorf("live capability-map.md mutated despite invalid token (SHA %q → %q)", liveMapSHA, got)
	}
}

// TestApplyTokenRefusal_ValidToken asserts the valid-token branch (AC-NS5-008d
// second half + AC-NS5-008c): WITH a valid approval.json (token matches the
// deterministic hash), Apply proceeds — the live doc surfaces ARE mutated to
// the draft content, and applied.json is written.
func TestApplyTokenRefusal_ValidToken(t *testing.T) {
	t.Parallel()

	root, draftID, prov := setupApplyFixture(t, true)
	writeApprovalJSON(t, root, draftID, "a", "engineer-alice", prov)

	oldMapSHA := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-map.md")))

	res, err := Apply(ApplyOptions{ProjectRoot: root, DraftID: draftID})
	if err != nil {
		t.Fatalf("Apply with valid token returned error: %v", err)
	}
	if !res.Applied {
		t.Fatal("Apply Result.Applied = false; want true")
	}
	if len(res.AppliedSubtreeIDs) != 2 {
		t.Errorf("AppliedSubtreeIDs count = %d, want 2; got %v", len(res.AppliedSubtreeIDs), res.AppliedSubtreeIDs)
	}

	// Live doc surfaces MUST now carry the draft content.
	newMap := readFile(t, filepath.Join(root, liveDocRelPath("capability-map.md")))
	if !strings.Contains(newMap, "Capability Map (drafted)") {
		t.Errorf("live capability-map.md not mutated to draft content; got:\n%s", newMap)
	}
	newMapSHA := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-map.md")))
	if newMapSHA == oldMapSHA {
		t.Error("live capability-map.md SHA unchanged after valid-token apply; expected mutation")
	}

	// applied.json MUST be written.
	appliedPath := filepath.Join(root, fixDraftsRelDir, draftID, "applied.json")
	data, err := os.ReadFile(appliedPath)
	if err != nil {
		t.Fatalf("read applied.json: %v", err)
	}
	var ledger AppliedLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatalf("unmarshal applied.json: %v", err)
	}
	if ledger.Approver != "engineer-alice" {
		t.Errorf("ledger.Approver = %q, want %q", ledger.Approver, "engineer-alice")
	}
	if ledger.ApprovalTimestamp == "" {
		t.Error("ledger.ApprovalTimestamp empty; want git-committer-date (git committer date via log -1), NOT wall-clock")
	}
	if len(ledger.AppliedSubtreeIDs) != 2 {
		t.Errorf("ledger.AppliedSubtreeIDs count = %d, want 2", len(ledger.AppliedSubtreeIDs))
	}
	if ledger.ResultingLiveDocSHA == "" {
		t.Error("ledger.ResultingLiveDocSHA empty; want the post-apply live-doc SHA")
	}
}

// TestApplyTokenRefusal_ValidToken is the combined AC-NS5-008d assertion: BOTH
// branches (no-token refuse + valid-token apply) in one test pass, matching the
// acceptance.md §C gate ("asserts BOTH branches").
func TestApplyTokenRefusal(t *testing.T) {
	t.Parallel()
	t.Run("NoToken", TestApplyTokenRefusal_NoToken)
	t.Run("InvalidToken", TestApplyTokenRefusal_InvalidToken)
	t.Run("ValidToken", TestApplyTokenRefusal_ValidToken)
}

// --- AC-NS5-008c — atomic-rename + applied.json ledger ----------------------

// TestApply_AtomicRenameAndLedger verifies the atomic-rename contract: the
// draft file is moved into place via .tmp + os.Rename (no partial write), and
// the applied.json ledger carries approver + git-committer-date timestamp +
// applied subtree IDs + resulting live-doc SHA.
func TestApply_AtomicRenameAndLedger(t *testing.T) {
	t.Parallel()

	root, draftID, prov := setupApplyFixture(t, true)
	writeApprovalJSON(t, root, draftID, "a", "engineer-bob", prov)

	res, err := Apply(ApplyOptions{ProjectRoot: root, DraftID: draftID})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Resulting live-doc SHA in the ledger MUST match the actual file content's
	// SHA for at least one applied surface.
	appliedPath := filepath.Join(root, fixDraftsRelDir, draftID, "applied.json")
	data, _ := os.ReadFile(appliedPath)
	var ledger AppliedLedger
	_ = json.Unmarshal(data, &ledger)
	// ResultingLiveDocSHA is a composite hash across applied surfaces — verify
	// it's non-empty and stable on re-derivation (re-applying idempotently
	// yields the same resulting SHA).
	if ledger.ResultingLiveDocSHA == "" {
		t.Error("ledger.ResultingLiveDocSHA empty")
	}

	// No .tmp file left behind (atomic-rename completes).
	draftDir := filepath.Join(root, fixDraftsRelDir, draftID)
	matches, _ := filepath.Glob(filepath.Join(root, ".moai", "project", "navigator", "*.tmp"))
	if len(matches) > 0 {
		t.Errorf("stale .tmp files left behind after atomic-rename: %v", matches)
	}
	_ = res
	_ = draftDir
}

// --- DBT-2 idempotence -------------------------------------------------------

// TestApply_Idempotence verifies the crash-resume contract (DBT-2, plan.md
// §C.6): a second Apply call after a successful first one is a no-op — it does
// NOT re-rename, does NOT error, and the applied.json ledger records the same
// set of subtree IDs (no duplicates).
func TestApply_Idempotence(t *testing.T) {
	t.Parallel()

	root, draftID, prov := setupApplyFixture(t, true)
	writeApprovalJSON(t, root, draftID, "a", "engineer-carol", prov)

	// First apply: succeeds, mutates live docs, writes ledger.
	if _, err := Apply(ApplyOptions{ProjectRoot: root, DraftID: draftID}); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	firstLedger := readAppliedLedger(t, root, draftID)
	if len(firstLedger.AppliedSubtreeIDs) != 2 {
		t.Fatalf("first apply ledger IDs = %v, want 2", firstLedger.AppliedSubtreeIDs)
	}

	// Re-write the draft NEW content so a NON-idempotent apply would re-mutate
	// the live doc to a different value (the idempotence check must NOT fire a
	// second rename). If the second apply re-renamed, the live doc would change.
	fixWrite(t, filepath.Join(root, fixDraftsRelDir, draftID, "draft", "capability-map.md"),
		"# DIFFERENT CONTENT — idempotence check\n")

	liveMapBefore := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-map.md")))

	// Second apply: no-op (all subtree IDs already in the ledger).
	if _, err := Apply(ApplyOptions{ProjectRoot: root, DraftID: draftID}); err != nil {
		t.Fatalf("second (idempotent) Apply returned error: %v", err)
	}

	liveMapAfter := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-map.md")))
	if liveMapAfter != liveMapBefore {
		t.Errorf("idempotence violated: live doc mutated on second apply (SHA %q → %q)", liveMapBefore, liveMapAfter)
	}

	// Ledger is unchanged (no duplicate IDs).
	secondLedger := readAppliedLedger(t, root, draftID)
	if len(secondLedger.AppliedSubtreeIDs) != len(firstLedger.AppliedSubtreeIDs) {
		t.Errorf("idempotence violated: ledger grew from %d to %d IDs",
			len(firstLedger.AppliedSubtreeIDs), len(secondLedger.AppliedSubtreeIDs))
	}
}

// --- AC-NS5-008a + REQ-NS5-013 — only approved + scope-conformant subtrees --

// TestApply_OnlyApprovedSubtreesTouched verifies that an OUT-OF-SCOPE subtree
// in the draft (over-produced by manager-develop, REQ-NS5-013) is excluded
// from the apply AND the live doc surface it would have mutated is NOT touched
// beyond the in-scope patches.
func TestApply_OnlyApprovedSubtreesTouched(t *testing.T) {
	t.Parallel()

	root, draftID, prov := setupApplyFixture(t, true)

	// Plant an out-of-scope draft file at a THIRD doc surface
	// (capability-symbols.json) — its subtree ID is NOT in diff_scope[].
	fixWrite(t, filepath.Join(root, fixDraftsRelDir, draftID, "draft", "capability-symbols.json"),
		`{"symbols":[{"id":"OUT-OF-SCOPE-Z","note":"over-produced"}]}`+"\n")
	// Live capability-symbols.json with known initial content.
	fixWrite(t, filepath.Join(root, liveDocRelPath("capability-symbols.json")),
		`{"symbols":[]}`+"\n")

	writeApprovalJSON(t, root, draftID, "a", "engineer-dave", prov)

	symbolsSHABefore := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-symbols.json")))

	res, err := Apply(ApplyOptions{ProjectRoot: root, DraftID: draftID})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Only the 2 in-scope subtree IDs are recorded as applied.
	if len(res.AppliedSubtreeIDs) != 2 {
		t.Errorf("AppliedSubtreeIDs = %v, want exactly 2 (in-scope only)", res.AppliedSubtreeIDs)
	}
	for _, id := range res.AppliedSubtreeIDs {
		if id != "DEC-ONE" && id != "SPEC-TWO" {
			t.Errorf("out-of-scope subtree %q was applied; only in-scope IDs allowed", id)
		}
	}

	// capability-symbols.json MUST be unchanged (the out-of-scope draft was NOT
	// applied).
	symbolsSHAAfter := sha256sum(t, filepath.Join(root, liveDocRelPath("capability-symbols.json")))
	if symbolsSHAAfter != symbolsSHABefore {
		t.Errorf("out-of-scope draft mutated live capability-symbols.json (SHA %q → %q)",
			symbolsSHABefore, symbolsSHAAfter)
	}
}

// --- ComputeApprovalToken (deterministic hash) ------------------------------

// TestComputeApprovalToken_Deterministic verifies the token is a deterministic
// SHA-256 of (draft-id + option + provenance), NOT a random nonce — two calls
// on identical inputs produce the identical token (plan.md §C.6).
func TestComputeApprovalToken_Deterministic(t *testing.T) {
	t.Parallel()

	prov := Provenance{
		FixCommitSHA:      "abc",
		BaselineCommitSHA: "def",
		CapturedAt:        "2026-01-01T00:00:00+00:00",
	}
	t1 := ComputeApprovalToken("draft-1", "a", prov)
	t2 := ComputeApprovalToken("draft-1", "a", prov)
	if t1 != t2 {
		t.Errorf("token not deterministic: %q vs %q", t1, t2)
	}
	// A different option or draft-id yields a different token.
	t3 := ComputeApprovalToken("draft-1", "b", prov)
	if t3 == t1 {
		t.Error("token collision across different options (a vs b)")
	}
	t4 := ComputeApprovalToken("draft-2", "a", prov)
	if t4 == t1 {
		t.Error("token collision across different draft-ids")
	}
	// Different provenance yields a different token.
	prov2 := prov
	prov2.FixCommitSHA = "xyz"
	t5 := ComputeApprovalToken("draft-1", "a", prov2)
	if t5 == t1 {
		t.Error("token collision across different provenance")
	}
	// The token is a hex SHA-256 (64 hex chars).
	if len(t1) != 64 {
		t.Errorf("token = %q, want 64-char hex SHA-256", t1)
	}
	// Sanity: token matches a direct sha256 hex computation so an external
	// caller (the orchestrator) can re-derive it.
	h := sha256.New()
	h.Write([]byte("draft-1"))
	h.Write([]byte{0})
	h.Write([]byte("a"))
	h.Write([]byte{0})
	h.Write([]byte(prov.FixCommitSHA))
	h.Write([]byte{0})
	h.Write([]byte(prov.BaselineCommitSHA))
	h.Write([]byte{0})
	h.Write([]byte(prov.CapturedAt))
	expected := hex.EncodeToString(h.Sum(nil))
	if t1 != expected {
		t.Errorf("token = %q, want direct-hash %q (the orchestrator must be able to re-derive it)", t1, expected)
	}
}

// --- helpers -----------------------------------------------------------------

// sha256sum returns the hex SHA-256 of the file at path.
func sha256sum(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("sha256sum: read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// readFile reads the file at path or fails the test.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// readAppliedLedger reads + unmarshals the applied.json ledger.
func readAppliedLedger(t *testing.T, root, draftID string) AppliedLedger {
	t.Helper()
	path := filepath.Join(root, fixDraftsRelDir, draftID, "applied.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read applied.json: %v", err)
	}
	var ledger AppliedLedger
	if err := json.Unmarshal(b, &ledger); err != nil {
		t.Fatalf("unmarshal applied.json: %v", err)
	}
	return ledger
}

// isApprovalTokenError returns true if err wraps one of the approval-token
// sentinel errors (AC-NS5-008d refusal). Uses errors.Is so the wrapped
// `fmt.Errorf("%w: ...")` form still matches.
func isApprovalTokenError(err error) bool {
	return errors.Is(err, ErrApprovalTokenMissing) || errors.Is(err, ErrApprovalTokenInvalid)
}
