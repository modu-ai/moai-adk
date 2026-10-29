// Package notify keeps the assigned agent responsible to the end (Issue #1948). It tells the Issue's
// holder when its pull request changes (CI failed, a review, merged, the Issue closed), and when the
// holder's wait record passes its recheck time with no sign of progress it tells the holder again, spends
// one retry of the wait, and at the retry limit escalates to the target the wait names. It never changes
// the owner: reassignment is the leader's decision, and nothing here can make it.
//
// It reads claim records and pull-request views and writes only through the interfaces it is given: the
// notifier, the escalator, the retry of the wait record (the runner's RetryWait, which verifies holder
// and epoch), and its own small per-Issue memory. A forge is read, never written.
package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// Records lists the claim records (claim.Store).
type Records interface {
	List() ([]claim.Record, error)
}

// Waiter spends one retry of a wait record (runner.Runner). The ticket is the record's own holder and
// epoch, so a write for a record that has moved on is refused.
type Waiter interface {
	RetryWait(t runner.Ticket, diagnosis, next string, recheck time.Time) (claim.Record, error)
}

// Repairs records the signal that an open pull request needs a repair (claim.Store.MarkRepairNeeded). It
// writes the signal only: holder, epoch and state stay, so the owner repairs its own PR.
type Repairs interface {
	MarkRepairNeeded(i core.IssueRef, reason string) (claim.Record, error)
}

// PRView is what the host says about an Issue's pull request.
type PRView struct {
	Number      int
	Head        string
	CI          string // "success", "failure", "pending" or ""
	Review      string // "approved", "changes_requested" or ""
	Merged      bool
	IssueClosed bool
}

// PRSource reads the pull request of an Issue. found is false when there is none.
type PRSource interface {
	PR(ctx context.Context, issue int) (view PRView, found bool, err error)
}

// Kind says what a notice is about.
type Kind string

const (
	CIFailed         Kind = "ci-failed"
	ChangesRequested Kind = "review-changes-requested"
	Approved         Kind = "review-approved"
	PRMerged         Kind = "pr-merged"
	IssueClosed      Kind = "issue-closed"
	WaitOverdue      Kind = "wait-overdue"
)

// Notice is one message to the Issue's holder.
type Notice struct {
	Issue   int    `json:"issue"`
	Holder  string `json:"holder"` // the session id of the agent that owns the Issue: the only addressee
	Agent   string `json:"agent"`  // its display name
	Kind    Kind   `json:"kind"`
	Detail  string `json:"detail"`
	Attempt int    `json:"attempt"` // for WaitOverdue: which retry this notice is
}

// Notifier delivers a notice to the holder. An error means it is known not to have been delivered; a
// notice can also be lost silently, which is what the retries are for.
type Notifier interface {
	Notify(ctx context.Context, n Notice) error
}

// Escalation tells the escalation target that the retries ran out. It carries facts; the target decides.
type Escalation struct {
	Issue      int    `json:"issue"`
	Holder     string `json:"holder"`
	To         string `json:"to"`
	Retries    int    `json:"retries"`
	MaxRetries int    `json:"max_retries"`
	Diagnosis  string `json:"diagnosis"`
	NextAction string `json:"next_action"`
}

// Escalator delivers an escalation.
type Escalator interface {
	Escalate(ctx context.Context, e Escalation) error
}

// Service is the resident reminder loop; Tick is one pass.
type Service struct {
	Records   Records
	Waits     Waiter
	PRs       PRSource // optional: without it only waits are watched
	Repairs   Repairs  // optional: where a failed check or a change request is recorded as a repair signal
	Notifier  Notifier
	Escalator Escalator
	// State is the per-Issue memory (what was already notified, the record revision last seen). It survives
	// a restart; a state that cannot be read stops the pass for that Issue rather than guessing.
	State forge.CursorStore
	Now   func() time.Time
	// ReNotifyAfter is how long after a reminder the next recheck is due. It comes from configuration;
	// zero refuses to run.
	ReNotifyAfter time.Duration
}

// Action is one thing a pass did (or declined to do) for one Issue.
type Action struct {
	Issue  int
	What   string // "pr-notice", "reminder", "escalation", "progress-seen", "state-error", ...
	Kind   Kind
	Detail string
	Err    error
}

