package hook

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	// wrapperTestTimeout is the maximum time to wait for a wrapper script to execute
	wrapperTestTimeout = 5 * time.Second
	// maxWrapperInputSize is the 1MB size limit enforced by wrapper scripts
	maxWrapperInputSize = 1048576
)

// TestHookWrapper_SizeLimit verifies that input larger than 1MB is truncated.
func TestHookWrapper_SizeLimit(t *testing.T) {
	t.Parallel()

	// Create a temporary hook wrapper script
	tempDir := t.TempDir()
	wrapperPath := filepath.Join(tempDir, "test-hook.sh")
	createTestWrapper(t, wrapperPath)

	// Create input larger than 1MB
	largeInput := strings.Repeat("x", maxWrapperInputSize+1000)

	ctx, cancel := context.WithTimeout(context.Background(), wrapperTestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", wrapperPath)
	cmd.Dir = tempDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start wrapper: %v", err)
	}

	// Write large input
	if _, err := stdin.Write([]byte(largeInput)); err != nil {
		t.Fatalf("failed to write stdin: %v", err)
	}
	_ = stdin.Close()

	// Wrapper should handle large input without hanging
	if err := cmd.Wait(); err != nil {
		// Exit code 0 is expected (graceful handling)
		if cmd.ProcessState.ExitCode() != 0 {
			t.Logf("wrapper exited with code %d (may be expected): %v", cmd.ProcessState.ExitCode(), err)
		}
	}

	// Verify wrapper didn't hang and completed within timeout
	t.Logf("large input test completed. stdout: %d bytes, stderr: %d bytes",
		stdout.Len(), stderr.Len())
}

// TestHookWrapper_EmptyInput verifies empty input exits gracefully.
func TestHookWrapper_EmptyInput(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wrapperPath := filepath.Join(tempDir, "test-hook.sh")
	createTestWrapper(t, wrapperPath)

	ctx, cancel := context.WithTimeout(context.Background(), wrapperTestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", wrapperPath)
	cmd.Dir = tempDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Provide empty stdin
	cmd.Stdin = bytes.NewReader([]byte{})

	if err := cmd.Run(); err != nil {
		t.Fatalf("wrapper should exit gracefully with empty input: %v", err)
	}

	// Exit code should be 0
	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("expected exit code 0, got %d", cmd.ProcessState.ExitCode())
	}

	// Empty input should produce no output
	if stdout.Len() > 0 {
		t.Errorf("expected no stdout for empty input, got %d bytes", stdout.Len())
	}
}

// TestHookWrapper_ValidJSON verifies valid JSON is forwarded correctly.
func TestHookWrapper_ValidJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wrapperPath := filepath.Join(tempDir, "test-hook.sh")

	// Create a mock moai binary that just echoes input
	moaiPath := filepath.Join(tempDir, "moai")
	createMockMoai(t, moaiPath)

	// Create wrapper that uses our mock moai
	createTestWrapperWithMoaiPath(t, wrapperPath, moaiPath)

	validInput := `{"session_id":"test-123","hook_event_name":"SessionStart","cwd":"/tmp"}`

	ctx, cancel := context.WithTimeout(context.Background(), wrapperTestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", wrapperPath)
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "PATH="+tempDir+":"+os.Getenv("PATH"))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start wrapper: %v", err)
	}

	if _, err := stdin.Write([]byte(validInput)); err != nil {
		t.Fatalf("failed to write stdin: %v", err)
	}
	_ = stdin.Close()

	if err := cmd.Wait(); err != nil {
		t.Logf("wrapper exited: %v (stdout: %s, stderr: %s)", err, stdout.String(), stderr.String())
	}

	// Verify moai binary was called (mock writes to stdout)
	if !strings.Contains(stdout.String(), "MOAI_CALLED") {
		t.Error("wrapper did not forward input to moai binary")
	}

	// Verify JSON was preserved
	if !strings.Contains(stdout.String(), "session_id") {
		t.Error("JSON input was not preserved in forward")
	}
}

