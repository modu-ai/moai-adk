package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
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

	t.Run("relaunch supervising loop starts one session per card", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
		// SPEC-TODO-CLASSIFY-DISPATCH-001: this scenario pins the LOOP's
		// continuation mechanics, and its stub children exit WITHOUT working
		// their card (the crashed-lane shape, whose recovery is lease
		// expiry). Unclassified cards read serial by default, so the
		// serial-vs-serial gate would stop the loop after one card — the
		// mode-neutral intent of this scenario maps to parallelizable.
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
		wantLabel := sdNextFreeLaneLabel(t, root)
		t.Chdir(root)
		t.Setenv(config.EnvClaudeProjectDir, root)
		sdScrubLauncherEnv(t)

		var got []sdCardLaunchCapture
		prevLook := claudeLookPath
		claudeLookPath = func(string) (string, error) { return "/sentinel/claude", nil }
		prevLaunch := factoryLaneCardLaunchFn
		factoryLaneCardLaunchFn = func(c *exec.Cmd) error {
			got = append(got, sdCardLaunchCapture{
				argv: append([]string(nil), c.Args...),
				dir:  c.Dir,
				env:  sdEnvOf(t, c.Env),
			})
			return nil
		}
		prevUnified := unifiedLaunchFunc
		unifiedLaunchFunc = func(string, string, []string) error {
			t.Error("the exec launch ran; the relaunch loop must start the children itself")
			return nil
		}
		prevRoot := findProjectRootFn
		findProjectRootFn = func() (string, error) { return root, nil }
		prevDeps := deps
		deps = nil
		t.Cleanup(func() {
			claudeLookPath, factoryLaneCardLaunchFn, unifiedLaunchFunc = prevLook, prevLaunch, prevUnified
			findProjectRootFn, deps = prevRoot, prevDeps
		})

		if err := sdCCEntry([]string{"-l", "--clear-policy", "relaunch"}); err != nil {
			t.Fatalf("cc lane relaunch: %v", err)
		}

		if len(got) != 2 {
			t.Fatalf("the supervising loop started %d sessions, want 2 (one per operator-picked card)", len(got))
		}
		for i, cardID := range []string{"t1", "t2"} {
			card := fcCard(t, root, cardID)
			rec := got[i]
			if card.WorktreePath == "" {
				t.Fatalf("%s recorded no worktree", cardID)
			}
			if rec.dir != card.WorktreePath {
				t.Errorf("session %d: child working directory = %q, want card %s's worktree %q", i, rec.dir, cardID, card.WorktreePath)
			}
			if rec.env[config.EnvFactoryCard] != cardID {
				t.Errorf("session %d: child env %s = %q, want card %s's id", i, config.EnvFactoryCard, rec.env[config.EnvFactoryCard], cardID)
			}
			if rec.env[config.EnvFactoryClearPolicy] != config.FactoryClearPolicyRelaunch {
				t.Errorf("session %d: child env %s = %q, want %q", i, config.EnvFactoryClearPolicy, rec.env[config.EnvFactoryClearPolicy], config.FactoryClearPolicyRelaunch)
			}
			if rec.env[config.EnvFactoryRole] != config.FactoryRoleLane {
				t.Errorf("session %d: child env %s = %q, want %q", i, config.EnvFactoryRole, rec.env[config.EnvFactoryRole], config.FactoryRoleLane)
			}
			if rec.env[config.EnvMoaiFactoryWorker] != wantLabel {
				t.Errorf("session %d: child env %s = %q, want the lane label %q", i, config.EnvMoaiFactoryWorker, rec.env[config.EnvMoaiFactoryWorker], wantLabel)
			}
			if !containsPair(rec.argv, nameFlagLong, wantLabel) {
				t.Errorf("session %d: argv %v does not carry --name with the lane label", i, rec.argv)
			}
		}
	})

	t.Run("complete prints the policy line end-to-end", func(t *testing.T) {
		sdClearLaneEnv(t)
		root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		if _, err := factory.RunRemeasure(root, cards[0].wt, "develop", "true"); err != nil {
			t.Fatalf("place candidate re-measure: %v", err)
		}
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

	t.Run("relaunch without the claude binary is refused", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked)
		sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
		t.Chdir(root)
		t.Setenv(config.EnvClaudeProjectDir, root)
		sdScrubLauncherEnv(t)
		prevLook := claudeLookPath
		claudeLookPath = func(string) (string, error) { return "", errors.New("not installed") }
		prevLaunch := factoryLaneCardLaunchFn
		factoryLaneCardLaunchFn = func(*exec.Cmd) error {
			t.Error("a session started without the binary")
			return nil
		}
		prevRoot := findProjectRootFn
		findProjectRootFn = func() (string, error) { return root, nil }
		prevDeps := deps
		deps = nil
		t.Cleanup(func() {
			claudeLookPath, factoryLaneCardLaunchFn = prevLook, prevLaunch
			findProjectRootFn, deps = prevRoot, prevDeps
		})
		if err := sdCCEntry([]string{"-l", "--clear-policy", "relaunch"}); err == nil {
			t.Fatal("relaunch without the claude binary succeeded, want a refusal")
		}
	})

	t.Run("relaunch continues after a failed child session", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
		// SPEC-TODO-CLASSIFY-DISPATCH-001: same mode-neutral mapping as the
		// per-card subtest above — the subject is the loop's continuation
		// after a failed child, not exclusivity; the failed card stays leased
		// until expiry either way.
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
		t.Chdir(root)
		t.Setenv(config.EnvClaudeProjectDir, root)
		sdScrubLauncherEnv(t)
		var started int
		prevLook := claudeLookPath
		claudeLookPath = func(string) (string, error) { return "/sentinel/claude", nil }
		prevLaunch := factoryLaneCardLaunchFn
		factoryLaneCardLaunchFn = func(*exec.Cmd) error {
			started++
			if started == 1 {
				return errors.New("child exited non-zero")
			}
			return nil
		}
		prevRoot := findProjectRootFn
		findProjectRootFn = func() (string, error) { return root, nil }
		prevDeps := deps
		deps = nil
		t.Cleanup(func() {
			claudeLookPath, factoryLaneCardLaunchFn = prevLook, prevLaunch
			findProjectRootFn, deps = prevRoot, prevDeps
		})
		if err := sdCCEntry([]string{"-l", "--clear-policy", "relaunch"}); err != nil {
			t.Fatalf("relaunch loop: %v", err)
		}
		if started != 2 {
			t.Fatalf("the loop started %d sessions after one failed child, want 2 (the failure continues the loop)", started)
		}
	})

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
		"factory_lane_relaunch.go":      sdReadSource(t, filepath.Join(cliDir, "factory_lane_relaunch.go")),
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
