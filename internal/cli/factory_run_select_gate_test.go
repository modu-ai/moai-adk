package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/discovery"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// SPEC-LANE-JOIN-RUN-SELECT-001 (card t1853) gate tests: the lane-join gate's
// interactive chooser loop, driven through enterFactoryLaneRun with a scripted
// spy chooser standing in for the operator. No test needs a terminal.

// A dead-owner stamp no process can match: the pid exceeds every platform's
// pid_max and the fingerprint names a process that never ran, so the
// production batch probe classifies the run dead without an injected
// classifier.
const (
	selDeadPID   = 99999999
	selDeadStart = "gone-before-select"
)

// selSeedDead records an active dead-owner run. A worker row or a
// non-terminal card makes it a run the automatic reconciler PRESERVES (a
// records-free dead run is retired by the gate's own resolution before the
// chooser can open).
func selSeedDead(t *testing.T, root, runID string, workers bool, cards int) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: "claude", ManifestJSON: "{}",
		LeadPID: selDeadPID, LeadProcessStart: selDeadStart,
		LaneCapacity: homestate.LaneCapacityDerived,
	}); err != nil {
		t.Fatalf("seed dead run %s: %v", runID, err)
	}
	if workers {
		if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,run_id,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`,
			"w-"+runID, 1, runID, "2026-10-18T00:00:00Z", "2026-10-18T00:00:00Z"); err != nil {
			t.Fatalf("seed worker for %s: %v", runID, err)
		}
	}
	for i := 0; i < cards; i++ {
		if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,updated_at) VALUES(?,?,?,?)`,
			runID, fmt.Sprintf("card-%s-%d", runID, i), "run", fmt.Sprintf("2026-10-18T02:%02d:00Z", i)); err != nil {
			t.Fatalf("seed card for %s: %v", runID, err)
		}
	}
}

func selExec(t *testing.T, root, query string, args ...any) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func selQueryInt(t *testing.T, root, query string, args ...any) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return n
}

func selRunStatus(t *testing.T, root, runID string) string {
	t.Helper()
	status, _, _ := runRowOwner(t, root, runID)
	return status
}

func selRetiredEvents(t *testing.T, root, runID string) int {
	t.Helper()
	return selQueryInt(t, root, `SELECT COUNT(*) FROM events WHERE run_id=? AND kind='run.retired'`, runID)
}

// laneTables is a snapshot of the two tables the chooser must not write
// behind the operator's back.
type laneTables struct{ Runs, Events [][]string }

func selSnapshot(t *testing.T, root string) laneTables {
	t.Helper()
	return selSnapshotNoT(root)
}

// spyEntry is what the spy recorded on one presentation entry, BEFORE it
// returns its scripted choice.
type spyEntry struct {
	Env  string
	Snap laneTables
	Rows []laneRunRow
}

// spyChooser is the chooser double: on every entry it records the factory-run
// environment variable, a snapshot of the runs and events tables and the
// rows it was shown, then runs the step scripted for that entry. A step may
// change the world (adopt a run, retire one) AFTER the recording. An entry
// beyond the script cancels.
type spyChooser struct {
	root    string
	steps   []func(*spyEntry) runChoice
	entries []*spyEntry
}

func (s *spyChooser) ChooseLaneRun(rows []laneRunRow) (runChoice, error) {
	e := &spyEntry{Env: os.Getenv(config.EnvFactoryRunID), Rows: append([]laneRunRow(nil), rows...)}
	e.Snap = selSnapshotNoT(s.root)
	s.entries = append(s.entries, e)
	if len(s.entries) > len(s.steps) {
		return runChoice{Kind: choiceCancel}, nil
	}
	return s.steps[len(s.entries)-1](e), nil
}

