package migration

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// ImplementationType.String() tests
// ---------------------------------------------------------------------------

func TestImplementationType_String_Go(t *testing.T) {
	if got := ImplementationGo.String(); got != "go" {
		t.Errorf("expected 'go', got %q", got)
	}
}

func TestImplementationType_String_Python(t *testing.T) {
	if got := ImplementationPython.String(); got != "python" {
		t.Errorf("expected 'python', got %q", got)
	}
}

func TestImplementationType_String_Unknown(t *testing.T) {
	unknown := ImplementationType(99)
	if got := unknown.String(); got != "unknown" {
		t.Errorf("expected 'unknown', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// extractVersion() tests
// ---------------------------------------------------------------------------

func TestExtractVersion_ValidFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard version output",
			input:    "moai version 1.14.0",
			expected: "1.14.0",
		},
		{
			name:     "moai-adk version output",
			input:    "moai-adk version 2.0.0-beta\nsome extra line",
			expected: "2.0.0-beta",
		},
		{
			name:     "version with more fields",
			input:    "moai version 1.0.0 (build 123)",
			expected: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVersion(tt.input)
			if got != tt.expected {
				t.Errorf("extractVersion(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractVersion_InsufficientParts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "only two words",
			input:    "moai version",
			expected: "unknown",
		},
		{
			name:     "single word",
			input:    "moai",
			expected: "unknown",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "only newlines",
			input:    "\n\n",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVersion(tt.input)
			if got != tt.expected {
				t.Errorf("extractVersion(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getCommonBinaryPaths() tests
// ---------------------------------------------------------------------------

func TestGetCommonBinaryPaths_ReturnsNonEmpty(t *testing.T) {
	paths := getCommonBinaryPaths()
	if len(paths) == 0 {
		t.Error("expected at least one binary path")
	}
}

func TestGetCommonBinaryPaths_ContainsMoaiAdk(t *testing.T) {
	paths := getCommonBinaryPaths()
	found := false
	for _, p := range paths {
		if strings.Contains(filepath.Base(p), "moai-adk") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected paths to contain moai-adk binary name")
	}
}

func TestGetCommonBinaryPaths_IncludesLocalBin(t *testing.T) {
	paths := getCommonBinaryPaths()
	found := false
	for _, p := range paths {
		if strings.Contains(p, ".local/bin") || strings.Contains(p, ".local\\bin") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected paths to include .local/bin directory")
	}
}

func TestGetCommonBinaryPaths_IncludesGopathBin(t *testing.T) {
	paths := getCommonBinaryPaths()
	found := false
	for _, p := range paths {
		if strings.Contains(p, "go/bin") || strings.Contains(p, "go\\bin") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected paths to include GOPATH/bin directory")
	}
}

func TestGetCommonBinaryPaths_CustomGOPATH(t *testing.T) {
	customGopath := t.TempDir()
	t.Setenv("GOPATH", customGopath)

	paths := getCommonBinaryPaths()
	found := false
	for _, p := range paths {
		if strings.HasPrefix(p, customGopath) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected paths to include custom GOPATH %s", customGopath)
	}
}

// ---------------------------------------------------------------------------
// Hook template tests
// ---------------------------------------------------------------------------

func TestGetGoHookTemplate(t *testing.T) {
	template := getGoHookTemplate("/usr/local/bin/moai-adk")
	if !strings.Contains(template, "/usr/local/bin/moai-adk") {
		t.Errorf("expected template to contain binary path, got %q", template)
	}
	if !strings.Contains(template, "hook") {
		t.Errorf("expected template to contain 'hook' subcommand, got %q", template)
	}
	if !strings.Contains(template, "--project-dir") {
		t.Errorf("expected template to contain --project-dir flag, got %q", template)
	}
}

func TestGetPythonHookTemplate_UV(t *testing.T) {
	template := getPythonHookTemplate("uv")
	// When pythonCmd is "uv", it should return the default template
	defaultTemplate := getDefaultPythonHookTemplate()
	if template != defaultTemplate {
		t.Errorf("expected uv to return default template\ngot:  %q\nwant: %q", template, defaultTemplate)
	}
}

func TestGetPythonHookTemplate_Python3(t *testing.T) {
	template := getPythonHookTemplate("python3")
	if !strings.Contains(template, "python3") {
		t.Errorf("expected template to contain python3, got %q", template)
	}
	if !strings.Contains(template, "CLAUDE_PROJECT_DIR") {
		t.Errorf("expected template to contain CLAUDE_PROJECT_DIR, got %q", template)
	}
}

func TestGetPythonHookTemplate_Python(t *testing.T) {
	template := getPythonHookTemplate("python")
	if !strings.Contains(template, "python") {
		t.Errorf("expected template to contain python, got %q", template)
	}
}

func TestGetDefaultPythonHookTemplate(t *testing.T) {
	template := getDefaultPythonHookTemplate()
	if !strings.Contains(template, "uv run") {
		t.Errorf("expected default template to use 'uv run', got %q", template)
	}
	if !strings.Contains(template, "CLAUDE_PROJECT_DIR") {
		t.Errorf("expected default template to reference CLAUDE_PROJECT_DIR, got %q", template)
	}
	if !strings.Contains(template, "${SHELL:-/bin/bash}") {
		t.Errorf("expected default template to use SHELL variable, got %q", template)
	}
}

// ---------------------------------------------------------------------------
// DetectionResult.GetHookCommandTemplate() tests
// ---------------------------------------------------------------------------

func TestGetHookCommandTemplate_Go(t *testing.T) {
	result := &DetectionResult{
		Found:      true,
		Type:       ImplementationGo,
		BinaryPath: "/usr/local/bin/moai-adk",
	}

	template := result.GetHookCommandTemplate()
	if !strings.Contains(template, "/usr/local/bin/moai-adk") {
		t.Errorf("expected Go template with binary path, got %q", template)
	}
}

func TestGetHookCommandTemplate_Python(t *testing.T) {
	result := &DetectionResult{
		Found:     true,
		Type:      ImplementationPython,
		PythonCmd: "python3",
	}

	template := result.GetHookCommandTemplate()
	if !strings.Contains(template, "python3") {
		t.Errorf("expected Python template with python3, got %q", template)
	}
}

func TestGetHookCommandTemplate_PythonUV(t *testing.T) {
	result := &DetectionResult{
		Found:     true,
		Type:      ImplementationPython,
		PythonCmd: "uv",
	}

	template := result.GetHookCommandTemplate()
	if !strings.Contains(template, "uv run") {
		t.Errorf("expected default uv template, got %q", template)
	}
}

func TestGetHookCommandTemplate_NotFound(t *testing.T) {
	result := &DetectionResult{
		Found: false,
	}

	template := result.GetHookCommandTemplate()
	// Should fall back to default Python template
	defaultTemplate := getDefaultPythonHookTemplate()
	if template != defaultTemplate {
		t.Errorf("expected default template for not-found\ngot:  %q\nwant: %q", template, defaultTemplate)
	}
}

func TestGetHookCommandTemplate_UnknownType(t *testing.T) {
	result := &DetectionResult{
		Found: true,
		Type:  ImplementationType(99),
	}

	template := result.GetHookCommandTemplate()
	defaultTemplate := getDefaultPythonHookTemplate()
	if template != defaultTemplate {
		t.Errorf("expected default template for unknown type\ngot:  %q\nwant: %q", template, defaultTemplate)
	}
}

// ---------------------------------------------------------------------------
// DetectionResult.GetImplementationLogInfo() tests
// ---------------------------------------------------------------------------

func TestGetImplementationLogInfo_GoImplementation(t *testing.T) {
	result := &DetectionResult{
		Found:      true,
		Type:       ImplementationGo,
		BinaryPath: "/usr/local/bin/moai-adk",
		Version:    "1.14.0",
	}

	info := result.GetImplementationLogInfo()

	if info["implementation"] != "go" {
		t.Errorf("expected implementation 'go', got %v", info["implementation"])
	}
	if info["binary_path"] != "/usr/local/bin/moai-adk" {
		t.Errorf("expected binary_path, got %v", info["binary_path"])
	}
	if info["version"] != "1.14.0" {
		t.Errorf("expected version '1.14.0', got %v", info["version"])
	}
	if _, exists := info["python_cmd"]; exists {
		t.Error("expected no python_cmd for Go implementation")
	}
}

func TestGetImplementationLogInfo_PythonImplementation(t *testing.T) {
	result := &DetectionResult{
		Found:     true,
		Type:      ImplementationPython,
		PythonCmd: "python3",
	}

	info := result.GetImplementationLogInfo()

	if info["implementation"] != "python" {
		t.Errorf("expected implementation 'python', got %v", info["implementation"])
	}
	if info["python_cmd"] != "python3" {
		t.Errorf("expected python_cmd 'python3', got %v", info["python_cmd"])
	}
	if info["fallback_reason"] != "go_binary_not_found" {
		t.Errorf("expected fallback_reason, got %v", info["fallback_reason"])
	}
	if _, exists := info["binary_path"]; exists {
		t.Error("expected no binary_path for Python implementation")
	}
}

func TestGetImplementationLogInfo_NotFound(t *testing.T) {
	result := &DetectionResult{
		Found: false,
	}

	info := result.GetImplementationLogInfo()

	if info["implementation"] != "none" {
		t.Errorf("expected implementation 'none', got %v", info["implementation"])
	}
	if info["error"] != "no_valid_implementation_found" {
		t.Errorf("expected error message, got %v", info["error"])
	}
}

// ---------------------------------------------------------------------------
// DetectImplementationWithOverride() tests
// ---------------------------------------------------------------------------

func TestDetectImplementationWithOverride_ForceGoWithValidPath(t *testing.T) {
	// Create a fake binary file
	dir := t.TempDir()
	fakeBinary := filepath.Join(dir, "moai-adk")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	result := DetectImplementationWithOverride(true, false, fakeBinary)
	if !result.Found {
		t.Error("expected Found=true for valid binary path")
	}
	if result.Type != ImplementationGo {
		t.Errorf("expected ImplementationGo, got %v", result.Type)
	}
	if result.BinaryPath != fakeBinary {
		t.Errorf("expected BinaryPath=%s, got %s", fakeBinary, result.BinaryPath)
	}
}

func TestDetectImplementationWithOverride_ForceGoWithInvalidPath(t *testing.T) {
	// Point to a nonexistent path - should fall through to auto-detection
	result := DetectImplementationWithOverride(true, false, "/nonexistent/moai-adk")
	// Result depends on system state but should not panic
	_ = result
}

func TestDetectImplementationWithOverride_ForceGoWithDirectory(t *testing.T) {
	// A directory should not be treated as a valid binary
	dir := t.TempDir()
	result := DetectImplementationWithOverride(true, false, dir)
	// Should fall through since dir is a directory, not a file
	// Result depends on further detection but should not be Go with the dir path
	if result.Found && result.Type == ImplementationGo && result.BinaryPath == dir {
		t.Error("directory should not be detected as Go binary")
	}
}

func TestDetectImplementationWithOverride_ForceGoEmptyPath(t *testing.T) {
	// Empty goBinaryPath with forceGo should fall through to auto-detection
	result := DetectImplementationWithOverride(true, false, "")
	// Should not panic and should return a valid result
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestDetectImplementationWithOverride_ForcePython(t *testing.T) {
	result := DetectImplementationWithOverride(false, true, "")
	// If Python/uv is available on the system, should find it
	if result.Found && result.Type == ImplementationPython {
		if result.PythonCmd == "" {
			t.Error("expected non-empty PythonCmd when Python is found")
		}
	}
	// If Python is not available, falls through to auto-detection
}

func TestDetectImplementationWithOverride_BothFalse(t *testing.T) {
	// Should fall through to auto-detection (same as DetectImplementation)
	result := DetectImplementationWithOverride(false, false, "")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CommonPaths == nil {
		t.Error("expected CommonPaths to be set")
	}
}

// ---------------------------------------------------------------------------
// DetectImplementation() tests
// ---------------------------------------------------------------------------

func TestDetectImplementation_ReturnsNonNil(t *testing.T) {
	result := DetectImplementation()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestDetectImplementation_HasCommonPaths(t *testing.T) {
	result := DetectImplementation()
	if len(result.CommonPaths) == 0 {
		t.Error("expected CommonPaths to be populated")
	}
}

// ---------------------------------------------------------------------------
// splitLines() tests
// ---------------------------------------------------------------------------

func TestSplitLines_MultipleLines(t *testing.T) {
	data := []byte("line1\nline2\nline3\n")
	lines := splitLines(data)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if string(lines[0]) != "line1" {
		t.Errorf("expected 'line1', got %q", string(lines[0]))
	}
	if string(lines[1]) != "line2" {
		t.Errorf("expected 'line2', got %q", string(lines[1]))
	}
	if string(lines[2]) != "line3" {
		t.Errorf("expected 'line3', got %q", string(lines[2]))
	}
}

func TestSplitLines_NoTrailingNewline(t *testing.T) {
	data := []byte("line1\nline2")
	lines := splitLines(data)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if string(lines[1]) != "line2" {
		t.Errorf("expected 'line2', got %q", string(lines[1]))
	}
}

func TestSplitLines_SingleLine(t *testing.T) {
	data := []byte("single line")
	lines := splitLines(data)
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}
	if string(lines[0]) != "single line" {
		t.Errorf("expected 'single line', got %q", string(lines[0]))
	}
}

func TestSplitLines_Empty(t *testing.T) {
	lines := splitLines([]byte{})
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestSplitLines_OnlyNewlines(t *testing.T) {
	data := []byte("\n\n\n")
	lines := splitLines(data)
	// Should produce 3 empty lines (empty before each \n)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if len(line) != 0 {
			t.Errorf("expected empty line at index %d, got %q", i, string(line))
		}
	}
}

// ---------------------------------------------------------------------------
// NewHookLogger() tests
// ---------------------------------------------------------------------------

func TestNewHookLogger_CreatesLogDir(t *testing.T) {
	dir := t.TempDir()

	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatalf("NewHookLogger failed: %v", err)
	}

	logDir := filepath.Join(dir, ".moai", "logs")
	info, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("expected log directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected .moai/logs to be a directory")
	}

	// Verify log file path contains date
	logFile := logger.GetLogFile()
	dateStr := time.Now().Format("2006-01-02")
	if !strings.Contains(logFile, dateStr) {
		t.Errorf("expected log file to contain date %s, got %s", dateStr, logFile)
	}
	if !strings.Contains(logFile, "hook-implementation-") {
		t.Errorf("expected log file to contain 'hook-implementation-', got %s", logFile)
	}
}

func TestNewHookLogger_ExistingLogDir(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, ".moai", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}

	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatalf("NewHookLogger should succeed with existing dir: %v", err)
	}
	if logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestNewHookLogger_InvalidPath(t *testing.T) {
	dir := t.TempDir()
	// Create a file where the directory should be
	blockingFile := filepath.Join(dir, ".moai")
	if err := os.WriteFile(blockingFile, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := NewHookLogger(dir)
	if err == nil {
		t.Error("expected error when .moai is a file")
	}
}

// ---------------------------------------------------------------------------
// HookLogger.GetLogFile() tests
// ---------------------------------------------------------------------------

func TestGetLogFile_ReturnsCorrectPath(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	logFile := logger.GetLogFile()
	if logFile == "" {
		t.Error("expected non-empty log file path")
	}
	if !filepath.IsAbs(logFile) {
		t.Errorf("expected absolute path, got %s", logFile)
	}
}

// ---------------------------------------------------------------------------
// HookLogger.LogHookExecution() tests
// ---------------------------------------------------------------------------

func TestLogHookExecution_Success(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	entry := &HookLogEntry{
		Hook:           "session_start",
		Implementation: "go",
		BinaryPath:     "/usr/local/bin/moai-adk",
		Version:        "1.14.0",
		Success:        true,
		Duration:       "100ms",
	}

	if err := logger.LogHookExecution(entry); err != nil {
		t.Fatalf("LogHookExecution failed: %v", err)
	}

	// Verify timestamp was set
	if entry.Timestamp == "" {
		t.Error("expected Timestamp to be set automatically")
	}

	// Read and verify log file contents
	data, err := os.ReadFile(logger.GetLogFile())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var logged HookLogEntry
	if err := json.Unmarshal(data[:len(data)-1], &logged); err != nil { // -1 for trailing newline
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if logged.Hook != "session_start" {
		t.Errorf("expected hook 'session_start', got %q", logged.Hook)
	}
	if logged.Implementation != "go" {
		t.Errorf("expected implementation 'go', got %q", logged.Implementation)
	}
	if !logged.Success {
		t.Error("expected Success=true")
	}
}

func TestLogHookExecution_WithPresetTimestamp(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	customTimestamp := "2025-12-25T12:00:00Z"
	entry := &HookLogEntry{
		Hook:           "test_hook",
		Implementation: "python",
		Timestamp:      customTimestamp,
		Success:        true,
	}

	if err := logger.LogHookExecution(entry); err != nil {
		t.Fatalf("LogHookExecution failed: %v", err)
	}

	// Verify custom timestamp was preserved
	if entry.Timestamp != customTimestamp {
		t.Errorf("expected custom timestamp to be preserved, got %q", entry.Timestamp)
	}
}

func TestLogHookExecution_WithError(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	entry := &HookLogEntry{
		Hook:           "failing_hook",
		Implementation: "go",
		Success:        false,
		Error:          "something went wrong",
	}

	if err := logger.LogHookExecution(entry); err != nil {
		t.Fatalf("LogHookExecution failed: %v", err)
	}

	// Read back and verify error field
	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Error != "something went wrong" {
		t.Errorf("expected error message, got %q", entries[0].Error)
	}
	if entries[0].Success {
		t.Error("expected Success=false")
	}
}

func TestLogHookExecution_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	hooks := []string{"session_start", "pre_tool_use", "post_tool_use"}
	for _, hook := range hooks {
		entry := &HookLogEntry{
			Hook:           hook,
			Implementation: "go",
			Success:        true,
		}
		if err := logger.LogHookExecution(entry); err != nil {
			t.Fatalf("LogHookExecution failed for %s: %v", hook, err)
		}
	}

	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	for i, hook := range hooks {
		if entries[i].Hook != hook {
			t.Errorf("entry %d: expected hook %q, got %q", i, hook, entries[i].Hook)
		}
	}
}

// ---------------------------------------------------------------------------
// HookLogger.LogHookExecutionFromResult() tests
// ---------------------------------------------------------------------------

func TestLogHookExecutionFromResult_GoSuccess(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	result := &DetectionResult{
		Found:      true,
		Type:       ImplementationGo,
		BinaryPath: "/usr/local/bin/moai-adk",
		Version:    "1.14.0",
	}

	duration := 150 * time.Millisecond
	if err := logger.LogHookExecutionFromResult("session_start", result, duration, nil); err != nil {
		t.Fatalf("LogHookExecutionFromResult failed: %v", err)
	}

	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Hook != "session_start" {
		t.Errorf("expected hook 'session_start', got %q", entry.Hook)
	}
	if entry.Implementation != "go" {
		t.Errorf("expected implementation 'go', got %q", entry.Implementation)
	}
	if entry.BinaryPath != "/usr/local/bin/moai-adk" {
		t.Errorf("expected binary_path, got %q", entry.BinaryPath)
	}
	if entry.Version != "1.14.0" {
		t.Errorf("expected version '1.14.0', got %q", entry.Version)
	}
	if !entry.Success {
		t.Error("expected Success=true")
	}
	if entry.Duration != duration.String() {
		t.Errorf("expected duration %q, got %q", duration.String(), entry.Duration)
	}
}

func TestLogHookExecutionFromResult_PythonSuccess(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	result := &DetectionResult{
		Found:     true,
		Type:      ImplementationPython,
		PythonCmd: "python3",
	}

	duration := 200 * time.Millisecond
	if err := logger.LogHookExecutionFromResult("pre_tool_use", result, duration, nil); err != nil {
		t.Fatalf("LogHookExecutionFromResult failed: %v", err)
	}

	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Implementation != "python" {
		t.Errorf("expected implementation 'python', got %q", entry.Implementation)
	}
	// Python entries should not have BinaryPath or Version
	if entry.BinaryPath != "" {
		t.Errorf("expected empty BinaryPath for Python, got %q", entry.BinaryPath)
	}
	if entry.Version != "" {
		t.Errorf("expected empty Version for Python, got %q", entry.Version)
	}
}

func TestLogHookExecutionFromResult_WithError(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	result := &DetectionResult{
		Found:      true,
		Type:       ImplementationGo,
		BinaryPath: "/usr/local/bin/moai-adk",
		Version:    "1.14.0",
	}

	hookErr := errors.New("hook execution failed")
	duration := 50 * time.Millisecond
	if err := logger.LogHookExecutionFromResult("failing_hook", result, duration, hookErr); err != nil {
		t.Fatalf("LogHookExecutionFromResult failed: %v", err)
	}

	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Success {
		t.Error("expected Success=false")
	}
	if entry.Error != "hook execution failed" {
		t.Errorf("expected error message, got %q", entry.Error)
	}
}

// ---------------------------------------------------------------------------
// HookLogger.ReadLogs() tests
// ---------------------------------------------------------------------------

func TestReadLogs_NoLogFile(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatalf("ReadLogs should not error for nonexistent file: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestReadLogs_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create empty log file
	if err := os.WriteFile(logger.GetLogFile(), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatalf("ReadLogs should not error for empty file: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty file, got %d", len(entries))
	}
}

func TestReadLogs_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Write invalid JSON lines mixed with valid ones
	content := `{"hook":"valid","implementation":"go","success":true,"timestamp":"2025-01-01T00:00:00Z"}
{invalid json line}
{"hook":"also_valid","implementation":"python","success":false,"timestamp":"2025-01-01T00:00:01Z"}
`
	if err := os.WriteFile(logger.GetLogFile(), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatalf("ReadLogs should skip invalid entries: %v", err)
	}
	// Should have 2 valid entries, skipping the invalid one
	if len(entries) != 2 {
		t.Errorf("expected 2 valid entries, got %d", len(entries))
	}
}

// ---------------------------------------------------------------------------
// HookLogger.RotateLogs() tests
// ---------------------------------------------------------------------------

func TestRotateLogs_RemovesOldFiles(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	logDir := filepath.Join(dir, ".moai", "logs")

	// Create an old log file (simulate 10 days ago by modifying modtime)
	oldFile := filepath.Join(logDir, "hook-implementation-2020-01-01.log")
	if err := os.WriteFile(oldFile, []byte("old data"), 0644); err != nil {
		t.Fatal(err)
	}
	// Set modification time to 10 days ago
	oldTime := time.Now().AddDate(0, 0, -10)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Create a recent log file
	recentFile := filepath.Join(logDir, "hook-implementation-recent.log")
	if err := os.WriteFile(recentFile, []byte("recent data"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := logger.RotateLogs(); err != nil {
		t.Fatalf("RotateLogs failed: %v", err)
	}

	// Old file should be removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("expected old log file to be removed")
	}

	// Recent file should still exist
	if _, err := os.Stat(recentFile); err != nil {
		t.Error("expected recent log file to be preserved")
	}
}

func TestRotateLogs_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Should not error on empty directory
	if err := logger.RotateLogs(); err != nil {
		t.Fatalf("RotateLogs should not error on empty dir: %v", err)
	}
}

func TestRotateLogs_SkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	logDir := filepath.Join(dir, ".moai", "logs")

	// Create a subdirectory in logs - should be skipped
	subDir := filepath.Join(logDir, "archive")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := logger.RotateLogs(); err != nil {
		t.Fatalf("RotateLogs should skip directories: %v", err)
	}

	// Subdirectory should still exist
	if _, err := os.Stat(subDir); err != nil {
		t.Error("expected subdirectory to be preserved")
	}
}

func TestRotateLogs_InvalidDir(t *testing.T) {
	logger := &HookLogger{
		logDir:  "/nonexistent/path/logs",
		logFile: "/nonexistent/path/logs/test.log",
	}

	err := logger.RotateLogs()
	if err == nil {
		t.Error("expected error for nonexistent log directory")
	}
}

// ---------------------------------------------------------------------------
// HookLogEntry JSON serialization tests
// ---------------------------------------------------------------------------

func TestHookLogEntry_JSONRoundTrip(t *testing.T) {
	entry := HookLogEntry{
		Hook:           "session_start",
		Implementation: "go",
		BinaryPath:     "/usr/local/bin/moai-adk",
		Version:        "1.14.0",
		Timestamp:      "2025-12-25T12:00:00Z",
		Duration:       "150ms",
		Success:        true,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded HookLogEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Hook != entry.Hook {
		t.Errorf("Hook mismatch: %q != %q", decoded.Hook, entry.Hook)
	}
	if decoded.Implementation != entry.Implementation {
		t.Errorf("Implementation mismatch: %q != %q", decoded.Implementation, entry.Implementation)
	}
	if decoded.BinaryPath != entry.BinaryPath {
		t.Errorf("BinaryPath mismatch: %q != %q", decoded.BinaryPath, entry.BinaryPath)
	}
	if decoded.Version != entry.Version {
		t.Errorf("Version mismatch: %q != %q", decoded.Version, entry.Version)
	}
	if decoded.Success != entry.Success {
		t.Errorf("Success mismatch: %v != %v", decoded.Success, entry.Success)
	}
}

func TestHookLogEntry_JSONOmitsEmpty(t *testing.T) {
	entry := HookLogEntry{
		Hook:           "test",
		Implementation: "python",
		Timestamp:      "2025-01-01T00:00:00Z",
		Success:        true,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, "binary_path") {
		t.Error("expected binary_path to be omitted when empty")
	}
	if strings.Contains(s, "version") {
		t.Error("expected version to be omitted when empty")
	}
	if strings.Contains(s, `"error"`) {
		t.Error("expected error to be omitted when empty")
	}
}

// ---------------------------------------------------------------------------
// LogHookExecution error path tests
// ---------------------------------------------------------------------------

func TestLogHookExecution_OpenFileError(t *testing.T) {
	// Create a logger pointing to a path where the file cannot be opened
	logger := &HookLogger{
		logDir:  "/nonexistent/dir",
		logFile: "/nonexistent/dir/test.log",
	}

	entry := &HookLogEntry{
		Hook:           "test",
		Implementation: "go",
		Success:        true,
	}

	err := logger.LogHookExecution(entry)
	if err == nil {
		t.Error("expected error when log file cannot be opened")
	}
}

// ---------------------------------------------------------------------------
// ReadLogs error path tests
// ---------------------------------------------------------------------------

func TestReadLogs_UnreadableFile(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create log file and make it unreadable
	logFile := logger.GetLogFile()
	if err := os.WriteFile(logFile, []byte("some data\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(logFile, 0000); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(logFile, 0644) // Restore for cleanup
	}()

	_, err = logger.ReadLogs()
	if err == nil {
		t.Error("expected error when log file is unreadable")
	}
}

// ---------------------------------------------------------------------------
// getCommonBinaryPaths edge cases
// ---------------------------------------------------------------------------

func TestGetCommonBinaryPaths_EmptyGOPATH(t *testing.T) {
	// Clear GOPATH to exercise the default path
	t.Setenv("GOPATH", "")

	paths := getCommonBinaryPaths()
	if len(paths) == 0 {
		t.Error("expected paths even with empty GOPATH")
	}

	// Should fall back to ~/go
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}
	expectedGoPath := filepath.Join(homeDir, "go", "bin", "moai-adk")
	found := false
	for _, p := range paths {
		if p == expectedGoPath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected default GOPATH path %s in results", expectedGoPath)
	}
}

// ---------------------------------------------------------------------------
// DetectImplementation additional coverage
// ---------------------------------------------------------------------------

func TestDetectImplementation_ResultContainsFoundOrNot(t *testing.T) {
	result := DetectImplementation()
	// We cannot control what's installed, but we can verify the result is consistent
	if result.Found {
		switch result.Type {
		case ImplementationGo:
			if result.BinaryPath == "" {
				t.Error("Go implementation found but BinaryPath is empty")
			}
		case ImplementationPython:
			if result.PythonCmd == "" {
				t.Error("Python implementation found but PythonCmd is empty")
			}
		}
	}
}

// ---------------------------------------------------------------------------
// DetectImplementationWithOverride additional coverage
// ---------------------------------------------------------------------------

func TestDetectImplementationWithOverride_ForceGoMatchesCommonPathBasename(t *testing.T) {
	// Create a fake binary at a temp path that matches a common basename
	dir := t.TempDir()
	fakeBinary := filepath.Join(dir, "moai-adk")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	// Force Go with a path that does NOT exist as-is but basename matches
	// The code tries CommonPaths when the exact path fails
	result := DetectImplementationWithOverride(true, false, "/nonexistent/moai-adk")
	// This will fall through to auto-detection since /nonexistent doesn't exist
	// and none of the CommonPaths will match either
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestDetectImplementationWithOverride_BothForceFlags(t *testing.T) {
	// When both flags are true, forceGo is checked first
	dir := t.TempDir()
	fakeBinary := filepath.Join(dir, "moai-adk")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	result := DetectImplementationWithOverride(true, true, fakeBinary)
	// forceGo with valid path should win
	if !result.Found || result.Type != ImplementationGo {
		t.Errorf("expected Go implementation with both flags set, got found=%v type=%v",
			result.Found, result.Type)
	}
}

// ---------------------------------------------------------------------------
// Integration tests: full logger workflow
// ---------------------------------------------------------------------------

func TestLoggerWorkflow_CreateLogReadRotate(t *testing.T) {
	dir := t.TempDir()

	// Create logger
	logger, err := NewHookLogger(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Log several entries
	for i := 0; i < 5; i++ {
		entry := &HookLogEntry{
			Hook:           "hook_" + string(rune('a'+i)),
			Implementation: "go",
			BinaryPath:     "/usr/local/bin/moai-adk",
			Success:        true,
		}
		if err := logger.LogHookExecution(entry); err != nil {
			t.Fatalf("LogHookExecution failed: %v", err)
		}
	}

	// Read logs
	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}

	// Rotate (nothing should be removed since all are recent)
	if err := logger.RotateLogs(); err != nil {
		t.Fatal(err)
	}

	// Logs should still be readable
	entries, err = logger.ReadLogs()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries after rotation, got %d", len(entries))
	}
}
