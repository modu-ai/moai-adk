package cli

// Card t2018 measurement of what a run-less `moai factory renew <card>` writes besides its own card
// (leader order, point 4; ledger event 5980 is the suspicion this answers).
//
// A run-less renew picks its run in resolveFactoryRenewRunRule (factory_card.go). With two or more
// active runs that goes through factorymsg.ResolveActiveRun, which calls ReconcileActiveRuns
// (store.go:433): a write that retires dead-owner runs and, through RebindRun, moves their
// unfinished cards into the one live run. Two tests pin the answer:
//
//   - TestRenewWithoutRunMovesOnlyItsOwnCard: every shape where no other row moves (a dead run whose
//     row is not active, and the guards that refuse a consolidation). Passes.
//   - TestRenewWithoutRunLeavesDeadOwnerRunRowsAlone: the shapes where a dead-owner run is still an
//     ACTIVE row next to the live one. RED: the bare renew moves that run's rows, even when the renew
//     itself is refused.
//
// Every fixture lives under t.TempDir(); the factory DB path is asserted to lie under the test's
// temp tree before anything opens it, and the lane label is bound as in factory_renew_test.go.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// renewMoveVariant describes one ledger shape: the dead run's row status, its cards, an optional
// dead-worker row, the lane's roster binding and the recorded pid of the lane's own worker row.
type renewMoveVariant struct {
	name       string
	deadStatus string // "active" (dead owner stamp, row still active) or "retired" (row not active)
	deadCards  []homestate.Card
	deadWorker int    // pid of a worker row "lane-9" bound to run-dead; 0 = none
	laneBound  string // roster binding of lane-1: "run-live", "run-dead" or ""
	lanePID    int    // recorded pid of lane-1's worker row
	t1Holder   string // lease holder of t1 in run-live; "" means lane-1
	renewCard  string // card the renew names; "" means t1
	refused    bool   // the renew itself is expected to be refused
}

// renewMoveDeadPID exceeds every platform's pid_max, so the owner stamp can never classify live.
const renewMoveDeadPID = 99999999

// renewMoveEvalExisting resolves symlinks of the longest existing ancestor of p.
func renewMoveEvalExisting(p string) string {
	rest := ""
	for cur := filepath.Clean(p); ; cur = filepath.Dir(cur) {
		if r, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(r, rest)
		}
		if filepath.Dir(cur) == cur {
			return filepath.Clean(p)
		}
		rest = filepath.Join(filepath.Base(cur), rest)
	}
}

// renewMoveAssertIsolated fails loudly unless the resolved factory DB path (and HOME) lie under this
// test's temp tree, the parent of every t.TempDir() the test creates.
func renewMoveAssertIsolated(t *testing.T, root string) {
	t.Helper()
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatalf("resolve factory DB path: %v", err)
	}
	tmp := renewMoveEvalExisting(filepath.Dir(t.TempDir())) + string(filepath.Separator)
	if got := renewMoveEvalExisting(path); !strings.HasPrefix(got, tmp) {
		t.Fatalf("ISOLATION VIOLATION: factory DB path %s (resolved %s) is not under the test temp tree %s", path, got, tmp)
	}
	if home := renewMoveEvalExisting(os.Getenv("HOME")); !strings.HasPrefix(home, tmp) {
		t.Fatalf("ISOLATION VIOLATION: HOME %s is not under the test temp tree %s", home, tmp)
	}
}

func renewMoveExec(t *testing.T, db *homestate.FactoryDB, query string, args ...any) {
	t.Helper()
	if _, err := db.DB.Exec(query, args...); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
}

