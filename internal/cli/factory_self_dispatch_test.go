// factory_self_dispatch_test.go — SPEC-FACTORY-SELF-DISPATCH-001 M1 AC tests
// (card t1240): the lane predicate, `moai factory next` selection order and
// output, --wait, the parent-checkout refusal, the lane queue allowlist, and
// the lane `decide` refusal. AC-SD-008, -009, -010 (CLI half), -015, -016
// (CLI half), -023.
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (§B of acceptance.md); ./internal/cli runs only
// through the anchored -run selectors naming one of these tests.
package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// sdRegisterLane inserts a lane label into the factory workers roster, the
// way the launcher's ClaimFactoryLane does — the T3 lease guard reads it.
func sdRegisterLane(t *testing.T, root, label string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES(?,1,'2026-09-28T00:00:00Z','2026-09-28T00:00:00Z')`, label); err != nil {
		t.Fatalf("register lane %s: %v", label, err)
	}
}

// sdLaneEnv stamps the full lane environment (admission + label + backend).
func sdLaneEnv(t *testing.T, label, backend string) {
	t.Helper()
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	t.Setenv(config.EnvMoaiKanbanLabel, label)
	t.Setenv(config.EnvMoaiKanbanBackend, backend)
}

// sdClearLaneEnv removes every variable the lane predicates read, so a test
// can assert the no-lane-variable behaviour deterministically.
func sdClearLaneEnv(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvFactoryRole, "")
	t.Setenv(config.EnvMoaiKanbanLabel, "")
	t.Setenv(config.EnvMoaiKanbanBackend, "")
}

// sdQueueBytes snapshots the queue file the verbs share.
func sdQueueBytes(t *testing.T, store *kanban.BacklogStore) string {
	t.Helper()
	raw, err := os.ReadFile(store.EnginePath())
	if err != nil {
		t.Fatalf("read queue engine: %v", err)
	}
	return string(raw)
}

// sdExit3 asserts err carries the intentional exit code 3 (no card).
func sdExit3(t *testing.T, where string, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected an error, got nil", where)
	}
	code, ok := ResolveExitCode(err)
	if !ok || code != 3 {
		t.Errorf("%s: exit code = (%d, %v), want (3, true); err=%v", where, code, ok, err)
	}
}

// sdNextPRLine returns the single row `moai todo pr <id>` prints for cardID.
func sdNextPRLine(t *testing.T, cardID string) string {
	t.Helper()
	out, _, err := runTodo(t, "pr", cardID)
	if err != nil {
		t.Fatalf("todo pr %s: %v", cardID, err)
	}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(line, cardID+"\t") {
			return line
		}
	}
	t.Fatalf("todo pr %s printed no row for the card: %q", cardID, out)
	return ""
}

// sdAssertLeasedOutput checks the `next` output contract: the card id, its
// stage, its worktree name, and a PR/landed line equal to `moai todo pr`'s.
func sdAssertLeasedOutput(t *testing.T, out, cardID, stage, worktree string) {
	t.Helper()
	head := ""
	prLine := ""
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(line, cardID+" ") && head == "" {
			head = line
			continue
		}
		if strings.HasPrefix(line, cardID+"\t") {
			prLine = line
		}
	}
	if head == "" {
		t.Fatalf("next output carries no head line for %s: %q", cardID, out)
	}
	wantHead := cardID + " stage=" + stage + " worktree=" + worktree
	if head != wantHead {
		t.Errorf("next head line = %q, want %q", head, wantHead)
	}
	if want := sdNextPRLine(t, cardID); prLine != want {
		t.Errorf("next PR/landed line = %q, want the todo pr row %q", prLine, want)
	}
}

// AC-SD-008 — selection order, output, no-card status.
func TestSD_AC008_NextSelectionOrderAndOutput(t *testing.T) {
	t.Run("assigned-to-this-lane wins over unowned picked", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root,
			homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
			homestate.Card{CardID: "t2", State: homestate.CardPicked},
		)
		sdLaneEnv(t, "lane-1", "")
		// M3 (REQ-SD-011): a leased card with no recorded worktree gains one
		// through the materializer, anchored at the parent checkout — the
		// cwd `next` runs from, per REQ-SD-010.
		t.Chdir(root)
		out, _, err := runFactory(t, "next", "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v\nstderr: %s", err, out)
		}
		if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
			t.Fatalf("t1 = %s holder=%q, want leased/lane-1", c.State, c.LeaseHolder)
		}
		if c := fcCard(t, root, "t2"); c.State != homestate.CardPicked {
			t.Fatalf("t2 = %s, want still picked", c.State)
		}
		sdAssertLeasedOutput(t, out, "t1", homestate.CardRun, "t1")
	})

	t.Run("unowned picked wins over older queued", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStatePicked)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardPicked})
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out, _, err := runFactory(t, "next", "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v", err)
		}
		if c := fcCard(t, root, "t2"); c.State != homestate.CardLeased {
			t.Fatalf("t2 = %s, want leased", c.State)
		}
		if fcHasCard(t, root, "t1") {
			t.Fatalf("t1 gained a record row although a picked card existed")
		}
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if rec.Items[0].State != kanban.BacklogStateQueued {
			t.Errorf("queued card promoted although a picked card existed: %s", rec.Items[0].State)
		}
		sdAssertLeasedOutput(t, out, "t2", "-", "t2")
	})

	t.Run("card assigned to another lane is never taken; exit 3", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStatePicked)
		sdRegisterLane(t, root, "lane-1")
		sdRegisterLane(t, root, "lane-2")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-2"})
		sdLaneEnv(t, "lane-1", "")
		before := fcCard(t, root, "t1")
		out, _, err := runFactory(t, "next", "--run", fcRun)
		sdExit3(t, "other lane's card", err)
		if !strings.Contains(out, "no card") {
			t.Errorf("stdout = %q, want a no-card line", out)
		}
		if after := fcCard(t, root, "t1"); after.State != before.State || after.Version != before.Version {
			t.Errorf("card row changed: %s v%d → %s v%d", before.State, before.Version, after.State, after.Version)
		}
	})

	t.Run("oldest queued promoted and leased; newer stays queued", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStateQueued)
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out, _, err := runFactory(t, "next", "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v", err)
		}
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if rec.Items[0].State != kanban.BacklogStatePicked || rec.Items[1].State != kanban.BacklogStateQueued {
			t.Fatalf("queue = (%s, %s), want (picked, queued)", rec.Items[0].State, rec.Items[1].State)
		}
		if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
			t.Fatalf("t1 = %s holder=%q, want leased/lane-1", c.State, c.LeaseHolder)
		}
		sdAssertLeasedOutput(t, out, "t1", "-", "t1")
	})

	t.Run("empty queue prints no-card and exits 3", func(t *testing.T) {
		root, _ := fcFixture(t)
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		out, _, err := runFactory(t, "next", "--run", fcRun)
		sdExit3(t, "empty queue", err)
		if !strings.Contains(out, "no card") {
			t.Errorf("stdout = %q, want a no-card line", out)
		}
	})

	t.Run("two lanes concurrently: exactly one leases the picked card", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStatePicked)
		sdRegisterLane(t, root, "lane-1")
		sdRegisterLane(t, root, "lane-2")
		fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardPicked})
		// The lane label reaches the selection through the environment on the
		// CLI path, which is process-global — two goroutines cannot each hold
		// their own label there. The concurrent race the AC names is the
		// selection+lease contention, so the arms call the selection directly
		// with their lane labels (the CLI wrapper is covered by every other
		// arm of this test).
		var wg sync.WaitGroup
		for _, lane := range []string{"lane-1", "lane-2"} {
			wg.Add(1)
			go func(lane string) {
				defer wg.Done()
				_, _, _ = factoryNextLeaseOnce(context.Background(), root, fcRun, lane)
			}(lane)
		}
		wg.Wait()
		c := fcCard(t, root, "t2")
		if c.State != homestate.CardLeased || (c.LeaseHolder != "lane-1" && c.LeaseHolder != "lane-2") {
			t.Fatalf("t2 = %s holder=%q, want leased by exactly one lane", c.State, c.LeaseHolder)
		}
	})
}

// AC-SD-009 — --wait leases once a card appears, and exits 3 at the bound.
func TestSD_AC009_NextWaitLeasesOrTimesOut(t *testing.T) {
	t.Run("card queued after two intervals is leased", func(t *testing.T) {
		root, store := fcFixture(t)
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		sleeps := 0
		prevSleep := factoryNextWaitSleep
		factoryNextWaitSleep = func(time.Duration) {
			sleeps++
			if sleeps == 2 {
				if _, _, err := store.Add("arrived while waiting"); err != nil {
					t.Errorf("queue add during wait: %v", err)
				}
			}
		}
		t.Cleanup(func() { factoryNextWaitSleep = prevSleep })
		_, _, err := runFactory(t, "next", "--wait", "--run", fcRun)
		if err != nil {
			t.Fatalf("next --wait: %v", err)
		}
		if sleeps != 2 {
			t.Errorf("wait slept %d times, want 2", sleeps)
		}
		if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
			t.Fatalf("t1 = %s holder=%q, want leased/lane-1", c.State, c.LeaseHolder)
		}
	})

	t.Run("bound elapsing with no card exits 3", func(t *testing.T) {
		root, _ := fcFixture(t)
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		prevClock := factoryCardNow
		prevSleep := factoryNextWaitSleep
		clock := fcNow
		sleeps := 0
		factoryCardNow = func() time.Time { return clock }
		factoryNextWaitSleep = func(d time.Duration) {
			sleeps++
			clock = clock.Add(d)
		}
		t.Cleanup(func() {
			factoryCardNow = prevClock
			factoryNextWaitSleep = prevSleep
		})
		out, _, err := runFactory(t, "next", "--wait", "--wait-bound", "10s", "--run", fcRun)
		sdExit3(t, "wait bound", err)
		if sleeps != 2 {
			t.Errorf("wait slept %d times before the 10s bound with 5s interval, want 2", sleeps)
		}
		if !strings.Contains(out, "no card") {
			t.Errorf("stdout = %q, want a no-card line", out)
		}
	})
}

// AC-SD-010 — `next` only on the parent checkout (CLI half; the project_root
// variant lands with the MCP tools in M3 and consumes the same function).
func TestSD_AC010_NextRefusedOutsideParent(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
	sdLaneEnv(t, "lane-1", "")

	wt := filepath.Join(filepath.Dir(root), filepath.Base(root)+"-wt")
	fcGit(t, root, "worktree", "add", "-q", wt, "-b", "WT-fixture")

	primary, _, err := identifyPrimaryCheckout(wt)
	if err != nil {
		t.Fatalf("identify primary: %v", err)
	}
	queueBefore := sdQueueBytes(t, store)
	cardBefore := fcCard(t, root, "t1")

	t.Setenv(config.EnvClaudeProjectDir, wt)
	_, stderr, err := runFactory(t, "next", "--run", fcRun)
	if err == nil {
		t.Fatalf("next from a linked worktree: expected refusal, got success")
	}
	if !strings.Contains(stderr+err.Error(), primary) {
		t.Errorf("refusal does not name the parent checkout %s: %v / %s", primary, err, stderr)
	}
	if got := sdQueueBytes(t, store); got != queueBefore {
		t.Errorf("queue file changed on the refused next")
	}
	if after := fcCard(t, root, "t1"); after.State != cardBefore.State || after.Version != cardBefore.Version {
		t.Errorf("card row changed on the refused next: %s v%d → %s v%d", cardBefore.State, cardBefore.Version, after.State, after.Version)
	}

	// The reusable check the MCP tools will evaluate against project_root
	// (M3): the linked worktree is refused, the parent checkout is accepted.
	if err := factoryAssertParentCheckout(wt); err == nil || !strings.Contains(err.Error(), primary) {
		t.Errorf("factoryAssertParentCheckout(worktree) = %v, want a refusal naming %s", err, primary)
	}
	if err := factoryAssertParentCheckout(primary); err != nil {
		t.Errorf("factoryAssertParentCheckout(parent) = %v, want nil", err)
	}
}

// AC-SD-015 — lane predicate and queue allowlist: the walk.
func TestSD_AC015_LaneQueueAllowlistWalk(t *testing.T) {
	_, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStateQueued)
	sdLaneEnv(t, "lane-1", "")

	// The walked set is derived from the tree itself: a verb without test
	// args fails the test, so a subcommand added later is covered without
	// editing the walk.
	verbArgs := map[string][]string{
		"add":         {"lane card"},
		"list":        nil,
		"done":        {"t1"},
		"next":        {"1"},
		"unpick":      {"t1"},
		"edit":        {"1", "text"},
		"move":        {"1"},
		"drop":        {"1", "reason"},
		"undone":      {"t1"},
		"undrop":      {"1"},
		"analyze":     nil,
		"relate":      {"t1", "t2"},
		"unrelate":    {"1"},
		"why":         {"t1"},
		"pr":          nil,
		"landed":      {"t1"},
		"auto-done":   nil,
		"export-json": nil,
		"history":     nil,
		"triage":      {"t1"},
	}
	allow := map[string]bool{"list": true, "history": true, "why": true, "pr": true, "triage": true}

	todoRoot := newTodoCmd()
	walked := 0
	for _, sub := range todoRoot.Commands() {
		name := sub.Name()
		if name == "help" || name == "completion" || sub.Hidden {
			continue
		}
		args, known := verbArgs[name]
		if !known {
			t.Fatalf("verb %q has no test args in the walk — extend verbArgs", name)
		}
		walked++
		before := sdQueueBytes(t, store)
		_, stderr, err := runTodo(t, append([]string{name}, args...)...)
		if allow[name] {
			if err != nil {
				t.Errorf("allowlisted verb %q refused: %v / %s", name, err, stderr)
			}
			continue
		}
		if err == nil {
			t.Errorf("verb %q succeeded under lane refusal; want the lane-boundary refusal", name)
			continue
		}
		if !strings.Contains(err.Error(), "lane boundary") {
			t.Errorf("verb %q refused without the lane boundary: %v", name, err)
		}
		if got := sdQueueBytes(t, store); got != before {
			t.Errorf("verb %q refusal changed the queue bytes", name)
		}
	}
	if walked != len(verbArgs) {
		t.Errorf("walked %d verbs, verbArgs covers %d — the walk and the tree disagree", walked, len(verbArgs))
	}

	// The bare render is allowed; the parent's add fallthrough is not.
	if _, _, err := runTodo(t); err != nil {
		t.Errorf("bare todo refused under lane refusal: %v", err)
	}
	before := sdQueueBytes(t, store)
	if _, _, err := runTodo(t, "natural", "language", "card"); err == nil || !strings.Contains(err.Error(), "lane boundary") {
		t.Errorf("parent fallthrough add not refused: %v", err)
	}
	if got := sdQueueBytes(t, store); got != before {
		t.Errorf("parent fallthrough refusal changed the queue bytes")
	}

	// `todo next <n>` is explicitly refused (plan B6).
	before = sdQueueBytes(t, store)
	if _, _, err := runTodo(t, "next", "1"); err == nil || !strings.Contains(err.Error(), "lane boundary") {
		t.Errorf("todo next <n> not refused: %v", err)
	}
	if got := sdQueueBytes(t, store); got != before {
		t.Errorf("todo next refusal changed the queue bytes")
	}
}

// AC-SD-015 — label-only is not a lane; refusal holds on the label and the
// Codex backend value (the Codex MCP env carries no role marker).
func TestSD_AC015_LabelOnlyIsNotALane(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStateQueued)
	sdRegisterLane(t, root, "lane-1")

	// Label-only environment: the marker is unset.
	sdClearLaneEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "lane-1")
	for _, tc := range []struct {
		verb string
		args []string
	}{
		{"next", nil},
		{"stage", []string{"t1", "run"}},
		{"complete", []string{"t1"}},
	} {
		full := append([]string{tc.verb}, tc.args...)
		full = append(full, "--run", fcRun)
		_, _, err := runFactory(t, full...)
		if err == nil || !strings.Contains(err.Error(), "not a lane session") {
			t.Errorf("factory %s under label-only env: err = %v, want the not-a-lane refusal", tc.verb, err)
		}
	}
	before := sdQueueBytes(t, store)
	if _, _, err := runTodo(t, "add", "x"); err == nil || !strings.Contains(err.Error(), "lane boundary") {
		t.Errorf("todo add under label-only env: err = %v, want the lane-boundary refusal", err)
	}
	if got := sdQueueBytes(t, store); got != before {
		t.Errorf("label-only todo add refusal changed the queue bytes")
	}

	// Codex MCP environment: lane label + backend gpt, no role marker. The
	// todo_add MCP tool refusal is exercised when the tool lands (M3); the
	// CLI todo surface and the not-a-lane factory refusal are covered here.
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGPT)
	before = sdQueueBytes(t, store)
	if _, _, err := runTodo(t, "add", "x"); err == nil || !strings.Contains(err.Error(), "lane boundary") {
		t.Errorf("todo add under Codex MCP env: err = %v, want the lane-boundary refusal", err)
	}
	if got := sdQueueBytes(t, store); got != before {
		t.Errorf("Codex MCP todo add refusal changed the queue bytes")
	}
	if _, _, err := runFactory(t, "next", "--run", fcRun); err == nil || !strings.Contains(err.Error(), "not a lane session") {
		t.Errorf("factory next under Codex MCP env: err = %v, want the not-a-lane refusal", err)
	}
}

// AC-SD-016 — a lane cannot decide (CLI half; factory_decide lands with the
// MCP tools in M3).
func TestSD_AC016_LaneDecideRefused(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardKickoff, DecisionGate: homestate.DecisionGateKickoff})
	before := fcCard(t, root, "t1")
	eventsBefore := fcEventCount(t, root, "card.transition")

	sdLaneEnv(t, "lane-1", "")
	_, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun)
	if err == nil || !strings.Contains(err.Error(), "refused") {
		t.Fatalf("lane decide: err = %v, want a refusal", err)
	}
	if after := fcCard(t, root, "t1"); after.State != before.State || after.Version != before.Version {
		t.Errorf("card row changed on the refused decide: %s v%d → %s v%d", before.State, before.Version, after.State, after.Version)
	}
	if got := fcEventCount(t, root, "card.transition"); got != eventsBefore {
		t.Errorf("events changed on the refused decide: %d → %d", eventsBefore, got)
	}

	// None of the three lane variables: the same call succeeds.
	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun); err != nil {
		t.Fatalf("operator decide: %v", err)
	}
	if after := fcCard(t, root, "t1"); after.State != homestate.CardAssigned || after.Stage != homestate.CardRun {
		t.Errorf("card = %s stage=%s, want assigned/run", after.State, after.Stage)
	}
}

// AC-SD-023 — a Codex lane never re-leases a card it cannot advance.
func TestSD_AC023_CodexNextSkipsUnadvanceableCard(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	// A card returned to `assigned` by lease expiry, stage merge-ready.
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardMergeReady})

	sdLaneEnv(t, "lane-1", kanban.BackendGPT)
	before := fcCard(t, root, "t1")
	out, _, err := runFactory(t, "next", "--run", fcRun)
	sdExit3(t, "codex unadvanceable", err)
	if !strings.Contains(out, "no card") {
		t.Errorf("stdout = %q, want a no-card line", out)
	}
	if after := fcCard(t, root, "t1"); after.State != before.State || after.Version != before.Version || after.LeaseHolder != "" {
		t.Errorf("card row changed: %+v → %+v", before, after)
	}

	// A Claude lane owning the same card leases it (M3: and gains its
	// worktree, anchored at the parent checkout the CLI runs from).
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("claude next on merge-ready-returned card: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t1 = %s holder=%q, want leased/lane-1", c.State, c.LeaseHolder)
	}
}
