package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeCitationsFixture lays down a minimal tree with real paths the cited
// tokens can resolve against, and one codemaps doc whose content the test
// controls (REQ-CMA-002 fixture surface).
func writeCitationsFixture(t *testing.T, doc string) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"internal/cli", "internal/graph", "internal/kanban", "cmd/moai"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"cmd/moai/main.go", "internal/cli/factory.go", "internal/graph/check.go", "internal/kanban/record.go"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(f)), []byte("package x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	dir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "modules.md"), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// AC-CMA-005 direction 1 + the mutant round-trip: a positive (non-blockquote)
// citation of an absent path must turn the citations layer stale and make
// CheckResult.Failed() true; removing the phantom restores fresh. The
// injection→red→restore→green cycle is the observation record that proves the
// check cannot be vacuously green.
func TestCheckCitations_PositivePhantomRed(t *testing.T) {
	base := `# modules

### internal/kanban

Entry points live in internal/cli/factory.go and cmd/moai/main() calls into
internal/graph/check.go.
`
	withPhantom := base + "Cross-check internal/zzz-phantom for the legacy flow.\n"

	// RED side: fabricated positive phantom (mutant) → stale, Failed.
	root := writeCitationsFixture(t, withPhantom)
	rep := checkCitations(root)
	if rep.Layer != LayerCitations {
		t.Fatalf("layer = %q, want %q", rep.Layer, LayerCitations)
	}
	if rep.Metric != MetricPositiveCitedPathAbsence {
		t.Fatalf("metric = %q, want %q", rep.Metric, MetricPositiveCitedPathAbsence)
	}
	if rep.Verdict != VerdictStale {
		t.Fatalf("verdict = %q, want stale (positive phantom must be red)", rep.Verdict)
	}
	if rep.Value != 1 {
		t.Fatalf("value = %d, want 1 (exactly the injected phantom)", rep.Value)
	}
	if rep.Threshold != 0 {
		t.Fatalf("threshold = %d, want 0 (any positive phantom is red)", rep.Threshold)
	}
	res := CheckResult{TreeRoot: root, Layers: []LayerReport{rep}}
	if !res.Failed() {
		t.Fatal("CheckResult.Failed() = false, want true when citations layer is stale")
	}
	if len(res.OffendingLayers()) != 1 || res.OffendingLayers()[0].Layer != LayerCitations {
		t.Fatalf("OffendingLayers() did not name the citations layer: %+v", res.OffendingLayers())
	}
	if len(rep.DrivingPaths) != 1 || !strings.Contains(rep.DrivingPaths[0], "internal/zzz-phantom") {
		t.Fatalf("driving paths = %v, want the phantom path named", rep.DrivingPaths)
	}

	// GREEN side of the round-trip: restore the doc → fresh.
	root2 := writeCitationsFixture(t, base)
	rep2 := checkCitations(root2)
	if rep2.Verdict != VerdictFresh {
		t.Fatalf("verdict after restore = %q, want fresh; reason=%q", rep2.Verdict, rep2.Reason)
	}
	if rep2.Value != 0 {
		t.Fatalf("value after restore = %d, want 0", rep2.Value)
	}
}

// AC-CMA-005 direction 2 (D2 exemption): absent paths cited ONLY on
// blockquote (`>`-prefixed) lines are negative-context citations and must not
// count — the layer stays fresh.
func TestCheckCitations_BlockquoteExemption(t *testing.T) {
	doc := `# modules

> **[REMOVED]** ` + "`internal/design`" + ` 는 제거된 패키지입니다 — 존재하지 않음.

> ` + "`internal/bodp`" + ` 는 #1278 worktree surface redesign에서 제거되었다.

### internal/kanban

See internal/kanban/record.go for the registry record.
`
	root := writeCitationsFixture(t, doc)
	rep := checkCitations(root)
	if rep.Verdict != VerdictFresh {
		t.Fatalf("verdict = %q, want fresh (blockquote-only absence is exempt); reason=%q driving=%v",
			rep.Verdict, rep.Reason, rep.DrivingPaths)
	}
}

// AC-CMA-005 direction 3 (normalization rules, REQ-CMA-002): trailing-slash
// strip, trailing punctuation trim, `cmd/moai/main` → `cmd/moai/main.go`, and
// `.go`-suffix restore each resolve a real tree path instead of a phantom.
func TestNormalizeCitedPath(t *testing.T) {
	root := writeCitationsFixture(t, "# doc\n")
	cases := []struct {
		raw  string
		want string
	}{
		{"internal/cli/", "internal/cli"},                           // trailing slash
		{"internal/cli,", "internal/cli"},                           // trailing punctuation
		{"internal/kanban/record.go.", "internal/kanban/record.go"}, // trailing period after .go
		{"cmd/moai/main", "cmd/moai/main.go"},                       // call-chain map (spec §1.1 P8)
		{"internal/graph/checkgo", "internal/graph/check.go"},       // .go-suffix restore (stripped period artifact)
		{"internal/zzz-phantom", "internal/zzz-phantom"},            // no rule invents existence
	}
	for _, tc := range cases {
		if got := normalizeCitedPath(root, tc.raw); got != tc.want {
			t.Errorf("normalizeCitedPath(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

// A missing codemaps directory leaves the layer unjudgeable — absent, never
// silently fresh (the existing layer contract).
func TestCheckCitations_AbsentDirMissing(t *testing.T) {
	root := t.TempDir()
	rep := checkCitations(root)
	if rep.Verdict != VerdictAbsent {
		t.Fatalf("verdict = %q, want absent when codemaps directory is missing", rep.Verdict)
	}
	if rep.Reason == "" {
		t.Fatal("absent verdict must carry a reason")
	}
}

// The citations row rides the existing CheckFreshness layer walk: four layer
// rows, citations last, and a phantom in the doc propagates to
// CheckResult.Failed() through the same consumer path (REQ-CMA-002 D1).
func TestCheckFreshness_IncludesCitationsLayer(t *testing.T) {
	doc := "# modules\n\nCross-check internal/zzz-phantom for the legacy flow.\n"
	root := writeCitationsFixture(t, doc)
	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness on a fixture tree must not system-error: %v", err)
	}
	if len(res.Layers) != 4 {
		t.Fatalf("layer count = %d, want 4 (codemaps, mx-index, edges, citations): %+v", len(res.Layers), res.Layers)
	}
	last := res.Layers[len(res.Layers)-1]
	if last.Layer != LayerCitations || last.Verdict != VerdictStale {
		t.Fatalf("last layer = %+v, want citations/stale", last)
	}
	if !res.Failed() {
		t.Fatal("Failed() = false, want true with a positive phantom present")
	}
	names := make([]string, 0, len(res.Layers))
	for _, l := range res.Layers {
		names = append(names, l.Layer)
	}
	if got, want := strings.Join(names, ","), "codemaps,mx-index,edges,citations"; got != want {
		t.Fatalf("layer order = %q, want %q", got, want)
	}
}
