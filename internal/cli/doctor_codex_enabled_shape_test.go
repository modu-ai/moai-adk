package cli

// t508 (SPEC-CODEX-ENABLED-FATAL-001) M3 — fatal-shape detection on the
// `enabled` axis, and the controls that stop it from asserting nothing.
//
// The fixture family here is the six-row acceptance lab
// (.moai/reports/t508/codex-enabled-lab.md) transcribed one-to-one: four fatal
// shapes, two accepting controls, plus a scope control for `path`.
//
// Every control carries a POSITIVE-READ clause as well as a negative one. "Not
// CheckFail" is satisfied just as well by a check that never reached the config
// at all — the codex-not-in-play informational skip returns CheckOK — so each
// control also asserts the check's detail text names the fixture's own config.
// Without that clause the controls would pass against a check that reads
// nothing.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// codexEnabledFixtureText runs the check against a fixture home and returns the
// check plus its message-plus-detail text.
func codexEnabledFixtureText(t *testing.T, entries []codexSkillEntrySpec) (DiagnosticCheck, string) {
	t.Helper()
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, entries)
	stubCodexHome(t, home)
	check := checkCodexWiring(wireProjectForDoctor(t), false)
	return check, check.Message + " " + codexDetailText(check)
}

// TestCheckCodexWiring_NonBooleanEnabled (AC-CEF-002, 003, 004) — a declared
// `enabled` whose value is not a bare TOML boolean is fatal.
//
// The three rows are the three shapes actually measured against codex 0.153.4.
// REQ-CEF-004 generalises from them to "not a bare TOML boolean"; that
// generalisation is an INDUCTION, grounded in codex's uniform “invalid type:
// … expected a boolean“ error text, and unmeasured members of the class (a
// float, an array, an inline table) are not pinned here.
func TestCheckCodexWiring_NonBooleanEnabled(t *testing.T) {
	cases := []struct {
		name       string
		enabledKey string
	}{
		{"integer", "1"},
		{"double_quoted_true", `"true"`},
		{"single_quoted_false", `'false'`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			check, text := codexEnabledFixtureText(t, []codexSkillEntrySpec{
				{Path: liveSkillFile(t), EnabledKey: c.enabledKey},
			})
			if check.Status != uikit.CheckFail {
				t.Errorf("status = %v, want CheckFail — codex refuses `enabled = %s`: %+v",
					check.Status, c.enabledKey, check)
			}
			if !strings.Contains(text, "enabled") {
				t.Errorf("the finding never names the `enabled` key: %+v", check)
			}
		})
	}
}

// TestCheckCodexWiring_PathAbsentIsNotFatal (AC-CEF-007) is the SCOPE control.
//
// Codex accepts a `path`-less entry — measured, rc=0 — where it refuses a
// key-less `enabled`. A fix that widened fatality to `path` would be a scope
// breach that none of the detection cases above would notice.
func TestCheckCodexWiring_PathAbsentIsNotFatal(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{EnabledKey: "true"}, // no path key: accepted by codex
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)

	if check.Status == uikit.CheckFail {
		t.Errorf("an entry declaring no `path` must not be fatal — codex accepts it: %+v", check)
	}
	if !strings.Contains(codexDetailText(check), home) {
		t.Errorf("positive-read clause: the detail never names the fixture's config: %+v", check)
	}
}

// TestCheckCodexWiring_FatalFindingNamesMeasuredVersion (AC-CEF-015) — the
// finding attributes the behaviour to the codex release actually measured, not
// to codex in general.
//
// Only codex-cli 0.153.4 was probed. In which release `enabled` became required
// is UNMEASURED, so a sentence asserting the behaviour holds across releases
// would outrun the evidence this SPEC has.
func TestCheckCodexWiring_FatalFindingNamesMeasuredVersion(t *testing.T) {
	check, text := codexEnabledFixtureText(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t)}, // no enabled key: the fatal shape
	})
	if check.Status != uikit.CheckFail {
		t.Fatalf("fixture did not produce a fatal finding, so there is no text to check: %+v", check)
	}
	if !strings.Contains(text, "0.153.4") {
		t.Errorf("the fatal finding does not name the measured codex version: %+v", check)
	}
	for _, overclaim := range []string{"all codex versions", "every codex", "codex always"} {
		if strings.Contains(strings.ToLower(text), overclaim) {
			t.Errorf("the finding overclaims across codex releases (%q): %+v", overclaim, check)
		}
	}
}

// TestCheckCodexWiring_FatalSummaryStaysInsideWidthBand keeps the new finding
// inside the panel width convention this file already observes. One long
// Message widens the whole doctor box and breaks every other row's alignment.
func TestCheckCodexWiring_FatalSummaryStaysInsideWidthBand(t *testing.T) {
	check, _ := codexEnabledFixtureText(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t)},
		{Path: liveSkillFile(t), EnabledKey: "1"},
	})
	if n := len([]rune(check.Message)); n > codexMessageWidthCeiling {
		t.Errorf("Message is %d runes, ceiling is %d: %q", n, codexMessageWidthCeiling, check.Message)
	}
}
