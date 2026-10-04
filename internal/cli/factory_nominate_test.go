// factory_nominate_test.go — SPEC-TODO-AUTO-PICK-001 (card t1448) M1-M4
// acceptance tests: the nominated lease (`moai factory next --card <id>` and
// its MCP form), the shared keep-set predicate, the bare-path golden, the
// lane refusal of `moai todo --auto`, and the corrected fallback instruction.
//
// Every fixture is built under t.TempDir() with the lane environment scrubbed
// first (sdClearLaneEnv) — the queue is seeded through `todo add`, which a lane
// session is refused. The refusal exit code is asserted as the literal 4; the
// run phase introduces the named constant in M2.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/pflag"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/jev"
)

// ---------------------------------------------------------------------------
// fixtures and assertions
// ---------------------------------------------------------------------------

const (
	nmHoldMarker = "[보류"
	nmInjected   = "injected seam failure"
)

// nmBase builds a queue root holding one card per state (t1..tN), every card
// parallelizable and normal priority, with lanes lane-1 and lane-2 registered
// in the factory roster. The lane environment is scrubbed before the queue is
// seeded and left scrubbed; callers stamp a lane with nmLaneEnv.
func nmBase(t *testing.T, states ...factory.BacklogState) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := sdMoaiFixture(t)
	sdClearLaneEnv(t)
	fcQueue(t, store, states...)
	for i := range states {
		fcClassify(t, store, fmt.Sprintf("t%d", i+1), factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)
	return root, store
}

// nmLaneEnv stamps the full lane environment for label with the given backend.
func nmLaneEnv(t *testing.T, label, backend string) {
	t.Helper()
	sdLaneEnv(t, label, backend)
}

// nmSetText replaces one queue card's text.
func nmSetText(t *testing.T, store *factory.BacklogStore, cardID, text string) {
	t.Helper()
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == cardID {
				r.Items[i].Text = text
				return nil
			}
		}
		return fmt.Errorf("no card %s", cardID)
	}); err != nil {
		t.Fatalf("set text of %s: %v", cardID, err)
	}
}

// nmSetState moves one queue card to the named queue state.
func nmSetState(t *testing.T, store *factory.BacklogStore, cardID string, state factory.BacklogState) {
	t.Helper()
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == cardID {
				r.Items[i].State = state
				return nil
			}
		}
		return fmt.Errorf("no card %s", cardID)
	}); err != nil {
		t.Fatalf("set state of %s: %v", cardID, err)
	}
}

// nmQueueState reads one queue card's state.
func nmQueueState(t *testing.T, store *factory.BacklogStore, cardID string) factory.BacklogState {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID == cardID {
			return it.State
		}
	}
	t.Fatalf("card %s is not in the queue", cardID)
	return ""
}

// nmSnapshot is what "the queue and the factory record are unchanged" compares:
// the queue file bytes plus every card row and the event count.
func nmSnapshot(t *testing.T, root string, store *factory.BacklogStore) string {
	t.Helper()
	return sdQueueBytes(t, store) + "\n--- record ---\n" + qasRecordDump(t, root)
}

