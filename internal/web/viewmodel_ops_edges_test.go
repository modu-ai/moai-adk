package web

// viewmodel_ops_edges_test.go — TG-6 (SPEC-WEB-CONSOLE-018 REQ-010).
//
// Defect answer: these tests catch a viewmodel returning a degradation or edge
// value that the screen layer does not distinguish — a dead session promoted to
// live, an unrecorded context drawn as 0%, a terminal SPEC landing on the
// pipeline board, a must-fix remediation mapped to the wrong field, or a goal
// past its ceiling rendering as merely "armed". Idiom: unit tests over the real
// logic (viewmodel_degradation_test.go), fixtures under t.TempDir().

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// TestSessionStateOrGate pins the live/stale decision: an alive PID is live on
// its own (GH #1711 — AND-gating buried 144 of 147 live sessions under stale
// heartbeats), a fresh heartbeat is the conservative fallback, and a dead PID
// with a stale heartbeat stays stale.
func TestSessionStateOrGate(t *testing.T) {
	now := time.Now()
	self := os.Getpid() // a pid the OS guarantees is alive

	if got := sessionState(now.Add(-time.Hour), self, now); got != StateLive {
		t.Errorf("alive PID with stale heartbeat = %q, want live (the OR gate)", got)
	}
	if got := sessionState(now.Add(-time.Minute), 0, now); got != StateLive {
		t.Errorf("fresh heartbeat without a PID = %q, want live (conservative fallback)", got)
	}
	if got := sessionState(now.Add(-time.Hour), 0, now); got != StateStale {
		t.Errorf("dead PID with stale heartbeat = %q, want stale", got)
	}
}

// TestEstimateStageBranches pins the stage mapping: live estimates active,
// stale estimates wait — both tagged as estimates by the bool — and a missing
// session is the one blocked stage that is a fact, not an estimate.
func TestEstimateStageBranches(t *testing.T) {
	stage, estimated := estimateStage(SessionVM{State: StateLive})
	if stage != StageActive || !estimated {
		t.Errorf("live → (%q, %v), want (active, true)", stage, estimated)
	}
	stage, estimated = estimateStage(SessionVM{State: StateStale})
	if stage != StageWait || !estimated {
		t.Errorf("stale → (%q, %v), want (wait, true)", stage, estimated)
	}
	stage, estimated = estimateStage(SessionVM{})
	if stage != StageBlocked || estimated {
		t.Errorf("missing session → (%q, %v), want (blocked, false)", stage, estimated)
	}
}

// TestTelemetryCellsAbsence pins the telemetry contract: no record yields empty
// cells and -1 (the "draw missing" value), a record without a context window
// keeps -1, and a recorded percentage clamps at 100 — an unrecorded context
// drawn as 0% reads as an empty context.
func TestTelemetryCellsAbsence(t *testing.T) {
	model, effort, pct := telemetryCells(nil)
	if model != "" || effort != "" || pct != -1 {
		t.Errorf("nil record → (%q, %q, %d), want empty/-1", model, effort, pct)
	}
	model, effort, pct = telemetryCells(&statusline.SessionTelemetryRecord{})
	if pct != -1 {
		t.Errorf("record without a context window → pct %d, want -1", pct)
	}
	full := &statusline.SessionTelemetryRecord{ContextWindowSize: 200000, RawPct: 187.5}
	if _, _, pct = telemetryCells(full); pct != 100 {
		t.Errorf("an over-range percentage = %d, want clamped to 100", pct)
	}
}