type prMark struct {
	// RepairKey is the failure the repair signal was last recorded for ("ci:<head>" or "review:<head>").
	RepairKey    string `json:"repair_key,omitempty"`
	CIFailedHead string `json:"ci_failed_head,omitempty"`
	Review       string `json:"review,omitempty"`
	Merged       bool   `json:"merged,omitempty"`
	IssueClosed  bool   `json:"issue_closed,omitempty"`
}

type state struct {
	// Seen is the record revision (Seq) the monitor last saw or wrote. A larger Seq later means someone
	// else wrote the record since: progress.
	Seen int64  `json:"seen_seq,omitempty"`
	PR   prMark `json:"pr"`
	// QuietUntil holds reminders back after progress was seen: the owner is working, and a wait it has
	// not yet updated gets one more period before it is reminded.
	QuietUntil time.Time `json:"quiet_until,omitempty"`
	// EscalationSent is set once the escalation for the current wait was delivered.
	EscalationSent bool `json:"escalation_sent,omitempty"`
}

func stateKey(issue int) string { return "notify/issue-" + strconv.Itoa(issue) }

func (s *Service) load(issue int) (state, error) {
	var st state
	raw, found, err := s.State.LoadCursor(stateKey(issue))
	if err != nil {
		return st, err
	}
	if found {
		if err := json.Unmarshal(raw, &st); err != nil {
			return state{}, fmt.Errorf("notify: the memory of issue-%d is not readable: %w", issue, err)
		}
	}
	return st, nil
}

func (s *Service) save(issue int, st state) error {
	raw, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return s.State.SaveCursor(stateKey(issue), raw)
}

// Tick runs one pass over every record. A failure for one Issue is reported and does not stop the others.
func (s *Service) Tick(ctx context.Context) ([]Action, error) {
	if s.ReNotifyAfter <= 0 || s.Now == nil || s.Records == nil || s.Notifier == nil || s.State == nil {
		return nil, errors.New("notify: ReNotifyAfter (configuration), Now, Records, Notifier and State are required")
	}
	recs, err := s.Records.List()
	if err != nil {
		return nil, fmt.Errorf("notify: the claim records could not be read: %w", err)
	}
	var out []Action
	for _, rec := range recs {
		if rec.Holder == "" { // nobody holds it: there is no one to tell
			continue
		}
		out = append(out, s.issue(ctx, rec)...)
	}
	return out, nil
}

func (s *Service) issue(ctx context.Context, rec claim.Record) []Action {
	st, err := s.load(rec.Issue)
	if err != nil {
		return []Action{{Issue: rec.Issue, What: "state-error", Err: err}}
	}
	var acts []Action
	if s.PRs != nil {
		acts = append(acts, s.prNotices(ctx, &rec, &st)...)
	}
	acts = append(acts, s.wait(ctx, rec, &st)...)
	return acts
}

// prNotices tells the holder about each notable change of the pull request that it was not told about yet.
func (s *Service) prNotices(ctx context.Context, rec *claim.Record, st *state) []Action {
	view, found, err := s.PRs.PR(ctx, rec.Issue)
	if err != nil {
		return []Action{{Issue: rec.Issue, What: "pr-read-error", Err: err}}
	}
	if !found {
		return nil
	}
	acts := s.repairSignal(rec, st, view)
	type event struct {
		kind   Kind
		detail string
		mark   func(*prMark)
	}
	var evs []event
	if view.CI == "failure" && st.PR.CIFailedHead != view.Head {
		evs = append(evs, event{CIFailed, fmt.Sprintf("a required check failed on %.12s of PR #%d", view.Head, view.Number),
			func(m *prMark) { m.CIFailedHead = view.Head }})
	}
	if view.Review != "" && view.Review != st.PR.Review {
		k := Approved
		if view.Review == "changes_requested" {
			k = ChangesRequested
		}
		evs = append(evs, event{k, fmt.Sprintf("review of PR #%d: %s", view.Number, view.Review),
			func(m *prMark) { m.Review = view.Review }})
	}
	if view.Merged && !st.PR.Merged {
		evs = append(evs, event{PRMerged, fmt.Sprintf("PR #%d was merged", view.Number), func(m *prMark) { m.Merged = true }})
	}
	if view.IssueClosed && !st.PR.IssueClosed {
		evs = append(evs, event{IssueClosed, "the Issue was closed", func(m *prMark) { m.IssueClosed = true }})
	}
	for _, ev := range evs {
		n := Notice{Issue: rec.Issue, Holder: rec.Holder, Agent: rec.AgentName, Kind: ev.kind, Detail: ev.detail}
		if err := s.Notifier.Notify(ctx, n); err != nil {
			// Not marked: it is tried again on the next pass, and the later events wait behind it.
			return append(acts, Action{Issue: rec.Issue, What: "pr-notice", Kind: ev.kind, Detail: ev.detail, Err: err})
		}
		ev.mark(&st.PR)
		if err := s.save(rec.Issue, *st); err != nil {
			return append(acts, Action{Issue: rec.Issue, What: "state-error", Kind: ev.kind, Err: err})
		}
		acts = append(acts, Action{Issue: rec.Issue, What: "pr-notice", Kind: ev.kind, Detail: ev.detail})
	}
	return acts
}