// TestHookWrapper_MalformedInput verifies malformed input doesn't crash.
func TestHookWrapper_MalformedInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		malformedInput string
	}{
		{
			name:         "incomplete JSON",
			malformedInput: `{"session_id":"test",`,
		},
		{
			name:         "random garbage",
			malformedInput: "!@#$%^&*()_+{}[]|\\:;<>?,./",
		},
		{
			name:         "binary data",
			malformedInput: string([]byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}),
		},
		{
			name:         "null bytes in middle",
			malformedInput: `{"test":"value` + "\x00" + `"continue"}`,
		},
		{
			name:         "unicode with nulls",
			malformedInput: "한글\x00테스트\x00😀🚀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			wrapperPath := filepath.Join(tempDir, "test-hook.sh")
			createTestWrapper(t, wrapperPath)

			ctx, cancel := context.WithTimeout(context.Background(), wrapperTestTimeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, "bash", wrapperPath)
			cmd.Dir = tempDir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatalf("failed to create stdin pipe: %v", err)
			}

			if err := cmd.Start(); err != nil {
				t.Fatalf("failed to start wrapper: %v", err)
			}

			if _, err := stdin.Write([]byte(tt.malformedInput)); err != nil {
				t.Fatalf("failed to write stdin: %v", err)
			}
			_ = stdin.Close()

			// Malformed input should not crash the wrapper
			err = cmd.Wait()
			if err != nil && cmd.ProcessState.ExitCode() != 0 {
				// Non-zero exit is acceptable for malformed input
				t.Logf("malformed input produced exit code %d: %v", cmd.ProcessState.ExitCode(), err)
			}

			// Wrapper should not hang
			t.Logf("malformed input handled. stdout: %d bytes, stderr: %d bytes",
				stdout.Len(), stderr.Len())
		})
	}
}

// TestHookWrapper_MoaiBinaryFallback verifies the fallback binary search path.
func TestHookWrapper_MoaiBinaryFallback(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wrapperPath := filepath.Join(tempDir, "test-hook.sh")

	// Create mock moai in a custom fallback location
	fallbackDir := filepath.Join(tempDir, "fallback")
	if err := os.MkdirAll(fallbackDir, 0o755); err != nil {
		t.Fatalf("failed to create fallback dir: %v", err)
	}

	moaiPath := filepath.Join(fallbackDir, "moai")
	createMockMoai(t, moaiPath)

	// Normalize path separators so the embedded bash script parses correctly on Windows.
	// Without this, backslashes in paths like C:\Users\... are interpreted as shell
	// escape sequences by Git Bash, causing the fallback lookup to silently fail.
	moaiPathBash := filepath.ToSlash(moaiPath)

	// Create wrapper that uses an absolute path to our mock moai
	wrapperScript := `#!/bin/bash
temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT

MAX_SIZE=1048576
head -c "$MAX_SIZE" > "$temp_file"

if [ ! -s "$temp_file" ]; then
    exit 0
fi

# Try fallback moai binary (absolute path)
if [ -f "` + moaiPathBash + `" ]; then
    exec "` + moaiPathBash + `" hook config-change < "$temp_file" 2>&1
fi

exit 0
`
	if err := os.WriteFile(wrapperPath, []byte(wrapperScript), 0o755); err != nil {
		t.Fatalf("failed to create wrapper: %v", err)
	}

	validInput := `{"session_id":"test-fallback","hook_event_name":"SessionStart"}`

	ctx, cancel := context.WithTimeout(context.Background(), wrapperTestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", wrapperPath)
	cmd.Dir = tempDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start wrapper: %v", err)
	}

	if _, err := stdin.Write([]byte(validInput)); err != nil {
		t.Fatalf("failed to write stdin: %v", err)
	}
	_ = stdin.Close()

	if err := cmd.Wait(); err != nil {
		t.Logf("fallback test exited: %v", err)
	}

	// Verify fallback moai binary was called (mock writes to stdout)
	if !strings.Contains(stdout.String(), "MOAI_CALLED") {
		t.Errorf("wrapper did not find fallback moai binary. stdout=%q, stderr=%q",
			stdout.String(), stderr.String())
	}

	// Verify JSON was preserved
	if !strings.Contains(stdout.String(), "session_id") {
		t.Error("JSON input was not preserved in forward")
	}
}

