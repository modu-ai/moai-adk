package importer

// Defect answer: these tests catch a dry-run that writes to a source, that loses a row without saying so, that
// maps a state to the wrong stage (a dropped card becoming claimable, a held one escaping its hold, a
// finished one coming back), that lets a picked card sort behind a queued one, and that reads a factory.db
// older than the current schema as an error. Fixtures are real SQLite files with the old ledgers' tables.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func mkdb(t *testing.T, name, ddl string, stmts ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(ddl); err != nil {
		t.Fatal(err)
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	return path
}

const backlogDDL = `CREATE TABLE items (
  seq INTEGER PRIMARY KEY, id TEXT NOT NULL UNIQUE, text TEXT NOT NULL, added_at TEXT NOT NULL, spec_id TEXT,
  state TEXT NOT NULL CHECK (state IN ('queued','picked','dropped','hold')), landing TEXT)`

const factoryDDL = `CREATE TABLE cards (
  run_id TEXT NOT NULL, card_id TEXT NOT NULL, owner_label TEXT NOT NULL DEFAULT '', state TEXT NOT NULL,
  version INTEGER NOT NULL DEFAULT 1, evidence_path TEXT NOT NULL DEFAULT '', updated_at TEXT NOT NULL,
  stage TEXT NOT NULL DEFAULT '', spec_id TEXT NOT NULL DEFAULT '', worktree_path TEXT NOT NULL DEFAULT '',
  PRIMARY KEY(run_id, card_id))`

func item(seq int, id, text, state, spec string) string {
	if spec == "" {
		spec = "NULL"
	} else {
		spec = "'" + spec + "'"
	}
	return fmt.Sprintf("INSERT INTO items(seq,id,text,added_at,spec_id,state) VALUES(%d,'%s','%s','2026-10-01T00:00:00Z',%s,'%s')", seq, id, text, spec, state)
}

func byID(p Plan) map[string]Proposal {
	m := map[string]Proposal{}
	for _, pr := range p.Proposals {
		m[pr.Card.ID] = pr
	}
	return m
}

func sum(t *testing.T, path string) [32]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(b)
}

func TestDry_MapsEveryBacklogStateAndCountsThem(t *testing.T) {
	bl := mkdb(t, "backlog.db", backlogDDL,
		item(1, "t1", "first queued", "queued", ""),
		item(2, "t2", "approved", "picked", "SPEC-X-001"),
		item(3, "t3", "on hold", "hold", ""),
		item(4, "t4", "dropped", "dropped", ""),
		item(5, "t5", "second queued", "queued", ""),
		item(6, "t6", "second approved", "picked", ""),
	)
	plan, err := Dry(context.Background(), Sources{BacklogDB: bl})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Proposals) != 6 || len(plan.Skips) != 0 {
		t.Fatalf("proposals %d skips %d, want 6 and 0: %+v", len(plan.Proposals), len(plan.Skips), plan.Skips)
	}
	for k, want := range map[string]int{"backlog:queued": 2, "backlog:picked": 2, "backlog:hold": 1, "backlog:dropped": 1} {
		if plan.Counts[k] != want {
			t.Errorf("count %s = %d, want %d", k, plan.Counts[k], want)
		}
	}
	m := byID(plan)
	if got := m["t1"].Card.Position; got.Stage != core.StageReady || got.Park != nil {
		t.Errorf("queued -> %+v, want READY unparked", got)
	}
	if got := m["t3"].Card.Position; got.Stage != core.StageReady || got.Park == nil || got.Park.Reason != core.ReasonOperatorHold {
		t.Errorf("hold -> %+v, want READY parked with operator-hold", got)
	}
	if got := m["t4"].Card.Position.Stage; got != core.StageCancelled {
		t.Errorf("dropped -> %s, want CANCELLED (never claimable)", got)
	}
	if !strings.Contains(m["t2"].Card.Title, "SPEC-X-001") {
		t.Errorf("the spec id was lost: title %q", m["t2"].Card.Title)
	}
	for id, pr := range m {
		if pr.Card.Spec.Class != DefaultClass {
			t.Errorf("%s class %q, want the careful default %s", id, pr.Card.Spec.Class, DefaultClass)
		}
	}
}

