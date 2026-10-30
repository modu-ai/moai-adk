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
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
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
	KindQA                 Kind = "qa"
	KindHandoffWriter      Kind = "handoff-writer"
	KindHandoffCoordinator Kind = "handoff-coordinator"
)

var requiredKinds = []Kind{KindCheck, KindReview, KindAudit, KindHandoffWriter, KindHandoffCoordinator}

// Result is one observed outcome tied to the HEAD it judged. Session and Epoch
// are what the registering service observed for the writer handoff; they are never
// read from the evidence file, and for the other producer-checked kinds they are
// ignored in favour of the attribution lookup.
type Result struct {
	Kind    Kind
	Name    string
	Head    string
	Verdict string
	Session string
	Epoch   int64
}

// Attribution is the answer of the read-only attribution lookup of
// SPEC-AGENTIC-FACTORY-001's identity sub-SPEC: either who produced a result
// and in which role, or why that cannot be said.
type Attribution struct {
	Attributed bool
	Role       string
	Session    string
	Attempt    string
	Epoch      int64
	Reason     string // set when not attributed
}

// TicketAttributor is an Attributor that needs to know the Issue: the lookup of
// the identity records is per Issue (a grant is for one Issue, a writer is a writer
// of one). The runner prefers it to Attribute when the Attributor has it.
type TicketAttributor interface {
	AttributeFor(t Ticket, res Result) Attribution
}

// Attributor looks up the producer of a result. The implementation belongs to
// the identity SPEC; until it exists the Runner has none, and every
// producer-checked result is unattributable.
type Attributor interface {
	Attribute(res Result) Attribution
}

