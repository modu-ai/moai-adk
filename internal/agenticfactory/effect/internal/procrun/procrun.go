// Package procrun is the one place the agentic-factory command wrappers (ghexec, gitexec) start a
// process: no shell, standard output and standard error kept apart and each capped, the process stopped
// when the context ends or an output limit is passed, a non-zero exit returned as an error together with
// the standard output. What a wrapper allows to run is the wrapper's business; this only runs it safely.
package procrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// ErrOutputTooLarge is returned when a stream passed its limit; the process is stopped.
var ErrOutputTooLarge = errors.New("procrun: output exceeded the configured limit")

// Spec describes one process.
type Spec struct {
	Bin  string
	Args []string
	Dir  string
	// Env is added to the inherited environment (later entries win), after Base.
	Base, Env []string
	// MaxOutput bounds each of stdout and stderr, in bytes. It must be positive.
	MaxOutput int64
	// WaitDelay bounds how long to wait for the pipes after the process was stopped; zero waits as long as
	// they stay open.
	WaitDelay time.Duration
}

// ExitError is a non-zero exit.
type ExitError struct {
	Code   int
	Stderr string
}

func (e *ExitError) Error() string { return fmt.Sprintf("exited %d: %s", e.Code, e.Stderr) }

type capWriter struct {
	buf  bytes.Buffer
	max  int64
	over bool
	stop func()
}

// Write keeps at most max bytes and notes when more arrived; it never errors, so the process is not cut
// off by a broken pipe before the caller decides.
func (w *capWriter) Write(p []byte) (int, error) {
	if room := w.max - int64(w.buf.Len()); int64(len(p)) > room {
		if room > 0 {
			w.buf.Write(p[:room])
		}
		if !w.over {
			w.over = true
			w.stop()
		}
		return len(p), nil
	}
	w.buf.Write(p)
	return len(p), nil
}

// Run starts the process and waits for it. On success it returns stdout. On a non-zero exit it returns the
// stdout and an *ExitError. When ctx ends it returns the context's error; past a limit,
// ErrOutputTooLarge; in both of those nothing partial is returned as if it were complete.
func Run(ctx context.Context, s Spec) ([]byte, error) {
	if s.MaxOutput <= 0 {
		return nil, errors.New("procrun: MaxOutput must come from configuration")
	}
	rctx, stop := context.WithCancel(ctx)
	defer stop()
	cmd := exec.CommandContext(rctx, s.Bin, s.Args...)
	cmd.Dir = s.Dir
	cmd.Env = append(append(os.Environ(), s.Base...), s.Env...)
	cmd.WaitDelay = s.WaitDelay
	out, errw := &capWriter{max: s.MaxOutput, stop: stop}, &capWriter{max: s.MaxOutput, stop: stop}
	cmd.Stdout, cmd.Stderr = out, errw
	err := cmd.Run()
	switch {
	case ctx.Err() != nil:
		return nil, ctx.Err()
	case out.over || errw.over:
		return nil, ErrOutputTooLarge
	case err == nil:
		return out.buf.Bytes(), nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return out.buf.Bytes(), &ExitError{Code: ee.ExitCode(), Stderr: errw.buf.String()}
	}
	return nil, err
}
