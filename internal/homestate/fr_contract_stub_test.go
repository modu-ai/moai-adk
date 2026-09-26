package homestate

// RED-phase contract stub for the F1 transition API (t1239 M2). It fixes the
// names and signatures the M2 tests compile against while the implementation
// does not exist yet; every call fails with errF1NotImplemented, so the tests
// fail on their assertions, not on the build. The GREEN commit deletes this
// file and supplies the real implementation under the same names.

import (
	"context"
	"errors"
	"time"
)

var errF1NotImplemented = errors.New("F1 transition API not implemented")

var (
	ErrStaleVersion        = errors.New("stale card version")
	ErrIllegalTransition   = errors.New("illegal card transition")
	ErrReservedEdge        = errors.New("reserved card transition")
	ErrLegacyState         = errors.New("legacy card state")
	ErrEvidence            = errors.New("card evidence refused")
	ErrLeaseExpired        = errors.New("card lease expired")
	ErrLeaseHolder         = errors.New("card lease holder refused")
	ErrUnknownPredecessor  = errors.New("unknown predecessor card")
	ErrPredecessorUnmerged = errors.New("predecessor card not merged")
	ErrDecider             = errors.New("card decider refused")
	ErrInvalidCardInput    = errors.New("invalid card transition input")
)

type TransitionRequest struct {
	RunID, CardID, To string
	ExpectedVersion   int64
	Actor             string
	Decider           string
	Owner             string
	SHA               string
	ArtifactPath      string
	MergeSHA          string
	RemeasurePath     string
	IntegrationBranch string
	Question          string
	Reason            string
	Now               time.Time
}

type TransitionEdge struct{ ID, From, To string }

func TransitionEdges() []TransitionEdge { return nil }

var cardTransitionFault func(stage string) error

func (f *FactoryDB) Transition(context.Context, TransitionRequest) (Card, error) {
	return Card{}, errF1NotImplemented
}
