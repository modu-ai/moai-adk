// Package browser resolves which browser the AgenticFactory adapter may use.
//
// The operator's explicit selection always wins. With no selection the
// adapter uses chrome without saving it as a selection. A selected browser
// that is not installed, or whose connection check fails, blocks with a
// resume condition; no other browser is tried (SPEC-AGENTIC-FACTORY-001
// REQ-AF-016, Issue #1948).
package browser

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Choice is one of the three browsers an operator can select.
type Choice string

const (
	Egolite Choice = "egolite"
	Aside   Choice = "aside"
	Chrome  Choice = "chrome"
)

// Choices returns the exact set of selectable browsers, in display order.
func Choices() []Choice { return []Choice{Egolite, Aside, Chrome} }

// Reason is a block reason from the SPEC's blocked-reason set.
type Reason string

const (
	ReasonNotInstalled     Reason = "not-installed"
	ReasonConnectionFailed Reason = "connection-failed"
)

// ErrUnknownChoice reports a selection outside the three allowed values.
var ErrUnknownChoice = errors.New("browser: unknown choice")

// Probe runs the external commands the resolver needs. Tests supply a fake;
// production uses ExecProbe.
type Probe interface {
	// LookPath reports whether name is on PATH (the `command -v` check).
	LookPath(name string) error
	// Run executes name with args and returns its exit code and combined output.
	// A non-nil error means the command could not be started.
	Run(ctx context.Context, name string, args ...string) (code int, output string, err error)
}

// Resolution is the outcome of Resolve.
type Resolution struct {
	// Selected echoes the configured selection ("" when none); it is never
	// rewritten, so a block leaves the operator's choice intact.
	Selected string
	// Use is the only browser the adapter may call; empty when blocked.
	Use Choice
	// Reason and Resume are set together when the resolution is blocked.
	Reason Reason
	Resume string
	// Evidence is a short excerpt of the command output behind a block.
	Evidence string
}

// Blocked reports whether the adapter must stop instead of calling a browser.
func (r Resolution) Blocked() bool { return r.Reason != "" }

const (
	egoBinary   = "ego-browser"
	egoReady    = "ego-browser ready"
	asideBin    = "aside"
	evidenceMax = 200
)

// Resolve maps the configured selection to the one browser the adapter may use.
func Resolve(ctx context.Context, selected string, p Probe) (Resolution, error) {
	res := Resolution{Selected: selected}
	switch Choice(selected) {
	case "", Chrome:
		res.Use = Chrome
	case Egolite:
		if err := p.LookPath(egoBinary); err != nil {
			return blockedRes(res, ReasonNotInstalled, "", "install ego-lite so `command -v ego-browser` succeeds, then retry"), nil
		}
		code, out, err := p.Run(ctx, egoBinary, "nodejs", "-e", "console.log('ego-browser ready')")
		if err != nil || code != 0 || !strings.Contains(out, egoReady) {
			return blockedRes(res, ReasonConnectionFailed, out, "start the ego lite app so `ego-browser nodejs` prints `ego-browser ready`, then retry"), nil
		}
		res.Use = Egolite
	case Aside:
		code, out, err := p.Run(ctx, asideBin, "--version")
		if err != nil || code != 0 {
			return blockedRes(res, ReasonNotInstalled, out, "install aside so `aside --version` succeeds, then retry"), nil
		}
		res.Use = Aside
	default:
		return res, fmt.Errorf("%w: %q", ErrUnknownChoice, selected)
	}
	return res, nil
}

func blockedRes(r Resolution, reason Reason, evidence, resume string) Resolution {
	evidence = strings.TrimSpace(evidence)
	if len(evidence) > evidenceMax {
		evidence = evidence[:evidenceMax]
	}
	r.Reason, r.Evidence, r.Resume = reason, evidence, resume
	return r
}

// ExecProbe is the production Probe backed by os/exec.
type ExecProbe struct{}

// LookPath implements Probe.
func (ExecProbe) LookPath(name string) error {
	_, err := exec.LookPath(name)
	return err
}

// Run implements Probe with a bounded timeout.
func (ExecProbe) Run(ctx context.Context, name string, args ...string) (int, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if ee, ok := errors.AsType[*exec.ExitError](err); ok {
		return ee.ExitCode(), string(out), nil
	}
	if err != nil {
		return -1, string(out), err
	}
	return 0, string(out), nil
}
