package spec

import (
	"fmt"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Performance-budget regression guard (SPEC-SESSIONSTART-PERF-001 M3 —
// REQ-SSP-014 / AC-SSP-014).
//
// The session-start block this SPEC removed was O(n) `git log` subprocess
// spawns: ~2×N git processes for N SPECs. The single-pass design bounds that to
// exactly ONE subprocess. This file is the standing guard that the O(n) pattern
// cannot silently return.
//
// Two complementary assertions, per the design's "primary deterministic +
// secondary timed" contract:
//
//   - PRIMARY (deterministic): drive detectDrift through the counting driftDeps
//     seam at N=DefaultDriftPerfFixtureSpecs and assert the git-log invocation
//     count is EXACTLY 1, independent of N. This is the mechanical proof — a
//     per-SPEC subprocess regression would make this count ~N. It cannot flake:
//     it counts calls, not wall-clock.
//   - SECONDARY (real timed run): run the REAL DetectDriftFresh path against a
//     real (tiny) git repo carrying N SPEC dirs, and assert it completes under
//     the configured budget. This is a real timed measurement of the actual
//     algorithm — not a mocked claim — with a generous margin (the in-memory
//     walk of N SPECs against a small history is milliseconds vs a 2s budget).
//
// Guard-catches-regression proof: reintroducing a per-SPEC `deps.logAll` call
// inside detectDrift's active-SPEC loop makes the PRIMARY assertion read
// logCalls == N and the test FAILS. That is the regression this guard exists to
// catch.

// TestDetectDrift_PerfBudget_ConstantSubprocessAtScale is the PRIMARY guard: the
// git-log subprocess count stays 1 at N=500. Deterministic — it counts calls.
func TestDetectDrift_PerfBudget_ConstantSubprocessAtScale(t *testing.T) {
	n := config.DefaultDriftPerfFixtureSpecs // 500

	root := t.TempDir()
	commits := writeActiveSpecCorpus(t, root, n)
	fg := &fakeGit{head: "head-perf", commits: commits}

	start := time.Now()
	report, err := detectDrift(root, fg.deps(false)) // --no-cache: force the full compute path
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("detectDrift at N=%d: %v", n, err)
	}

	// PRIMARY: exactly one git-log subprocess, whatever N is.
	if fg.logCalls != 1 {
		t.Errorf("AC-SSP-014: git log invocations = %d at N=%d, want EXACTLY 1 "+
			"(a per-SPEC subprocess regression would make this ~N — the O(n) pattern this SPEC removed)",
			fg.logCalls, n)
	}
	if len(report.Records) != n {
		t.Errorf("len(Records) = %d, want %d (every active SPEC is classified)", len(report.Records), n)
	}

	// SECONDARY (in-memory walk timing under the seam): even with git mocked, the
	// per-SPEC in-memory walk is the part that would go O(n²) if the index were
	// mis-built. A generous budget catches such a regression while never flaking.
	if elapsed > config.DefaultDriftPerfBudget {
		t.Errorf("AC-SSP-014: seam-driven detection of N=%d SPECs took %v, want < %v (perf budget)",
			n, elapsed, config.DefaultDriftPerfBudget)
	}
}

// TestDetectDrift_PerfBudget_RealRepoWithinBudget is the SECONDARY guard: a REAL
// timed run of the production DetectDriftFresh path against a real git repo
// holding N SPEC dirs, asserting completion under the budget. This exercises the
// actual `git log` subprocess + parser + in-memory walk end-to-end — not a stub.
//
// Uses os.Chdir (via setupDriftCorpusFixture), so it MUST NOT run in parallel.
func TestDetectDrift_PerfBudget_RealRepoWithinBudget(t *testing.T) {
	n := config.DefaultDriftPerfFixtureSpecs // 500

	specs := make([]fixtureSpec, 0, n)
	for i := 1; i <= n; i++ {
		id := fmt.Sprintf("SPEC-PERF%04d-001", i)
		specs = append(specs, fixtureSpec{id: id, status: "in-progress", era: "V3R6"})
	}

	// A small real history: a handful of commits, one of which classifies a few
	// SPECs. The point is that the real `git log` pass runs once against a real
	// repo — not that every SPEC has a matching commit.
	commits := []fixtureCommit{
		{title: "chore: bootstrap"},
		{title: "feat(SPEC-PERF0001-001): M1 implementation"},
		{title: "feat(SPEC-PERF0002-001): M1 implementation"},
	}

	root := setupDriftCorpusFixture(t, specs, commits) // inits git repo + writes specs + chdir

	start := time.Now()
	report, err := DetectDriftFresh(root) // real path, cold cache
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("DetectDriftFresh on a real N=%d fixture: %v", n, err)
	}

	// All N active SPECs go through the in-memory walk (each scans the shared
	// commit index), so the O(n)-surface cost is incurred for the full N even
	// though only the committed SPECs produce a record — a SPEC with no matching
	// git history is correctly DROPPED from the report (pre-existing behavior).
	// The two feat(SPEC-PERF...) commits classify their SPECs as "implemented"
	// vs frontmatter "in-progress" → 2 drifted records.
	const wantCommittedDrift = 2
	if report.Count != wantCommittedDrift {
		t.Errorf("report.Count = %d, want %d (the two committed SPECs drift; uncommitted SPECs are dropped)",
			report.Count, wantCommittedDrift)
	}

	// The load-bearing assertion: the REAL end-to-end path (git log subprocess +
	// parser + N in-memory walks) completes under the budget. A blow-out at this
	// scale signals the O(n)-subprocess pattern has regressed.
	if elapsed > config.DefaultDriftPerfBudget {
		t.Errorf("AC-SSP-014: real-repo detection of N=%d SPECs took %v, want < %v (perf budget)",
			n, elapsed, config.DefaultDriftPerfBudget)
	}
}
