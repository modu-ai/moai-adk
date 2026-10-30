// board.go — the read-only local board (Issue #1948, local-only redesign).
// It renders one BoardReader and nothing else: cards by stage, leases, conflict
// keys, the merge queue and notes. It writes nothing. Flow metrics belong to
// lane 20's package and are shown as not provided until that is wired.
package web

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// boardGap marks a value the ledger did not give.
const boardGap = "—"

// boardStageOrder is the card flow, named by core's constants; a stage outside it
// is kept as written after these, never folded into a known one.
var boardStageOrder = []core.Stage{core.StageReady, core.StageBuilding, core.StageReviewing, core.StageMerging, core.StageDone, core.StageCancelled}

type BoardCardVM struct {
	ID, Title, Sub, ParkReason, Priority, Reclaims string
	Parked, Redacted                               bool
}

type BoardStageVM struct {
	Stage string
	Cards []BoardCardVM
}

type BoardLeaseVM struct {
	CardID, Session, Epoch, Worktree, Heartbeat, NextWake, Expires string
	Expired, Redacted                                              bool
}

type BoardKeyVM struct{ Key, CardID string }

type BoardQueueVM struct {
	CardID, State, Candidate, FailReason string
	Redacted                             bool
}

type BoardNoteVM struct {
	Seq              int64
	Kind, CardID, At string
	Text             string
	Redacted         bool
}

// BoardGroupVM is one flow group: how many cards sit in it and how long the
// longest has been there.
type BoardGroupVM struct{ Key, Count, Oldest string }

// BoardStalledVM is one lease whose holder shows no sign of life.
type BoardStalledVM struct{ CardID, HeartbeatAge, Liveness string }

// BoardFlowVM is the flow read model of the same ledger (monitor/flow). Known is
// false when it was not measured; nothing in it is then to be read as zero.
type BoardFlowVM struct {
	Known   bool
	Note    string
	WIP     string
	Stages  []BoardGroupVM
	Waiting []BoardGroupVM
	Stalled []BoardStalledVM
}

type BoardVM struct {
	Available bool
	Note      string
	Stages    []BoardStageVM
	Leases    []BoardLeaseVM
	Keys      []BoardKeyVM
	Queue     []BoardQueueVM
	Landed    []BoardNoteVM
	Notes     []BoardNoteVM
	Flow      BoardFlowVM
}

func boardTime(t time.Time, layout string) string {
	if t.IsZero() {
		return boardGap
	}
	return t.UTC().Format(layout)
}

func (a *app) buildBoard(ctx context.Context, now time.Time) BoardVM {
	r := a.cfg.Ledger
	if r == nil {
		return BoardVM{Note: "ledger reader not provided"}
	}
	cards, err1 := r.Cards(ctx)
	leases, err2 := r.Leases(ctx)
	queue, err3 := r.Queue(ctx)
	notes, err4 := r.Notes(ctx, 0)
	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return BoardVM{Note: "ledger unreadable: " + refineText(err.Error())}
		}
	}
	vm := BoardVM{Available: true}

	by := map[core.Stage][]BoardCardVM{}
	for _, c := range cards {
		pos := c.Position
		var parkReason string
		if pos.Park != nil {
			parkReason = string(pos.Park.Reason)
		}
		by[pos.Stage] = append(by[pos.Stage], BoardCardVM{
			ID: c.ID, Title: refineText(c.Title), Sub: refineText(string(pos.Sub)),
			Parked: pos.Park != nil, ParkReason: refineText(parkReason),
			Priority: itoa(c.Priority), Reclaims: itoa(c.ReclaimCount),
			Redacted: anyMasked(c.Title, string(pos.Sub), parkReason),
		})
	}
	var extra []core.Stage
	known := map[core.Stage]bool{}
	for _, s := range boardStageOrder {
		known[s] = true
	}
	for s := range by {
		if !known[s] {
			extra = append(extra, s)
		}
	}
	sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })
	for _, s := range append(append([]core.Stage{}, boardStageOrder...), extra...) {
		if len(by[s]) > 0 {
			vm.Stages = append(vm.Stages, BoardStageVM{Stage: refineText(string(s)), Cards: by[s]})
		}
	}

	for _, l := range leases {
		var nextWake time.Time
		if l.NextWakeAt != nil {
			nextWake = *l.NextWakeAt
		}
		vm.Leases = append(vm.Leases, BoardLeaseVM{
			CardID: l.CardID, Session: refineText(l.SessionID), Epoch: strconv.FormatInt(l.Epoch, 10),
			Worktree:  refineText(l.Worktree),
			Heartbeat: boardTime(l.HeartbeatAt, "2006-01-02 15:04:05Z"),
			NextWake:  boardTime(nextWake, "2006-01-02 15:04Z"),
			Expires:   boardTime(l.ExpiresAt, "2006-01-02 15:04Z"),
			Expired:   !l.ExpiresAt.IsZero() && l.ExpiresAt.Before(now),
			Redacted:  anyMasked(l.SessionID, l.Worktree),
		})
		for _, k := range l.Keys {
			vm.Keys = append(vm.Keys, BoardKeyVM{Key: refineText(k), CardID: l.CardID})
		}
	}

	sort.SliceStable(queue, func(i, j int) bool { return queue[i].EnqueuedAt.Before(queue[j].EnqueuedAt) })
	for _, q := range queue {
		sha := q.CandidateSHA
		if len(sha) > 8 {
			sha = sha[:8]
		}
		fail := boardGap
		if q.FailReason != "" {
			fail = refineText(q.FailReason)
		}
		vm.Queue = append(vm.Queue, BoardQueueVM{
			CardID: q.CardID, State: refineText(string(q.State)), Candidate: refineText(sha),
			FailReason: fail, Redacted: anyMasked(q.FailReason, string(q.State)),
		})
	}

	for _, n := range notes {
		nv := BoardNoteVM{Seq: n.Seq, Kind: refineText(string(n.Kind)), CardID: n.CardID,
			At: boardTime(n.At, "2006-01-02 15:04Z"), Text: refineText(n.Text), Redacted: anyMasked(n.Text)}
		if n.Kind == ledger.NoteLanded {
			vm.Landed = append(vm.Landed, nv)
		} else {
			vm.Notes = append(vm.Notes, nv)
		}
	}
	return vm
}

func (a *app) handleBoard(w http.ResponseWriter, r *http.Request) {
	if !a.readOnly(w, r) {
		return
	}
	vm := a.shellVM(r, "board", "Board", "local ledger")
	a.renderPage(w, Board(vm, a.buildBoard(r.Context(), time.Now())))
}

func (a *app) handleBoardAPI(w http.ResponseWriter, r *http.Request) {
	if !a.readOnly(w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a.buildBoard(r.Context(), time.Now()))
}