// nmRowCount counts the factory-record rows of one card across every run.
func nmRowCount(t *testing.T, root, cardID string) int {
	t.Helper()
	db := fcOpen(t, root)
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM cards WHERE card_id = ?`, cardID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// nmExit reports the exit code an error carries (-1 when it carries none).
func nmExit(err error) int {
	if err == nil {
		return 0
	}
	if code, ok := ResolveExitCode(err); ok {
		return code
	}
	return -1
}

// nmNonEmptyLines splits text into its non-empty lines.
func nmNonEmptyLines(text string) []string {
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// nmCardHeadRE matches the head line of a leased card's output.
var nmCardHeadRE = regexp.MustCompile(`^t[0-9]+ stage=`)

// nmLeasedHead returns the leased card's head line from `next` stdout; the
// worktree materializer writes its own lines to the same stream first.
func nmLeasedHead(out string) string {
	for _, l := range nmNonEmptyLines(out) {
		if nmCardHeadRE.MatchString(l) {
			return l
		}
	}
	return ""
}

// nmAssertRefused asserts the refusal contract of one nomination: exit status
// 4, one stderr line `factory next: refused <token>: <detail>`, nothing on
// stdout.
func nmAssertRefused(t *testing.T, out, stderr string, err error, token string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the nomination was not refused (want exit 4, token %q): stdout=%q stderr=%q", token, out, stderr)
	}
	if code := nmExit(err); code != 4 {
		t.Fatalf("exit code = %d, want 4 for refused %s; err=%v stderr=%q", code, token, err, stderr)
	}
	if out != "" {
		t.Errorf("a refused nomination wrote stdout %q, want nothing", out)
	}
	prefix := "factory next: refused " + token + ": "
	lines := nmNonEmptyLines(stderr)
	if len(lines) != 1 || !strings.HasPrefix(lines[0], prefix) || strings.TrimSpace(strings.TrimPrefix(lines[0], prefix)) == "" {
		t.Errorf("stderr = %q, want exactly one line %q<detail>", stderr, prefix)
	}
}

// nmAssertLeased asserts the nominee ended leased to lane in the record.
func nmAssertLeased(t *testing.T, root, cardID, lane string) {
	t.Helper()
	if !fcHasCard(t, root, cardID) {
		t.Fatalf("%s has no record row, want a lease held by %s", cardID, lane)
	}
	if c := fcCard(t, root, cardID); c.State != homestate.CardLeased || c.LeaseHolder != lane {
		t.Fatalf("%s = %s holder=%q, want leased/%s", cardID, c.State, c.LeaseHolder, lane)
	}
}

// ---------------------------------------------------------------------------
// the refusal / leasable case table (shared by several tests)
// ---------------------------------------------------------------------------

// nmCase is one nomination scenario: prepare builds the fixture and stamps the
// lane environment; token names the expected refusal, "" means the nominee is
// leased to lane-1.
type nmCase struct {
	name    string
	token   string
	prepare func(t *testing.T) (root string, store *factory.BacklogStore, nominee string)
}

// nmCases returns the scenarios keyed by name. Each decided-before-write
// refusal of the spec's token set appears once; marker-mid-text is the
// leasable control of the marker predicate.
func nmCases() []nmCase {
	single := func(state factory.BacklogState, mut func(t *testing.T, root string, store *factory.BacklogStore)) func(t *testing.T) (string, *factory.BacklogStore, string) {
		return func(t *testing.T) (string, *factory.BacklogStore, string) {
			root, store := nmBase(t, state, factory.BacklogStateQueued)
			if mut != nil {
				mut(t, root, store)
			}
			nmLaneEnv(t, "lane-1", "")
			return root, store, "t1"
		}
	}
	return []nmCase{
		{"unknown-card", "unknown-card", func(t *testing.T) (string, *factory.BacklogStore, string) {
			root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			return root, store, "t9"
		}},
		{"dropped", "dropped", single(factory.BacklogStateDropped, nil)},
		{"held", "held", single(factory.BacklogStateHold, nil)},
		{"hold-marker", "hold-marker", single(factory.BacklogStateQueued, func(t *testing.T, _ string, store *factory.BacklogStore) {
			nmSetText(t, store, "t1", nmHoldMarker+" waiting for the operator's decision")
		})},
		{"marker-leading-space", "hold-marker", single(factory.BacklogStateQueued, func(t *testing.T, _ string, store *factory.BacklogStore) {
			nmSetText(t, store, "t1", "   "+nmHoldMarker+" waiting for the operator's decision")
		})},
		{"marker-mid-text", "", single(factory.BacklogStateQueued, func(t *testing.T, _ string, store *factory.BacklogStore) {
			nmSetText(t, store, "t1", "notes that mention "+nmHoldMarker+" only in the middle")
		})},
		{"blocked", "blocked", single(factory.BacklogStateQueued, func(t *testing.T, _ string, store *factory.BacklogStore) {
			fcClassify(t, store, "t1", factory.ClassPriorityNormal, true, factory.ClassModeParallelizable)
		})},
		// Run-phase deviation (a): `blocked` applies in any queue state, an
		// operator-picked nominee included.
		{"blocked-picked", "blocked", single(factory.BacklogStatePicked, func(t *testing.T, _ string, store *factory.BacklogStore) {
			fcClassify(t, store, "t1", factory.ClassPriorityNormal, true, factory.ClassModeParallelizable)
		})},
		{"serial-slot", "serial-slot", func(t *testing.T) (string, *factory.BacklogStore, string) {
			// tS1 (t1) is a serial card in flight under lane-2; tS2 (t2) is the
			// serial card nominated while it holds the slot.
			root, store := nmBase(t, factory.BacklogStatePicked, factory.BacklogStateQueued)
			fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
			fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", Stage: homestate.CardRun})
			nmLaneEnv(t, "lane-1", "")
			return root, store, "t2"
		}},
		{"owned", "owned", func(t *testing.T) (string, *factory.BacklogStore, string) {
			root, store := nmBase(t, factory.BacklogStatePicked, factory.BacklogStateQueued)
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", Stage: homestate.CardRun})
			nmLaneEnv(t, "lane-1", "")
			return root, store, "t1"
		}},
		{"recorded", "recorded", func(t *testing.T) (string, *factory.BacklogStore, string) {
			root, store := nmBase(t, factory.BacklogStatePicked, factory.BacklogStateQueued)
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardDone, OwnerLabel: "lane-2"})
			nmLaneEnv(t, "lane-1", "")
			return root, store, "t1"
		}},
		{"foreign-worktree", "foreign-worktree", func(t *testing.T) (string, *factory.BacklogStore, string) {
			root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
			if err := os.MkdirAll(filepath.Join(root, sessionWorktreeSubdir, "t1"), 0o755); err != nil {
				t.Fatal(err)
			}
			nmLaneEnv(t, "lane-1", "")
			return root, store, "t1"
		}},
		{"quota-hold", "quota-hold", func(t *testing.T) (string, *factory.BacklogStore, string) {
			root, store := qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5)})
			return root, store, "t1"
		}},
		{"backend-skip", "backend-skip", func(t *testing.T) (string, *factory.BacklogStore, string) {
			root, store := nmBase(t, factory.BacklogStatePicked, factory.BacklogStateQueued)
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardMergeReady})
			nmLaneEnv(t, "lane-1", factory.BackendGPT)
			return root, store, "t1"
		}},
	}
}

// nmCaseByName selects scenarios from the table, in the order asked.
func nmCaseByName(t *testing.T, names ...string) []nmCase {
	t.Helper()
	all := nmCases()
	var out []nmCase
	for _, n := range names {
		found := false
		for _, c := range all {
			if c.name == n {
				out = append(out, c)
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("no nomination case %q", n)
		}
	}
	return out
}

// nmRunCase nominates the case's nominee and asserts its outcome: a refusal
// with the case's token and no write, or a lease to lane-1.
func nmRunCase(t *testing.T, c nmCase) {
	t.Helper()
	root, store, nominee := c.prepare(t)
	before := nmSnapshot(t, root, store)
	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", nominee)
	if c.token == "" {
		if err != nil {
			t.Fatalf("nominating %s: %v (stderr %q), want a lease", nominee, err, stderr)
		}
		nmAssertLeased(t, root, nominee, "lane-1")
		return
	}
	nmAssertRefused(t, out, stderr, err, c.token)
	if after := nmSnapshot(t, root, store); after != before {
		t.Errorf("a refused nomination (%s) changed the queue or the record:\nbefore:\n%s\nafter:\n%s", c.token, before, after)
	}
}

// ---------------------------------------------------------------------------
// AC-TAU-001 — the nomination, and the interface exactly one flag wider
// ---------------------------------------------------------------------------

// TestFactoryNextNominateLeasesNominee — a lane takes the card it nominates,
// not the first by priority order, and nothing else moves.
func TestFactoryNextNominateLeasesNominee(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmLaneEnv(t, "lane-1", "")

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2")
	if err != nil {
		t.Fatalf("next --card t2: %v (stderr %q)", err, stderr)
	}
	nmAssertLeased(t, root, "t2", "lane-1")
	sdAssertLeasedOutput(t, out, "t2", "-", "t2")
	if got := nmQueueState(t, store, "t2"); got != factory.BacklogStatePicked {
		t.Errorf("t2 queue state = %s, want picked", got)
	}
	for _, id := range []string{"t1", "t3"} {
		if got := nmQueueState(t, store, id); got != factory.BacklogStateQueued {
			t.Errorf("%s queue state = %s, want still queued", id, got)
		}
		if fcHasCard(t, root, id) {
			t.Errorf("%s gained a record row although only t2 was nominated", id)
		}
	}
}

// TestFactoryNextNominateUnknownCard — an id in no queue is refused with the
// unknown-card token and the stores are byte-identical.
func TestFactoryNextNominateUnknownCard(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmLaneEnv(t, "lane-1", "")
	before := nmSnapshot(t, root, store)

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t9")
	nmAssertRefused(t, out, stderr, err, "unknown-card")
	if after := nmSnapshot(t, root, store); after != before {
		t.Errorf("an unknown-card refusal changed the stores:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// nmNextFlagNames lists the flags of `moai factory next`, help excluded.
func nmNextFlagNames(t *testing.T) []string {
	t.Helper()
	next, _, err := newFactoryCommand().Find([]string{"next"})
	if err != nil || next == nil || next.Name() != "next" {
		t.Fatalf("factory next command not found: %v", err)
	}
	var names []string
	next.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			names = append(names, f.Name)
		}
	})
	sort.Strings(names)
	return names
}

// TestFactoryNextFlagSet — the CLI interface is exactly the baseline plus
// --card: nothing else was added.
func TestFactoryNextFlagSet(t *testing.T) {
	want := []string{"card", "run", "wait", "wait-bound"}
	if got := nmNextFlagNames(t); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("factory next flags = %v, want %v (the baseline --run/--wait/--wait-bound plus --card)", got, want)
	}
}

// nmMCPNextTool captures the registered factory_next tool.
func nmMCPNextTool(t *testing.T) mcp.Tool {
	t.Helper()
	var tool mcp.Tool
	found := false
	registerFactoryCardMCPTools(func(name string, tl mcp.Tool, _ server.ToolHandlerFunc) {
		if name == "factory_next" {
			tool, found = tl, true
		}
	})
	if !found {
		t.Fatal("the factory_next MCP tool is not registered")
	}
	return tool
}

// TestFactoryNextNominateMCPParity — the MCP form takes the same nomination:
// its inputs are exactly run, project_root and card; with card it leases the
// nominee or returns the CLI's refusal line as an error result, and it honors
// the quota hold and the Codex skip; without card it behaves as before.
func TestFactoryNextNominateMCPParity(t *testing.T) {
	t.Run("inputs", func(t *testing.T) {
		var got []string
		for name := range nmMCPNextTool(t).InputSchema.Properties {
			got = append(got, name)
		}
		sort.Strings(got)
		if want := "card,project_root,run"; strings.Join(got, ",") != want {
			t.Errorf("factory_next inputs = %v, want exactly %s", got, want)
		}
	})
	t.Run("leases the nominee and refuses an unknown card", func(t *testing.T) {
		root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		text, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun, "card": "t2"})
		if err != nil {
			t.Fatalf("factory_next card=t2: %v", err)
		}
		if !strings.HasPrefix(text, "t2 stage=") {
			t.Errorf("factory_next card=t2 text = %q, want the leased t2 line", text)
		}
		nmAssertLeased(t, root, "t2", "lane-1")
		if fcHasCard(t, root, "t1") {
			t.Errorf("t1 was leased: the MCP card input was ignored")
		}
		before := nmSnapshot(t, root, store)
		_, err = sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun, "card": "t9"})
		if err == nil || !strings.Contains(err.Error(), "factory next: refused unknown-card: ") {
			t.Errorf("factory_next card=t9 error = %v, want the refusal line `factory next: refused unknown-card: …`", err)
		}
		if after := nmSnapshot(t, root, store); after != before {
			t.Errorf("a refused MCP nomination changed the stores")
		}
		// Without card the tool is the unnominated lease: the first card by order.
		if _, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun}); err != nil {
			t.Fatalf("factory_next without card: %v", err)
		}
		nmAssertLeased(t, root, "t1", "lane-1")
	})
	for _, name := range []string{"quota-hold", "backend-skip"} {
		c := nmCaseByName(t, name)[0]
		t.Run("refuses "+name, func(t *testing.T) {
			root, store, nominee := c.prepare(t)
			before := nmSnapshot(t, root, store)
			_, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun, "card": nominee})
			if err == nil || !strings.Contains(err.Error(), "factory next: refused "+c.token+": ") {
				t.Fatalf("factory_next card=%s error = %v, want refused %s", nominee, err, c.token)
			}
			if after := nmSnapshot(t, root, store); after != before {
				t.Errorf("a refused MCP nomination (%s) changed the stores", c.token)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-TAU-002 — concurrent lanes never double-pick
// ---------------------------------------------------------------------------

// nmLaneRun is one lane's nomination in a race.
type nmLaneRun struct{ label, card string }

// nmLaneResult is what one raced invocation produced.
type nmLaneResult struct {
	lane, card  string
	out, stderr string
	err         error
}

// nmRaceAtSeamTolerant runs each lane's `factory next --card <card>`
// concurrently and holds every invocation at the nomination seam — after any
// promotion, before the first record write — until all of them have arrived or
// the tolerant gate's grace window has passed (SPEC-FACTORY-ATOMIC-LEASE-001
// plan §5). The seam sits inside the lease section, where only one lane can be:
// the strict form of this helper (release only when every lane has arrived)
// would wait forever on the second lane, so the gate lets the first lane go
// after the grace window and the second proceeds in order once the lock is free.
// A lane label is read once at the start of an invocation, so the label is
// switched between launches (flRaceLanes).
func nmRaceAtSeamTolerant(t *testing.T, lanes []nmLaneRun) []nmLaneResult {
	t.Helper()
	g := newFLGate(len(lanes))
	prev := factoryNominateBeforeRecord
	factoryNominateBeforeRecord = func(string) error {
		g.hold(strconv.Itoa(flGoroutineID()))
		return nil
	}
	t.Cleanup(func() { factoryNominateBeforeRecord = prev })
	raced := flRaceLanes(t, g, lanes)
	results := make([]nmLaneResult, 0, len(raced))
	for _, r := range raced {
		results = append(results, nmLaneResult(r))
	}
	return results
}

// nmIsolatedWorktrees stubs the shared worktree materializer so each card named
// gets its OWN independent repository as its "worktree". The post-lease step of
// a successful lease (factoryEnsureCardWorktree) renames the new tree's branch;
// two lanes doing that at the same instant in ONE repository collide on the
// shared reflog temp file — a pre-existing hazard of that step, outside the
// lease invariant AC-TAU-002 asserts. Separate repositories remove the
// collision while the contention under test (the queue promotion and the
// version-checked record edges) stays untouched. The repositories are built up
// front on the test goroutine; the stub only looks them up.
func nmIsolatedWorktrees(t *testing.T, cards ...string) {
	t.Helper()
	trees := make(map[string]string, len(cards))
	for _, id := range cards {
		dir := t.TempDir()
		initGitRepo(t, dir)
		trees[id] = dir
	}
	prev := worktree.WorktreeCreator
	worktree.WorktreeCreator = func(name string, _ io.Writer) (string, error) {
		dir, ok := trees[name]
		if !ok {
			return "", fmt.Errorf("no isolated worktree prepared for %s", name)
		}
		return dir, nil
	}
	t.Cleanup(func() { worktree.WorktreeCreator = prev })
}

// TestFactoryNextNominateConcurrentLanes — two lanes nominating different
// cards at the same moment each hold their own card.
func TestFactoryNextNominateConcurrentLanes(t *testing.T) {
	root, _ := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmIsolatedWorktrees(t, "t1", "t2")
	results := nmRaceAtSeamTolerant(t, []nmLaneRun{{"lane-1", "t1"}, {"lane-2", "t2"}})
	for _, r := range results {
		if r.err != nil {
			t.Errorf("lane %s nominating %s errored: %v (stderr %q)", r.lane, r.card, r.err, r.stderr)
		}
	}
	nmAssertLeased(t, root, "t1", "lane-1")
	nmAssertLeased(t, root, "t2", "lane-2")
}

// TestFactoryNextNominateSameCardExactlyOne — two lanes nominating the same
// card end with exactly one holder and one refusal; the loser re-nominates a
// different candidate and succeeds.
func TestFactoryNextNominateSameCardExactlyOne(t *testing.T) {
	root, _ := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	results := nmRaceAtSeamTolerant(t, []nmLaneRun{{"lane-1", "t1"}, {"lane-2", "t1"}})

	var winners, losers []nmLaneResult
	for _, r := range results {
		if r.err == nil {
			winners = append(winners, r)
		} else {
			losers = append(losers, r)
		}
	}
	if len(winners) != 1 || len(losers) != 1 {
		t.Fatalf("t1 nominated by two lanes: %d won, %d lost, want exactly one of each (results %+v)", len(winners), len(losers), results)
	}
	loser := losers[0]
	if code := nmExit(loser.err); code != 4 {
		t.Fatalf("the losing lane exited %d (%v), want the refusal status 4", code, loser.err)
	}
	if !strings.Contains(loser.stderr, "refused raced: ") && !strings.Contains(loser.stderr, "refused owned: ") {
		t.Errorf("loser stderr = %q, want a `raced` or `owned` refusal line", loser.stderr)
	}
	if n := nmRowCount(t, root, "t1"); n != 1 {
		t.Errorf("t1 has %d record rows, want exactly 1", n)
	}
	nmAssertLeased(t, root, "t1", winners[0].lane)

	// The losing lane re-selects: a different candidate leases normally.
	nmLaneEnv(t, loser.lane, "")
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2"); err != nil {
		t.Fatalf("%s re-nominating t2: %v (stderr %q)", loser.lane, err, stderr)
	}
	nmAssertLeased(t, root, "t2", loser.lane)
}

// ---------------------------------------------------------------------------
// AC-TAU-004 — keep-set cards are never leased
// ---------------------------------------------------------------------------

// TestFactoryNextNominateRefusesKeepSet — each keep-set card is refused with
// its own token and changes no state; a card that merely mentions the marker
// mid-text is not a marker card and leases.
func TestFactoryNextNominateRefusesKeepSet(t *testing.T) {
	for _, c := range nmCaseByName(t, "held", "hold-marker", "marker-leading-space", "marker-mid-text", "blocked", "blocked-picked", "serial-slot", "dropped", "owned") {
		t.Run(c.name, func(t *testing.T) { nmRunCase(t, c) })
	}
	t.Run("ordinary card leases", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2"); err != nil {
			t.Fatalf("next --card t2: %v (stderr %q)", err, stderr)
		}
		nmAssertLeased(t, root, "t2", "lane-1")
	})
}

// TestFactoryNextArmCSkipsHoldMarker — the bare lease path skips a queued card
// whose trimmed text opens with the marker (with or without leading space) and
// leases the card that only mentions it mid-text.
func TestFactoryNextArmCSkipsHoldMarker(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmSetText(t, store, "t1", nmHoldMarker+" parked card")
	nmSetText(t, store, "t2", "   "+nmHoldMarker+" parked card with leading space")
	nmSetText(t, store, "t3", "ordinary notes that mention "+nmHoldMarker+" mid-text")
	nmLaneEnv(t, "lane-1", "")

	for _, want := range []string{"t3", "t4"} {
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("bare next (%s expected): %v (stderr %q)", want, err, stderr)
		}
		if head := nmLeasedHead(out); !strings.HasPrefix(head, want+" stage=") {
			t.Fatalf("bare next leased %q, want %s (a marker card must be skipped)", head, want)
		}
	}
	out, _, err := qasRunNext(t, "--run", fcRun)
	sdExit3(t, "only marker cards remain", err)
	if !strings.Contains(out, "no card") {
		t.Errorf("stdout = %q, want a no-card line", out)
	}
	for _, id := range []string{"t1", "t2"} {
		if got := nmQueueState(t, store, id); got != factory.BacklogStateQueued {
			t.Errorf("%s queue state = %s, want still queued", id, got)
		}
		if fcHasCard(t, root, id) {
			t.Errorf("%s gained a record row although its text opens with the marker", id)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-TAU-014 — state semantics, quota, Codex, record states
// ---------------------------------------------------------------------------

// TestFactoryNextNominateQuotaHold — the nominated lease honors the quota
// hold exactly as the unnominated new-card arms do: a queued nominee is
// refused, a card already assigned to the lane still leases, and --wait keeps
// waiting through the hold while any other refusal ends it at once.
func TestFactoryNextNominateQuotaHold(t *testing.T) {
	t.Run("a new card is refused and nothing changes", func(t *testing.T) {
		nmRunCase(t, nmCaseByName(t, "quota-hold")[0])
	})
	t.Run("a card already assigned to the lane still leases", func(t *testing.T) {
		root, _ := qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5), assigned: true})
		if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t4"); err != nil {
			t.Fatalf("next --card t4 under a hold: %v (stderr %q)", err, stderr)
		}
		nmAssertLeased(t, root, "t4", "lane-1")
	})
	t.Run("--wait waits through the hold until the bound", func(t *testing.T) {
		qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5)})
		sleeps := qasFakeClock(t, nil, nil)
		out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1", "--wait", "--wait-bound", "10s")
		nmAssertRefused(t, out, stderr, err, "quota-hold")
		if *sleeps != 2 {
			t.Errorf("wait slept %d times before the 10s bound with a 5s interval, want 2 (the hold is waited through)", *sleeps)
		}
	})
	t.Run("--wait does not wait on a permanent refusal", func(t *testing.T) {
		nmBase(t, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		sleeps := qasFakeClock(t, nil, nil)
		out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t9", "--wait", "--wait-bound", "10s")
		nmAssertRefused(t, out, stderr, err, "unknown-card")
		if *sleeps != 0 {
			t.Errorf("wait slept %d times on an unknown-card refusal, want 0", *sleeps)
		}
	})
}

// TestFactoryNextNominateBackendSkip — a Codex lane is refused a nominee at
// or past merge-ready and the stores do not change; the same card leases to a
// Claude lane, and an earlier-stage card leases to the Codex lane.
func TestFactoryNextNominateBackendSkip(t *testing.T) {
	t.Run("codex lane, merge-ready nominee", func(t *testing.T) {
		nmRunCase(t, nmCaseByName(t, "backend-skip")[0])
	})
	t.Run("claude lane leases the same card", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStatePicked)
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardMergeReady})
		nmLaneEnv(t, "lane-1", factory.BackendClaude)
		if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
			t.Fatalf("claude lane next --card t1: %v (stderr %q)", err, stderr)
		}
		nmAssertLeased(t, root, "t1", "lane-1")
	})
	t.Run("codex lane leases an earlier-stage card", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStatePicked)
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun})
		nmLaneEnv(t, "lane-1", factory.BackendGPT)
		if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
			t.Fatalf("codex lane next --card t1: %v (stderr %q)", err, stderr)
		}
		nmAssertLeased(t, root, "t1", "lane-1")
	})
}

// TestFactoryNextNominateRecordStateTokens — every factory-record state maps
// to a token or to a lease: in-flight states are `owned` for any holder,
// terminal and parked states (and a legacy value) are `recorded`, a card with
// a worktree directory nobody owns is `foreign-worktree`, and the leasable
// shapes (no row, a picked row with no owner, a row assigned to this lane)
// lease.
func TestFactoryNextNominateRecordStateTokens(t *testing.T) {
	place := func(t *testing.T, row homestate.Card) (string, *factory.BacklogStore) {
		t.Helper()
		root, store := nmBase(t, factory.BacklogStatePicked)
		row.CardID = "t1"
		fcPlace(t, root, row)
		nmLaneEnv(t, "lane-1", "")
		return root, store
	}
	refuse := func(t *testing.T, root string, store *factory.BacklogStore, token string) {
		t.Helper()
		before := nmSnapshot(t, root, store)
		out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
		nmAssertRefused(t, out, stderr, err, token)
		if after := nmSnapshot(t, root, store); after != before {
			t.Errorf("a %s refusal changed the stores", token)
		}
	}
	lease := func(t *testing.T, root string) {
		t.Helper()
		if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
			t.Fatalf("next --card t1: %v (stderr %q)", err, stderr)
		}
		nmAssertLeased(t, root, "t1", "lane-1")
	}

	inFlight := []string{
		homestate.CardLeased, homestate.CardPlan, homestate.CardPlanAudit, homestate.CardKickoff, homestate.CardRun,
		homestate.CardSync, homestate.CardSyncAudit, homestate.CardMergeReady, homestate.CardMerging,
		homestate.CardMergedLocal, homestate.CardPushed, homestate.CardCIGreen,
	}
	for _, state := range inFlight {
		t.Run("owned/"+state, func(t *testing.T) {
			root, store := place(t, homestate.Card{State: state, OwnerLabel: "lane-2", LeaseHolder: "lane-2"})
			refuse(t, root, store, "owned")
		})
	}
	t.Run("owned/leased-by-this-lane", func(t *testing.T) {
		root, store := place(t, homestate.Card{State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1"})
		refuse(t, root, store, "owned")
	})
	// A `picked` row is leasable only while it has no owner (an operator pick);
	// once an owner is recorded it is `owned`, for any lane — this one included.
	for _, owner := range []string{"lane-2", "lane-1"} {
		t.Run("owned/picked-with-owner-"+owner, func(t *testing.T) {
			root, store := place(t, homestate.Card{State: homestate.CardPicked, OwnerLabel: owner})
			refuse(t, root, store, "owned")
		})
	}
	t.Run("owned/assigned-to-another-lane", func(t *testing.T) {
		root, store := place(t, homestate.Card{State: homestate.CardAssigned, OwnerLabel: "lane-2"})
		refuse(t, root, store, "owned")
	})
	for _, state := range []string{homestate.CardDone, homestate.CardAbandoned, homestate.CardFailed, homestate.CardBlocked, homestate.CardNeedsDecision, "legacy-state"} {
		t.Run("recorded/"+state, func(t *testing.T) {
			root, store := place(t, homestate.Card{State: state, OwnerLabel: "lane-2"})
			refuse(t, root, store, "recorded")
		})
	}
	t.Run("foreign-worktree", func(t *testing.T) { nmRunCase(t, nmCaseByName(t, "foreign-worktree")[0]) })
	t.Run("leasable/no-row", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStatePicked)
		nmLaneEnv(t, "lane-1", "")
		lease(t, root)
	})
	t.Run("leasable/queued-no-row", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		lease(t, root)
	})
	t.Run("leasable/picked-unowned", func(t *testing.T) {
		root, _ := place(t, homestate.Card{State: homestate.CardPicked})
		lease(t, root)
	})
	t.Run("leasable/assigned-to-this-lane", func(t *testing.T) {
		root, _ := place(t, homestate.Card{State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun})
		lease(t, root)
	})
}

// TestFactoryNextNominateRefusalLeavesStateUnchanged — every refusal decided
// by the read-only validation (every token but `raced`) leaves the queue file
// and the factory record byte-identical.
func TestFactoryNextNominateRefusalLeavesStateUnchanged(t *testing.T) {
	for _, c := range nmCases() {
		if c.token == "" {
			continue
		}
		t.Run(c.token+"/"+c.name, func(t *testing.T) { nmRunCase(t, c) })
	}
}

// nmQueuedNominee builds the two-card queue the seam tests use (t1 is the
// queued nominee) and stamps lane-1.
func nmQueuedNominee(t *testing.T) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmLaneEnv(t, "lane-1", "")
	return root, store
}

// nmSetSeam installs fn as the nomination seam for the test.
func nmSetSeam(t *testing.T, fn func(cardID string) error) {
	t.Helper()
	prev := factoryNominateBeforeRecord
	factoryNominateBeforeRecord = fn
	t.Cleanup(func() { factoryNominateBeforeRecord = prev })
}

// TestFactoryNextNominatePromoteThenLose — the nominee is promoted to picked,
// a competing lane leases it at the seam; the invocation is refused `raced`,
// the competitor holds the card, the queue shows the competitor's picked
// state, and no second record row exists.
func TestFactoryNextNominatePromoteThenLose(t *testing.T) {
	root, store := nmQueuedNominee(t)
	nmSetSeam(t, func(cardID string) error {
		db, err := homestate.OpenFactory(root)
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
		_, leased, _, err := factoryNextRecordAndClaim(context.Background(), db, root, fcRun, cardID, "lane-2")
		if err != nil {
			return err
		}
		if !leased {
			return errors.New("the competing lane did not lease the nominee at the seam")
		}
		return nil
	})

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, out, stderr, err, "raced")
	nmAssertLeased(t, root, "t1", "lane-2")
	if got := nmQueueState(t, store, "t1"); got != factory.BacklogStatePicked {
		t.Errorf("t1 queue state = %s, want picked (the competitor's state is left alone)", got)
	}
	if n := nmRowCount(t, root, "t1"); n != 1 {
		t.Errorf("t1 has %d record rows, want exactly 1", n)
	}
	if got := nmQueueState(t, store, "t2"); got != factory.BacklogStateQueued {
		t.Errorf("t2 queue state = %s, want untouched", got)
	}
}

// TestFactoryNextNominateClaimRefusedRollsBack — a failure injected through
// the seam, with no other holder and before any record row exists, ends with
// a non-token error status, the queue item back to queued, and no record row.
func TestFactoryNextNominateClaimRefusedRollsBack(t *testing.T) {
	root, store := nmQueuedNominee(t)
	before := nmSnapshot(t, root, store)
	nmSetSeam(t, func(string) error { return errors.New(nmInjected) })

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err == nil {
		t.Fatalf("a claim failure was swallowed: stdout=%q stderr=%q", out, stderr)
	}
	if code := nmExit(err); code == 4 {
		t.Errorf("a seam failure exited 4 (a refusal status); it is not a refusal of the card — err=%v", err)
	}
	if !strings.Contains(err.Error(), nmInjected) {
		t.Errorf("the error = %v, want the injected failure reported", err)
	}
	if got := nmQueueState(t, store, "t1"); got != factory.BacklogStateQueued {
		t.Errorf("t1 queue state = %s, want restored to queued", got)
	}
	if fcHasCard(t, root, "t1") {
		t.Errorf("t1 has a record row after a failure injected before RecordPicked")
	}
	if after := nmSnapshot(t, root, store); after != before {
		t.Errorf("the failed nomination left a trace:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestFactoryNextNominateCompensationFailure — the compensation acts only on
// a promotion that is still its own.
func TestFactoryNextNominateCompensationFailure(t *testing.T) {
	t.Run("item-moved", func(t *testing.T) {
		root, store := nmQueuedNominee(t)
		var op *flOp
		inside := false
		nmSetSeam(t, func(cardID string) error {
			// An operator drop between the promotion and the compensation. The
			// write is started from a goroutine (the section holds the queue lock
			// here, so a synchronous write would wait out the whole budget): it
			// applies after the verb, and the compensation must not overwrite it.
			op = flStartState(store, cardID, factory.BacklogStateDropped)
			inside = op.within(flWindow)
			return errors.New(nmInjected)
		})
		_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
		if op == nil {
			t.Fatal("the nomination seam was never reached")
		}
		flJoin(t, []*flOp{op})
		if err == nil || !strings.Contains(err.Error(), nmInjected) {
			t.Fatalf("the original failure was not reported: err=%v stderr=%q", err, stderr)
		}
		if inside {
			t.Errorf("the operator's drop completed inside the section (the seam returned only after the write finished)")
		}
		if got := nmQueueState(t, store, "t1"); got != factory.BacklogStateDropped {
			t.Errorf("t1 queue state = %s, want dropped (the compensation must not overwrite the operator's change)", got)
		}
		if fcHasCard(t, root, "t1") {
			t.Errorf("t1 has a record row")
		}
	})
}

// TestFactoryNextNominateCompensateRechecksRecord — the compensation reads the
// factory-record row under the queue lock and acts on it: the queue item is not
// restored to `queued` under a card the factory holds (a restored `queued` item
// could be claimed a second time). The table calls the compensation directly,
// inside a lease section opened by the test, with the state pre-arranged.
//
// SPEC-FACTORY-ATOMIC-LEASE-001 plan §5 removed this test's lock-wait subtest:
// it modeled a lease landing while the compensation waits for the queue lock, a
// window the section removes (the compensation no longer takes the lock; it
// runs inside the hold that made the promotion). The operator-write analogue is
// TestFactoryLeaseCompensationKeepsOperatorPick.
func TestFactoryNextNominateCompensateRechecksRecord(t *testing.T) {
	cases := []struct {
		name         string
		row          *homestate.Card
		promoted     bool
		wantOther    bool
		wantQueueEnd factory.BacklogState
	}{
		{"no row restores", nil, true, false, factory.BacklogStateQueued},
		{"unowned picked row restores", &homestate.Card{State: homestate.CardPicked}, true, false, factory.BacklogStateQueued},
		{"another lane's lease is left alone", &homestate.Card{State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", Stage: homestate.CardRun}, true, true, factory.BacklogStatePicked},
		{"another lane's assignment is left alone", &homestate.Card{State: homestate.CardAssigned, OwnerLabel: "lane-2"}, true, true, factory.BacklogStatePicked},
		{"this lane's own lease is left alone and is not another holder", &homestate.Card{State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", Stage: homestate.CardRun}, true, false, factory.BacklogStatePicked},
		{"a promotion this invocation did not make is never undone", nil, false, false, factory.BacklogStatePicked},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root, store := nmBase(t, factory.BacklogStatePicked)
			if c.row != nil {
				row := *c.row
				row.CardID = "t1"
				fcPlace(t, root, row)
			}
			db := fcOpen(t, root)
			var other bool
			var err error
			if lockErr := todoStoreAt(root).WithLock(func(l *factory.LockedBacklog) error {
				other, err = factoryNominateCompensate(context.Background(), l, db, fcRun, "lane-1", "t1", c.promoted)
				return nil
			}); lockErr != nil {
				t.Fatalf("the lease section: %v", lockErr)
			}
			if err != nil {
				t.Fatalf("compensation: %v", err)
			}
			if other != c.wantOther {
				t.Errorf("another holder = %v, want %v", other, c.wantOther)
			}
			if got := nmQueueState(t, store, "t1"); got != c.wantQueueEnd {
				t.Errorf("t1 queue state = %s, want %s", got, c.wantQueueEnd)
			}
		})
	}
}

// TestFactoryNextNominateBlankCardIsAnError — a blank `--card` (CLI) or a blank
// `card` (MCP) is an error that names the missing id; it never falls back to
// the bare priority-order lease the session did not choose.
func TestFactoryNextNominateBlankCardIsAnError(t *testing.T) {
	assertNothingLeased := func(t *testing.T, root string, store *factory.BacklogStore, before string) {
		t.Helper()
		if after := nmSnapshot(t, root, store); after != before {
			t.Errorf("a blank card changed the queue or the record:\nbefore:\n%s\nafter:\n%s", before, after)
		}
		if fcHasCard(t, root, "t1") {
			t.Errorf("t1 was leased: a blank card fell back to the bare lease")
		}
	}
	for _, blank := range []string{"", "   "} {
		t.Run(fmt.Sprintf("cli %q", blank), func(t *testing.T) {
			root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			before := nmSnapshot(t, root, store)
			out, stderr, err := qasRunNext(t, "--run", fcRun, "--card="+blank)
			if err == nil || !strings.Contains(err.Error(), "--card needs a card id") {
				t.Fatalf("blank --card: err = %v (stdout %q stderr %q), want `--card needs a card id`", err, out, stderr)
			}
			if code := nmExit(err); code == 4 {
				t.Errorf("a blank --card exited 4 (a refusal status); it names no card to refuse: %v", err)
			}
			assertNothingLeased(t, root, store, before)
		})
		t.Run(fmt.Sprintf("mcp %q", blank), func(t *testing.T) {
			root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			before := nmSnapshot(t, root, store)
			_, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun, "card": blank})
			if err == nil || !strings.Contains(err.Error(), "card needs a card id") {
				t.Fatalf("blank card: err = %v, want `card needs a card id`", err)
			}
			assertNothingLeased(t, root, store, before)
		})
	}
}

// ---------------------------------------------------------------------------
// AC-TAU-006 — the bare path is unchanged, and an all-marker queue ends on 3
// ---------------------------------------------------------------------------

// TestFactoryNextAllMarkerQueueExitsNoCard — a queue whose only queued cards
// open with the marker ends on the no-card exit, not on a retry or a lease.
func TestFactoryNextAllMarkerQueueExitsNoCard(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmSetText(t, store, "t1", nmHoldMarker+" first parked card")
	nmSetText(t, store, "t2", "  "+nmHoldMarker+" second parked card")
	nmLaneEnv(t, "lane-1", "")
	before := nmSnapshot(t, root, store)

	out, _, err := qasRunNext(t, "--run", fcRun)
	sdExit3(t, "all-marker queue", err)
	if strings.TrimSpace(out) != "no card is available" {
		t.Errorf("stdout = %q, want the no-card line", out)
	}
	if after := nmSnapshot(t, root, store); after != before {
		t.Errorf("an all-marker queue changed the stores")
	}
}

// nmStep renders one bare `next` run as a golden transcript line. The head is
// the leased card's line (the worktree materializer and the pull-request
// lookup also write to the streams, and what they say depends on the machine:
// the commit identity, whether `gh` answers); the stderr part keeps only what
// the verb itself writes — the quota hold line and refusals.
func nmStep(label, out, stderr string, err error) string {
	head := ""
	for _, l := range nmNonEmptyLines(out) {
		if nmCardHeadRE.MatchString(l) || strings.HasPrefix(l, "no card") {
			head = l
			break
		}
	}
	var own []string
	for _, l := range nmNonEmptyLines(stderr) {
		if strings.HasPrefix(l, "quota hold") || strings.HasPrefix(l, "factory next:") {
			own = append(own, strings.TrimSpace(l))
		}
	}
	return fmt.Sprintf("%s exit=%d head=%q stderr=%q", label, nmExit(err), head, strings.Join(own, "\n"))
}

// nmBareGolden is the transcript the bare `moai factory next` produced on the
// tree BEFORE the nominated lease was implemented (the unmodified tree plus the
// inert seam declaration): the minimum case set of AC-TAU-006 — the default arm
// order (a)→(b)→(b2)→(c), a bounded --wait, the quota hold with an assigned
// card still leasing, the Codex skip, the serial slot, and the no-card exit.
var nmBareGolden = map[string]string{
	"arm-order": `next#1 exit=0 head="t5 stage=run worktree=t5" stderr=""
next#2 exit=0 head="t4 stage=- worktree=t4" stderr=""
next#3 exit=0 head="t3 stage=- worktree=t3" stderr=""
next#4 exit=0 head="t1 stage=- worktree=t1" stderr=""
next#5 exit=0 head="t2 stage=- worktree=t2" stderr=""
next#6 exit=3 head="no card is available" stderr=""`,
	"wait-bound": `next --wait exit=3 head="no card is available" stderr="" sleeps=2`,
	"quota-hold": `next#1 exit=0 head="t4 stage=run worktree=t4" stderr=""
next#2 exit=3 head="" stderr="quota hold: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z"`,
	"codex-skip": `next#1 exit=0 head="t2 stage=run worktree=t2" stderr=""
next#2 exit=3 head="no card is available" stderr=""`,
	"serial-slot": `next#1 exit=0 head="t1 stage=- worktree=t1" stderr=""
next#2 exit=0 head="t3 stage=- worktree=t3" stderr=""
next#3 exit=3 head="no card is available" stderr=""`,
}

