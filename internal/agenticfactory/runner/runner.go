// Package runner drives one Issue along the minimum execution path of
// SPEC-AGENTIC-FACTORY-001 design §2.2 (Issue #1948):
//
//	claimed -> implementing -> verifying -> ready -> landing -> done
//
// blocked and on-hold are pauses, not stages: they record the stage they left
// and resume only to it. Every step goes through claim.Store.Update, so a
// write by anyone who does not hold the claim at the recorded epoch is refused
// and leaves the record untouched. No step is skipped, and no evidence bound to
// a different HEAD is accepted.
package runner

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Verdict values a result may carry. Anything else (skipped, cancelled,
// missing, unreadable, stale) is not a pass.
const (
	VerdictPass         = "PASS"
	VerdictPassWithDebt = "PASS-WITH-DEBT"
	VerdictFail         = "FAIL"
)

// Kind is the evidence category a Result belongs to.
type Kind string

const (
	KindCheck              Kind = "check"
	KindReview             Kind = "review"
	KindAudit              Kind = "audit"
	KindHandoffWriter      Kind = "handoff-writer"
	KindHandoffCoordinator Kind = "handoff-coordinator"
)

var requiredKinds = []Kind{KindCheck, KindReview, KindAudit, KindHandoffWriter, KindHandoffCoordinator}

// Result is one observed outcome tied to the HEAD it judged.
type Result struct {
	Kind    Kind
	Name    string
	Head    string
	Verdict string
}

// Verification is the evidence set for one candidate HEAD.
type Verification struct {
	Head    string
	Results []Result
}

// Landing is what must be observed before done (REQ-AF-002).
type Landing struct {
	RemoteMerged      bool
	LocalSyncRecorded bool
	IssueClosed       bool
	Deploy            string // evidence path, or "N/A: <reason>"
}

// Ticket names a claim: issue, holder and the epoch the holder was admitted at.
type Ticket struct {
	Issue  core.IssueRef
	Holder string
	Epoch  int64
}

// Runner applies transitions to claims in Store.
type Runner struct{ Store *claim.Store }

// New returns a Runner over st.
func New(st *claim.Store) *Runner { return &Runner{Store: st} }

// Claim admits holder for the Issue (queued -> claimed).
func (r *Runner) Claim(i core.IssueRef, holder, dispatch string, run core.RunID) (Ticket, error) {
	rec, err := r.Store.Acquire(i, holder, dispatch, run)
	if err != nil {
		return Ticket{}, err
	}
	return Ticket{Issue: i, Holder: holder, Epoch: rec.Epoch}, nil
}

// ClaimAs admits a registered agent: the claim stores its session id and its
// display name separately.
func (r *Runner) ClaimAs(i core.IssueRef, a core.Agent, dispatch string, run core.RunID) (Ticket, error) {
	rec, err := r.Store.AcquireFor(i, a, dispatch, run)
	if err != nil {
		return Ticket{}, err
	}
	return Ticket{Issue: i, Holder: a.SessionID, Epoch: rec.Epoch}, nil
}

// step applies fn when the record is in state from; otherwise it refuses.
func (r *Runner) step(t Ticket, from core.State, fn func(*claim.Record) error) (claim.Record, error) {
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		if rec.State != from {
			return fmt.Errorf("%w: %s -> expected %s", core.ErrTransitionInvalid, rec.State, from)
		}
		return fn(rec)
	})
}

// StartAttempt moves claimed -> implementing. The command must have reached
// executed; a sent or received command is not a running attempt.
func (r *Runner) StartAttempt(t Ticket, attempt string, delivery core.Delivery) (claim.Record, error) {
	id, err := core.ParseAttemptID(attempt)
	if err != nil {
		return claim.Record{}, err
	}
	return r.step(t, core.StateClaimed, func(rec *claim.Record) error {
		// An Issue that came through intake may not start until its worktree
		// has the preparer's READY receipt.
		if err := r.Store.ReadyGate(rec.Issue); err != nil {
			return err
		}
		if !delivery.AtLeast(core.DeliveryExecuted) {
			return fmt.Errorf("%w: delivery %q has not reached executed", core.ErrTransitionInvalid, delivery)
		}
		rec.State, rec.Attempt, rec.Delivery = core.StateImplementing, id, delivery
		rec.AttemptSeq++
		return nil
	})
}

