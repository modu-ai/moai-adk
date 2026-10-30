//go:build afred

package web

// board_test.go — the read-only local board (Issue #1948, local-only redesign,
// contract .moai/reports/issue-1948/local-adk-design-20261030.md §3-§4 lane-21 row).
// The board reads one ledger.Reader-shaped source and only displays: cards by
// stage with park, priority and reclaim count; leases with session, epoch,
// heartbeat and next wake; conflict keys; the merge queue; landed news and other
// notes. Until ledger/api.go exists the tests run against a fake reader with the
// same method set.
//
// Defect answer: these tests catch a card shown under the wrong stage or an
// unknown stage guessed into a known one, a parked card that reads as running, a
// lease without its epoch or next wake, a conflict key shown without its holder, a
// queue entry that hides its failure reason, a landed note lost among others, an
// unwired or failing reader read as an empty board, text that reaches the page
// unrefined, and a board that accepts a write.

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fakeBoardReader struct {
	cards      []LedgerCard
	leases     []LedgerLease
	queue      []LedgerQueueItem
	notes      []LedgerNote
	err        error
	notesSince []int64
}

func (f *fakeBoardReader) Cards(context.Context) ([]LedgerCard, error)      { return f.cards, f.err }
func (f *fakeBoardReader) Leases(context.Context) ([]LedgerLease, error)    { return f.leases, f.err }
func (f *fakeBoardReader) Queue(context.Context) ([]LedgerQueueItem, error) { return f.queue, f.err }
func (f *fakeBoardReader) Notes(_ context.Context, since int64) ([]LedgerNote, error) {
	f.notesSince = append(f.notesSince, since)
	return f.notes, f.err
}

var boardNow = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

func boardFixture() *fakeBoardReader {
	return &fakeBoardReader{
		cards: []LedgerCard{
			{ID: "c1", Title: "ready one", Stage: "READY", Priority: 5},
			{ID: "c2", Title: "building spec", Stage: "BUILDING", SubStage: "SPEC", Priority: 3},
			{ID: "c3", Title: "building fix", Stage: "BUILDING", SubStage: "FIX", Priority: 3, ReclaimCount: 2},
			{ID: "c4", Title: "parked review", Stage: "REVIEWING", Parked: true, ParkReason: "fix-limit", Priority: 1, ReclaimCount: 3},
			{ID: "c5", Title: "in queue", Stage: "MERGING", Priority: 2},
			{ID: "c6", Title: "finished", Stage: "DONE"},
			{ID: "c7", Title: "from a future stage", Stage: "QUARANTINE"},
		},
		leases: []LedgerLease{
			{CardID: "c2", SessionID: "sess-a", Epoch: 4, PID: 111, Worktree: "/work/c2",
				HeartbeatAt: boardNow.Add(-40 * time.Second), NextWakeAt: boardNow.Add(5 * time.Minute), ExpiresAt: boardNow.Add(10 * time.Minute),
				Keys: []string{"pkg:internal/web", "shared:go.mod"}},
			{CardID: "c3", SessionID: "sess-b", Epoch: 2, PID: 222, Worktree: "/work/c3",
				HeartbeatAt: boardNow.Add(-20 * time.Minute), ExpiresAt: boardNow.Add(-5 * time.Minute), Keys: []string{"pkg:internal/cli"}},
		},
		queue: []LedgerQueueItem{
			{CardID: "c5", CandidateSHA: "0123456789abcdef0123456789abcdef01234567", State: "waiting", EnqueuedAt: boardNow.Add(-2 * time.Minute)},
			{CardID: "c8", CandidateSHA: "fedcba9876543210fedcba9876543210fedcba98", State: "failed", FailReason: "vet failed in the candidate worktree", EnqueuedAt: boardNow.Add(-9 * time.Minute)},
		},
		notes: []LedgerNote{
			{Seq: 10, Kind: "landed", CardID: "c6", Text: "c6 landed at 89abcdef", At: boardNow.Add(-1 * time.Hour)},
			{Seq: 11, Kind: "gate_wait", CardID: "c4", Text: "waiting for the operator", At: boardNow.Add(-30 * time.Minute)},
			{Seq: 12, Kind: "stalled", CardID: "c3", Text: "owner liveness unknown", At: boardNow.Add(-10 * time.Minute)},
			{Seq: 13, Kind: "decision", Text: "park after three reclaims", At: boardNow.Add(-5 * time.Minute)},
		},
	}
}

