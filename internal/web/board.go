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
	"time"
)

// boardStageOrder is the card flow; a stage outside it is kept as written after
// these, never folded into a known one.
var boardStageOrder = []string{"READY", "BUILDING", "REVIEWING", "MERGING", "DONE"}

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

type BoardFlowVM struct{ Known bool }

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
		return prGap
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

	by := map[string][]BoardCardVM{}
	for _, c := range cards {
		by[c.Stage] = append(by[c.Stage], BoardCardVM{
			ID: c.ID, Title: refineText(c.Title), Sub: refineText(c.SubStage),
			Parked: c.Parked, ParkReason: refineText(c.ParkReason),
			Priority: itoa(c.Priority), Reclaims: itoa(c.ReclaimCount),
			Redacted: anyMasked(c.Title, c.SubStage, c.ParkReason),
		})
	}
	var extra []string
	known := map[string]bool{}
	for _, s := range boardStageOrder {
		known[s] = true
	}
	for s := range by {
		if !known[s] {
			extra = append(extra, s)
		}
	}
	sort.Strings(extra)
	for _, s := range append(append([]string{}, boardStageOrder...), extra...) {
		if len(by[s]) > 0 {
			vm.Stages = append(vm.Stages, BoardStageVM{Stage: refineText(s), Cards: by[s]})
		}
	}

	for _, l := range leases {
		vm.Leases = append(vm.Leases, BoardLeaseVM{
			CardID: l.CardID, Session: refineText(l.SessionID), Epoch: itoa(int(l.Epoch)),
			Worktree:  refineText(l.Worktree),
			Heartbeat: boardTime(l.HeartbeatAt, "2006-01-02 15:04:05Z"),
			NextWake:  boardTime(l.NextWakeAt, "2006-01-02 15:04Z"),
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
		fail := prGap
		if q.FailReason != "" {
			fail = refineText(q.FailReason)
		}
		vm.Queue = append(vm.Queue, BoardQueueVM{
			CardID: q.CardID, State: refineText(q.State), Candidate: refineText(sha),
			FailReason: fail, Redacted: anyMasked(q.FailReason, q.State),
		})
	}

	for _, n := range notes {
		nv := BoardNoteVM{Seq: n.Seq, Kind: refineText(n.Kind), CardID: n.CardID,
			At: boardTime(n.At, "2006-01-02 15:04Z"), Text: refineText(n.Text), Redacted: anyMasked(n.Text)}
		if n.Kind == "landed" {
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