// selSnapshotNoT snapshots without a *testing.T (the spy has none in hand
// when the gate calls it); a read failure yields an empty snapshot that no
// equality assertion can match, so it fails loudly there.
func selSnapshotNoT(root string) laneTables {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return laneTables{Runs: [][]string{{"snapshot open failed: " + err.Error()}}}
	}
	defer func() { _ = db.Close() }()
	dump := func(query string) [][]string {
		rows, err := db.DB.Query(query)
		if err != nil {
			return [][]string{{"snapshot query failed: " + err.Error()}}
		}
		defer func() { _ = rows.Close() }()
		cols, err := rows.Columns()
		if err != nil {
			return [][]string{{"snapshot columns failed: " + err.Error()}}
		}
		var out [][]string
		for rows.Next() {
			cells := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range cells {
				ptrs[i] = &cells[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				return [][]string{{"snapshot scan failed: " + err.Error()}}
			}
			line := make([]string, len(cols))
			for i, c := range cells {
				line[i] = fmt.Sprint(c)
			}
			out = append(out, line)
		}
		return out
	}
	return laneTables{
		Runs:   dump(`SELECT * FROM runs ORDER BY run_id`),
		Events: dump(`SELECT * FROM events ORDER BY seq`),
	}
}

// installLaneSeams makes the session interactive for this test, captures the
// loop's notices and installs the chooser. Everything is restored on cleanup
// (TestMain's pinned-false interactive seam comes back).
func installLaneSeams(t *testing.T, chooser runChooser) *bytes.Buffer {
	t.Helper()
	origInteractive, origChooser, origNotices := runSelectionInteractive, activeRunChooser, laneRunNotices
	t.Cleanup(func() {
		runSelectionInteractive, activeRunChooser, laneRunNotices = origInteractive, origChooser, origNotices
	})
	notices := &bytes.Buffer{}
	runSelectionInteractive = func() bool { return true }
	activeRunChooser = chooser
	laneRunNotices = notices
	return notices
}

func newSpy(root string, steps ...func(*spyEntry) runChoice) *spyChooser {
	return &spyChooser{root: root, steps: steps}
}

func choose(kind runChoiceKind, runID string) func(*spyEntry) runChoice {
	return func(*spyEntry) runChoice { return runChoice{Kind: kind, RunID: runID} }
}

// laneGateRoot is the sandboxed project plus the premise every test asserts
// before entering the gate: the join variable is empty, so a leaked ambient
// value fails the test instead of masking a premature set.
func laneGateRoot(t *testing.T) string {
	t.Helper()
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	if v := os.Getenv(config.EnvFactoryRunID); v != "" {
		t.Fatalf("%s = %q before entering the gate; the premise needs it empty", config.EnvFactoryRunID, v)
	}
	return root
}

func rowRunIDs(rows []laneRunRow) []string {
	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.RunID
	}
	return ids
}

func selAdoptToSelf(t *testing.T, root, runID string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	owner := homestate.RunOwner{RunID: runID, LeadPID: selDeadPID, LeadProcessStart: selDeadStart}
	if err := db.AdoptRun(context.Background(), owner, os.Getpid(), homestate.CurrentProcessFingerprint(), "test-adopt"); err != nil {
		t.Fatalf("adopt %s to this process: %v", runID, err)
	}
}

func selRetire(t *testing.T, root, runID string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.RetireRunIfDead(context.Background(), runID, factorymsg.ReconcileOptionsFor(root)); err != nil {
		t.Fatalf("retire %s: %v", runID, err)
	}
}

func requireEnv(t *testing.T, want string) {
	t.Helper()
	if got := os.Getenv(config.EnvFactoryRunID); got != want {
		t.Errorf("%s = %q, want %q", config.EnvFactoryRunID, got, want)
	}
}

