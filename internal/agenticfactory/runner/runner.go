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
// code; it makes no write.
type Facts interface {
	IssueFacts(issue core.IssueRef) (IssueFacts, error)
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
}

// guard decides whether the acting user may perform op on the Issue, through the
// one permission decision in core. A failed or missing lookup is an unknown
// identity.
func (r *Runner) guard(i core.IssueRef, op core.Op) error {
	if r.Facts == nil {
		return core.DenyIdentityUnknown.Err()
	}
	f, err := r.Facts.IssueFacts(i)
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
			if r.Attributor != nil {
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
		rec.Receipt = &claim.Receipt{
			HandoffID: h.HandoffID, LeaderAccepted: true, PR: h.PR, Head: head, Branch: h.Branch, Worktree: h.Worktree,
			Attempt: rec.Attempt, AttemptSeq: rec.AttemptSeq, Agent: rec.AgentName, Session: rec.Holder,
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

// ReopenOwn is Reopen for the owner: the receipt did not release the Issue, so
// the same agent that published the PR takes it back to verification when CI or
// review asks for a repair. The receipt, epoch and holder stay.
func (r *Runner) ReopenOwn(t Ticket, reason string) (claim.Record, error) {
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
		rec.State, rec.Pending = core.StateVerifying, ""
		rec.AddEvent(claim.Event{Kind: "reopen", Head: string(rec.Head), PR: rec.Receipt.PR, Note: reason})
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
	f, err := r.Facts.IssueFacts(ref)
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
