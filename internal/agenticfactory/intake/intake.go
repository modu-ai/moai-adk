// Package intake accepts background work and carries it to a ready worktree:
// ACK -> ISSUE_PUBLISHED -> WT_READY, each with evidence (Issue #1948, S46).
// STARTED is not a stored step; it is read from the claim when the agent
// begins its attempt (claim.Store.ViewRequest). State lives in claim.Store;
// issue publishing and worktree preparation are injected, so the code here
// performs no GitHub or filesystem effect of its own.
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

// Ready is what a preparer returns once the worktree exists. WriterStopped
// must be true: the preparer has stopped writing before the agent may.
type Ready struct {
	Path          string
	Head          string
	WriterStopped bool
}

// Publisher creates (or finds) the Issue for a request. It receives the
// request key and must be idempotent on it: Advance may call it again after a
// timeout or a crash that lost the response.
type Publisher interface {
	PublishIssue(ctx context.Context, req claim.Request) (issue int, ref string, err error)
}

// Preparer builds the worktree in an exclusively held slot.
type Preparer interface {
	PrepareWorktree(ctx context.Context, req claim.Request, slot string) (Ready, error)
}

// Config bounds intake. There are no defaults: every value is validated.
type Config struct {
	MaxInFlight int           // requests between ACK and STARTED
	MaxParked   int           // requests held behind them; 0 rejects at once
	Slots       []string      // worktree slot names
	StepTimeout time.Duration // per publish or prepare call
}

// Intake drives requests through the stages.
type Intake struct {
	Store *claim.Store
	pub   Publisher
	prep  Preparer
	cfg   Config
}

// New validates the configuration and the injected effects.
func New(st *claim.Store, pub Publisher, prep Preparer, cfg Config) (*Intake, error) {
	if st == nil || pub == nil || prep == nil {
		return nil, errors.New("agenticfactory/intake: store, publisher and preparer are required")
	}
	if cfg.MaxInFlight < 1 || cfg.MaxParked < 0 || len(cfg.Slots) == 0 || cfg.StepTimeout <= 0 {
		return nil, fmt.Errorf("%w: in-flight %d, parked %d, slots %d, timeout %s", core.ErrLimitInvalid, cfg.MaxInFlight, cfg.MaxParked, len(cfg.Slots), cfg.StepTimeout)
	}
	for _, s := range cfg.Slots {
		if _, err := core.ParseRunID(s); err != nil {
			return nil, fmt.Errorf("%w: slot %q", core.ErrLimitInvalid, s)
		}
	}
	return &Intake{Store: st, pub: pub, prep: prep, cfg: cfg}, nil
}

func (in *Intake) policy() claim.Policy {
	return claim.Policy{MaxInFlight: in.cfg.MaxInFlight, MaxParked: in.cfg.MaxParked}
}

// Accept records the request and acknowledges it. Only the payload hash is
// stored. A repeat is idempotent; another payload under the same key is
// ErrConflict; a full intake is ErrIntakeFull and records nothing.
func (in *Intake) Accept(key string, payload []byte, priority int) (claim.Request, error) {
	sum := sha256.Sum256(payload)
	req, _, err := in.Store.PutRequest(key, hex.EncodeToString(sum[:]), priority, in.policy())
	return req, err
}

// View returns the request with STARTED derived from the claim.
func (in *Intake) View(key string) (claim.Request, error) { return in.Store.ViewRequest(key) }

// Advance performs the one next step of a request. A failed step is recorded
// as a retryable failure and returned; it is not retried here.
func (in *Intake) Advance(ctx context.Context, key string) (claim.Request, error) {
	req, err := in.Store.GetRequest(key)
	if err != nil {
		return claim.Request{}, err
	}
	if req.Parked != "" {
		return req, fmt.Errorf("%w: %s is parked (%s)", core.ErrTransitionInvalid, key, req.Parked)
	}
	switch req.Stage {
	case claim.StageACK:
		return in.publish(ctx, req)
	case claim.StageIssuePublished:
		return in.prepare(ctx, req)
	}
	return req, fmt.Errorf("%w: %s is %s, waiting for the agent to begin", core.ErrTransitionInvalid, key, req.Stage)
}

