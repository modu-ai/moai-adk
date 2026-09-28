package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

const fcUnavailableTag = "FACTORY_RECORD_UNAVAILABLE"

var errFcInjected = errors.New("injected factory write error")

// fcFailWritesFor makes the factory-record write fail for the named cards
// (all cards when none is named) and restores the seam at cleanup.
func fcFailWritesFor(t *testing.T, cards ...string) func() {
	t.Helper()
	prev := factoryAssignmentWriter
	factoryAssignmentWriter = func(ctx context.Context, root string, store *kanban.BacklogStore, runID, cardID, lane string) error {
		if len(cards) == 0 {
			return errFcInjected
		}
		for _, c := range cards {
			if c == cardID {
				return errFcInjected
			}
		}
		return prev(ctx, root, store, runID, cardID, lane)
	}
	restore := func() { factoryAssignmentWriter = prev }
	t.Cleanup(restore)
	return restore
}

type fcLogEntry struct {
	RunID      string `json:"run_id"`
	CardID     string `json:"card_id"`
	Lane       string `json:"lane"`
	Error      string `json:"error"`
	Reconciled bool   `json:"reconciled"`
}

func fcUnavailableLog(t *testing.T, root string) []fcLogEntry {
	t.Helper()
	dir, err := homestate.FactoryDir(root)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(dir, "record-unavailable.jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	var out []fcLogEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) == "" {
			continue
		}
		var e fcLogEntry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatalf("log line %q: %v", sc.Text(), err)
		}
		out = append(out, e)
	}
	return out
}

