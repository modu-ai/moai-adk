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
	State           string
	Stage           string
	Attempt         int
	Epoch           int
	HeadSHA         string
	StoreSeq        int
	MirrorState     string
	LastCheckedAt   time.Time
	BlockReason     string
	ResumeCondition string
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
	State           Field[string]
	Stage           Field[string]
	Attempt         Field[int]
	Epoch           Field[int]
	HeadSHA         Field[string]
	LastCheckedAt   Field[time.Time]
	BlockReason     Field[string]
	ResumeCondition Field[string]
	MirrorState     Field[string]
	Checks          Field[[]CheckRun]
	NextAction      Field[string]

	// ChecksStale is set when any returned check ran against a HEAD other than the claim's HEAD.
	ChecksStale bool
}

// Collector gathers rows. All fields are read-only sources.
type Collector struct {
	Claims ClaimReader
	Remote RemoteReader
	Ledger LedgerReader
	Now    func() time.Time
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
		r.State = textField("state", cr.State, SrcClaimStore, at)
		r.Stage = textField("stage", cr.Stage, SrcClaimStore, at)
		r.Attempt = Field[int]{cr.Attempt, SrcClaimStore, at, true, false}
		r.Epoch = Field[int]{cr.Epoch, SrcClaimStore, at, true, false}
		r.HeadSHA = textField("head", cr.HeadSHA, SrcClaimStore, at)
		r.LastCheckedAt = Field[time.Time]{cr.LastCheckedAt, SrcClaimStore, at, true, false}
		r.BlockReason = textField("block_reason", cr.BlockReason, SrcClaimStore, at)
		r.ResumeCondition = textField("resume_condition", cr.ResumeCondition, SrcClaimStore, at)
		r.MirrorState = textField("mirror_state", cr.MirrorState, SrcClaimStore, at)
	} else {
		r.Owner.Source, r.State.Source, r.Stage.Source = SrcClaimStore, SrcClaimStore, SrcClaimStore
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
