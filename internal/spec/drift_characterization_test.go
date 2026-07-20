package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Characterization tests for DetectDrift (SPEC-SESSIONSTART-PERF-001 M1 —
// AC-SSP-003 / AC-SSP-005a / AC-SSP-005b / AC-SSP-020 / AC-SSP-021).
//
// These tests are the DDD PRESERVE safety net for the M1 single-pass refactor:
// they pin the CURRENT observable behavior of DetectDrift (per-SPEC records plus
// the drift count) BEFORE any algorithmic change, and must keep passing verbatim
// AFTER it. They are the gate for HARD REQ-SSP-005.
//
// The fixture deliberately exercises every classification category the refactor
// must preserve:
//
//	(1) drifted        — frontmatter disagrees with the git-implied status
//	(2) non-drifted    — frontmatter agrees with the git-implied status
//	(3) terminal       — terminal frontmatter is authoritative; sentinel "terminal-exempt"
//	(4) grandfather    — era-final SPEC is drift-exempt; sentinel "era-exempt"
//	(5) combined-scope — closed by a commit naming only the scope-prefix, reachable
//	                     ONLY through the secondary prefix fallback
//
// @MX:NOTE: [AUTO] characterization baseline — do not relax these assertions.
// @MX:REASON: SPEC-SESSIONSTART-PERF-001 REQ-SSP-005 [HARD] — M1 is a
//
//	behavior-preserving refactor, so these expectations ARE the contract. A failure
//	here means drift semantics changed, NOT that the fixture needs adjusting.

// fixtureSpec describes one SPEC directory to materialize inside a drift fixture.
type fixtureSpec struct {
	// id is the SPEC-ID, used as the directory name under .moai/specs/.
	id string
	// status is the frontmatter `status:` value.
	status string
	// era is the frontmatter `era:` value. A non-empty value pins era
	// classification through the H-override rule, keeping the fixture independent
	// of progress.md heuristics. Empty omits the field.
	era string
	// progress is the progress.md content. Empty omits the file.
	progress string
}

// chdirForTest switches the working directory for the duration of the test and
// restores it on cleanup. The drift walker queries git from the CURRENT working
// directory (not from baseDir), so a fixture must chdir into its own repo.
//
// Tests using this helper MUST NOT call t.Parallel(): os.Chdir mutates
// process-global state.
func chdirForTest(t *testing.T, dir string) {
	t.Helper()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() {
		if cerr := os.Chdir(orig); cerr != nil {
			t.Logf("restoring original working directory failed (ignored): %v", cerr)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %s failed: %v", dir, err)
	}
}

// writeFixtureSpecs materializes each fixtureSpec as .moai/specs/<id>/spec.md
// (plus progress.md when supplied) under root.
func writeFixtureSpecs(t *testing.T, root string, specs []fixtureSpec) {
	t.Helper()

	for _, s := range specs {
		dir := filepath.Join(root, ".moai", "specs", s.id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s failed: %v", dir, err)
		}
		specPath := filepath.Join(dir, "spec.md")
		if err := os.WriteFile(specPath, []byte(fixtureSpecMD(s)), 0o644); err != nil {
			t.Fatalf("write %s failed: %v", specPath, err)
		}
		if s.progress != "" {
			progressPath := filepath.Join(dir, "progress.md")
			if err := os.WriteFile(progressPath, []byte(s.progress), 0o644); err != nil {
				t.Fatalf("write %s failed: %v", progressPath, err)
			}
		}
	}
}

// setupDriftCorpusFixture builds a temporary project root holding BOTH a git
// repository (carrying the given commits, oldest→newest) and a .moai/specs
// corpus, then chdirs into it. Returns the project-root path for DetectDrift.
func setupDriftCorpusFixture(t *testing.T, specs []fixtureSpec, commits []fixtureCommit) string {
	t.Helper()

	root := setupGitFixture(t, commits)
	writeFixtureSpecs(t, root, specs)
	chdirForTest(t, root)

	return root
}

// fixtureSpecMD renders a minimal, schema-valid spec.md for a fixtureSpec.
func fixtureSpecMD(s fixtureSpec) string {
	var b strings.Builder

	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", s.id)
	fmt.Fprintf(&b, "title: %q\n", s.id+" characterization fixture")
	b.WriteString("version: \"0.1.0\"\n")
	fmt.Fprintf(&b, "status: %s\n", s.status)
	b.WriteString("created: 2026-07-11\n")
	b.WriteString("updated: 2026-07-11\n")
	b.WriteString("author: characterization-fixture\n")
	b.WriteString("priority: P1\n")
	b.WriteString("phase: \"v3.0.0\"\n")
	b.WriteString("module: \"internal/spec\"\n")
	b.WriteString("lifecycle: spec-anchored\n")
	b.WriteString("tags: \"fixture\"\n")
	if s.era != "" {
		fmt.Fprintf(&b, "era: %s\n", s.era)
	}
	b.WriteString("---\n\n# " + s.id + "\n")

	return b.String()
}

// assertDriftReport compares a DriftReport against the expected record set and
// drift count. Records are compared positionally: DetectDrift sorts by SPEC-ID,
// so want must be supplied in SPEC-ID order.
func assertDriftReport(t *testing.T, got *DriftReport, want []DriftRecord, wantCount int) {
	t.Helper()

	if got.Count != wantCount {
		t.Errorf("DriftReport.Count = %d, want %d", got.Count, wantCount)
	}
	if len(got.Records) != len(want) {
		t.Fatalf("len(DriftReport.Records) = %d, want %d\ngot: %+v", len(got.Records), len(want), got.Records)
	}
	for i, w := range want {
		if g := got.Records[i]; g != w {
			t.Errorf("Records[%d] = %+v, want %+v", i, g, w)
		}
	}
}

// TestDetectDrift_Characterization_FiveCategories pins the classification of all
// five categories the M1 refactor must preserve (AC-SSP-005a, AC-SSP-005b).
//
// It also pins the two record-emission sentinels the cheap pre-filter MUST keep
// emitting verbatim (AC-SSP-003): "terminal-exempt" and "era-exempt" — NOT empty
// strings, and NOT dropped records.
func TestDetectDrift_Characterization_FiveCategories(t *testing.T) {
	specs := []fixtureSpec{
		// (1) drifted — frontmatter in-progress, git implies implemented
		{id: "SPEC-DRIFTED-001", status: "in-progress", era: "V3R6"},
		// (2) non-drifted — frontmatter implemented, git implies implemented
		{id: "SPEC-ALIGNED-001", status: "implemented", era: "V3R6"},
		// (3) terminal — superseded frontmatter is authoritative (mechanism ③)
		{id: "SPEC-TERMINAL-001", status: "superseded", era: "V3R6"},
		// (4) grandfather — era-final (V3R5) SPECs are drift-exempt (mechanism ④)
		{id: "SPEC-GRAND-001", status: "in-progress", era: "V3R5"},
		// (5) combined-scope close — closed by a scope-prefix commit (mechanism ①)
		{id: "SPEC-CCSYNC-CLAUDEMD-001", status: "completed", era: "V3R6"},
	}

	// oldest → newest
	commits := []fixtureCommit{
		{title: "feat(SPEC-DRIFTED-001): M1 implementation"},
		{title: "feat(SPEC-ALIGNED-001): M1 implementation"},
		{title: "feat(SPEC-CCSYNC-CLAUDEMD-001): M1 implementation"},
		// Combined-scope close: names the scope-prefix SPEC-CCSYNC (NOT the full ID)
		// plus the distinguishing-segment group (CLAUDEMD + TOOLCAT). The per-SPEC
		// full-ID walk can never reach it; only the secondary prefix fallback can.
		{title: "chore(SPEC-CCSYNC): sync-phase artifacts + 3-phase close (CLAUDEMD + TOOLCAT)"},
	}

	root := setupDriftCorpusFixture(t, specs, commits)

	report, err := DetectDrift(root)
	if err != nil {
		t.Fatalf("DetectDrift returned unexpected error: %v", err)
	}

	// Sorted by SPEC-ID: ALIGNED < CCSYNC < DRIFTED < GRAND < TERMINAL
	want := []DriftRecord{
		{SPECID: "SPEC-ALIGNED-001", FrontmatterStatus: "implemented", GitImpliedStatus: "implemented", Drifted: false},
		{SPECID: "SPEC-CCSYNC-CLAUDEMD-001", FrontmatterStatus: "completed", GitImpliedStatus: "completed", Drifted: false},
		{SPECID: "SPEC-DRIFTED-001", FrontmatterStatus: "in-progress", GitImpliedStatus: "implemented", Drifted: true},
		{SPECID: "SPEC-GRAND-001", FrontmatterStatus: "in-progress", GitImpliedStatus: "era-exempt", Drifted: false},
		{SPECID: "SPEC-TERMINAL-001", FrontmatterStatus: "superseded", GitImpliedStatus: "terminal-exempt", Drifted: false},
	}

	assertDriftReport(t, report, want, 1)
}

// TestDetectDrift_Characterization_ChoreSkipAndWordBoundary pins the two hard-won
// stage-2 filters the single-pass rewrite must reproduce verbatim (AC-SSP-005b):
// chore-skip (LSCSK-001) and the SPEC-ID word-boundary filter (LSGF-001).
//
// Both filters are body-dependent: the candidate set comes from a FULL-message
// match (subject OR body), while the filters run against the SUBJECT only. This
// is precisely the two-stage divergence hazard the refactor must not flatten.
func TestDetectDrift_Characterization_ChoreSkipAndWordBoundary(t *testing.T) {
	specs := []fixtureSpec{
		// chore-skip: the newest matching commit is a metadata sweep that must be
		// skipped so the walker reaches the real feat commit beneath it.
		{id: "SPEC-SWEEP-001", status: "in-progress", era: "V3R6"},
		// word-boundary: SPEC-HARNESS-001 must NOT adopt the classification of a
		// commit that names only its NAMESPACE sibling.
		{id: "SPEC-HARNESS-001", status: "in-progress", era: "V3R6"},
	}

	// oldest → newest
	commits := []fixtureCommit{
		{title: "feat(SPEC-SWEEP-001): M1 implementation"},
		// Newest for SPEC-SWEEP-001: a chore(spec) sweep whose BODY names the SPEC.
		// The body match makes it a --grep candidate; the chore-skip filter must
		// drop it at stage 2 so the feat commit above wins.
		{title: "chore(spec): status drift sweep", body: "SPEC-SWEEP-001 in-progress -> implemented"},
		// Names only the NAMESPACE sibling. SPEC-HARNESS-001 is a SUBSTRING of
		// SPEC-HARNESS-NAMESPACE-001, so this commit IS a --grep candidate for
		// SPEC-HARNESS-001 — the word-boundary filter must reject it, leaving
		// SPEC-HARNESS-001 with no classifiable commit (its record is dropped).
		{title: "feat(SPEC-HARNESS-NAMESPACE-001): unrelated sibling work"},
	}

	root := setupDriftCorpusFixture(t, specs, commits)

	report, err := DetectDrift(root)
	if err != nil {
		t.Fatalf("DetectDrift returned unexpected error: %v", err)
	}

	// SPEC-HARNESS-001 emits NO record: its only --grep candidate is rejected by
	// the word-boundary filter, the walker exhausts, and DetectDrift skips it.
	want := []DriftRecord{
		{SPECID: "SPEC-SWEEP-001", FrontmatterStatus: "in-progress", GitImpliedStatus: "implemented", Drifted: true},
	}

	assertDriftReport(t, report, want, 1)
}

// TestDetectDrift_Characterization_EmptySpecsDir pins AC-SSP-021: an absent
// .moai/specs directory yields an empty report with count 0 and no error.
func TestDetectDrift_Characterization_EmptySpecsDir(t *testing.T) {
	root := t.TempDir()

	report, err := DetectDrift(root)
	if err != nil {
		t.Fatalf("DetectDrift returned unexpected error: %v", err)
	}
	if report.Count != 0 {
		t.Errorf("DriftReport.Count = %d, want 0", report.Count)
	}
	if len(report.Records) != 0 {
		t.Errorf("len(DriftReport.Records) = %d, want 0", len(report.Records))
	}
}

// TestDetectDrift_Characterization_NonGitEnvironment pins AC-SSP-020: in a
// checkout without git history, drift detection degrades gracefully — no panic,
// no error. SPECs needing git classification are skipped; the cheap pre-filter
// records (terminal / era-exempt) are still emitted.
func TestDetectDrift_Characterization_NonGitEnvironment(t *testing.T) {
	root := t.TempDir()

	specs := []fixtureSpec{
		// Pre-filtered: emitted without any git work.
		{id: "SPEC-TERMINAL-001", status: "archived", era: "V3R6"},
		// Requires git classification: skipped when git is unavailable.
		{id: "SPEC-ACTIVE-001", status: "in-progress", era: "V3R6"},
	}
	writeFixtureSpecs(t, root, specs)
	chdirForTest(t, root) // NOT a git repository

	report, err := DetectDrift(root)
	if err != nil {
		t.Fatalf("DetectDrift returned unexpected error in non-git environment: %v", err)
	}

	want := []DriftRecord{
		{SPECID: "SPEC-TERMINAL-001", FrontmatterStatus: "archived", GitImpliedStatus: "terminal-exempt", Drifted: false},
	}
	assertDriftReport(t, report, want, 0)
}