// Submit moves implementing -> verifying with a candidate HEAD. A dirty work
// tree is refused: the HEAD would not describe what was built.
func (r *Runner) Submit(t Ticket, head string, dirty bool) (claim.Record, error) {
	h, err := core.ParseHeadSHA(head)
	if err != nil {
		return claim.Record{}, err
	}
	return r.step(t, core.StateImplementing, func(rec *claim.Record) error {
		if dirty {
			return fmt.Errorf("%w: work tree is dirty", core.ErrTransitionInvalid)
		}
		rec.State, rec.Head = core.StateVerifying, h
		return nil
	})
}

// Verify judges the evidence set against the candidate HEAD. Stale or
// incomplete evidence is an error and changes nothing. A FAIL sends the Issue
// back to implementing; a non-pass pauses it as blocked; all-pass makes it
// ready.
func (r *Runner) Verify(t Ticket, v Verification) (claim.Record, error) {
	return r.step(t, core.StateVerifying, func(rec *claim.Record) error {
		if v.Head != string(rec.Head) {
			return fmt.Errorf("%w: evidence head %.12s, candidate %.12s", core.ErrStaleHead, v.Head, rec.Head)
		}
		seen := map[Kind]bool{}
		for _, res := range v.Results {
			if res.Head != string(rec.Head) {
				return fmt.Errorf("%w: %s %q judged %.12s, candidate %.12s", core.ErrStaleHead, res.Kind, res.Name, res.Head, rec.Head)
			}
			seen[res.Kind] = true
		}
		for _, k := range requiredKinds {
			if !seen[k] {
				return fmt.Errorf("%w: no %s result", core.ErrEvidenceIncomplete, k)
			}
		}
		reason, failed := judge(v.Results)
		switch {
		case failed:
			rec.State = core.StateImplementing
		case reason != "":
			pauseIn(rec, core.StateBlocked, reason, "same-HEAD rerun passes without debt")
		default:
			rec.State = core.StateReady
		}
		return nil
	})
}

// judge returns (reason, false) for a pause, ("", true) when any result is a
// FAIL, and ("", false) when everything passed.
func judge(results []Result) (core.BlockReason, bool) {
	var reason core.BlockReason
	set := func(r core.BlockReason) {
		if reason == "" {
			reason = r
		}
	}
	for _, res := range results {
		switch res.Verdict {
		case VerdictPass:
		case VerdictFail:
			return "", true
		case VerdictPassWithDebt:
			set(core.ReasonDebtPresent)
		default:
			if res.Kind == KindCheck {
				// A skipped, missing, cancelled or stale check outranks a debt verdict.
				reason = core.ReasonCheckNotPassed
			} else {
				set(core.ReasonVerdictNotPass)
			}
		}
	}
	return reason, false
}

// Land moves ready -> landing after re-confirming the HEAD.
func (r *Runner) Land(t Ticket, head string) (claim.Record, error) {
	return r.step(t, core.StateReady, func(rec *claim.Record) error {
		if head != string(rec.Head) {
			return fmt.Errorf("%w: landing %.12s, ready %.12s", core.ErrStaleHead, head, rec.Head)
		}
		rec.State = core.StateLanding
		return nil
	})
}