// nmCompareGolden fails when the transcript differs from the recorded golden.
func nmCompareGolden(t *testing.T, key string, steps []string) {
	t.Helper()
	got := strings.Join(steps, "\n")
	if want := nmBareGolden[key]; got != want {
		t.Errorf("bare `factory next` transcript %q changed:\n--- got ---\n%s\n--- want (golden) ---\n%s", key, got, want)
	}
}

// TestFactoryNextBareUnchanged — the golden of the bare path over a
// marker-free queue. It is green on the tree before the nominated lease
// exists (it pins current behavior); its non-vacuity is the seeded
// perturbation recorded in progress.md.
func TestFactoryNextBareUnchanged(t *testing.T) {
	t.Run("arm-order", func(t *testing.T) {
		// Queue order is deliberately NOT arm order: two queued cards first,
		// then a queue-picked card with no row (b2), an operator-picked card with
		// a row (b), and a card assigned to this lane (a).
		root, _ := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked)
		fcPlace(t, root,
			homestate.Card{CardID: "t4", State: homestate.CardPicked},
			homestate.Card{CardID: "t5", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
		)
		nmLaneEnv(t, "lane-1", "")
		var steps []string
		for i := 1; i <= 6; i++ {
			out, stderr, err := qasRunNext(t, "--run", fcRun)
			steps = append(steps, nmStep(fmt.Sprintf("next#%d", i), out, stderr, err))
		}
		nmCompareGolden(t, "arm-order", steps)
	})
	t.Run("wait-bound", func(t *testing.T) {
		nmBase(t)
		nmLaneEnv(t, "lane-1", "")
		sleeps := qasFakeClock(t, nil, nil)
		out, stderr, err := qasRunNext(t, "--run", fcRun, "--wait", "--wait-bound", "10s")
		nmCompareGolden(t, "wait-bound", []string{fmt.Sprintf("%s sleeps=%d", nmStep("next --wait", out, stderr, err), *sleeps)})
	})
	t.Run("quota-hold", func(t *testing.T) {
		qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5), assigned: true})
		var steps []string
		for i := 1; i <= 2; i++ {
			out, stderr, err := qasRunNext(t, "--run", fcRun)
			steps = append(steps, nmStep(fmt.Sprintf("next#%d", i), out, stderr, err))
		}
		nmCompareGolden(t, "quota-hold", steps)
	})
	t.Run("codex-skip", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStatePicked, factory.BacklogStatePicked)
		fcPlace(t, root,
			homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardMergeReady},
			homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
		)
		nmLaneEnv(t, "lane-1", factory.BackendGPT)
		var steps []string
		for i := 1; i <= 2; i++ {
			out, stderr, err := qasRunNext(t, "--run", fcRun)
			steps = append(steps, nmStep(fmt.Sprintf("next#%d", i), out, stderr, err))
		}
		nmCompareGolden(t, "codex-skip", steps)
	})
	t.Run("serial-slot", func(t *testing.T) {
		_, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
		fcClassify(t, store, "t1", factory.ClassPriorityHigh, false, factory.ClassModeSerial)
		fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
		fcClassify(t, store, "t3", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		nmLaneEnv(t, "lane-1", "")
		var steps []string
		for i := 1; i <= 3; i++ {
			out, stderr, err := qasRunNext(t, "--run", fcRun)
			steps = append(steps, nmStep(fmt.Sprintf("next#%d", i), out, stderr, err))
		}
		nmCompareGolden(t, "serial-slot", steps)
	})
}

