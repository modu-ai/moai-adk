package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/mx"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// marshalCodemapsProvenance serializes a clean codemaps provenance sidecar
// via json.Marshal (CR round-2 3855001906): hand-interpolated JSON breaks on
// Windows temp paths' backslashes — a path like C:\Users\... yields invalid
// escape sequences the check cannot parse.
func marshalCodemapsProvenance(t *testing.T, root, commit string) []byte {
	t.Helper()
	data, err := json.Marshal(&mx.Provenance{
		SchemaVersion: mx.ProvenanceSchemaVersion,
		TreeRoot:      root,
		CommitSHA:     commit,
		GeneratedBy:   "codemaps-gen",
	})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// checkFixtureGit runs git in the CLI check fixture.
func checkFixtureGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v\nstderr: %s", args, err, errBuf.String())
	}
	return strings.TrimSpace(string(out))
}

// newCheckCLIRepo creates a committed fixture with described sources.
func newCheckCLIRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	checkFixtureGit(t, root, "init", "-q")
	checkFixtureGit(t, root, "config", "user.email", "fixture@example.com")
	checkFixtureGit(t, root, "config", "user.name", "Fixture")
	if err := os.MkdirAll(filepath.Join(root, "internal", "alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "alpha", "alpha.go"),
		[]byte("package alpha\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "-A")
	checkFixtureGit(t, root, "commit", "-q", "-m", "base")
	return root
}

// stampAllLayers stamps an in-sync provenance state for all four layers
// (citations joined as the fourth, SPEC-CODEMAPS-ACCURACY-001 REQ-CMA-002).
// The codemaps doc is part of a genuinely fully-stamped tree: the citations
// layer judges the .md docs themselves, and a doc-less codemaps directory is
// unjudgeable (absent), not fresh.
func stampAllLayers(t *testing.T, root string) {
	t.Helper()
	head := checkFixtureGit(t, root, "rev-parse", "HEAD")

	cmDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(cmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pvJSON := marshalCodemapsProvenance(t, root, head)
	if err := os.WriteFile(filepath.Join(cmDir, "provenance.json"), pvJSON, 0o644); err != nil {
		t.Fatal(err)
	}
	// One accurate doc: cites only the fixture's real described source.
	if err := os.WriteFile(filepath.Join(cmDir, "modules.md"),
		[]byte("# modules\n\ninternal/alpha/alpha.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sum, err := mx.HashFile(filepath.Join(root, "internal", "alpha", "alpha.go"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := mx.NewManager(filepath.Join(root, ".moai", "state"))
	if err := mgr.Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		ScannedAt:     time.Now(),
		Provenance:    mx.StampMXScan(root, map[string]string{"internal/alpha/alpha.go": sum}),
	}); err != nil {
		t.Fatal(err)
	}

	graphDir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(graphDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(graphDir, "edges.jsonl"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := graph.WriteEdgesMeta(filepath.Join(graphDir, graph.MetaFileName),
		root, graph.SourceFingerprintsForEdges(root), 0); err != nil {
		t.Fatal(err)
	}
}

// AC-GF-001 (CLI surface): in-sync fixture exits cleanly with a per-layer
// numeric report on stdout.
func TestGraphCheckCmd_AllFreshExitZero(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"check", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("in-sync check must exit clean, got: %v (stderr: %s)", err, errOut.String())
	}
	body := out.String()
	for _, want := range []string{"codemaps", "mx-index", "edges", "verdict=fresh"} {
		if !strings.Contains(body, want) {
			t.Errorf("report missing %q:\n%s", want, body)
		}
	}
	if !strings.Contains(body, "metric=") || !strings.Contains(body, "value=") || !strings.Contains(body, "threshold=") {
		t.Errorf("report must be numeric per layer (metric/value/threshold):\n%s", body)
	}
}

// AC-GF-004 (CLI surface, MUTANT A): a stale layer produces an error whose
// ExitCoder code is exactly 1 — the binary observable CI and moai gate
// consume. A reporting-only implementation returns nil here and fails.
func TestGraphCheckCmd_StaleExitsOne(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)

	// Age codemaps past the threshold: 40+ described-source files changed
	// since the stamped commit.
	if err := os.MkdirAll(filepath.Join(root, "internal", "aged"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 41; i++ {
		p := filepath.Join(root, "internal", "aged", "f"+string(rune('a'+i%26))+string(rune('a'+i/26))+".go")
		if err := os.WriteFile(p, []byte("package aged\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"check", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("MUTANT A: stale check must return an error (exit 1 path) — reporting-only implementations return nil")
	}
	var codeErr interface{ ExitCode() int }
	if !errors.As(err, &codeErr) {
		t.Fatalf("stale check error must carry an exit code, got %T: %v", err, err)
	}
	if got := codeErr.ExitCode(); got != 1 {
		t.Errorf("exit code = %d, want 1 (stale/absent)", got)
	}
	if !strings.Contains(errOut.String(), "codemaps") {
		t.Errorf("stderr must name the offending layer:\n%s", errOut.String())
	}
}

// CI run 32775609689 regression (F4): an unresolvable stamped commit exits 2
// (system error) with the missing commit named — never a fabricated stale.
func TestGraphCheckCmd_NotComparableExitsTwo(t *testing.T) {
	root := newCheckCLIRepo(t)
	// Stamp a plausible-looking SHA no history holds.
	cmDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(cmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pvJSON := marshalCodemapsProvenance(t, root, "1234567890abcdef1234567890abcdef12345678")
	if err := os.WriteFile(filepath.Join(cmDir, "provenance.json"), pvJSON, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"check", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err := cmd.Execute()
	var codeErr interface{ ExitCode() int }
	if !errors.As(err, &codeErr) || codeErr.ExitCode() != 2 {
		t.Fatalf("unresolvable stamp must exit 2 (system error), got %v", err)
	}
	if !strings.Contains(errOut.String(), "not comparable") {
		t.Errorf("exit-2 stderr must name the not-comparable condition:\n%s", errOut.String())
	}
}

// CR round-2 3855149230 (t261: failing input + observed red): a PRESENT but
// malformed gate.yaml exits 2 — never a silent defaults-based pass — while an
// absent config keeps the defaults path (exit reflects the check, not config).
func TestGraphCheckCmd_MalformedGateYamlExitsTwo(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)

	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "gate.yaml"),
		[]byte("gate: [unclosed\n  bad yaml: {{{"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"check", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.Execute()

	var codeErr interface{ ExitCode() int }
	if !errors.As(err, &codeErr) || codeErr.ExitCode() != 2 {
		t.Fatalf("malformed gate.yaml must exit 2 (system error), got %v", err)
	}
	if !strings.Contains(errOut.String(), "gate.yaml") {
		t.Errorf("exit-2 stderr must name the config: %s", errOut.String())
	}
}

// AC-GF-005 (CLI surface): absent untracked layers exit 1 with explicit
// absent verdicts — a fresh worktree state.
func TestGraphCheckCmd_AbsentExitsOne(t *testing.T) {
	root := newCheckCLIRepo(t)
	// codemaps provenance stamped; mx-index and edges deliberately absent.
	cmDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(cmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	head := checkFixtureGit(t, root, "rev-parse", "HEAD")
	pvJSON := marshalCodemapsProvenance(t, root, head)
	if err := os.WriteFile(filepath.Join(cmDir, "provenance.json"), pvJSON, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"check", "--root", root, "--json"})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err := cmd.Execute()
	var codeErr interface{ ExitCode() int }
	if !errors.As(err, &codeErr) || codeErr.ExitCode() != 1 {
		t.Fatalf("absent layers must exit 1, got %v", err)
	}
	if !strings.Contains(out.String(), `"absent"`) {
		t.Errorf("JSON report must carry absent verdicts:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "fresh-worktree state") &&
		!strings.Contains(out.String(), "untracked") {
		t.Errorf("absent verdicts must carry explicit reasons:\n%s", out.String())
	}
}

// CR round-2 3855001906 — the Windows-path breakage the marshal migration
// fixes: a backslash tree root marshals to VALID JSON and round-trips, while
// the retired hand-interpolation form on the same path is invalid JSON (the
// `\U` escape does not exist), which is exactly how the fixtures broke on
// Windows temp paths.
func TestProvenanceFixtureJSONRoundTripsWindowsPaths(t *testing.T) {
	const winRoot = `C:\Users\dev\proj`

	data, err := json.Marshal(&mx.Provenance{
		SchemaVersion: mx.ProvenanceSchemaVersion,
		TreeRoot:      winRoot,
		CommitSHA:     "1234567890abcdef1234567890abcdef12345678",
		GeneratedBy:   "codemaps-gen",
	})
	if err != nil {
		t.Fatal(err)
	}
	var back mx.Provenance
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("marshaled provenance must parse back: %v\nraw: %s", err, data)
	}
	if back.TreeRoot != winRoot {
		t.Errorf("tree root round-trip = %q, want %q", back.TreeRoot, winRoot)
	}

	// Negative control — the hand-interpolated form on the same root is not
	// valid JSON; a fixture built that way fails at parse time on Windows.
	handInterp := "{\"tree_root\": \"" + winRoot + "\"}"
	if json.Valid([]byte(handInterp)) {
		t.Errorf("hand-interpolated form unexpectedly valid: %s", handInterp)
	}
}
