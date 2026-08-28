package guardliveness

import (
	"context"
	"errors"
	"log/slog"
)

// Activation describes one activation of the already-attended host surface.
//
// It carries the root the subject queries are taken against and nothing that
// could be read as a reason to skip: no changed-file list, no diff summary, no
// subject-matter hint. A refresh entailed by someone working is only entailed
// if the act that PRODUCES a verdict is unconditional on that activation
// (spec.md §D.1), and a field describing what the activation touched is where a
// condition grows.
type Activation struct {
	Root string
}

// Producer produces the next evaluation result. It is the seam to
// SPEC-GUARD-STATE-MODEL-001: issuing the subject queries, classifying each
// entry, and designating the clean value all happen behind it. Reaching
// Produce IS reaching the query layer, which is where AC-GDL-003 counts.
type Producer interface {
	Produce(ctx context.Context, act Activation) (Result, error)
}

// ErrProducerUnwired reports that no evaluation-result producer is wired.
var ErrProducerUnwired = errors.New("guard liveness: no evaluation-result producer is wired")

type unwiredProducer struct{}

func (unwiredProducer) Produce(context.Context, Activation) (Result, error) {
	return Result{}, ErrProducerUnwired
}

// Unwired returns the producer used until the state model lands.
//
// It is a producer rather than a nil check at the call site on purpose. A nil
// check would be a branch short of the query layer, so the production path
// would be conditional on the seam's existence and every activation would pass
// through it having asked nothing — the absent-execution shape this SPEC is
// about, rebuilt inside the deliverable. Reached, it says plainly that there is
// no producer; the count at the query layer stays the count of activations.
func Unwired() Producer { return unwiredProducer{} }

// Evaluator runs the refresh half of the advisory: it asks the producer for the
// next result. It is pull-based and has no cadence of its own — a scheduled
// watcher would be subject to the very defect it watches for (REQ-GDL-002).
type Evaluator struct {
	producer Producer
}

// New returns an Evaluator over the given producer. A nil producer degrades to
// Unwired rather than panicking at session start.
func New(p Producer) *Evaluator {
	if p == nil {
		p = Unwired()
	}
	return &Evaluator{producer: p}
}

// Refresh asks the producer for the next result.
//
// There is no branch between entry and the producer call, and there is nothing
// here to add one to: a refresh that returns early on a subject-matter test
// satisfies a call-site reading of "invoked unconditionally" while being the
// same defect one frame inward.
func (e *Evaluator) Refresh(ctx context.Context, act Activation) (Result, error) {
	return e.producer.Produce(ctx, act)
}

// Refresh is the handle on one initiated refresh.
type Refresh struct {
	done   chan struct{}
	result Result
	err    error
}

// Wait blocks until the refresh completes and returns its outcome. It is
// observability for tests, not part of the host surface's path — the render
// reads a persisted result rather than waiting for one (REQ-GDL-011/012).
func (r *Refresh) Wait() (Result, error) {
	<-r.done
	return r.result, r.err
}

// OnActivation initiates a refresh for this activation and returns immediately.
//
// Both halves are load-bearing. Initiating on EVERY activation is what makes
// the verdict entailed by someone working rather than by a cadence of its own;
// returning without awaiting is what makes that compatible with a
// latency-bounded host surface. The result is carried on the handle for a
// caller that wants it; the host surface discards it and will read the
// persisted result instead.
//
// @MX:WARN: [AUTO] unjoined background goroutine — the refresh outlives this
// call by construction.
// @MX:REASON: the host surface is latency-bounded and the refresh issues one
// query per subject, so it cannot be awaited (spec.md §B.3). Safe because the
// refresh is read-only with respect to the working tree and the forge
// (REQ-GDL-008), so an abandoned one has no effect to undo.
func (e *Evaluator) OnActivation(ctx context.Context, act Activation) *Refresh {
	r := &Refresh{done: make(chan struct{})}
	go func() {
		defer close(r.done)
		defer func() {
			if rec := recover(); rec != nil {
				slog.Debug("guard liveness: refresh panicked (non-blocking)", "recover", rec)
				r.err = errors.New("guard liveness: refresh panicked")
			}
		}()
		res, err := e.Refresh(ctx, act)
		r.result = res
		r.err = err
	}()
	return r
}
