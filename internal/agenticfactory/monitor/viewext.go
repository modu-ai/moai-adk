package monitor

import (
	"context"
	"time"
)

// IntakeStage is the stored stage of a background intake request. The names are the canonical S51
// stages, owned by lane 18's claim record; this package only displays them.
type IntakeStage string

const (
	StageAccepted         IntakeStage = "ACCEPTED"
	StageIssueConfirmed   IntakeStage = "ISSUE_CONFIRMED"
	StageBranchRegistered IntakeStage = "BRANCH_REGISTERED"
	StageAgentReserved    IntakeStage = "AGENT_RESERVED"
	StageWTPrepared       IntakeStage = "WT_PREPARED"
	StageReadyAcked       IntakeStage = "READY_ACKED"
)

// Displayed row states that are not stages. A request that is waiting is PARKED (a wait for capacity,
// not a failure); BLOCKED needs the recorded resume condition; EFFECT_UNKNOWN means an effect may have
// happened and must be looked up, never repeated.
const (
	StateParked        = "PARKED"
	StateBlocked       = "BLOCKED"
	StateEffectUnknown = "EFFECT_UNKNOWN"
)

// IntakeBlock is the recorded reason a request stopped.
type IntakeBlock struct {
	Step      IntakeStage `json:"step"`
	Detail    string      `json:"detail"`
	Retryable bool        `json:"retryable"`
	Unknown   bool        `json:"unknown"`
	Resume    string      `json:"resume,omitempty"`
}

// IntakeItem is one background intake request. State is derived here, from the fields, so no reader has
// to guess: EFFECT_UNKNOWN when Block.Unknown, BLOCKED when Block is set, PARKED when Parked is set,
// otherwise the stage. Started comes from the claim (an attempt exists), never from the request; it is
// not a stage, and BUILDING follows READY_ACKED.
type IntakeItem struct {
	Key     string       `json:"key"`
	Issue   int          `json:"issue,omitempty"`
	Stage   IntakeStage  `json:"stage"`
	State   string       `json:"state"`
	Parked  string       `json:"parked,omitempty"` // backpressure | no-slot
	Block   *IntakeBlock `json:"block,omitempty"`
	Started bool         `json:"started"`
	Agent   string       `json:"agent,omitempty"` // reserved agent's session id
	Slot    string       `json:"slot,omitempty"`
}

// DeriveIntakeState computes the displayed state of an item.
func DeriveIntakeState(it IntakeItem) string {
	switch {
	case it.Block != nil && it.Block.Unknown:
		return StateEffectUnknown
	case it.Block != nil:
		return StateBlocked
	case it.Parked != "":
		return StateParked
	}
	return string(it.Stage)
}

// ContextState tells a stale context item from one that cannot be reached.
type ContextState string

const (
	CtxCurrent      ContextState = "current"
	CtxStale        ContextState = "stale"
	CtxInaccessible ContextState = "inaccessible"
)

// LinkAvailability is the state of an evidence link. It is "available" only after a remote lookup
// confirmed it; uploading (pushing) alone proves nothing.
type LinkAvailability string

const (
	LinkUnverified   LinkAvailability = "unverified"
	LinkAvailable    LinkAvailability = "available"
	LinkInaccessible LinkAvailability = "inaccessible"
	LinkRestricted   LinkAvailability = "restricted" // limited material: never shown as an open link
)

// EvidenceStage is how far an evidence item has travelled toward being publicly checkable.
type EvidenceStage string

const (
	StagePrepared       EvidenceStage = "prepared"
	StageMasked         EvidenceStage = "masked"
	StageCommitted      EvidenceStage = "committed"
	StagePushed         EvidenceStage = "pushed"
	StageRemoteVerified EvidenceStage = "remote-verified"
	StageRestricted     EvidenceStage = "restricted"
	StageUnknown        EvidenceStage = "unknown"
)

func (s EvidenceStage) valid() bool {
	switch s {
	case StagePrepared, StageMasked, StageCommitted, StagePushed, StageRemoteVerified, StageRestricted:
		return true
	}
	return false
}

// ContextItem is one entry of an Issue's context ledger.
type ContextItem struct {
	Section    string        `json:"section"`
	Link       string        `json:"link"`
	TargetHead string        `json:"target_head,omitempty"`
	Hash       string        `json:"hash,omitempty"`
	Owner      string        `json:"owner,omitempty"`
	State      ContextState  `json:"state"`
	Stage      EvidenceStage `json:"stage"`
	// RemoteVerified is set by the reader only when a remote lookup found the link. The collector
	// derives Availability from it and ignores whatever the reader put in Availability.
	RemoteVerified bool             `json:"remote_verified"`
	Availability   LinkAvailability `json:"availability"`
	At             time.Time        `json:"at"`
}

// QueueStats are the PR queue figures beyond the open count.
type QueueStats struct {
	AwaitingHandoff   int
	RepairBacklog     int
	OldestWaitSeconds int
}

// IntakeReader, ContextReader and QueueReader are read-only sources; each is optional.
type IntakeReader interface {
	ReadIntake(ctx context.Context) ([]IntakeItem, error)
}

type ContextReader interface {
	ReadContext(ctx context.Context, issue int) ([]ContextItem, error)
}

type QueueReader interface {
	ReadQueue(ctx context.Context) (QueueStats, error)
}

func (c *Collector) fillContext(ctx context.Context, r *Row, at time.Time) {
	r.Context.Source = SrcLedger
	if c.Context == nil {
		return
	}
	items, err := c.Context.ReadContext(ctx, r.Issue)
	if err != nil {
		return
	}
	for i := range items {
		items[i].Link, _ = Mask(items[i].Link)
		items[i].Owner, _ = Mask(items[i].Owner)
		if !items[i].Stage.valid() {
			items[i].Stage = StageUnknown
		}
		switch {
		case items[i].State == CtxInaccessible:
			items[i].Availability = LinkInaccessible
		case items[i].Stage == StageRestricted:
			items[i].Availability = LinkRestricted
		case items[i].Stage == StageRemoteVerified && items[i].RemoteVerified:
			items[i].Availability = LinkAvailable
		default:
			items[i].Availability = LinkUnverified
		}
	}
	r.Context = Field[[]ContextItem]{items, SrcLedger, at, true, false}
}

func (c *Collector) fillIntake(ctx context.Context, v *View) {
	v.Intake.Source = SrcClaimStore
	if c.Intake == nil {
		return
	}
	items, err := c.Intake.ReadIntake(ctx)
	if err != nil {
		return
	}
	now := c.Now()
	for i := range items {
		it := &items[i]
		it.Key = textField("key", it.Key, SrcClaimStore, now).Value
		it.Slot = textField("slot", it.Slot, SrcClaimStore, now).Value
		it.Agent = idField(it.Agent, SrcClaimStore, now).Value
		if it.Block != nil {
			b := *it.Block
			b.Detail = textField("detail", b.Detail, SrcClaimStore, now).Value
			b.Resume = textField("resume", b.Resume, SrcClaimStore, now).Value
			it.Block = &b
		}
		it.State = DeriveIntakeState(*it)
	}
	v.Intake = Field[[]IntakeItem]{items, SrcClaimStore, c.Now(), true, false}
}

func (c *Collector) fillQueue(ctx context.Context, b *Backlog) {
	if c.Queue == nil {
		return
	}
	if q, err := c.Queue.ReadQueue(ctx); err == nil {
		b.AwaitingHandoff, b.RepairBacklog, b.OldestWaitSeconds, b.QueueKnown = q.AwaitingHandoff, q.RepairBacklog, q.OldestWaitSeconds, true
	}
}
