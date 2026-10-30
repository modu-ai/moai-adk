package procrun

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// The shared runner is exercised here on the local shell; the wrappers that use it are tested with fake
// gh and git scripts. Nothing reaches a network.

func sh(script string) Spec {
	return Spec{Bin: "/bin/sh", Args: []string{"-c", script}, MaxOutput: 10_000, WaitDelay: time.Second}
}

func TestStdoutAndStderrAreSeparateAndTheExitCodeIsAnError(t *testing.T) {
	out, err := Run(t.Context(), sh(`printf OUT; printf ERR >&2`))
	if err != nil || string(out) != "OUT" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	out, err = Run(t.Context(), sh(`printf partial; printf 'bad thing' >&2; exit 7`))
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 7 || ee.Stderr != "bad thing" || string(out) != "partial" {
		t.Errorf("out=%q err=%v", out, err)
	}
}

func TestTheLimitIsRequiredAndAStreamPastItStopsTheProcess(t *testing.T) {
	s := sh(`printf x`)
	s.MaxOutput = 0
	if _, err := Run(t.Context(), s); err == nil {
		t.Error("a run without a limit was accepted")
	}
	s = sh(`head -c 100000 /dev/zero | tr '\0' x`)
	s.MaxOutput = 1000
	if out, err := Run(t.Context(), s); !errors.Is(err, ErrOutputTooLarge) || out != nil {
		t.Errorf("stdout past the limit: %d bytes err=%v", len(out), err)
	}
	s = sh(`head -c 100000 /dev/zero | tr '\0' x >&2`)
	s.MaxOutput = 1000
	if _, err := Run(t.Context(), s); !errors.Is(err, ErrOutputTooLarge) {
		t.Errorf("stderr past the limit: %v", err)
	}
}

func TestTheContextEndsTheProcessAndTheArgumentsAndEnvironmentArriveAsGiven(t *testing.T) {
	s := sh(`exec sleep 30`)
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := Run(ctx, s); !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 10*time.Second {
		t.Errorf("err=%v after %v", err, time.Since(start))
	}
	e := sh(`printf '%s|%s|%s' "$BASE_ONE" "$EXTRA_ONE" "$1"`)
	e.Args = append(e.Args, "arg0", `$(echo no shell) "q"`)
	e.Base, e.Env = []string{"BASE_ONE=b", "EXTRA_ONE=base"}, []string{"EXTRA_ONE=extra"}
	out, err := Run(t.Context(), e)
	if err != nil || string(out) != `b|extra|$(echo no shell) "q"` {
		t.Errorf("out=%q err=%v", out, err)
	}
	if _, err := Run(t.Context(), Spec{Bin: "/no/such/program", MaxOutput: 10}); err == nil || strings.Contains(err.Error(), "exited") {
		t.Errorf("a program that cannot start: %v", err)
	}
}
