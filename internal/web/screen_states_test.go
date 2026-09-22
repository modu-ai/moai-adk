package web

// screen_states_test.go — TG-2 (SPEC-WEB-CONSOLE-018 REQ-006).
//
// Defect answer: these tests catch the wrong conditional branch rendered for an
// item state — a done card taking the active-badge branch, a must-fix panel
// omitting its entries, a lane rendering another lane's record, specRowLink
// losing its target, or the todo queue presenting a read failure as an empty
// queue.

import (
	"strings"
	"testing"
)

// tg2ShellVM is the minimal shell state the four screens need.
func tg2ShellVM(area string) ShellVM {
	return ShellVM{
		Area: area, Title: "T", Crumb: "c",
		Host: "127.0.0.1:3041", Profile: "default", Project: "proj",
		ProjectPath: "/tmp/proj", Lang: "en", Live: "on", RenderedAt: "12:00:00",
	}
}

// tg2PopulatedRole is a chain role with a live session and full telemetry.
func tg2PopulatedRole(role string) RoleVM {
	return RoleVM{Role: role, Session: "sess-" + role, Backend: "claude", Model: "opus", Effort: "high",
		ContextPct: 42, State: StateLive, Stage: StageActive, StageEstimated: true, Heartbeat: "1m"}
}

// TestKanbanChainRoleStates pins the chain board: an idle role renders the
// idle/blocked branch with the "not started" vocabulary, a populated role
// renders its telemetry, and an unrecorded model/effort/context renders the
// missing glyph instead of a plausible substitute.
func TestKanbanChainRoleStates(t *testing.T) {
	k := KanbanVM{
		CardID: "t1079", IdleRole: "sync",
		Roles: []RoleVM{
			tg2PopulatedRole("lead"), tg2PopulatedRole("plan"), tg2PopulatedRole("run"),
			{Role: "sync", State: StateIdle, Stage: StageBlocked, ContextPct: -1},
		},
	}
	html := renderTempl(t, Kanban(tg2ShellVM("kanban"), k))

	for _, want := range []string{
		`t1079`,                                  // the chain card id
		`chain.stopped`,                          // the idle-role warning branch fired
		`role--idle`,                             // idle role marked on the card
		`state--live`, `state--idle`,             // both state marks present
		`stage--active`, `stage--blocked`,        // both stages drawn
		`mark.estimated`,                         // estimated stages carry the estimate tag
		`sess-lead`, `opus`, `42`,                // telemetry flows into the row
		`No session`,                             // ...and the missing-session branch is absent here
	} {
		if !strings.Contains(html, want) {
			t.Errorf("kanban chain board missing %q:\n%s", want, html)
		}
	}
	if !strings.Contains(html, `data-i18n="kanban.noSession"`) {
		t.Errorf("a missing session must draw the no-session branch:\n%s", html)
	}
	if !strings.Contains(html, `class="missing"`) {
		t.Errorf("the sync row's unrecorded values must draw the missing glyph:\n%s", html)
	}
}

// TestKanbanRoleWithNoTelemetryRecord pins the "no start record" branch: a role
// whose session exists but carries no heartbeat renders the warn instead of a
// blank footer.
func TestKanbanRoleWithNoTelemetryRecord(t *testing.T) {
	k := KanbanVM{Roles: []RoleVM{{Role: "lead", State: StateStale, Stage: StageWait, StageEstimated: true, ContextPct: -1}}}
	html := renderTempl(t, Kanban(tg2ShellVM("kanban"), k))
	if !strings.Contains(html, `data-i18n="kanban.noStart"`) {
		t.Errorf("a role with no heartbeat lost its no-start-record warn:\n%s", html)
	}
}

