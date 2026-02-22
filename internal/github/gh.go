package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ghBin caches the resolved gh binary path to avoid repeated exec.LookPath calls.
var (
	ghBinOnce sync.Once
	ghBinPath string
	ghBinErr  error
)

// gitBin caches the resolved git binary path to avoid repeated exec.LookPath calls.
var (
	gitBinOnce sync.Once
	gitBinPath string
	gitBinErr  error
)

// CheckConclusion represents the overall CI/CD check result.
type CheckConclusion string

const (
	// CheckPass indicates all CI/CD checks passed.
	CheckPass CheckConclusion = "pass"

	// CheckFail indicates one or more CI/CD checks failed.
	CheckFail CheckConclusion = "fail"

	// CheckPending indicates CI/CD checks are still running.
	CheckPending CheckConclusion = "pending"
)

// MergeMethod represents the Git merge strategy for a PR.
type MergeMethod string

const (
	// MergeMethodMerge creates a merge commit.
	MergeMethodMerge MergeMethod = "merge"

	// MergeMethodSquash squashes all commits into one.
	MergeMethodSquash MergeMethod = "squash"

	// MergeMethodRebase rebases commits onto the base branch.
	MergeMethodRebase MergeMethod = "rebase"
)

// PRCreateOptions holds parameters for creating a pull request.
type PRCreateOptions struct {
	Title       string
	Body        string
	BaseBranch  string
	HeadBranch  string
	Labels      []string
	IssueNumber int
}

// PRDetails holds information about an existing pull request.
type PRDetails struct {
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	State      string    `json:"state"`
	Mergeable  string    `json:"mergeable"`
	HeadBranch string    `json:"headRefName"`
	BaseBranch string    `json:"baseRefName"`
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Check represents a single CI/CD status check.
type Check struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// CheckStatus holds the aggregated CI/CD check results for a PR.
type CheckStatus struct {
	Overall CheckConclusion
	Checks  []Check
}

// GHClient abstracts GitHub CLI (gh) operations for testability.
type GHClient interface {
	// PRCreate creates a pull request and returns the PR number.
	PRCreate(ctx context.Context, opts PRCreateOptions) (int, error)

	// PRView retrieves PR details by number.
	PRView(ctx context.Context, number int) (*PRDetails, error)

	// PRMerge merges a PR by number using the specified method.
	// If deleteBranch is true, the head branch is deleted after merge.
	PRMerge(ctx context.Context, number int, method MergeMethod, deleteBranch bool) error

	// PRChecks returns the CI/CD check status for a PR.
	PRChecks(ctx context.Context, number int) (*CheckStatus, error)

	// Push pushes the current branch to the remote.
	Push(ctx context.Context, dir string) error

	// IsAuthenticated checks whether gh is authenticated.
	IsAuthenticated(ctx context.Context) error
}

// execFunc is the function signature for executing gh CLI commands.
// Used for dependency injection in tests.
type execFunc func(ctx context.Context, dir string, args ...string) (string, error)

// pushFunc is the function signature for pushing to remote.
// Used for dependency injection in tests.
type pushFunc func(ctx context.Context, workDir string) error

// ghClient implements GHClient using the gh CLI binary.
type ghClient struct {
	root   string
	logger *slog.Logger
	// execFn is the function used to execute gh commands.
	// If nil, the package-level execGH function is used.
	execFn execFunc
	// pushFn is the function used to push to remote.
	// If nil, the default git push implementation is used.
	pushFn pushFunc
}

// Compile-time interface compliance check.
var _ GHClient = (*ghClient)(nil)

// NewGHClient creates a new GitHub CLI client rooted at the given directory.
func NewGHClient(root string) *ghClient {
	return &ghClient{
		root:   root,
		logger: slog.Default().With("module", "github"),
	}
}

// newGHClientWithExec creates a ghClient with a custom exec function for testing.
func newGHClientWithExec(root string, fn execFunc) *ghClient {
	return &ghClient{
		root:   root,
		logger: slog.Default().With("module", "github"),
		execFn: fn,
	}
}

// newGHClientWithPush creates a ghClient with a custom push function for testing.
func newGHClientWithPush(root string, fn pushFunc) *ghClient {
	return &ghClient{
		root:   root,
		logger: slog.Default().With("module", "github"),
		pushFn: fn,
	}
}

// exec runs a gh command using execFn if set, otherwise falls back to execGH.
func (c *ghClient) exec(ctx context.Context, args ...string) (string, error) {
	if c.execFn != nil {
		return c.execFn(ctx, c.root, args...)
	}
	return execGH(ctx, c.root, args...)
}

// IsAuthenticated checks whether the gh CLI is authenticated.
func (c *ghClient) IsAuthenticated(ctx context.Context) error {
	_, err := c.exec(ctx, "auth", "status")
	if err != nil {
		return fmt.Errorf("check auth: %w", ErrGHNotAuthenticated)
	}
	return nil
}

// PRCreate creates a pull request and returns the new PR number.
func (c *ghClient) PRCreate(ctx context.Context, opts PRCreateOptions) (int, error) {
	args := []string{
		"pr", "create",
		"--title", opts.Title,
		"--body", opts.Body,
	}
	if opts.BaseBranch != "" {
		args = append(args, "--base", opts.BaseBranch)
	}
	if opts.HeadBranch != "" {
		args = append(args, "--head", opts.HeadBranch)
	}
	for _, label := range opts.Labels {
		args = append(args, "--label", label)
	}

	c.logger.Debug("creating pull request", "title", opts.Title, "base", opts.BaseBranch)

	output, err := c.exec(ctx, args...)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return 0, fmt.Errorf("create PR: %w", ErrPRAlreadyExists)
		}
		return 0, fmt.Errorf("create PR: %w", err)
	}

	// gh pr create returns the PR URL; extract the number from it.
	number, err := extractPRNumber(output)
	if err != nil {
		return 0, fmt.Errorf("parse PR number from %q: %w", output, err)
	}

	c.logger.Info("pull request created", "number", number, "title", opts.Title)
	return number, nil
}

