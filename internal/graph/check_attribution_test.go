package graph

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// stampAtHEAD writes a clean codemaps provenance stamped at the fixture's
// current HEAD and returns that sha.
func stampAtHEAD(t *testing.T, root string) string {
	t.Helper()
	sha := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenance(t, root, sha)
	return sha
}

// AC-GFC-008 (REQ-GFC-007) — the codemaps report carries the change's own
// described-worthy contribution alongside the cumulative count, and a checkout
// with no first parent reports that contribution ABSENT rather than 0.
//
// Mutant this kills: an implementation that defaults a missing first parent to
// 0. Under it the no-first-parent sub-test reads a present contribution of 0 —
// indistinguishable from a measured zero, which is exactly the inheriting
// signature this SPEC exists to make legible.
func TestCheckCodemaps_Contribution(t *testing.T) {
	t.Run("inheriting merge contributes nothing while the cumulative is red", func(t *testing.T) {
		root := newCheckFixture(t)
		stamp := stampAtHEAD(t, root)
		base := gitFix(t, root, "rev-parse", "--abbrev-ref", "HEAD")

		// Five described-worthy files land on the mainline AFTER the stamp —
		// these are the cumulative the merge will inherit.
		for i := 1; i <= 5; i++ {
			writeFixtureFile(t, root, fmt.Sprintf("internal/gamma/g%d.go", i), "package gamma\n")
		}
		gitFix(t, root, "add", "-A")
		gitFix(t, root, "commit", "-q", "-m", "five described-worthy files")

		// A side branch carrying NOTHING described-worthy, merged with a real
		// merge commit. Its own contribution against its first parent is 0.
		gitFix(t, root, "checkout", "-q", "-b", "side")
		writeFixtureFile(t, root, "internal/astgrep/testdata/rule-tests/a.yml", "id: a\n")
		writeFixtureFile(t, root, "internal/gamma/g1_test.go", "package gamma\n")
		gitFix(t, root, "add", "-A")
		gitFix(t, root, "commit", "-q", "-m", "testdata and tests only")
		gitFix(t, root, "checkout", "-q", base)
		gitFix(t, root, "merge", "-q", "--no-ff", "-m", "merge side", "side")

		if parents := strings.Fields(gitFix(t, root, "rev-list", "--parents", "-n1", "HEAD")); len(parents) != 3 {
			t.Fatalf("fixture is not a merge commit: rev-list --parents returned %d fields, want 3", len(parents))
		}

		rep, err := checkCodemaps(root, Thresholds{CodemapsChangedFiles: 3})
		if err != nil {
			t.Fatalf("checkCodemaps: %v", err)
		}
		if rep.Verdict != VerdictStale {
			t.Fatalf("verdict = %q, want %q (cumulative %d vs threshold 3)", rep.Verdict, VerdictStale, rep.Value)
		}
		if rep.Value != 5 {
			t.Fatalf("cumulative value = %d, want 5 (stamp %s)", rep.Value, shortHash(stamp))
		}
		if rep.Contribution == nil {
			t.Fatalf("contribution is absent, want a measured 0 — the first parent resolves on a merge commit")
		}
		if *rep.Contribution != 0 {
			t.Fatalf("contribution = %d, want 0 — the merge introduced no described-worthy file", *rep.Contribution)
		}
		if rep.ContributionAbsentReason != "" {
			t.Fatalf("contribution absent-reason = %q on a measured contribution, want empty", rep.ContributionAbsentReason)
		}
	})

	t.Run("no first parent reports the contribution absent, never 0", func(t *testing.T) {
		root := newCheckFixture(t)
		// newCheckFixture leaves exactly one commit, so HEAD is the root
		// commit and HEAD^1 does not resolve.
		if out, err := gitOutput(root, "rev-parse", "--verify", "HEAD^1"); err == nil {
			t.Fatalf("fixture HEAD has a first parent (%s) — the no-first-parent case is not exercised", strings.TrimSpace(out))
		}
		stampAtHEAD(t, root)
		writeFixtureFile(t, root, "internal/delta/d.go", "package delta\n")

		rep, err := checkCodemaps(root, Thresholds{CodemapsChangedFiles: 40})
		if err != nil {
			t.Fatalf("checkCodemaps: %v", err)
		}
		if rep.Value != 1 {
			t.Fatalf("cumulative value = %d, want 1 — the layer must still be measured", rep.Value)
		}
		if rep.Contribution != nil {
			t.Fatalf("contribution = %d (present), want absent — HEAD has no first parent, so a present 0 is fabricated", *rep.Contribution)
		}
		if rep.ContributionAbsentReason == "" {
			t.Fatalf("contribution is absent but carries no reason — absent must be legible, not merely a missing field")
		}
	})
}

