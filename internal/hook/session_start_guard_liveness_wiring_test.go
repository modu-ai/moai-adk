package hook

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// The wiring site is one line, and one line is exactly how much of this family
// the sibling SPEC (card t333) shipped broken: its producer seam existed, its
// tests were green, and the real tree rendered nothing because no producer was
// ever named there. That failure is invisible to every test that constructs its
// subject directly — the seam works, each side works, and the two are not
// joined.
//
// So the wiring gets its own guard, and the guard is written to go RED when the
// wiring is reverted to the stub. TestWiredProducerIsNotTheStub below is the
// whole of it.

// TestWiredProducerIsNotTheStub is the wiring mutant's guard.
//
// It reaches the wired producer through the same package variable session start
// reads and asks it to produce. The assertion is NEGATIVE and deliberately so:
// what must not happen is the unwired producer's answer. Asserting a concrete
// success instead would need a manifest, an enumeration and a forge query, and
// would then be a test of those rather than of the join.
//
// What it establishes: the variable names a producer that attempts the work.
// What it does NOT establish: that the work succeeds against a real repository
// — that is the producer's own package's business, and it is measured there.
func TestWiredProducerIsNotTheStub(t *testing.T) {
	if guardLivenessProducer == nil {
		t.Fatal("no producer is wired at the session-start site")
	}

	// An empty tree: the producer gets far enough to look for its manifest and
	// fails on its absence. No network, no subprocess, no fixture.
	_, err := guardLivenessProducer.Produce(
		context.Background(),
		guardliveness.Activation{Root: t.TempDir()},
	)

	if errors.Is(err, guardliveness.ErrProducerUnwired) {
		t.Fatal("the session-start site still names the stub producer: " +
			"every activation reaches the seam and asks nothing, which is the " +
			"absent-execution shape this family is about")
	}
	if err == nil {
		t.Fatal("the wired producer returned a result for a tree with no manifest, " +
			"so it evaluated nothing and reported it as something")
	}
}

// TestWiringIsUnconditional records the second half, which is the half the
// wiring site's own comment argues for at length: the stub is a PRODUCER rather
// than a nil check, so that replacing it cannot introduce a branch short of the
// query layer. A conditional wiring — `if x != nil { … }` at the call site —
// would make the production path depend on the seam existing, rebuilding the
// defect inside the fix.
//
// Measured as: the refresh path reaches the producer on an activation carrying
// nothing that could be read as a reason to skip.
func TestWiringIsUnconditional(t *testing.T) {
	var reached int
	original := guardLivenessProducer
	t.Cleanup(func() { guardLivenessProducer = original })

	guardLivenessProducer = producerFunc(func(context.Context, guardliveness.Activation) (guardliveness.Result, error) {
		reached++
		return guardliveness.Result{}, errors.New("counted")
	})

	guardLivenessRefresh(context.Background(), t.TempDir(), false)

	if reached != 1 {
		t.Fatalf("the refresh reached the producer %d times, want exactly 1", reached)
	}
}

// producerFunc adapts a function to the producer seam.
type producerFunc func(context.Context, guardliveness.Activation) (guardliveness.Result, error)

func (f producerFunc) Produce(ctx context.Context, act guardliveness.Activation) (guardliveness.Result, error) {
	return f(ctx, act)
}