func TestDry_PickedClaimsBeforeQueuedAndTheOldOrderStands(t *testing.T) {
	bl := mkdb(t, "backlog.db", backlogDDL,
		item(1, "q1", "queued early", "queued", ""),
		item(2, "p1", "picked early", "picked", ""),
		item(3, "q2", "queued late", "queued", ""),
		item(4, "p2", "picked late", "picked", ""),
	)
	plan, err := Dry(context.Background(), Sources{BacklogDB: bl})
	if err != nil {
		t.Fatal(err)
	}
	m := byID(plan)
	pr := func(id string) int { return m[id].Card.Priority }
	if !(pr("p1") > pr("p2") && pr("p2") > pr("q1") && pr("q1") > pr("q2")) {
		t.Errorf("priorities p1=%d p2=%d q1=%d q2=%d: want every picked above every queued and the earlier row above the later",
			pr("p1"), pr("p2"), pr("q1"), pr("q2"))
	}
}

func TestDry_NeverWritesToASource(t *testing.T) {
	bl := mkdb(t, "backlog.db", backlogDDL, item(1, "t1", "x", "queued", ""))
	fa := mkdb(t, "factory.db", factoryDDL, "INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('r1','t1','run','2026-10-02T00:00:00Z')")
	before := [2][32]byte{sum(t, bl), sum(t, fa)}
	if _, err := Dry(context.Background(), Sources{BacklogDB: bl, FactoryDB: fa}); err != nil {
		t.Fatal(err)
	}
	if after := [2][32]byte{sum(t, bl), sum(t, fa)}; after != before {
		t.Error("a source file changed during the dry-run")
	}
	// control: the connection really is read-only, so a write attempt through the same DSN fails.
	db, err := openRO(bl)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("UPDATE items SET text='y'"); err == nil {
		t.Error("control: the read-only connection accepted a write, so the check above proves nothing")
	}
}

func TestDry_FactoryRowsSettleOrRefineBacklogRows(t *testing.T) {
	bl := mkdb(t, "backlog.db", backlogDDL,
		item(1, "t1", "finished in the factory", "picked", ""),
		item(2, "t2", "in flight", "picked", ""),
		item(3, "t3", "abandoned", "queued", ""),
	)
	fa := mkdb(t, "factory.db", factoryDDL,
		"INSERT INTO cards(run_id,card_id,state,updated_at,worktree_path) VALUES('r1','t1','merged-local','2026-10-02T00:00:00Z','/wt/t1')",
		"INSERT INTO cards(run_id,card_id,state,updated_at,owner_label,stage,worktree_path) VALUES('r1','t2','run','2026-10-02T00:00:00Z','lane-5','run','/wt/t2')",
		"INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('r1','t3','abandoned','2026-10-02T00:00:00Z')",
		"INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('r1','t9','done','2026-10-02T00:00:00Z')", // only the factory knows it
	)
	plan, err := Dry(context.Background(), Sources{BacklogDB: bl, FactoryDB: fa})
	if err != nil {
		t.Fatal(err)
	}
	m := byID(plan)
	if m["t1"].Card.Position.Stage != core.StageDone || m["t1"].Source != "backlog+factory" {
		t.Errorf("t1 %+v: a card the factory merged must not come back as READY", m["t1"].Card.Position)
	}
	if m["t2"].Card.Position.Stage != core.StageReady || !strings.Contains(strings.Join(m["t2"].Notes, " "), "/wt/t2") {
		t.Errorf("t2 %+v notes %v: an in-flight card restarts at READY and says where its old tree is", m["t2"].Card.Position, m["t2"].Notes)
	}
	if m["t3"].Card.Position.Stage != core.StageCancelled {
		t.Errorf("t3 %+v, want CANCELLED", m["t3"].Card.Position)
	}
	if m["t9"].Card.Position.Stage != core.StageDone || m["t9"].Source != "factory" {
		t.Errorf("t9 %+v source %s, want DONE from the factory alone", m["t9"].Card.Position, m["t9"].Source)
	}
	if len(plan.Proposals) != 4 {
		t.Errorf("proposals %d, want 4 (each card id once)", len(plan.Proposals))
	}
}

func TestDry_TheLatestFactoryRowWinsForACardInSeveralRuns(t *testing.T) {
	fa := mkdb(t, "factory.db", factoryDDL,
		"INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('old','t1','run','2026-10-01T00:00:00Z')",
		"INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('new','t1','done','2026-10-05T00:00:00Z')",
	)
	plan, err := Dry(context.Background(), Sources{FactoryDB: fa})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Proposals) != 1 || plan.Proposals[0].Card.Position.Stage != core.StageDone {
		t.Errorf("proposals %+v, want one DONE card (the newer run)", plan.Proposals)
	}
}