// ---------------------------------------------------------------------------
// AC-TAU-005 — a lane is refused `moai todo --auto`; a non-lane GPT session is not
// ---------------------------------------------------------------------------

// nmAutoFixture seeds a two-card queue (lane environment scrubbed first) and
// installs hermetic ranking seams: the live landed seam runs `gh`.
func nmAutoFixture(t *testing.T) *factory.BacklogStore {
	t.Helper()
	_, store := todoFixture(t)
	sdClearLaneEnv(t)
	for _, text := range []string{"lane auto card one", "lane auto card two"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("todo add %q: %v", text, err)
		}
	}
	origLanded, origJev := todoAutoLandedLookup, todoAutoJevRanker
	t.Cleanup(func() { todoAutoLandedLookup, todoAutoJevRanker = origLanded, origJev })
	todoAutoLandedLookup = func(*factory.BacklogRecord) (map[string]factory.PRLinkKind, error) { return nil, nil }
	todoAutoJevRanker = func(jev.Request) jev.Result { return jev.Result{Availability: jev.Disabled} }
	return store
}

// nmLaneVariants are the three lane definitions of the refusal: the label
// alone, the role marker alone, and both.
var nmLaneVariants = []struct {
	name string
	set  func(t *testing.T)
}{
	{"label-only", func(t *testing.T) { t.Setenv(config.EnvMoaiFactoryWorker, "lane-1") }},
	{"role-only", func(t *testing.T) { t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane) }},
	{"role-and-label", func(t *testing.T) {
		t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
	}},
}

