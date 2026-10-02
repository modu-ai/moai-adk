package cli

// managed_failure_paths_test.go — SPEC-FACTORY-MANAGED-SESSION-001 M5
// failure-path coverage for the managed layer: the refusals and teardown
// branches a launch hits when its inputs or its backend misbehave.

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

// managedClosedPipe fails every write, standing in for a closed child stdin or a
// closed operator stdout.
type managedClosedPipe struct{}

func (managedClosedPipe) Write([]byte) (int, error) { return 0, errors.New("pipe closed") }

func TestManagedStreamSessionRefusals(t *testing.T) {
	t.Run("operator takeover of the stream flags is refused at construction", func(t *testing.T) {
		for _, flag := range []string{"-p", "--print", "--input-format", "--output-format=text"} {
			if _, err := newManagedStreamSession(BackendClaude, "claude", []string{"claude", flag}, nil); err == nil {
				t.Errorf("operator flag %q was accepted", flag)
			}
		}
	})
	t.Run("a turn before Start fails instead of writing to a dead child", func(t *testing.T) {
		s, err := newManagedStreamSession(BackendClaude, "claude", []string{"claude"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.DeliverTurn("hello"); err == nil || !strings.Contains(err.Error(), "not started") {
			t.Fatalf("DeliverTurn before Start = %v, want the not-started refusal", err)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("Close of a never-started session = %v, want nil", err)
		}
	})
}

func TestManagedStreamTurnPumpFailures(t *testing.T) {
	t.Run("a failed prompt write ends the turn with the write error", func(t *testing.T) {
		err := pumpManagedStreamTurn(strings.NewReader(`{"type":"result","is_error":false}`+"\n"), io.Discard, managedClosedPipe{}, "p")
		if err == nil || !strings.Contains(err.Error(), "pipe closed") {
			t.Fatalf("err = %v, want the write failure", err)
		}
	})
	t.Run("assistant text that cannot reach the operator fails the turn", func(t *testing.T) {
		out := `{"type":"assistant","message":{"content":[{"type":"text","text":"hi"}]}}` + "\n"
		if err := pumpManagedStreamTurn(strings.NewReader(out), managedClosedPipe{}, io.Discard, "p"); err == nil {
			t.Fatal("stdout write failure was swallowed")
		}
	})
	t.Run("an error result is surfaced with its message", func(t *testing.T) {
		out := `{"type":"result","is_error":true,"result":"model overloaded"}` + "\n"
		err := pumpManagedStreamTurn(strings.NewReader(out), io.Discard, io.Discard, "p")
		if err == nil || !strings.Contains(err.Error(), "model overloaded") {
			t.Fatalf("err = %v, want the result message", err)
		}
	})
	t.Run("an oversize line is a scan error, not a silent end", func(t *testing.T) {
		huge := strings.Repeat("x", managedStreamMaxLineBytes+1)
		err := pumpManagedStreamTurn(strings.NewReader(huge), io.Discard, io.Discard, "p")
		if err == nil || errors.Is(err, errManagedStreamClosed) {
			t.Fatalf("err = %v, want a scanner error", err)
		}
	})
}

func TestManagedDriverFailureBranches(t *testing.T) {
	noClaim := func() ([]factorymsg.Claim, error) { return nil, nil }
	prompt := func([]factorymsg.Claim) string { return "inbox" }

	t.Run("a priming turn that fails ends the session before any poll", func(t *testing.T) {
		sess := &fakeManagedSession{failOnPrompt: managedPrimingPrompt}
		err := driveManagedFactorySession(sess, strings.NewReader(""), make(chan time.Time), noClaim, prompt)
		if err == nil {
			t.Fatal("priming failure was swallowed")
		}
	})
	t.Run("operator stdin ending leaves the broker poll running", func(t *testing.T) {
		sess := &fakeManagedSession{failOnPrompt: "inbox"}
		idle := make(chan time.Time, 4)
		idle <- time.Now()
		idle <- time.Now()
		var calls int
		claim := func() ([]factorymsg.Claim, error) {
			calls++
			if calls < 2 {
				return nil, nil
			}
			return []factorymsg.Claim{{ClaimToken: "tok"}}, nil
		}
		// EOF stdin: the driver must keep polling and still deliver the
		// inbox turn, which then fails and ends the run.
		err := driveManagedFactorySession(sess, strings.NewReader(""), idle, claim, prompt)
		if err == nil || !strings.Contains(err.Error(), "stream died") {
			t.Fatalf("err = %v, want the inbox delivery failure after stdin EOF", err)
		}
		if calls < 2 {
			t.Fatalf("claim called %d times, want polling to continue past stdin EOF", calls)
		}
	})
}

func TestManagedLaunchRefusesWithoutFactoryRun(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, glm := range []bool{false, true} {
		if err := managedFactoryLaunch(glm, "claude", []string{"claude"}, nil); err == nil {
			t.Errorf("managed launch (glm=%v) with no factory run succeeded", glm)
		}
	}
	// An operator flag that takes over the stream is refused by the owner
	// before any child exists.
	if err := runManagedFactoryStreamSession(BackendClaude, "claude", []string{"claude", "--print"}, nil, strings.NewReader("")); err == nil {
		t.Error("owner accepted an operator --print")
	}
}

func TestManagedCodexAppReadyGuards(t *testing.T) {
	if err := managedCodexAppReady(context.Background(), "ws://10.1.2.3:9000"); !errors.Is(err, errManagedCodexNonLoopback) {
		t.Fatalf("readiness probe to a non-loopback target = %v, want the loopback refusal", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Port 1 on loopback has no listener: a canceled context must end the
	// wait instead of polling forever.
	if err := managedCodexAppReady(ctx, "ws://127.0.0.1:1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled readiness wait = %v, want context.Canceled", err)
	}
}
