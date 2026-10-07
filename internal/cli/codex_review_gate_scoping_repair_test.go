package cli

// SPEC-CODEX-GATE-SCOPING-001 — the card-review repair round (card t1404).
//
// The advisory codex card review (base 2de0a2cb6, verdict recorded at
// .moai/reports/t1404/card-review.md) returned five findings against the M3
// Facet-2 implementation. Each test below pins one finding's behavior through
// an EXISTING entry point, so the file compiles and runs RED on the
// pre-repair tree and GREEN after it:
//
//	R1  the tree-only config exclusion leaked into the SHARED change detector,
//	    so the multi-review gate ALLOWs over a stored required FAIL when the
//	    only changes are runtime-config surfaces (the multi-review handler and
//	    Codex Stop-chain member 7 both consult reviewGateChangeDetector)
//	R2  the runtime-drift reclassification ran on the Claude Stop path only —
//	    the receipt producer still recorded a fail the Codex Stop chain DENYs on
//	R3  finding paths were matched by relative prefix without normalization, so
//	    an absolute-path config finding kept the block
//	R4  the .claude/settings.json prefix also matched sibling names that merely
//	    extend the file name (.claude/settings.json.template)
//	R5  a rename porcelain record kept only the source path, so the destination
//	    inherited the source's exclusion
//
// REQ-CGSC-008's two-arm intent (decision-index Q3: both arms kept) and
// REQ-CGSC-007's tree-only scope decide the placement: the exclusions and the
// reclassification live on the tree-scope-only surfaces; the shared baseline
// detector, the card path and the multi-review paths keep full reviewability
// (REQ-CGSC-005 / AC-CGSC-009).

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// newConfigOnlyRepo builds a real git repository whose only UNCOMMITTED change
// is a tracked file under the MoAI managed config tree — the working-tree
// shape R1 claims slips past the shared detector. The config file is COMMITTED
// in the seed and then modified, so porcelain names it verbatim (an untracked
// .moai/ tree would collapse to "?? .moai/" and defeat any prefix filter, the
// cardChangedPaths precedent).
func newConfigOnlyRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	cardScopeGit(t, dir, "init", "-q", "-b", "main")
	writeCardFile(t, dir, "main.go", "package main\n")
	writeCardFile(t, dir, filepath.Join(".moai", "config", "sections", "workflow.yaml"), "workflow:\n")
	cardScopeGit(t, dir, "add", "-A")
	cardScopeGit(t, dir, "commit", "-q", "-m", "seed")
	writeCardFile(t, dir, filepath.Join(".moai", "config", "sections", "workflow.yaml"), "workflow:\n  codex:\n")
	return dir
}

// --- R1: the shared detector carries no tree-only exclusion ----------------

