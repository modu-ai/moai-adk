package reconcile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// PR handoff reconcile (S46, HANDOFF01-03 and 08). A published PR is not completion: the agent's
// claim is released only after the leader's acceptance is seen against the same remote PR HEAD,
// and only once. Everything is read from durable receipt plus remote facts; nothing is inferred
// from a reply or an idle notice.

// Receipt is the durable PR handoff record, looked up by handoff id.
type Receipt struct {
	HandoffID      string
	Issue          int
	PR             int
	Head           string
	LeaderAccepted bool
	ClaimReleased  bool
}

// RemotePR is the PR as the remote reports it.
type RemotePR struct {
	Found  bool
	Number int
	Head   string
	Merged bool
}

// HandoffLookup is the read-only fact source.
type HandoffLookup interface {
	LookupHandoff(ctx context.Context, handoffID string) (Receipt, bool, error)
	RemotePR(ctx context.Context, issue int) (RemotePR, error)
}

// HandoffID is the deterministic id of a handoff: the same Issue, PR and HEAD always give the same id,
// so a lost response is resolved by looking the id up.
func HandoffID(issue, pr int, head string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("handoff\x00%d\x00%d\x00%s", issue, pr, head)))
	return hex.EncodeToString(h[:])[:32]
}

type HandoffKind string

const (
	HandoffUnconfirmed           HandoffKind = "unconfirmed"
	HandoffNone                  HandoffKind = "no-handoff"
	HandoffAcceptanceUnconfirmed HandoffKind = "acceptance-unconfirmed"
	HandoffPRMissing             HandoffKind = "pr-missing-on-remote"
	HandoffStale                 HandoffKind = "stale-after-head-change"
	HandoffAccepted              HandoffKind = "accepted"
)

// Closure owners.
const (
	OwnerAgent  = "agent"
	OwnerLeader = "leader"
)

// HandoffDecision is a recommendation. It has no field that finishes the Issue or starts a writer.
type HandoffDecision struct {
	Kind             HandoffKind
	ReleaseClaim     bool // release the agent's claim now (true at most once per handoff)
	NextDispatch     bool // the freed agent may take a new Issue
	NewWriterAllowed bool // always false here: a handoff never creates a second writer
	NeedsNewHandoff  bool // the HEAD moved; earlier writer/audit evidence cannot be reused
	Done             bool // always false here: completion needs the done evidence
	ClosureOwner     string
	Detail           string
}

// ReconcileHandoff decides from the receipt and the remote PR.
func ReconcileHandoff(ctx context.Context, lk HandoffLookup, issue int, handoffID string) HandoffDecision {
	unconfirmed := func(why string) HandoffDecision {
		return HandoffDecision{Kind: HandoffUnconfirmed, ClosureOwner: OwnerAgent, Detail: why}
	}
	rec, haveRec, err := lk.LookupHandoff(ctx, handoffID)
	if err != nil {
		return unconfirmed("handoff receipt could not be read")
	}
	pr, err := lk.RemotePR(ctx, issue)
	if err != nil {
		return unconfirmed("remote PR could not be read")
	}
	switch {
	case !haveRec && !pr.Found:
		return HandoffDecision{Kind: HandoffNone, ClosureOwner: OwnerAgent}
	case !haveRec:
		return HandoffDecision{Kind: HandoffAcceptanceUnconfirmed, ClosureOwner: OwnerAgent,
			Detail: fmt.Sprintf("PR #%d is published but no handoff receipt exists; not done, claim kept", pr.Number)}
	case !pr.Found:
		return HandoffDecision{Kind: HandoffPRMissing, ClosureOwner: OwnerAgent, Detail: "receipt exists but the remote shows no PR"}
	case pr.Head != rec.Head:
		return HandoffDecision{Kind: HandoffStale, NeedsNewHandoff: true, ClosureOwner: OwnerAgent,
			Detail: fmt.Sprintf("receipt HEAD %.12s, remote PR HEAD %.12s", rec.Head, pr.Head)}
	case !rec.LeaderAccepted:
		return HandoffDecision{Kind: HandoffAcceptanceUnconfirmed, ClosureOwner: OwnerAgent, Detail: "leader acceptance not recorded"}
	}
	return HandoffDecision{Kind: HandoffAccepted, ReleaseClaim: !rec.ClaimReleased, NextDispatch: true, ClosureOwner: OwnerLeader}
}
