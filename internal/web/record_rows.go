// record_rows.go — PR tracking rows filled from the records the claim store
// already persists (Issue #1948, backlog B1). The monitor Collector needs the
// GitHub remote for checks, reviews and the open-PR count, so a console refresh
// would call the network; this source reads only claim.Record values from disk:
// no network, no store creation, no write. A fact the store does not persist is
// left unconfirmed, so the panel shows a gap instead of a guess:
//   - the PR phase is derived only from persisted facts (closed, merged, repair
//     evidence recorded, repair signal); a PR that is merely handed off has no persisted live phase;
//   - the checks and the review verdict are remote facts and stay unconfirmed;
//   - the open-PR count follows the store's own rule (handed off, not done).
//
// The rows are encoded with the monitor's own types and MarshalView, so the one
// decoder in pr_tracking.go reads them like a Collector's output.
package web

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// RecordSource is the read side of claim.Store that the rows need.
type RecordSource interface {
	List() ([]claim.Record, error)
	MaxOpenPRs() int
}

// recordView is the monitor View plus the one thing the monitor Row has no key
// for: the PR URL the receipt records, by Issue number. The key is this
// package's own (producer and decoder both live here); a Collector's JSON does
// not carry it, so that source shows the URL as a gap.
type recordView struct {
	monitor.View
	PRURLs map[string]string `json:"web_pr_urls,omitempty"`
	// RepairBacklog counts open PRs that carry the repair signal or the repair
	// evidence; nil when not counted. Also a web-local key.
	RepairBacklog *int `json:"web_repair_backlog,omitempty"`
	// PositionUnknown holds, by Issue number, the claim store's reason when it
	// could not derive the S46 position of a record. Also a web-local key.
	PositionUnknown map[string]string `json:"web_position_unknown,omitempty"`
}

// RecordRowsFrom builds a MonitorRows source from persisted claim records. A
// nil source yields a source that fails, which the panels render as unavailable.
func RecordRowsFrom(s RecordSource) func(context.Context) ([]byte, error) {
	return func(context.Context) ([]byte, error) {
		if s == nil {
			return nil, errors.New("record source is not configured")
		}
		recs, err := s.List()
		if err != nil {
			return nil, err
		}
		at := time.Now()
		view := monitor.View{Rows: make([]monitor.Row, 0, len(recs))}
		open := 0
		urls := map[string]string{}
		repair := 0
		unknown := map[string]string{}
		for _, r := range recs {
			view.Rows = append(view.Rows, recordRow(r, at))
			if r.PositionUnknown != "" {
				unknown[strconv.Itoa(r.Issue)] = recText(r.PositionUnknown)
			}
			if r.Receipt != nil && r.Receipt.PRURL != "" {
				urls[strconv.Itoa(r.Issue)] = recText(r.Receipt.PRURL)
			}
			if r.Receipt != nil && r.State != core.StateDone {
				open++
				if r.RepairNeeded != "" || r.Repair != nil {
					repair++
				}
			}
		}
		limit := s.MaxOpenPRs()
		view.Backlog = recField(monitor.Backlog{Open: open, Limit: limit, State: monitor.BacklogStateOf(open, limit)}, at, true)
		out := recordView{View: view, PRURLs: urls, RepairBacklog: &repair, PositionUnknown: unknown}
		return json.Marshal(out)
	}
}

func recField[T any](v T, at time.Time, confirmed bool) monitor.Field[T] {
	return monitor.Field[T]{Value: v, Source: monitor.SrcClaimStore, ReadAt: at, Confirmed: confirmed}
}

// recText masks a persisted text with core.Redact. It is not the monitor's own
// mask (which also drops opaque blobs and marks the Field redacted); that
// function is unexported, and this stays until the monitor publishes it.
func recText(s string) string {
	red, _ := core.Redact(s)
	return red
}

func recordRow(r claim.Record, at time.Time) monitor.Row {
	// The cached position is shown as confirmed only while the cache matches the
	// authoritative copy (design §8.1: a stale or blocked mirror confirms
	// nothing it carries), and the S46 detail only when the position could be
	// derived (the claim store marks that with Record.PositionUnknown): such a
	// record says nothing about park, gate, refusal or wait, and an absent value
	// there would be an invented "none".
	fresh := r.Mirror == "" || r.Mirror == "current"
	derived := fresh && r.PositionUnknown == ""
	row := monitor.Row{
		Issue:       r.Issue,
		Agent:       recField(r.AgentName, at, r.AgentName != ""),
		SessionID:   recField(r.Holder, at, r.Holder != ""),
		State:       recField(string(r.State), at, fresh && r.State != ""),
		Stage:       recField(string(r.Stage), at, derived),
		SubStage:    recField(string(r.Sub), at, derived && r.Sub != ""),
		MirrorState: recField(r.Mirror, at, r.Mirror != ""),
		Completed:   fresh && r.State == core.StateDone && r.Merged && r.Closed,
	}
	row.Park = recField[*monitor.ParkInfo](nil, at, derived)
	if p := r.Park; derived && p != nil {
		row.Park.Value = &monitor.ParkInfo{Reason: string(p.Reason), Evidence: recText(p.Evidence), Resume: recText(p.Resume)}
	}
	row.Gate = recField[*monitor.GateInfo](nil, at, derived)
	if g := r.Gate; derived && g != nil {
		row.Gate.Value = &monitor.GateInfo{Reason: string(g.Reason), Detail: recText(g.Detail), Resume: recText(g.Resume), Head: string(g.Head)}
	}
	row.Refusal = recField[*monitor.RefusalInfo](nil, at, derived)
	if f := r.Refusal; derived && f != nil {
		row.Refusal.Value = &monitor.RefusalInfo{Reason: string(f.Reason), Detail: recText(f.Detail)}
	}
	row.Wait = recField[*monitor.WaitInfo](nil, at, derived)
	if w := r.Wait; derived && w != nil {
		row.Wait.Value = &monitor.WaitInfo{
			Owner: w.Owner, NextAction: recText(w.NextAction), RecheckAt: w.RecheckAt, Diagnosis: recText(w.Diagnosis),
			Retries: w.Retries, MaxRetries: w.MaxRetries, EscalateTo: w.Escalate, Escalated: w.Escalated,
		}
	}
	head := string(r.Head)
	if r.Receipt != nil && r.Receipt.Head != "" {
		head = string(r.Receipt.Head)
	}
	row.HeadSHA = recField(head, at, fresh && head != "")
	if rc := r.Receipt; fresh && rc != nil {
		row.PR = recField(monitor.PRStatus{Number: rc.PR, Phase: recordPhase(r)}, at, true)
	}
	return row
}

// recordPhase derives the PR phase from persisted facts only. A handed-off PR
// with none of them has no persisted live phase: it is left empty (a gap).
func recordPhase(r claim.Record) monitor.PRPhase {
	switch {
	case r.Closed:
		return monitor.PhaseIssueClosed
	case r.Merged:
		return monitor.PhaseMergedIssueOpen
	case r.Repair != nil:
		return monitor.PhaseRepairDispatched
	case r.RepairNeeded != "":
		return monitor.PhaseRepairNeeded
	}
	return ""
}
