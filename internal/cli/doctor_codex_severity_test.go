package cli

// t508 (SPEC-CODEX-ENABLED-FATAL-001) M1 — the per-finding severity axis.
//
// The Codex Wiring check had NO severity axis before this SPEC: codexFinding
// was {summary, detail} and every problem landed on uikit.CheckWarn. The axis
// is BUILT here, not selected — and the whole point of building it rather than
// flipping the check is that an advisory finding must still surface its text in
// a run that also carries a fatal one.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// TestCodexFindingZeroValueIsAdvisory (AC-CEF-016) pins the ONE property that
// makes the severity axis safe to add to a file with 13 pre-existing
// construction sites: the Go zero value must be the ADVISORY grade.
//
// Every one of those sites builds a codexFinding by composite literal naming no
// severity, so each takes the zero value. Were the enum declared fatal-first —
// the iota ordering its neighbour SkillEnabled uses at skills.go:34, which is
// what copying the adjacent style would produce — all 13 would silently
// re-grade to fatal and `moai doctor` would exit 1 on every advisory machine in
// existence.
//
// This is distinct from the advisory-only run in doctor_exitcode_codex_test.go:
// that case is satisfiable by explicitly setting the advisory grade at all 13
// sites, which leaves the invariant unpinned — the 14th site added later would
// re-grade in silence. This case pins the zero value itself.
func TestCodexFindingZeroValueIsAdvisory(t *testing.T) {
	var zero codexFinding
	if zero.severity != codexSeverityAdvisory {
		t.Errorf("the zero-value codexFinding severity = %v, want codexSeverityAdvisory", zero.severity)
	}

	// The shape all 13 pre-existing sites use: a composite literal naming no
	// severity.
	literal := codexFinding{summary: "s", detail: "d"}
	if literal.severity != codexSeverityAdvisory {
		t.Errorf("a severity-less composite literal = %v, want codexSeverityAdvisory", literal.severity)
	}
	if got := codexCheckStatus([]codexFinding{literal, {summary: "s2", detail: "d2"}}); got != uikit.CheckWarn {
		t.Errorf("a run of severity-less findings = %v, want CheckWarn", got)
	}
}

// TestCodexCheckStatusFoldsAnyFatal pins the fold itself: any fatal finding
// makes the check fail, and an all-advisory run keeps the pre-change behaviour
// byte-identical (uikit.CheckWarn).
func TestCodexCheckStatusFoldsAnyFatal(t *testing.T) {
	cases := []struct {
		name     string
		problems []codexFinding
		want     uikit.CheckStatus
	}{
		{"advisory_only", []codexFinding{{summary: "a"}}, uikit.CheckWarn},
		{"fatal_first", []codexFinding{{summary: "f", severity: codexSeverityFatal}, {summary: "a"}}, uikit.CheckFail},
		{"fatal_last", []codexFinding{{summary: "a"}, {summary: "f", severity: codexSeverityFatal}}, uikit.CheckFail},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := codexCheckStatus(c.problems); got != c.want {
				t.Errorf("codexCheckStatus(%s) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}

// TestCheckCodexWiring_MixedFindingSurfacesAdvisoryText (AC-CEF-010, severity
// half) is the case that proves the axis was BUILT rather than the whole check
// being flipped to fatal: a run producing one fatal finding AND one advisory
// finding reports fatal status while STILL surfacing the advisory finding's
// text.
//
// Flipping the check to CheckFail wholesale would satisfy the fatal assertions
// elsewhere in this suite and drop nothing here — so the load-bearing half is
// the advisory text, not the status.
func TestCheckCodexWiring_MixedFindingSurfacesAdvisoryText(t *testing.T) {
	stubCodexLookup(t, true, true)
	root := wireProjectForDoctor(t)
	if err := os.Remove(filepath.Join(root, codexwiring.HooksRelPath)); err != nil {
		t.Fatalf("removing hooks.json from the wired fixture: %v", err)
	}
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t)}, // no enabled key: the fatal shape
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(root, false)

	if check.Status != uikit.CheckFail {
		t.Errorf("status = %v, want CheckFail — a fatal finding is present: %+v", check.Status, check)
	}
	text := check.Message + " " + codexDetailText(check)
	if !strings.Contains(text, "enabled") {
		t.Errorf("the fatal finding never names the `enabled` key: %+v", check)
	}
	if !strings.Contains(text, "hooks.json missing") {
		t.Errorf("the advisory finding's text was dropped by the fatal grade: %+v", check)
	}
}
