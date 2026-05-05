package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ErrGhUnavailable is returned when the gh CLI is not available or not authenticated.
var ErrGhUnavailable = errors.New("gh CLI unavailable or not authenticated")

// Confirmer is an interface for asking the user a yes/no question.
// The orchestrator uses an interactive TTY confirmer by default; the --yes-*
// flags inject a yesConfirmer so subagents can drive the flow non-interactively.
type Confirmer interface {
	Confirm(prompt string) (bool, error)
}

// GhClient is an interface for invoking the gh CLI.
// The real implementation shells out; tests inject stubs.
type GhClient interface {
	// Run executes gh with the given arguments and returns stdout.
	Run(ctx context.Context, args ...string) ([]byte, error)
	// RunWithStdin executes gh with the given arguments, writing stdin as the request body.
	RunWithStdin(ctx context.Context, stdin string, args ...string) ([]byte, error)
}

// ttyConfirmer reads a y/n answer from the user's terminal.
type ttyConfirmer struct{}

// Confirm prompts the user on stderr and reads a single line from stdin.
func (t *ttyConfirmer) Confirm(prompt string) (bool, error) {
	fmt.Printf("%s [y/N]: ", prompt)
	var answer string
	if _, err := fmt.Scanln(&answer); err != nil {
		return false, nil
	}
	return strings.EqualFold(strings.TrimSpace(answer), "y"), nil
}

// yesConfirmer always returns true — used with --yes-branch-protection flag.
type yesConfirmer struct{}

// Confirm always returns true without prompting.
func (y *yesConfirmer) Confirm(_ string) (bool, error) {
	return true, nil
}

// realGhClient invokes the gh CLI via os/exec.
// Tests should inject a stub via the GhClient interface instead.
type realGhClient struct{}

// NewRealGhClient returns a GhClient that shells out to the real gh CLI.
func NewRealGhClient() GhClient { return &realGhClient{} }

// Run executes gh with the given arguments and returns stdout.
func (r *realGhClient) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	return cmd.Output()
}

// RunWithStdin executes gh with the given arguments, writing stdin as the request body.
func (r *realGhClient) RunWithStdin(ctx context.Context, stdin string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Stdin = strings.NewReader(stdin)
	return cmd.Output()
}

// DiscoverOwnerRepo derives owner+repo by invoking `gh repo view --json nameWithOwner`.
// Returns ErrGhUnavailable if gh is missing/unauthenticated.
func DiscoverOwnerRepo(ctx context.Context, gh GhClient) (owner, repo string, err error) {
	out, err := gh.Run(ctx, "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	if err != nil {
		return "", "", fmt.Errorf("%w: gh repo view failed: %v", ErrGhUnavailable, err)
	}
	full := strings.TrimSpace(string(out))
	parts := strings.SplitN(full, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected gh output %q (want owner/repo)", full)
	}
	return parts[0], parts[1], nil
}

// branchProtectionTemplate is the Go text/template for the gh API JSON payload.
// It renders the required_status_checks contexts from the SSoT YAML.
const branchProtectionTemplate = `{
  "required_status_checks": {
    "strict": true,
    "contexts": [{{range $i, $c := .Contexts}}{{if $i}}, {{end}}{{printf "%q" $c}}{{end}}]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_linear_history": false,
  "required_conversation_resolution": true
}`

// RenderBranchProtectionJSON renders the gh API JSON payload for the given branch
// using contexts from the RequiredChecks SSoT. Returns an error if the branch
// is not found in the checks.
func RenderBranchProtectionJSON(checks *config.RequiredChecks, branch string) (string, error) {
	bc, ok := checks.Branches[branch]
	if !ok {
		return "", fmt.Errorf("branch %q not found in required-checks.yml; available: %v",
			branch, branchKeys(checks))
	}

	tmpl, err := template.New("bp").Parse(branchProtectionTemplate)
	if err != nil {
		return "", fmt.Errorf("parse branch protection template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, bc); err != nil {
		return "", fmt.Errorf("render branch protection template: %w", err)
	}
	return buf.String(), nil
}

// PreflightGh checks that the gh CLI is available and authenticated.
// Returns ErrGhUnavailable if gh is missing, unauthenticated, or errors.
func PreflightGh(ctx context.Context, gh GhClient) error {
	_, err := gh.Run(ctx, "auth", "status")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGhUnavailable, err)
	}
	return nil
}

// ApplyBranchProtection applies branch protection rules to the given branch
// using the gh CLI. If gh is unavailable or unauthenticated, it returns an
// error containing the exact manual command for copy-paste.
func ApplyBranchProtection(ctx context.Context, gh GhClient, owner, repo, branch string, checks *config.RequiredChecks) error {
	// Preflight: verify gh is available and authenticated.
	if err := PreflightGh(ctx, gh); err != nil {
		json, renderErr := RenderBranchProtectionJSON(checks, branch)
		if renderErr != nil {
			json = "<render error: " + renderErr.Error() + ">"
		}
		return fmt.Errorf(
			"%w\n\nRun manually:\n  gh api -X PUT /repos/%s/%s/branches/%s/protection --input - <<'EOF'\n%s\nEOF",
			err, owner, repo, branch, json,
		)
	}

	// Render the JSON payload.
	json, err := RenderBranchProtectionJSON(checks, branch)
	if err != nil {
		return fmt.Errorf("render branch protection JSON: %w", err)
	}

	// URL-encode the branch name. GitHub branch protection API requires that
	// `/` in branch names (e.g. release/*) be encoded as %2F or the path will
	// be misinterpreted as additional URL segments and return 404.
	encodedBranch := strings.ReplaceAll(branch, "/", "%2F")

	// Apply via gh api.
	apiPath := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", owner, repo, encodedBranch)
	_, err = gh.RunWithStdin(ctx, json, "api", "-X", "PUT", apiPath, "--input", "-")
	if err != nil {
		return fmt.Errorf(
			"gh api PUT %s failed: %w\n\nRun manually:\n  gh api -X PUT /repos/%s/%s/branches/%s/protection --input - <<'EOF'\n%s\nEOF",
			apiPath, err, owner, repo, encodedBranch, json,
		)
	}

	return nil
}

// branchKeys returns the sorted keys from the RequiredChecks branches map.
func branchKeys(checks *config.RequiredChecks) []string {
	keys := make([]string, 0, len(checks.Branches))
	for k := range checks.Branches {
		keys = append(keys, k)
	}
	return keys
}