// PRView retrieves pull request details by number.
func (c *ghClient) PRView(ctx context.Context, number int) (*PRDetails, error) {
	output, err := c.exec(ctx,
		"pr", "view", strconv.Itoa(number),
		"--json", "number,title,state,mergeable,headRefName,baseRefName,url,createdAt",
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Could not resolve") {
			return nil, fmt.Errorf("view PR #%d: %w", number, ErrPRNotFound)
		}
		return nil, fmt.Errorf("view PR #%d: %w", number, err)
	}

	var details PRDetails
	if err := json.Unmarshal([]byte(output), &details); err != nil {
		return nil, fmt.Errorf("parse PR #%d JSON: %w", number, err)
	}

	return &details, nil
}

// PRMerge merges a pull request using the specified method.
// If deleteBranch is true, the head branch is deleted after merge.
func (c *ghClient) PRMerge(ctx context.Context, number int, method MergeMethod, deleteBranch bool) error {
	args := []string{"pr", "merge", strconv.Itoa(number)}

	switch method {
	case MergeMethodMerge:
		args = append(args, "--merge")
	case MergeMethodSquash:
		args = append(args, "--squash")
	case MergeMethodRebase:
		args = append(args, "--rebase")
	default:
		return fmt.Errorf("merge PR #%d: unsupported merge method %q", number, method)
	}

	if deleteBranch {
		args = append(args, "--delete-branch")
	}

	c.logger.Debug("merging pull request", "number", number, "method", method)

	_, err := c.exec(ctx, args...)
	if err != nil {
		if strings.Contains(err.Error(), "conflict") {
			return fmt.Errorf("merge PR #%d: %w", number, ErrMergeConflict)
		}
		return fmt.Errorf("merge PR #%d: %w", number, err)
	}

	c.logger.Info("pull request merged", "number", number, "method", method)
	return nil
}

