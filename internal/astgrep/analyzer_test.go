package astgrep

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockExecutor is a test double for CommandExecutor.
type mockExecutor struct {
	// outputs maps "name args..." to the response.
	outputs map[string]mockResponse
	// calls records all executed commands.
	calls []string
}

type mockResponse struct {
	output []byte
	err    error
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		outputs: make(map[string]mockResponse),
	}
}

func (m *mockExecutor) on(key string, output []byte, err error) {
	m.outputs[key] = mockResponse{output: output, err: err}
}

func (m *mockExecutor) Execute(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
	key := name
	for _, a := range args {
		key += " " + a
	}
	m.calls = append(m.calls, key)

	if resp, ok := m.outputs[key]; ok {
		return resp.output, resp.err
	}

	// Default: return error for unknown commands.
	return nil, fmt.Errorf("mock: unknown command %q", key)
}

// --- DetectLanguage Tests ---

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filePath string
		expected string
	}{
		{"main.py", "python"},
		{"app.js", "javascript"},
		{"lib.mjs", "javascript"},
		{"util.cjs", "javascript"},
		{"component.jsx", "javascriptreact"},
		{"app.ts", "typescript"},
		{"lib.mts", "typescript"},
		{"util.cts", "typescript"},
		{"component.tsx", "typescriptreact"},
		{"main.c", "c"},
		{"header.h", "c"},
		{"main.cpp", "cpp"},
		{"main.cc", "cpp"},
		{"main.cxx", "cpp"},
		{"header.hpp", "cpp"},
		{"Program.cs", "csharp"},
		{"main.go", "go"},
		{"main.rs", "rust"},
		{"Main.java", "java"},
		{"Main.kt", "kotlin"},
		{"script.kts", "kotlin"},
		{"app.rb", "ruby"},
		{"main.swift", "swift"},
		{"script.lua", "lua"},
		{"index.html", "html"},
		{"App.vue", "vue"},
		{"App.svelte", "svelte"},
		// Unknown extensions
		{"readme.txt", "text"},
		{"data.csv", "text"},
		{"image.png", "text"},
		{"noextension", "text"},
		// Path with directories
		{"/path/to/main.py", "python"},
		{"src/components/App.tsx", "typescriptreact"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := DetectLanguage(tt.filePath)
			if got != tt.expected {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filePath, got, tt.expected)
			}
		})
	}
}

// --- IsSGAvailable Tests ---

func TestIsSGAvailable_True(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0\n"), nil)

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	if !a.IsSGAvailable(ctx) {
		t.Error("expected sg to be available")
	}
}

func TestIsSGAvailable_False(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	if a.IsSGAvailable(ctx) {
		t.Error("expected sg to be unavailable")
	}
}

func TestIsSGAvailable_Caching(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0\n"), nil)

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	// First call.
	a.IsSGAvailable(ctx)
	// Second call should use cache.
	a.IsSGAvailable(ctx)

	// Should only have been called once.
	versionCalls := 0
	for _, c := range mock.calls {
		if c == "sg --version" {
			versionCalls++
		}
	}
	if versionCalls != 1 {
		t.Errorf("expected 1 version call, got %d", versionCalls)
	}
}

func TestIsSGAvailable_Timeout(t *testing.T) {
	mock := newMockExecutor()
	// Simulate timeout by returning an error.
	mock.on("sg --version", nil, context.DeadlineExceeded)

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	if a.IsSGAvailable(ctx) {
		t.Error("expected sg to be unavailable on timeout")
	}
}

// --- ShouldIncludeFile Tests ---

func TestShouldIncludeFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		config   *ScanConfig
		expected bool
	}{
		{
			name:     "nil config includes all",
			filePath: "main.go",
			config:   nil,
			expected: true,
		},
		{
			name:     "exclude node_modules",
			filePath: "node_modules/package/index.js",
			config:   &ScanConfig{ExcludePatterns: []string{"node_modules"}},
			expected: false,
		},
		{
			name:     "exclude .git",
			filePath: ".git/config",
			config:   &ScanConfig{ExcludePatterns: []string{".git"}},
			expected: false,
		},
		{
			name:     "include only go files",
			filePath: "main.go",
			config:   &ScanConfig{IncludePatterns: []string{"*.go"}},
			expected: true,
		},
		{
			name:     "include filter rejects non-matching",
			filePath: "main.py",
			config:   &ScanConfig{IncludePatterns: []string{"*.go"}},
			expected: false,
		},
		{
			name:     "no exclude no include includes all",
			filePath: "anything.txt",
			config:   &ScanConfig{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldIncludeFile(tt.filePath, tt.config)
			if got != tt.expected {
				t.Errorf("ShouldIncludeFile(%q) = %v, want %v", tt.filePath, got, tt.expected)
			}
		})
	}
}