func TestDry_UnknownStatesAreSkippedAndSaidNotDropped(t *testing.T) {
	fa := mkdb(t, "factory.db", factoryDDL,
		"INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('r','t1','teleported','2026-10-01T00:00:00Z')")
	plan, err := Dry(context.Background(), Sources{FactoryDB: fa})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Proposals) != 0 || len(plan.Skips) != 1 || !strings.Contains(plan.Skips[0].Reason, "teleported") {
		t.Errorf("proposals %d skips %+v: an unknown state must be reported, not silently lost", len(plan.Proposals), plan.Skips)
	}
	if plan.Counts["factory:teleported"] != 1 {
		t.Errorf("counts %v: the unknown state is still counted", plan.Counts)
	}
}

func TestDry_ReadsAFactoryTableThatLacksTheNewerColumns(t *testing.T) {
	old := `CREATE TABLE cards (run_id TEXT NOT NULL, card_id TEXT NOT NULL, state TEXT NOT NULL, updated_at TEXT NOT NULL, PRIMARY KEY(run_id, card_id))`
	fa := mkdb(t, "factory.db", old, "INSERT INTO cards VALUES('r','t1','abandoned','2026-10-01T00:00:00Z')")
	plan, err := Dry(context.Background(), Sources{FactoryDB: fa})
	if err != nil || len(plan.Proposals) != 1 {
		t.Fatalf("plan %+v err %v: a schema without stage/spec_id/worktree_path must still be readable", plan, err)
	}
}

// The operator's expected totals (101 picked + 64 queued held cards) are compared with what was read.
func TestPlanCheck_ComparesTheCountsWithWhatIsExpected(t *testing.T) {
	bl := mkdb(t, "backlog.db", backlogDDL,
		item(1, "a", "x", "picked", ""), item(2, "b", "y", "picked", ""), item(3, "c", "z", "queued", ""))
	plan, err := Dry(context.Background(), Sources{BacklogDB: bl})
	if err != nil {
		t.Fatal(err)
	}
	if bad := plan.Check(map[string]int{"backlog:picked": 2, "backlog:queued": 1}); len(bad) != 0 {
		t.Errorf("matching expectations reported %v", bad)
	}
	bad := plan.Check(map[string]int{"backlog:picked": 101, "backlog:queued": 64, "backlog:hold": 3})
	if len(bad) != 3 || !strings.Contains(bad[0], "found 0, expected 3") {
		t.Errorf("mismatches %v: want three lines, an absent state counted as zero", bad)
	}
	rep := plan.Report()
	for _, want := range []string{"dry-run", "backlog:picked", "READY", "Cards that would be created: 3", "class C"} {
		if !strings.Contains(rep, want) {
			t.Errorf("the report lacks %q:\n%s", want, rep)
		}
	}
}

// The shape of the operator's held set (101 picked + 64 queued = 165) at full size: the totals check passes,
// every card is READY and unparked, and every card id appears once.
func TestDry_TheHeldSetAtFullSize(t *testing.T) {
	var stmts []string
	for i := 1; i <= 165; i++ {
		state := "queued"
		if i <= 101 {
			state = "picked"
		}
		stmts = append(stmts, item(i, fmt.Sprintf("t%d", 1000+i), "held card", state, ""))
	}
	bl := mkdb(t, "backlog.db", backlogDDL, stmts...)
	plan, err := Dry(context.Background(), Sources{BacklogDB: bl})
	if err != nil {
		t.Fatal(err)
	}
	if bad := plan.Check(map[string]int{"backlog:picked": 101, "backlog:queued": 64}); len(bad) != 0 {
		t.Errorf("totals mismatch: %v", bad)
	}
	seen := map[string]bool{}
	for _, pr := range plan.Proposals {
		if seen[pr.Card.ID] {
			t.Errorf("card %s proposed twice", pr.Card.ID)
		}
		seen[pr.Card.ID] = true
		if pr.Card.Position.Stage != core.StageReady || pr.Card.Position.Park != nil {
			t.Errorf("%s -> %+v, want READY unparked", pr.Card.ID, pr.Card.Position)
		}
	}
	if len(plan.Proposals) != 165 {
		t.Errorf("proposals %d, want 165", len(plan.Proposals))
	}
}

