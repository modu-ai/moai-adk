package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestHookWrapper_LargeStdin_DoesNotExceedTimeout verifies that hook wrappers
// process 50KB+ JSON payloads without approaching the 5-second timeout.
// This is the regression test for the 22 PreToolUse timeout cancellations
// observed in the diagnostic session (SPEC-V3R4-HOOK-HARDEN-001 §0).
func TestHookWrapper_LargeStdin_DoesNotExceedTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		SkipIfNoBash(t)
	}

	wrapperDir := filepath.Join("..", "template", "templates", ".claude", "hooks", "moai")
	preToolWrapper := filepath.Join(wrapperDir, "handle-pre-tool.sh.tmpl")

	if _, err := os.Stat(preToolWrapper); err != nil {
		t.Skipf("template wrapper not found: %s", preToolWrapper)
	}

	tmpDir := t.TempDir()
	wrapperCopy := filepath.Join(tmpDir, "handle-pre-tool.sh")
	data, err := os.ReadFile(preToolWrapper)
	if err != nil {
		t.Fatalf("read wrapper: %v", err)
	}
	if err := os.WriteFile(wrapperCopy, data, 0755); err != nil {
		t.Fatalf("write wrapper copy: %v", err)
	}

	payload := buildLargePayload(t, 50000)

	stderrLog := filepath.Join(tmpDir, "hook-stderr.log")

	// Cold-start: first execution to warm caches
	runWrapperOnce(t, wrapperCopy, payload, stderrLog)
	os.Remove(stderrLog)

	// Measured execution
	start := time.Now()
	runWrapperOnce(t, wrapperCopy, payload, stderrLog)
	elapsed := time.Since(start)

	if elapsed >= 1*time.Second {
		t.Errorf("wrapper execution took %v, want < 1s (regression: timeout risk)", elapsed)
	}

	t.Logf("wrapper execution time: %v", elapsed)
}

// TestHookWrapper_OptOutStderrLog verifies that setting
// MOAI_HOOK_STDERR_LOG=/dev/null suppresses log file creation.
func TestHookWrapper_OptOutStderrLog(t *testing.T) {
	if runtime.GOOS == "windows" {
		SkipIfNoBash(t)
	}

	wrapperDir := filepath.Join("..", "template", "templates", ".claude", "hooks", "moai")
	sessionStartWrapper := filepath.Join(wrapperDir, "handle-session-start.sh.tmpl")

	if _, err := os.Stat(sessionStartWrapper); err != nil {
		t.Skipf("template wrapper not found: %s", sessionStartWrapper)
	}

	tmpDir := t.TempDir()
	wrapperCopy := filepath.Join(tmpDir, "handle-session-start.sh")
	data, err := os.ReadFile(sessionStartWrapper)
	if err != nil {
		t.Fatalf("read wrapper: %v", err)
	}
	if err := os.WriteFile(wrapperCopy, data, 0755); err != nil {
		t.Fatalf("write wrapper copy: %v", err)
	}

	cmd := exec.Command("bash", wrapperCopy)
	cmd.Stdin = strings.NewReader(`{"session_id":"test"}`)
	cmd.Env = append(os.Environ(),
		"MOAI_HOOK_STDERR_LOG=/dev/null",
		"HOME="+tmpDir,
	)
	cmd.CombinedOutput()

	logFile := filepath.Join(tmpDir, ".moai", "logs", "hook-stderr.log")
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Error("log file was created despite MOAI_HOOK_STDERR_LOG=/dev/null")
	}
}

// TestHookWrapper_NoHardcodedUserPath is a sentinel regression test
// ensuring no wrapper template contains hardcoded user paths.
func TestHookWrapper_NoHardcodedUserPath(t *testing.T) {
	templateDir := filepath.Join("..", "template", "templates", ".claude", "hooks", "moai")
	templates, err := filepath.Glob(filepath.Join(templateDir, "handle-*.sh.tmpl"))
	if err != nil {
		t.Fatalf("glob templates: %v", err)
	}
	if len(templates) == 0 {
		t.Skip("no template wrappers found")
	}

	for _, tmpl := range templates {
		content, err := os.ReadFile(tmpl)
		if err != nil {
			t.Fatalf("read %s: %v", tmpl, err)
		}
		s := string(content)

		if strings.Contains(s, "/Users/") {
			t.Errorf("%s: hardcoded /Users/ path detected", filepath.Base(tmpl))
		}
		// Check /home/<user> but allow $HOME
		for _, line := range strings.Split(s, "\n") {
			if strings.Contains(line, "/home/") && !strings.Contains(line, "$HOME") {
				t.Errorf("%s: hardcoded /home/ path (not $HOME): %s", filepath.Base(tmpl), strings.TrimSpace(line))
			}
		}
	}
}

// buildLargePayload creates a JSON payload of approximately targetBytes.
func buildLargePayload(t *testing.T, targetBytes int) []byte {
	t.Helper()
	fillSize := targetBytes - len(`{"tool_input":{"command":""}}`)
	if fillSize < 0 {
		fillSize = 0
	}
	payload := fmt.Sprintf(`{"tool_input":{"command":"%s"}}`, strings.Repeat("x", fillSize))
	return []byte(payload)
}

// runWrapperOnce executes a wrapper script with the given payload as stdin.
func runWrapperOnce(t *testing.T, wrapper string, payload []byte, stderrLog string) {
	t.Helper()
	cmd := exec.Command("bash", wrapper)
	cmd.Stdin = strings.NewReader(string(payload))
	cmd.Env = append(os.Environ(),
		"MOAI_HOOK_STDERR_LOG="+stderrLog,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("wrapper output: %s", out)
		// moai binary not found is expected in test environments
		if !strings.Contains(string(out), "moai") && !strings.Contains(err.Error(), "exit") {
			t.Fatalf("wrapper execution failed: %v", err)
		}
	}
}

// SkipIfNoBash skips the test if bash is not available.
func SkipIfNoBash(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
}
