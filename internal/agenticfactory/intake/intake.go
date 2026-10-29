// Package intake accepts background work and carries it through the S51
// readiness stages, each with evidence (Issue #1948):
//
//	ACCEPTED -> ISSUE_CONFIRMED -> BRANCH_REGISTERED -> AGENT_RESERVED
//	         -> WT_PREPARED -> READY_ACKED
//
// plus BLOCKED or EFFECT_UNKNOWN, which keep the failed step and how to resume.
// These are readiness stages, not product states: BUILDING starts only after
// READY_ACKED, and whether the agent has begun is read from the claim
// (claim.Store.ViewRequest), never stored. State lives in claim.Store; issue
// creation, branch registration and worktree preparation are injected, so this
// package performs no GitHub, git or filesystem effect of its own.
//
// An effect whose outcome is unknown is never repeated. If creating the Issue
// or registering the branch fails in a way that does not prove nothing was sent,
// the request becomes EFFECT_UNKNOWN: Advance only looks the effect up, and a
// new attempt needs explicit evidence that it did not happen (ConfirmNotCreated).
package intake

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// ErrNotSent is wrapped by a Publisher or Brancher that can prove its request
// never left (for example it was refused before any call). Any other failure is
// treated as an unknown effect.
var ErrNotSent = errors.New("agenticfactory/intake: the request was not sent")

// Publisher creates the Issue for a request and looks it up. PublishIssue
// receives the request key and should carry a correlation marker; FindIssue must
// find it again by that marker. GitHub does not make creation idempotent, so the
// caller never repeats a PublishIssue whose result is unknown.
type Publisher interface {
	PublishIssue(ctx context.Context, req claim.Request) (issue int, ref string, err error)
	FindIssue(ctx context.Context, req claim.Request) (issue int, ref string, found bool, err error)
}

// Brancher registers the one branch of an Issue: it checks whether a remote
// branch is needed, compares it with an existing branch, and records the ref and
// base OID. It must not overwrite an existing branch or force an empty commit.
type Brancher interface {
	RegisterBranch(ctx context.Context, req claim.Request) (claim.Branch, error)
	FindBranch(ctx context.Context, req claim.Request) (claim.Branch, bool, error)
}

// Ready is what a preparer returns once the local worktree is verified.
// WriterStopped must be true: the preparer has stopped writing before the agent may.
type Ready struct {
	Host          string
	Path          string
	Head          string
	WriterStopped bool
}

// Preparer builds and verifies the worktree in an exclusively held slot. It only
// prepares: it writes no product code or SPEC.
type Preparer interface {
	PrepareWorktree(ctx context.Context, req claim.Request, slot string) (Ready, error)
}

// Config bounds intake. There are no defaults: every value is validated.
type Config struct {
	MaxInFlight int           // requests between ACCEPTED and the agent's first attempt
	MaxParked   int           // requests held behind them; 0 rejects at once
	Slots       []string      // worktree slot names
	StepTimeout time.Duration // per injected call
}

// Intake drives requests through the stages.
type Intake struct {
	Store *claim.Store
	pub   Publisher
	br    Brancher
	prep  Preparer
	cfg   Config
}

// New validates the configuration and the injected effects.
func New(st *claim.Store, pub Publisher, br Brancher, prep Preparer, cfg Config) (*Intake, error) {
	if st == nil || pub == nil || br == nil || prep == nil {
		return nil, errors.New("agenticfactory/intake: store, publisher, brancher and preparer are required")
	}
	if cfg.MaxInFlight < 1 || cfg.MaxParked < 0 || len(cfg.Slots) == 0 || cfg.StepTimeout <= 0 {
		return nil, fmt.Errorf("%w: in-flight %d, parked %d, slots %d, timeout %s", core.ErrLimitInvalid, cfg.MaxInFlight, cfg.MaxParked, len(cfg.Slots), cfg.StepTimeout)
	}
	for _, s := range cfg.Slots {
		if _, err := core.ParseRunID(s); err != nil {
			return nil, fmt.Errorf("%w: slot %q", core.ErrLimitInvalid, s)
		}
	}
	return &Intake{Store: st, pub: pub, br: br, prep: prep, cfg: cfg}, nil
}

func (in *Intake) policy() claim.Policy {
	return claim.Policy{MaxInFlight: in.cfg.MaxInFlight, MaxParked: in.cfg.MaxParked}
}

// Accept records the request durably and returns its receipt. Only the payload
// hash is stored. A repeat is idempotent; another payload under the same key is
// ErrConflict; a full intake is ErrIntakeFull and records nothing.
func (in *Intake) Accept(key string, payload []byte, priority int) (claim.Request, error) {
	sum := sha256.Sum256(payload)
	req, _, err := in.Store.PutRequest(key, hex.EncodeToString(sum[:]), priority, in.policy())
	return req, err
}

// View returns the request with Started derived from the claim.
func (in *Intake) View(key string) (claim.Request, error) { return in.Store.ViewRequest(key) }