// --- Scan Tests ---

func makeSGJSON(t *testing.T, matches []sgMatch) []byte {
	t.Helper()
	data, err := json.Marshal(matches)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestScan_WithMatches(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:  "console.log(msg)",
			Range: sgRange{Start: sgPosition{Line: 2, Column: 4}, End: sgPosition{Line: 2, Column: 20}},
			File:  "test.js",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern console.log($MSG) --json /project", makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.Scan(ctx, []string{"console.log($MSG)"}, []string{"/project"})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Matches))
	}

	m := result.Matches[0]
	if m.File != "test.js" {
		t.Errorf("expected file test.js, got %q", m.File)
	}
	if m.Line != 3 { // 0-indexed line 2 -> 1-indexed line 3
		t.Errorf("expected line 3, got %d", m.Line)
	}
	if m.Text != "console.log(msg)" {
		t.Errorf("expected text console.log(msg), got %q", m.Text)
	}
}

func TestScan_SGNotAvailable(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.Scan(ctx, []string{"pattern"}, []string{"/tmp"})
	if err != nil {
		t.Fatalf("Scan should not error when sg unavailable: %v", err)
	}
	if len(result.Matches) != 0 {
		t.Error("expected empty matches when sg unavailable")
	}
}

func TestScan_EmptyOutput(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern nonexistent --json /project", []byte("[]"), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.Scan(ctx, []string{"nonexistent"}, []string{"/project"})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(result.Matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(result.Matches))
	}
}

// --- FindPattern Tests ---