// AC-GFC-009 / AC-GFC-010 (REQ-GFC-008, REQ-GFC-010) — a stale codemaps
// verdict names the described-worthy paths driving the count, bounded to a
// readable maximum with an explicit overflow count.
//
// Mutant this kills: a report that carries the count alone. Both sub-tests
// fail on an empty driving-path list.
func TestCheckCodemaps_DrivingPaths(t *testing.T) {
	t.Run("over the display bound, the listing truncates and declares the overflow", func(t *testing.T) {
		root := newCheckFixture(t)
		stampAtHEAD(t, root)

		const total = drivingPathDisplayBound + 7
		for i := 1; i <= total; i++ {
			writeFixtureFile(t, root, fmt.Sprintf("internal/eps/e%02d.go", i), "package eps\n")
		}
		// Noise the predicate must keep out of the listing.
		writeFixtureFile(t, root, "internal/eps/e01_test.go", "package eps\n")
		writeFixtureFile(t, root, "internal/eps/testdata/x.go", "package testdata\n")

		rep, err := checkCodemaps(root, Thresholds{CodemapsChangedFiles: 3})
		if err != nil {
			t.Fatalf("checkCodemaps: %v", err)
		}
		if rep.Verdict != VerdictStale {
			t.Fatalf("verdict = %q, want %q", rep.Verdict, VerdictStale)
		}
		if rep.Value != total {
			t.Fatalf("cumulative value = %d, want %d — the predicate must exclude the _test.go and testdata noise", rep.Value, total)
		}
		if len(rep.DrivingPaths) != drivingPathDisplayBound {
			t.Fatalf("driving paths listed = %d, want %d (0 = the report carries the count alone)", len(rep.DrivingPaths), drivingPathDisplayBound)
		}
		if rep.DrivingPathsOmitted != total-drivingPathDisplayBound {
			t.Fatalf("driving paths omitted = %d, want %d — the overflow must be explicit", rep.DrivingPathsOmitted, total-drivingPathDisplayBound)
		}
		if !sort.StringsAreSorted(rep.DrivingPaths) {
			t.Fatalf("driving paths are not sorted: %v — an unstable listing cannot be diffed between runs", rep.DrivingPaths)
		}
		for _, p := range rep.DrivingPaths {
			if strings.HasSuffix(p, "_test.go") || strings.Contains(p, "/testdata/") {
				t.Fatalf("driving path %q is not described-worthy — the listing must apply the same predicate as the count", p)
			}
		}
	})

	t.Run("under the display bound, every driving path is listed and nothing is omitted", func(t *testing.T) {
		root := newCheckFixture(t)
		stampAtHEAD(t, root)

		for i := 1; i <= 4; i++ {
			writeFixtureFile(t, root, fmt.Sprintf("internal/zeta/z%d.go", i), "package zeta\n")
		}

		rep, err := checkCodemaps(root, Thresholds{CodemapsChangedFiles: 3})
		if err != nil {
			t.Fatalf("checkCodemaps: %v", err)
		}
		if rep.Verdict != VerdictStale {
			t.Fatalf("verdict = %q, want %q", rep.Verdict, VerdictStale)
		}
		if len(rep.DrivingPaths) != 4 {
			t.Fatalf("driving paths listed = %d, want 4 (0 = the report carries the count alone)", len(rep.DrivingPaths))
		}
		if rep.DrivingPathsOmitted != 0 {
			t.Fatalf("driving paths omitted = %d, want 0 — nothing overflowed", rep.DrivingPathsOmitted)
		}
	})
}