// roleFor is the role design 11 requires for each producer-checked kind.
var roleFor = map[Kind]string{
	KindReview:             core.RoleReviewer,
	KindAudit:              core.RoleAuditor,
	KindQA:                 core.RoleQA,
	KindHandoffCoordinator: core.RoleCoordinator,
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

// IssueFacts is what the host says about an Issue, for the permission decision:
// the repository asked about and the one answered for (and whether a redirect
// intervened), the authenticated user, the Issue's author, and who owns the work.
// An id that could not be established is passed as 0.
type IssueFacts struct {
	RequestedRepo string
	ActualRepo    string
	Redirected    bool
	UserID        int64
	AuthorID      int64
	Owner         core.OwnerState
}

// Facts looks an Issue up on the host. The implementation lives with the forge
// code; it makes no write. The context is the caller's: cancelling the pass that
// holds it must cancel the lookup, so an implementation derives its own time limit
// from it and never from context.Background().
type Facts interface {
	IssueFacts(ctx context.Context, issue core.IssueRef) (IssueFacts, error)
}

// Runner applies transitions to claims in Store.
//
// Two lookups are injected and both fail closed: with no Attributor every
// review, audit, QA and coordinator result is unattributable and is refused; with
// no Facts nobody's identity is established and no Issue is claimed. Collab holds
// the collaboration settings that decide when another author's Issue may be taken.
type Runner struct {
	Store      *claim.Store
	Attributor Attributor
	Facts      Facts
	Collab     core.Collaboration
	// Ctx is the context of the pass this Runner serves; a pass takes its own copy of
	// the Runner and sets it. Nil means context.Background(). It reaches the host
	// lookups (Facts) and nothing else: store operations stay local and short.
	Ctx context.Context

	// beforeClaimMoves is a test seam: it runs after TakeOver has decided and before
	// the store moves the claim, where the owner could still act.
	beforeClaimMoves func()
}

func (r *Runner) ctx() context.Context {
	if r.Ctx != nil {
		return r.Ctx
	}
	return context.Background()
}

// guard decides whether the acting user may perform op on the Issue, through the
// one permission decision in core. A failed or missing lookup is an unknown
// identity.
func (r *Runner) guard(i core.IssueRef, op core.Op) error {
	if r.Facts == nil {
		return core.DenyIdentityUnknown.Err()
	}
	f, err := r.Facts.IssueFacts(r.ctx(), i)
	if err != nil {
		return fmt.Errorf("%w: host lookup of issue-%d failed: %v", core.ErrIdentityUnknown, i.Number, err)
	}
	return r.Collab.Authorize(core.AuthzRequest{
		Op: op, RequestedRepo: f.RequestedRepo, ActualRepo: f.ActualRepo, Redirected: f.Redirected,
		UserID: f.UserID, IssueAuthorID: f.AuthorID, Owner: f.Owner,
	}).Err()
}

// New returns a Runner over st.
func New(st *claim.Store) *Runner { return &Runner{Store: st} }

// Claim admits holder for the Issue (queued -> claimed).
func (r *Runner) Claim(i core.IssueRef, holder, dispatch string, run core.RunID) (Ticket, error) {
	if err := r.guard(i, core.OpTakeOver); err != nil {
		return Ticket{}, err
	}
	rec, err := r.Store.Acquire(i, holder, dispatch, run)
	if err != nil {
		return Ticket{}, err
	}
	return Ticket{Issue: i, Holder: holder, Epoch: rec.Epoch}, nil
}

// ClaimAs admits a registered agent: the claim stores its session id and its
// display name separately.
func (r *Runner) ClaimAs(i core.IssueRef, a core.Agent, dispatch string, run core.RunID) (Ticket, error) {
	if err := r.guard(i, core.OpTakeOver); err != nil {
		return Ticket{}, err
	}
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
		// An Issue that came through intake may not start until the reserved
		// agent acknowledged the preparer's receipt (READY_ACKED).
		if err := r.Store.ReadyGate(rec.Issue, rec.Holder); err != nil {
			return err
		}
		// Another PR of this agent is open: this Issue is written in its own worktree.
		if rec.Worktree == "" {
			if open, err := r.Store.HasOpenHandoff(rec.Holder, core.IssueRef{Number: rec.Issue}); err != nil {
				return err
			} else if open {
				return fmt.Errorf("%w: issue-%d", core.ErrWorktreeUnbound, rec.Issue)
			}
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
// back to implementing; a non-pass records a merge gate on the verifying stage
// (the state does not change and nothing is paused); all-pass makes it ready.
// Judging the same HEAD with the same evidence records the same gate again and
// nothing else: only a new HEAD or new evidence can turn it into a pass.
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
		if detail := r.producerProblem(t, v.Results); detail != "" {
			rec.Gate = &core.Gate{Reason: core.ReasonProducerNotAllowed, Detail: detail,
				Resume: "a producer in the required role registers the evidence on the same HEAD", Head: rec.Head}
			return nil
		}
		reason, failed := judge(v.Results)
		switch {
		case failed:
			rec.State, rec.Gate = core.StateImplementing, nil
		case reason != "":
			rec.Gate = &core.Gate{Reason: reason, Detail: fmt.Sprintf("verification of %.12s did not pass: %s", rec.Head, reason),
				Resume: "same-HEAD rerun passes without debt", Head: rec.Head}
		default:
			rec.State, rec.Gate = core.StateReady, nil
		}
		return nil
	})
}