// TestKanbanLaneStates pins the factory lanes beside the chain: a resolved lane
// renders its own card/spec rows, an unresolved lane renders the
// lane-unresolved warning with its reason (never another lane's record), and an
// empty lane list renders the no-lanes banner rather than a silent nothing.
func TestKanbanLaneStates(t *testing.T) {
	resolved := KanbanVM{Lanes: []LaneVM{{Lane: 1, State: StateLive, Stage: StageActive, StageEstimated: true,
		Session: "sess-lane1", CardID: "t999", SpecID: "SPEC-X-001", Backend: "glm"}}}
	html := renderTempl(t, Kanban(tg2ShellVM("kanban"), resolved))
	for _, want := range []string{`data-lane="1"`, `t999`, `SPEC-X-001`, `backend--metered`} {
		if !strings.Contains(html, want) {
			t.Errorf("resolved lane missing %q:\n%s", want, html)
		}
	}

	unresolved := KanbanVM{Lanes: []LaneVM{{Lane: 2, Unresolved: true, UnresolvedReason: "no-session"}}}
	html = renderTempl(t, Kanban(tg2ShellVM("kanban"), unresolved))
	for _, want := range []string{`data-lane="2"`, `data-lane-unresolved="no-session"`, `kanban.laneUnresolved`} {
		if !strings.Contains(html, want) {
			t.Errorf("unresolved lane missing %q:\n%s", want, html)
		}
	}
	if strings.Contains(html, `role__s"`) && strings.Contains(html, "t999") {
		t.Errorf("an unresolved lane must not render a foreign record:\n%s", html)
	}

	empty := renderTempl(t, Kanban(tg2ShellVM("kanban"), KanbanVM{}))
	if !strings.Contains(empty, `data-i18n="kanban.noLanes"`) {
		t.Errorf("an empty lane list lost the no-lanes banner:\n%s", empty)
	}
}

// TestKanbanPipelineColumns pins view B: each SPEC routes to the column its
// status selects, terminal statuses stay off the board, and a card's link keeps
// its target with tier/drift badges following the row's state.
func TestKanbanPipelineColumns(t *testing.T) {
	k := KanbanVM{
		Total: 3,
		Columns: []PipeColumnVM{
			{ID: "plan", Status: "draft", Cards: []SpecRowVM{{ID: "SPEC-A-001", Title: "Alpha", Tier: "M", Updated: "2026-09-01"}}},
			{ID: "run", Status: "in-progress", Cards: []SpecRowVM{{ID: "SPEC-B-002", Title: "Beta", Drift: "MUST-FIX", Updated: "2026-09-02"}}},
			{ID: "sync", Status: "implemented"},
			{ID: "done", Status: "completed", Cards: []SpecRowVM{{ID: "SPEC-C-003", Title: "Gamma", Updated: "2026-09-03"}}},
		},
	}
	html := renderTempl(t, Kanban(tg2ShellVM("kanban"), k))
	for _, want := range []string{
		`4 status columns · 3`,
		`SPEC-A-001`, `SPEC-B-002`, `SPEC-C-003`,
		`/specs?id=SPEC-A-001`,                   // kcard keeps its route target
		`badge--outline`,                         // tier badge branch
		`badge--danger`,                          // MUST-FIX drift badge branch
	} {
		if !strings.Contains(html, want) {
			t.Errorf("pipeline board missing %q:\n%s", want, html)
		}
	}
}

// TestTodoScreenStates pins the three queue states: a read failure renders the
// unavailable status (never as an empty queue), an empty queue renders its own
// empty state, and populated items render rows whose state badge follows the
// card state — with a missing SPEC drawing the missing glyph.
func TestTodoScreenStates(t *testing.T) {
	unavailable := renderTempl(t, Todo(tg2ShellVM("todo"), TodoVM{Root: "/q", Unavailable: true}))
	for _, want := range []string{`data-todo-unavailable`, `todo.unavailable`} {
		if !strings.Contains(unavailable, want) {
			t.Errorf("an unreadable queue must render the unavailable state, missing %q:\n%s", want, unavailable)
		}
	}
	if strings.Contains(unavailable, `data-todo-row`) {
		t.Errorf("an unavailable queue rendered card rows:\n%s", unavailable)
	}

	empty := renderTempl(t, Todo(tg2ShellVM("todo"), TodoVM{Root: "/q"}))
	if !strings.Contains(empty, `data-todo-empty`) {
		t.Errorf("an empty queue lost its empty state:\n%s", empty)
	}

	populated := renderTempl(t, Todo(tg2ShellVM("todo"), TodoVM{Root: "/q", Items: []TodoItemVM{
		{ID: "t1", Text: "first", State: "queued", SpecID: "SPEC-A-001"},
		{ID: "t2", Text: "second", State: "picked"},
		{ID: "t3", Text: "third", State: "dropped"},
	}}))
	for _, want := range []string{
		`data-todo-row`, `t1`, `first`, `t2`, `t3`,
		`badge--outline`, `badge--strong`, `badge--danger`, // one badge branch per state
		`data-todo-state="queued"`, `data-todo-state="picked"`, `data-todo-state="dropped"`,
		`SPEC-A-001`, `class="missing"`, // spec present vs absent
	} {
		if !strings.Contains(populated, want) {
			t.Errorf("populated queue missing %q:\n%s", want, populated)
		}
	}
}