func TestLaneJoinChooserJoinAndCancel(t *testing.T) {
	t.Run("join-sets-the-variable-only-after-the-choice", func(t *testing.T) {
		root := laneGateRoot(t)
		seedLiveRun(t, root, "rx-live")
		seedLiveRun(t, root, "ry-live")
		spy := newSpy(root, choose(choiceJoin, "ry-live"))
		installLaneSeams(t, spy)
		restore, err := enterFactoryLaneRun(root, "", "", nil)
		if err != nil {
			t.Fatalf("join: %v", err)
		}
		defer restore()
		requireEnv(t, "ry-live")
		if len(spy.entries) != 1 {
			t.Fatalf("spy entered %d times, want 1", len(spy.entries))
		}
		if spy.entries[0].Env != "" {
			t.Errorf("variable was %q at the presentation entry, want empty (set before the choice)", spy.entries[0].Env)
		}
		if got := selSnapshot(t, root); !reflect.DeepEqual(got, spy.entries[0].Snap) {
			t.Errorf("runs/events changed between the presentation entry and the join:\n entry %v\n after %v", spy.entries[0].Snap, got)
		}
	})

	cancels := map[string]runChooser{
		"scripted-cancel": newSpy("", choose(choiceCancel, "")),
		"empty-line":      textRunChooser{in: strings.NewReader("\n"), out: &bytes.Buffer{}},
		"end-of-input":    textRunChooser{in: strings.NewReader(""), out: &bytes.Buffer{}},
	}
	for name, chooser := range cancels {
		t.Run(name, func(t *testing.T) {
			root := laneGateRoot(t)
			seedLiveRun(t, root, "rx-live")
			seedLiveRun(t, root, "ry-live")
			if spy, ok := chooser.(*spyChooser); ok {
				spy.root = root
			}
			notices := installLaneSeams(t, chooser)
			_, err := enterFactoryLaneRun(root, "", "", nil)
			if err == nil || !strings.Contains(err.Error(), joinCancelledSentinel) {
				t.Fatalf("err = %v, want %s", err, joinCancelledSentinel)
			}
			requireEnv(t, "")
			if !strings.Contains(notices.String(), laneMsgCancelled) {
				t.Errorf("notices %q lack the cancel confirmation", notices.String())
			}
		})
	}

	t.Run("cancel-keeps-the-reconciliation-and-adds-nothing", func(t *testing.T) {
		root := laneGateRoot(t)
		seedLiveRun(t, root, "rx-live")
		seedLiveRun(t, root, "ry-live")
		selSeedDead(t, root, "rz-dead", false, 0) // records-free: the gate's own resolution retires it
		spy := newSpy(root, choose(choiceCancel, ""))
		installLaneSeams(t, spy)
		_, err := enterFactoryLaneRun(root, "", "", nil)
		if err == nil || !strings.Contains(err.Error(), joinCancelledSentinel) {
			t.Fatalf("err = %v, want %s", err, joinCancelledSentinel)
		}
		if got := selRunStatus(t, root, "rz-dead"); got != "retired" {
			t.Errorf("rz-dead status = %q, want retired (the reconciliation precedes the chooser and persists)", got)
		}
		if n := selRetiredEvents(t, root, "rz-dead"); n != 1 {
			t.Errorf("rz-dead has %d run.retired events, want exactly the reconciler's one", n)
		}
		if ids := rowRunIDs(spy.entries[0].Rows); !reflect.DeepEqual(ids, []string{"rx-live", "ry-live"}) {
			t.Errorf("rows = %v, want the two live runs only", ids)
		}
		if got := selSnapshot(t, root); !reflect.DeepEqual(got, spy.entries[0].Snap) {
			t.Errorf("the cancel path wrote to runs/events after the presentation entry:\n entry %v\n after %v", spy.entries[0].Snap, got)
		}
	})

	t.Run("join-of-a-vanished-run-returns-the-explicit-join-error", func(t *testing.T) {
		root := laneGateRoot(t)
		seedLiveRun(t, root, "rx-live")
		seedLiveRun(t, root, "ry-live")
		spy := newSpy(root, func(*spyEntry) runChoice {
			selExec(t, root, `UPDATE runs SET status='retired' WHERE run_id='ry-live'`)
			return runChoice{Kind: choiceJoin, RunID: "ry-live"}
		})
		installLaneSeams(t, spy)
		_, err := enterFactoryLaneRun(root, "", "", nil)
		if err == nil || err.Error() != noActiveFactorySentinel {
			t.Fatalf("err = %v, want the explicit-join answer %s", err, noActiveFactorySentinel)
		}
		if len(spy.entries) != 1 {
			t.Errorf("spy entered %d times, want 1 (no re-presentation after a failed join)", len(spy.entries))
		}
		requireEnv(t, "")
	})
}