// TestPipelineColumnsRouting pins view B's routing: each status lands in its
// column, the planned legacy value joins draft, terminal statuses stay off the
// board, and within a column the newest card sorts first.
func TestPipelineColumnsRouting(t *testing.T) {
	cols := pipelineColumns([]SpecRowVM{
		{ID: "S-DONE", Status: "completed", Updated: "2026-09-01"},
		{ID: "S-TERM", Status: "superseded", Updated: "2026-09-02"},
		{ID: "S-OLD", Status: "in-progress", Updated: "2026-09-03"},
		{ID: "S-NEW", Status: "in-progress", Updated: "2026-09-04"},
		{ID: "S-PLANNED", Status: "planned", Updated: "2026-09-05"},
	})
	if got := colIDs(cols[1]); got[0] != "S-NEW" || got[1] != "S-OLD" {
		t.Errorf("run column ordered %v, want newest first", got)
	}
	if got := colIDs(cols[0]); len(got) != 1 || got[0] != "S-PLANNED" {
		t.Errorf("plan column = %v, want the planned legacy value routed beside draft", got)
	}
	for _, col := range cols {
		for _, id := range colIDs(col) {
			if id == "S-TERM" {
				t.Errorf("terminal status %s reached the board (column %s)", id, col.ID)
			}
		}
	}
}

func colIDs(col PipeColumnVM) []string {
	out := make([]string, len(col.Cards))
	for i, c := range col.Cards {
		out[i] = c.ID
	}
	return out
}

// TestHumanSinceBranches pins the relative-time vocabulary: an absent timestamp
// renders empty, and each magnitude band renders its unit — a day-old value
// rendered "just now" lies about freshness.
func TestHumanSinceBranches(t *testing.T) {
	now := time.Now()
	cases := []struct {
		age  time.Duration
		want string
	}{
		{30 * time.Second, "just now"},
		{5 * time.Minute, "5m"},
		{3 * time.Hour, "3h"},
		{2 * 24 * time.Hour, "2d"},
	}
	for _, tc := range cases {
		if got := humanSince(now.Add(-tc.age), now); got != tc.want {
			t.Errorf("humanSince(%s) = %q, want %q", tc.age, got, tc.want)
		}
	}
	if got := humanSince(time.Time{}, now); got != "" {
		t.Errorf("humanSince(zero) = %q, want empty", got)
	}
}

// TestClampPct pins the clamp: negatives pass through untouched (they mean
// "not recorded" downstream), over-range clamps to 100.
func TestClampPct(t *testing.T) {
	if got := clampPct(-7); got != -1 {
		t.Errorf("clampPct(-7) = %d, want -1 (pass-through)", got)
	}
	if got := clampPct(187); got != 100 {
		t.Errorf("clampPct(187) = %d, want 100", got)
	}
	if got := clampPct(42); got != 42 {
		t.Errorf("clampPct(42) = %d, want 42", got)
	}
}

// TestBuildAttentionOrderAndCap pins the attention list: the chain's idle role
// leads, MUST-FIX findings follow with their spec-targeted links, non-must
// severities are skipped, and the list caps so a catastrophic audit cannot
// drown the kanban warning.
func TestBuildAttentionOrderAndCap(t *testing.T) {
	chain := ChainVM{Present: true, IdleRole: "sync"}
	rows := []SpecRowVM{{ID: "SPEC-A-001"}, {ID: "SPEC-B-002"}}
	findings := map[string][]FindingVM{}
	for i := 0; i < 12; i++ {
		findings["SPEC-A-001"] = append(findings["SPEC-A-001"], FindingVM{Severity: "MUST-FIX", Message: "fix", File: "type"})
	}
	findings["SPEC-B-002"] = []FindingVM{{Severity: "SHOULD-FIX", Message: "minor", File: "type"}}

	got := buildAttention(rows, findings, chain)
	if len(got) != maxOverviewRows {
		t.Fatalf("attention rows = %d, want capped at %d", len(got), maxOverviewRows)
	}
	first := got[0]
	if first.Source != "kanban" || first.Role != "sync" || first.Href != "/kanban" {
		t.Errorf("the idle-role warning did not lead: %+v", first)
	}
	if got[1].Href != "/specs?id=SPEC-A-001" {
		t.Errorf("must-fix rows lost their spec target: %+v", got[1])
	}
	for _, a := range got {
		if a.Text == "minor" {
			t.Errorf("a SHOULD-FIX finding reached the attention list: %+v", a)
		}
	}

	quiet := buildAttention(rows, map[string][]FindingVM{}, ChainVM{Present: true})
	if len(quiet) != 0 {
		t.Errorf("a healthy chain and clean audit produced attention rows: %+v", quiet)
	}
}