// TestSpecsRowsPanelsAndDetail pins the SPEC list screen: row selection, drift
// and tier badges on the row branch, the two attention panels in their empty
// and populated branches (including the truncation notice and the copy-only
// remediation contract), and the detail slide rendering only when a row is
// selected.
func TestSpecsRowsPanelsAndDetail(t *testing.T) {
	l := SpecListVM{
		SelectedID: "SPEC-B-002",
		Rows: []SpecRowVM{
			{ID: "SPEC-A-001", Title: "Alpha", Status: "draft", Tier: "S", Updated: "2026-09-01"},
			{ID: "SPEC-B-002", Title: "Beta", Status: "in-progress", Drift: "MUST-FIX", Updated: "2026-09-02"},
		},
		CloseDebt: []SpecRowVM{{ID: "SPEC-D-004", Title: "Debt", Status: "implemented", Updated: "2026-09-03"}},
		MustFix: []MustFixVM{
			{SpecID: "SPEC-B-002", FindingType: "owner-drift", Remediation: "moai spec audit --fix SPEC-B-002"},
			{SpecID: "SPEC-E-005", FindingType: "stale-era"},
		},
		Detail: &SpecDetailVM{
			ID: "SPEC-B-002", Title: "Beta", Status: "in-progress", Tier: "M", Era: "V3R6",
			Path: ".moai/specs/SPEC-B-002", Docs: []string{"spec.md", "plan.md"},
			Findings: []FindingVM{{Severity: "MUST-FIX", Message: "fix the owner", File: "owner-drift"}},
		},
	}
	html := renderTempl(t, Specs(tg2ShellVM("specs"), l))

	for _, want := range []string{
		`tr--sel`,                                      // the selected row is marked
		`badge--danger`,                                // MUST-FIX drift badge on the row
		`/specs?id=SPEC-A-001`,                         // row link target preserved
		`class="slide"`,                                // the detail slide markup present
		`SPEC-B-002`, `V3R6`, `spec.md`,                // detail fields carried
		`drift 1`,                                      // the findings count header
		`data-copy="moai spec audit --fix SPEC-B-002"`, // remediation is copy-only
		`board.copy`,                                   // the copy button rendered
	} {
		if !strings.Contains(html, want) {
			t.Errorf("specs screen missing %q:\n%s", want, html)
		}
	}
}

// TestSpecsPanelsEmptyBranches pins the panel empty states and the finding
// without a remediation: an empty close-debt panel says so, a must-fix finding
// with no remediation renders without a copy control (never with an empty
// command to copy), and no selection means no detail slide.
func TestSpecsPanelsEmptyBranches(t *testing.T) {
	l := SpecListVM{
		Rows:      []SpecRowVM{{ID: "SPEC-A-001", Title: "Alpha", Status: "draft"}},
		CloseDebt: []SpecRowVM{},
		MustFix:   []MustFixVM{{SpecID: "SPEC-E-005", FindingType: "stale-era"}},
	}
	html := renderTempl(t, Specs(tg2ShellVM("specs"), l))
	if !strings.Contains(html, `board.closedebt.empty`) {
		t.Errorf("an empty close-debt list must say so:\n%s", html)
	}
	if !strings.Contains(html, `SPEC-E-005`) {
		t.Errorf("a must-fix finding without remediation must still name its spec:\n%s", html)
	}
	if strings.Contains(html, `class="slide"`) {
		t.Errorf("no selection must not render a detail slide:\n%s", html)
	}
	if strings.Contains(html, "data-copy") {
		t.Errorf("a finding without remediation must not render a copy control:\n%s", html)
	}
}