func boardApp(t *testing.T, r BoardReader) *app {
	t.Helper()
	a := newApp(Config{ProjectRoot: t.TempDir(), Ledger: r})
	a.recordLastProfile = func(string) error { return nil }
	return a
}

func TestBoard_CardsGroupedByStageWithParkPriorityReclaim(t *testing.T) {
	vm := boardApp(t, boardFixture()).buildBoard(context.Background(), boardNow)
	if !vm.Available {
		t.Fatalf("board = %+v", vm)
	}
	stages := map[string][]BoardCardVM{}
	var order []string
	for _, g := range vm.Stages {
		stages[g.Stage] = g.Cards
		order = append(order, g.Stage)
	}
	want := []string{"READY", "BUILDING", "REVIEWING", "MERGING", "DONE", "QUARANTINE"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("stage order = %v, want the flow order then unknown stages as written: %v", order, want)
	}
	if len(stages["BUILDING"]) != 2 || stages["BUILDING"][0].Sub != "SPEC" || stages["BUILDING"][1].Sub != "FIX" {
		t.Errorf("BUILDING = %+v", stages["BUILDING"])
	}
	rv := stages["REVIEWING"][0]
	if !rv.Parked || rv.ParkReason != "fix-limit" || rv.Priority != "1" || rv.Reclaims != "3" {
		t.Errorf("a parked card must say it is parked and why: %+v", rv)
	}
	if c := stages["BUILDING"][1]; c.Parked || c.Reclaims != "2" || c.Priority != "3" {
		t.Errorf("c3 = %+v", c)
	}
	if len(stages["QUARANTINE"]) != 1 {
		t.Errorf("an unknown stage is kept as written, not folded into a known one: %v", order)
	}
}

func TestBoard_LeasesShowSessionEpochHeartbeatNextWake(t *testing.T) {
	vm := boardApp(t, boardFixture()).buildBoard(context.Background(), boardNow)
	by := map[string]BoardLeaseVM{}
	for _, l := range vm.Leases {
		by[l.CardID] = l
	}
	l := by["c2"]
	if l.Session != "sess-a" || l.Epoch != "4" || l.Heartbeat != "2026-10-30 11:59:20Z" || l.NextWake != "2026-10-30 12:05Z" || l.Expires != "2026-10-30 12:10Z" || l.Expired {
		t.Errorf("c2 lease = %+v", l)
	}
	if by["c3"].NextWake != prGap || !by["c3"].Expired {
		t.Errorf("a lease with no next wake shows a gap and one past its expiry says so: %+v", by["c3"])
	}
}

func TestBoard_ConflictKeysNameTheirHolder(t *testing.T) {
	vm := boardApp(t, boardFixture()).buildBoard(context.Background(), boardNow)
	holder := map[string]string{}
	for _, k := range vm.Keys {
		holder[k.Key] = k.CardID
	}
	if holder["pkg:internal/web"] != "c2" || holder["shared:go.mod"] != "c2" || holder["pkg:internal/cli"] != "c3" || len(vm.Keys) != 3 {
		t.Errorf("keys = %+v", vm.Keys)
	}
}

func TestBoard_QueueAndNotes(t *testing.T) {
	f := boardFixture()
	vm := boardApp(t, f).buildBoard(context.Background(), boardNow)
	if len(vm.Queue) != 2 || vm.Queue[0].CardID != "c8" || vm.Queue[0].FailReason != "vet failed in the candidate worktree" || vm.Queue[0].Candidate != "fedcba98" {
		t.Errorf("the queue is oldest first and a failure keeps its reason: %+v", vm.Queue)
	}
	if vm.Queue[1].State != "waiting" || vm.Queue[1].FailReason != prGap {
		t.Errorf("queue[1] = %+v", vm.Queue[1])
	}
	if len(vm.Landed) != 1 || vm.Landed[0].CardID != "c6" {
		t.Errorf("landed news = %+v", vm.Landed)
	}
	kinds := map[string]bool{}
	for _, n := range vm.Notes {
		kinds[n.Kind] = true
	}
	if !kinds["gate_wait"] || !kinds["stalled"] || !kinds["decision"] || kinds["landed"] {
		t.Errorf("other notes keep their kind and landed ones are not repeated here: %+v", vm.Notes)
	}
	if len(f.notesSince) != 1 || f.notesSince[0] != 0 {
		t.Errorf("the board reads all notes (since 0), got %v", f.notesSince)
	}
	if vm.Flow.Known {
		t.Errorf("flow metrics are not provided yet and must not be invented: %+v", vm.Flow)
	}
}

