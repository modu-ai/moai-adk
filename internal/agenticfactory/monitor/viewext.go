package monitor

import (
	"context"
	"time"
)

// IntakeStage is where background acceptance of an Issue stands.
type IntakeStage string

const (
	StageAck          IntakeStage = "ack"
	StageIssueCreated IntakeStage = "issue-created"
	StageWTReady      IntakeStage = "wt-ready"
	StageStarted      IntakeStage = "started"
	StageFailed       IntakeStage = "failed"
	StageParked       IntakeStage = "parked"
)

// IntakeItem is one Issue's acceptance record.
type IntakeItem struct {
	Issue  int         `json:"issue"`
	Stage  IntakeStage `json:"stage"`
	State  string      `json:"state"`
	Reason string      `json:"reason,omitempty"`
	At     time.Time   `json:"at"`
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
)

// ContextItem is one entry of an Issue's context ledger.
type ContextItem struct {
	Section    string       `json:"section"`
	Link       string       `json:"link"`
	TargetHead string       `json:"target_head,omitempty"`
	Hash       string       `json:"hash,omitempty"`
	Owner      string       `json:"owner,omitempty"`
	State      ContextState `json:"state"`
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
		switch {
		case items[i].State == CtxInaccessible:
			items[i].Availability = LinkInaccessible
		case items[i].RemoteVerified:
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
	for i := range items {
		items[i].Reason = textField("reason", items[i].Reason, SrcClaimStore, c.Now()).Value
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