func fcDriftEvents(t *testing.T, root, runID string) []map[string]any {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT payload_json FROM events WHERE kind='record.drift' AND run_id=? ORDER BY seq`, runID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var out []map[string]any
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(p), &m); err != nil {
			t.Fatal(err)
		}
		out = append(out, m)
	}
	return out
}

func fcCardIn(t *testing.T, root, runID, cardID string) (homestate.Card, bool) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	c, err := db.LoadCard(context.Background(), runID, cardID)
	if errors.Is(err, homestate.ErrCardNotFound) {
		return c, false
	}
	if err != nil {
		t.Fatal(err)
	}
	return c, true
}

// fcStatusReportsUnavailable reports whether `moai factory status` (text and
// JSON) lists an unreconciled record-unavailable entry for cardID.
func fcStatusReportsUnavailable(t *testing.T, runID, cardID string) (bool, bool) {
	t.Helper()
	text, _, err := runFactory(t, "status", "--run", runID)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	inText := false
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, fcUnavailableTag) && strings.Contains(line, "card="+cardID) {
			inText = true
		}
	}
	jsonOut, _, err := runFactory(t, "status", "--json", "--run", runID)
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	var st struct {
		Unavailable []struct {
			CardID string `json:"card_id"`
		} `json:"unavailable"`
	}
	if err := json.Unmarshal([]byte(jsonOut), &st); err != nil {
		t.Fatalf("status json: %v\n%s", err, jsonOut)
	}
	inJSON := false
	for _, e := range st.Unavailable {
		if e.CardID == cardID {
			inJSON = true
		}
	}
	return inText, inJSON
}

// AC-024 — the queue dispatch path mirrors the assignment into factory.db;
// a failed mirror write leaves the dispatch intact, is logged and reported,
// and is reconciled by exactly one record.drift event on the next write.
func TestFR_AC024_GTDDispatchMirrorsFactoryRecord(t *testing.T) {
	root, store := todoFixture(t)
	const run = "run-gtd"
	stdout, _, d1 := fcGTDDispatch(t, store, "mirror gtd dispatch one", "fr-gtd-1", "worker-2", run)
	if strings.TrimSpace(stdout) != fcGTDEngageOutput {
		t.Fatalf("engage output changed: %q", stdout)
	}
	if c, ok := fcCardIn(t, root, run, d1); !ok || c.State != homestate.CardAssigned || c.OwnerLabel != "worker-2" {
		t.Fatalf("factory record for %s = %+v (present=%v), want assigned to worker-2", d1, c, ok)
	}
	if got := fcRuntimeRows(t, store); len(got) != 1 || got[0] != (fcRuntimeRow{run, d1, "worker-2", "picked", "card.assigned"}) {
		t.Fatalf("runtime rows = %+v", got)
	}

	restore := fcFailWritesFor(t, "t2")
	_, stderr, d3 := fcGTDDispatch(t, store, "mirror gtd dispatch two", "fr-gtd-2", "worker-2", run)
	if d3 != "t2" {
		t.Fatalf("second dispatched card = %q, want t2", d3)
	}
	if !strings.Contains(stderr, fcUnavailableTag) {
		t.Fatalf("stderr lacks %s:\n%s", fcUnavailableTag, stderr)
	}
	rows := fcRuntimeRows(t, store)
	if len(rows) != 2 || rows[1] != (fcRuntimeRow{run, d3, "worker-2", "picked", "card.assigned"}) {
		t.Fatalf("runtime rows after failed mirror = %+v", rows)
	}
	log := fcUnavailableLog(t, root)
	if len(log) != 1 || log[0].CardID != d3 || log[0].Lane != "worker-2" || log[0].RunID != run || !strings.Contains(log[0].Error, errFcInjected.Error()) || log[0].Reconciled {
		t.Fatalf("unavailable log = %+v", log)
	}
	if inText, inJSON := fcStatusReportsUnavailable(t, run, d3); !inText || !inJSON {
		t.Fatalf("status does not report the unreconciled entry (text=%v json=%v)", inText, inJSON)
	}

	restore()
	_, _, d5 := fcGTDDispatch(t, store, "mirror gtd dispatch three", "fr-gtd-3", "worker-2", run)
	if _, ok := fcCardIn(t, root, run, d5); !ok {
		t.Fatalf("third dispatch %s not recorded", d5)
	}
	drift := fcDriftEvents(t, root, run)
	if len(drift) != 1 || drift[0]["card_id"] != d3 || drift[0]["lane"] != "worker-2" || drift[0]["factory_state"] != "absent" {
		t.Fatalf("record.drift events = %+v, want one naming %s / worker-2 / absent", drift, d3)
	}
	if log := fcUnavailableLog(t, root); len(log) != 1 || !log[0].Reconciled {
		t.Fatalf("log after reconciliation = %+v, want the entry reconciled", log)
	}
	if inText, inJSON := fcStatusReportsUnavailable(t, run, d3); inText || inJSON {
		t.Fatal("status still reports a reconciled entry")
	}
	fcGTDDispatch(t, store, "mirror gtd dispatch four", "fr-gtd-4", "worker-2", run)
	if drift := fcDriftEvents(t, root, run); len(drift) != 1 {
		t.Fatalf("a second successful write appended more drift events: %+v", drift)
	}
}

// AC-025 — the auto-mission dispatch path mirrors the assignment too, with
// the same fail-open, logging, and reconciliation behavior.
func TestFR_AC025_GoalDispatchMirrorsFactoryRecord(t *testing.T) {
	rootA, storeA := todoFixture(t)
	const sessionA = "018f4f4a-7b7c-7a11-8f4d-f33333333333"
	op, d2, _ := fcGoalDispatch(t, rootA, storeA, "mirror goal dispatch", "fr-goal-1", sessionA, "worker-3", "auto-run-3")
	if c, ok := fcCardIn(t, rootA, "auto-run-3", d2); !ok || c.State != homestate.CardAssigned || c.OwnerLabel != "worker-3" {
		t.Fatalf("factory record for %s = %+v (present=%v), want assigned to worker-3", d2, c, ok)
	}
	if got := fcRuntimeRows(t, storeA); len(got) != 1 || got[0] != (fcRuntimeRow{"auto-run-3", d2, "worker-3", "picked", "card.assigned"}) {
		t.Fatalf("runtime rows = %+v", got)
	}
	fcAssertGoalDispatchOp(t, op, sessionA)

	rootB, storeB := todoFixture(t)
	const sessionB = "018f4f4a-7b7c-7a11-8f4d-f44444444444"
	const runB = "auto-run-4"
	restore := fcFailWritesFor(t)
	opB, d4, stderr := fcGoalDispatch(t, rootB, storeB, "mirror goal dispatch failing", "fr-goal-2", sessionB, "worker-3", runB)
	restore()
	fcAssertGoalDispatchOp(t, opB, sessionB)
	if got := fcRuntimeRows(t, storeB); len(got) != 1 || got[0] != (fcRuntimeRow{runB, d4, "worker-3", "picked", "card.assigned"}) {
		t.Fatalf("runtime rows after failed mirror = %+v", got)
	}
	if !strings.Contains(stderr, fcUnavailableTag) {
		t.Fatalf("stderr lacks %s:\n%s", fcUnavailableTag, stderr)
	}
	log := fcUnavailableLog(t, rootB)
	if len(log) != 1 || log[0].CardID != d4 || log[0].Lane != "worker-3" || log[0].Reconciled {
		t.Fatalf("unavailable log = %+v", log)
	}
	// The next successful factory-record write for the run: record another
	// queue-picked card through the command.
	if _, _, err := runTodo(t, "add", "next write card"); err != nil {
		t.Fatal(err)
	}
	var next string
	if err := storeB.Mutate(func(r *kanban.BacklogRecord) error {
		last := &r.Items[len(r.Items)-1]
		last.State = kanban.BacklogStatePicked
		next = last.ID
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runFactory(t, "assign", next, "--run", runB); err != nil {
		t.Fatalf("assign %s: %v", next, err)
	}
	drift := fcDriftEvents(t, rootB, runB)
	if len(drift) != 1 || drift[0]["card_id"] != d4 || drift[0]["lane"] != "worker-3" || drift[0]["factory_state"] != "absent" {
		t.Fatalf("record.drift events = %+v, want one naming %s / worker-3 / absent", drift, d4)
	}
	if log := fcUnavailableLog(t, rootB); len(log) != 1 || !log[0].Reconciled {
		t.Fatalf("log after reconciliation = %+v", log)
	}
}