// Advance performs the one next step of a request that needs no agent. A failed
// step is recorded in the request's Blocked row and returned; it is not retried
// here. A request whose effect is unknown is only looked up.
func (in *Intake) Advance(ctx context.Context, key string) (claim.Request, error) {
	req, err := in.Store.GetRequest(key)
	if err != nil {
		return claim.Request{}, err
	}
	if req.Parked != "" {
		return req, fmt.Errorf("%w: %s is parked (%s)", core.ErrTransitionInvalid, key, req.Parked)
	}
	if req.Blocked != nil && req.Blocked.Unknown {
		return in.lookup(ctx, req)
	}
	switch req.Stage {
	case claim.StageAccepted:
		return in.publish(ctx, req)
	case claim.StageIssueConfirmed:
		return in.register(ctx, req)
	case claim.StageAgentReserved:
		return in.prepare(ctx, req)
	case claim.StageBranchRegistered:
		return req, fmt.Errorf("%w: %s waits for an agent reservation", core.ErrTransitionInvalid, key)
	case claim.StageWTPrepared:
		return req, fmt.Errorf("%w: %s waits for the reserved agent to acknowledge the receipt", core.ErrTransitionInvalid, key)
	}
	return req, fmt.Errorf("%w: %s is %s, waiting for the agent to begin", core.ErrTransitionInvalid, key, req.Stage)
}

func (in *Intake) step(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, in.cfg.StepTimeout)
}

// confirmIssue binds an Issue number to the request after a lookup agrees.
func (in *Intake) confirmIssue(key string, issue int, ref string) (claim.Request, error) {
	return in.Store.UpdateRequest(key, func(r *claim.Request) error {
		r.Stage, r.Issue, r.Blocked = claim.StageIssueConfirmed, issue, nil
		r.Evidence = append(r.Evidence, claim.StageEvidence{Stage: claim.StageIssueConfirmed, Ref: ref})
		return nil
	})
}

func (in *Intake) publish(ctx context.Context, req claim.Request) (claim.Request, error) {
	sctx, cancel := in.step(ctx)
	defer cancel()
	issue, ref, err := in.pub.PublishIssue(sctx, req)
	if err != nil {
		return in.fail(req.Key, claim.StageAccepted, err, !errors.Is(err, ErrNotSent))
	}
	if issue <= 0 {
		return in.fail(req.Key, claim.StageAccepted, fmt.Errorf("publisher returned issue %d", issue), true)
	}
	// ISSUE_CONFIRMED needs the lookup to agree: GitHub gives no idempotency.
	lctx, lcancel := in.step(ctx)
	defer lcancel()
	found, fref, ok, ferr := in.pub.FindIssue(lctx, req)
	if ferr != nil || !ok || found != issue {
		cause := ferr
		if cause == nil {
			cause = fmt.Errorf("lookup found issue %d (found=%v), publisher said %d", found, ok, issue)
		}
		return in.fail(req.Key, claim.StageAccepted, cause, true)
	}
	if fref == "" {
		fref = ref
	}
	return in.confirmIssue(req.Key, issue, fref)
}

func (in *Intake) register(ctx context.Context, req claim.Request) (claim.Request, error) {
	sctx, cancel := in.step(ctx)
	defer cancel()
	b, err := in.br.RegisterBranch(sctx, req)
	if err != nil {
		return in.fail(req.Key, claim.StageIssueConfirmed, err, !errors.Is(err, ErrNotSent))
	}
	return in.recordBranch(req.Key, b)
}

func (in *Intake) recordBranch(key string, b claim.Branch) (claim.Request, error) {
	if _, err := core.ParseHeadSHA(b.BaseOID); err != nil || b.Ref == "" || (b.Remote && b.RemoteOID == "") {
		return in.fail(key, claim.StageIssueConfirmed, fmt.Errorf("%w: branch receipt needs a ref, a full base OID and, for a remote branch, its OID", core.ErrEvidenceIncomplete), true)
	}
	return in.Store.UpdateRequest(key, func(r *claim.Request) error {
		r.Stage, r.Branch, r.Blocked = claim.StageBranchRegistered, &b, nil
		r.Evidence = append(r.Evidence, claim.StageEvidence{Stage: claim.StageBranchRegistered, Ref: b.Ref + "@" + b.BaseOID})
		return nil
	})
}

// lookup resolves an EFFECT_UNKNOWN request by reading, never by repeating.
func (in *Intake) lookup(ctx context.Context, req claim.Request) (claim.Request, error) {
	sctx, cancel := in.step(ctx)
	defer cancel()
	switch req.Blocked.Step {
	case claim.StageAccepted:
		issue, ref, ok, err := in.pub.FindIssue(sctx, req)
		if err != nil {
			return req, fmt.Errorf("%w: lookup of the issue failed: %v", core.ErrEffectUnknown, err)
		}
		if !ok {
			return req, fmt.Errorf("%w: no issue is found for %s; confirm it was not created to allow another attempt", core.ErrEffectUnknown, req.Key)
		}
		return in.confirmIssue(req.Key, issue, ref)
	case claim.StageIssueConfirmed:
		b, ok, err := in.br.FindBranch(sctx, req)
		if err != nil {
			return req, fmt.Errorf("%w: lookup of the branch failed: %v", core.ErrEffectUnknown, err)
		}
		if !ok {
			return req, fmt.Errorf("%w: no branch is found for %s; confirm it was not registered to allow another attempt", core.ErrEffectUnknown, req.Key)
		}
		return in.recordBranch(req.Key, b)
	}
	return req, fmt.Errorf("%w: step %s has no lookup", core.ErrEffectUnknown, req.Blocked.Step)
}

