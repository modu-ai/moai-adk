// Package monitor builds the read-only view of the factory (SPEC-AGENTIC-FACTORY-001 REQ-AF-005,
// design §8.1): per project/Issue/owner it shows state, last-checked time, HEAD, required checks,
// progress, block reason, next action and mirror state, each with its source and read time.
//
// The package has no write path: its source interfaces carry no mutating methods and it imports
// only read-safe standard-library packages (enforced by TestReadBoundaryHasNoWriteVerbs).
package monitor

import (
	"context"
	"errors"
	"time"
)

// Source names the authoritative place a field was read from.
type Source string

const (
	SrcClaimStore Source = "claim-store"
	SrcRemote     Source = "remote"
	SrcLedger     Source = "issue-ledger"
)

// ErrNotFound is returned by a ClaimReader when the Issue has no claim row.
var ErrNotFound = errors.New("monitor: no claim row")

// ClaimRow is what the local claim store holds for one Issue.
type ClaimRow struct {
	Project         string
	Issue           int
	Owner           string
	AgentName       string // display name, e.g. agent-3; not an identity
	SessionID       string // immutable session id; the identity
	State           string // the runner's old progress state (still recorded); display only
	Stage           string // S46 stage: READY, BUILDING, REVIEWING, MERGING, DONE, CANCELLED
	SubStage        string // SPEC, IMPLEMENT or FIX while BUILDING, else ""
	Park            *ParkInfo
	Gate            *GateInfo
	Refusal         *RefusalInfo
	Wait            *WaitInfo
	Attempt         int
	Epoch           int
	HeadSHA         string
	StoreSeq        int
	MirrorState     string
	LastCheckedAt   time.Time
	BlockReason     string
	ResumeCondition string
}

// ParkInfo is the PARKED overlay: the position is stopped, the stage and owner are unchanged.
type ParkInfo struct {
	Reason   string `json:"reason"`
	Evidence string `json:"evidence"`
	Resume   string `json:"resume_condition"`
}

// GateInfo is the merge-gate record on a stage. It does not move the stage.
type GateInfo struct {
	Reason string `json:"reason"`
	Detail string `json:"detail"`
	Resume string `json:"resume_condition"`
	Head   string `json:"head,omitempty"`
}

