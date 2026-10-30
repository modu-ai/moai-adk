package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// HostIssue is what a host identity lookup says about one Issue: the numeric id of
// its author, the repository the host really answered for and whether a redirect
// intervened. An author id that is not a positive number is not an identity.
type HostIssue struct {
	AuthorID   int64
	ActualRepo string
	Redirected bool
}

// HostFacts is the production Facts: it answers the permission decision from a
// host identity lookup, read-only. The lookups are functions so the code that talks
// to the host (the GitHub identity reader) is wired in where it is built: its
// CurrentUserID is User, and its IssueFacts, with the fields copied into HostIssue,
// is Issue. Every shortfall is an error, which the guard reads as an unknown
// identity, and every host lookup (the user, the Issue) is cut off by its own Timeout, derived from the
// caller's context (Runner.Ctx; Background when nil). The claim-store owner read is local and unbounded here.
type HostFacts struct {
	// Repo is the repository this store works in ("owner/name").
	Repo string
	// Timeout bounds each lookup; there is no default, and zero is refused.
	Timeout time.Duration
	User    func(ctx context.Context) (int64, error)
	Issue   func(ctx context.Context, repo string, issue int) (HostIssue, error)
	// Owner says who owns the work on the Issue now; StoreOwner reads it from the
	// claim store.
	Owner func(issue core.IssueRef) (core.OwnerState, error)
}

// IssueFacts implements Facts.
func (h HostFacts) IssueFacts(parent context.Context, issue core.IssueRef) (IssueFacts, error) {
	switch {
	case h.Repo == "":
		return IssueFacts{}, errors.New("hostfacts: no repository is configured")
	case h.Timeout <= 0:
		return IssueFacts{}, errors.New("hostfacts: no lookup time limit is configured")
	case h.User == nil || h.Issue == nil || h.Owner == nil:
		return IssueFacts{}, errors.New("hostfacts: a lookup is not wired")
	}
	if parent == nil {
		parent = context.Background()
	}
	// Each host lookup gets the whole Timeout to itself, derived from the caller's context: a slow user
	// lookup must not shorten the Issue lookup, and cancelling the caller's context reaches both.
	uctx, ucancel := context.WithTimeout(parent, h.Timeout)
	defer ucancel()
	user, err := h.User(uctx)
	if err != nil {
		return IssueFacts{}, fmt.Errorf("hostfacts: user lookup: %w", err)
	}
	if user <= 0 {
		return IssueFacts{}, errors.New("hostfacts: the acting user has no numeric id")
	}
	ictx, icancel := context.WithTimeout(parent, h.Timeout)
	defer icancel()
	hi, err := h.Issue(ictx, h.Repo, issue.Number)
	if err != nil {
		return IssueFacts{}, fmt.Errorf("hostfacts: issue lookup: %w", err)
	}
	if hi.AuthorID <= 0 {
		return IssueFacts{}, errors.New("hostfacts: the Issue's author has no numeric id")
	}
	owner, err := h.Owner(issue)
	if err != nil {
		return IssueFacts{}, fmt.Errorf("hostfacts: owner lookup: %w", err)
	}
	return IssueFacts{
		RequestedRepo: h.Repo, ActualRepo: hi.ActualRepo, Redirected: hi.Redirected,
		UserID: user, AuthorID: hi.AuthorID, Owner: owner,
	}, nil
}

// StoreOwner reads who owns the work on an Issue from the claim record: no record,
// or one with neither holder nor receipt, is nobody; a holder is active; a released
// record that carries a receipt was handed over. A record that cannot be read is an
// error, not "nobody".
func StoreOwner(st *claim.Store) func(core.IssueRef) (core.OwnerState, error) {
	return func(i core.IssueRef) (core.OwnerState, error) {
		rec, err := st.Read(i)
		switch {
		case errors.Is(err, core.ErrNotHeld):
			return core.OwnerNone, nil
		case err != nil:
			return core.OwnerUnknown, err
		case rec.Holder != "":
			return core.OwnerActive, nil
		case rec.Receipt != nil:
			return core.OwnerHandedOver, nil
		}
		return core.OwnerNone, nil
	}
}
