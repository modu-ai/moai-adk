package spec

import (
	"fmt"
	"strings"
	"testing"
)

// Seam tests for the O(1)-subprocess guarantee and the HEAD-SHA cache
// (SPEC-SESSIONSTART-PERF-001 M1 — AC-SSP-001, AC-SSP-002, AC-SSP-003,
// AC-SSP-004, AC-SSP-006, AC-SSP-006a, AC-SSP-022).
//
// These drive detectDrift through the injectable driftDeps seam so the git-log
// invocation COUNT is directly observable. That count is the ONLY mechanical proof
// that drift detection no longer scales its subprocess spawns with the SPEC count —
// wall-clock alone would be corroborating evidence, not a verifiable assertion.

// fakeGit is a counting stand-in for the three git operations detectDrift needs.
// logCalls counts ONLY the single-pass git log, which is the expensive operation
// the refactor exists to bound.
type fakeGit struct {
	logCalls int
	head     string
	commits  []commitRecord
}

// deps builds a driftDeps backed by this fake.
func (f *fakeGit) deps(useCache bool) driftDeps {
	return driftDeps{
		useCache: useCache,
		headSHA:  func() (string, error) { return f.head, nil },
		branch:   func() string { return "main" },
		logAll: func(string) ([]commitRecord, error) {
			f.logCalls++
			return f.commits, nil
		},
	}
}

// seamSpecID is the SPEC-ID scheme used by the scaling fixtures.
func seamSpecID(i int) string {
	return fmt.Sprintf("SPEC-SEAM%03d-001", i)
}

// writeActiveSpecCorpus materializes n SPECs that all survive the pre-filter (so
// they all require git classification) and returns a matching newest-first commit
// index in which each SPEC has exactly one classifiable commit.
func writeActiveSpecCorpus(t *testing.T, root string, n int) []commitRecord {
	t.Helper()

	specs := make([]fixtureSpec, 0, n)
	commits := make([]commitRecord, 0, n)

	for i := 1; i <= n; i++ {
		id := seamSpecID(i)
		specs = append(specs, fixtureSpec{id: id, status: "in-progress", era: "V3R6"})

		subject := fmt.Sprintf("feat(%s): M1 implementation", id)
		commits = append(commits, commitRecord{subject: subject, fullMsg: subject})
	}
	writeFixtureSpecs(t, root, specs)

	return commits
}

// TestDetectDrift_ConstantGitLogInvocations is the AC-SSP-001 proof: the git-log
// subprocess count is a small constant, independent of the SPEC count N.
//
// The pre-refactor implementation spawned 1-2 `git log --grep` subprocesses PER
// active SPEC, so this counter would have read N (or ~2N) instead of 1.
func TestDetectDrift_ConstantGitLogInvocations(t *testing.T) {
	for _, n := range []int{1, 10, 100} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			root := t.TempDir()
			commits := writeActiveSpecCorpus(t, root, n)

			fg := &fakeGit{head: "head-sha-constant", commits: commits}

			report, err := detectDrift(root, fg.deps(true))
			if err != nil {
				t.Fatalf("detectDrift: %v", err)
			}

			if fg.logCalls != 1 {
				t.Errorf("git log invocations = %d, want exactly 1 (must NOT grow with N=%d)", fg.logCalls, n)
			}
			if len(report.Records) != n {
				t.Errorf("len(Records) = %d, want %d", len(report.Records), n)
			}
			// Every fixture SPEC is frontmatter=in-progress vs git=implemented.
			if report.Count != n {
				t.Errorf("Count = %d, want %d (all fixtures drift)", report.Count, n)
			}
		})
	}
}

// TestDetectDrift_PreFilterPerformsNoGitWork is the AC-SSP-003 proof: terminal and
// grandfather-era SPECs are excluded from the git-checked working set BEFORE any git
// work happens. With a corpus of only such SPECs, the git log pass never runs at all.
//
// It also pins the record-emission contract: pre-filtered SPECs still emit a record,
// carrying the ACTUAL sentinel ("terminal-exempt" / "era-exempt"), never an empty
// string and never a dropped record.
func TestDetectDrift_PreFilterPerformsNoGitWork(t *testing.T) {
	root := t.TempDir()

	writeFixtureSpecs(t, root, []fixtureSpec{
		{id: "SPEC-TERM-001", status: "superseded", era: "V3R6"},
		{id: "SPEC-ARCH-001", status: "archived", era: "V3R6"},
		{id: "SPEC-OLD-001", status: "in-progress", era: "V3R5"}, // era-final → grandfathered
	})

	fg := &fakeGit{head: "head-sha-prefilter"}

	report, err := detectDrift(root, fg.deps(false))
	if err != nil {
		t.Fatalf("detectDrift: %v", err)
	}

	if fg.logCalls != 0 {
		t.Errorf("git log invocations = %d, want 0 (a fully pre-filtered corpus needs no git work)", fg.logCalls)
	}

	// Sorted by SPEC-ID: ARCH < OLD < TERM
	want := []DriftRecord{
		{SPECID: "SPEC-ARCH-001", FrontmatterStatus: "archived", GitImpliedStatus: "terminal-exempt", Drifted: false},
		{SPECID: "SPEC-OLD-001", FrontmatterStatus: "in-progress", GitImpliedStatus: "era-exempt", Drifted: false},
		{SPECID: "SPEC-TERM-001", FrontmatterStatus: "superseded", GitImpliedStatus: "terminal-exempt", Drifted: false},
	}
	assertDriftReport(t, report, want, 0)
}