// Complete moves landing -> done only when every obligation is met: remote
// merge, local sync, a deploy record (or a reasoned N/A) and the Issue close.
// Publishing a PR is none of these. When the merge was observed but another
// obligation is open, the merge fact is kept, the open obligation is named in
// the record (the merged-and-pending state) and done stays unreached.
func (r *Runner) Complete(t Ticket, l Landing) (claim.Record, error) {
	var open error
	rec, err := r.step(t, core.StateLanding, func(rec *claim.Record) error {
		var e error
		if open, e = settle(rec, l); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return rec, err
	}
	return rec, open
}

// settle applies Landing to rec. It returns (pending, nil) when the merge is
// observed but an obligation is open (the record is updated and persisted by
// the caller, who then reports pending), and (nil, err) when nothing may be
// recorded because the merge itself was not observed.
func settle(rec *claim.Record, l Landing) (pending error, err error) {
	name, miss := firstMissing(l)
	if miss == nil {
		rec.State, rec.Deploy, rec.Closed, rec.Merged, rec.Pending = core.StateDone, l.Deploy, true, true, ""
		return nil, nil
	}
	if !l.RemoteMerged {
		return nil, miss
	}
	rec.Merged, rec.Pending = true, name
	return miss, nil
}

// firstMissing names the first open obligation, in the order they are
// observed in practice.
func firstMissing(l Landing) (string, error) {
	switch {
	case !l.RemoteMerged:
		return "remote-merge", fmt.Errorf("%w: remote main merge not observed", core.ErrEvidenceIncomplete)
	case !l.LocalSyncRecorded:
		return "local-sync", fmt.Errorf("%w: local sync not recorded", core.ErrEvidenceIncomplete)
	case !deployOK(l.Deploy):
		return "deploy", fmt.Errorf("%w: deploy record or reasoned N/A missing", core.ErrEvidenceIncomplete)
	case !l.IssueClosed:
		return "issue-close", fmt.Errorf("%w: issue close not recorded", core.ErrEvidenceIncomplete)
	}
	return "", nil
}

func deployOK(d string) bool {
	if rest, ok := strings.CutPrefix(d, "N/A:"); ok {
		return strings.TrimSpace(rest) != ""
	}
	return strings.TrimSpace(d) != "" && d != "N/A"
}

// Pause moves any in-flight stage to blocked or on-hold, recording the stage
// it left, the reason and the resume condition.
func (r *Runner) Pause(t Ticket, to core.State, reason core.BlockReason, resume string) (claim.Record, error) {
	if err := core.ValidatePause(to, reason); err != nil {
		return claim.Record{}, err
	}
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		if !rec.State.IsProgress() || rec.State.IsTerminal() {
			return fmt.Errorf("%w: cannot pause from %s", core.ErrTransitionInvalid, rec.State)
		}
		if strings.TrimSpace(resume) == "" {
			return fmt.Errorf("%w: resume condition is required", core.ErrTransitionInvalid)
		}
		pauseIn(rec, to, reason, resume)
		return nil
	})
}

func pauseIn(rec *claim.Record, to core.State, reason core.BlockReason, resume string) {
	rec.Return, rec.State, rec.Reason, rec.Resume = rec.State, to, reason, resume
}

// Resume leaves a pause for the recorded return stage, and only for it, once
// the caller has observed the resume condition.
func (r *Runner) Resume(t Ticket, conditionMet bool) (claim.Record, error) {
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		if !rec.State.IsPause() {
			return fmt.Errorf("%w: %s is not paused", core.ErrTransitionInvalid, rec.State)
		}
		if !conditionMet {
			return fmt.Errorf("%w: resume condition not met: %s", core.ErrTransitionInvalid, rec.Resume)
		}
		rec.State, rec.Return, rec.Reason, rec.Resume = rec.Return, "", "", ""
		return nil
	})
}

// Handoff is the PR handoff evidence an agent supplies when it publishes a PR.
// BoundaryHead is the branch HEAD observed after the writer stopped; it must
// equal the PR head, which proves nothing was written after publication.
type Handoff struct {
	PR            int
	Head          string
	Branch        string
	Worktree      string
	WriterStopped bool
	BoundaryHead  string
}

