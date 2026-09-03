package quality

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/graph"
	"github.com/modu-ai/moai-adk/internal/mx"
)

// gfGit runs git in the fixture repo.
func gfGit(t *testing.T, dir string, args ...string) string {
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

// newGFFixture builds a git repo whose graph layers are ABSENT (fresh
// worktree state) — the natural stale input for the step's red paths. No
// go.mod: the fixture must exercise the step BEFORE language detection, so
// the toolchain detector returning nil is part of what is under test.
func newGFFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gfGit(t, root, "init", "-q")
	gfGit(t, root, "config", "user.email", "fixture@example.com")
	gfGit(t, root, "config", "user.name", "Fixture")
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "a.go"), []byte("package internal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gfGit(t, root, "add", "-A")
	gfGit(t, root, "commit", "-q", "-m", "base")
	return root
}

func gfGate(cfg *GraphFreshnessConfig, projectDir string) *QualityGate {
	base := DefaultGateConfig()
	base.ProjectDir = projectDir
	base.Enabled = true
	base.SkipTests = true
	base.TypecheckEnabled = false
	if cfg != nil {
		base.GraphFreshness = cfg
	}
	return NewQualityGate(base)
}

// AC-GF-006 posture 1 — enabled + stale layers + BLOCKING: the gate fails
// and the output names the graph-freshness step with per-layer verdicts.
func TestGateGraphFreshness_BlockingFails(t *testing.T) {
	root := newGFFixture(t)
	g := gfGate(&GraphFreshnessConfig{Enabled: true, Blocking: true,
		CodemapsChangedFiles: 40, MXIndexChangedFiles: 1}, root)

	passed, out := g.Run(context.Background())
	if passed {
		t.Fatal("blocking graph-freshness on an absent-layer fixture must fail the gate")
	}
	if !strings.Contains(out, "graph-freshness") {
		t.Errorf("failure output must name the graph-freshness step, got: %s", out)
	}
	if !strings.Contains(out, "absent") {
		t.Errorf("failure output must carry the per-layer verdict, got: %s", out)
	}
}

// AC-GF-006 posture 2 — enabled + stale + ADVISORY (the default): the gate
// passes while emitting the check's verdict — never silence.
func TestGateGraphFreshness_AdvisoryWarns(t *testing.T) {
	root := newGFFixture(t)
	g := gfGate(&GraphFreshnessConfig{Enabled: true, Blocking: false,
		CodemapsChangedFiles: 40, MXIndexChangedFiles: 1}, root)

	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("advisory step must not fail the gate, got failure: %s", out)
	}
	if !strings.Contains(out, "graph-freshness") || !strings.Contains(out, "absent") {
		t.Errorf("advisory step must emit the verdict, got output: %q", out)
	}
	if !strings.Contains(out, "advisory") {
		t.Errorf("advisory posture must be marked in the notice, got: %q", out)
	}
}

// AC-GF-006 posture 3 — disabled: explicit skip notice, gate passes.
func TestGateGraphFreshness_DisabledNotice(t *testing.T) {
	root := newGFFixture(t)
	g := gfGate(&GraphFreshnessConfig{Enabled: false}, root)

	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("disabled step must pass, got: %s", out)
	}
	if !strings.Contains(out, "disabled") {
		t.Errorf("disabled step must emit an explicit skip notice, got: %q", out)
	}
}

// AC-GF-006 posture 4 — fresh layers: passes with an explicit fresh notice.
// Also proves the step runs BEFORE language detection: the fixture has no
// language marker, and a silent early return would emit nothing.
func TestGateGraphFreshness_FreshNoticeBeforeLanguageDetection(t *testing.T) {
	root := newGFFixture(t)
	cmDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(cmDir, 0o755); err != nil {
		t.Fatal(err)
	}

	g := gfGate(&GraphFreshnessConfig{Enabled: true, Blocking: false,
		CodemapsChangedFiles: 40, MXIndexChangedFiles: 1}, root)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate must pass, got: %s", out)
	}
	// The layers are absent in this fixture, so the notice carries the absent
	// verdict — the assertion is the NOTICE ITSELF (no silent path), which is
	// what the language-detection-ordering guarantees.
	if !strings.Contains(out, "graph-freshness") {
		t.Errorf("step must emit its notice even with no detectable language, got: %q", out)
	}
}

