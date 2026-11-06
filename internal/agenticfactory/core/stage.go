package core

import (
	"errors"
	"fmt"
)

// Stage is the S46 execution stage of an Issue (Issue #1948, comment 6489610367,
// transitions E01-E27). It replaces State (the contract 0.4.0 model) in the
// 2nd step of the migration; until then nothing imports it.
type Stage string

const (
	StageReady     Stage = "READY"
	StageBuilding  Stage = "BUILDING"
	StageReviewing Stage = "REVIEWING"
	StageMerging   Stage = "MERGING"
	StageDone      Stage = "DONE"
	StageCancelled Stage = "CANCELLED"
)

// SubStage refines BUILDING and nothing else.
type SubStage string

const (
	SubSpec      SubStage = "SPEC"
	SubImplement SubStage = "IMPLEMENT"
	SubFix       SubStage = "FIX"
)

var (
	ErrGuardMissing     = errors.New("agenticfactory: a transition guard has no evidence")
	ErrUnknownEdge      = errors.New("agenticfactory: no such transition")
	ErrUnknownOutcome   = errors.New("agenticfactory: no such transition outcome")
	ErrNotParked        = errors.New("agenticfactory: the position is not parked")
	ErrLegacyUnmappable = errors.New("agenticfactory: legacy record cannot be mapped to the S46 model")
)

func (s Stage) valid() bool {
	switch s {
	case StageReady, StageBuilding, StageReviewing, StageMerging, StageDone, StageCancelled:
		return true
	}
	return false
}

func (s Stage) terminal() bool { return s == StageDone || s == StageCancelled }

// Position is a stage plus the PARKED overlay. PARKED is not a stage: it never
// changes the stage or the owner, it only marks the position as stopped.
type Position struct {
	Stage Stage
	Sub   SubStage
	// RemoteIntegrated marks a MERGING position whose remote merge was observed
	// (E12). It is meaningful in MERGING only.
	RemoteIntegrated bool
	Park             *Park
}

// Validate checks the shape of a position, not whether it is reachable.
func (p Position) Validate() error {
	bad := func(why string) error { return fmt.Errorf("%w: %s", ErrTransitionInvalid, why) }
	switch {
	case !p.Stage.valid():
		return bad(fmt.Sprintf("unknown stage %q", p.Stage))
	case p.Stage == StageBuilding && p.Sub != SubSpec && p.Sub != SubImplement && p.Sub != SubFix:
		return bad("BUILDING needs a sub stage")
	case p.Stage != StageBuilding && p.Sub != "":
		return bad("only BUILDING has a sub stage")
	case p.RemoteIntegrated && p.Stage != StageMerging:
		return bad("remote_integrated belongs to MERGING")
	case p.Park != nil && p.Stage.terminal():
		return bad("a finished position is never parked")
	}
	if p.Park != nil {
		return p.Park.Validate()
	}
	return nil
}