// producerProblem returns why some result must not be counted, or "" when every
// producer is allowed. Review, audit, QA and coordinator results need the
// lookup to say "attributed" with the right role; the writer handoff must come
// from the holder and epoch the attempt was issued to. Nothing the evidence says
// about its own producer is used.
func (r *Runner) producerProblem(t Ticket, results []Result) string {
	for _, res := range results {
		switch res.Kind {
		case KindHandoffWriter:
			if res.Session == "" || res.Session != t.Holder || res.Epoch != t.Epoch {
				return fmt.Sprintf("writer handoff %q is not from the issued holder and epoch (%s)", res.Name, core.ReasonProducerNotAllowed)
			}
		case KindReview, KindAudit, KindQA, KindHandoffCoordinator:
			att := Attribution{Reason: core.AttributionLocked}
			if ta, ok := r.Attributor.(TicketAttributor); ok {
				att = ta.AttributeFor(t, res)
			} else if r.Attributor != nil {
				att = r.Attributor.Attribute(res)
			}
			switch {
			case !att.Attributed:
				return fmt.Sprintf("%s %q is unattributable (%s); a producer in the required role must register it on the same HEAD", res.Kind, res.Name, att.Reason)
			case att.Role != roleFor[res.Kind]:
				return fmt.Sprintf("%s %q was produced in role %q, %q is required", res.Kind, res.Name, att.Role, roleFor[res.Kind])
			}
		}
	}
	return ""
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

// Land moves ready -> landing after re-confirming the HEAD and that the PR text
// was checked on it (PreflightPR). Publishing the PR is this step.
func (r *Runner) Land(t Ticket, head string) (claim.Record, error) {
	return r.step(t, core.StateReady, func(rec *claim.Record) error {
		if head != string(rec.Head) {
			return fmt.Errorf("%w: landing %.12s, ready %.12s", core.ErrStaleHead, head, rec.Head)
		}
		if string(rec.PreflightHead) != head {
			return fmt.Errorf("%w: the PR text was not checked for closing keywords on %.12s (PreflightPR)", core.ErrEvidenceIncomplete, head)
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
		rec.Wait, rec.Repair = nil, nil
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
// it left, the reason and the resume condition. A park carries a wait record
// (who answers, the next action, the recheck time, the diagnosis, the retries and
// the escalation target); a pause without a valid one is refused.
func (r *Runner) Pause(t Ticket, to core.State, reason core.BlockReason, resume string, w core.Wait) (claim.Record, error) {
	if err := core.ValidatePause(to, reason); err != nil {
		return claim.Record{}, err
	}
	if !core.IsParkable(reason) {
		// Gate and refusal reasons are records on a stage, not pauses.
		return claim.Record{}, fmt.Errorf("%w: %q cannot pause an Issue", core.ErrReasonNotAllowed, reason)
	}
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		if !rec.State.IsProgress() || rec.State.IsTerminal() {
			return fmt.Errorf("%w: cannot pause from %s", core.ErrTransitionInvalid, rec.State)
		}
		if strings.TrimSpace(resume) == "" {
			return fmt.Errorf("%w: resume condition is required", core.ErrTransitionInvalid)
		}
		pauseIn(rec, to, reason, resume)
		rec.Wait = &w // a park always carries its wait record (design 2.1c); the store checks it
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
	// HandoffID names the receipt; delivering the same id again is a replay and
	// converges on one acceptance. LeaderAccepted is true only when the leader
	// has accepted the receipt: the claim is returned after that, not before.
	HandoffID      string
	LeaderAccepted bool
	PR             int
	Head           string
	Branch         string
	Worktree       string
	WriterStopped  bool
	BoundaryHead   string
	// The rest of the receipt (spec 148), all required of a new handoff: the PR's
	// URL, the base it targets, the spec revision adopted, what the task produced
	// and the state of the checks.
	PRURL       string
	Base        string
	AdoptedSpec string
	TaskResult  string
	CheckStatus string
}

// HandOff records the PR receipt. It does not release the Issue: the agent that
// published the PR stays its owner until the Issue is closed, so it answers CI
// and review itself (ReopenOwn) and finishes the merge (Complete). It may take
// another Issue under the open-PR limit; the original worktree and branch stay
// untouched. It is refused unless the writer has stopped and the write boundary
// matches the PR head. The Issue stays in landing: a PR is not done.
func (r *Runner) HandOff(t Ticket, h Handoff) (claim.Record, error) {
	if _, err := core.ParseRunID(h.HandoffID); err != nil {
		return claim.Record{}, fmt.Errorf("%w: a handoff id is required", core.ErrEvidenceIncomplete)
	}
	if !h.LeaderAccepted {
		return claim.Record{}, fmt.Errorf("%w: the claim is returned only after the leader accepted the receipt", core.ErrEvidenceIncomplete)
	}
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
	case !validPRURL(h.PRURL, h.PR):
		return claim.Record{}, fmt.Errorf("%w: the PR URL is missing, not an http(s) URL, or not the URL of PR #%d", core.ErrEvidenceIncomplete, h.PR)
	case strings.TrimSpace(h.Base) == "" || strings.TrimSpace(h.AdoptedSpec) == "" || strings.TrimSpace(h.TaskResult) == "" || strings.TrimSpace(h.CheckStatus) == "":
		return claim.Record{}, fmt.Errorf("%w: base, adopted spec, task result and check status are required", core.ErrEvidenceIncomplete)
	case h.Branch == "" || h.Worktree == "":
		return claim.Record{}, fmt.Errorf("%w: branch and worktree are required", core.ErrEvidenceIncomplete)
	case !h.WriterStopped:
		return claim.Record{}, fmt.Errorf("%w: writer stop not confirmed", core.ErrEvidenceIncomplete)
	case boundary != head:
		return claim.Record{}, fmt.Errorf("%w: branch moved to %.12s after PR head %.12s", core.ErrStaleHead, boundary, head)
	}
	// A receipt that was already accepted is a replay: same id and same content
	// converge on the stored record, a different content under that id conflicts.
	if cur, err := r.Store.Read(t.Issue); err == nil && cur.Receipt != nil && cur.Receipt.HandoffID == h.HandoffID {
		if cur.Receipt.PR == h.PR && cur.Receipt.Head == head && cur.Receipt.Branch == h.Branch {
			return cur, nil
		}
		return claim.Record{}, fmt.Errorf("%w: receipt %s was accepted with other content", core.ErrDuplicateHandoff, h.HandoffID)
	}
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		if rec.State != core.StateLanding {
			return fmt.Errorf("%w: %s, handoff needs a published PR (landing)", core.ErrTransitionInvalid, rec.State)
		}
		if head != rec.Head {
			return fmt.Errorf("%w: PR head %.12s, candidate %.12s", core.ErrStaleHead, head, rec.Head)
		}
		if rec.Receipt != nil && rec.Receipt.Head == head {
			return fmt.Errorf("%w: PR #%d", core.ErrDuplicateHandoff, rec.Receipt.PR)
		}
		// The receipt's worktree is held like a bound one: while the PR is open no
		// other Issue may bind it, and a finished Issue's reservation is taken over.
		// It is reserved only now, after the owner and every check have passed, so a
		// refused handoff leaves nothing behind.
		if err := r.Store.ReserveWorktree(t.Issue, h.Worktree); err != nil {
			return err
		}
		rec.Repair, rec.Wait, rec.Revision = nil, nil, "" // a new receipt ends the repair, the wait and the revision
		rec.Receipt = &claim.Receipt{
			HandoffID: h.HandoffID, LeaderAccepted: true, PR: h.PR, Head: head, Branch: h.Branch, Worktree: h.Worktree,
			Attempt: rec.Attempt, AttemptSeq: rec.AttemptSeq, Agent: rec.AgentName, Session: rec.Holder,
			PRURL: h.PRURL, Base: h.Base, AdoptedSpec: h.AdoptedSpec, TaskResult: h.TaskResult, CheckStatus: h.CheckStatus,
		}
		rec.AddEvent(claim.Event{Kind: "handoff", Agent: rec.AgentName, Session: rec.Holder, Attempt: string(rec.Attempt), Head: string(head), PR: h.PR, Note: "receipt " + h.HandoffID + " accepted by the leader"})
		return nil
	})
}

// ClaimRepair admits the single repair holder of a handed-off Issue, on the
// branch its receipt names. The repair attempt then runs the normal steps.
func (r *Runner) ClaimRepair(i core.IssueRef, holder, dispatch string, run core.RunID, branch string) (Ticket, error) {
	if err := r.guard(i, core.OpTakeOver); err != nil {
		return Ticket{}, err
	}
	rec, err := r.Store.AcquireRepair(i, holder, dispatch, run, branch)
	if err != nil {
		return Ticket{}, err
	}
	return Ticket{Issue: i, Holder: holder, Epoch: rec.Epoch}, nil
}

// ClaimRepairAs is ClaimRepair for a registered agent.
func (r *Runner) ClaimRepairAs(i core.IssueRef, a core.Agent, dispatch string, run core.RunID, branch string) (Ticket, error) {
	if err := r.guard(i, core.OpTakeOver); err != nil {
		return Ticket{}, err
	}
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

// RepairEvidence is what canon E06 (REVIEWING -> BUILDING/FIX) needs besides the
// repair decision and the exclusive owner, which the reason and the ticket give.
type RepairEvidence struct {
	ScopeReserved  bool
	OldWriterEnded bool
}

// ReopenOwn is Reopen for the owner: the receipt did not release the Issue, so
// the same agent that published the PR takes it back into repair when CI or
// review asks for one. reason is the repair decision; the ticket is the
// exclusive owner; ev carries the rest of the E06 evidence. Without all four the
// record does not move. The receipt, epoch and holder stay, and the position
// becomes BUILDING/FIX.
func (r *Runner) ReopenOwn(t Ticket, reason string, ev RepairEvidence) (claim.Record, error) {
	if strings.TrimSpace(reason) == "" {
		return claim.Record{}, fmt.Errorf("%w: a reopen reason is required", core.ErrEvidenceIncomplete)
	}
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		switch {
		case rec.Receipt == nil:
			return fmt.Errorf("%w: no PR handoff receipt", core.ErrTransitionInvalid)
		case rec.State != core.StateLanding:
			return fmt.Errorf("%w: %s, reopen applies to landing", core.ErrTransitionInvalid, rec.State)
		case rec.Merged:
			return fmt.Errorf("%w: merge already observed, not reverted", core.ErrTransitionInvalid)
		}
		rep := claim.Repair{Decision: reason, Owner: t.Holder, ScopeReserved: ev.ScopeReserved, OldWriterEnded: ev.OldWriterEnded}
		if !ev.ScopeReserved {
			return fmt.Errorf("%w: the scope was not reserved again", core.ErrEvidenceIncomplete)
		}
		if !ev.OldWriterEnded {
			return fmt.Errorf("%w: the previous writer has not ended", core.ErrEvidenceIncomplete)
		}
		rec.State, rec.Pending, rec.Repair, rec.RepairNeeded = core.StateVerifying, "", &rep, ""
		// A repair is a new attempt: the next number and a fresh id, so the earlier
		// attempt's receipt and this one can be told apart.
		rec.AttemptSeq++
		rec.Attempt = core.AttemptID(fmt.Sprintf("repair-%d", rec.AttemptSeq))
		rec.AddEvent(claim.Event{Kind: "reopen", Head: string(rec.Head), PR: rec.Receipt.PR, Note: reason})
		return nil
	})
}

// Takeover is the evidence that an owner is really gone before its Issue moves.
// Contract: every field is a FACT THE CALLER SUPPLIES. Nothing in this package
// probes the owner's process, the worktree or git; the wiring that observes them
// (the monitor and the integration side, lanes 20 and 22) is what makes the
// facts true, and a TakeOver call is only as sound as that observation.
// OwnerGone: the owner's session was observed gone (not merely quiet); SafeStop
// may stand in for it, one of the two is required.
// WriterStopped: nothing can still write the branch or the worktree.
// UncommittedPreserved and PreservedRef: what the owner left uncommitted was
// committed or snapshotted, and PreservedRef names where. What the store does
// check is the rest: all four are present, and the owner and epoch the decision
// was taken about are still the recorded ones.
// All four are required; a missing one leaves the record as it was.
type Takeover struct {
	OwnerGone bool
	// SafeStop is the reference of a stop record the old owner left (REQ-AF-009):
	// it stands in for OwnerGone. One of the two is required.
	SafeStop             string
	WriterStopped        bool
	UncommittedPreserved bool
	PreservedRef         string
	// Wait is the new holder's wait record, required when the Issue is parked
	// (design 2.1c) and ignored otherwise; its owner must be the new holder.
	Wait *core.Wait
}

// TakeOver reassigns a held Issue after its owner died. old is the ticket of the
// owner the decision was taken about; the new holder gets the next epoch and the
// old ticket stops writing. This is not the repair claim: that one serves a
// released record.
func (r *Runner) TakeOver(i core.IssueRef, old Ticket, newHolder, dispatch string, run core.RunID, ev Takeover) (claim.Record, error) {
	switch {
	case !ev.WriterStopped:
		return claim.Record{}, fmt.Errorf("%w: writer stop not confirmed", core.ErrEvidenceIncomplete)
	case !ev.UncommittedPreserved || strings.TrimSpace(ev.PreservedRef) == "":
		return claim.Record{}, fmt.Errorf("%w: the owner's uncommitted work is not preserved", core.ErrEvidenceIncomplete)
	}
	if err := r.guard(i, core.OpTakeOver); err != nil {
		return claim.Record{}, err
	}
	// A safe stop is checked first and settles it: the owner said it stopped and has
	// not acted since, so its process is not consulted. Without one, where the owner
	// recorded its process the departure is read from the process and the caller's
	// OwnerGone is not consulted (REQ-AF-009).
	note := "preserved at " + ev.PreservedRef
	stop := strings.TrimSpace(ev.SafeStop)
	if stop != "" {
		// The store checks the stop under the Issue lock, in the critical section that
		// moves the claim (TakeOverOnStop); no earlier check here could be relied on.
		note += "; safe stop " + stop
	} else {
		st, err := r.Store.OwnerLiveness(i)
		if err != nil {
			return claim.Record{}, err
		}
		switch {
		case st.HasPresence && st.Class == homestate.OwnerLive:
			return claim.Record{}, fmt.Errorf("%w: the owner's process %d is still running", core.ErrHeld, st.PID)
		case st.HasPresence && st.Damaged:
			return claim.Record{}, fmt.Errorf("%w: the presence file %s cannot be read: once the owner's process is confirmed gone, remove that file (the owner's next Heartbeat writes a new one)", core.ErrOwnerUnconfirmed, st.File)
		case st.HasPresence && st.Class == homestate.OwnerIndeterminate:
			return claim.Record{}, fmt.Errorf("%w: the owner's process %d cannot be told running or gone", core.ErrOwnerUnconfirmed, st.PID)
		case st.HasPresence:
			note += fmt.Sprintf("; owner process %d is gone", st.PID)
		case ev.OwnerGone:
			note += "; owner gone (caller-asserted, no process on record)"
		default:
			return claim.Record{}, fmt.Errorf("%w: neither the owner's departure nor a safe stop record is given", core.ErrEvidenceIncomplete)
		}
	}
	if r.beforeClaimMoves != nil {
		r.beforeClaimMoves()
	}
	if stop == "" {
		// The departure was established above from the owner's process (or the caller's
		// word, marked in the note); the store checks holder and epoch only.
		return r.Store.TakeOverUnverified(i, old.Holder, old.Epoch, newHolder, dispatch, run, note, ev.Wait)
	}
	return r.Store.TakeOverOnStop(i, old.Holder, old.Epoch, newHolder, dispatch, run, note, ev.Wait, stop)
}

// RecordSafeStop writes the ticket holder's safe-stop record, the evidence that
// lets a later takeover go ahead on a process that is still running (REQ-AF-009),
// and returns its reference.
func (r *Runner) RecordSafeStop(t Ticket, note string) (string, error) {
	return r.Store.RecordSafeStop(t.Issue, t.Holder, t.Epoch, note)
}

// ContractRevision is the evidence of canon E18 adopt: what changed in the
// contract and the revision that was approved.
type ContractRevision struct {
	Diff     string
	Approved string
}

// AdoptRevision applies an approved contract revision to a NEW attempt of a PR
// that is parked for contract-change in BUILDING/FIX (canon E24, then E18): the
// park is released, the attempt is reset and the position becomes BUILDING/SPEC.
// The receipt, the owner and the epoch stay. Resuming alone (E20) never does
// this. The position decision is core's: any other position or park is refused.
func (r *Runner) AdoptRevision(t Ticket, rev ContractRevision) (claim.Record, error) {
	if strings.TrimSpace(rev.Diff) == "" || strings.TrimSpace(rev.Approved) == "" {
		return claim.Record{}, fmt.Errorf("%w: adopting needs the contract diff and the approved revision", core.ErrEvidenceIncomplete)
	}
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		pos := core.Position{Stage: rec.Stage, Sub: rec.Sub, Park: rec.Park}
		if _, err := core.Transition(pos, core.Event{ID: "E18", Outcome: "adopt", Evidence: []string{"contract-diff", "approved-revision"}}); err != nil {
			return err
		}
		rec.State, rec.Return, rec.Reason, rec.Resume = core.StateClaimed, "", "", ""
		rec.Attempt, rec.Delivery, rec.Repair, rec.Gate, rec.Refusal = "", "", nil, nil, nil
		rec.Revision = rev.Approved
		rec.AddEvent(claim.Event{Kind: "revision", Head: string(rec.Head), Note: "approved revision " + rev.Approved + " applied to a new attempt"})
		return nil
	})
}

