package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/statusline"
)

// factory_m6_test.go covers the lane clear policies (SPEC-FACTORY-SELF-
// DISPATCH-001 REQ-SD-020, AC-SD-020): the line the `complete` output prints
// per policy, the context-usage record comparison for clear-when-full, and
// the relaunch supervising loop that starts one session per card.

// sdWriteContextRecord injects a context-usage telemetry record for one
// session under root — the record the clear-when-full comparison reads.
func sdWriteContextRecord(t *testing.T, root, sessionID string, window int, rawPct float64) {
	t.Helper()
	path := statusline.SessionTelemetryPath(filepath.Join(root, ".moai", "state"), sessionID)
	if path == "" {
		t.Fatalf("session key %q refused", sessionID)
	}
	rec := statusline.SessionTelemetryRecord{
		SchemaVersion:     2,
		SessionID:         sessionID,
		WriterPID:         os.Getpid(),
		CapturedAt:        time.Now().Format(time.RFC3339Nano),
		ContextWindowSize: window,
		TokensUsed:        int(float64(window) * rawPct / 100),
		RawPct:            rawPct,
		Stage:             "none",
		Band:              "standard",
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write record: %v", err)
	}
}

// sdClearPolicyOutput renders the policy line the complete output appends.
func sdClearPolicyOutput(t *testing.T, root string) string {
	t.Helper()
	var buf bytes.Buffer
	factoryPrintClearPolicyLine(&buf, root)
	return buf.String()
}

// sdWantOneClearLine asserts the output is exactly one line asking the
// operator to /clear, naming the policy.
func sdWantOneClearLine(t *testing.T, out, policy string) {
	t.Helper()
	trimmed := strings.TrimRight(out, "\n")
	if trimmed == "" {
		t.Fatalf("policy %s: no line printed, want one /clear request", policy)
	}
	if strings.Contains(trimmed, "\n") {
		t.Fatalf("policy %s: printed more than one line:\n%s", policy, out)
	}
	if strings.Count(out, "/clear") != 1 {
		t.Fatalf("policy %s: output %q does not carry exactly one /clear request", policy, out)
	}
	if !strings.Contains(out, policy) {
		t.Fatalf("policy %s: output %q does not name the policy", policy, out)
	}
}

