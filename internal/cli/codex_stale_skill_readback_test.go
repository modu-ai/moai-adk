package cli

// codex_stale_skill_readback_test.go — SPEC-CODEX-SKILL-PATH-READBACK-001
// (t562) M1: AC-CSRB-006, the doctor-side regression guard.
//
// The doctor half of the card is verified at OUTPUT level: this guard pins the
// finding's counter decomposition on a fixture whose declared paths are REAL
// (one existing file, one missing, one relative, one backslash-shaped) and
// must stay byte-identical after M2 changes the doctor's stat target.
//
// CORRECTED post-close (t562, sync-audit F1). This comment previously stated
// that doctor_codex.go has zero osStatFn seams. That was true when written at
// M1 and became FALSE at the absorb merge 2d1dad058, which brought in card
// t563's seam: doctor_codex.go now calls osStatFn at :459 and :861. REQ-CSRB-006
// still forbids THIS card from authoring such a seam — that half of the original
// sentence stands, and AC-CSRB-007 pins this card's diff to zero added osStatFn
// lines — but the seam's ABSENCE is no longer a fact, and it must not be read as
// the reason the doctor side is checked at output level only.
//
// The honest statement is that output-level verification here is a CHOICE, not a
// necessity: an injectable seam exists in the tree, so a stat-target assertion on
// the doctor half is writable (stubStatRecording, doctor_codex_stale_skill_test.go).
// It was not written. Consequence, recorded as sync-audit F2: reverting the
// doctor-side conversion to `statPath = e.Path` is caught by nothing here —
// darwin's separator makes the conversion the identity, so this guard is blind to
// it. Closing that is a follow-up card, not a silent edit to a closed one.
//
// The expected values were derived from this PRE-CHANGE run and recorded as
// the M1 baseline (`.moai/reports/t562/ac-csrb-006-baseline.log` +
// `ac-csrb-006-baseline.md`); M3 re-runs this test and diffs against that
// baseline — any diff is a FAIL, never a new expectation.
//
// The fixture rides stubCodexHome (the codexUserHomeDir seam + a blanked
// CODEX_HOME), so the developer's real ~/.codex/config.toml is never read,
// and writeCodexHomeConfig's unwritten-hash guard binds this fixture too. No
// test here performs a deletion.

import (
	"strings"
	"testing"
)

// TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard (AC-CSRB-006) pins
// the finding's counter decomposition: the existing-file entry is NOT counted
// missing, the backslash declaration counts as oddly-formed, the relative one
// as relative, and the missing one lands in its declared `enabled` bucket.
func TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard(t *testing.T) {
	live := liveSkillFile(t)
	absent := absentSkillPath(t, "readback")

	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		// Exists on disk: the healthy registration a wrong stat target
		// would misread as absent. Must resolve — never counted missing.
		{Path: live, EnabledKey: "true"},
		// Missing: the one ghost, landing in the enabled bucket.
		{Path: absent, EnabledKey: "true"},
		// Relative shape: reported, never statted.
		{Path: "relative-skills/SKILL.md", EnabledKey: "true"},
		// Backslash shape on this '/'-separator host: oddly-formed,
		// reported, never statted.
		{Path: `C:\Users\u\SKILL.md`, EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	finding, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatal("codexStaleSkillFinding produced no finding — the fixture did not traverse the finding path, so a baseline captured from it would be empty")
	}

	t.Logf("summary: %s", finding.summary)
	t.Logf("detail: %s", finding.detail)

	// The exercised-count, next to the raw output (the lead's requirement):
	// 4 declared entries flowed through the finding — decomposition below.
	// The "declares 4" phrase IS the parsed count: codexUserSkillConfig
	// renders len(entries) it read from the fixture.
	const declaredCount = 4
	// resolved(1, counted neither missing nor unresolved) +
	// missing(1) + relative(1) + oddlyFormed(1) == declaredCount.
	const resolvedCount = 1
	const wantMissing = 1
	const wantRelative = 1
	const wantOddlyFormed = 1
	if resolvedCount+wantMissing+wantRelative+wantOddlyFormed != declaredCount {
		t.Fatalf("fixture decomposition does not sum to the declared count — the baseline note would be unaccounted")
	}

	detail := finding.detail
	if !strings.Contains(detail, "declares 4 [[skills.config]] entries") {
		t.Errorf("detail does not name 4 declared entries — the fixture did not parse as built:\n%s", detail)
	}
	if !strings.Contains(detail, "1 with a path that no longer exists (1 enabled, 0 disabled, 0 unspecified, 0 non-boolean)") {
		t.Errorf("existing-file entry misread, or the missing split moved:\n%s", detail)
	}
	if !strings.Contains(detail, "1 relative entry (not checked: the resolution base is not observed)") {
		t.Errorf("relative counter moved:\n%s", detail)
	}
	if !strings.Contains(detail, "1 oddly-formed entry (not checked: backslash or ~other-user shape)") {
		t.Errorf("oddly-formed counter moved:\n%s", detail)
	}
	wantSummary := "1 stale skill entry"
	if !strings.Contains(finding.summary, wantSummary) {
		t.Errorf("summary = %q, want it to carry %q", finding.summary, wantSummary)
	}
}