// TestDetectDrift_CacheHitPerformsZeroGitLogWork is the AC-SSP-004 proof: while HEAD
// is unchanged, a second run is served entirely from cache — zero git-log work — and
// returns a result identical to the computed one.
func TestDetectDrift_CacheHitPerformsZeroGitLogWork(t *testing.T) {
	root := t.TempDir()
	commits := writeActiveSpecCorpus(t, root, 5)

	fg := &fakeGit{head: "head-sha-A", commits: commits}

	// Cold: computes and persists the cache (AC-SSP-006).
	cold, err := detectDrift(root, fg.deps(true))
	if err != nil {
		t.Fatalf("cold detectDrift: %v", err)
	}
	if fg.logCalls != 1 {
		t.Fatalf("cold run git log invocations = %d, want 1", fg.logCalls)
	}

	// Warm: same HEAD → cache hit.
	fg.logCalls = 0
	warm, err := detectDrift(root, fg.deps(true))
	if err != nil {
		t.Fatalf("warm detectDrift: %v", err)
	}

	if fg.logCalls != 0 {
		t.Errorf("cache hit performed %d git log invocations, want 0", fg.logCalls)
	}
	if warm.Count != cold.Count {
		t.Errorf("cached Count = %d, want %d", warm.Count, cold.Count)
	}
	if len(warm.Records) != len(cold.Records) {
		t.Fatalf("cached len(Records) = %d, want %d", len(warm.Records), len(cold.Records))
	}
	for i := range cold.Records {
		if warm.Records[i] != cold.Records[i] {
			t.Errorf("cached Records[%d] = %+v, want %+v", i, warm.Records[i], cold.Records[i])
		}
	}
}

// TestDetectDrift_CacheInvalidatedWhenHeadAdvances is the AC-SSP-022 proof: a cache
// entry keyed on a stale HEAD is never served — the result is recomputed against the
// current HEAD.
func TestDetectDrift_CacheInvalidatedWhenHeadAdvances(t *testing.T) {
	root := t.TempDir()
	commits := writeActiveSpecCorpus(t, root, 3)

	fg := &fakeGit{head: "head-sha-A", commits: commits}

	if _, err := detectDrift(root, fg.deps(true)); err != nil {
		t.Fatalf("priming run: %v", err)
	}

	// A new commit lands: HEAD advances, so the cached entry is stale.
	fg.logCalls = 0
	fg.head = "head-sha-B"

	if _, err := detectDrift(root, fg.deps(true)); err != nil {
		t.Fatalf("post-advance run: %v", err)
	}

	if fg.logCalls != 1 {
		t.Errorf("git log invocations after HEAD advanced = %d, want 1 (stale cache must NOT be served)", fg.logCalls)
	}
}

// TestDetectDrift_NoCacheBypassesValidCacheEntry is the AC-SSP-006a proof: --no-cache
// recomputes even when a valid entry exists for the CURRENT HEAD. This is what makes
// the HEAD-only cache key safe — it is the operator's escape from the
// uncommitted-frontmatter stale window.
func TestDetectDrift_NoCacheBypassesValidCacheEntry(t *testing.T) {
	root := t.TempDir()
	commits := writeActiveSpecCorpus(t, root, 3)

	fg := &fakeGit{head: "head-sha-A", commits: commits}

	if _, err := detectDrift(root, fg.deps(true)); err != nil {
		t.Fatalf("priming run: %v", err)
	}

	// Same HEAD — the cache entry is valid — but --no-cache must ignore it.
	fg.logCalls = 0

	if _, err := detectDrift(root, fg.deps(false)); err != nil {
		t.Fatalf("--no-cache run: %v", err)
	}

	if fg.logCalls != 1 {
		t.Errorf("--no-cache git log invocations = %d, want 1 (a valid same-HEAD cache must be bypassed)", fg.logCalls)
	}
}