// TestMultiReviewGateConfigOnlyChangesStillReadStoredFail is the R1 repro at
// the leaked consumer: the multi-review gate consults the SHARED detector, so
// a tree whose only changes are runtime-config surfaces must still reach the
// stored convergence result and BLOCK on a required FAIL — the baseline gate
// source blocked this exact fixture before the tree-only exclusion landed
// inside reviewableFromPorcelain.
func TestMultiReviewGateConfigOnlyChangesStillReadStoredFail(t *testing.T) {
	root := newConfigOnlyRepo(t)
	stateDir := filepath.Join(root, ".moai", "state", "audit-multi")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"overall_verdict":"fail","residual_risk_note":"required FAIL (R1 fixture)"}`
	if err := os.WriteFile(filepath.Join(stateDir, "s-r1.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := HandleMultiReviewGate(&hook.HookInput{SessionID: "s-r1"}, true, root, "s-r1")
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("config-only changes must not silence a stored required FAIL; got %+v", out)
	}
	if !strings.Contains(out.Reason, "required FAIL") {
		t.Errorf("the block must carry the stored result's note, got %q", out.Reason)
	}
}

// TestTreeScopedSelfGateSkipsConfigOnlyTree pins REQ-CGSC-007 end to end after
// the R1 relocation: a real git tree whose only change is the managed config
// file reads NOT reviewable through the scoped self-gate, and one source
// change makes it reviewable again. The exclusion lives in the tree-scope-only
// caller, never in the shared parser.
func TestTreeScopedSelfGateSkipsConfigOnlyTree(t *testing.T) {
	root := newConfigOnlyRepo(t)
	scope := reviewScope{Class: reviewScopeTree, Dir: root}
	if reviewGateScopedChangeDetector(scope) {
		t.Errorf("a config-only tree must not be reviewable through the scoped self-gate (REQ-CGSC-007)")
	}
	writeCardFile(t, root, "extra.go", "package main\n")
	if !reviewGateScopedChangeDetector(scope) {
		t.Errorf("a tree with a source change must stay reviewable through the scoped self-gate")
	}
}

// --- R2: the receipt producer reclassifies too -----------------------------
// (its repro lives in codex_review_receipt_scoping_repair_test.go, beside the
// producer it repairs — the two commits stay independently compiling.)

// --- R3: finding paths are normalized before the exclusion check -----------

// absoluteDriftReviewText is the R3 shape: the same config-only finding the
// reclassification admits, carrying the config file as an ABSOLUTE path — the
// un-normalized prefix comparison let it keep the block.
const absoluteDriftReviewText = "- [P1] `/proj/.claude/settings.json:13` personal PATH entry drifted\n"

// TestCodexReviewGateRuntimeDriftFindingsAbsolutePathsReclassified pins R3 at
// its M2 home, the receipt producer (the gate's live arm moved with the review
// — REQ-GBN-002): the finding anchor is normalized against the reviewed
// scope's tree before the exclusion comparison, so an absolute-path config
// finding reclassifies exactly as its relative twin does. The fixture composes
// the anchor from its resolved root (macOS temp dirs sit behind a symlink).
func TestCodexReviewGateRuntimeDriftFindingsAbsolutePathsReclassified(t *testing.T) {
	root := cacheTestRoot(t)
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("eval fixture root: %v", err)
	}
	absoluteText := "- [P1] `" + filepath.ToSlash(filepath.Join(resolvedRoot, ".claude", "settings.json")) +
		":13` personal PATH entry drifted\n"

	prem := synthesizeReviewOutput(absoluteText, codexMethodReviewStart)
	if prem.Verdict != "fail" || len(prem.Findings) != 1 ||
		prem.Findings[0].File != filepath.ToSlash(filepath.Join(resolvedRoot, ".claude", "settings.json")) {
		t.Fatalf("premise: the fixture must synthesize fail with the absolute config path, got verdict %q findings %+v", prem.Verdict, prem.Findings)
	}

	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
	withCodexSession(t, codexSessionScript(absoluteText))
	read := captureGateDiagnostics(t)
	r, err := produceCodexReviewReceipt(context.Background(), root)
	if err != nil {
		t.Fatalf("receipt producer error: %v", err)
	}
	diagnostics := read()
	if r.Verdict != codexReviewVerdictPass {
		t.Errorf("an absolute-path config finding must normalize into the reclassification, got verdict %q", r.Verdict)
	}
	if !strings.Contains(diagnostics, "runtime-managed") || !strings.Contains(diagnostics, ".claude/settings.json") {
		t.Errorf("the reclassification row must name the normalized target; diagnostics: %q", diagnostics)
	}
}

// --- R4: the settings file matches exactly, the directory by prefix --------

// TestTreeRuntimeConfigPathExactSettingsMatch pins R4: the local settings FILE
// is matched exactly, so a sibling name that merely extends it
// (.claude/settings.json.template — a template-source shape, or any other
// suffix a user or tool appends) stays reviewable; only the managed config
// DIRECTORY matches by prefix.
func TestTreeRuntimeConfigPathExactSettingsMatch(t *testing.T) {
	if !isTreeRuntimeConfigPath(".claude/settings.json") {
		t.Errorf("the exact settings surface must still be excluded (REQ-CGSC-007)")
	}
	if !isTreeRuntimeConfigPath(".moai/config/sections/workflow.yaml") {
		t.Errorf("the managed config directory must still be excluded by prefix (REQ-CGSC-007)")
	}
	for _, keep := range []string{
		".claude/settings.json.template",
		".claude/settings.json.bak",
		"cmd/moai/main.go",
	} {
		if isTreeRuntimeConfigPath(keep) {
			t.Errorf("path %q must stay reviewable — only the exact settings file and the config directory are excluded", keep)
		}
	}
}

// --- R5: a rename does not inherit the source's exclusion ------------------

// TestReviewableFromPorcelainRenameKeepsDestination is the R5 repro on the
// shared parser: a rename record carrying the config file as its SOURCE must
// not exclude the DESTINATION — `.claude/settings.json -> main.go` is a real
// source change (the destination file) and stays reviewable.
func TestReviewableFromPorcelainRenameKeepsDestination(t *testing.T) {
	if !reviewableFromPorcelain("R  .claude/settings.json -> cmd/moai/main.go\n") {
		t.Fatalf("a rename into an ordinary source path must stay reviewable — the destination must not inherit the source's exclusion")
	}
}
