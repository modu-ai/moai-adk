package cli

// SPEC-CODEX-GATE-SCOPING-001 — the card-review repair round 2 (card t1404).
//
// The round-1 repair commits (2c25df957, d2f3b1d1e, 12152cd5b) were
// re-reviewed by codex and five NEW defects were returned, all in the
// path-resolution / reclassification-target family. Each test below pins one
// finding through an EXISTING entry point, authored BEFORE its repair and
// observed RED on the pre-repair tree:
//
//	N1  codexFindingsOf takes the FIRST path:line in a finding message as
//	    File, so a headline naming the config surface while the actual
//	    location is a source file reclassifies a real defect as drift and
//	    silently ALLOWs — the anchor must stay unset when the message carries
//	    several distinct path candidates (REQ-CGSC-008 reclassifies only
//	    findings targeting ONLY those surfaces)
//	N3  finding-path normalization anchors on the session tree, so a session
//	    sitting in a subdirectory relativizes an absolute repo-root config
//	    path to a "../" form that escapes the exclusion sets and a config-only
//	    FAIL stays BLOCK — the comparison anchors on the GIT REPOSITORY ROOT
//	N4  isPrimaryCheckoutGit compares spellings, and through a symlinked
//	    subdirectory git reports --git-dir as the REAL absolute path while
//	    --git-common-dir comes back relative — the same primary directory
//	    judges not-primary through the link and the default skip is missed
//	N5  porcelain collapses a fully-untracked .moai/ tree to "?? .moai/", so
//	    the config-only probe cannot see .moai/config/ inside the collapsed
//	    entry and a config-only untracked change counts as reviewable — the
//	    probe collects untracked files at file level (--untracked-files=all)
//	(N2 is the sync gate script's scan/key parity; its repro lives in
//	internal/template/hook_gate_reports_scan_parity_test.go beside the M4
//	tests it repairs.)

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

)

// --- N1: the parser claims no anchor it cannot defend ----------------------

// multiPathDriftReviewText is the N1 shape: the finding's HEADLINE names the
// runtime-config surface while its ACTUAL location is a source file — two
// distinct path:line candidates in one message. Taking the first candidate
// misreads the finding as config drift and silently ALLOWs a source defect.
const multiPathDriftReviewText = "- [P1] `.moai/config/sections/workflow.yaml:25` gate key drifted — fix the loader guard at `internal/config/loader.go:125`\n"

// TestCodexFindingsOfAmbiguousMultiPathMessageLeavesAnchorUnset pins N1 at the
// parser: a message carrying SEVERAL distinct path candidates has no
// defensible single location, so the anchor stays unset and consumers that
// require an unambiguous target (the REQ-CGSC-008 reclassification) keep the
// strict disposition. A reference inside a title is not the target; ambiguity
// is not resolved by position.
//
// Controls pin the two shapes that stay anchored: exactly one candidate (the
// message's location), and one DISTINCT candidate repeated (the same path at
// two lines is still one location).
func TestCodexFindingsOfAmbiguousMultiPathMessageLeavesAnchorUnset(t *testing.T) {
	out := synthesizeReviewOutput(multiPathDriftReviewText, codexMethodReviewStart)
	if out.Verdict != "fail" || len(out.Findings) != 1 {
		t.Fatalf("premise: the fixture must synthesize fail with one finding, got verdict %q findings %+v", out.Verdict, out.Findings)
	}
	f := out.Findings[0]
	if f.File != "" || f.Line != 0 {
		t.Errorf("a message with two distinct path candidates has no defensible anchor — File:Line = %q:%d, want unset so the reclassification never claims it", f.File, f.Line)
	}

	single := synthesizeReviewOutput("- [P1] `.claude/settings.json:13` personal PATH entry drifted\n", codexMethodReviewStart)
	if len(single.Findings) != 1 || single.Findings[0].File != ".claude/settings.json" || single.Findings[0].Line != 13 {
		t.Errorf("control: an exactly-one-candidate message keeps its anchor, got %+v", single.Findings)
	}
	repeat := synthesizeReviewOutput("- [P1] `main.go:10` duplicated key, also at `main.go:44`\n", codexMethodReviewStart)
	if len(repeat.Findings) != 1 || repeat.Findings[0].File != "main.go" {
		t.Errorf("control: one DISTINCT candidate repeated keeps its anchor (first occurrence's line), got %+v", repeat.Findings)
	}
}

