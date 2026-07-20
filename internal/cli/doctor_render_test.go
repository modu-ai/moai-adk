package cli

// M4c contract tests for doctor live progress + result tables
// (SPEC-CLI-TUX-V3-004 REQ-TUX4-001/002/003). Test names match the AC run
// patterns: 'DoctorLiveProgress|DoctorStep' (AC-TUX4-001) and
// 'DoctorTable|DoctorSectionResult' (AC-TUX4-002).

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestDoctorLiveProgress_SeamPreservesVerdicts verifies the progress observer
// seam wraps every check exactly once, sees the correct group titles, and
// returns verdicts unchanged (AC-TUX4-014: reporter seam only — no verdict
// logic drift).
func TestDoctorLiveProgress_SeamPreservesVerdicts(t *testing.T) {
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")

	type event struct{ group, name string }
	var events []event
	observed := runGroupedChecksObserved(false, "", func(group, name string, run func() DiagnosticCheck) DiagnosticCheck {
		events = append(events, event{group, name})
		return run()
	})
	baseline := runGroupedChecks(false, "")

	// Observer fired once per check.
	total := 0
	for _, g := range baseline {
		total += len(g.checks)
	}
	if len(events) != total {
		t.Errorf("observer fired %d times, want once per check (%d)", len(events), total)
	}
	// Group titles propagate to the observer.
	seenGroups := map[string]bool{}
	for _, ev := range events {
		seenGroups[ev.group] = true
	}
	for _, want := range []string{"System", "MoAI-ADK", "Workspace"} {
		if !seenGroups[want] {
			t.Errorf("observer never saw group %q", want)
		}
	}
	// Verdicts identical to the unobserved run (names + statuses).
	if len(observed) != len(baseline) {
		t.Fatalf("group count drifted: observed %d, baseline %d", len(observed), len(baseline))
	}
	for i := range baseline {
		if len(observed[i].checks) != len(baseline[i].checks) {
			t.Errorf("group %q check count drifted", baseline[i].title)
			continue
		}
		for j := range baseline[i].checks {
			if observed[i].checks[j].Name != baseline[i].checks[j].Name {
				t.Errorf("check name drifted at %s[%d]: %q vs %q",
					baseline[i].title, j, observed[i].checks[j].Name, baseline[i].checks[j].Name)
			}
		}
	}
}

// TestDoctorStep_ProgressOnStderr verifies the doctor command emits per-check
// step feedback through the Printer handles onto stderr, while stdout carries
// only the result surface (REQ-TUX4-001 + channel discipline).
func TestDoctorStep_ProgressOnStderr(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")

	stdout, stderr := captureDoctorCmd(t)

	// Per-check step lines land on stderr.
	for _, name := range []string{"Go Runtime", "Git", "Claude Code"} {
		if !strings.Contains(stderr, name) {
			t.Errorf("stderr should carry a step line for check %q, stderr:\n%s", name, stderr)
		}
	}
	// stdout still carries the result surface (not the step markers).
	if !strings.Contains(stdout, "System Diagnostics") {
		t.Errorf("stdout should carry the diagnostics result surface, got:\n%s", stdout)
	}
	// Plain mode: zero ANSI on both channels (REQ-TUX4-003).
	if strings.Contains(stdout, "\x1b") || strings.Contains(stderr, "\x1b") {
		t.Error("NO_COLOR doctor output must carry zero ANSI on both channels")
	}
}

// TestDoctorTable_PlainAlignedColumns verifies the non-TTY result table is an
// aligned plain-text table with a header row (REQ-TUX4-002 non-TTY pair).
func TestDoctorTable_PlainAlignedColumns(t *testing.T) {
	groups := []checkGroup{
		{title: "System", checks: []DiagnosticCheck{
			{Name: "Alpha", Status: uikit.CheckOK, Message: "alpha fine"},
			{Name: "Beta Longer Name", Status: uikit.CheckWarn, Message: "beta warned"},
		}},
	}
	th := tui.MonochromeTheme()
	body := renderDoctorGroups(nil, groups, false, th)

	if !strings.Contains(body, "STATUS") || !strings.Contains(body, "CHECK") || !strings.Contains(body, "MESSAGE") {
		t.Errorf("plain table should carry STATUS/CHECK/MESSAGE header, got:\n%s", body)
	}
	for _, want := range []string{"Alpha", "alpha fine", "Beta Longer Name", "beta warned", "ok", "warn"} {
		if !strings.Contains(body, want) {
			t.Errorf("plain table lost %q, got:\n%s", want, body)
		}
	}
	if strings.Contains(body, "\x1b") {
		t.Error("monochrome plain table must carry zero ANSI")
	}
	// Alignment: the two check-name cells start at the same column.
	lines := strings.Split(body, "\n")
	colOf := func(sub string) int {
		for _, l := range lines {
			if idx := strings.Index(l, sub); idx >= 0 {
				return idx
			}
		}
		return -1
	}
	if a, b := colOf("Alpha"), colOf("Beta Longer Name"); a < 0 || b < 0 || a != b {
		t.Errorf("check-name column misaligned: Alpha@%d vs Beta@%d\n%s", a, b, body)
	}
}

// TestDoctorSectionResult_Counts verifies per-section counts and overall
// counts render with the table (REQ-TUX4-002).
func TestDoctorSectionResult_Counts(t *testing.T) {
	groups := []checkGroup{
		{title: "System", checks: []DiagnosticCheck{
			{Name: "A", Status: uikit.CheckOK, Message: "m"},
			{Name: "B", Status: uikit.CheckWarn, Message: "m"},
		}},
		{title: "Workspace", checks: []DiagnosticCheck{
			{Name: "C", Status: uikit.CheckFail, Message: "m"},
		}},
	}
	th := tui.MonochromeTheme()
	body := renderDoctorGroups(nil, groups, false, th)

	// Per-section counts.
	if !strings.Contains(body, "1 ok, 1 warn, 0 fail") {
		t.Errorf("System section counts missing, got:\n%s", body)
	}
	if !strings.Contains(body, "0 ok, 0 warn, 1 fail") {
		t.Errorf("Workspace section counts missing, got:\n%s", body)
	}
	// Overall summary pills (mono degrade: [Pass N] form).
	for _, want := range []string{"Pass 1", "Warn 1", "Fail 1"} {
		if !strings.Contains(body, want) {
			t.Errorf("overall summary lost %q, got:\n%s", want, body)
		}
	}
}

// TestDoctorTable_RichUsesBubblesTable verifies the rich path routes through
// the bubbles v2 table component (AC-TUX4-002 reachability at behavior level:
// styled table output with ANSI, carrying the check rows).
func TestDoctorTable_RichUsesBubblesTable(t *testing.T) {
	groups := []checkGroup{
		{title: "System", checks: []DiagnosticCheck{
			{Name: "Alpha", Status: uikit.CheckOK, Message: "alpha fine"},
		}},
	}
	th := tui.DarkTheme()
	body := renderDoctorRichTable(groups[0], th)
	if !strings.Contains(body, "Alpha") || !strings.Contains(body, "alpha fine") {
		t.Errorf("rich table lost row content, got:\n%s", body)
	}
	if !strings.Contains(body, "\x1b[") {
		t.Error("rich bubbles table should carry ANSI styling from theme tokens")
	}
}