// renewMoveFixture builds the ledger of variant v under a temp root and returns the project root.
func renewMoveFixture(t *testing.T, v renewMoveVariant) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	root, _ := fcFixture(t)
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	renewMoveAssertIsolated(t, root)

	seedLiveRun(t, root, "run-live")
	func() {
		db, err := homestate.OpenFactory(root)
		if err != nil {
			t.Fatalf("open factory: %v", err)
		}
		defer func() { _ = db.Close() }()
		if err := db.RecordRun(context.Background(), homestate.FactoryRun{
			RunID: "run-dead", Backend: "claude", ManifestJSON: "{}",
			LeadPID: renewMoveDeadPID, LeadProcessStart: "gone-before-resume",
			LaneCapacity: homestate.LaneCapacityDerived,
		}); err != nil {
			t.Fatalf("record run-dead: %v", err)
		}
		if v.deadStatus != "active" {
			renewMoveExec(t, db, `UPDATE runs SET status=? WHERE run_id='run-dead'`, v.deadStatus)
		}
		if v.deadWorker > 0 {
			renewMoveExec(t, db, `INSERT INTO workers(label,pid,registered_at,heartbeat_at,run_id) VALUES('lane-9',?,'2026-09-28T00:00:00Z','2026-09-28T00:00:00Z','run-dead')`, v.deadWorker)
		}
	}()
	sdRegisterLane(t, root, "lane-1")
	func() {
		db, err := homestate.OpenFactory(root)
		if err != nil {
			t.Fatalf("open factory: %v", err)
		}
		defer func() { _ = db.Close() }()
		renewMoveExec(t, db, `UPDATE workers SET pid=? WHERE label='lane-1'`, v.lanePID)
	}()
	if v.laneBound != "" {
		factoryRenewBindLane(t, root, "lane-1", v.laneBound)
	}

	live := fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano)
	t1Holder := v.t1Holder
	if t1Holder == "" {
		t1Holder = "lane-1"
	}
	fcPlace(t, root,
		homestate.Card{RunID: "run-live", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
			OwnerLabel: t1Holder, LeaseHolder: t1Holder, LeaseExpiresAt: live, WorktreePath: root},
		homestate.Card{RunID: "run-live", CardID: "tb", State: homestate.CardSync, Stage: homestate.CardSync,
			OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root},
	)
	for _, c := range v.deadCards {
		c.RunID = "run-dead"
		fcPlace(t, root, c)
	}
	sdLaneEnv(t, "lane-1", "")
	return root
}

// renewMoveSnap is every row of every table a renew could plausibly touch, one text line per row.
type renewMoveSnap struct {
	cards, runs, workers, dispatch, events []string
}

// renewMoveCardLine renders every column of the row that is not the zero value (a zero column
// carries no information), so two lines are equal exactly when the rows are equal in every column.
func renewMoveCardLine(c homestate.Card) string {
	v := reflect.ValueOf(c)
	typ := v.Type()
	parts := make([]string, 0, typ.NumField())
	for i := range typ.NumField() {
		if v.Field(i).IsZero() {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", typ.Field(i).Name, v.Field(i).Interface()))
	}
	return strings.Join(parts, " ")
}

func renewMoveRows(t *testing.T, db *homestate.FactoryDB, query string, cols int) []string {
	t.Helper()
	rows, err := db.DB.Query(query)
	if err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		vals := make([]any, cols)
		ptrs := make([]any, cols)
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatal(err)
		}
		parts := make([]string, cols)
		for i, v := range vals {
			parts[i] = fmt.Sprint(v)
		}
		out = append(out, strings.Join(parts, "|"))
	}
	return out
}

func renewMoveSnapshot(t *testing.T, root string) renewMoveSnap {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var s renewMoveSnap
	cards, err := db.ListCards(context.Background(), "")
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	for _, c := range cards {
		s.cards = append(s.cards, renewMoveCardLine(c))
	}
	s.runs = renewMoveRows(t, db, `SELECT run_id,status,lead_pid,lead_process_start,updated_at FROM runs ORDER BY run_id`, 5)
	s.workers = renewMoveRows(t, db, `SELECT label,pid,run_id,heartbeat_at FROM workers ORDER BY label`, 4)
	s.dispatch = renewMoveRows(t, db, `SELECT card_id,run_id FROM card_dispatch ORDER BY card_id`, 2)
	s.events = renewMoveRows(t, db, `SELECT run_id,kind,payload_json FROM events ORDER BY rowid`, 3)
	return s
}

