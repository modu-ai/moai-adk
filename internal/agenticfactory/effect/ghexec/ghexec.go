// Package ghexec runs the gh command line for the agentic-factory components (Issue #1948): one os/exec
// wrapper that fits ghpr.Runner, ghintake.Runner, adapt.GHRunner and ctxledger.Runner.
//
// It is off by default in every sense that matters. Nothing wires it in; the caller passes it where a
// Runner is wanted. It refuses, before any process exists, every gh call that is not on its read
// allowlist, and the one write it can ever run - creating an Issue, through the protection guard - only
// when AllowWrite is set. Comment writes are never run here: the protection guard leaves them to the
// author policy (effect.Enforcer), which needs an identity this layer does not have, so the context
// ledger's comment writes cannot go through this wrapper (its reads can). The components in front of it
// keep their own allowlists and guards; this is the last gate, not the only one. Real use against GitHub,
// writes above all, needs the user's approval.
//
// Arguments are passed to the process as they are, never through a shell. Standard output and standard
// error are kept apart and both are capped. A non-zero exit is an error, returned together with whatever
// standard output the process produced (gh pr checks prints its JSON and then exits non-zero while a check
// fails or is pending).
package ghexec

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/internal/procrun"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// ErrNotAllowed is returned, with no process started, for a gh call outside the allowlist.
var ErrNotAllowed = errors.New("ghexec: this gh call is not allowed")

// ErrOutputTooLarge is returned when gh printed more than MaxOutput bytes; the process is stopped and
// nothing is returned as if it were complete.
var ErrOutputTooLarge = fmt.Errorf("ghexec: gh output exceeded the configured limit: %w", procrun.ErrOutputTooLarge)

// Exec runs gh.
type Exec struct {
	// Bin is the program to run; empty means "gh" found on PATH.
	Bin string
	Dir string
	// Env is added to the environment the process inherits (later entries win). The wrapper itself adds
	// only what keeps gh from stopping to ask a question or paint the output.
	Env []string
	// MaxOutput bounds standard output and, separately, standard error, in bytes. It comes from
	// configuration and is required: zero refuses every call.
	MaxOutput int64
	// WaitDelay bounds how long to wait for the output pipes after the process was stopped. Zero waits as
	// long as the pipes stay open.
	WaitDelay time.Duration
	// AllowWrite lets `gh issue create` run (with a --repo that passes the protection guard). It is false
	// by default; no other write is ever allowed.
	AllowWrite bool
}

// ExitError is a non-zero exit of gh.
type ExitError struct {
	Verb   string // the first two arguments only: the rest may carry text that must not be echoed
	Code   int
	Stderr string
	status *forge.StatusError
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("ghexec: gh %s exited %d: %s", e.Verb, e.Code, strings.TrimSpace(e.Stderr))
}

// Unwrap exposes the typed HTTP failure when gh reported one, so callers decide on the status and never
// on wording.
func (e *ExitError) Unwrap() error {
	if e.status == nil {
		return nil
	}
	return e.status
}

var httpStatus = regexp.MustCompile(`\(HTTP (\d{3})\)`)

func verbOf(args []string) string {
	if len(args) >= 2 {
		return args[0] + " " + args[1]
	}
	return strings.Join(args, " ")
}

// Run runs gh with args. It returns the standard output; on a non-zero exit it returns it together with an
// *ExitError.
func (e Exec) Run(ctx context.Context, args ...string) ([]byte, error) {
	if e.MaxOutput <= 0 {
		return nil, errors.New("ghexec: MaxOutput must come from configuration")
	}
	if !e.allowed(args) {
		return nil, fmt.Errorf("%w: gh %s", ErrNotAllowed, verbOf(args))
	}
	bin := e.Bin
	if bin == "" {
		bin = "gh"
	}
	out, err := procrun.Run(ctx, procrun.Spec{
		Bin: bin, Args: args, Dir: e.Dir, MaxOutput: e.MaxOutput, WaitDelay: e.WaitDelay,
		Base: []string{"GH_PROMPT_DISABLED=1", "GH_NO_UPDATE_NOTIFIER=1", "NO_COLOR=1"}, Env: e.Env,
	})
	var pe *procrun.ExitError
	switch {
	case err == nil:
		return out, nil
	case errors.Is(err, procrun.ErrOutputTooLarge):
		return nil, fmt.Errorf("%w: gh %s", ErrOutputTooLarge, verbOf(args))
	case errors.As(err, &pe):
		x := &ExitError{Verb: verbOf(args), Code: pe.Code, Stderr: pe.Stderr}
		if m := httpStatus.FindStringSubmatch(x.Stderr); m != nil {
			code, _ := strconv.Atoi(m[1])
			x.status = &forge.StatusError{Status: code, Body: []byte(x.Stderr)}
		}
		return out, x
	}
	return nil, fmt.Errorf("ghexec: gh %s: %w", verbOf(args), err)
}

// --- the allowlist -----------------------------------------------------------------------------------

var (
	// flags of `gh api` this wrapper understands; the value says whether the flag takes a value.
	apiFlags = map[string]bool{
		"--paginate": false, "--slurp": false, "--silent": false, "-i": false, "--include": false,
		"--jq": true, "-q": true, "--template": true, "-t": true, "-H": true, "--header": true,
		"--cache": true, "-X": true, "--method": true,
		"-f": true, "-F": true, "--field": true, "--raw-field": true, "--input": true,
	}
	bodyFlags = map[string]bool{"-f": true, "-F": true, "--field": true, "--raw-field": true, "--input": true}
)

func (e Exec) allowed(args []string) bool {
	if len(args) < 2 {
		return false
	}
	for _, a := range args {
		if a == "--web" || a == "-w" { // opens a browser
			return false
		}
	}
	if args[0] == "api" { // its second argument is a path or a flag, so it cannot be matched as a pair
		method, _, hasBody, ok := parseAPI(args)
		if !ok {
			return false
		}
		return method == "GET" && !hasBody // any other method, or any body flag (-f turns a call into a POST), is a write
	}
	switch args[0] + " " + args[1] {
	case "pr view", "pr list", "pr checks", "issue view", "issue list":
		return true
	case "issue create":
		repo := flagValue(args, "--repo")
		return e.AllowWrite && repo != "" && effect.CheckRequest("POST", "/repos/"+repo+"/issues", nil) == nil
	}
	return false
}

// parseAPI reads the arguments after "api". ok is false for anything it does not recognise: an unknown
// flag fails closed, because gh turns some flags (-f) into a write by themselves.
func parseAPI(a []string) (method, path string, hasBody, ok bool) {
	method = "GET"
	for i := 1; i < len(a); i++ {
		arg := a[i]
		if !strings.HasPrefix(arg, "-") {
			if path != "" {
				return "", "", false, false
			}
			path = arg
			continue
		}
		name, val, hasEq := strings.Cut(arg, "=")
		takes, known := apiFlags[name]
		if !known {
			return "", "", false, false
		}
		if takes && !hasEq {
			i++
			if i >= len(a) {
				return "", "", false, false
			}
			val = a[i]
		}
		switch {
		case name == "-X" || name == "--method":
			method = strings.ToUpper(val)
		case bodyFlags[name]:
			hasBody = true
		}
	}
	return method, path, hasBody, path != ""
}

// flagValue returns the value of --name given as "--name value" or "--name=value".
func flagValue(args []string, name string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := strings.CutPrefix(a, name+"="); ok {
			return v
		}
	}
	return ""
}
