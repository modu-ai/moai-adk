package graph

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// gitFix runs a git command inside dir, failing the test on error. Fixture
// repositories live under t.TempDir() (test-isolation contract) and are real
// git repos so the codemaps layer's endpoint comparison exercises real git.
func gitFix(t *testing.T, dir string, args ...string) string {
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

// newCheckFixture creates a git repo with described sources (internal/, cmd/,
// pkg/), commits them, and returns the root. The mx sidecar and edges
// artifacts are NOT created — each test builds the state it needs.
func newCheckFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitFix(t, root, "init", "-q")
	gitFix(t, root, "config", "user.email", "fixture@example.com")
	gitFix(t, root, "config", "user.name", "Fixture")

	for _, p := range []string{
		"internal/alpha/alpha.go",
		"internal/beta/beta.go",
		"cmd/tool/main.go",
		"pkg/lib/lib.go",
	} {
		abs := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("package "+filepath.Base(filepath.Dir(abs))+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "fixture base")
	return root
}

// writeCodemapsProvenance writes a clean codemaps provenance sidecar stamped
// at the current HEAD.
func writeCodemapsProvenance(t *testing.T, root, commit string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "{\n  \"schema_version\": 1,\n  \"tree_root\": " + quoteJSON(root) +
		",\n  \"commit_sha\": " + quoteJSON(commit) + ",\n  \"dirty\": false,\n" +
		"  \"generated_by\": \"codemaps-gen\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "provenance.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func quoteJSON(s string) string {
	return `"` + strings.ReplaceAll(s, `\`, `\\`) + `"`
}

// AC-GF-001 — numeric per-layer report on a fully in-sync fixture: every layer
// reports name + metric kind + integer value + threshold + verdict fresh, and
// CheckFreshness reports no failure.
func TestCheckFreshness_AllFresh(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")

	writeCodemapsProvenance(t, root, head)
	writeSyncedMXIndex(t, root)
	writeSyncedEdgesMeta(t, root)

	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	if res.Failed() {
		t.Fatalf("expected no failure, got %+v", res.Layers)
	}
	byLayer := map[string]LayerReport{}
	for _, l := range res.Layers {
		byLayer[l.Layer] = l
	}
	if len(res.Layers) != 3 {
		t.Fatalf("expected 3 layer reports, got %d: %+v", len(res.Layers), res.Layers)
	}
	for _, name := range []string{"codemaps", "mx-index", "edges"} {
		l, ok := byLayer[name]
		if !ok {
			t.Fatalf("missing layer report for %q", name)
		}
		if l.Verdict != VerdictFresh {
			t.Errorf("layer %s verdict = %q, want fresh (reason: %s)", name, l.Verdict, l.Reason)
		}
		if l.Metric == "" {
			t.Errorf("layer %s metric kind is empty", name)
		}
		if l.Value != 0 {
			t.Errorf("layer %s value = %d, want 0 on an in-sync fixture", name, l.Value)
		}
	}
}

// AC-GF-002 — per-layer metric correctness: each layer derives its value from
// its own metric, not a shared formula.
func TestCheckFreshness_PerLayerMetrics(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenance(t, root, head)
	// The scanner read only alpha — inventory is the subset actually read.
	writeSyncedMXIndex(t, root, "internal/alpha/alpha.go")
	writeSyncedEdgesMeta(t, root)

	// (a) 3 described-source files change after the stamped generation commit
	// (uncommitted — endpoint content differs from the stamped commit).
	for _, p := range []string{
		"internal/alpha/alpha.go",
		"internal/beta/beta.go",
		"cmd/tool/main.go",
	} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(p)),
			[]byte("package changed\n// edited\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// (c) the codemaps content moves past the edges build's stamped
	// fingerprint: edit a codemaps file AFTER the meta was stamped.
	if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps", "dependencies.md"),
		[]byte("# Deps\n// regenerated later\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	byLayer := map[string]LayerReport{}
	for _, l := range res.Layers {
		byLayer[l.Layer] = l
	}

	// codemaps counts endpoint-differing described-source files vs the stamped
	// commit: alpha, beta, cmd/tool (3).
	if got := byLayer["codemaps"].Value; got != 3 {
		t.Errorf("codemaps value = %d, want 3 (endpoint diff vs stamped commit, reason %s)", got, byLayer["codemaps"].Reason)
	}
	// mx-index counts inventory files whose current content hash differs:
	// only alpha is inventoried, and it changed (1).
	if got := byLayer["mx-index"].Value; got != 1 {
		t.Errorf("mx-index value = %d, want 1 (one inventoried file changed)", got)
	}
	// edges: the codemaps source set's fingerprint no longer matches the
	// stamped one — its own metric, not a shared formula.
	if v := byLayer["edges"]; v.Verdict != VerdictStale || v.Value < 1 {
		t.Errorf("edges = %+v, want stale with >=1 fingerprint mismatch", v)
	}
}

// AC-GF-004 / MUTANT A red input — a deliberately aged layer must FAIL the
// check (Failed() true), the single binary observable the exit code consumes.
// A reporting-only implementation (prints staleness, always passes) fails here.
func TestCheckFreshness_AgedLayerFails(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenance(t, root, head)
	writeSyncedMXIndex(t, root)
	writeSyncedEdgesMeta(t, root)

	// Age codemaps past the default threshold (>= 40 changed described-source
	// files since the stamped commit).
	for i := 0; i < DefaultThresholds().CodemapsChangedFiles; i++ {
		p := filepath.Join(root, "internal", "alpha", filepath.FromSlash("gen"+strings.Repeat("0", 3-len(itoa(i)))+itoa(i)+".go"))
		if err := os.WriteFile(p, []byte("package alpha\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	if !res.Failed() {
		t.Fatal("MUTANT A red input: aged codemaps must fail the check — a reporting-only implementation always exits 0")
	}
	// The failing layer names value and threshold.
	for _, l := range res.Layers {
		if l.Layer == "codemaps" {
			if l.Verdict != VerdictStale {
				t.Errorf("codemaps verdict = %q, want stale", l.Verdict)
			}
			if l.Value < l.Threshold {
				t.Errorf("codemaps value %d < threshold %d — verdict must follow the measured value", l.Value, l.Threshold)
			}
		}
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

// AC-GF-005 — absent is a distinct, failing verdict: a fresh worktree state
// (no mx-index, no edges.jsonl) reports absent per layer with a reason, never
// 0-as-fresh.
func TestCheckFreshness_AbsentLayers(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenance(t, root, head)
	// No mx-index, no edges artifacts — the fresh-worktree state.

	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	if !res.Failed() {
		t.Fatal("absent layers must fail the check (REQ-GF-004 absent clause)")
	}
	byLayer := map[string]LayerReport{}
	for _, l := range res.Layers {
		byLayer[l.Layer] = l
	}
	for _, name := range []string{"mx-index", "edges"} {
		l := byLayer[name]
		if l.Verdict != VerdictAbsent {
			t.Errorf("layer %s verdict = %q, want absent", name, l.Verdict)
		}
		if l.Reason == "" {
			t.Errorf("layer %s absent verdict carries no reason string", name)
		}
		if l.Verdict == VerdictAbsent && l.Value == 0 && l.Metric == "" {
			t.Errorf("layer %s absent must not be reported as a bare zero value", name)
		}
	}
}

// AC-GF-003 tail — an artifact whose provenance block is removed is
// unjudgeable (absent-equivalent), never silently fresh.
func TestCheckFreshness_NoProvenanceIsNotFresh(t *testing.T) {
	root := newCheckFixture(t)

	// codemaps dir exists, provenance sidecar absent.
	dir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dependencies.md"), []byte("# Deps\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSyncedMXIndex(t, root)
	writeSyncedEdgesMeta(t, root)

	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	if !res.Failed() {
		t.Fatal("codemaps without a provenance block must not be treated as fresh")
	}
	for _, l := range res.Layers {
		if l.Layer == "codemaps" && l.Verdict != VerdictAbsent {
			t.Errorf("codemaps verdict = %q, want absent (no provenance — unjudgeable)", l.Verdict)
		}
	}
}

// REQ-GF-003 dirty-anchor clause — a dirty-generation stamp is compared via
// its content fingerprint, not a named commit.
func TestCheckFreshness_DirtyGenerationAnchor(t *testing.T) {
	root := newCheckFixture(t)

	// Build the aggregate fingerprint the dirty stamp would carry: hash of the
	// described roots' current content. Mutate one file first so the stamped
	// fingerprint (computed AFTER the edit) matches the current tree → fresh.
	if err := os.WriteFile(filepath.Join(root, "internal", "alpha", "alpha.go"),
		[]byte("package alpha\n// post-stamp edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fp, err := mx.AggregateDescribedFingerprint(root, mx.DefaultDescribedRoots)
	if err != nil {
		t.Fatalf("AggregateDescribedFingerprint: %v", err)
	}
	dir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "{\n  \"schema_version\": 1,\n  \"tree_root\": " + quoteJSON(root) +
		",\n  \"commit_sha\": \"\",\n  \"dirty\": true,\n" +
		"  \"content_fingerprint\": " + quoteJSON(fp) + ",\n  \"generated_by\": \"codemaps-gen\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "provenance.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSyncedMXIndex(t, root)
	writeSyncedEdgesMeta(t, root)

	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	for _, l := range res.Layers {
		if l.Layer == "codemaps" && l.Verdict != VerdictFresh {
			t.Errorf("dirty-stamped codemaps with matching fingerprint = %q (%s), want fresh", l.Verdict, l.Reason)
		}
	}

	// Now move the content past the stamped fingerprint — mismatch must stale.
	if err := os.WriteFile(filepath.Join(root, "internal", "beta", "beta.go"),
		[]byte("package beta\n// later edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res2, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness(2): %v", err)
	}
	staled := false
	for _, l := range res2.Layers {
		if l.Layer == "codemaps" && l.Verdict == VerdictStale {
			staled = true
		}
	}
	if !staled {
		t.Error("dirty-stamped codemaps with moved content must go stale via fingerprint mismatch")
	}
}

// EdgesSourcesMoved probe — the cheap rebuild-free staleness signal the
// query paths consult (REQ-GF-007). No meta ⇒ moved; fresh meta ⇒ not moved;
// a moved source ⇒ moved.
func TestEdgesSourcesMoved_ProbeStates(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")

	if !EdgesSourcesMoved(root) {
		t.Error("no meta sidecar must probe as moved (unjudgeable artifact)")
	}

	writeCodemapsProvenance(t, root, head)
	writeSyncedMXIndex(t, root)
	writeSyncedEdgesMeta(t, root)
	if EdgesSourcesMoved(root) {
		t.Error("in-sync meta must probe as not moved")
	}

	// Move a source the meta covers: edit the codemaps directory content.
	if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps", "dependencies.md"),
		[]byte("# moved\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !EdgesSourcesMoved(root) {
		t.Error("a moved codemaps source must probe as moved")
	}
}

// MXIndexNeedsRefresh probe — absent ⇒ refresh (scan sources unindexed);
// pre-provenance ⇒ NOT refreshed (legacy contract); drift / wrong-tree ⇒
// refresh; in-sync ⇒ not.
func TestMXIndexNeedsRefresh_ProbeStates(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")

	if !MXIndexNeedsRefresh(root) {
		t.Error("absent sidecar must need refresh (scan sources unindexed)")
	}

	// Pre-provenance sidecar: the legacy answering contract — never refreshed.
	stateDir := filepath.Join(root, ".moai", "state")
	mgr := mx.NewManager(stateDir)
	if err := mgr.Write(&mx.Sidecar{SchemaVersion: mx.SchemaVersion}); err != nil {
		t.Fatal(err)
	}
	if MXIndexNeedsRefresh(root) {
		t.Error("pre-provenance sidecar must NOT be refreshed (legacy contract)")
	}

	// In-sync provenance: not needed.
	writeSyncedMXIndex(t, root)
	if MXIndexNeedsRefresh(root) {
		t.Error("in-sync index must not need refresh")
	}

	// Content drift in an inventoried file: needed.
	if err := os.WriteFile(filepath.Join(root, "internal", "alpha", "alpha.go"),
		[]byte("package changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !MXIndexNeedsRefresh(root) {
		t.Error("inventory drift must need refresh")
	}

	// Wrong-tree anchor: needed (wrong-tree defect family).
	writeMXIndexProvenance(t, root, "/some/other/tree", head)
	if !MXIndexNeedsRefresh(root) {
		t.Error("wrong-tree index must need refresh")
	}
}

// Wrong-tree mx-index (provenance tree_root names a different tree) is stale,
// never fresh — the CR #8/t246 wrong-tree defect family.
func TestCheckFreshness_WrongTreeIndex(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenance(t, root, head)
	writeSyncedEdgesMeta(t, root)
	writeMXIndexProvenance(t, root, "/some/other/tree", head)

	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	for _, l := range res.Layers {
		if l.Layer == "mx-index" {
			if l.Verdict != VerdictStale {
				t.Errorf("wrong-tree index verdict = %q, want stale", l.Verdict)
			}
		}
	}
}

// writeSyncedMXIndex writes an mx sidecar whose provenance inventory matches
// the CURRENT content of the fixture's described files (in-sync state), using
// the production stamping path so fixture state and product agree on shape.
// Only the listed rels are inventoried — the real scanner inventories the
// files it actually read, which is a subset of the described sources.
func writeSyncedMXIndex(t *testing.T, root string, rels ...string) {
	t.Helper()
	if len(rels) == 0 {
		rels = []string{"internal/alpha/alpha.go", "internal/beta/beta.go", "cmd/tool/main.go", "pkg/lib/lib.go"}
	}
	inv := map[string]string{}
	for _, rel := range rels {
		sum, err := mx.HashFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("hash %s: %v", rel, err)
		}
		inv[rel] = sum
	}
	pv := mx.StampMXScan(root, inv)
	stateDir := filepath.Join(root, ".moai", "state")
	mgr := mx.NewManager(stateDir)
	if err := mgr.Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		Tags:          nil,
		ScannedAt:     time.Now(),
		Provenance:    pv,
	}); err != nil {
		t.Fatal(err)
	}
}

// writeMXIndexProvenance writes an mx sidecar whose provenance names an
// explicit (possibly wrong) tree root — the wrong-tree fixture.
func writeMXIndexProvenance(t *testing.T, root, treeRoot, commit string) {
	t.Helper()
	stateDir := filepath.Join(root, ".moai", "state")
	mgr := mx.NewManager(stateDir)
	if err := mgr.Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		Provenance: &mx.Provenance{
			SchemaVersion: 1,
			TreeRoot:      treeRoot,
			CommitSHA:     commit,
			FileInventory: map[string]string{"internal/alpha/alpha.go": "00"},
		},
	}); err != nil {
		t.Fatal(err)
	}
}

// writeSyncedEdgesMeta writes edges.jsonl plus a .meta.json sidecar whose
// source fingerprints match the CURRENT sources (in-sync derived state).
func writeSyncedEdgesMeta(t *testing.T, root string) {
	t.Helper()
	graphDir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(graphDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(graphDir, "edges.jsonl"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	fp := SourceFingerprintsForEdges(root)
	if err := WriteEdgesMeta(filepath.Join(graphDir, "edges.meta.json"), root, fp); err != nil {
		t.Fatal(err)
	}
}