// renewMoveDiff returns "table - line" / "table + line" for rows present on one side only.
func renewMoveDiff(table string, before, after []string) []string {
	count := map[string]int{}
	for _, l := range before {
		count[l]++
	}
	for _, l := range after {
		count[l]--
	}
	var out []string
	for l, n := range count {
		for ; n > 0; n-- {
			out = append(out, table+" - "+l)
		}
		for ; n < 0; n++ {
			out = append(out, table+" + "+l)
		}
	}
	sort.Strings(out)
	return out
}

// renewMoveAnyDiff lists every row that differs between two snapshots, in any table.
func renewMoveAnyDiff(before, after renewMoveSnap) []string {
	var out []string
	out = append(out, renewMoveDiff("cards", before.cards, after.cards)...)
	out = append(out, renewMoveDiff("runs", before.runs, after.runs)...)
	out = append(out, renewMoveDiff("workers", before.workers, after.workers)...)
	out = append(out, renewMoveDiff("card_dispatch", before.dispatch, after.dispatch)...)
	out = append(out, renewMoveDiff("events", before.events, after.events)...)
	return out
}

// renewMoveStrip drops the columns a renewal legitimately rewrites on its own card.
func renewMoveStrip(line string) string {
	var keep []string
	for _, tok := range strings.Split(line, " ") {
		switch {
		case strings.HasPrefix(tok, "LeaseExpiresAt="), strings.HasPrefix(tok, "HeartbeatAt="), strings.HasPrefix(tok, "UpdatedAt="):
		default:
			keep = append(keep, tok)
		}
	}
	return strings.Join(keep, " ")
}

func renewMoveCardKey(line string) string {
	f := strings.SplitN(line, " ", 3)
	return strings.TrimPrefix(f[0], "RunID=") + "/" + strings.TrimPrefix(f[1], "CardID=")
}

func renewMoveBrief(line string) string {
	if line == "" {
		return "(absent)"
	}
	var keep []string
	for _, tok := range strings.Split(line, " ") {
		if strings.HasPrefix(tok, "State=") || strings.HasPrefix(tok, "Version=") {
			keep = append(keep, tok)
		}
	}
	return strings.Join(keep, " ")
}

// renewMoveForeign returns every change other than the lane's own renewal of run-live/<renewed>:
// the three lease columns of that one card row and the heartbeat of lane-1's roster row.
func renewMoveForeign(before, after renewMoveSnap, renewed string) []string {
	var out []string
	bm, am := map[string]string{}, map[string]string{}
	keys := map[string]bool{}
	for _, l := range before.cards {
		bm[renewMoveCardKey(l)] = l
		keys[renewMoveCardKey(l)] = true
	}
	for _, l := range after.cards {
		am[renewMoveCardKey(l)] = l
		keys[renewMoveCardKey(l)] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	for _, k := range sorted {
		b, a := bm[k], am[k]
		if b == a {
			continue
		}
		if k == "run-live/"+renewed && b != "" && a != "" && renewMoveStrip(b) == renewMoveStrip(a) {
			continue
		}
		out = append(out, fmt.Sprintf("card %s: %s -> %s", k, renewMoveBrief(b), renewMoveBrief(a)))
	}
	strip := func(w string) string { return w[:strings.LastIndex(w, "|")] }
	wb, wa := map[string]string{}, map[string]string{}
	for _, w := range before.workers {
		wb[strings.SplitN(w, "|", 2)[0]] = w
	}
	for _, w := range after.workers {
		wa[strings.SplitN(w, "|", 2)[0]] = w
	}
	for _, label := range []string{"lane-1", "lane-2", "lane-9"} {
		b, a := wb[label], wa[label]
		if b == a || (label == "lane-1" && b != "" && a != "" && strip(b) == strip(a)) {
			continue
		}
		out = append(out, fmt.Sprintf("worker %s: %q -> %q", label, b, a))
	}
	out = append(out, renewMoveDiff("runs", before.runs, after.runs)...)
	out = append(out, renewMoveDiff("card_dispatch", before.dispatch, after.dispatch)...)
	out = append(out, renewMoveDiff("events", before.events, after.events)...)
	return out
}

// renewMoveDeadCards is the dead run's cards: expired leases in four lease-holding states, one card
// with no lease, one terminal card, and (lived) one card whose lease is still valid.
func renewMoveDeadCards(lived bool) []homestate.Card {
	past := time.Now().Add(-time.Hour).Format(time.RFC3339Nano)
	cards := []homestate.Card{
		{CardID: "d-run", State: homestate.CardRun, Stage: homestate.CardRun, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: past},
		{CardID: "d-sync", State: homestate.CardSync, Stage: homestate.CardSync, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: past},
		{CardID: "d-mr", State: homestate.CardMergeReady, Stage: homestate.CardMergeReady, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: past},
		{CardID: "d-merging", State: homestate.CardMerging, Stage: homestate.CardMergeReady, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: past},
		{CardID: "d-assigned", State: homestate.CardAssigned, OwnerLabel: "lane-9"},
		{CardID: "d-done", State: homestate.CardDone, Stage: homestate.CardSync, OwnerLabel: "lane-9"},
	}
	if lived {
		cards = append(cards, homestate.Card{CardID: "d-live", State: homestate.CardRun, Stage: homestate.CardRun,
			OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339Nano)})
	}
	return cards
}