// AC-LJR-006: cleanup-then-join retires through RetireRunIfDead only, then
// re-reads; the join half fires when one run remains; a refusal is printed
// and the facts re-read; a live owner is never retired; a run that became
// live between render and choice shows live on the next render.
func TestLaneJoinChooserRetireThenJoin(t *testing.T) {
	t.Run("cleanup-then-join-after-a-successful-cleanup", func(t *testing.T) {
		root := laneGateRoot(t)
		selSeedDead(t, root, "ra-dead", true, 0)
		seedLiveRun(t, root, "rb-live")
		selSeedDead(t, root, "rc-dead", true, 0)
		spy := newSpy(root,
			func(e *spyEntry) runChoice {
				requireEnv(t, "")
				if ids := rowRunIDs(e.Rows); !reflect.DeepEqual(ids, []string{"ra-dead", "rb-live", "rc-dead"}) {
					t.Errorf("first rows = %v", ids)
				}
				return runChoice{Kind: choiceCleanupThenJoin, RunID: "ra-dead"}
			},
			func(e *spyEntry) runChoice {
				if ids := rowRunIDs(e.Rows); !reflect.DeepEqual(ids, []string{"rb-live", "rc-dead"}) {
					t.Errorf("second rows = %v, want the re-read remainder [rb-live rc-dead]", ids)
				}
				return runChoice{Kind: choiceCleanupThenJoin, RunID: "rc-dead"}
			},
		)
		installLaneSeams(t, spy)
		restore, err := enterFactoryLaneRun(root, "", "", nil)
		if err != nil {
			t.Fatalf("gate: %v", err)
		}
		requireEnv(t, "rb-live")
		if len(spy.entries) != 2 {
			t.Errorf("spy entered %d times, want 2", len(spy.entries))
		}
		for _, id := range []string{"ra-dead", "rc-dead"} {
			if got := selRunStatus(t, root, id); got != "retired" {
				t.Errorf("%s status = %q, want retired", id, got)
			}
			if n := selRetiredEvents(t, root, id); n != 1 {
				t.Errorf("%s has %d run.retired events, want exactly 1", id, n)
			}
		}
		payload := ""
		db, derr := homestate.OpenFactory(root)
		if derr != nil {
			t.Fatal(derr)
		}
		if qerr := db.DB.QueryRow(`SELECT payload_json FROM events WHERE run_id='ra-dead' AND kind='run.retired'`).Scan(&payload); qerr != nil {
			t.Fatal(qerr)
		}
		_ = db.Close()
		if !strings.Contains(payload, `"classification":"dead"`) {
			t.Errorf("run.retired payload %q lacks the dead classification", payload)
		}
		if got := selRunStatus(t, root, "rb-live"); got != "active" {
			t.Errorf("rb-live status = %q, want active", got)
		}
		restore()
		requireEnv(t, "")
	})

	t.Run("cancel-after-a-completed-cleanup-adds-nothing", func(t *testing.T) {
		root := laneGateRoot(t)
		selSeedDead(t, root, "ra-dead", true, 0)
		seedLiveRun(t, root, "rb-live")
		selSeedDead(t, root, "rc-dead", true, 0)
		spy := newSpy(root, choose(choiceCleanupThenJoin, "ra-dead"), choose(choiceCancel, ""))
		installLaneSeams(t, spy)
		_, err := enterFactoryLaneRun(root, "", "", nil)
		if err == nil || !strings.Contains(err.Error(), joinCancelledSentinel) {
			t.Fatalf("err = %v, want %s", err, joinCancelledSentinel)
		}
		if got := selRunStatus(t, root, "ra-dead"); got != "retired" {
			t.Errorf("ra-dead status = %q, want retired (the completed cleanup persists)", got)
		}
		if n := selRetiredEvents(t, root, "ra-dead"); n != 1 {
			t.Errorf("ra-dead has %d run.retired events, want 1", n)
		}
		if len(spy.entries) != 2 {
			t.Fatalf("spy entered %d times, want 2", len(spy.entries))
		}
		if got := selSnapshot(t, root); !reflect.DeepEqual(got, spy.entries[1].Snap) {
			t.Errorf("the cancel path wrote after the second presentation entry:\n entry %v\n after %v", spy.entries[1].Snap, got)
		}
		requireEnv(t, "")
	})

	t.Run("a-chooser-cannot-retire-a-live-owner", func(t *testing.T) {
		root := laneGateRoot(t)
		seedLiveRun(t, root, "rl-live")
		seedLiveRun(t, root, "rm-live")
		if err := retireChosenRun(root, "rl-live"); !errors.Is(err, homestate.ErrRunOwnerNotDead) {
			t.Fatalf("retireChosenRun(live) = %v, want it to wrap ErrRunOwnerNotDead", err)
		}
		// A fake chooser bypasses the parser's eligibility check.
		spy := newSpy(root, choose(choiceCleanupThenJoin, "rl-live"), choose(choiceCancel, ""))
		notices := installLaneSeams(t, spy)
		_, err := enterFactoryLaneRun(root, "", "", nil)
		if err == nil || !strings.Contains(err.Error(), joinCancelledSentinel) {
			t.Fatalf("err = %v, want %s", err, joinCancelledSentinel)
		}
		if got := selRunStatus(t, root, "rl-live"); got != "active" {
			t.Errorf("rl-live status = %q, want active", got)
		}
		if n := selRetiredEvents(t, root, "rl-live"); n != 0 {
			t.Errorf("rl-live has %d run.retired events, want 0", n)
		}
		if !strings.Contains(notices.String(), fmt.Sprintf("run %s 정리에 실패했습니다", "rl-live")) {
			t.Errorf("notices %q lack the printed refusal", notices.String())
		}
		if len(spy.entries) != 2 {
			t.Fatalf("spy entered %d times, want 2 (presented again after the refusal)", len(spy.entries))
		}
		if ids := rowRunIDs(spy.entries[1].Rows); !reflect.DeepEqual(ids, []string{"rl-live", "rm-live"}) {
			t.Errorf("second rows = %v, want rl-live still listed", ids)
		}
	})

	t.Run("a-run-that-became-live-shows-live-on-the-next-render", func(t *testing.T) {
		root := laneGateRoot(t)
		selSeedDead(t, root, "rd-dead", true, 0)
		selSeedDead(t, root, "re-dead", true, 0)
		retiredBefore := -1
		spy := newSpy(root,
			func(e *spyEntry) runChoice {
				// After the first render and before the choice returns: a
				// competing write adopts the run to a live identity.
				selAdoptToSelf(t, root, "rd-dead")
				retiredBefore = selRetiredEvents(t, root, "rd-dead")
				return runChoice{Kind: choiceCleanupThenJoin, RunID: "rd-dead"}
			},
			choose(choiceCancel, ""),
		)
		notices := installLaneSeams(t, spy)
		_, err := enterFactoryLaneRun(root, "", "", nil)
		if err == nil || !strings.Contains(err.Error(), joinCancelledSentinel) {
			t.Fatalf("err = %v, want %s", err, joinCancelledSentinel)
		}
		if !strings.Contains(notices.String(), fmt.Sprintf("run %s 정리에 실패했습니다", "rd-dead")) {
			t.Errorf("notices %q lack the printed refusal", notices.String())
		}
		if got := selRunStatus(t, root, "rd-dead"); got != "active" {
			t.Errorf("rd-dead status = %q, want active", got)
		}
		if n := selRetiredEvents(t, root, "rd-dead"); n != retiredBefore {
			t.Errorf("rd-dead gained run.retired events after the callback: %d -> %d", retiredBefore, n)
		}
		if len(spy.entries) != 2 {
			t.Fatalf("spy entered %d times, want 2", len(spy.entries))
		}
		first, second := spy.entries[0].Rows[0], spy.entries[1].Rows[0]
		if first.RunID != "rd-dead" || first.Liveness != laneLivenessDead {
			t.Errorf("first render row = %+v, want rd-dead dead", first)
		}
		if second.RunID != "rd-dead" || second.Liveness != laneLivenessLive {
			t.Errorf("second render row = %+v, want rd-dead live (a cell that exists only if the facts were re-read)", second)
		}
	})

	// DEBT-2 of the plan audit: when the cleanup leaves no active run the
	// gate fails with NO_ACTIVE_FACTORY and joins nothing.
	t.Run("cleanup-leaving-no-run-fails-no-active-factory", func(t *testing.T) {
		root := laneGateRoot(t)
		selSeedDead(t, root, "rp-dead", true, 0)
		selSeedDead(t, root, "rq-dead", true, 0)
		spy := newSpy(root, func(*spyEntry) runChoice {
			// After the chooser opened, the only other run is retired.
			selRetire(t, root, "rq-dead")
			return runChoice{Kind: choiceCleanupThenJoin, RunID: "rp-dead"}
		})
		installLaneSeams(t, spy)
		_, err := enterFactoryLaneRun(root, "", "", nil)
		if err == nil || err.Error() != noActiveFactorySentinel {
			t.Fatalf("err = %v, want %s", err, noActiveFactorySentinel)
		}
		if got := selRunStatus(t, root, "rp-dead"); got != "retired" {
			t.Errorf("rp-dead status = %q, want retired (the cleanup succeeded)", got)
		}
		if len(spy.entries) != 1 {
			t.Errorf("spy entered %d times, want 1", len(spy.entries))
		}
		requireEnv(t, "")
	})
}