// TestChainCardIDAndRoleFilter pins the chain helpers: the first recorded SPEC
// becomes the card id, a chain with no SPEC has none, and lane records are
// filtered out of the chain feed while chain roles survive.
func TestChainCardIDAndRoleFilter(t *testing.T) {
	if got := chainCardID([]KanbanRecord{{SpecID: ""}, {SpecID: "SPEC-A-001"}}); got != "SPEC-A-001" {
		t.Errorf("chainCardID = %q, want the first recorded SPEC", got)
	}
	if got := chainCardID([]KanbanRecord{{SpecID: ""}}); got != "" {
		t.Errorf("chainCardID = %q, want empty for a plan-stage chain", got)
	}

	records := []KanbanRecord{
		{SessionID: "s1", Role: "Lead"}, // case-insensitive role match
		{SessionID: "s2", Role: "lane"},
		{SessionID: "s3", Role: ""},
	}
	filtered := chainRoleRecords(records)
	if len(filtered) != 1 || filtered[0].SessionID != "s1" {
		t.Errorf("chainRoleRecords kept %d records, want only the chain role", len(filtered))
	}
}

// TestCloseDebtRowsAndBound pins the close-debt derivation: only implemented
// SPECs, newest first, and the panel bound cuts the tail without dropping the
// truncation notice's inputs.
func TestCloseDebtRowsAndBound(t *testing.T) {
	rows := []SpecRowVM{
		{ID: "A", Status: "implemented", Updated: "2026-09-01"},
		{ID: "B", Status: "completed", Updated: "2026-09-02"},
		{ID: "C", Status: "implemented", Updated: "2026-09-03"},
	}
	got := closeDebtRows(rows)
	if len(got) != 2 || got[0].ID != "C" {
		t.Errorf("closeDebtRows = %+v, want C then A, newest first", got)
	}

	many := make([]SpecRowVM, closeDebtShown+5)
	for i := range many {
		many[i] = SpecRowVM{ID: "X", Status: "implemented"}
	}
	if got := boundedRows(many); len(got) != closeDebtShown {
		t.Errorf("boundedRows cut to %d, want %d", len(got), closeDebtShown)
	}
	if got := boundedRows(rows[:1]); len(got) != 1 {
		t.Errorf("boundedRows changed a short list: %d rows", len(got))
	}
}

// TestMustFixFindingsMapping pins the in-place rename contract: the finding's
// Remediation arrives in Message and its FindingType in File (the loadSpecRows
// packing), and mustFixFindings unpacks them to the correctly named fields —
// swapped, the console would copy a finding type as the fix command.
func TestMustFixFindingsMapping(t *testing.T) {
	rows := []SpecRowVM{{ID: "SPEC-A-001"}, {ID: "SPEC-B-002"}}
	findings := map[string][]FindingVM{
		"SPEC-A-001": {{Severity: "MUST-FIX", Message: "moai spec audit --fix", File: "owner-drift"}},
		"SPEC-B-002": {{Severity: "SHOULD-FIX", Message: "minor", File: "type"}},
	}
	got := mustFixFindings(rows, findings)
	if len(got) != 1 {
		t.Fatalf("mustFixFindings returned %d rows, want 1", len(got))
	}
	if got[0].SpecID != "SPEC-A-001" || got[0].Remediation != "moai spec audit --fix" || got[0].FindingType != "owner-drift" {
		t.Errorf("must-fix fields mapped wrong: %+v", got[0])
	}
}

