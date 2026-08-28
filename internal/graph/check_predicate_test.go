package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// writeFixtureFile writes a repo-relative slash path under root.
func writeFixtureFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// AC-GFC-002 — gitDiffNameCount counts only described-worthy paths, across
// BOTH of its branches. The fixture differs from the stamp by one production
// .go file and forty testdata fixtures, and one of the forty is left
// UNTRACKED so that a predicate wired into the `git diff` branch alone (the
// mutant) reaches the untracked file unfiltered and returns 2.
func TestGitDiffNameCount_Predicate(t *testing.T) {
	root := newCheckFixture(t)
	stamp := gitFix(t, root, "rev-parse", "HEAD")

	// One production .go change — the only described-worthy difference.
	writeFixtureFile(t, root, "internal/alpha/alpha.go", "package alpha\n\nfunc A() {}\n")

	// Thirty-nine tracked testdata fixtures, committed after the stamp so the
	// `git diff --name-only <stamp>` branch sees them.
	for i := 1; i <= 39; i++ {
		writeFixtureFile(t, root,
			fmt.Sprintf("internal/astgrep/testdata/rule-tests/f%02d.yml", i),
			fmt.Sprintf("id: f%02d\n", i))
	}
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "fixture churn")

	// The fortieth fixture stays untracked — and is a .go file, so the
	// untracked branch must apply the testdata rule, not merely the suffix.
	writeFixtureFile(t, root, "internal/astgrep/testdata/rule-tests/f40.go", "package ruletests\n")

	got, err := gitDiffNameCount(root, stamp, mx.DefaultDescribedRoots)
	if err != nil {
		t.Fatalf("gitDiffNameCount: %v", err)
	}
	if got != 1 {
		t.Fatalf("gitDiffNameCount = %d, want 1 (41 = no predicate; 2 = predicate on the git diff branch only)", got)
	}
}

// AC-GFC-014 — the edges layer's source-set fingerprints are computed over the
// UNFILTERED aggregate. The mutant this kills is the predicate pushed down into
// aggregateFingerprint: .moai/project/codemaps and .moai/specs hold zero .go
// files, so under a .go-only filter both collapse to the empty-entry hash.
func TestSourceFingerprintsForEdges_Unchanged(t *testing.T) {
	const emptyEntriesHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	root := newCheckFixture(t)
	writeFixtureFile(t, root, ".moai/project/codemaps/overview.md", "# overview\n")
	writeFixtureFile(t, root, ".moai/specs/SPEC-X-001/spec.md", "# spec\n")
	writeFixtureFile(t, root, ".moai/reports/t000/verdict.md", "# verdict\n")

	fp := SourceFingerprintsForEdges(root)
	for _, key := range []string{srcCodemaps, srcSpecs, srcReports} {
		if fp[key] == "" {
			t.Fatalf("source set %q missing from the edges fingerprint map", key)
		}
	}
	for _, key := range []string{srcCodemaps, srcSpecs} {
		if fp[key] == emptyEntriesHash {
			t.Errorf("source set %q collapsed to the empty-entry hash — the predicate reached aggregateFingerprint", key)
		}
	}

	// A non-.go edit inside a source set must still move its fingerprint: the
	// inequality above pins the empty case, this pins the live behaviour.
	writeFixtureFile(t, root, ".moai/project/codemaps/overview.md", "# overview (edited)\n")
	after := SourceFingerprintsForEdges(root)
	if after[srcCodemaps] == fp[srcCodemaps] {
		t.Errorf("codemaps source fingerprint did not move on a non-.go edit — the edges layer is permanently green")
	}
	if after[srcSpecs] != fp[srcSpecs] {
		t.Errorf("specs source fingerprint moved on an unrelated edit")
	}
}
