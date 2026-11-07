package graph

import (
	"fmt"
	"testing"
)

// Card t2001 — the base branch and the pull request can touch the SAME
// described-worthy files. The cumulative count (Value) is the union of both
// change sets measured from the codemaps anchor; the change's own contribution
// (HEAD^1..HEAD) is a subset of it, so Value - Contribution is not the base's
// own count when the two overlap. The base alone is already over the line here,
// so the layer must classify inherited (warn on a pull_request).
func TestClassifyVerdicts_BaseAndChangeOverlapIsInherited(t *testing.T) {
	const threshold = 5
	root := newCheckFixture(t)
	stampAtHEAD(t, root)

	// Base: six described-worthy files land after the stamp (base alone >= 5).
	for i := 1; i <= 6; i++ {
		writeFixtureFile(t, root, fmt.Sprintf("internal/gamma/g%d.go", i), "package gamma\n")
	}
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "base: six described-worthy files")

	// PR head: re-edits two of those same files, leaving stamp and body alone.
	for i := 1; i <= 2; i++ {
		writeFixtureFile(t, root, fmt.Sprintf("internal/gamma/g%d.go", i), "package gamma\n\nvar _ = 1\n")
	}
	gitFix(t, root, "add", "-A")
	gitFix(t, root, "commit", "-q", "-m", "pr: re-edit two of the base files")

	th := Thresholds{CodemapsChangedFiles: threshold}
	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("checkCodemaps: %v", err)
	}
	if rep.Contribution == nil || rep.CodemapsInputsStable == nil {
		t.Fatalf("contribution/stability unmeasured: %+v", rep)
	}
	t.Logf("Value=%d Contribution=%d threshold=%d stable=%v", rep.Value, *rep.Contribution, threshold, *rep.CodemapsInputsStable)
	if rep.Verdict != VerdictStale {
		t.Fatalf("verdict = %q, want stale", rep.Verdict)
	}

	res := CheckResult{Layers: []LayerReport{rep}}
	res.ClassifyVerdicts(EventPullRequest, InheritedDriftWarn, th)
	got := res.Layers[0]
	if got.DriftOrigin != DriftInherited || got.VerdictClass != VerdictClassWarn {
		t.Fatalf("origin=%q class=%q, want %q/%q (Value=%d Contribution=%d threshold=%d)",
			got.DriftOrigin, got.VerdictClass, DriftInherited, VerdictClassWarn, got.Value, *got.Contribution, threshold)
	}
}