// TestInMemImpliedStatus_BodyOnlyMatchesExhaustWindow guards the single most dangerous
// divergence in the refactor (design.md §M1.3 / AP-6).
//
// `git log --grep=<specID> -50` applies its -50 window to the FULL-MESSAGE-matched set.
// So gitLogWindowSize commits that mention a SPEC-ID only in their BODY — while their
// SUBJECT names something else — completely exhaust the walker's window, and a deeper
// subject-close commit is NEVER reached.
//
// An index keyed on SUBJECT tokens alone would skip those body-only commits entirely,
// reach the close commit, and return "implemented" — flipping the drift count. This
// test fails loudly if anyone ever makes that "simplification".
func TestInMemImpliedStatus_BodyOnlyMatchesExhaustWindow(t *testing.T) {
	specID := "SPEC-WINDOW-001"

	var commits []commitRecord

	// Newest-first: exactly gitLogWindowSize commits whose BODY references specID but
	// whose SUBJECT names an unrelated SPEC. Each is a stage-1 candidate (consuming
	// window) and a stage-2 rejection (word-boundary filter).
	for i := 0; i < gitLogWindowSize; i++ {
		subject := fmt.Sprintf("chore(SPEC-OTHER-%03d): unrelated sweep", i+1)
		commits = append(commits, commitRecord{
			subject: subject,
			fullMsg: subject + "\n\nrelated: " + specID,
		})
	}

	// One commit deeper than the window: the SPEC's real, classifiable commit.
	closeSubject := fmt.Sprintf("feat(%s): M1 implementation", specID)
	commits = append(commits, commitRecord{subject: closeSubject, fullMsg: closeSubject})

	status, err := inMemImpliedStatus(commits, specID)
	if err == nil {
		t.Fatalf("inMemImpliedStatus = (%q, nil), want a window-exhaustion error.\n"+
			"Reaching the close commit means the candidate window was applied to SUBJECT matches "+
			"instead of FULL-MESSAGE matches — the exact semantic drift that flips the drift count.", status)
	}
}

// TestInMemImpliedStatus_WindowBoundaryReachesCloseCommit is the complement: with one
// FEWER body-only match, the close commit lands as the last in-window candidate and IS
// classified. Together the two tests pin the window boundary exactly where git puts it.
func TestInMemImpliedStatus_WindowBoundaryReachesCloseCommit(t *testing.T) {
	specID := "SPEC-WINDOW-002"

	var commits []commitRecord

	for i := 0; i < gitLogWindowSize-1; i++ {
		subject := fmt.Sprintf("chore(SPEC-OTHER-%03d): unrelated sweep", i+1)
		commits = append(commits, commitRecord{
			subject: subject,
			fullMsg: subject + "\n\nrelated: " + specID,
		})
	}

	closeSubject := fmt.Sprintf("feat(%s): M1 implementation", specID)
	commits = append(commits, commitRecord{subject: closeSubject, fullMsg: closeSubject})

	status, err := inMemImpliedStatus(commits, specID)
	if err != nil {
		t.Fatalf("inMemImpliedStatus returned unexpected error: %v", err)
	}
	if status != "implemented" {
		t.Errorf("inMemImpliedStatus = %q, want %q (the close commit is the last in-window candidate)", status, "implemented")
	}
}

// TestInMemImpliedStatus_NoMatchReportsNoHistory pins the two distinct exhaustion
// errors the original walker produced, since DetectDrift skips a SPEC on either.
func TestInMemImpliedStatus_NoMatchReportsNoHistory(t *testing.T) {
	commits := []commitRecord{
		{subject: "feat(SPEC-OTHER-001): unrelated", fullMsg: "feat(SPEC-OTHER-001): unrelated"},
	}

	if _, err := inMemImpliedStatus(commits, "SPEC-ABSENT-001"); err == nil {
		t.Error("inMemImpliedStatus on a SPEC with no matching commit = nil error, want error")
	}
}

// TestParseCommitRecords verifies the single-pass git log output parser: subject and
// full raw message are split on the unit separator, records on the record separator,
// and multi-line bodies survive intact (the reason a newline-delimited format cannot
// be used here).
func TestParseCommitRecords(t *testing.T) {
	output := "feat(SPEC-A-001): first" + gitLogFieldSep +
		"feat(SPEC-A-001): first\n\nbody line 1\nbody line 2\n" + gitLogRecordSep +
		"\nchore(SPEC-B-001): second" + gitLogFieldSep +
		"chore(SPEC-B-001): second\n" + gitLogRecordSep + "\n"

	records := parseCommitRecords(output)

	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2\ngot: %+v", len(records), records)
	}
	if records[0].subject != "feat(SPEC-A-001): first" {
		t.Errorf("records[0].subject = %q", records[0].subject)
	}
	if !strings.Contains(records[0].fullMsg, "body line 2") {
		t.Errorf("records[0].fullMsg lost its multi-line body: %q", records[0].fullMsg)
	}
	if records[1].subject != "chore(SPEC-B-001): second" {
		t.Errorf("records[1].subject = %q", records[1].subject)
	}
}
