package guardliveness

import (
	"context"
	"errors"
	"log/slog"
	"time"
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

// Sink records a completed refresh for a later activation to read. *Store is
// the implementation; the interface exists so the evaluator does not know where
// the result lands, and so a test can watch it land.
type Sink interface {
	Save(root string, r Result, takenAt time.Time) error
}

// Evaluator runs the refresh half of the advisory: it asks the producer for the
// next result and persists it. It is pull-based and has no cadence of its own —
// a scheduled watcher would be subject to the very defect it watches for
// (REQ-GDL-002).
type Evaluator struct {
	producer Producer
	sink     Sink
}

// Option configures an Evaluator.
type Option func(*Evaluator)

// WithSink persists each completed refresh, which is what carries a result
// across the render/refresh split: the host surface is latency-bounded, so the
// activation that PRODUCES a verdict is never the one that renders it
// (REQ-GDL-011/012). Without a sink the evaluator still refreshes, and nothing
// downstream ever sees the answer.
func WithSink(s Sink) Option {
	return func(e *Evaluator) { e.sink = s }
}

// New returns an Evaluator over the given producer. A nil producer degrades to
// Unwired rather than panicking at session start.
func New(p Producer, opts ...Option) *Evaluator {
	if p == nil {
		p = Unwired()
	}
	e := &Evaluator{producer: p}
	for _, opt := range opts {
		opt(e)
	}
	return e
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

// Run refreshes and persists, which is the whole of one activation's
// production half.
//
// Only a refresh that ANSWERED is persisted. Recording a failed one would
// replace the last verdict anything actually measured with an empty result,
// and an empty result partitions to nothing non-clean — an all-clear about a
// set nobody evaluated (spec.md §A.0). The previous result stays, and its age
// keeps growing where the reader can see it (REQ-GDL-006).
//
// A persistence failure is reported and is not the refresh's failure: the
// verdict was produced correctly, and only its carriage to the next activation
// broke.
func (e *Evaluator) Run(ctx context.Context, act Activation) (Result, error) {
	res, err := e.Refresh(ctx, act)
	if err != nil {
		return res, err
	}
	if e.sink != nil {
		if saveErr := e.sink.Save(act.Root, res, time.Now()); saveErr != nil {
			slog.Debug("guard liveness: persisting the refresh failed (non-blocking)", "error", saveErr.Error())
		}
	}
	return res, nil
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
// latency-bounded host surface. The completed result is persisted (Run) for a
// LATER activation to render from; the handle carries it for a caller that
// wants to observe this one, and the host surface does not.
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
		res, err := e.Run(ctx, act)
		r.result = res
		r.err = err
	}()
	return r
}
