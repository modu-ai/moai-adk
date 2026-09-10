package spec_test

// lint_syncsha_test.go — SyncSHASlotFormatRule acceptance suite
// (SPEC-SYNC-SHA-SLOT-FORMAT-001 M3, card t299).
//
// THE REGRESSION TRIAD. AC-SSF-001, AC-SSF-002, and AC-SSF-003 are ONE
// instrument in three parts, and they are reported together or not at all:
//
//	(a) fires on prose               — alone, satisfied by a rule that flags everything
//	(b) silent on a well-formed SHA  — alone, satisfied by a rule that flags nothing
//	(c) silent on the sanctioned
//	    mid-backfill state           — alone, also satisfied by a rule that flags nothing
//
// (a) with (b) still permits a rule that has quietly closed the D3 backfill
// window. (b) with (c) still permits a switched-off rule. Only all three
// together separate a working guard from every degenerate one, which is why
// partial reporting of this triad is treated as no report.
//
// Every test below names the MUTATION that must turn it red. Each was planted,
// observed red, and reverted; the verbatim failing output is recorded in
// progress.md §E.2.
//
// FIXTURE ERA PRECONDITION (acceptance.md §A, [HARD]). Fixtures under
// testdata/syncsha/<case>/ carry `status: in-progress` (non-terminal) and a
// progress.md satisfying era heuristic H-4 (§E.2 + §E.4 + non-empty
// sync_commit_sha), so nothing is demoted. lint.go:220 marks every warning on a
// grandfathered or terminal-status document advisory and --strict escalates only
// non-advisory warnings, so a mis-built fixture would take AC-SSF-006's --strict
// half down for a reason having nothing to do with the rule under test. The
// prose fixture uses `TBD (filled post-commit)` specifically: `cleanFieldValue`
// maps a bare `pending` / `<pending>` / `null` / `none` / `tbd` to the empty
// string, and a fixture using one of those would classify V3R5 under H-3 and be
// demoted. TestSyncSHASlot_FlagsProse asserts Advisory == false, which is what
// makes that precondition a measurement rather than a comment.
//
// Fixtures are linted ONE AT A TIME (an explicit spec.md path per call), so they
// deliberately share a SPEC id; DuplicateSPECIDRule never sees two at once.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// syncSHAFindings lints one fixture case and returns only its
// SyncSHASlotFormat findings.
func syncSHAFindings(t *testing.T, c string) []spec.Finding {
	t.Helper()
	dir := filepath.Join(testdataDir, "syncsha", c)
	// Guard against the silent-empty-sweep hazard (card t397): Lint on a
	// missing fixture file returns no findings and no error, so a deleted or
	// mistyped fixture directory would read as a pass rather than as broken
	// test infrastructure. A fixture that does not exist is a failure here,
	// before the lint can produce its vacuous zero.
	fixtureMD := filepath.Join(dir, "spec.md")
	if st, statErr := os.Stat(fixtureMD); statErr != nil || st.IsDir() {
		t.Fatalf("fixture %q has no spec.md (%v) — the zero-findings result below would be a vacuous pass over missing infrastructure, not evidence", c, statErr)
	}
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      dir,
	})
	report, err := linter.Lint([]string{fixtureMD})
	if err != nil {
		t.Fatalf("Lint(%s) returned unexpected error: %v", c, err)
	}
	return findingsForCode(report.Findings, "SyncSHASlotFormat")
}

// syncSHAFixtureLine reads line n (1-indexed) of a fixture artifact, so a
// criterion can assert WHICH line was flagged without hardcoding a number that
// any edit to the fixture's prose would invalidate.
func syncSHAFixtureLine(t *testing.T, path string, n int) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(string(data), "\n")
	if n < 1 || n > len(lines) {
		t.Fatalf("line %d out of range in %s (%d lines)", n, path, len(lines))
	}
	return lines[n-1]
}

// TestSyncSHASlot_FlagsProse decides AC-SSF-001 — triad leg (a).
//
// The finding must land on the SIBLING progress.md, not on spec.md:
// `SPECDoc.Body` carries spec.md alone and this field lives exclusively in
// progress.md §E.4, so a body-only rule would see none of the corpus while
// appearing to work.
//
// Mutation that must turn it red: replace the fixture's value with `a6bbbf82b` —
// the finding count must drop to 0. A criterion that stays green under that
// mutation is measuring nothing.
func TestSyncSHASlot_FlagsProse(t *testing.T) {
	got := syncSHAFindings(t, "prose")
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 SyncSHASlotFormat finding, got %d: %v", len(got), got)
	}
	f := got[0]
	if filepath.Base(f.File) != "progress.md" {
		t.Errorf("expected the finding against the sibling progress.md, got %s", f.File)
	}
	if f.Severity != spec.SeverityWarning {
		t.Errorf("expected severity %q, got %q", spec.SeverityWarning, f.Severity)
	}
	// The era precondition, measured rather than asserted in prose: a demoted
	// fixture would carry Advisory == true and silently break AC-SSF-006's
	// --strict half.
	if f.Advisory {
		t.Errorf("finding is advisory — the fixture is being demoted (grandfathered era or terminal status), which is not what this criterion measures")
	}
	line := syncSHAFixtureLine(t, f.File, f.Line)
	if !strings.HasPrefix(line, "sync_commit_sha: TBD") {
		t.Errorf("finding points at %s:%d, whose content is not the flagged slot: %q", f.File, f.Line, line)
	}
}