// TestBuildFiltersPinsCountsAndActive pins the filter rail: counts come from
// the status census, the active filter is marked, and the href preserves an
// active query while a no-status href drops the trailing separator.
func TestBuildFiltersPinsCountsAndActive(t *testing.T) {
	counts := map[string]int{"draft": 2, "in-progress": 1}
	got := buildFilters(counts, "draft", "cov", 7)
	if len(got) != 5 {
		t.Fatalf("buildFilters returned %d filters, want 5", len(got))
	}
	// An active status filter means the all-filter is NOT active — a filter
	// rail where "all" and "draft" light up together reads as two selections.
	if got[0].Active || got[0].Count != 7 || got[0].Href != "/specs?q=cov" {
		t.Errorf("the all-filter is wrong: %+v", got[0])
	}
	draft := got[1]
	if draft.Label != "draft" || draft.Count != 2 || !draft.Active || draft.Href != "/specs?q=cov&status=draft" {
		t.Errorf("the draft filter is wrong: %+v", draft)
	}
	noQuery := buildFilters(counts, "", "", 3)
	// Pin the current all-filter target for a query-less request: "/specs?"
	// (the trailing separator is deliberately untrimmed). A change here moves
	// the default filter landing route — that is the regression this pins.
	if noQuery[0].Href != "/specs?" {
		t.Errorf("the query-less all-filter target moved: %q, want /specs?", noQuery[0].Href)
	}
}

// TestLoadGoalsReadsStalledAndSkipsJunk pins the goal loader: an armed goal
// under its ceiling reads with its turn percentage, a goal at its ceiling is
// flagged stalled, and verdict/malformed files in the state dir are skipped —
// a skipped file must degrade to absence, never to a wrong goal row.
func TestLoadGoalsReadsStalledAndSkipsJunk(t *testing.T) {
	root := t.TempDir()

	armed := goal.NewGoal("sess-goala1", "coverage converges", nil)
	armed.TurnsUsed = 5
	if err := goal.SaveGoal(root, armed); err != nil {
		t.Fatalf("save armed goal: %v", err)
	}
	stalled := goal.NewGoal("sess-goalb2", "spec lands", nil)
	stalled.TurnsUsed = stalled.Ceiling.MaxTurns // at the ceiling
	if err := goal.SaveGoal(root, stalled); err != nil {
		t.Fatalf("save stalled goal: %v", err)
	}
	// Junk that must be skipped: a verdict sibling and an unparseable state file.
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state", "goal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "state", "goal", "zzz.verdict.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "state", "goal", "broken.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := loadGoals(root)
	if len(got) != 2 {
		t.Fatalf("loadGoals returned %d goals, want 2 (junk skipped)", len(got))
	}
	byCond := map[string]GoalVM{}
	for _, g := range got {
		byCond[g.Condition] = g
	}
	armedRow, ok := byCond["coverage converges"]
	if !ok {
		t.Fatalf("the armed goal is missing: %+v", got)
	}
	if armedRow.Turns != 5 || armedRow.Stalled || armedRow.TurnPct != 16 {
		t.Errorf("armed goal row wrong: %+v (want turns 5, pct 16, not stalled)", armedRow)
	}
	stalledRow := byCond["spec lands"]
	if !stalledRow.Stalled {
		t.Errorf("a goal at its ceiling did not flag stalled: %+v", stalledRow)
	}
}

