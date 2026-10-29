package review

import (
	"errors"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

var (
	// ErrRepairIssue: the repair result belongs to a different Issue.
	ErrRepairIssue = errors.New("review: repair result is for a different issue")
	// ErrRepairBranch: the repair result was made on a different branch.
	ErrRepairBranch = errors.New("review: repair result is from a different branch")
	// ErrRepairBinding: the binding itself is incomplete, so it accepts nothing.
	ErrRepairBinding = errors.New("review: repair binding is incomplete")
)

// RepairBinding ties a repair dispatch to its original Issue, its branch, and
// that Issue's single claim holder (#1948 DELTA4).
type RepairBinding struct {
	Issue  core.IssueRef
	Branch string
	// HolderSession is the holder's immutable session ID; it is the only
	// holder field compared. A display name such as agent-N is not an identity.
	HolderSession string
	DisplayName   string // informational only, never compared
}

// RepairResult is what a repair attempt reports back.
type RepairResult struct {
	Issue         core.IssueRef
	Branch        string
	HolderSession string
	DisplayName   string // informational only, never compared
	Head          core.HeadSHA
}

// Accept returns nil only for a result from the bound Issue, branch, and
// holder with a full HEAD. A result from another holder is core.ErrForeign.
func (b RepairBinding) Accept(r RepairResult) error {
	if b.Issue.Number <= 0 || b.Branch == "" || b.HolderSession == "" {
		return ErrRepairBinding
	}
	switch {
	case r.Issue != b.Issue:
		return fmt.Errorf("%w: got #%d, bound #%d", ErrRepairIssue, r.Issue.Number, b.Issue.Number)
	case r.Branch != b.Branch:
		return fmt.Errorf("%w: got %q, bound %q", ErrRepairBranch, r.Branch, b.Branch)
	case r.HolderSession != b.HolderSession:
		return fmt.Errorf("%w: result holder differs from the bound holder", core.ErrForeign)
	}
	if _, err := core.ParseHeadSHA(string(r.Head)); err != nil {
		return err
	}
	return nil
}