// TestTodoLaneRefusesAutoCycle — a lane session, however it is identified, is
// refused `moai todo --auto` and the queue file is byte-identical; the
// read-only forms still run.
func TestTodoLaneRefusesAutoCycle(t *testing.T) {
	for _, v := range nmLaneVariants {
		t.Run(v.name, func(t *testing.T) {
			store := nmAutoFixture(t)
			v.set(t)
			before := sdQueueBytes(t, store)
			out, _, err := runTodo(t, "--auto", "--auto-wait", "1ms")
			if err == nil {
				t.Fatalf("a lane session ran `moai todo --auto`: output %q", out)
			}
			if after := sdQueueBytes(t, store); after != before {
				t.Errorf("the refused lane `--auto` changed the queue file")
			}
		})
	}
	t.Run("read-only forms still run", func(t *testing.T) {
		nmAutoFixture(t)
		t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
		for _, args := range [][]string{nil, {"list"}} {
			if _, _, err := runTodo(t, args...); err != nil {
				t.Errorf("lane `moai todo %s`: %v, want the read-only form to run", strings.Join(args, " "), err)
			}
		}
	})
}

// TestTodoLaneAutoRefusalText — the refusal is the dedicated text: it names
// the lease path and says the --auto authorization is exercised through it,
// so a lane that reads it proceeds instead of asking.
func TestTodoLaneAutoRefusalText(t *testing.T) {
	for _, v := range nmLaneVariants {
		t.Run(v.name, func(t *testing.T) {
			nmAutoFixture(t)
			v.set(t)
			_, _, err := runTodo(t, "--auto", "--auto-wait", "1ms")
			if err == nil {
				t.Fatal("a lane session ran `moai todo --auto`")
			}
			for _, want := range []string{"the --auto authorization is exercised through", "moai factory next --card <id>"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("refusal = %q, want it to contain %q", err.Error(), want)
				}
			}
			if strings.Contains(err.Error(), "cannot mutate the queue") {
				t.Errorf("refusal reuses the queue-mutation text: %q", err.Error())
			}
		})
	}
}

