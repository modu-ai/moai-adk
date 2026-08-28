package hook

import (
	"context"
	"log/slog"
	"time"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// guardLivenessJoinBound caps how long session start waits for the persisted
// guard-liveness verdict before giving up on it for this session.
//
// The bound belongs to the CALLER, for the reason the comparable landed
// contributor on this surface records a few files over: a timeout placed inside
// the read would be replaced along with it, so it could not protect the handler
// from a read that misbehaves. No shared bound exists on this surface — each
// contributor carries its own — which makes an undeclared bound an unbounded
// one.
//
// The value matches binaryLagJoinBound because this is the same shape of work:
// a small local read that normally finishes in a fraction of a millisecond,
// with a quarter-second capping the worst case a pathological filesystem can
// add to session start. It is a ceiling, not a budget.
//
// What is NOT behind this bound is the reason the split exists at all: the
// subject queries. The refresh issues one per subject and no sequence of
// network round-trips fits inside a quarter-second, so the render reads a
// PERSISTED result and the refresh runs unawaited beside it (REQ-GDL-011/012).
const guardLivenessJoinBound = 250 * time.Millisecond

// guardLivenessProducer is the seam to the state model
// (SPEC-GUARD-STATE-MODEL-001), which has not landed. Until it does, the wired
// producer reports that it is unwired — reached on every activation, so the
// count at the query layer stays the count of activations rather than dropping
// silently to zero. When the state model lands, this one line names its
// producer instead.
var guardLivenessProducer guardliveness.Producer = guardliveness.Unwired()

// newGuardLivenessStore resolves where verdicts persist between activations.
//
// A variable rather than a direct call so a test can point it at a throwaway
// directory. The default is under the user's ~/.moai state tree, which is
// OUTSIDE every evaluated working tree: a cache written into the tree would
// show up as drift for the next reader, and the advisory path leaves the tree
// byte-identical (REQ-GDL-008).
var newGuardLivenessStore = guardliveness.DefaultStore

// guardLivenessRefresh initiates the guard-liveness refresh for one session
// start (SPEC-GUARD-LIVENESS-001 REQ-GDL-002/003/012, card t333).
//
// Three properties, each of which the SPEC pays for elsewhere if lost:
//
//   - PULL-BASED. The evaluator runs at a moment that already happens for other
//     reasons rather than on a cadence of its own, so it is not itself subject
//     to the defect it watches for and no scheduled workflow is added
//     (REQ-GDL-002).
//   - UNCONDITIONAL. There is no path filter, no changed-file test, and no
//     subject-matter condition here or one frame inward — a condition that
//     stops matching is exactly how the guards of spec.md §A.3 went quiet
//     without being removed. Note what is NOT here: no early return on an empty
//     root, because "nothing to query" is the producer's judgement to make and
//     a caller-side skip is a condition wearing a guard clause's clothes.
//   - NEVER AWAITED. The refresh is initiated and abandoned for this session;
//     session start is latency-bounded and the refresh issues one query per
//     subject, so awaiting it would put the network on the input-lag budget.
//     Its result is persisted on completion and read by a LATER activation
//     (guardLivenessAdvisory), which is what keeps the unconditional binding
//     compatible with the bound.
func guardLivenessRefresh(ctx context.Context, root string) {
	act := guardliveness.Activation{Root: root}
	evaluator := guardliveness.New(guardLivenessProducer, guardLivenessSinkOption())

	if !deferredScansAsyncEnabled() {
		// Test path (TestMain sets deferredScansAsync=false): run inline so no
		// goroutine outlives the test boundary, matching the binary-lag
		// contributor on this surface and for the same reason.
		if _, err := evaluator.Run(ctx, act); err != nil {
			slog.Debug("guard liveness: refresh failed (non-blocking)", "error", err.Error())
		}
		return
	}

	evaluator.OnActivation(ctx, act)
}

// guardLivenessSinkOption wires the persistence the refresh writes to. A store
// that cannot be resolved degrades to no sink: the refresh still runs and still
// reaches the query layer, and only the carriage to the next activation is lost
// — a degradation the render reports as an absent verdict rather than as an
// all-clear.
func guardLivenessSinkOption() guardliveness.Option {
	store, err := newGuardLivenessStore()
	if err != nil {
		slog.Debug("guard liveness: persistence unavailable (non-blocking)", "error", err.Error())
		return func(*guardliveness.Evaluator) {}
	}
	return guardliveness.WithSink(store)
}

// guardLivenessAdvisory returns the guard firing-liveness notice for this tree,
// or the empty string when there is nothing to say
// (SPEC-GUARD-LIVENESS-001 REQ-GDL-004/005/006/007/008/011/013, card t333).
//
// It leads with what CHANGED since the previous render and carries the rest as
// a count (REQ-GDL-007), which needs a memory of what the reader was last
// shown. That memory is written beside the persisted verdict — OUTSIDE every
// evaluated working tree, so the render leaves the tree byte-identical
// (REQ-GDL-008). Nothing on this path calls a forge.
//
// It reads a PERSISTED verdict and issues no subject query of its own. That
// split is the constraint this surface imposes rather than a preference: one
// query per subject cannot fit inside a quarter-second join bound, so the act
// that produces a verdict and the act that renders one are separate, and the
// reader sees the most recent COMPLETED refresh with its age disclosed.
//
// The operator supplies nothing. No guard name, no workflow file, no query —
// which is precisely the input the lead session of spec.md §A.4 did not have
// and could not have produced. A verdict that answers only when asked has
// relocated the defect into whoever is expected to already know the question.
//
// Silence covers exactly two outcomes and neither of them is a failure to read:
// no verdict has been persisted yet, or the verdict says every subject is
// clean. A verdict that could not be READ renders the contract violation by
// name (guardliveness.Advisory), because reporting that as silence would be
// this card's own subject at the consumer's layer.
func guardLivenessAdvisory(root string) string {
	if root == "" {
		return ""
	}

	read := func() string {
		store, err := newGuardLivenessStore()
		if err != nil {
			slog.Debug("session start: guard liveness persistence unavailable (non-blocking)", "error", err.Error())
			return ""
		}
		snapshot, err := store.Load(root)
		if err != nil {
			slog.Debug("session start: no persisted guard-liveness verdict (non-blocking)", "error", err.Error())
			return ""
		}

		// What the PREVIOUS render announced, which is what the reader can be
		// assumed to already know. A read that fails degrades to an empty
		// record: every non-clean entry is then announced again, which is
		// noisier than intended and never quieter — the failure direction that
		// costs a reader nothing.
		previous, err := store.LoadRendered(root)
		if err != nil {
			slog.Debug("session start: previous guard-liveness render record unavailable (non-blocking)", "error", err.Error())
			previous = guardliveness.RenderRecord{}
		}

		text, rendered := guardliveness.Advisory(snapshot, previous, time.Now())
		if rendered != nil {
			if err := store.SaveRendered(root, *rendered); err != nil {
				slog.Debug("session start: persisting the guard-liveness render record failed (non-blocking)", "error", err.Error())
			}
		}
		return text
	}

	if !deferredScansAsyncEnabled() {
		// Test path (TestMain sets deferredScansAsync=false): run inline so no
		// goroutine outlives the test boundary.
		return read()
	}

	// @MX:WARN: [AUTO] unjoined background goroutine — a read that outruns
	// guardLivenessJoinBound is abandoned by the select below and keeps running.
	// @MX:REASON: session start is latency-bounded, so the read cannot be
	// awaited without spending the input-lag budget. Bounded because the read is
	// non-mutating with respect to the working tree and the forge (REQ-GDL-008)
	// and its channel is buffered, so an abandoned goroutine sends once and
	// exits rather than leaking. NOT harmless: an abandoned read still writes
	// its render record, so the operator sees nothing while the entry is marked
	// announced and appears as a standing count next session (progress.md §E.2
	// M3 residual risk).
	//
	// Buffered so a read that finishes after the bound has elapsed can still
	// send and exit rather than leaking blocked on an abandoned channel.
	rendered := make(chan string, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Debug("session start: guard liveness render panicked (non-blocking)", "recover", r)
			}
		}()
		rendered <- read()
	}()

	timer := time.NewTimer(guardLivenessJoinBound)
	defer timer.Stop()
	select {
	case advisory := <-rendered:
		return advisory
	case <-timer.C:
		slog.Debug("session start: guard liveness render exceeded join bound (non-blocking)",
			"bound", guardLivenessJoinBound.String())
		return ""
	}
}