// BindWorktree binds the write worktree of the ticket's Issue. It is required
// before writing when the same agent already has another PR open, and it must be
// a worktree no other open Issue uses.
func (r *Runner) BindWorktree(t Ticket, path string) (claim.Record, error) {
	return r.Store.BindWorktree(t.Issue, t.Holder, t.Epoch, path)
}

// SetWait records what the owner waits for. The wait belongs to the ticket's
// holder; a wait owned by anyone else is refused.
func (r *Runner) SetWait(t Ticket, w core.Wait) (claim.Record, error) {
	if err := w.Validate(); err != nil {
		return claim.Record{}, err
	}
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		if rec.State.IsTerminal() {
			return fmt.Errorf("%w: %s has nothing left to wait for", core.ErrTransitionInvalid, rec.State)
		}
		rec.Wait = &w // the store refuses a wait that its holder does not own
		return nil
	})
}

// RetryWait spends one retry of the recorded wait. The retry that reaches the
// limit is recorded with Escalated set and reported as ErrRetryLimit; nothing is
// retried past it.
func (r *Runner) RetryWait(t Ticket, diagnosis, next string, recheck time.Time) (claim.Record, error) {
	return r.spendRetry(r.Store.Update, t, diagnosis, next, recheck)
}

// RemindWait is RetryWait written as a monitor: the reminder keeps the owner's
// ticket but is not the owner acting, so it never stales the owner's safe stop.
func (r *Runner) RemindWait(t Ticket, diagnosis, next string, recheck time.Time) (claim.Record, error) {
	return r.spendRetry(r.Store.UpdateAsMonitor, t, diagnosis, next, recheck)
}