// wait carries the reminder and escalation protocol of one wait record.
func (s *Service) wait(ctx context.Context, rec claim.Record, st *state) []Action {
	w := rec.Wait
	if w == nil {
		if st.EscalationSent || st.Seen != 0 || !st.QuietUntil.IsZero() {
			st.EscalationSent, st.Seen, st.QuietUntil = false, 0, time.Time{} // the next wait starts clean
			if err := s.save(rec.Issue, *st); err != nil {
				return []Action{{Issue: rec.Issue, What: "state-error", Err: err}}
			}
		}
		return nil
	}
	if w.Escalated {
		return s.escalate(ctx, rec, *w, st)
	}
	now := s.Now()
	if now.Before(w.RecheckAt) {
		return s.see(rec, st)
	}
	if st.Seen != 0 && rec.Seq != st.Seen { // someone wrote the record since: the owner is working
		st.Seen, st.QuietUntil = rec.Seq, now.Add(s.ReNotifyAfter)
		if err := s.save(rec.Issue, *st); err != nil {
			return []Action{{Issue: rec.Issue, What: "state-error", Err: err}}
		}
		return []Action{{Issue: rec.Issue, What: "progress-seen"}}
	}
	if now.Before(st.QuietUntil) {
		return nil
	}
	return s.remind(ctx, rec, *w, st, now)
}

// see records the revision the monitor saw, without acting.
func (s *Service) see(rec claim.Record, st *state) []Action {
	if st.Seen == rec.Seq {
		return nil
	}
	st.Seen = rec.Seq
	if err := s.save(rec.Issue, *st); err != nil {
		return []Action{{Issue: rec.Issue, What: "state-error", Err: err}}
	}
	return nil
}

func (s *Service) remind(ctx context.Context, rec claim.Record, w core.Wait, st *state, now time.Time) []Action {
	attempt := w.Retries + 1
	n := Notice{
		Issue: rec.Issue, Holder: rec.Holder, Agent: rec.AgentName, Kind: WaitOverdue, Attempt: attempt,
		Detail: fmt.Sprintf("waiting since %s: %s (next: %s); reminder %d of %d", w.RecheckAt.Format(time.RFC3339), w.Diagnosis, w.NextAction, attempt, w.MaxRetries),
	}
	diag := fmt.Sprintf("no progress by %s; the owner was reminded (%d of %d)", w.RecheckAt.Format(time.RFC3339), attempt, w.MaxRetries)
	notifyErr := s.Notifier.Notify(ctx, n)
	if errors.Is(notifyErr, core.ErrBusy) {
		// A held lock is "retry", not a failed delivery: the same reminder is tried on the next pass and
		// no retry of the wait is spent, so contention alone never walks a wait to its escalation.
		return []Action{{Issue: rec.Issue, What: "reminder-deferred", Kind: WaitOverdue, Detail: n.Detail, Err: notifyErr}}
	}
	if notifyErr != nil {
		// A reminder that is known not to have been delivered still spends the retry, so a broken
		// channel ends in an escalation instead of an endless loop.
		diag = fmt.Sprintf("no progress by %s; the reminder could not be delivered (%d of %d): %v", w.RecheckAt.Format(time.RFC3339), attempt, w.MaxRetries, notifyErr)
	}
	ref, _ := core.NewIssueRef(rec.Issue)
	next, err := s.Waits.RetryWait(runner.Ticket{Issue: ref, Holder: rec.Holder, Epoch: rec.Epoch}, diag,
		"wait for the owner's progress; escalate to "+w.Escalate+" when the retries run out", now.Add(s.ReNotifyAfter))
	act := Action{Issue: rec.Issue, What: "reminder", Kind: WaitOverdue, Detail: n.Detail, Err: notifyErr}
	acts := []Action{act}
	switch {
	case err == nil:
	case errors.Is(err, core.ErrRetryLimit):
		// recorded as escalated by the retry that reached the limit
	case errors.Is(err, core.ErrOwnershipUnverified):
		// the record moved (new epoch, released, or another holder) between the read and the write
		return append(acts, Action{Issue: rec.Issue, What: "record-moved", Err: err})
	default:
		return append(acts, Action{Issue: rec.Issue, What: "retry-not-recorded", Err: err})
	}
	st.Seen = next.Seq
	if err := s.save(rec.Issue, *st); err != nil {
		return append(acts, Action{Issue: rec.Issue, What: "state-error", Err: err})
	}
	if next.Wait != nil && next.Wait.Escalated {
		acts = append(acts, s.escalate(ctx, next, *next.Wait, st)...)
	}
	return acts
}