// renewMoveT1 is a t1 row of the dead run. With a lease it diverges from run-live's t1 (a refused
// consolidation); without one it is a structural duplicate (abandoned with a pointer).
func renewMoveT1(leased bool) []homestate.Card {
	c := homestate.Card{CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun, OwnerLabel: "lane-1"}
	if leased {
		c.LeaseHolder, c.LeaseExpiresAt = "lane-1", time.Now().Add(-time.Hour).Format(time.RFC3339Nano)
	}
	return []homestate.Card{c}
}

// renewMoveCheck runs one variant: snapshot, bare renew through the in-process CLI entry, snapshot,
// and the "nothing but the lane's own renewal changed" assertion.
func renewMoveCheck(t *testing.T, v renewMoveVariant) {
	root := renewMoveFixture(t, v)
	card := v.renewCard
	if card == "" {
		card = "t1"
	}
	before := renewMoveSnapshot(t, root)
	out, _, err := runFactory(t, "renew", card, "--json")
	after := renewMoveSnapshot(t, root)
	renewMoveAssertIsolated(t, root)

	if v.refused {
		if err == nil {
			t.Errorf("renew was expected to be refused, got stdout %q", strings.TrimSpace(out))
		}
		// A refusal writes nothing (factoryRenewCard's contract): no row of any table differs.
		if diff := renewMoveAnyDiff(before, after); len(diff) > 0 {
			t.Errorf("a refused bare renew (%v) changed %d rows:\n  %s", err, len(diff), strings.Join(renewMoveBrief2(diff), "\n  "))
		}
		return
	}
	if err != nil || !strings.Contains(out, `"result":"renewed"`) {
		t.Fatalf("renew = %q, err %v, want renewed", strings.TrimSpace(out), err)
	}
	if foreign := renewMoveForeign(before, after, card); len(foreign) > 0 {
		t.Errorf("a bare renew of %s changed %d rows beyond its own card:\n  %s", card, len(foreign), strings.Join(foreign, "\n  "))
	}
}

// renewMoveBrief2 shortens diff lines for failure output; the row identity stays in the first 150 bytes.
func renewMoveBrief2(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		if len(l) > 150 {
			l = l[:150] + "..."
		}
		out[i] = l
	}
	return out
}

