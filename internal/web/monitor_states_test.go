package web

// monitor_states_test.go — TG-1 (SPEC-WEB-CONSOLE-018 REQ-005).
//
// Defect answer: these tests catch the monitor screen rendering wrong row
// content, omitting a state's row entirely, or dropping the live/i18n wiring
// when the viewmodel state changes — a populated session row losing its
// backend badge branch, a stalled goal losing its badge, a failed verify
// rendering "Pass", or the empty registry rendering a stale row.

import (
	"strings"
	"testing"
)

// tg1ShellVM is the minimal shell state every Monitor render needs.
func tg1ShellVM() ShellVM {
	return ShellVM{
		Area: "monitor", Title: "Monitor", Crumb: "sessions · goals · verification",
		Host: "127.0.0.1:3041", Profile: "default", Project: "proj",
		ProjectPath: "/tmp/proj", Lang: "en", Live: "on", RenderedAt: "12:00:00",
	}
}

// TestMonitorSessionRowsRenderPerState pins the session-table row contract:
// each registry entry contributes a row carrying its ID, backend badge branch
// (glm → metered, claude → flat rate, unrecorded → the missing glyph), state
// mark, and heartbeat — and the panel keeps its data-live wiring while doing it.
func TestMonitorSessionRowsRenderPerState(t *testing.T) {
	vm := MonitorVM{
		Sessions: []SessionVM{
			{ID: "sess-aaa1", SpecID: "SPEC-A-001", Backend: "glm", State: StateLive, Heartbeat: "2m"},
			{ID: "sess-bbb2", SpecID: "", Backend: "claude", State: StateStale, Heartbeat: "31m"},
			{ID: "sess-ccc3", Backend: "", State: StateStale},
		},
		VerifyKeys: 3, Cwd: "/tmp/proj",
	}
	html := renderTempl(t, Monitor(tg1ShellVM(), vm))

	for _, want := range []string{
		`data-live="session"`,     // the htmx refresh marker survives
		`sess-aaa1`, `SPEC-A-001`, // populated row content per session
		`sess-bbb2`, `sess-ccc3`, // every entry renders a row
		`backend--metered`, `metered`, // glm branch
		`flat rate`,                   // claude branch
		`class="missing"`,             // unrecorded backend draws "—"
		`state--live`, `state--stale`, // state mark follows the viewmodel
		`2m`, `31m`, // heartbeat text carried through
		`registry 3`,                  // panel meta counts the entries
		`cwd /tmp/proj`,               // the cwd provenance line
		`data-i18n="monitor.session"`, // header i18n wiring intact
	} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor session table missing %q:\n%s", want, html)
		}
	}
}

// TestMonitorSessionTableEmptyState pins the empty registry: the header row and
// live marker remain, but no data row is invented for a session that does not
// exist — a fabricated row would read as a live session.
func TestMonitorSessionTableEmptyState(t *testing.T) {
	html := renderTempl(t, Monitor(tg1ShellVM(), MonitorVM{VerifyKeys: 0, Cwd: "/tmp/proj"}))

	if strings.Contains(html, "tr--monitor-row") {
		t.Errorf("an empty registry rendered a session row:\n%s", html)
	}
	for _, want := range []string{`data-live="session"`, `tr--monitor-head`, `cwd /tmp/proj`, `registry 0`} {
		if !strings.Contains(html, want) {
			t.Errorf("empty monitor lost %q:\n%s", want, html)
		}
	}
}

// TestMonitorGoalRowsPinStalledAndProgress pins the goals panel: an armed goal
// renders its condition and turn budget, a ceiling-exhausted goal gains the
// "stalled" badge, and the bar width follows the turn percentage.
func TestMonitorGoalRowsPinStalledAndProgress(t *testing.T) {
	vm := MonitorVM{
		Goals: []GoalVM{
			{Session: "sess-goa1", Condition: "coverage converges", Turns: 5, TurnPct: 16, Verdict: "armed"},
			{Session: "sess-gob2", Condition: "spec lands", Turns: 30, TurnPct: 100, Stalled: true, Verdict: "armed"},
		},
		VerifyKeys: 1, Cwd: "/tmp/proj",
	}
	html := renderTempl(t, Monitor(tg1ShellVM(), vm))

	for _, want := range []string{
		`sess-goa1`, `coverage converges`, `sess-gob2`, `spec lands`,
		`data-live="goal"`,        // live wiring on the panel
		`width:16%`, `width:100%`, // bar width = turn percentage
		`badge--outline`, `>stalled<`, // the stalled badge marks exhaustion
		`Turns`, `5`, `30`, // turn counts rendered
	} {
		if !strings.Contains(html, want) {
			t.Errorf("goals panel missing %q:\n%s", want, html)
		}
	}
}

// TestMonitorVerifyRowsPassFailAndSpark pins the verification panel: a snapshot
// whose checks all exited 0 renders "Pass", one with a failure renders the warn
// "Fail", and the spark cells follow the per-check history — a flat all-on
// spark on a failed key would hide the regression the panel exists to show.
func TestMonitorVerifyRowsPassFailAndSpark(t *testing.T) {
	vm := MonitorVM{
		Verify: []VerifyVM{
			{Key: "HEAD:aaaa1111", When: "3m", OK: true, History: []bool{true, true}},
			{Key: "HEAD:bbbb2222", When: "9m", OK: false, History: []bool{true, false}},
		},
		VerifyKeys: 2, Cwd: "/tmp/proj",
	}
	html := renderTempl(t, Monitor(tg1ShellVM(), vm))

	for _, want := range []string{
		`HEAD:aaaa1111`, `HEAD:bbbb2222`, `data-live="verify"`,
		`data-i18n="monitor.pass"`, // clean snapshot renders Pass
		`data-i18n="monitor.fail"`, // failed snapshot renders Fail (not Pass)
		`class="spark"`,            // the spark strip rendered at all
		`spark__b--on`,             // on-cells for passing checks
		`2 keys`,                   // panel meta carries the total key count
	} {
		if !strings.Contains(html, want) {
			t.Errorf("verify panel missing %q:\n%s", want, html)
		}
	}
	// The failing row's spark must carry at least one off-cell: 4 history
	// entries total — 3 on-cells plus exactly one bare off-cell. Counting the
	// bare class (`spark__b"` with the closing quote) avoids matching the
	// on-cell modifier (`spark__b--on`), which contains the same prefix.
	if n := strings.Count(html, `spark__b--on`); n != 3 {
		t.Errorf("spark rendered %d on-cells, want 3", n)
	}
	if n := strings.Count(html, `spark__b"`); n != 1 {
		t.Errorf("spark rendered %d off-cells, want 1", n)
	}
}

// TestMonitorEpicsPanel pins the epic progress rows: prefix, percentage bar
// width, and the progress label flow through unchanged — a dropped field here
// reads as an epic that never advanced.
func TestMonitorEpicsPanel(t *testing.T) {
	vm := MonitorVM{
		Epics:      []EpicVM{{Prefix: "WEB", Progress: "3/7", Pct: 43}, {Prefix: "HARNESS", Progress: "0/4", Pct: 0}},
		VerifyKeys: 1, Cwd: "/tmp/proj",
	}
	html := renderTempl(t, Monitor(tg1ShellVM(), vm))
	for _, want := range []string{`WEB`, `HARNESS`, `3/7`, `0/4`, `width:43%`, `width:0%`} {
		if !strings.Contains(html, want) {
			t.Errorf("epics panel missing %q:\n%s", want, html)
		}
	}
}
