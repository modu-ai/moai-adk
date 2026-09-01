package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// REQ-SR-009 regression locks (SPEC-STAMP-REACHABILITY-001 M3). These are
// characterization pins on the EXISTING freshness contract, added because a
// run-phase coverage sweep measured three gaps: reverted-churn-zero,
// described-roots diff scoping, and the one-below-threshold boundary
// (AgedLayerFails already pins exactly-at-threshold → stale). Each pin names
// the mutant it discriminates in its comment.

// TestCheckFreshness_RevertedChurnCountsZero pins the endpoint-diff property:
// churn that REVERTS to the stamped commit's content counts zero. A
// touch-based or reflog-based metric would count the two writes; only an
// endpoint diff (working-tree content vs content at the named commit) counts
// zero — the mutant this test discriminates.
func TestCheckFreshness_RevertedChurnCountsZero(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenance(t, root, head)

	target := filepath.Join(root, "internal", "alpha", "alpha.go")
	original, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, append([]byte("// churn\n"), original...), 0o644); err != nil {
		t.Fatal(err)
	}
	// Revert to the exact committed bytes — the endpoint state equals the
	// stamped commit again.
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	for _, l := range res.Layers {
		if l.Layer == LayerCodemaps {
			if l.Value != 0 || l.Verdict != VerdictFresh {
				t.Errorf("reverted churn: codemaps = (value %d, %q), want (0, fresh) — endpoint diff must ignore reverted writes (reason %s)", l.Value, l.Verdict, l.Reason)
			}
		}
	}
}

// TestCheckFreshness_DescribedRootsScopeFidelity pins that the checker
// re-diffes EXACTLY what the provenance claimed to describe: a stamp whose
// described_roots is ["internal"] ignores changes under cmd/ and pkg/, and
// counts a change under internal/ the moment one appears. The discriminated
// mutant hardcodes DefaultDescribedRoots instead of the stamp's roots.
func TestCheckFreshness_DescribedRootsScopeFidelity(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenanceBlock(t, root, &mx.Provenance{
		SchemaVersion:  mx.ProvenanceSchemaVersion,
		TreeRoot:       root,
		CommitSHA:      head,
		DescribedRoots: []string{"internal"},
		GeneratedBy:    "codemaps-gen",
	})

	// Changes OUTSIDE the described roots must not count.
	for _, p := range []string{"cmd/tool/main.go", "pkg/lib/lib.go"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(p)),
			[]byte("package changed // outside described roots\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	res, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness: %v", err)
	}
	for _, l := range res.Layers {
		if l.Layer == LayerCodemaps {
			if l.Value != 0 || l.Verdict != VerdictFresh {
				t.Errorf("outside-roots changes counted: codemaps = (value %d, %q), want (0, fresh) (reason %s)", l.Value, l.Verdict, l.Reason)
			}
		}
	}

	// One change INSIDE the described roots must count.
	if err := os.WriteFile(filepath.Join(root, "internal", "beta", "beta.go"),
		[]byte("package beta // inside described roots\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res2, err := CheckFreshness(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("CheckFreshness(2): %v", err)
	}
	counted := false
	for _, l := range res2.Layers {
		if l.Layer == LayerCodemaps && l.Value == 1 {
			counted = true
		}
	}
	if !counted {
		t.Errorf("inside-roots change not counted: codemaps row = %+v", res2.Layers)
	}
}

// TestCheckFreshness_ThresholdBoundaryOneBelowFresh pins the comparison
// direction at check.go:213 (`count >= th.CodemapsChangedFiles` → stale):
// one-below the threshold stays FRESH and exactly-at goes STALE. The
// discriminated mutants are `>` (would keep one-below fresh but ALSO keep
// exactly-at fresh) and `<=` (inverted).
func TestCheckFreshness_ThresholdBoundaryOneBelowFresh(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenance(t, root, head)

	th := DefaultThresholds()
	th.CodemapsChangedFiles = 3

	verdict := func() (int, string) {
		t.Helper()
		res, err := CheckFreshness(root, th)
		if err != nil {
			t.Fatalf("CheckFreshness: %v", err)
		}
		for _, l := range res.Layers {
			if l.Layer == LayerCodemaps {
				return l.Value, string(l.Verdict)
			}
		}
		t.Fatalf("no codemaps row: %+v", res.Layers)
		return 0, ""
	}

	// One-below: 2 changed described-source files (< 3) → fresh.
	for _, p := range []string{"internal/alpha/alpha.go", "internal/beta/beta.go"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(p)),
			[]byte("package changed\n// one below\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if v, vr := verdict(); v != 2 || vr != string(VerdictFresh) {
		t.Errorf("one-below boundary: codemaps = (value %d, %q), want (2, fresh)", v, vr)
	}

	// Exactly-at: a third file tips count to 3 (>= 3) → stale.
	if err := os.WriteFile(filepath.Join(root, "cmd", "tool", "main.go"),
		[]byte("package changed\n// at threshold\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if v, vr := verdict(); v != 3 || vr != string(VerdictStale) {
		t.Errorf("at-threshold boundary: codemaps = (value %d, %q), want (3, stale)", v, vr)
	}
}