// ConfirmNotCreated records evidence that the unknown effect did not happen, so
// the next Advance may try the step again. Without that evidence a request
// stays EFFECT_UNKNOWN: no second Issue, branch or agent is ever made on a guess.
func (in *Intake) ConfirmNotCreated(key, evidence string) (claim.Request, error) {
	if evidence == "" {
		return claim.Request{}, fmt.Errorf("%w: evidence that the effect did not happen is required", core.ErrEvidenceIncomplete)
	}
	redacted, _ := core.Redact(evidence)
	return in.Store.UpdateRequest(key, func(r *claim.Request) error {
		if r.Blocked == nil || !r.Blocked.Unknown {
			return fmt.Errorf("%w: %s has no unknown effect", core.ErrTransitionInvalid, key)
		}
		r.Blocked = &claim.Blocked{Step: r.Blocked.Step, Detail: "confirmed not created: " + redacted, Retryable: true}
		return nil
	})
}

// Reserve reserves the agent and a worktree slot (AGENT_RESERVED). It does not
// permit development.
func (in *Intake) Reserve(key string, a core.Agent) (claim.Request, error) {
	return in.Store.ReserveRequest(key, a, in.cfg.Slots)
}

func (in *Intake) prepare(ctx context.Context, req claim.Request) (claim.Request, error) {
	sctx, cancel := in.step(ctx)
	defer cancel()
	rd, perr := in.prep.PrepareWorktree(sctx, req, req.Slot)
	if perr == nil {
		if _, herr := core.ParseHeadSHA(rd.Head); herr != nil || rd.Path == "" || rd.Host == "" || !rd.WriterStopped {
			perr = fmt.Errorf("%w: the preparer must give a host, a path and a full head, and have stopped writing", core.ErrEvidenceIncomplete)
		}
	}
	if perr != nil {
		// Preparing a local worktree is repeatable and keeps whatever it made.
		return in.fail(req.Key, claim.StageAgentReserved, perr, false)
	}
	return in.Store.UpdateRequest(req.Key, func(r *claim.Request) error {
		r.Stage, r.Blocked = claim.StageWTPrepared, nil
		r.Ready = &claim.Ready{Host: rd.Host, Slot: req.Slot, Path: rd.Path, Head: rd.Head, WriterStopped: true}
		r.Evidence = append(r.Evidence, claim.StageEvidence{Stage: claim.StageWTPrepared, Ref: rd.Host + ":" + rd.Path})
		return nil
	})
}

// AckReady is the reserved agent's explicit acceptance of the preparation
// receipt (READY_ACKED). Only after it may that agent begin an attempt.
func (in *Intake) AckReady(key string, a core.Agent) (claim.Request, error) {
	return in.Store.AckRequestReady(key, a)
}

// fail records the failed step. unknown marks an effect that may have happened.
func (in *Intake) fail(key string, step claim.Stage, cause error, unknown bool) (claim.Request, error) {
	msg := cause.Error()
	if errors.Is(cause, context.DeadlineExceeded) {
		msg = fmt.Sprintf("timeout: step exceeded %s", in.cfg.StepTimeout)
	}
	msg, _ = core.Redact(msg)
	if len(msg) > 200 {
		msg = msg[:200]
	}
	b := claim.Blocked{Step: step, Detail: msg, Retryable: !unknown, Unknown: unknown}
	if unknown {
		b.Resume = "look the effect up (Advance); another attempt needs ConfirmNotCreated"
	}
	rec, err := in.Store.UpdateRequest(key, func(r *claim.Request) error {
		r.Attempts++
		r.Blocked = &b
		return nil
	})
	if err != nil {
		return claim.Request{}, errors.Join(cause, err)
	}
	if unknown {
		return rec, errors.Join(core.ErrEffectUnknown, cause)
	}
	return rec, cause
}

// Pump unparks requests that now fit, highest priority first.
func (in *Intake) Pump() ([]claim.Request, error) {
	free := 0
	for _, s := range in.cfg.Slots {
		owner, err := in.Store.SlotOwner(s)
		if err != nil {
			return nil, err
		}
		if owner == "" {
			free++
		}
	}
	return in.Store.PromoteParked(in.policy(), free)
}

// ReleaseWorktree returns the request's slot once the agent has begun, so the
// next request can use it.
func (in *Intake) ReleaseWorktree(key string) error {
	v, err := in.Store.ViewRequest(key)
	if err != nil {
		return err
	}
	if !v.Started || v.Slot == "" {
		return fmt.Errorf("%w: %s is %s, the agent has not begun", core.ErrTransitionInvalid, key, v.Stage)
	}
	return in.Store.ReleaseSlot(v.Slot, key)
}
