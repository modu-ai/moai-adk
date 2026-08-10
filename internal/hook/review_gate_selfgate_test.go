package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The two review-gate wrappers are registered in the Stop array, and both gates
// ship OFF. A wrapper that resolved and exec'd the moai binary unconditionally
// would therefore add two cold starts to every turn-end for every user. These
// tests are the runtime counterpart of the static ordering check in
// internal/template: they run each wrapper against a temp project and count
// whether the binary was reached.
//
// The counting-stub approach mirrors the Stop-chain trim tests: a stub `moai`
// goes first in PATH, HOME is redirected so the 3-tier fallback finds no real
// binary, and the counter file records whether the stub ran.

// reviewGateWorkflowYAML builds a workflow.yaml in the deployed nested shape
// with the named gate set to enabled.
func reviewGateWorkflowYAML(gate string, enabled bool) string {
	state := "false"
	if enabled {
		state = "true"
	}
	return "workflow:\n" +
		"    execution_mode: auto\n" +
		"    " + gate + ":\n" +
		"        review_gate:\n" +
		"            enabled: " + state + "\n"
}

// runReviewGateWrapper runs the wrapper from the repo against projectDir and
// returns how many times the stub moai binary was invoked.
func runReviewGateWrapper(t *testing.T, script, projectDir string) int {
	t.Helper()
	repoRoot := filepath.Join(mustGetwd(t), "..", "..")
	path := filepath.Join(repoRoot, ".claude", "hooks", "moai", script)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("wrapper not found: %s (%v)", path, err)
	}

	tmp := t.TempDir()
	stubBin := filepath.Join(tmp, "bin")
	homeDir := filepath.Join(tmp, "home")
	for _, d := range []string{stubBin, homeDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	counter := filepath.Join(tmp, "moai.count")
	writeCountingStub(t, stubBin, "moai", counter)

	cmd := exec.Command("bash", path)
	cmd.Stdin = strings.NewReader(`{"session_id":"sess-review-gate","cwd":"` + projectDir + `","stop_hook_active":false}`)
	cmd.Dir = projectDir
	cmd.Env = []string{
		"PATH=" + stubBin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + homeDir,
		"CLAUDE_PROJECT_DIR=" + projectDir,
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run %s: %v; output=%s", script, err, out)
	}
	return readCounter(t, counter)
}

// newReviewGateProject creates a temp project whose workflow.yaml sets the
// named gate to enabled. Passing an empty body leaves workflow.yaml absent.
func newReviewGateProject(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(cfgDir, "workflow.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestReviewGateWrappers_DisabledCostsZeroColdStarts is the cost constraint:
// with the gate off — or its config absent, or the key spelled some way the
// conservative parser cannot read — the wrapper must exit 0 without ever
// reaching the moai binary.
func TestReviewGateWrappers_DisabledCostsZeroColdStarts(t *testing.T) {
	requireBash(t)
	for _, tc := range []struct {
		script, gate, name, body string
	}{
		{"handle-codex-review-gate.sh", "codex", "explicit false", reviewGateWorkflowYAML("codex", false)},
		{"handle-codex-review-gate.sh", "codex", "config absent", ""},
		{"handle-codex-review-gate.sh", "codex", "sibling gate on only", reviewGateWorkflowYAML("multi", true)},
		{"handle-codex-review-gate.sh", "codex", "flat key not honoured", "codex:\n    review_gate:\n        enabled: true\n"},
		{"handle-multi-review-gate.sh", "multi", "explicit false", reviewGateWorkflowYAML("multi", false)},
		{"handle-multi-review-gate.sh", "multi", "config absent", ""},
		{"handle-multi-review-gate.sh", "multi", "sibling gate on only", reviewGateWorkflowYAML("codex", true)},
		{"handle-multi-review-gate.sh", "multi", "flat key not honoured", "multi:\n    review_gate:\n        enabled: true\n"},
	} {
		t.Run(tc.gate+"/"+tc.name, func(t *testing.T) {
			project := newReviewGateProject(t, tc.body)
			if got := runReviewGateWrapper(t, tc.script, project); got != 0 {
				t.Fatalf("%s invoked the moai binary %d time(s) with the gate off; want 0 cold starts", tc.script, got)
			}
		})
	}
}

// TestReviewGateWrappers_EnabledReachesBinary is the other direction: once the
// opt-in is set at its real nested key, the wrapper must actually reach the
// binary. Without this, a self-gate that always short-circuits would look like
// a pass above while leaving the gate as unreachable as before.
func TestReviewGateWrappers_EnabledReachesBinary(t *testing.T) {
	requireBash(t)
	for _, tc := range []struct{ script, gate string }{
		{"handle-codex-review-gate.sh", "codex"},
		{"handle-multi-review-gate.sh", "multi"},
	} {
		t.Run(tc.gate, func(t *testing.T) {
			project := newReviewGateProject(t, reviewGateWorkflowYAML(tc.gate, true))
			if got := runReviewGateWrapper(t, tc.script, project); got != 1 {
				t.Fatalf("%s invoked the moai binary %d time(s) with the gate ON; want exactly 1", tc.script, got)
			}
		})
	}
}

// TestReviewGateWrappers_ToleratesCommentsAndIndentation pins that the
// conservative parser still reads a true through the shapes a real config
// carries: trailing comments and a differently-indented document.
func TestReviewGateWrappers_ToleratesCommentsAndIndentation(t *testing.T) {
	requireBash(t)
	for _, tc := range []struct{ script, gate, body, name string }{
		{
			"handle-codex-review-gate.sh", "codex",
			"workflow:\n  codex:\n    review_gate:\n      enabled: true # opted in\n",
			"two-space indent with trailing comment",
		},
		{
			"handle-multi-review-gate.sh", "multi",
			"# leading comment\nworkflow:\n\n    multi:\n        review_gate:\n            enabled: true\n",
			"blank lines and a leading comment",
		},
	} {
		t.Run(tc.gate+"/"+tc.name, func(t *testing.T) {
			project := newReviewGateProject(t, tc.body)
			if got := runReviewGateWrapper(t, tc.script, project); got != 1 {
				t.Fatalf("%s: parser missed an enabled gate (%s); binary invoked %d time(s), want 1", tc.script, tc.name, got)
			}
		})
	}
}