// TestSyncSHASlot_SilentOnSHA decides AC-SSF-002 — triad leg (b).
//
// THE FOUR FIXTURES ARE NOT REDUNDANT. Each isolates one clause of the §D.1
// grammar: the length band (short vs full), quote stripping on the token, and
// annotation tolerance. A single bare-SHA fixture would stay green under a rule
// that rejects quotes or rejects annotations — and 99 of the corpus's 334
// conforming values carry an annotation, so annotation intolerance would flag
// nearly a third of a healthy corpus while this criterion reported success.
//
// THE TWO BAND-INSIDE FIXTURES (SPEC-SYNCSHA-BAND-BOUNDARY-001, card t380).
// `sha-min7` sits at the inside edge of the floor (7 characters) and `sha-full`
// at the inside edge of the ceiling (40). Before `sha-min7` existed the corpus
// sat at 9 / 40 / 9 / 9, so a floor narrowed by ONE ({7,40} -> {8,40}) was
// invisible to every fixture while the {40} narrowing this comment already
// anticipated was caught — the anticipated mutation killed, the one-step
// mutation surviving. The outside-band half of the pair lives in
// TestSyncSHASlot_FlagsOutOfBand; neither direction of an edge is observable
// from one side alone.
//
// Mutations that must turn it red:
//   - narrow the SHA pattern to {40} — `sha-short` and `sha-quoted` must produce findings
//   - narrow the floor by one, {7,40} -> {8,40} — `sha-min7` must produce a finding
//   - narrow the ceiling by one, {7,40} -> {7,39} — `sha-full` must produce a finding
//   - narrow the alphabet, [0-9a-fA-F] -> [0-9a-f] — `sha-uppercase` must produce a
//     finding. That axis had no fixture before card t397 (SPEC-SYNCSHA-BAND-BOUNDARY-001
//     §C's declared debt): every fixture was lowercase, so the narrowing mutant was
//     indistinguishable from the shipped rule.
func TestSyncSHASlot_SilentOnSHA(t *testing.T) {
	for _, c := range []string{"sha-min7", "sha-short", "sha-full", "sha-quoted", "sha-annotated", "sha-uppercase"} {
		if got := syncSHAFindings(t, c); len(got) != 0 {
			t.Errorf("%s: expected 0 findings on a well-formed SHA, got %d: %v", c, len(got), got)
		}
	}
}

// TestSyncSHASlot_SilentOnPlaceholder decides AC-SSF-003 — triad leg (c).
//
// This is the criterion that keeps the D3 backfill window open. A commit cannot
// cite its own hash, so a placeholder in the phase commit followed by a backfill
// in a later commit is the only workable procedure; a rule failing this forbids
// it, and would do so while AC-SSF-001 and AC-SSF-002 both still read green.
//
// The suffixed fixture is the one that matters for the corpus: refusing the
// `-`-suffixed family would flag the 24 occurrences of `pending-backfill-sync`
// alone.
//
// Mutation that must turn it red: delete the placeholder branch from the lint
// predicate — both fixtures must produce findings.
func TestSyncSHASlot_SilentOnPlaceholder(t *testing.T) {
	for _, c := range []string{"placeholder", "placeholder-suffixed"} {
		if got := syncSHAFindings(t, c); len(got) != 0 {
			t.Errorf("%s: expected 0 findings on a sanctioned backfill placeholder, got %d: %v", c, len(got), got)
		}
	}
}

// TestSyncSHASlot_FlagsOutOfBand decides AC-SBB-003 and AC-SBB-004
// (SPEC-SYNCSHA-BAND-BOUNDARY-001, card t380).
//
// It carries the OUTSIDE half of the band pair: one character below the floor
// and one above the ceiling. It is a separate function rather than a case added
// to TestSyncSHASlot_FlagsProse because that function's name is quoted verbatim
// in a live criterion of a completed SPEC — overloading it would mean a failure
// there no longer identifies AC-SSF-001, and renaming it would falsify that
// criterion's command string. It asserts the same five properties that function
// does, over its own case list, parameterized by the expected flagged-line
// prefix rather than hardcoding one value.
//
// Mutations that must turn it red:
//   - widen the floor by one, {7,40} -> {6,40} — `sha-below6`'s count drops to 0
//   - widen the ceiling by one, {7,40} -> {7,41} — `sha-above41`'s count drops to 0
func TestSyncSHASlot_FlagsOutOfBand(t *testing.T) {
	for _, tc := range []struct {
		fixture string
		prefix  string
	}{
		{"sha-below6", "sync_commit_sha: 19b6f7"},
		{"sha-above41", "sync_commit_sha: a6bbbf82b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f"},
	} {
		got := syncSHAFindings(t, tc.fixture)
		if len(got) != 1 {
			t.Errorf("%s: expected exactly 1 SyncSHASlotFormat finding, got %d: %v", tc.fixture, len(got), got)
			continue
		}
		f := got[0]
		if filepath.Base(f.File) != "progress.md" {
			t.Errorf("%s: expected the finding against the sibling progress.md, got %s", tc.fixture, f.File)
		}
		if f.Severity != spec.SeverityWarning {
			t.Errorf("%s: expected severity %q, got %q", tc.fixture, spec.SeverityWarning, f.Severity)
		}
		// The era precondition, measured rather than asserted in prose: a
		// demoted fixture would carry Advisory == true and the criterion would
		// be measuring the demotion path instead of the band.
		if f.Advisory {
			t.Errorf("%s: finding is advisory — the fixture is being demoted (grandfathered era or terminal status), which is not what this criterion measures", tc.fixture)
		}
		line := syncSHAFixtureLine(t, f.File, f.Line)
		if !strings.HasPrefix(line, tc.prefix) {
			t.Errorf("%s: finding points at %s:%d, whose content is not the flagged slot: %q", tc.fixture, f.File, f.Line, line)
		}
	}
}
