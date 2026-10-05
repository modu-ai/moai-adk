package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The ordering-attribution doctrine guard for backlog card t300.
//
// verification-claim-integrity.md §2.3 ("the commit graph is the only
// sequencing witness") closes a recorded defect class: a baseline artifact
// that an acceptance criterion's ordering premise depends on is committed in
// the SAME commit as the change it measures, leaving the ordering claim
// permanently unverifiable — git snapshots trees per commit and cannot
// witness authoring order inside one commit.
//
// Doctrine text cannot execute; its only mechanical surface is presence in
// the trees actors actually read. Following the doctrine-surface pattern of
// evidence_citation_guard_test.go, this guard reads BOTH the repository copy
// and the template mirror so deleting the clause from either tree fails CI.
// Two vacuity traps are closed explicitly:
//
//   - Per-tree expectations diverge by design: the repository copy MUST name
//     the motivating instance (internal provenance), while the mirror MUST
//     NOT carry it (§25 template neutrality). A single shared literal would
//     let one tree satisfy the guard while the other goes stale.
//   - The section is located by heading, and a heading that is not found is
//     a hard failure naming the tree — a guard that silently scans an empty
//     range proves nothing.
const (
	vciOrderingRelPath = ".claude/rules/moai/core/verification-claim-integrity.md"
	// The clause heading and its one operative sentence. Present in BOTH trees.
	vciOrderingHeading   = "### 2.3 Ordering attribution"
	vciOrderingOperative = "only the commit graph witnesses it"
	// Provenance the LOCAL copy carries and the MIRROR must not (§25).
	vciOrderingProvenance = "SPEC-V3R6-GRAPH-FRESHNESS-001"
)

// vciOrderingMirrorForbidden are the internal tokens §25 keeps out of the
// distribution mirror, instantiated by this clause's motivating instance.
var vciOrderingMirrorForbidden = []string{
	"SPEC-V3R6-GRAPH-FRESHNESS-001",
	"AC-GF-022",
	"7f2e9e77d",
	"m5-baseline.md",
}

// vciOrderingSection extracts the §2.3 section body (heading line through the
// line before the next `## `-level heading or EOF). It fails the test naming
// the tree when the heading is absent — absence of the heading IS the clause
// having been deleted, the failure shape this guard exists to catch.
func vciOrderingSection(t *testing.T, path, tree string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ordering-clause guard: %s copy unreadable at %s: %v", tree, path, err)
	}
	lines := strings.Split(string(raw), "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, vciOrderingHeading) {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf(
			"VCI_ORDERING_CLAUSE_MISSING: %s copy of %s carries no %q heading — the ordering-attribution "+
				"clause was deleted or retitled. Restore it in BOTH trees (repo copy + internal/template/templates mirror).",
			tree, vciOrderingRelPath, vciOrderingHeading,
		)
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

// TestVCIOrderingClausePresence asserts the §2.3 clause survives in both trees
// with its per-tree expectations intact. It reports what it visited, not just
// what it found: the sweep is always exactly these two files.
func TestVCIOrderingClausePresence(t *testing.T) {
	t.Parallel()

	root := findProjectRootForMirrorTest(t)
	visited := 0

	local := vciOrderingSection(t, filepath.Join(root, vciOrderingRelPath), "repo-root")
	visited++
	if !strings.Contains(local, vciOrderingOperative) {
		t.Errorf("VCI_ORDERING_CLAUSE_MISSING: repo-root §2.3 lost its operative sentence %q — restore it", vciOrderingOperative)
	}
	if !strings.Contains(local, vciOrderingProvenance) {
		t.Errorf("VCI_ORDERING_CLAUSE_MISSING: repo-root §2.3 lost the motivating-instance provenance %q — restore it", vciOrderingProvenance)
	}

	mirror := vciOrderingSection(t, filepath.Join(root, "internal", "template", "templates", vciOrderingRelPath), "template-mirror")
	visited++
	if !strings.Contains(mirror, vciOrderingOperative) {
		t.Errorf(
			"VCI_ORDERING_CLAUSE_MISSING: template-mirror §2.3 lost its operative sentence %q — propagate the clause to %s (sanitized per §25)",
			vciOrderingOperative, filepath.Join("internal", "template", "templates", vciOrderingRelPath),
		)
	}
	for _, tok := range vciOrderingMirrorForbidden {
		if strings.Contains(mirror, tok) {
			t.Errorf("VCI_ORDERING_CLAUSE_NEUTRALITY: template-mirror §2.3 leaks internal token %q (§25 template neutrality) — strip it", tok)
		}
	}

	if visited != 2 {
		t.Fatalf("ordering-clause guard swept %d trees, expected 2 — a tree was silently skipped", visited)
	}
}