func TestBoard_UnwiredOrFailingReaderIsNotAnEmptyBoard(t *testing.T) {
	if vm := boardApp(t, nil).buildBoard(context.Background(), boardNow); vm.Available || !strings.Contains(vm.Note, "not provided") {
		t.Errorf("an unwired reader = %+v", vm)
	}
	bad := &fakeBoardReader{err: errors.New("open /Users/alice/proj/.ledger: busy")}
	vm := boardApp(t, bad).buildBoard(context.Background(), boardNow)
	if vm.Available || !strings.Contains(vm.Note, "unreadable") || len(vm.Stages) != 0 {
		t.Errorf("a failing reader = %+v", vm)
	}
	if strings.Contains(vm.Note, "/Users/alice") {
		t.Errorf("the failure text must be refined: %q", vm.Note)
	}
	if empty := boardApp(t, &fakeBoardReader{}).buildBoard(context.Background(), boardNow); !empty.Available || len(empty.Stages) != 0 {
		t.Errorf("a readable, empty ledger is an available, empty board: %+v", empty)
	}
}

func TestBoard_TextIsRefinedAndMarkedMasked(t *testing.T) {
	p := bannerPayloads
	f := &fakeBoardReader{
		cards:  []LedgerCard{{ID: "c1", Title: "mail " + p["email"], Stage: "READY"}, {ID: "c2", Title: "clean", Stage: "READY"}},
		leases: []LedgerLease{{CardID: "c1", SessionID: "s", Epoch: 1, Worktree: p["home-path"], Keys: []string{"k"}}},
		queue:  []LedgerQueueItem{{CardID: "c1", CandidateSHA: "abc", State: "failed", FailReason: "tok " + p["token"]}},
		notes:  []LedgerNote{{Seq: 1, Kind: "landed", CardID: "c1", Text: "call " + p["phone"]}},
	}
	a := boardApp(t, f)
	vm := a.buildBoard(context.Background(), boardNow)
	if !vm.Stages[0].Cards[0].Redacted || vm.Stages[0].Cards[1].Redacted {
		t.Errorf("the masked card carries the marker, the clean one does not: %+v", vm.Stages[0].Cards)
	}
	html := renderTempl(t, Board(a.shellVM(httptestGet("/board"), "board", "Board", "local ledger"), vm))
	for cat, payload := range map[string]string{"email": p["email"], "home-path": p["home-path"], "token": p["token"], "phone": p["phone"]} {
		if strings.Contains(html, payload) {
			t.Errorf("the %s reached the board page", cat)
		}
	}
	if !strings.Contains(html, "masked") {
		t.Errorf("the page must show that something was masked")
	}
}

func TestBoard_RouteIsReadOnlyAndRendersEveryPanel(t *testing.T) {
	a := boardApp(t, boardFixture())
	rec := serveGet(t, a.routes(), "/board")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /board = %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`id="board-cards"`, `id="board-leases"`, `id="board-keys"`, `id="board-queue"`, `id="board-landed"`, `id="board-flow"`, "fix-limit", "sess-a", "pkg:internal/web", "c6 landed at 89abcdef", "vet failed in the candidate worktree"} {
		if !strings.Contains(body, want) {
			t.Errorf("board page missing %q", want)
		}
	}
	for _, bad := range []string{"<input", "<select", "<button", "<form"} {
		if strings.Contains(body[strings.Index(body, `id="board-cards"`):], bad) {
			t.Errorf("the board must be display-only; found %q", bad)
		}
	}
	if post := servePostEmpty(t, a, "/board"); post != http.StatusMethodNotAllowed {
		t.Errorf("POST /board = %d, want 405", post)
	}
	api := serveGet(t, a.routes(), "/api/board")
	if api.Code != http.StatusOK || !strings.Contains(api.Body.String(), `"c2"`) {
		t.Errorf("GET /api/board = %d", api.Code)
	}
	if post := servePostEmpty(t, a, "/api/board"); post != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/board = %d, want 405", post)
	}
}

func httptestGet(path string) *http.Request {
	r, _ := http.NewRequest(http.MethodGet, path, nil)
	return r
}

func servePostEmpty(t *testing.T, a *app, path string) int {
	t.Helper()
	return servePost(t, a.routes(), path, nil).Code
}