func (s *Service) escalate(ctx context.Context, rec claim.Record, w core.Wait, st *state) []Action {
	if st.EscalationSent {
		return nil
	}
	if s.Escalator == nil {
		return []Action{{Issue: rec.Issue, What: "escalation", Err: errors.New("notify: no escalator configured")}}
	}
	e := Escalation{
		Issue: rec.Issue, Holder: rec.Holder, To: w.Escalate, Retries: w.Retries, MaxRetries: w.MaxRetries,
		Diagnosis: w.Diagnosis, NextAction: w.NextAction,
	}
	if err := s.Escalator.Escalate(ctx, e); err != nil {
		return []Action{{Issue: rec.Issue, What: "escalation", Err: err}} // tried again on the next pass
	}
	st.EscalationSent = true
	if err := s.save(rec.Issue, *st); err != nil {
		return []Action{{Issue: rec.Issue, What: "escalation", Err: fmt.Errorf("delivered, but not remembered: %w", err)}}
	}
	return []Action{{Issue: rec.Issue, What: "escalation", Detail: fmt.Sprintf("%d of %d retries spent; told %s", w.Retries, w.MaxRetries, w.Escalate)}}
}

// repairSignal records, once per failure, that the open pull request needs a repair. The write moves the
// record's revision, so the monitor takes the new revision as its own and does not read it as the owner's
// progress.
func (s *Service) repairSignal(rec *claim.Record, st *state, view PRView) []Action {
	if s.Repairs == nil {
		return nil
	}
	var key, reason string
	switch {
	case view.CI == "failure":
		key, reason = "ci:"+view.Head, fmt.Sprintf("a required check failed on %.12s of PR #%d", view.Head, view.Number)
	case view.Review == "changes_requested":
		key, reason = "review:"+view.Head, fmt.Sprintf("changes were requested on PR #%d", view.Number)
	default:
		return nil
	}
	if st.PR.RepairKey == key {
		return nil
	}
	ref, err := core.NewIssueRef(rec.Issue)
	if err != nil {
		return []Action{{Issue: rec.Issue, What: "repair-signal", Err: err}}
	}
	out, err := s.Repairs.MarkRepairNeeded(ref, reason)
	switch {
	case err == nil:
		if st.Seen == 0 || st.Seen == rec.Seq {
			st.Seen = out.Seq
		}
		*rec = out
	case errors.Is(err, core.ErrTransitionInvalid):
		// there is no open pull request to repair on this record: nothing to signal, and nothing to retry
	default:
		return []Action{{Issue: rec.Issue, What: "repair-signal", Detail: reason, Err: err}}
	}
	st.PR.RepairKey = key
	if err := s.save(rec.Issue, *st); err != nil {
		return []Action{{Issue: rec.Issue, What: "state-error", Err: err}}
	}
	return []Action{{Issue: rec.Issue, What: "repair-signal", Detail: reason}}
}