// TestCloseDebtPanelTruncationNotice pins the truncation contract: more rows
// than the panel shows renders the showing-N-of-M notice and a see-all link —
// a silent cut reads as "this is everything".
func TestCloseDebtPanelTruncationNotice(t *testing.T) {
	rows := make([]SpecRowVM, closeDebtShown+3)
	for i := range rows {
		rows[i] = SpecRowVM{ID: "SPEC-T-" + itoa(i), Status: "implemented", Updated: "2026-09-0" + itoa(i%9+1)}
	}
	html := renderTempl(t, closeDebtPanel(rows))
	if !strings.Contains(html, "board.closedebt.showing") {
		t.Errorf("an over-long close-debt list lost its truncation notice:\n%s", html)
	}
	if !strings.Contains(html, `/specs?status=implemented`) {
		t.Errorf("the truncation notice lost its see-all link:\n%s", html)
	}
	// Exactly closeDebtShown row links render, not the whole list. The closing
	// quote keeps the row class apart from its __id/__title/__updated child
	// classes, which share the same prefix.
	if n := strings.Count(html, `spec-summary-row"`); n != closeDebtShown {
		t.Errorf("close-debt panel rendered %d rows, want %d", n, closeDebtShown)
	}
}

// TestSpecRowLinkBadges pins the row-link badge branches: drift renders a
// badge, tier renders a badge, and neither renders nothing — a row with no
// drift must not carry a danger badge.
func TestSpecRowLinkBadges(t *testing.T) {
	full := renderTempl(t, specRowLink(SpecRowVM{ID: "SPEC-A-001", Title: "A", Drift: "MUST-FIX", Tier: "L", Updated: "u"}))
	for _, want := range []string{`/specs?id=SPEC-A-001`, `badge--danger`, `badge--outline`, `MUST-FIX`, `L`} {
		if !strings.Contains(full, want) {
			t.Errorf("spec row link missing %q:\n%s", want, full)
		}
	}
	plain := renderTempl(t, specRowLink(SpecRowVM{ID: "SPEC-B-002", Title: "B", Updated: "u"}))
	if strings.Contains(plain, "badge") {
		t.Errorf("a clean row must not render badges:\n%s", plain)
	}
}

// TestSpecDetailWithoutFindings pins the detail slide's clean branch: no
// findings section is emitted when the audit found nothing — an empty "drift 0"
// header would read as a suppressed result.
func TestSpecDetailWithoutFindings(t *testing.T) {
	html := renderTempl(t, specDetail(SpecDetailVM{ID: "SPEC-C-003", Title: "C", Status: "draft", Path: ".moai/specs/SPEC-C-003", Docs: []string{"spec.md"}}))
	if strings.Contains(html, "drift ") {
		t.Errorf("a finding-free detail rendered a drift section:\n%s", html)
	}
	for _, want := range []string{`SPEC-C-003`, `spec.md`, `/specs`} {
		if !strings.Contains(html, want) {
			t.Errorf("detail slide missing %q:\n%s", want, html)
		}
	}
}

// TestOverviewStateBranches pins the overview's three-way todo branch and the
// health-grid empty branch: an unavailable queue, an empty queue, and a
// populated queue each render their own panel state, and an empty stats list
// renders the empty note instead of an empty grid.
func TestOverviewStateBranches(t *testing.T) {
	base := OverviewVM{}
	unavailable := renderTempl(t, Overview(tg2ShellVM("overview"), base, TodoVM{Root: "/q", Unavailable: true}))
	if !strings.Contains(unavailable, `data-todo-unavailable`) {
		t.Errorf("overview lost the unavailable-queue state:\n%s", unavailable)
	}
	empty := renderTempl(t, Overview(tg2ShellVM("overview"), base, TodoVM{Root: "/q"}))
	if !strings.Contains(empty, `data-i18n="todo.empty"`) {
		t.Errorf("overview lost the empty-queue state:\n%s", empty)
	}
	populated := renderTempl(t, Overview(tg2ShellVM("overview"), OverviewVM{
		Stats: []StatVM{{Label: "work tracked", Value: "3", Note: "1 in-progress", NoteKey: "statNote.in-progress", NoteParams: "1"}},
	}, TodoVM{Root: "/q", Items: []TodoItemVM{{ID: "t1", Text: "x", State: "queued"}}}))
	for _, want := range []string{`health-stat__label`, `statNote.in-progress`, `todo-summary__row`, `t1`} {
		if !strings.Contains(populated, want) {
			t.Errorf("populated overview missing %q:\n%s", want, populated)
		}
	}
	if !strings.Contains(populated, `data-i18n="stat.work-tracked"`) {
		t.Errorf("a health stat lost its derived i18n key:\n%s", populated)
	}
}