func (r *Runner) spendRetry(update func(core.IssueRef, string, int64, func(*claim.Record) error) (claim.Record, error), t Ticket, diagnosis, next string, recheck time.Time) (claim.Record, error) {
	var limit error
	rec, err := update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		if rec.Wait == nil {
			return fmt.Errorf("%w: no wait is recorded", core.ErrTransitionInvalid)
		}
		// Only a retry that makes the wait escalated in THIS call is written. One
		// against a wait that was already exhausted or escalated is refused and
		// writes nothing, whatever Retry hands back.
		exhausted := rec.Wait.Escalated || rec.Wait.Retries >= rec.Wait.MaxRetries
		w, e := rec.Wait.Retry(diagnosis, next, recheck)
		if e != nil && !errors.Is(e, core.ErrRetryLimit) {
			return e
		}
		if exhausted || (errors.Is(e, core.ErrRetryLimit) && !w.Escalated) {
			return e
		}
		rec.Wait, limit = &w, e
		return nil
	})
	if err != nil {
		return rec, err
	}
	return rec, limit
}

// ClearWait drops the wait once what it waited for has happened.
func (r *Runner) ClearWait(t Ticket) (claim.Record, error) {
	return r.Store.Update(t.Issue, t.Holder, t.Epoch, func(rec *claim.Record) error {
		rec.Wait = nil // the store refuses this while the Issue is parked
		return nil
	})
}