// refusedRetireRun stages the N2 fixture: dead-owner runs G and H (each with a
// worker row, no cards). The first presentation's callback adopts G to this
// process's live identity and retires H, so G is the only active run when
// the operator's cleanup of G is refused; with withZ it also inserts a
// records-free dead run Z. secondStep scripts the second presentation.
// refusedRetireResult is what refusedRetireRun observed.
type refusedRetireResult struct {
	spy      *spyChooser
	notices  *bytes.Buffer
	firstEnd laneTables // snapshot at the end of the first callback
	restore  func()
	err      error
	root     string
}

func refusedRetireRun(t *testing.T, withZ bool, secondStep func(*spyEntry) runChoice) refusedRetireResult {
	t.Helper()
	res := refusedRetireResult{root: laneGateRoot(t)}
	root := res.root
	selSeedDead(t, root, "rg-dead", true, 0)
	selSeedDead(t, root, "rh-dead", true, 0)
	res.spy = newSpy(root,
		func(e *spyEntry) runChoice {
			requireEnv(t, "")
			selAdoptToSelf(t, root, "rg-dead")
			selRetire(t, root, "rh-dead")
			if withZ {
				selSeedDead(t, root, "rz-dead", false, 0)
			}
			res.firstEnd = selSnapshot(t, root)
			return runChoice{Kind: choiceCleanupThenJoin, RunID: "rg-dead"}
		},
		secondStep,
	)
	res.notices = installLaneSeams(t, res.spy)
	res.restore, res.err = enterFactoryLaneRun(root, "", "", nil)
	return res
}