// TestCodexReviewGateMultiPathFindingKeepsStrictDisposition pins N1 end to end
// at its M2 home, the receipt producer (the gate's live arm moved with the
// review — REQ-GBN-002): the ambiguous-anchor finding must keep the strict
// disposition in the recorded receipt, and no reclassification row may be
// recorded for it.
func TestCodexReviewGateMultiPathFindingKeepsStrictDisposition(t *testing.T) {
	prem := synthesizeReviewOutput(multiPathDriftReviewText, codexMethodReviewStart)
	if prem.Verdict != "fail" || len(prem.Findings) != 1 {
		t.Fatalf("premise: the fixture must synthesize fail with one finding, got verdict %q findings %+v", prem.Verdict, prem.Findings)
	}

	root := cacheTestRoot(t)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
	withCodexSession(t, codexSessionScript(multiPathDriftReviewText))

	read := captureGateDiagnostics(t)
	r, err := produceCodexReviewReceipt(context.Background(), root)
	if err != nil {
		t.Fatalf("receipt producer error: %v", err)
	}
	diagnostics := read()
	if r.Verdict != codexReviewVerdictFail {
		t.Errorf("a fail finding whose message cannot pin an unambiguous config target must keep the strict disposition, got %q", r.Verdict)
	}
	if strings.Contains(diagnostics, "runtime-managed") {
		t.Errorf("no reclassification may be recorded for an ambiguous-anchor finding; diagnostics: %q", diagnostics)
	}
}

// --- N3: the exclusion comparison anchors on the repo root -----------------

// TestCodexReviewGateRuntimeDriftFindingAnchoredToRepoRoot pins N3 at its M2
// home, the receipt producer (the gate's live arm moved with the review —
// REQ-GBN-002): an absolute config finding reported against the repo root must
// normalize into the reclassification — the comparison anchors on the GIT
// REPOSITORY ROOT of the reviewed tree. macOS temp dirs sit behind a symlink,
// so the fixture composes the anchor from the resolved root.
//
// Control: a finding anchored OUTSIDE the repository relativizes to a "../"
// form no exclusion set matches and keeps the strict disposition — anchoring
// on the root must not widen the exclusion to every absolute path.
func TestCodexReviewGateRuntimeDriftFindingAnchoredToRepoRoot(t *testing.T) {
	root := cacheTestRoot(t)
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("eval fixture root: %v", err)
	}

	anchoredText := "- [P1] `" + filepath.ToSlash(filepath.Join(resolvedRoot, ".claude", "settings.json")) +
		":13` personal PATH entry drifted\n"
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
	withCodexSession(t, codexSessionScript(anchoredText))
	read := captureGateDiagnostics(t)
	r, err := produceCodexReviewReceipt(context.Background(), root)
	if err != nil {
		t.Fatalf("receipt producer error: %v", err)
	}
	diagnostics := read()
	if r.Verdict != codexReviewVerdictPass {
		t.Errorf("an absolute config finding must anchor on the repo root and reclassify, got verdict %q", r.Verdict)
	}
	if !strings.Contains(diagnostics, "runtime-managed") || !strings.Contains(diagnostics, ".claude/settings.json") {
		t.Errorf("the reclassification row must name the config target; diagnostics: %q", diagnostics)
	}

	// Control: outside-the-repo anchors keep the strict disposition.
	outside := t.TempDir()
	outsideText := "- [P1] `" + filepath.ToSlash(filepath.Join(outside, ".claude", "settings.json")) +
		":1` some other tree's settings\n"
	withCodexSession(t, codexSessionScript(outsideText))
	readOut := captureGateDiagnostics(t)
	rOut, err := produceCodexReviewReceipt(context.Background(), root)
	if err != nil {
		t.Fatalf("receipt producer error (control): %v", err)
	}
	diagnosticsOut := readOut()
	if rOut.Verdict != codexReviewVerdictFail {
		t.Errorf("control: a finding anchored outside the repository must keep the strict disposition, got %q", rOut.Verdict)
	}
	if strings.Contains(diagnosticsOut, "runtime-managed") {
		t.Errorf("control: no reclassification may be recorded for an outside-the-repo anchor; diagnostics: %q", diagnosticsOut)
	}
}

// --- N4: the primary judgment compares locations, not spellings ------------