// PreflightPR checks the PR text (description and commit messages) before the PR
// is published. A closing keyword (close/fix/resolve and their forms) that would
// close an Issue the acting user may not close stops the PR: the operation is
// refused with the reason foreign-closing-keyword, the Issues are named, the
// state does not change and Land stays shut (behaviour change: it used to
// pause the Issue as blocked and wait for a resume). Each
// target is decided by the same permission decision as any other close
// (core.Authorize with OpClose), so there is no second rule: your own Issues pass,
// another author's Issue passes only where collaboration allows closing it. A
// target that cannot be checked (another repository, an unknown author, no host
// lookup) blocks. Passing records the HEAD; Land needs it. ownRepo is the PR's
// repository, used to tell same-repository references from others.
func (r *Runner) PreflightPR(t Ticket, head, ownRepo string, texts ...string) (claim.Record, error) {
	h, err := core.ParseHeadSHA(head)
	if err != nil {
		return claim.Record{}, err
	}
	var problems []string
	for _, c := range core.ClosingTargets(texts...) {
		if c.Repo != "" && core.NormalizeRepo(r.Collab.Provider, c.Repo) != core.NormalizeRepo(r.Collab.Provider, ownRepo) {
			problems = append(problems, fmt.Sprintf("%s#%d is in another repository", c.Repo, c.Issue))
			continue
		}
		if c.Issue == t.Issue.Number {
			continue
		}
		if why := r.mayClose(c.Issue); why != "" {
			problems = append(problems, fmt.Sprintf("#%d (%s)", c.Issue, why))
		}
	}
	var refused error
	rec, err := r.step(t, core.StateReady, func(rec *claim.Record) error {
		if h != rec.Head {
			return fmt.Errorf("%w: preflight on %.12s, candidate %.12s", core.ErrStaleHead, h, rec.Head)
		}
		if len(problems) == 0 {
			rec.PreflightHead, rec.Refusal = h, nil
			return nil
		}
		// A refusal (SPEC column R): the state does not change, the refusal is
		// recorded, and Land stays shut because no HEAD is preflighted.
		detail := "PR text would close: " + strings.Join(problems, ", ")
		rec.Refusal = &core.Refusal{Reason: core.ReasonForeignClosing, Detail: detail + "; reword the keyword to Refs #N and run the preflight again"}
		rec.PreflightHead = ""
		rec.AddEvent(claim.Event{Kind: "preflight-refused", Head: string(h), Note: detail})
		refused = fmt.Errorf("%w: %s", core.ErrForeignClosing, detail)
		return nil
	})
	if err != nil {
		return rec, err
	}
	return rec, refused
}