// RefusalInfo is the last refused operation. It does not move the stage.
type RefusalInfo struct {
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

// WaitInfo is what the assigned agent waits for (CI, a review, a host answer): who is responsible, what
// happens next on its own, when it is looked at again, why it waits, how many retries are spent of how
// many are allowed, and who is told when they run out. It sits beside Park and Gate and moves nothing.
type WaitInfo struct {
	Owner      string    `json:"owner"`
	NextAction string    `json:"next_action"`
	RecheckAt  time.Time `json:"recheck_at"`
	Diagnosis  string    `json:"diagnosis"`
	Retries    int       `json:"retries"`
	MaxRetries int       `json:"max_retries"`
	EscalateTo string    `json:"escalate_to"`
	Escalated  bool      `json:"escalated"`
}

// CheckRun is one required check as the remote reports it.
type CheckRun struct {
	Name       string
	Conclusion string
	HeadSHA    string
}

// ClaimReader, RemoteReader and LedgerReader are the only things the collector can call.
type ClaimReader interface {
	ReadClaim(ctx context.Context, issue int) (ClaimRow, error)
}

type RemoteReader interface {
	RequiredChecks(ctx context.Context, headSHA string) ([]CheckRun, error)
}

type LedgerReader interface {
	NextAction(ctx context.Context, issue int) (string, error)
}

// Field is one displayed value with where and when it was read. Confirmed is false when the source
// could not be read; the value is then empty, never a guess.
type Field[T any] struct {
	Value     T         `json:"value"`
	Source    Source    `json:"source"`
	ReadAt    time.Time `json:"read_at"`
	Confirmed bool      `json:"confirmed"`
	Redacted  bool      `json:"redacted"`
}

// Row is the monitor view of one Issue.
type Row struct {
	Project string
	Issue   int

	Owner           Field[string]
	Agent           Field[string] `json:"agent"`
	SessionID       Field[string] `json:"session_id"`
	State           Field[string]
	Stage           Field[string]
	SubStage        Field[string]       `json:"sub_stage"`
	Park            Field[*ParkInfo]    `json:"park"`
	Gate            Field[*GateInfo]    `json:"gate"`
	Refusal         Field[*RefusalInfo] `json:"refusal"`
	Wait            Field[*WaitInfo]    `json:"wait"`
	Attempt         Field[int]
	Epoch           Field[int]
	HeadSHA         Field[string]
	LastCheckedAt   Field[time.Time]
	BlockReason     Field[string]
	ResumeCondition Field[string]
	MirrorState     Field[string]
	Checks          Field[[]CheckRun]
	NextAction      Field[string]

	PR      Field[PRStatus]      `json:"pr"`
	Context Field[[]ContextItem] `json:"context"`
	// Completed is true only when the claim state is done AND the PR is merged AND the Issue is
	// closed. An open or merely merged PR is never completion.
	Completed bool `json:"completed"`

	// ChecksStale is set when any returned check ran against a HEAD other than the claim's HEAD.
	ChecksStale bool
}

// Collector gathers rows. All fields are read-only sources.
type Collector struct {
	Claims ClaimReader
	Remote RemoteReader
	Ledger LedgerReader
	Now    func() time.Time

	PRs          PRReader // optional; nil leaves PR fields unconfirmed
	Intake       IntakeReader
	Context      ContextReader
	Queue        QueueReader
	BacklogLimit int // open-PR cap; 0 = no cap configured
}

// Collect returns one row per Issue, in the order given.
func (c *Collector) Collect(ctx context.Context, issues []int) []Row {
	rows := make([]Row, 0, len(issues))
	for _, n := range issues {
		rows = append(rows, c.row(ctx, n))
	}
	return rows
}

func (c *Collector) row(ctx context.Context, issue int) Row {
	at := c.Now()
	r := Row{Issue: issue}
	cr, err := c.Claims.ReadClaim(ctx, issue)
	if err == nil {
		r.Project = cr.Project
		r.Owner = textField("owner", cr.Owner, SrcClaimStore, at)
		r.Agent = textField("agent", cr.AgentName, SrcClaimStore, at)
		r.SessionID = idField(cr.SessionID, SrcClaimStore, at)
		r.State = textField("state", cr.State, SrcClaimStore, at)
		r.Stage = textField("stage", cr.Stage, SrcClaimStore, at)
		r.SubStage = textField("sub_stage", cr.SubStage, SrcClaimStore, at)
		r.Park = Field[*ParkInfo]{maskPark(cr.Park, at), SrcClaimStore, at, true, false}
		r.Gate = Field[*GateInfo]{maskGate(cr.Gate, at), SrcClaimStore, at, true, false}
		r.Refusal = Field[*RefusalInfo]{maskRefusal(cr.Refusal, at), SrcClaimStore, at, true, false}
		r.Wait = Field[*WaitInfo]{maskWait(cr.Wait, at), SrcClaimStore, at, true, false}
		r.Attempt = Field[int]{cr.Attempt, SrcClaimStore, at, true, false}
		r.Epoch = Field[int]{cr.Epoch, SrcClaimStore, at, true, false}
		r.HeadSHA = textField("head", cr.HeadSHA, SrcClaimStore, at)
		r.LastCheckedAt = Field[time.Time]{cr.LastCheckedAt, SrcClaimStore, at, true, false}
		r.BlockReason = textField("block_reason", cr.BlockReason, SrcClaimStore, at)
		r.ResumeCondition = textField("resume_condition", cr.ResumeCondition, SrcClaimStore, at)
		r.MirrorState = textField("mirror_state", cr.MirrorState, SrcClaimStore, at)
	} else {
		r.Owner.Source, r.State.Source, r.Stage.Source = SrcClaimStore, SrcClaimStore, SrcClaimStore
		r.Agent.Source, r.SessionID.Source = SrcClaimStore, SrcClaimStore
		r.SubStage.Source, r.Park.Source, r.Gate.Source, r.Refusal.Source = SrcClaimStore, SrcClaimStore, SrcClaimStore, SrcClaimStore
		r.Wait.Source = SrcClaimStore
		r.Attempt.Source, r.Epoch.Source, r.HeadSHA.Source = SrcClaimStore, SrcClaimStore, SrcClaimStore
		r.LastCheckedAt.Source, r.BlockReason.Source = SrcClaimStore, SrcClaimStore
		r.ResumeCondition.Source, r.MirrorState.Source = SrcClaimStore, SrcClaimStore
	}

	r.Checks.Source = SrcRemote
	if err == nil && cr.HeadSHA != "" {
		if cs, cerr := c.Remote.RequiredChecks(ctx, cr.HeadSHA); cerr == nil {
			r.Checks = Field[[]CheckRun]{cs, SrcRemote, at, true, false}
			for _, k := range cs {
				if k.HeadSHA != cr.HeadSHA {
					r.ChecksStale = true
				}
			}
		}
	}

	r.NextAction.Source = SrcLedger
	if next, lerr := c.Ledger.NextAction(ctx, issue); lerr == nil {
		r.NextAction = textField("next_action", next, SrcLedger, at)
	}
	c.fillPR(ctx, &r, at)
	c.fillContext(ctx, &r, at)
	return r
}

// textField masks the text and omits it entirely when it still looks like a credential.
func textField(_ string, s string, src Source, at time.Time) Field[string] {
	out, masked := Mask(s)
	if looksOpaque(out) {
		return Field[string]{Source: src, ReadAt: at, Confirmed: true, Redacted: true}
	}
	return Field[string]{Value: out, Source: src, ReadAt: at, Confirmed: true, Redacted: masked}
}

// idField carries an identifier verbatim when it is a UUID; anything else goes through textField.
func idField(s string, src Source, at time.Time) Field[string] {
	if uuidRe.MatchString(s) {
		return Field[string]{Value: s, Source: src, ReadAt: at, Confirmed: true}
	}
	return textField("session_id", s, src, at)
}

// SameIdentity judges identity by session id alone. The display name never decides, and an
// unconfirmed or empty session id never matches anything.
func SameIdentity(a, b Row) bool {
	return a.SessionID.Confirmed && b.SessionID.Confirmed && a.SessionID.Value != "" && a.SessionID.Value == b.SessionID.Value
}

// The overlay, gate and refusal texts come from people and tools, so each text field is masked and
// omitted whole when it still looks like a credential (the same rule as every other text field).
func maskText(label, s string, at time.Time) string {
	return textField(label, s, SrcClaimStore, at).Value
}

func maskPark(p *ParkInfo, at time.Time) *ParkInfo {
	if p == nil {
		return nil
	}
	return &ParkInfo{Reason: p.Reason, Evidence: maskText("evidence", p.Evidence, at), Resume: maskText("resume", p.Resume, at)}
}

func maskGate(g *GateInfo, at time.Time) *GateInfo {
	if g == nil {
		return nil
	}
	return &GateInfo{Reason: g.Reason, Detail: maskText("detail", g.Detail, at), Resume: maskText("resume", g.Resume, at), Head: g.Head}
}

func maskWait(w *WaitInfo, at time.Time) *WaitInfo {
	if w == nil {
		return nil
	}
	return &WaitInfo{
		Owner: w.Owner, NextAction: maskText("next_action", w.NextAction, at), RecheckAt: w.RecheckAt,
		Diagnosis: maskText("diagnosis", w.Diagnosis, at), Retries: w.Retries, MaxRetries: w.MaxRetries,
		EscalateTo: w.EscalateTo, Escalated: w.Escalated,
	}
}

func maskRefusal(r *RefusalInfo, at time.Time) *RefusalInfo {
	if r == nil {
		return nil
	}
	return &RefusalInfo{Reason: r.Reason, Detail: maskText("detail", r.Detail, at)}
}