// TestLoadVerifySnapshotsPinsHistoryAndLimit pins the verify loader: a snapshot
// whose checks all exited 0 reads OK, one failing check flips the whole row,
// history trims to the last 8 cells, and the limit returns the newest rows with
// the total still counted — a limit that also shrank the total would mislabel
// the key count on the monitor panel.
//
// The fixture keys are filename-safe (no ":"): the loader re-derives each key
// from the snapshot filename, and verify.Save sanitizes a "HEAD:sha" key into a
// "HEAD-sha" filename, so a colon key saved via verify.Save does not round-trip
// through loadVerify (Snapshot.Key != filename-derived key → treated absent).
// That divergence is recorded in progress.md §E.2 as a discovered-defect
// candidate; fixing it is a product-code change outside this SPEC's tests-only
// scope.
func TestLoadVerifySnapshotsPinsHistoryAndLimit(t *testing.T) {
	root := t.TempDir()

	clean := &verify.Snapshot{Key: "head-aaaa1111", RecordedAt: time.Now().Add(-time.Hour)}
	for i := 0; i < 10; i++ {
		clean.Checks = append(clean.Checks, verify.CheckEntry{Command: "cmd", ExitCode: 0, RecordedAt: time.Now()})
	}
	if err := verify.Save(root, clean); err != nil {
		t.Fatalf("save clean snapshot: %v", err)
	}
	failed := &verify.Snapshot{Key: "head-bbbb2222", RecordedAt: time.Now()}
	failed.Checks = []verify.CheckEntry{{Command: "cmd", ExitCode: 1, RecordedAt: time.Now()}}
	if err := verify.Save(root, failed); err != nil {
		t.Fatalf("save failed snapshot: %v", err)
	}

	rows, total := loadVerify(root, maxVerifyRows)
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	// Newest first: the failed snapshot was recorded later.
	if rows[0].Key != "head-bbbb2222" || rows[0].OK {
		t.Errorf("newest row wrong: %+v (want the failed snapshot)", rows[0])
	}
	if rows[1].Key != "head-aaaa1111" || !rows[1].OK {
		t.Errorf("clean row wrong: %+v", rows[1])
	}
	if n := len(rows[1].History); n != 8 {
		t.Errorf("history kept %d cells, want the 8-cell window", n)
	}

	limited, limitedTotal := loadVerify(root, 1)
	if len(limited) != 1 || limitedTotal != 2 {
		t.Errorf("limit=1 → (%d rows, total %d), want (1, 2)", len(limited), limitedTotal)
	}

	if rows, total := loadVerify(filepath.Join(root, "nope"), 0); rows != nil || total != 0 {
		t.Errorf("an absent verify dir → (%v, %d), want (nil, 0)", rows, total)
	}
}

// TestBuildMonitorAssemblesRealState pins the monitor assembly end-to-end over
// fixture files: the cwd is the project root, goals and verify snapshots flow
// through their loaders, and the key count reflects the snapshot total.
func TestBuildMonitorAssemblesRealState(t *testing.T) {
	root := t.TempDir()
	a := newTestApp(t)
	a.cfg.ProjectRoot = root

	g := goal.NewGoal("sess-monit1", "converge", nil)
	if err := goal.SaveGoal(root, g); err != nil {
		t.Fatalf("save goal: %v", err)
	}
	snap := &verify.Snapshot{Key: "head-cccc3333", RecordedAt: time.Now()}
	snap.Checks = []verify.CheckEntry{{Command: "cmd", ExitCode: 0, RecordedAt: time.Now()}}
	if err := verify.Save(root, snap); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}

	vm, err := a.buildMonitor(time.Now())
	if err != nil {
		t.Fatalf("buildMonitor: %v", err)
	}
	if vm.Cwd != root {
		t.Errorf("cwd = %q, want the project root", vm.Cwd)
	}
	if len(vm.Goals) != 1 || vm.Goals[0].Condition != "converge" {
		t.Errorf("goals wrong: %+v", vm.Goals)
	}
	if vm.VerifyKeys != 1 || len(vm.Verify) != 1 || !vm.Verify[0].OK {
		t.Errorf("verify rows wrong: keys=%d rows=%+v", vm.VerifyKeys, vm.Verify)
	}
}