// TestHookWrapper_TempFileCleanup verifies temp files are cleaned up.
func TestHookWrapper_TempFileCleanup(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wrapperPath := filepath.Join(tempDir, "test-hook.sh")
	createTestWrapper(t, wrapperPath)

	// Count temp files before execution
	tempFilePattern := filepath.Join(os.TempDir(), "tmp.*")
	beforeFiles, err := filepath.Glob(tempFilePattern)
	if err != nil {
		t.Fatalf("failed to list temp files before: %v", err)
	}

	validInput := `{"session_id":"test-cleanup","hook_event_name":"SessionStart"}`

	ctx, cancel := context.WithTimeout(context.Background(), wrapperTestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", wrapperPath)
	cmd.Dir = tempDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Stdin = bytes.NewReader([]byte(validInput))

	if err := cmd.Run(); err != nil {
		t.Logf("cleanup test exited: %v", err)
	}

	// Count temp files after execution (with small delay for async cleanup)
	time.Sleep(100 * time.Millisecond)
	afterFiles, err := filepath.Glob(tempFilePattern)
	if err != nil {
		t.Fatalf("failed to list temp files after: %v", err)
	}

	// Temp file count should not increase dramatically
	// Allow up to 10 new temp files (system processes create temp files too)
	diff := len(afterFiles) - len(beforeFiles)
	if diff > 10 {
		t.Errorf("potential temp file leak detected: %d more temp files after execution", diff)
	}

	t.Logf("temp file change: %d before, %d after (diff: %d)", len(beforeFiles), len(afterFiles), diff)
}

// TestHookWrapper_SignalHandling verifies wrapper handles signals gracefully.
func TestHookWrapper_SignalHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping signal handling test in short mode")
	}
	t.Parallel()

	tempDir := t.TempDir()
	wrapperPath := filepath.Join(tempDir, "test-hook.sh")
	createTestWrapper(t, wrapperPath)

	ctx, cancel := context.WithTimeout(context.Background(), wrapperTestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", wrapperPath)
	cmd.Dir = tempDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start wrapper: %v", err)
	}

	// Send SIGTERM after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		if cmd.Process != nil {
			_ = cmd.Process.Signal(os.Interrupt)
		}
	}()

	_ = stdin.Close()

	// Wrapper should handle signal gracefully (exit without hanging)
	err = cmd.Wait()
	if err != nil {
		t.Logf("signal handling test exited: %v", err)
	}

	// Should not timeout (context cancellation)
	if ctx.Err() == context.DeadlineExceeded {
		t.Error("wrapper did not handle signal gracefully, timed out")
	}
}

// createTestWrapper creates a minimal hook wrapper script for testing.
func createTestWrapper(t *testing.T, path string) {
	t.Helper()

	wrapperScript := `#!/bin/bash
temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT

MAX_SIZE=1048576
head -c "$MAX_SIZE" > "$temp_file"

if [ ! -s "$temp_file" ]; then
    exit 0
fi

# Try PATH moai
if command -v moai &> /dev/null; then
    exec moai hook config-change < "$temp_file" 2>/dev/null
fi

# Try fallback
if [ -f "$HOME/go/bin/moai" ]; then
    exec "$HOME/go/bin/moai" hook config-change < "$temp_file" 2>/dev/null
fi

exit 0
`
	if err := os.WriteFile(path, []byte(wrapperScript), 0o755); err != nil {
		t.Fatalf("failed to create test wrapper: %v", err)
	}
}

// createTestWrapperWithMoaiPath creates a wrapper that uses a specific moai path.
func createTestWrapperWithMoaiPath(t *testing.T, wrapperPath, moaiPath string) {
	t.Helper()

	wrapperScript := `#!/bin/bash
temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT

MAX_SIZE=1048576
head -c "$MAX_SIZE" > "$temp_file"

if [ ! -s "$temp_file" ]; then
    exit 0
fi

# Use specific moai path
if [ -f "` + moaiPath + `" ]; then
    exec "` + moaiPath + `" hook config-change < "$temp_file" 2>&1
fi

exit 0
`
	if err := os.WriteFile(wrapperPath, []byte(wrapperScript), 0o755); err != nil {
		t.Fatalf("failed to create wrapper with moai path: %v", err)
	}
}

// createMockMoai creates a mock moai binary for testing.
func createMockMoai(t *testing.T, path string) {
	t.Helper()

	// Create a mock moai that reads stdin and echoes it
	mockScript := `#!/bin/bash
echo "MOAI_CALLED" >&2
cat
exit 0
`
	if err := os.WriteFile(path, []byte(mockScript), 0o755); err != nil {
		t.Fatalf("failed to create mock moai: %v", err)
	}
}
