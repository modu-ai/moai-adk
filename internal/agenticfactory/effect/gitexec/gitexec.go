// Package gitexec runs the git command line for the two read-only questions the intake lookups ask
// (Issue #1948): is a branch on the remote (`ls-remote --heads <remote> [<pattern>]`) and which worktrees
// exist (`worktree list [--porcelain] [-z]`). It fits ghintake.Runner. Nothing wires it in, and nothing
// else is allowed to run through it.
//
// It refuses, before any process exists, every call that is not exactly one of those two shapes: another
// subcommand, a remote that is not a plain remote name (a URL, a path, or a transport such as ext::),
// another option, or one of git's own leading options (-c, -C) that could change what a read does.
// Arguments are passed to the process as they are, never through a shell; the process shares the
// size-capped, context-bound runner of ghexec (effect/internal/procrun). It never prompts: it sets
// GIT_TERMINAL_PROMPT=0.
package gitexec

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/internal/procrun"
)

// ErrNotAllowed is returned, with no process started, for a git call outside the two read shapes.
var ErrNotAllowed = errors.New("gitexec: this git call is not allowed")

// ErrOutputTooLarge is returned when git printed more than MaxOutput bytes; the process is stopped.
var ErrOutputTooLarge = fmt.Errorf("gitexec: git output exceeded the configured limit: %w", procrun.ErrOutputTooLarge)

// Exec runs git.
type Exec struct {
	// Bin is the program to run; empty means "git" found on PATH.
	Bin string
	Dir string
	// Env is added to the inherited environment (later entries win).
	Env []string
	// MaxOutput bounds standard output and, separately, standard error, in bytes. It comes from
	// configuration and is required: zero refuses every call.
	MaxOutput int64
	// WaitDelay bounds how long to wait for the output pipes after the process was stopped. Zero waits as
	// long as the pipes stay open.
	WaitDelay time.Duration
}

// ExitError is a non-zero exit of git.
type ExitError struct {
	Verb   string // the subcommand only: the rest of the call is not echoed
	Code   int
	Stderr string
}

func (e *ExitError) Error() string {
	msg, _ := core.Redact(strings.TrimSpace(e.Stderr)) // stored and logged: through the one sanitizer
	return fmt.Sprintf("gitexec: git %s exited %d: %s", e.Verb, e.Code, msg)
}

// remoteName is a configured remote's name; anything with a scheme, a colon, a slash or a leading dash is
// not one.
var remoteName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func verbOf(args []string) string {
	if len(args) >= 2 {
		return args[0] + " " + args[1]
	}
	return strings.Join(args, " ")
}

// allowed admits exactly the two read shapes.
func allowed(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch args[0] {
	case "ls-remote":
		if args[1] != "--heads" || len(args) < 3 || len(args) > 4 || !remoteName.MatchString(args[2]) {
			return false
		}
		if len(args) == 4 { // one ref pattern, never an option
			p := args[3]
			return p != "" && !strings.HasPrefix(p, "-") && !strings.ContainsRune(p, 0)
		}
		return true
	case "worktree":
		if args[1] != "list" {
			return false
		}
		seen := map[string]bool{}
		for _, a := range args[2:] {
			if (a != "--porcelain" && a != "-z") || seen[a] {
				return false
			}
			seen[a] = true
		}
		return true
	}
	return false
}

// Run runs git with args. It returns the standard output; on a non-zero exit it returns it together with
// an *ExitError.
func (e Exec) Run(ctx context.Context, args ...string) ([]byte, error) {
	if e.MaxOutput <= 0 {
		return nil, errors.New("gitexec: MaxOutput must come from configuration")
	}
	if !allowed(args) {
		return nil, fmt.Errorf("%w: git %s", ErrNotAllowed, verbOf(args))
	}
	bin := e.Bin
	if bin == "" {
		bin = "git"
	}
	out, err := procrun.Run(ctx, procrun.Spec{
		Bin: bin, Args: args, Dir: e.Dir, MaxOutput: e.MaxOutput, WaitDelay: e.WaitDelay,
		Base: []string{"GIT_TERMINAL_PROMPT=0"}, Env: e.Env,
	})
	var pe *procrun.ExitError
	switch {
	case err == nil:
		return out, nil
	case errors.Is(err, procrun.ErrOutputTooLarge):
		return nil, fmt.Errorf("%w: git %s", ErrOutputTooLarge, args[0])
	case errors.As(err, &pe):
		return out, &ExitError{Verb: args[0], Code: pe.Code, Stderr: pe.Stderr}
	}
	return nil, fmt.Errorf("gitexec: git %s: %w", args[0], err)
}
