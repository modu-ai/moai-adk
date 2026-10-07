// Package publish is the publication half of the opt-in participation
// flow (SPEC-FEEDBACK-PARTICIPATION-001 design.md section 7): the sender
// that files through the USER'S OWN gh, the duplicate lookup, the issue
// contract, the deterministic template text, and (M6) the model seam.
//
// gh is the transport, full stop — no transport abstraction for anonymity,
// no net/http anywhere in this package, and os/exec in THIS FILE ONLY
// (REQ-ANON-025, enforced by the static import guard). Every GitHub
// operation runs under the caller's context, so the flush time box bounds
// the whole run.
package publish

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// ghSearchLimit bounds one duplicate lookup. The exact-title match needs
// no more; a fingerprint shared by more than ten open issues has already
// flooded past every cap this SPEC enforces locally.
const ghSearchLimit = "10"

// Runner is the gh seam: every GitHub operation the sender performs goes
// through it, so a test double proves the sender's discipline without a
// network (and M6's budget proofs count model calls at the same seam).
type Runner interface {
	// Available reports whether gh exists AND is authenticated. False
	// means the sender leaves every item queued, quietly.
	Available(ctx context.Context) bool
	// SearchIssues runs the duplicate lookup: issues whose TITLE matches
	// the search token (the fingerprint), any state, comments included.
	SearchIssues(ctx context.Context, repo, token string) ([]RemoteIssue, error)
	// CreateIssue files a new issue; body feeds --body-file -.
	CreateIssue(ctx context.Context, repo, title string, body io.Reader) error
	// CommentIssue adds one comment to an existing issue; body feeds
	// --body-file -.
	CommentIssue(ctx context.Context, repo string, number int, body io.Reader) error
}

// RemoteIssue is the slice of gh's issue JSON the contract needs.
type RemoteIssue struct {
	Number   int             `json:"number"`
	Title    string          `json:"title"`
	State    string          `json:"state"`
	URL      string          `json:"url"`
	Comments []RemoteComment `json:"comments"`
}

// RemoteComment is the slice of a comment object the occurrence count
// reads: the body only.
type RemoteComment struct {
	Body string `json:"body"`
}

// execRunner is the production Runner over the gh CLI. It is constructed
// ONLY here; tests substitute the interface.
type execRunner struct{}

// newExecRunner returns the production gh runner.
func newExecRunner() Runner { return &execRunner{} }

// ghTimeout bounds ONE gh invocation beyond the caller's context: a wedged
// gh must not hold a sender slot longer than this even when the caller
// passed a generous context.
const ghTimeout = 30 * time.Second

// rungh executes gh with args, feeding stdin, and returns its combined
// error state. Output is discarded: the sender needs exit codes, not gh's
// prose (quiet forms, bounded output).
func rungh(ctx context.Context, stdin io.Reader, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, ghTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", args...)
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("gh %s: %w", strings.Join(args[:min(2, len(args))], " "), ctx.Err())
		}
		return fmt.Errorf("gh %s: %w: %s", strings.Join(args[:min(2, len(args))], " "), err, firstLine(string(out)))
	}
	return nil
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 200 {
		s = s[:200]
	}
	return strings.TrimSpace(s)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Available probes gh's presence (exec.LookPath semantics via running
// `gh auth status`) — one call answering BOTH the missing-binary and the
// unauthenticated cases the sender must treat identically (leave queued,
// quietly).
func (r *execRunner) Available(ctx context.Context) bool {
	if _, err := exec.LookPath("gh"); err != nil {
		return false
	}
	return rungh(ctx, nil, "auth", "status") == nil
}

// SearchIssues runs the duplicate lookup (design section 7 step 3):
// `gh issue list --repo <repo> --state all --search "<token> in:title"
// --json number,title,state,url,comments --limit 10`. The query carries
// only the fingerprint, which leaks nothing.
func (r *execRunner) SearchIssues(ctx context.Context, repo, token string) ([]RemoteIssue, error) {
	args := []string{
		"issue", "list",
		"--repo", repo,
		"--state", "all",
		"--search", token + " in:title",
		"--json", "number,title,state,url,comments",
		"--limit", ghSearchLimit,
	}
	ctx, cancel := context.WithTimeout(ctx, ghTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue list: %w: %s", err, firstLine(string(out)))
	}
	var issues []RemoteIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, fmt.Errorf("gh issue list: parsing: %w", err)
	}
	return issues, nil
}

// CreateIssue files the report: `gh issue create --repo <repo> --title
// <title> --body-file -`. No --label is passed: labels depend on
// repository permission and a non-collaborator's labels are dropped or
// rejected (design section 7 step 5).
func (r *execRunner) CreateIssue(ctx context.Context, repo, title string, body io.Reader) error {
	return rungh(ctx, body, "issue", "create", "--repo", repo, "--title", title, "--body-file", "-")
}

// CommentIssue adds one occurrence comment: `gh issue comment <number>
// --repo <repo> --body-file -`.
func (r *execRunner) CommentIssue(ctx context.Context, repo string, number int, body io.Reader) error {
	return rungh(ctx, body, "issue", "comment", fmt.Sprintf("%d", number), "--repo", repo, "--body-file", "-")
}