// mayClose returns why the acting user may not close the Issue, or "" when the
// permission decision allows it.
func (r *Runner) mayClose(issue int) string {
	ref, err := core.NewIssueRef(issue)
	if err != nil {
		return "not a valid issue number"
	}
	if r.Facts == nil {
		return "no host lookup to check the author"
	}
	f, err := r.Facts.IssueFacts(r.ctx(), ref)
	if err != nil {
		return "the host lookup failed"
	}
	if v := r.Collab.Authorize(core.AuthzRequest{
		Op: core.OpClose, RequestedRepo: f.RequestedRepo, ActualRepo: f.ActualRepo, Redirected: f.Redirected,
		UserID: f.UserID, IssueAuthorID: f.AuthorID, Owner: f.Owner,
	}); v != core.Allow {
		return string(v)
	}
	return ""
}

// validPRURL accepts an absolute http(s) URL with a host and no blanks whose path
// ends in the change request's own address with the receipt's number: /pull/<pr>
// on GitHub, /-/merge_requests/<pr> on GitLab. The receipt's URL is the PR it
// carries, not another PR, an Issue or a sub-page of one; a trailing slash, a
// query or a fragment do not change which PR it names.
func validPRURL(raw string, pr int) bool {
	if raw == "" || strings.ContainsAny(raw, " \t\r\n") {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return false
	}
	p := strings.TrimSuffix(u.Path, "/")
	n := strconv.Itoa(pr)
	return strings.HasSuffix(p, "/pull/"+n) || strings.HasSuffix(p, "/-/merge_requests/"+n)
}