// A card held by the global hold comes back as what it was; a hold row the checkpoint does not list stays parked.
func TestDry_TheHeldCheckpointRestoresOnlyTheCardsItLists(t *testing.T) {
	bl := mkdb(t, "backlog.db", backlogDDL,
		item(1, "t1", "was picked", "hold", ""),
		item(2, "t2", "was queued", "hold", ""),
		item(3, "t3", "held by the operator, not in the checkpoint", "hold", ""),
		item(4, "t4", "listed but not in hold any more", "dropped", ""),
	)
	plan, err := Dry(context.Background(), Sources{BacklogDB: bl, HeldIDs: map[string]string{"t1": "picked", "t2": "queued", "t4": "picked"}})
	if err != nil {
		t.Fatal(err)
	}
	m := byID(plan)
	if m["t1"].Card.Position.Park != nil || m["t1"].Card.Priority < priorityPickedBase {
		t.Errorf("t1 %+v priority %d: want READY unparked in the picked band", m["t1"].Card.Position, m["t1"].Card.Priority)
	}
	if m["t2"].Card.Position.Park != nil || m["t2"].Card.Priority >= priorityPickedBase {
		t.Errorf("t2 %+v priority %d: want READY unparked in the queued band", m["t2"].Card.Position, m["t2"].Card.Priority)
	}
	if m["t3"].Card.Position.Park == nil {
		t.Error("t3 is not in the checkpoint and must stay parked")
	}
	if m["t4"].Card.Position.Stage != core.StageCancelled {
		t.Errorf("t4 %+v: only a row still in hold is restored; a dropped one stays CANCELLED", m["t4"].Card.Position)
	}
	if plan.Counts["backlog:hold"] != 3 || plan.Counts["backlog:picked"] != 0 {
		t.Errorf("counts %v: the source states are counted as they are, not as restored", plan.Counts)
	}
}

// The operator picked or queued these at the checkpoint, so a factory "abandoned" does not cancel them (decision
// OD-36..42): they come back READY with a note and triage decides. Without the checkpoint, abandoned still cancels.
func TestDry_ACheckpointedCardTheFactoryAbandonedStaysReady(t *testing.T) {
	bl := mkdb(t, "backlog.db", backlogDDL,
		item(1, "t1", "picked at the checkpoint, abandoned by the factory", "hold", ""),
		item(2, "t2", "not in the checkpoint, abandoned by the factory", "hold", ""),
	)
	fa := mkdb(t, "factory.db", factoryDDL,
		"INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('r1','t1','abandoned','2026-10-02T00:00:00Z')",
		"INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('r1','t2','abandoned','2026-10-02T00:00:00Z')",
	)
	plan, err := Dry(context.Background(), Sources{BacklogDB: bl, FactoryDB: fa, HeldIDs: map[string]string{"t1": "picked"}})
	if err != nil {
		t.Fatal(err)
	}
	m := byID(plan)
	if m["t1"].Card.Position.Stage != core.StageReady || m["t1"].Card.Position.Park != nil {
		t.Errorf("t1 %+v, want READY unparked", m["t1"].Card.Position)
	}
	if !strings.Contains(strings.Join(m["t1"].Notes, " "), "triage decides") {
		t.Errorf("t1 notes %v: the override must be recorded", m["t1"].Notes)
	}
	if m["t2"].Card.Position.Stage != core.StageCancelled {
		t.Errorf("t2 %+v: a card outside the checkpoint stays CANCELLED (control)", m["t2"].Card.Position)
	}
}

func TestReadHeldIDs_RefusesAHalfReadCheckpoint(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	got, err := ReadHeldIDs(write("ok.txt", "t1 picked\nt2 queued\n"))
	if err != nil || got["t1"] != "picked" || got["t2"] != "queued" || len(got) != 2 {
		t.Fatalf("control: %v, %v", got, err)
	}
	for name, body := range map[string]string{
		"bad-state.txt": "t1 sleeping\n", "one-field.txt": "t1\n", "dup.txt": "t1 picked\nt1 queued\n", "three.txt": "t1 picked extra\n",
	} {
		if _, err := ReadHeldIDs(write(name, body)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestDry_ASourceWithoutItsTableIsAnErrorNotAnEmptyPlan(t *testing.T) {
	bl := mkdb(t, "backlog.db", `CREATE TABLE unrelated (x INTEGER)`)
	if _, err := Dry(context.Background(), Sources{BacklogDB: bl}); err == nil {
		t.Error("a file with no items table produced a plan; an empty plan would read as 'nothing to import'")
	}
	if _, err := Dry(context.Background(), Sources{BacklogDB: filepath.Join(t.TempDir(), "absent.db")}); err == nil {
		t.Error("a missing file produced a plan")
	}
}