// AC-LJR-007 pins the N2 and N1 defects by the environment observed at every
// presentation entry: a refused cleanup that leaves one run presents it
// again (join/cancel only), never joins it on its own, and the re-read does
// not reconcile.
func TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining(t *testing.T) {
	check := func(t *testing.T, spy *spyChooser, notices *bytes.Buffer, firstEnd laneTables) *spyEntry {
		t.Helper()
		if !strings.Contains(notices.String(), fmt.Sprintf("run %s 정리에 실패했습니다", "rg-dead")) {
			t.Errorf("notices %q lack the printed refusal of the cleanup of rg-dead", notices.String())
		}
		if len(spy.entries) != 2 {
			t.Fatalf("spy entered %d times, want 2 (the gate must present again, not join)", len(spy.entries))
		}
		second := spy.entries[1]
		if second.Env != "" {
			t.Errorf("%s = %q at the second presentation entry, want empty (the gate joined without the operator)", config.EnvFactoryRunID, second.Env)
		}
		if !reflect.DeepEqual(second.Snap, firstEnd) {
			t.Errorf("runs/events changed between the end of the first callback and the second entry (the refused cleanup or the re-read wrote):\n first-end %v\n second    %v", firstEnd, second.Snap)
		}
		return second
	}

	t.Run("join-after-the-refusal", func(t *testing.T) {
		res := refusedRetireRun(t, false, choose(choiceJoin, "rg-dead"))
		if res.err != nil {
			t.Fatalf("gate: %v", res.err)
		}
		defer res.restore()
		second := check(t, res.spy, res.notices, res.firstEnd)
		if len(second.Rows) != 1 || second.Rows[0].RunID != "rg-dead" || second.Rows[0].Liveness != laneLivenessLive || second.Rows[0].CanCleanup {
			t.Errorf("second rows = %+v, want exactly rg-dead, live, no cleanup offered", second.Rows)
		}
		requireEnv(t, "rg-dead")
	})

	t.Run("cancel-after-the-refusal", func(t *testing.T) {
		res := refusedRetireRun(t, false, choose(choiceCancel, ""))
		if res.err == nil || !strings.Contains(res.err.Error(), joinCancelledSentinel) {
			t.Fatalf("err = %v, want %s", res.err, joinCancelledSentinel)
		}
		check(t, res.spy, res.notices, res.firstEnd)
		requireEnv(t, "")
	})

	// Sub-scenario (e): the re-read does not reconcile. A records-free dead
	// run that appeared between the two presentations is listed, dead, with
	// cleanup offered; it is not retired behind the operator's back.
	t.Run("the-reread-does-not-reconcile", func(t *testing.T) {
		res := refusedRetireRun(t, true, choose(choiceCancel, ""))
		if res.err == nil || !strings.Contains(res.err.Error(), joinCancelledSentinel) {
			t.Fatalf("err = %v, want %s", res.err, joinCancelledSentinel)
		}
		second := check(t, res.spy, res.notices, res.firstEnd)
		root := res.root
		if ids := rowRunIDs(second.Rows); !reflect.DeepEqual(ids, []string{"rg-dead", "rz-dead"}) {
			t.Fatalf("second rows = %v, want [rg-dead rz-dead]", ids)
		}
		z := second.Rows[1]
		if z.Liveness != laneLivenessDead || z.InProgress != 0 || !z.CanCleanup {
			t.Errorf("rz-dead row = %+v, want dead, 0 cards, cleanup offered", z)
		}
		if got := selRunStatus(t, root, "rz-dead"); got != "active" {
			t.Errorf("rz-dead status = %q, want active (the read path must not reconcile)", got)
		}
		if n := selRetiredEvents(t, root, "rz-dead"); n != 0 {
			t.Errorf("rz-dead has %d run.retired events, want 0", n)
		}
	})
}