// PRChecks returns the CI/CD check status for a pull request.
func (c *ghClient) PRChecks(ctx context.Context, number int) (*CheckStatus, error) {
	output, err := c.exec(ctx,
		"pr", "checks", strconv.Itoa(number),
		"--json", "name,status,conclusion",
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("checks PR #%d: %w", number, ErrPRNotFound)
		}
		return nil, fmt.Errorf("checks PR #%d: %w", number, err)
	}

	var checks []Check
	if err := json.Unmarshal([]byte(output), &checks); err != nil {
		return nil, fmt.Errorf("parse checks JSON for PR #%d: %w", number, err)
	}

	status := &CheckStatus{
		Overall: deriveOverallConclusion(checks),
		Checks:  checks,
	}

	return status, nil
}

// Push pushes the current branch to the remote repository.
func (c *ghClient) Push(ctx context.Context, dir string) error {
	workDir := dir
	if workDir == "" {
		workDir = c.root
	}

	c.logger.Debug("pushing to remote", "dir", workDir)

	// Use injected push function if available (for testing).
	if c.pushFn != nil {
		return c.pushFn(ctx, workDir)
	}

	return pushWithGit(ctx, workDir)
}

// pushWithGit performs the actual git push operation.
func pushWithGit(ctx context.Context, workDir string) error {
	gitBinOnce.Do(func() {
		gitBinPath, gitBinErr = exec.LookPath("git")
	})
	if gitBinErr != nil {
		return fmt.Errorf("push: git not found: %w", gitBinErr)
	}

	cmd := exec.CommandContext(ctx, gitBinPath, "push", "-u", "origin", "HEAD")
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = runErr.Error()
		}
		return fmt.Errorf("push: %s: %w", errMsg, runErr)
	}

	return nil
}

// execGH runs a gh CLI command and returns its stdout output.
func execGH(ctx context.Context, dir string, args ...string) (string, error) {
	ghBinOnce.Do(func() {
		ghBinPath, ghBinErr = exec.LookPath("gh")
	})
	if ghBinErr != nil {
		return "", fmt.Errorf("gh lookup: %w", ErrGHNotFound)
	}

	cmd := exec.CommandContext(ctx, ghBinPath, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		if len(args) == 0 {
			return "", fmt.Errorf("gh: %s: %w", errMsg, err)
		}
		return "", fmt.Errorf("gh %s: %s: %w", args[0], errMsg, err)
	}

	return strings.TrimRight(stdout.String(), "\n\r"), nil
}

// extractPRNumber parses a PR number from the gh pr create output URL.
// The URL format is: https://github.com/owner/repo/pull/123
func extractPRNumber(output string) (int, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return 0, fmt.Errorf("empty output")
	}

	parts := strings.Split(output, "/")
	// A valid GitHub PR URL has the form: https://github.com/owner/repo/pull/123
	// After splitting by "/" that yields at least 7 parts:
	// ["https:", "", "github.com", "owner", "repo", "pull", "123"]
	// We require at least 2 trailing segments: "pull" and a number.
	if len(parts) < 2 {
		return 0, fmt.Errorf("unexpected URL format: %q", output)
	}

	if parts[len(parts)-2] != "pull" {
		return 0, fmt.Errorf("URL missing /pull/ segment: %q", output)
	}

	lastPart := parts[len(parts)-1]
	number, err := strconv.Atoi(lastPart)
	if err != nil {
		return 0, fmt.Errorf("parse PR number %q: %w", lastPart, err)
	}

	if number <= 0 {
		return 0, fmt.Errorf("invalid PR number: %d", number)
	}

	return number, nil
}

// deriveOverallConclusion computes the aggregate check status.
func deriveOverallConclusion(checks []Check) CheckConclusion {
	if len(checks) == 0 {
		return CheckPending
	}

	hasFailure := false
	hasPending := false

	for _, check := range checks {
		switch {
		case check.Conclusion == "failure" || check.Conclusion == "cancelled" || check.Conclusion == "timed_out":
			hasFailure = true
		case check.Status != "completed":
			hasPending = true
		}
	}

	switch {
	case hasFailure:
		return CheckFail
	case hasPending:
		return CheckPending
	default:
		return CheckPass
	}
}