// HandOff records the PR receipt and releases the agent's claim in one write,
// so an agent can take a new Issue while the original worktree and branch stay
// untouched. It is refused unless the writer has stopped and the write boundary
// matches the PR head. The Issue stays in landing: a PR is not done.
func (r *Runner) HandOff(t Ticket, h Handoff) (claim.Record, error) {
	head, err := core.ParseHeadSHA(h.Head)
	if err != nil {
		return claim.Record{}, err
	}
	boundary, err := core.ParseHeadSHA(h.BoundaryHead)
	if err != nil {
		return claim.Record{}, err
	}
	switch {
	case h.PR <= 0:
		return claim.Record{}, fmt.Errorf("%w: PR number is required", core.ErrEvidenceIncomplete)
	case h.Branch == "" || h.Worktree == "":
		return claim.Record{}, fmt.Errorf("%w: branch and worktree are required", core.ErrEvidenceIncomplete)
	case !h.WriterStopped:
		return claim.Record{}, fmt.Errorf("%w: writer stop not confirmed", core.ErrEvidenceIncomplete)
	case boundary != head:
		return claim.Record{}, fmt.Errorf("%w: branch moved to %.12s after PR head %.12s", core.ErrStaleHead, boundary, head)
	}
	return r.Store.UpdateRelease(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		if rec.State != core.StateLanding {
			return fmt.Errorf("%w: %s, handoff needs a published PR (landing)", core.ErrTransitionInvalid, rec.State)
		}
		if head != rec.Head {
			return fmt.Errorf("%w: PR head %.12s, candidate %.12s", core.ErrStaleHead, head, rec.Head)
		}
		if rec.Receipt != nil && rec.Receipt.Head == head {
			return fmt.Errorf("%w: PR #%d", core.ErrDuplicateHandoff, rec.Receipt.PR)
		}
		rec.Receipt = &claim.Receipt{
			PR: h.PR, Head: head, Branch: h.Branch, Worktree: h.Worktree,
			Attempt: rec.Attempt, AttemptSeq: rec.AttemptSeq, Agent: rec.AgentName, Session: rec.Holder,
		}
		rec.AddEvent(claim.Event{Kind: "handoff", Agent: rec.AgentName, Session: rec.Holder, Attempt: string(rec.Attempt), Head: string(head), PR: h.PR})
		return nil
	})
}

// ClaimRepair admits the single repair holder of a handed-off Issue, on the
// branch its receipt names. The repair attempt then runs the normal steps.
func (r *Runner) ClaimRepair(i core.IssueRef, holder, dispatch string, run core.RunID, branch string) (Ticket, error) {
	rec, err := r.Store.AcquireRepair(i, holder, dispatch, run, branch)
	if err != nil {
		return Ticket{}, err
	}
	return Ticket{Issue: i, Holder: holder, Epoch: rec.Epoch}, nil
}

// ClaimRepairAs is ClaimRepair for a registered agent.
func (r *Runner) ClaimRepairAs(i core.IssueRef, a core.Agent, dispatch string, run core.RunID, branch string) (Ticket, error) {
	rec, err := r.Store.AcquireRepairFor(i, a, dispatch, run, branch)
	if err != nil {
		return Ticket{}, err
	}
	return Ticket{Issue: i, Holder: a.SessionID, Epoch: rec.Epoch}, nil
}

// CompleteHandedOff is Complete for an Issue whose agent already left: the
// leader or monitor records done once every obligation, the Issue close
// included, is observed. It refuses a record that is still held or never
// handed off.
func (r *Runner) CompleteHandedOff(i core.IssueRef, l Landing) (claim.Record, error) {
	var open error
	rec, err := r.Store.UpdateReleased(i, func(rec *claim.Record) error {
		if rec.Receipt == nil {
			return fmt.Errorf("%w: no PR handoff receipt", core.ErrTransitionInvalid)
		}
		if rec.State != core.StateLanding {
			return fmt.Errorf("%w: %s, expected %s", core.ErrTransitionInvalid, rec.State, core.StateLanding)
		}
		var e error
		open, e = settle(rec, l)
		return e
	})
	if err != nil {
		return rec, err
	}
	return rec, open
}

// Reopen is the E10 edge: a handed-off PR that is not merged goes back from
// landing to verifying (head, base or protection changed, or CI/review needs a
// repair). The merge fact is never reverted, so a merged Issue is refused. The
// next step is the single repair claim (E06).
func (r *Runner) Reopen(i core.IssueRef, reason string) (claim.Record, error) {
	if strings.TrimSpace(reason) == "" {
		return claim.Record{}, fmt.Errorf("%w: a reopen reason is required", core.ErrEvidenceIncomplete)
	}
	return r.Store.UpdateReleased(i, func(rec *claim.Record) error {
		switch {
		case rec.Receipt == nil:
			return fmt.Errorf("%w: no PR handoff receipt", core.ErrTransitionInvalid)
		case rec.State != core.StateLanding:
			return fmt.Errorf("%w: %s, reopen applies to landing", core.ErrTransitionInvalid, rec.State)
		case rec.Merged:
			return fmt.Errorf("%w: merge already observed, not reverted", core.ErrTransitionInvalid)
		}
		rec.State, rec.Pending = core.StateVerifying, ""
		rec.AddEvent(claim.Event{Kind: "reopen", Head: string(rec.Head), PR: rec.Receipt.PR, Note: reason})
		return nil
	})
}