// TestRenewWithoutRunMovesOnlyItsOwnCard pins the shapes where a run-less renew touches nothing but
// its own card row (lease expiry, heartbeat, updated_at) and the lane's roster heartbeat: the dead
// run's row is not active, or it is active but a consolidation guard refuses (a live lease, a live
// worker, a diverging duplicate). A refused renew leaves every table byte-equal.
func TestRenewWithoutRunMovesOnlyItsOwnCard(t *testing.T) {
	pid := os.Getpid()
	for _, v := range []renewMoveVariant{
		{name: "dead run row retired, lane bound to the live run", deadStatus: "retired", deadCards: renewMoveDeadCards(false), laneBound: "run-live", lanePID: pid},
		{name: "dead run row retired, lane bound only to it, refused", deadStatus: "retired", deadCards: renewMoveDeadCards(false), laneBound: "run-dead", lanePID: pid, refused: true},
		{name: "dead run row retired, card t1 held in both runs", deadStatus: "retired", deadCards: renewMoveT1(true), laneBound: "run-live", lanePID: pid},
		{name: "dead-owner run active, one card has a valid lease, consolidation refuses", deadStatus: "active", deadCards: renewMoveDeadCards(true), laneBound: "run-live", lanePID: pid},
		{name: "dead-owner run active, lane bound to it with a live worker pid, refused", deadStatus: "active", deadCards: renewMoveDeadCards(false), laneBound: "run-dead", lanePID: pid, refused: true},
		{name: "dead-owner run active, card t1 in both runs diverging, consolidation refuses", deadStatus: "active", deadCards: renewMoveT1(true), laneBound: "run-live", lanePID: pid},
	} {
		t.Run(v.name, func(t *testing.T) { renewMoveCheck(t, v) })
	}
}

// TestRenewWithoutRunLeavesDeadOwnerRunRowsAlone is RED. When a dead-owner run is still an ACTIVE row
// next to the live run, the run-less renew resolves its run through ReconcileActiveRuns, which
// retires the dead run and re-keys its non-terminal cards into the live run (RebindRun), deletes
// dead worker rows, and abandons a duplicate card row. The renew never asked for any of that, and it
// does it before it checks the holder or the card, so a refused renew still moved the rows.
func TestRenewWithoutRunLeavesDeadOwnerRunRowsAlone(t *testing.T) {
	pid := os.Getpid()
	for _, v := range []renewMoveVariant{
		{name: "expired cards of the dead-owner run move to the live run", deadStatus: "active", deadCards: renewMoveDeadCards(false), laneBound: "run-live", lanePID: pid},
		{name: "no roster binding, same move", deadStatus: "active", deadCards: renewMoveDeadCards(false), laneBound: "", lanePID: pid},
		{name: "dead worker row of the dead run is deleted", deadStatus: "active", deadCards: renewMoveDeadCards(false), deadWorker: renewMoveDeadPID, laneBound: "run-live", lanePID: pid},
		{name: "structural duplicate of t1 in the dead run is abandoned", deadStatus: "active", deadCards: renewMoveT1(false), laneBound: "run-live", lanePID: pid},
		// Card t2030: the expectation was "renewed", which held only while the reconciliation deleted the lane's
		// own worker row (the defect named here), leaving the lane looking unbound. With the row kept, the lane
		// is bound to run-dead and renewing in run-live is the binding mismatch that
		// TestFactoryRenewSoleRunExplicitMismatchRefuses pins as a refusal (F11), so the renew refuses naming
		// --run and every table, the lane's own roster row included, stays byte-equal.
		{name: "lane bound to the dead run with a dead recorded pid keeps its own roster row, renew refused", deadStatus: "active", deadCards: renewMoveDeadCards(false), laneBound: "run-dead", lanePID: renewMoveDeadPID, refused: true},
		{name: "refused renew, lane does not hold t1, still moves the rows", deadStatus: "active", deadCards: renewMoveDeadCards(false), laneBound: "run-live", lanePID: pid, t1Holder: "lane-2", refused: true},
		{name: "refused renew, unknown card, still moves the rows", deadStatus: "active", deadCards: renewMoveDeadCards(false), laneBound: "run-live", lanePID: pid, renewCard: "t-nope", refused: true},
	} {
		t.Run(v.name, func(t *testing.T) { renewMoveCheck(t, v) })
	}
}