// TestTodoNonLaneGPTSessionNotRefused — a non-lane session whose only marker
// is the Codex backend runs the cycle exactly as today: it has no lease
// alternative, so refusing it would strand its batch approval. Green on the
// tree before the guard exists; its non-vacuity is the seeded over-broad
// predicate recorded in progress.md.
func TestTodoNonLaneGPTSessionNotRefused(t *testing.T) {
	nmAutoFixture(t)
	t.Setenv(config.EnvFactoryBackend, factory.BackendGPT)
	out, _, err := runTodo(t, "--auto", "--auto-wait", "1ms")
	if err != nil {
		t.Fatalf("a non-lane Codex-backend session was refused `moai todo --auto`: %v", err)
	}
	if !strings.Contains(out, "accept t1") {
		t.Errorf("--auto output = %q, want the cycle to have run (accept t1)", out)
	}
}

// TestFactoryFallbackDeclarePrintsLeasePath — the declare confirmation routes
// the lane to the lease path, not to the serial cycle a lane may no longer run.
func TestFactoryFallbackDeclarePrintsLeasePath(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	laneEnv(t)
	out, _, err := runFactory(t, "fallback", "declare", "--trigger", "channel-unavailable")
	if err != nil {
		t.Fatalf("declare: %v", err)
	}
	if !strings.Contains(out, "moai factory next [--card <id>]") {
		t.Errorf("declare output = %q, want the lease path `moai factory next [--card <id>]`", out)
	}
	if strings.Contains(out, "/moai:todo --auto") {
		t.Errorf("declare output = %q, must no longer route a lane to /moai:todo --auto", out)
	}
}
