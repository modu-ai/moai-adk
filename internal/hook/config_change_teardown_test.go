package hook

// config_change_teardown_test.go — coverage for the one-shot-process teardown
// contract on the ConfigChange path: the handler's background work is joined
// before the process exits, and its verdict is written somewhere that survives
// the exit.
//
// The measured defect both halves refute: across 5 real `moai hook
// config-change` CLI processes fed an invalid YAML file, the async validation
// was reached 0 times, and even had it been reached its slog warning would
// have gone to io.Discard (the unconditional `moai hook` logging carve-out in
// internal/cli/logging.go). Asserting only that validation eventually runs
// would leave the second half unproven.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/testutil"
)

// readConfigChangeAudit returns the audit log's lines for a project dir, or an
// empty slice when the file was never created.
func readConfigChangeAudit(t *testing.T, projectDir string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectDir, configChangeAuditRelPath))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// writeConfigFixture writes a config file and returns a HookInput pointing at
// it, anchored on projectDir so the audit log lands inside the test's tempdir.
func writeConfigFixture(t *testing.T, projectDir, name, content string) *HookInput {
	t.Helper()
	path := filepath.Join(projectDir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return &HookInput{
		SessionID:      "sess-teardown",
		HookEventName:  "ConfigChange",
		ConfigFilePath: path,
		ConfigSource:   "project",
		CWD:            projectDir,
	}
}

// TestRegistryShutdownJoinsConfigChangeAsyncWork is the direct refutation of
// the measured defect: after Dispatch returns, teardown alone — with no test-
// side WaitGroup join — must be enough for the rejection verdict to exist on
// disk.
//
// The test deliberately does NOT call testutil.WaitForAsync. Waiting on the
// handler's WaitGroup here would reproduce the control condition rather than
// the production one: the control already passed before this change, while the
// real CLI process, which has no such join, produced nothing.
func TestRegistryShutdownJoinsConfigChangeAsyncWork(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	input := writeConfigFixture(t, projectDir, "broken.yaml", "a: [1, 2\n  bad: : :\n")

	reg := NewRegistry(config.NewConfigManager())
	reg.Register(NewConfigChangeHandler())

	if _, err := reg.Dispatch(context.Background(), EventConfigChange, input); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	// Stand-in for process exit: the one-shot CLI defers exactly this call
	// (internal/cli/hook.go).
	reg.Shutdown()

	lines := readConfigChangeAudit(t, projectDir)
	if len(lines) != 1 {
		t.Fatalf("audit lines after teardown = %d, want 1; lines=%v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "result="+configChangeResultRejected) {
		t.Errorf("audit line = %q, want result=%s", lines[0], configChangeResultRejected)
	}
	if !strings.Contains(lines[0], "invalid YAML") {
		t.Errorf("audit line = %q, want the validation error in the detail field", lines[0])
	}
}

// TestConfigChangeAuditRecordsBothOutcomes asserts the audit trail separates
// "the config was fine" from "the check never ran". A trail that recorded only
// rejections would leave those two indistinguishable, which is the ambiguity
// the file exists to remove — and the card's completion condition requires the
// control group to be observable too, not just the failure group.
func TestConfigChangeAuditRecordsBothOutcomes(t *testing.T) {
	t.Parallel()

	t.Run("valid config records reloaded", func(t *testing.T) {
		t.Parallel()
		projectDir := t.TempDir()
		input := writeConfigFixture(t, projectDir, "good.yaml", "development_mode: tdd\ncoverage_target: 85\n")

		h := NewConfigChangeHandler().(*configChangeHandler)
		if _, err := h.Handle(context.Background(), input); err != nil {
			t.Fatalf("handle: %v", err)
		}
		testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)

		lines := readConfigChangeAudit(t, projectDir)
		if len(lines) != 1 {
			t.Fatalf("audit lines = %d, want 1; lines=%v", len(lines), lines)
		}
		if !strings.Contains(lines[0], "result="+configChangeResultReloaded) {
			t.Errorf("audit line = %q, want result=%s", lines[0], configChangeResultReloaded)
		}
	})

	t.Run("invalid config records rejected", func(t *testing.T) {
		t.Parallel()
		projectDir := t.TempDir()
		input := writeConfigFixture(t, projectDir, "bad.yaml", "a: [1, 2\n  bad: : :\n")

		h := NewConfigChangeHandler().(*configChangeHandler)
		if _, err := h.Handle(context.Background(), input); err != nil {
			t.Fatalf("handle: %v", err)
		}
		testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)

		lines := readConfigChangeAudit(t, projectDir)
		if len(lines) != 1 {
			t.Fatalf("audit lines = %d, want 1; lines=%v", len(lines), lines)
		}
		if !strings.Contains(lines[0], "result="+configChangeResultRejected) {
			t.Errorf("audit line = %q, want result=%s", lines[0], configChangeResultRejected)
		}
	})
}

// TestConfigChangeJoinAsyncBoundsTheWait asserts the join is bounded rather
// than open-ended: a side effect that never finishes must not hang the exiting
// process, it must be abandoned with a diagnostic.
func TestConfigChangeJoinAsyncBoundsTheWait(t *testing.T) {
	t.Parallel()

	h := NewConfigChangeHandler().(*configChangeHandler)

	// Simulate a side effect still in flight. The release is deferred so the
	// abandoned helper goroutine inside joinAsync cannot outlive the test.
	var release sync.WaitGroup
	release.Add(1)
	h.wg.Add(1)
	go func() {
		release.Wait()
		h.wg.Done()
	}()
	defer release.Done()

	start := time.Now()
	err := h.joinAsync(20 * time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("joinAsync returned nil for work that never finished, want a timeout error")
	}
	if elapsed > time.Second {
		t.Errorf("joinAsync blocked for %v, want it bounded near the 20ms budget", elapsed)
	}
}

// TestConfigChangeJoinAsyncReturnsOnCompletion is the positive control for the
// test above: the same method must return nil, and promptly, once the work is
// actually done. Without it a joinAsync that always reported a timeout would
// still pass.
func TestConfigChangeJoinAsyncReturnsOnCompletion(t *testing.T) {
	t.Parallel()

	h := NewConfigChangeHandler().(*configChangeHandler)
	if err := h.joinAsync(2 * time.Second); err != nil {
		t.Fatalf("joinAsync with no in-flight work = %v, want nil", err)
	}
}

// TestRegistryShutdownTolerantOfPlainHandlers asserts teardown stays safe for
// handlers that start no background work and is repeatable — Shutdown is
// deferred by the CLI and documented as idempotent, and adding a second async
// axis to it must not narrow either property.
func TestRegistryShutdownTolerantOfPlainHandlers(t *testing.T) {
	t.Parallel()

	reg := NewRegistry(config.NewConfigManager())
	reg.Register(&lastTraceHandler{event: EventConfigChange})

	reg.Shutdown()
	reg.Shutdown()
}