// AC-SD-020 — each policy with an injected context-usage record; no policy
// given equals clear-each.
func TestSD_AC020_ClearPolicies(t *testing.T) {
	const sessionID = "sess-m6-clear"

	t.Run("clear-each prints exactly one /clear request", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeCodeSessionID, sessionID)
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyEach)
		sdWantOneClearLine(t, sdClearPolicyOutput(t, root), config.FactoryClearPolicyEach)
	})

	t.Run("absent policy equals clear-each", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeCodeSessionID, sessionID)
		t.Setenv(config.EnvFactoryClearPolicy, "")
		out := sdClearPolicyOutput(t, root)
		if out == "" || strings.Count(out, "/clear") != 1 {
			t.Fatalf("absent policy output = %q, want the clear-each /clear request", out)
		}
	})

	t.Run("clear-when-full below threshold prints none", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeCodeSessionID, sessionID)
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyWhenFull)
		sdWriteContextRecord(t, root, sessionID, 200_000, 10.0)
		if out := sdClearPolicyOutput(t, root); out != "" {
			t.Fatalf("clear-when-full below threshold printed %q, want nothing", out)
		}
	})

	t.Run("clear-when-full at threshold prints one", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeCodeSessionID, sessionID)
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyWhenFull)
		sdWriteContextRecord(t, root, sessionID, 200_000, 90.0)
		sdWantOneClearLine(t, sdClearPolicyOutput(t, root), config.FactoryClearPolicyWhenFull)
	})

	t.Run("clear-when-full large window at threshold prints one", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeCodeSessionID, sessionID)
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyWhenFull)
		sdWriteContextRecord(t, root, sessionID, 1_000_000, 50.0)
		sdWantOneClearLine(t, sdClearPolicyOutput(t, root), config.FactoryClearPolicyWhenFull)
	})

	t.Run("clear-when-full missing record prints none", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeCodeSessionID, sessionID)
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyWhenFull)
		if out := sdClearPolicyOutput(t, root); out != "" {
			t.Fatalf("clear-when-full with no record printed %q, want nothing (a missing record reads as below)", out)
		}
	})

	t.Run("clear-when-full refused session key prints none", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeCodeSessionID, "../escape")
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyWhenFull)
		if out := sdClearPolicyOutput(t, root); out != "" {
			t.Fatalf("refused session key printed %q, want nothing", out)
		}
	})

	t.Run("relaunch prints one end-session request", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeCodeSessionID, sessionID)
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyRelaunch)
		out := sdClearPolicyOutput(t, root)
		trimmed := strings.TrimRight(out, "\n")
		if trimmed == "" || strings.Contains(trimmed, "\n") {
			t.Fatalf("relaunch output = %q, want exactly one end-session request line", out)
		}
		if !strings.Contains(out, "end this session") {
			t.Fatalf("relaunch output %q does not ask to end the session", out)
		}
		if strings.Contains(out, "/clear") {
			t.Fatalf("relaunch output %q asks for /clear; it must ask to end the session", out)
		}
	})

	t.Run("relaunch degrades to the one-shot lane session", func(t *testing.T) {
		// Card t1554: the supervising lease loop is removed — card
		// consumption is the unified `moai todo --auto` engine's alone. The
		// flag stays accepted: one supersession note on the error stream, one
		// lane session started, and the queue untouched (no lease, no card
		// rows — the session consumes the queue itself).
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
		sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
		t.Chdir(root)
		t.Setenv(config.EnvClaudeProjectDir, root)
		sdScrubLauncherEnv(t)

		var launches int
		prevLook := claudeLookPath
		claudeLookPath = func(string) (string, error) { return "/sentinel/claude", nil }
		prevUnified := unifiedLaunchFunc
		unifiedLaunchFunc = func(string, string, []string) error {
			launches++
			return nil
		}
		prevRoot := findProjectRootFn
		findProjectRootFn = func() (string, error) { return root, nil }
		prevDeps := deps
		deps = nil
		t.Cleanup(func() {
			claudeLookPath, unifiedLaunchFunc = prevLook, prevUnified
			findProjectRootFn, deps = prevRoot, prevDeps
		})

		if err := sdCCEntry([]string{"-l", "--clear-policy", "relaunch"}); err != nil {
			t.Fatalf("cc lane relaunch: %v", err)
		}
		if launches != 1 {
			t.Fatalf("the degraded relaunch started %d sessions, want 1 (the one-shot lane session)", launches)
		}
		// The queue is untouched: the launcher leases nothing at boot.
		for _, cardID := range []string{"t1", "t2"} {
			if fcHasCard(t, root, cardID) {
				t.Errorf("%s gained a factory record row; the launcher must not lease", cardID)
			}
		}
		if s := nmQueueState(t, store, "t1"); s != factory.BacklogStatePicked {
			t.Errorf("t1 queue state = %s, want picked (the seeded operator pick stands)", s)
		}
	})

	t.Run("complete prints the policy line end-to-end", func(t *testing.T) {
		sdClearLaneEnv(t)
		root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", factory.BranchSourceConfig, integWT, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyEach)
		t.Chdir(root)
		out, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err != nil {
			t.Fatalf("cli complete: %v", err)
		}
		if n := strings.Count(out, "/clear"); n != 1 {
			t.Fatalf("complete output carries %d /clear requests, want 1:\n%s", n, out)
		}
		if !strings.Contains(out, config.FactoryClearPolicyEach) {
			t.Errorf("complete output does not name the policy:\n%s", out)
		}
	})

	// (The loop's own "relaunch without the claude binary is refused" pin is
	// gone with the loop: the binary precondition lived in the removed
	// supervisor. The one-shot launch's resolution is the plain lane
	// launch's, not the relaunch policy's.)

	t.Run("clear-policy outside a lane entry is refused", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvClaudeProjectDir, root)
		err := sdCCEntry([]string{"--clear-policy", config.FactoryClearPolicyRelaunch})
		if err == nil {
			t.Fatal("--clear-policy without a factory lane entry succeeded, want a refusal")
		}
		if strings.Contains(err.Error(), "\n") {
			t.Errorf("refusal is more than one line:\n%s", err.Error())
		}
	})

	t.Run("clear-policy on the leader shape is refused", func(t *testing.T) {
		root, _ := fcFixture(t)
		t.Setenv(config.EnvClaudeProjectDir, root)
		prevRoot := findProjectRootFn
		findProjectRootFn = func() (string, error) { return root, nil }
		prevDeps := deps
		deps = nil
		t.Cleanup(func() { findProjectRootFn, deps = prevRoot, prevDeps })
		err := sdCCEntry([]string{"-f", "--clear-policy", config.FactoryClearPolicyEach})
		if err == nil {
			t.Fatal("--clear-policy with the leader shape succeeded, want a refusal")
		}
	})

	t.Run("invalid policy value is refused", func(t *testing.T) {
		root, _ := fcFixture(t)
		t.Setenv(config.EnvClaudeProjectDir, root)
		prevRoot := findProjectRootFn
		findProjectRootFn = func() (string, error) { return root, nil }
		prevDeps := deps
		deps = nil
		t.Cleanup(func() { findProjectRootFn, deps = prevRoot, prevDeps })
		err := sdCCEntry([]string{"-l", "--clear-policy", "never"})
		if err == nil {
			t.Fatal("invalid clear-policy value succeeded, want a refusal")
		}
		for _, want := range []string{
			config.FactoryClearPolicyEach, config.FactoryClearPolicyWhenFull, config.FactoryClearPolicyRelaunch,
		} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("refusal %q does not name the accepted value %q", err.Error(), want)
			}
		}
	})
}

// AC-SD-019 source hygiene — the new carrier name exists only as the
// internal/config constant in the M6 production files (the M4 literal scan's
// M6 counterpart).
func TestSD_M6_ClearPolicyCarrierConstantOnly(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not name this file")
	}
	cliDir := filepath.Dir(thisFile)
	hookDir := filepath.Join(cliDir, "..", "hook")
	scanned := map[string]string{
		"factory.go":                    sdReadSource(t, filepath.Join(cliDir, "factory.go")),
		"factory_card.go":               sdReadSource(t, filepath.Join(cliDir, "factory_card.go")),
		"cc.go":                         sdReadSource(t, filepath.Join(cliDir, "cc.go")),
		"glm.go":                        sdReadSource(t, filepath.Join(cliDir, "glm.go")),
		"codex_launcher.go":             sdReadSource(t, filepath.Join(cliDir, "codex_launcher.go")),
		"session_start_factory.go":      sdReadSource(t, filepath.Join(hookDir, "session_start_factory.go")),
		"session_start_factory_i18n.go": sdReadSource(t, filepath.Join(hookDir, "session_start_factory_i18n.go")),
	}
	for name, body := range scanned {
		for _, literal := range []string{`"MOAI_FACTORY_CLEAR_POLICY"`} {
			if strings.Contains(body, literal) {
				t.Errorf("%s carries the raw literal %s — the env name exists only as the internal/config constant", name, literal)
			}
		}
	}
}