// TestBuildOverviewAndSpecListOverFixture pins the two heaviest assemblies over
// a real SPEC catalog: the overview counts rows and flags the implemented SPEC
// as in-progress attention-free, and the spec list filters by status, selects
// the detail slide, and derives close-debt from the catalog — not from the
// active filter.
func TestBuildOverviewAndSpecListOverFixture(t *testing.T) {
	root := t.TempDir()
	writeBoardSpec(t, root, "SPEC-EDGE-DONE-001", "id: SPEC-EDGE-DONE-001\nstatus: implemented\ntitle: Edge done\ntier: S\n", "")
	writeBoardSpec(t, root, "SPEC-EDGE-PLAN-002", "id: SPEC-EDGE-PLAN-002\nstatus: draft\ntitle: Edge plan\n", "")

	a := newTestApp(t)
	a.cfg.ProjectRoot = root

	overview, err := a.buildOverview(time.Now())
	if err != nil {
		t.Fatalf("buildOverview: %v", err)
	}
	if len(overview.Stats) == 0 {
		t.Error("the overview produced no stats")
	}
	if len(overview.InProgress) != 0 {
		t.Errorf("a draft-only catalog produced in-progress rows: %+v", overview.InProgress)
	}

	list, err := a.buildSpecList("edge", "", "SPEC-EDGE-DONE-001")
	if err != nil {
		t.Fatalf("buildSpecList: %v", err)
	}
	if len(list.Rows) != 2 {
		t.Errorf("query 'edge' matched %d rows, want 2", len(list.Rows))
	}
	if list.Detail == nil || list.Detail.ID != "SPEC-EDGE-DONE-001" {
		t.Errorf("the selected row produced no detail slide: %+v", list.Detail)
	}
	if len(list.CloseDebt) != 1 || list.CloseDebt[0].ID != "SPEC-EDGE-DONE-001" {
		t.Errorf("close debt is wrong: %+v (close debt spans the catalog, not the filter)", list.CloseDebt)
	}
	if got := buildSpecListSelectedAbsent(a, "", ""); got.Detail != nil {
		t.Errorf("no selection rendered a detail slide: %+v", got.Detail)
	}
}

// TestLoadVerifyColonKeyDivergence is a deliberate tripwire pinning a
// discovered divergence (SPEC-WEB-CONSOLE-018 run phase, 2026-09-22).
//
// Defect answer: catches the monitor verify panel reading EMPTY while
// snapshots exist. loadVerify re-derives each key from the snapshot filename,
// but verify.Save sanitizes the canonical "HEAD:<sha>" key into a
// "HEAD-<sha>" filename; the loader then consults verify.Load with the
// filename-derived key, the stored Snapshot.Key does not match, and the row
// is treated as absent. A fix that closes the gap MUST flip this test
// together with the loader — the assertion describes the divergence, not a
// desired contract.
func TestLoadVerifyColonKeyDivergence(t *testing.T) {
	root := t.TempDir()
	snap := &verify.Snapshot{Key: "HEAD:dddd4444", RecordedAt: time.Now()}
	snap.Checks = []verify.CheckEntry{{Command: "cmd", ExitCode: 0, RecordedAt: time.Now()}}
	if err := verify.Save(root, snap); err != nil {
		t.Fatalf("save colon-key snapshot: %v", err)
	}
	// The snapshot file exists on disk under its sanitized name...
	entries, err := os.ReadDir(filepath.Join(root, ".moai", "state", "verify", "snapshots"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("snapshot file missing: entries=%d err=%v", len(entries), err)
	}
	// ...yet the loader reports nothing for it. When the divergence is fixed,
	// this assertion flips to total==1 / rows visible — update it in the same
	// commit as the fix.
	rows, total := loadVerify(root, maxVerifyRows)
	if total != 0 || len(rows) != 0 {
		t.Errorf("colon-key snapshot became visible to loadVerify (total=%d rows=%d): "+
			"the divergence was fixed — flip this tripwire in the same commit as the fix", total, len(rows))
	}
}

// buildSpecListSelectedAbsent is a thin alias so the no-selection assertion
// above stays on one line; it calls the same real builder.
func buildSpecListSelectedAbsent(a *app, query, selected string) SpecListVM {
	vm, err := a.buildSpecList(query, "", selected)
	if err != nil {
		return SpecListVM{}
	}
	return vm
}
