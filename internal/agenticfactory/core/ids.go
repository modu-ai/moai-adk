package core

import (
	"fmt"
	"regexp"
)

var (
	idPattern   = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)
	headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// RunID identifies one factory run. It is used as a path segment, so it
// follows the same allowlist as slot-lease resource names.
type RunID string

// AttemptID identifies one attempt inside a run.
type AttemptID string

// ParseRunID validates s against ^[a-z0-9-]{1,64}$.
func ParseRunID(s string) (RunID, error) {
	if !idPattern.MatchString(s) {
		return "", fmt.Errorf("%w: run id %q", ErrIDInvalid, s)
	}
	return RunID(s), nil
}

// ParseAttemptID validates s against ^[a-z0-9-]{1,64}$.
func ParseAttemptID(s string) (AttemptID, error) {
	if !idPattern.MatchString(s) {
		return "", fmt.Errorf("%w: attempt id %q", ErrIDInvalid, s)
	}
	return AttemptID(s), nil
}

// HeadSHA is a full 40-character lowercase hex commit id. Abbreviated or
// mixed-case values are rejected so evidence can be compared byte-for-byte.
type HeadSHA string

// ParseHeadSHA validates s as a full lowercase SHA-1 hex string.
func ParseHeadSHA(s string) (HeadSHA, error) {
	if !headPattern.MatchString(s) {
		return "", fmt.Errorf("%w: %q", ErrHeadInvalid, s)
	}
	return HeadSHA(s), nil
}

// sessionPattern admits UUID-like session ids and nothing that could escape a
// path segment.
var sessionPattern = regexp.MustCompile(`^[0-9A-Za-z-]{8,64}$`)

// Agent is an agent session's identity: the immutable session id, and the
// number that gives it the display name agent-<number>. The two never stand in
// for each other.
type Agent struct {
	SessionID string `json:"session_id"`
	Number    int    `json:"number"`
}

// DisplayName is agent-<number>.
func (a Agent) DisplayName() string { return fmt.Sprintf("agent-%d", a.Number) }

// ParseSessionID validates a session id.
func ParseSessionID(s string) (string, error) {
	if !sessionPattern.MatchString(s) {
		return "", fmt.Errorf("%w: session id %q", ErrIDInvalid, s)
	}
	return s, nil
}

// IssueRef names the GitHub Issue a claim belongs to.
type IssueRef struct{ Number int }

// NewIssueRef rejects non-positive issue numbers.
func NewIssueRef(n int) (IssueRef, error) {
	if n <= 0 {
		return IssueRef{}, fmt.Errorf("%w: %d", ErrIssueInvalid, n)
	}
	return IssueRef{Number: n}, nil
}

// Resource is the claim key, e.g. "issue-1948".
func (r IssueRef) Resource() string { return fmt.Sprintf("issue-%d", r.Number) }