var kstShape = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} KST`)

// AC-LJR-008 (non-interactive half): off a terminal the gate keeps its
// fail-closed guided refusal, never consults the chooser, and renders each
// start in KST.
func TestLaneJoinNonTTYKeepsGuidedRefusalInKST(t *testing.T) {
	root := laneGateRoot(t)
	seedLiveRun(t, root, "ambigruna")
	seedLiveRun(t, root, "ambigrunb")
	selExec(t, root, `UPDATE runs SET created_at=? WHERE run_id=?`, "2026-10-17T09:31:00.5Z", "ambigruna")
	selExec(t, root, `UPDATE runs SET created_at=? WHERE run_id=?`, "2026-10-17T10:00:00Z", "ambigrunb")
	spy := newSpy(root, func(*spyEntry) runChoice {
		t.Error("the chooser was consulted on a non-interactive session")
		return runChoice{Kind: choiceCancel}
	})
	installLaneSeams(t, spy)
	runSelectionInteractive = func() bool { return false }

	_, err := enterFactoryLaneRun(root, "", "", nil)
	if err == nil {
		t.Fatal("non-interactive join among two live runs succeeded, want the ambiguity refusal")
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "AMBIGUOUS_FACTORY:") {
		t.Errorf("refusal %q does not begin with AMBIGUOUS_FACTORY:", msg)
	}
	for _, want := range []string{"ambigruna", "ambigrunb", "--factory-run", "--leader", "--retire",
		"started 2026-10-17 18:31:00 KST", "started 2026-10-17 19:00:00 KST"} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal %q misses %q", msg, want)
		}
	}
	if got := len(kstShape.FindAllString(msg, -1)); got != 2 {
		t.Errorf("refusal carries %d KST-shaped times, want 2: %q", got, msg)
	}
	for _, raw := range []string{"T09:31:00", "T10:00:00", "09:31:00.5Z"} {
		if strings.Contains(msg, raw) {
			t.Errorf("refusal %q still carries the raw UTC form %q", msg, raw)
		}
	}
	if len(spy.entries) != 0 {
		t.Errorf("chooser entered %d times, want 0", len(spy.entries))
	}
	requireEnv(t, "")
}

// AC-LJR-008 (selector half): an explicit run and a named leader are
// decisions; the chooser is never shown for them, interactive or not.
func TestLaneJoinInteractiveSkipsChooserForExplicitSelectors(t *testing.T) {
	forbidden := func(root string) *spyChooser {
		return newSpy(root, func(*spyEntry) runChoice { return runChoice{Kind: choiceCancel} })
	}
	t.Run("explicit-run", func(t *testing.T) {
		root := laneGateRoot(t)
		seedLiveRun(t, root, "ambigruna")
		seedLiveRun(t, root, "ambigrunb")
		spy := forbidden(root)
		installLaneSeams(t, spy)
		restore, err := enterFactoryLaneRun(root, "ambigrunb", "", nil)
		if err != nil {
			t.Fatalf("explicit join: %v", err)
		}
		defer restore()
		requireEnv(t, "ambigrunb")
		if len(spy.entries) != 0 {
			t.Errorf("chooser entered %d times for an explicit run, want 0", len(spy.entries))
		}
	})
	t.Run("named-leader", func(t *testing.T) {
		root := laneGateRoot(t)
		seedLiveRun(t, root, "ambigruna")
		seedLiveRun(t, root, "ambigrunb")
		stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("ambigruna")})
		spy := forbidden(root)
		installLaneSeams(t, spy)
		restore, err := enterFactoryLaneRun(root, "", "leader", nil)
		if err != nil {
			t.Fatalf("named-leader join: %v", err)
		}
		defer restore()
		requireEnv(t, "ambigruna")
		if len(spy.entries) != 0 {
			t.Errorf("chooser entered %d times for a named leader, want 0", len(spy.entries))
		}
	})
	t.Run("named-leader-nobody-verified-keeps-the-refusal", func(t *testing.T) {
		root := laneGateRoot(t)
		seedLiveRun(t, root, "ambigruna")
		seedLiveRun(t, root, "ambigrunb")
		stageDiscoveredLeaders(t, nil)
		spy := forbidden(root)
		installLaneSeams(t, spy)
		_, err := enterFactoryLaneRun(root, "", "leader-nowhere", nil)
		if err == nil || !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") {
			t.Fatalf("err = %v, want the ambiguity refusal to stand", err)
		}
		if len(spy.entries) != 0 {
			t.Errorf("chooser entered %d times for a named leader, want 0", len(spy.entries))
		}
		requireEnv(t, "")
	})
}