// TestIsPrimaryCheckoutGitSymlinkedSubdirectory pins N4: through a symlinked
// subdirectory of the primary checkout, git reports --git-dir as the REAL
// absolute path while --git-common-dir comes back relative to the link, so a
// Clean-only comparison spells two different locations for the same directory
// and the default primary skip is missed. Both sides resolve symlinks before
// the comparison.
func TestIsPrimaryCheckoutGitSymlinkedSubdirectory(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	cardScopeGit(t, root, "init", "-q", "-b", "main")
	sub := filepath.Join(root, "internal")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "internal-link")
	if err := os.Symlink(sub, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	// Premise: the spelling mismatch the repair resolves — if a future git
	// reports both answers in the same form, this premise fails loudly rather
	// than the assertion passing vacuously.
	gitDir := cardScopeGit(t, link, "rev-parse", "--git-dir")
	commonDir := cardScopeGit(t, link, "rev-parse", "--git-common-dir")
	if !filepath.IsAbs(gitDir) || filepath.IsAbs(commonDir) {
		t.Fatalf("premise: through a symlink git must report an absolute --git-dir %q and a relative --git-common-dir %q", gitDir, commonDir)
	}

	if !isPrimaryCheckoutGit(root) {
		t.Errorf("control: the primary checkout itself must judge primary")
	}
	if !isPrimaryCheckoutGit(link) {
		t.Errorf("the same primary directory through a symlinked subdirectory must judge primary too — the comparison must resolve symlinks before comparing (N4)")
	}

	// Control: a LINKED worktree is a different tree and stays non-primary —
	// symlink resolution must not flatten the worktree/common-dir distinction.
	wt := t.TempDir()
	cardScopeGit(t, root, "worktree", "add", "-q", "-b", "WT-n4-probe", wt)
	if isPrimaryCheckoutGit(wt) {
		t.Errorf("control: a linked worktree must stay non-primary")
	}
}

// --- N5: the config-only probe reads untracked files at file level ---------

// newUntrackedSeedRepo builds a seeded repository (one committed source file)
// so callers can layer untracked additions over a clean HEAD.
func newUntrackedSeedRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	cardScopeGit(t, root, "init", "-q", "-b", "main")
	writeCardFile(t, root, "main.go", "package main\n")
	cardScopeGit(t, root, "add", "-A")
	cardScopeGit(t, root, "commit", "-q", "-m", "seed")
	return root
}

// TestTreeConfigOnlyChangesSeesUntrackedConfigThroughCollapsedPorcelain pins
// N5: porcelain collapses a fully-untracked .moai/ tree to `?? .moai/`, an
// entry the exclusion sets cannot see .moai/config/ inside, so the config-only
// probe must collect untracked files at FILE level (--untracked-files=all) —
// the cardChangedPaths precedent.
func TestTreeConfigOnlyChangesSeesUntrackedConfigThroughCollapsedPorcelain(t *testing.T) {
	// Case 1: untracked managed-config additions only — config-only.
	root := newUntrackedSeedRepo(t)
	writeCardFile(t, root, filepath.Join(".moai", "config", "sections", "workflow.yaml"), "workflow:\n")
	if !treeConfigOnlyChanges(root) {
		t.Errorf("an untracked managed-config addition must read config-only even though porcelain collapses the .moai/ tree to one entry (N5)")
	}
	// The composition the defect reaches: the scoped self-gate must not call
	// this tree reviewable (real detector inside — no seam swap).
	if reviewGateScopedChangeDetector(reviewScope{Class: reviewScopeTree, Dir: root}) {
		t.Errorf("a config-only untracked tree must not be reviewable through the scoped self-gate (N5)")
	}

	// Case 2: control — an untracked ordinary source file keeps the tree
	// reviewable.
	rootSrc := newUntrackedSeedRepo(t)
	writeCardFile(t, rootSrc, "extra.go", "package main\n")
	if treeConfigOnlyChanges(rootSrc) {
		t.Errorf("control: an untracked ordinary source file must keep the tree reviewable")
	}

	// Case 3: mixed config + source — the source leg keeps it reviewable.
	rootMix := newUntrackedSeedRepo(t)
	writeCardFile(t, rootMix, filepath.Join(".moai", "config", "sections", "workflow.yaml"), "workflow:\n")
	writeCardFile(t, rootMix, "extra.go", "package main\n")
	if treeConfigOnlyChanges(rootMix) {
		t.Errorf("control: mixed config + source untracked changes must keep the tree reviewable")
	}
	if !reviewGateScopedChangeDetector(reviewScope{Class: reviewScopeTree, Dir: rootMix}) {
		t.Errorf("control: the mixed tree must stay reviewable through the scoped self-gate")
	}

	// Case 4: regression pin — a TRACKED-modified config file still reads
	// config-only (the round-1 shape, unchanged by the file-level listing).
	rootTracked := newUntrackedSeedRepo(t)
	writeCardFile(t, rootTracked, filepath.Join(".moai", "config", "sections", "workflow.yaml"), "workflow:\n")
	cardScopeGit(t, rootTracked, "add", "-A")
	cardScopeGit(t, rootTracked, "commit", "-q", "-m", "config")
	writeCardFile(t, rootTracked, filepath.Join(".moai", "config", "sections", "workflow.yaml"), "workflow:\n  codex:\n")
	if !treeConfigOnlyChanges(rootTracked) {
		t.Errorf("regression: a tracked-modified config file must still read config-only")
	}
}