func (in *Intake) publish(ctx context.Context, req claim.Request) (claim.Request, error) {
	sctx, cancel := context.WithTimeout(ctx, in.cfg.StepTimeout)
	defer cancel()
	issue, ref, err := in.pub.PublishIssue(sctx, req)
	if err == nil && issue <= 0 {
		err = fmt.Errorf("%w: publisher returned issue %d", core.ErrEvidenceIncomplete, issue)
	}
	if err != nil {
		return in.fail(req.Key, err)
	}
	return in.Store.UpdateRequest(req.Key, func(r *claim.Request) error {
		r.Stage, r.Issue, r.Retryable, r.LastError = claim.StageIssuePublished, issue, false, ""
		r.Evidence = append(r.Evidence, claim.StageEvidence{Stage: claim.StageIssuePublished, Ref: ref})
		return nil
	})
}

// takeSlot returns the slot this request holds, taking a free one if needed.
func (in *Intake) takeSlot(req claim.Request) (string, error) {
	if req.Slot != "" {
		return req.Slot, in.Store.AcquireSlot(req.Slot, req.Key)
	}
	for _, s := range in.cfg.Slots {
		err := in.Store.AcquireSlot(s, req.Key)
		if err == nil {
			return s, nil
		}
		if !errors.Is(err, core.ErrHeld) {
			return "", err
		}
	}
	return "", core.ErrNoSlot
}

func (in *Intake) prepare(ctx context.Context, req claim.Request) (claim.Request, error) {
	slot, err := in.takeSlot(req)
	if errors.Is(err, core.ErrNoSlot) {
		if _, uerr := in.Store.UpdateRequest(req.Key, func(r *claim.Request) error { r.Parked = claim.ParkNoSlot; return nil }); uerr != nil {
			return req, uerr
		}
		return req, fmt.Errorf("%w: %s", core.ErrNoSlot, req.Key)
	}
	if err != nil {
		return req, err
	}
	if req.Slot == "" {
		if req, err = in.Store.UpdateRequest(req.Key, func(r *claim.Request) error { r.Slot = slot; return nil }); err != nil {
			return req, err
		}
	}
	sctx, cancel := context.WithTimeout(ctx, in.cfg.StepTimeout)
	defer cancel()
	rd, perr := in.prep.PrepareWorktree(sctx, req, slot)
	if perr == nil {
		if _, herr := core.ParseHeadSHA(rd.Head); herr != nil || rd.Path == "" || !rd.WriterStopped {
			perr = fmt.Errorf("%w: the preparer has not stopped writing, or gave no path or full head", core.ErrEvidenceIncomplete)
		}
	}
	if perr != nil {
		return in.fail(req.Key, perr)
	}
	return in.Store.UpdateRequest(req.Key, func(r *claim.Request) error {
		r.Stage, r.Retryable, r.LastError = claim.StageWTReady, false, ""
		r.Ready = &claim.Ready{Slot: slot, Path: rd.Path, Head: rd.Head, WriterStopped: true}
		r.Evidence = append(r.Evidence, claim.StageEvidence{Stage: claim.StageWTReady, Ref: rd.Path})
		return nil
	})
}

// fail records a retryable failure (a timeout among them) and returns cause.
func (in *Intake) fail(key string, cause error) (claim.Request, error) {
	msg := cause.Error()
	if errors.Is(cause, context.DeadlineExceeded) {
		msg = fmt.Sprintf("timeout: step exceeded %s", in.cfg.StepTimeout)
	}
	msg, _ = core.Redact(msg)
	if len(msg) > 200 {
		msg = msg[:200]
	}
	rec, err := in.Store.UpdateRequest(key, func(r *claim.Request) error {
		r.Attempts++
		r.Retryable, r.LastError = true, msg
		return nil
	})
	if err != nil {
		return claim.Request{}, errors.Join(cause, err)
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
	if v.Stage != claim.StageStarted || v.Slot == "" {
		return fmt.Errorf("%w: %s is %s, the agent has not begun", core.ErrTransitionInvalid, key, v.Stage)
	}
	return in.Store.ReleaseSlot(v.Slot, key)
}