func TestFindPattern_WithMatches(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:  "func main()",
			Range: sgRange{Start: sgPosition{Line: 5, Column: 0}, End: sgPosition{Line: 5, Column: 11}},
			File:  "main.go",
		},
		{
			Text:  "func helper(x int)",
			Range: sgRange{Start: sgPosition{Line: 10, Column: 0}, End: sgPosition{Line: 10, Column: 18}},
			File:  "util.go",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern func $NAME($PARAMS) --lang go --json /project", makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	matches, err := a.FindPattern(ctx, "func $NAME($PARAMS)", "go")
	if err != nil {
		t.Fatalf("FindPattern failed: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}

	if matches[0].File != "main.go" {
		t.Errorf("expected file main.go, got %q", matches[0].File)
	}
	if matches[0].Line != 6 { // 0-indexed 5 -> 1-indexed 6
		t.Errorf("expected line 6, got %d", matches[0].Line)
	}
}

func TestFindPattern_SGNotAvailable(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	matches, err := a.FindPattern(ctx, "pattern", "go")
	if err != nil {
		t.Fatalf("FindPattern should not error when sg unavailable: %v", err)
	}
	if matches != nil {
		t.Error("expected nil matches when sg unavailable")
	}
}

// --- Replace Tests ---

func TestReplace_WithMatches(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:  "fmt.Println(msg)",
			Range: sgRange{Start: sgPosition{Line: 7, Column: 1}, End: sgPosition{Line: 7, Column: 17}},
			File:  "main.go",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern fmt.Println($MSG) --rewrite log.Info($MSG) --lang go --json /project",
		makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	changes, err := a.Replace(ctx, "fmt.Println($MSG)", "log.Info($MSG)", "go", []string{"/project"})
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	if changes[0].OldCode != "fmt.Println(msg)" {
		t.Errorf("expected old code fmt.Println(msg), got %q", changes[0].OldCode)
	}
	if changes[0].NewCode != "log.Info($MSG)" {
		t.Errorf("expected new code log.Info($MSG), got %q", changes[0].NewCode)
	}
}

func TestReplace_SGNotAvailable(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	changes, err := a.Replace(ctx, "pattern", "replacement", "go", nil)
	if err != nil {
		t.Fatalf("Replace should not error when sg unavailable: %v", err)
	}
	if changes != nil {
		t.Error("expected nil changes when sg unavailable")
	}
}

// --- ScanFile Tests ---

func TestScanFile_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	if err := os.WriteFile(testFile, []byte("eval(code)\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sgOutput := []sgMatch{
		{
			Text:     "eval(code)",
			Range:    sgRange{Start: sgPosition{Line: 0, Column: 0}, End: sgPosition{Line: 0, Column: 10}},
			File:     testFile,
			RuleID:   "no-eval",
			Severity: "error",
			Message:  "Do not use eval",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on(fmt.Sprintf("sg scan --json %s", testFile), makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.ScanFile(ctx, testFile, nil)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if result.Language != "python" {
		t.Errorf("expected language python, got %q", result.Language)
	}
	if result.Files != 1 {
		t.Errorf("expected 1 file scanned, got %d", result.Files)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Matches))
	}
	if result.Matches[0].Severity != "error" {
		t.Errorf("expected severity error, got %q", result.Matches[0].Severity)
	}
}

func TestScanFile_NotFound(t *testing.T) {
	mock := newMockExecutor()
	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.ScanFile(ctx, "/nonexistent/file.py", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestScanFile_SGNotAvailable(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.ScanFile(ctx, testFile, nil)
	if err != nil {
		t.Fatalf("ScanFile should not error when sg unavailable: %v", err)
	}
	if len(result.Matches) != 0 {
		t.Error("expected empty matches when sg unavailable")
	}
	if result.Language != "go" {
		t.Errorf("expected language go, got %q", result.Language)
	}
}

// --- ScanProject Tests ---

func TestScanProject_NotFound(t *testing.T) {
	mock := newMockExecutor()
	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.ScanProject(ctx, "/nonexistent/path", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent project path")
	}
}

func TestScanProject_ExcludePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure.
	srcDir := filepath.Join(tmpDir, "src")
	nodeDir := filepath.Join(tmpDir, "node_modules", "pkg")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nodeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create test files.
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeDir, "index.js"), []byte("module.exports = {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found")) // sg not available

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	config := &ScanConfig{ExcludePatterns: []string{"node_modules"}}
	result, err := a.ScanProject(ctx, tmpDir, config)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	// Only src/main.go should be scanned (node_modules excluded).
	if result.TotalFiles != 1 {
		t.Errorf("expected 1 file scanned (excluding node_modules), got %d", result.TotalFiles)
	}
}

func TestScanProject_IncludePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "app.py"), []byte("print('hello')\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	config := &ScanConfig{IncludePatterns: []string{"*.go"}}
	result, err := a.ScanProject(ctx, tmpDir, config)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	if result.TotalFiles != 1 {
		t.Errorf("expected 1 file scanned (only .go), got %d", result.TotalFiles)
	}
}

// --- PatternSearch Tests ---

func TestPatternSearch_SetsRuleID(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:  "func main()",
			Range: sgRange{Start: sgPosition{Line: 0, Column: 0}, End: sgPosition{Line: 0, Column: 11}},
			File:  "main.go",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern func $NAME($$$ARGS) --lang go --json /project",
		makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	matches, err := a.PatternSearch(ctx, "func $NAME($$$ARGS)", "go", "/project")
	if err != nil {
		t.Fatalf("PatternSearch failed: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	expectedRuleID := "pattern:func $NAME($$$ARGS)"
	if matches[0].Rule != expectedRuleID {
		t.Errorf("expected rule_id %q, got %q", expectedRuleID, matches[0].Rule)
	}
}

func TestPatternSearch_LongPatternTruncated(t *testing.T) {
	longPattern := "very_long_pattern_that_exceeds_thirty_characters_limit"

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on(fmt.Sprintf("sg run --pattern %s --lang go --json /project", longPattern),
		[]byte("[]"), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.PatternSearch(ctx, longPattern, "go", "/project")
	if err != nil {
		t.Fatalf("PatternSearch failed: %v", err)
	}
}

// --- PatternReplace Tests ---

func TestPatternReplace_DryRun(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:  "fmt.Println(msg)",
			Range: sgRange{Start: sgPosition{Line: 3, Column: 1}, End: sgPosition{Line: 3, Column: 17}},
			File:  "main.go",
		},
		{
			Text:  "fmt.Println(err)",
			Range: sgRange{Start: sgPosition{Line: 8, Column: 1}, End: sgPosition{Line: 8, Column: 17}},
			File:  "util.go",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern fmt.Println($MSG) --lang go --json /project",
		makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.PatternReplace(ctx, "fmt.Println($MSG)", "log.Info($MSG)", "go", "/project", true)
	if err != nil {
		t.Fatalf("PatternReplace failed: %v", err)
	}

	if !result.DryRun {
		t.Error("expected DryRun=true")
	}
	if result.MatchesFound != 2 {
		t.Errorf("expected 2 matches found, got %d", result.MatchesFound)
	}
	if result.FilesModified != 2 {
		t.Errorf("expected 2 files modified, got %d", result.FilesModified)
	}
	if len(result.Changes) != 2 {
		t.Errorf("expected 2 changes, got %d", len(result.Changes))
	}
}

func TestPatternReplace_SGNotAvailable(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.PatternReplace(ctx, "pattern", "replacement", "go", "/tmp", false)
	if err != nil {
		t.Fatalf("PatternReplace should not error when sg unavailable: %v", err)
	}
	if result.MatchesFound != 0 {
		t.Error("expected 0 matches when sg unavailable")
	}
	if result.FilesModified != 0 {
		t.Error("expected 0 files modified when sg unavailable")
	}
}

// --- DefaultScanConfig Tests ---

func TestDefaultScanConfig(t *testing.T) {
	config := DefaultScanConfig()

	expected := []string{"node_modules", ".git", "__pycache__"}
	if len(config.ExcludePatterns) != len(expected) {
		t.Fatalf("expected %d exclude patterns, got %d", len(expected), len(config.ExcludePatterns))
	}
	for i, pattern := range expected {
		if config.ExcludePatterns[i] != pattern {
			t.Errorf("exclude pattern %d: expected %q, got %q", i, pattern, config.ExcludePatterns[i])
		}
	}
}

// --- parseSGOutput Tests ---

func TestParseSGOutput_EmptyInput(t *testing.T) {
	matches, err := parseSGOutput(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseSGOutput_EmptyArray(t *testing.T) {
	matches, err := parseSGOutput([]byte("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestParseSGOutput_InvalidJSON(t *testing.T) {
	_, err := parseSGOutput([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseSGOutput_WithRuleInfo(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:     "eval(code)",
			Range:    sgRange{Start: sgPosition{Line: 4, Column: 0}, End: sgPosition{Line: 4, Column: 10}},
			File:     "test.py",
			RuleID:   "no-eval",
			Severity: "error",
			Message:  "Do not use eval",
		},
	}
	data, marshalErr := json.Marshal(sgOutput)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}

	matches, err := parseSGOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	m := matches[0]
	if m.Rule != "no-eval" {
		t.Errorf("expected rule no-eval, got %q", m.Rule)
	}
	if m.Severity != "error" {
		t.Errorf("expected severity error, got %q", m.Severity)
	}
	if m.Message != "Do not use eval" {
		t.Errorf("unexpected message: %q", m.Message)
	}
	if m.Line != 5 { // 0-indexed 4 -> 1-indexed 5
		t.Errorf("expected line 5, got %d", m.Line)
	}
}

// --- Truncate Tests ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		n        int
		expected string
	}{
		{"short", 30, "short"},
		{"exactly thirty characters long!", 30, "exactly thirty characters long"},
		{"this is a very long string that exceeds thirty characters", 30, "this is a very long string tha"},
		{"", 30, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.n)
		if got != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.expected)
		}
	}
}

// --- Duration Tests ---

func TestScan_IncludesDuration(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer("/tmp", WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.Scan(ctx, []string{"pattern"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if result.Duration < 0 {
		t.Error("duration should be non-negative")
	}
}

// --- BuildScanArgs Tests ---

func TestBuildScanArgs(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		config   *ScanConfig
		expected []string
	}{
		{
			name:     "no config",
			filePath: "test.py",
			config:   nil,
			expected: []string{"scan", "--json", "test.py"},
		},
		{
			name:     "with rules path",
			filePath: "test.py",
			config:   &ScanConfig{RulesPath: "/rules/sgconfig.yml"},
			expected: []string{"scan", "--json", "--config", "/rules/sgconfig.yml", "test.py"},
		},
		{
			name:     "empty rules path",
			filePath: "test.py",
			config:   &ScanConfig{RulesPath: ""},
			expected: []string{"scan", "--json", "test.py"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildScanArgs(tt.filePath, tt.config)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d args, got %d: %v", len(tt.expected), len(got), got)
			}
			for i, arg := range tt.expected {
				if got[i] != arg {
					t.Errorf("arg %d: expected %q, got %q", i, arg, got[i])
				}
			}
		})
	}
}

// --- PatternReplace Non-DryRun Tests ---

func TestPatternReplace_ActualReplace(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:  "fmt.Println(msg)",
			Range: sgRange{Start: sgPosition{Line: 3, Column: 1}, End: sgPosition{Line: 3, Column: 17}},
			File:  "main.go",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern fmt.Println($MSG) --lang go --json /project",
		makeSGJSON(t, sgOutput), nil)
	mock.on("sg run --pattern fmt.Println($MSG) --rewrite log.Info($MSG) --lang go /project",
		[]byte(""), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.PatternReplace(ctx, "fmt.Println($MSG)", "log.Info($MSG)", "go", "/project", false)
	if err != nil {
		t.Fatalf("PatternReplace failed: %v", err)
	}

	if result.DryRun {
		t.Error("expected DryRun=false")
	}
	if result.MatchesFound != 1 {
		t.Errorf("expected 1 match found, got %d", result.MatchesFound)
	}
	if result.FilesModified != 1 {
		t.Errorf("expected 1 file modified, got %d", result.FilesModified)
	}
}

func TestPatternReplace_ActualReplace_ExecutionError(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:  "eval(x)",
			Range: sgRange{Start: sgPosition{Line: 0, Column: 0}, End: sgPosition{Line: 0, Column: 7}},
			File:  "test.py",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern eval($X) --lang python --json /project",
		makeSGJSON(t, sgOutput), nil)
	mock.on("sg run --pattern eval($X) --rewrite safe_eval($X) --lang python /project",
		nil, fmt.Errorf("write permission denied"))

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.PatternReplace(ctx, "eval($X)", "safe_eval($X)", "python", "/project", false)
	if err == nil {
		t.Fatal("expected error when replacement execution fails")
	}
}

// --- ScanProject Additional Tests ---

func TestScanProject_DefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	// nil config should use defaults.
	result, err := a.ScanProject(ctx, tmpDir, nil)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	if result.TotalFiles != 1 {
		t.Errorf("expected 1 file scanned, got %d", result.TotalFiles)
	}
}

func TestScanProject_NotADirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notadir.go")
	if err := os.WriteFile(filePath, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mock := newMockExecutor()
	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.ScanProject(ctx, filePath, nil)
	if err == nil {
		t.Fatal("expected error for non-directory path")
	}
}

func TestScanProject_WithSGMatches(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "app.py")
	if err := os.WriteFile(testFile, []byte("eval(code)\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sgOutput := []sgMatch{
		{
			Text:     "eval(code)",
			Range:    sgRange{Start: sgPosition{Line: 0, Column: 0}, End: sgPosition{Line: 0, Column: 10}},
			File:     testFile,
			Severity: "error",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on(fmt.Sprintf("sg scan --json %s", testFile), makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.ScanProject(ctx, tmpDir, nil)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	if result.TotalFiles != 1 {
		t.Errorf("expected 1 file scanned, got %d", result.TotalFiles)
	}
	if result.TotalMatches != 1 {
		t.Errorf("expected 1 match, got %d", result.TotalMatches)
	}
	if result.BySeverity["error"] != 1 {
		t.Errorf("expected 1 error severity, got %d", result.BySeverity["error"])
	}
}

func TestScanProject_MatchWithEmptySeverity(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "lib.js")
	if err := os.WriteFile(testFile, []byte("console.log('hi')\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sgOutput := []sgMatch{
		{
			Text:  "console.log('hi')",
			Range: sgRange{Start: sgPosition{Line: 0, Column: 0}, End: sgPosition{Line: 0, Column: 17}},
			File:  testFile,
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on(fmt.Sprintf("sg scan --json %s", testFile), makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	result, err := a.ScanProject(ctx, tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Empty severity should default to "info".
	if result.BySeverity["info"] != 1 {
		t.Errorf("expected 1 info severity for empty severity, got %d", result.BySeverity["info"])
	}
}

// --- Scan with Default Paths ---

func TestScan_DefaultPaths(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern test_pattern --json /project", []byte("[]"), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	// Empty paths should default to workDir.
	result, err := a.Scan(ctx, []string{"test_pattern"}, nil)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// --- Replace with Default Paths ---

func TestReplace_DefaultPaths(t *testing.T) {
	sgOutput := []sgMatch{
		{
			Text:  "old_code()",
			Range: sgRange{Start: sgPosition{Line: 0, Column: 0}, End: sgPosition{Line: 0, Column: 10}},
			File:  "main.go",
		},
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern old_code() --rewrite new_code() --lang go --json /project",
		makeSGJSON(t, sgOutput), nil)

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	// Empty paths should default to workDir.
	changes, err := a.Replace(ctx, "old_code()", "new_code()", "go", nil)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}
}

// --- Scan Error from sg execution ---

func TestScan_ExecutionError(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern crash --json /project", nil, fmt.Errorf("sg crashed"))

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.Scan(ctx, []string{"crash"}, []string{"/project"})
	if err == nil {
		t.Fatal("expected error when sg execution fails")
	}
}

// --- FindPattern Error ---

func TestFindPattern_ExecutionError(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern bad --lang go --json /project", nil, fmt.Errorf("sg error"))

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.FindPattern(ctx, "bad", "go")
	if err == nil {
		t.Fatal("expected error from FindPattern execution failure")
	}
}

// --- ScanFile with config ---

func TestScanFile_WithRulesPath(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on(fmt.Sprintf("sg scan --json --config /rules/sg.yml %s", testFile), []byte("[]"), nil)

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	config := &ScanConfig{RulesPath: "/rules/sg.yml"}
	result, err := a.ScanFile(ctx, testFile, config)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}
	if result.Language != "go" {
		t.Errorf("expected language go, got %q", result.Language)
	}
}

// --- PatternSearch error paths ---

func TestPatternSearch_SGNotAvailable(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", nil, fmt.Errorf("not found"))

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	matches, err := a.PatternSearch(ctx, "pattern", "go", "/project")
	if err != nil {
		t.Fatalf("PatternSearch should not error when sg unavailable: %v", err)
	}
	if matches != nil {
		t.Error("expected nil matches")
	}
}

func TestPatternSearch_ExecutionError(t *testing.T) {
	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on("sg run --pattern bad --lang go --json /project", nil, fmt.Errorf("crash"))

	a := NewAnalyzer("/project", WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.PatternSearch(ctx, "bad", "go", "/project")
	if err == nil {
		t.Fatal("expected error from PatternSearch execution failure")
	}
}

// --- ScanFile execution error ---

func TestScanFile_ExecutionError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	if err := os.WriteFile(testFile, []byte("code\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mock := newMockExecutor()
	mock.on("sg --version", []byte("0.25.0"), nil)
	mock.on(fmt.Sprintf("sg scan --json %s", testFile), nil, fmt.Errorf("sg scan failed"))

	a := NewAnalyzer(tmpDir, WithCommandExecutor(mock))
	ctx := context.Background()

	_, err := a.ScanFile(ctx, testFile, nil)
	if err == nil {
		t.Fatal("expected error when sg scan fails")
	}
}

// --- ShouldIncludeFile with exclude glob pattern ---

func TestShouldIncludeFile_ExcludeGlob(t *testing.T) {
	config := &ScanConfig{ExcludePatterns: []string{"*.min.js"}}
	if ShouldIncludeFile("bundle.min.js", config) {
		t.Error("expected bundle.min.js to be excluded")
	}
	if !ShouldIncludeFile("bundle.js", config) {
		t.Error("expected bundle.js to be included")
	}
}

// Ensure unused imports are used.
var _ = time.Now