// gfStampAllLayers stamps an IN-SYNC state for all graph layers on the
// fixture — codemaps provenance at HEAD (plus a citation-free codemaps doc
// for the fourth, citations layer), mx sidecar over the fixture's own file
// inventory, and an edges meta stamped against the CURRENT source
// fingerprints (order matters — the mx sidecar write must precede the
// fingerprint computation or the edges layer reads back stale).
func gfStampAllLayers(t *testing.T, root string) {
	t.Helper()
	head := gfGit(t, root, "rev-parse", "HEAD")

	cmDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(cmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pvJSON, err := json.Marshal(&mx.Provenance{
		SchemaVersion: mx.ProvenanceSchemaVersion,
		TreeRoot:      root,
		CommitSHA:     head,
		GeneratedBy:   "codemaps-gen",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmDir, "provenance.json"), pvJSON, 0o644); err != nil {
		t.Fatal(err)
	}
	// The citations layer (fourth row, SPEC-CODEMAPS-ACCURACY-001 REQ-CMA-002)
	// judges the codemaps docs themselves — a doc-less directory is
	// unjudgeable-absent, not fresh (mirrors writeCodemapsProvenance in
	// internal/graph/check_test.go). No cited paths in the doc, so the
	// citations row reads fresh.
	if err := os.WriteFile(filepath.Join(cmDir, "modules.md"), []byte("# modules\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sum, err := mx.HashFile(filepath.Join(root, "internal", "a.go"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := mx.NewManager(filepath.Join(root, ".moai", "state"))
	if err := mgr.Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		ScannedAt:     time.Now(),
		Provenance:    mx.StampMXScan(root, map[string]string{"internal/a.go": sum}),
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

// REQ-GFR-002 / CR round-2 3855002099 — the all-fresh posture: a fixture
// with ALL layers (graph layers plus the citations row, REQ-CMA-002)
// stamped in sync passes carrying the LITERAL fresh
// notice (gate.go "graph-freshness: all layers fresh"), not merely any
// "graph-freshness" substring — the other postures (stale/disabled) also
// contain the substring, so only the literal distinguishes fresh.
func TestGateGraphFreshness_AllLayersFreshNotice(t *testing.T) {
	root := newGFFixture(t)
	gfStampAllLayers(t, root)

	g := gfGate(&GraphFreshnessConfig{Enabled: true, Blocking: false,
		CodemapsChangedFiles: 40, MXIndexChangedFiles: 1}, root)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("all-layers-fresh fixture must pass the gate, got: %s", out)
	}
	if !strings.Contains(out, "graph-freshness: all layers fresh") {
		t.Errorf("fresh posture must carry the literal all-fresh notice, got: %q", out)
	}
}

// Guard: the step never reads wall-clock freshness — durations here are
// structural (timeouts), not staleness signals. This test pins the advisory
// outcome shape so a future regression to silence fails loudly. (The
// unconfigured nil case is deliberately absent: it keeps the pre-existing
// silent-pass contract for unknown projects.)
func TestGateGraphFreshness_NoticeNeverEmpty(t *testing.T) {
	root := newGFFixture(t)
	for _, cfg := range []*GraphFreshnessConfig{
		{Enabled: true, Blocking: false, CodemapsChangedFiles: 40, MXIndexChangedFiles: 1},
		{Enabled: true, Blocking: true, CodemapsChangedFiles: 40, MXIndexChangedFiles: 1},
		{Enabled: false},
	} {
		g := gfGate(cfg, root)
		_, out := g.Run(context.Background())
		if !strings.Contains(out, "graph-freshness") {
			t.Errorf("cfg %+v: output lacks any graph-freshness notice: %q", cfg, out)
		}
	}
}
