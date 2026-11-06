package homestate

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// The evidence readers run read-only git subprocesses in the card's
// worktree. None of them fetches, checks out, merges, resets, or otherwise
// moves a ref: F1 reads git state, it never changes it.

// gitRead runs git in dir with the repository-scoping environment scrubbed
// and returns trimmed stdout.
func gitRead(ctx context.Context, dir string, args ...string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", errors.New("card has no worktree path")
	}
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = gitenv.Env()
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return "", fmt.Errorf("git %s: exit %d: %s", strings.Join(args, " "), exit.ExitCode(), strings.TrimSpace(string(exit.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// gitRemotes lists the configured remotes of the repository at dir.
func gitRemotes(ctx context.Context, dir string) ([]string, error) {
	out, err := gitRead(ctx, dir, "remote")
	if err != nil {
		return nil, err
	}
	var remotes []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			remotes = append(remotes, line)
		}
	}
	return remotes, nil
}

// hasRemote reports whether the repository at dir has any remote configured
// (the T17 / T18 discriminator).
func hasRemote(ctx context.Context, dir string) (bool, error) {
	remotes, err := gitRemotes(ctx, dir)
	if err != nil {
		return false, err
	}
	return len(remotes) > 0, nil
}

// PushGateTarget returns the state `decide --gate push` requests for a
// merged-local card by DEFAULT: `done`, whether or not the card's repository
// has a remote. The earlier answer — `pushed` when a remote exists — encoded
// the premise that a remote means push and CI come before done; the operator's
// Done standard (card t2059, root cause of t2057's follow-up) is local
// verification + the merge into local main + the leader's evidence receipt,
// with release CI, review and the remote merge managed apart. The store still
// requires the receipt on the done edge (REQ-FCR-002b). The pushed → ci-green
// → done route (T17/T19/T20) is taken only by an explicit choice, resolved by
// the caller (`--gate push --choice push`), not here.
func PushGateTarget(_ context.Context, _ Card) (string, error) {
	return CardDone, nil
}
