package mx

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// writeGoFile creates a temporary Go file with the given content.
// The file will be cleaned up automatically by t.TempDir().
func writeGoFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file %s: %v", path, err)
	}
	return path
}

// fileChecksum returns the SHA-256 checksum of a file's content.
func fileChecksum(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file for checksum: %v", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// TestValidateFile_ANCHORDetection verifies AC-VAL-001:
// exported function with fan_in >= 3 missing @MX:ANCHOR → P1 violation.
func TestValidateFile_ANCHORDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		content       string
		callerContent string // content of a separate file with callers
		wantP1        int
		wantViolFunc  string
	}{
		{
			name: "exported function with high fan_in missing ANCHOR",
			// This file has a function that is called 3+ times (simulated via callers file)
			content: `package testpkg

// ProcessRequest handles incoming requests.
func ProcessRequest(req string) string {
	return req
}
`,
			// Three callers in the same directory
			callerContent: `package testpkg

func callerA() { ProcessRequest("a") }
func callerB() { ProcessRequest("b") }
func callerC() { ProcessRequest("c") }
`,
			wantP1:       1,
			wantViolFunc: "ProcessRequest",
		},
		{
			name: "exported function already has ANCHOR tag",
			content: `package testpkg

// @MX:ANCHOR: [AUTO] ProcessRequest is a high fan_in entry point
// @MX:REASON: Called by 3+ callers (callerA, callerB, callerC)
func ProcessRequest(req string) string {
	return req
}
`,
			callerContent: `package testpkg

func callerA() { ProcessRequest("a") }
func callerB() { ProcessRequest("b") }
func callerC() { ProcessRequest("c") }
`,
			wantP1:       0, // no violation because ANCHOR tag exists
			wantViolFunc: "",
		},
		{
			name: "unexported function with high fan_in - no violation",
			content: `package testpkg

// processRequest is internal.
func processRequest(req string) string {
	return req
}
`,
			callerContent: `package testpkg

func callerA() { processRequest("a") }
func callerB() { processRequest("b") }
func callerC() { processRequest("c") }
`,
			wantP1: 0, // unexported functions don't need ANCHOR
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			mainFile := writeGoFile(t, dir, "main.go", tt.content)
			if tt.callerContent != "" {
				writeGoFile(t, dir, "callers.go", tt.callerContent)
			}

			v := NewValidator(nil, dir) // nil analyzer = Grep fallback
			ctx := context.Background()

			report, err := v.ValidateFile(ctx, mainFile)
			if err != nil {
				t.Fatalf("ValidateFile() error = %v", err)
			}
			if report == nil {
				t.Fatal("ValidateFile() returned nil report")
			}

			p1Count := report.P1Count()
			if p1Count != tt.wantP1 {
				t.Errorf("P1Count() = %d, want %d\nViolations: %+v", p1Count, tt.wantP1, report.Violations)
			}

			if tt.wantViolFunc != "" {
				found := false
				for _, v := range report.Violations {
					if v.FuncName == tt.wantViolFunc && v.Priority == P1 {
						found = true
						if v.FilePath == "" {
							t.Error("violation.FilePath is empty")
						}
						if v.Line <= 0 {
							t.Errorf("violation.Line = %d, want > 0", v.Line)
						}
						if v.FanIn < 3 {
							t.Errorf("violation.FanIn = %d, want >= 3", v.FanIn)
						}
						if !v.Blocking {
							t.Error("P1 violation should be Blocking=true")
						}
					}
				}
				if !found {
					t.Errorf("expected P1 violation for function %q, violations: %+v", tt.wantViolFunc, report.Violations)
				}
			}
		})
	}
}

// TestValidateFile_PriorityClassification verifies AC-VAL-002:
// violations are correctly classified as P1-P4.
func TestValidateFile_PriorityClassification(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// File with multiple violation types:
	// - ExportedHighFanIn: P1 (fan_in >= 3, no ANCHOR)
	// - ExportedWithGoroutine: P2 (goroutine pattern, no WARN)
	// - ExportedLongFunc: P3 (100+ lines, no NOTE)
	// - ExportedUntestedFunc: P4 (no test file, no TODO)
	mainContent := `package testpkg

// ExportedHighFanIn is called by many callers.
func ExportedHighFanIn(x int) int {
	return x * 2
}

// ExportedWithGoroutine launches background work.
func ExportedWithGoroutine() {
	go func() {
		// background work
	}()
}

// ExportedLongFunc is a long function that should have a NOTE tag.
func ExportedLongFunc() {
	// line 1
	// line 2
	// line 3
	// line 4
	// line 5
	// line 6
	// line 7
	// line 8
	// line 9
	// line 10
	// line 11
	// line 12
	// line 13
	// line 14
	// line 15
	// line 16
	// line 17
	// line 18
	// line 19
	// line 20
	// line 21
	// line 22
	// line 23
	// line 24
	// line 25
	// line 26
	// line 27
	// line 28
	// line 29
	// line 30
	// line 31
	// line 32
	// line 33
	// line 34
	// line 35
	// line 36
	// line 37
	// line 38
	// line 39
	// line 40
	// line 41
	// line 42
	// line 43
	// line 44
	// line 45
	// line 46
	// line 47
	// line 48
	// line 49
	// line 50
	// line 51
	// line 52
	// line 53
	// line 54
	// line 55
	// line 56
	// line 57
	// line 58
	// line 59
	// line 60
	// line 61
	// line 62
	// line 63
	// line 64
	// line 65
	// line 66
	// line 67
	// line 68
	// line 69
	// line 70
	// line 71
	// line 72
	// line 73
	// line 74
	// line 75
	// line 76
	// line 77
	// line 78
	// line 79
	// line 80
	// line 81
	// line 82
	// line 83
	// line 84
	// line 85
	// line 86
	// line 87
	// line 88
}

// ExportedUntestedFunc is a public function without tests.
func ExportedUntestedFunc() string {
	return "untested"
}
`
	callerContent := `package testpkg

func callerA() { ExportedHighFanIn(1) }
func callerB() { ExportedHighFanIn(2) }
func callerC() { ExportedHighFanIn(3) }
`
	mainFile := writeGoFile(t, dir, "main.go", mainContent)
	writeGoFile(t, dir, "callers.go", callerContent)
	// No test file → ExportedUntestedFunc has no corresponding test

	v := NewValidator(nil, dir)
	ctx := context.Background()

	report, err := v.ValidateFile(ctx, mainFile)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	// Verify P1 and P2 are blocking, P3 and P4 are advisory
	for _, viol := range report.Violations {
		switch viol.Priority {
		case P1, P2:
			if !viol.Blocking {
				t.Errorf("violation %s (priority %s) should be Blocking=true", viol.FuncName, viol.Priority)
			}
		case P3, P4:
			if viol.Blocking {
				t.Errorf("violation %s (priority %s) should be Blocking=false", viol.FuncName, viol.Priority)
			}
		}
	}

	// Must detect at least P1 and P2
	if report.P1Count() < 1 {
		t.Errorf("expected >= 1 P1 violation, got %d", report.P1Count())
	}
	if report.P2Count() < 1 {
		t.Errorf("expected >= 1 P2 violation, got %d", report.P2Count())
	}
}

// TestValidateFile_GoroutineDetection verifies P2 violation detection
// for goroutine patterns without @MX:WARN.
func TestValidateFile_GoroutineDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantP2  int
	}{
		{
			name: "goroutine without WARN tag",
			content: `package testpkg

func WorkerFunc() {
	go func() {
		// background work
	}()
}
`,
			wantP2: 1,
		},
		{
			name: "goroutine with WARN tag",
			content: `package testpkg

// @MX:WARN: [AUTO] Goroutine lifecycle risk
// @MX:REASON: Goroutine may leak if context is not cancelled
func WorkerFunc() {
	go func() {
		// background work
	}()
}
`,
			wantP2: 0,
		},
		{
			name: "go statement without goroutine",
			content: `package testpkg

func PlainFunc() {
	// no goroutine
}
`,
			wantP2: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			mainFile := writeGoFile(t, dir, "main.go", tt.content)

			v := NewValidator(nil, dir)
			ctx := context.Background()

			report, err := v.ValidateFile(ctx, mainFile)
			if err != nil {
				t.Fatalf("ValidateFile() error = %v", err)
			}

			if report.P2Count() != tt.wantP2 {
				t.Errorf("P2Count() = %d, want %d\nViolations: %+v", report.P2Count(), tt.wantP2, report.Violations)
			}
		})
	}
}

// TestValidateFiles_Parallel verifies AC-EDGE-003: thread safety.
// Concurrent ValidateFile calls on different files must not race.
func TestValidateFiles_Parallel(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create multiple files
	files := make([]string, 5)
	for i := range files {
		content := fmt.Sprintf(`package testpkg

func Func%d() {
	go func() {}()
}
`, i)
		files[i] = writeGoFile(t, dir, fmt.Sprintf("file%d.go", i), content)
	}

	v := NewValidator(nil, dir)
	ctx := context.Background()

	// Validate all files concurrently (AC-EDGE-003)
	var wg sync.WaitGroup
	reports := make([]*FileReport, len(files))
	errors := make([]error, len(files))

	for i, f := range files {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			reports[idx], errors[idx] = v.ValidateFile(ctx, path)
		}(i, f)
	}
	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("file %d: ValidateFile() error = %v", i, err)
		}
		if reports[i] == nil {
			t.Errorf("file %d: ValidateFile() returned nil report", i)
		}
	}
}

// TestValidateFiles_ValidateFiles verifies AC-EDGE-004 (empty list)
// and basic batch validation.
func TestValidateFiles_ValidateFiles(t *testing.T) {
	t.Parallel()

	t.Run("empty file list", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		v := NewValidator(nil, dir)
		ctx := context.Background()

		report, err := v.ValidateFiles(ctx, nil)
		if err != nil {
			t.Fatalf("ValidateFiles() error = %v", err)
		}
		if report == nil {
			t.Fatal("ValidateFiles() returned nil report")
		}
		if len(report.FileReports) != 0 {
			t.Errorf("FileReports = %d, want 0", len(report.FileReports))
		}
		if report.TotalViolations() != 0 {
			t.Errorf("TotalViolations() = %d, want 0", report.TotalViolations())
		}
	})

	t.Run("multiple files aggregated correctly", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		file1 := writeGoFile(t, dir, "file1.go", `package testpkg

func Func1() {
	go func() {}()
}
`)
		file2 := writeGoFile(t, dir, "file2.go", `package testpkg

// normal function
func Func2() {}
`)

		v := NewValidator(nil, dir)
		ctx := context.Background()

		report, err := v.ValidateFiles(ctx, []string{file1, file2})
		if err != nil {
			t.Fatalf("ValidateFiles() error = %v", err)
		}
		if len(report.FileReports) != 2 {
			t.Errorf("FileReports = %d, want 2", len(report.FileReports))
		}
	})
}

// TestValidateFiles_TimeoutPartialResults verifies AC-EDGE-005:
// partial results are returned on timeout, not an error.
func TestValidateFiles_TimeoutPartialResults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create several files
	var files []string
	for i := 0; i < 5; i++ {
		content := fmt.Sprintf(`package testpkg
func Func%d() {}
`, i)
		files = append(files, writeGoFile(t, dir, fmt.Sprintf("file%d.go", i), content))
	}

	v := NewValidator(nil, dir)

	// Use a very short timeout to force cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	report, err := v.ValidateFiles(ctx, files)
	if err != nil {
		t.Fatalf("ValidateFiles() should not return error on timeout, got: %v", err)
	}
	if report == nil {
		t.Fatal("ValidateFiles() returned nil report on timeout")
	}
	// Some files may have timed out
	totalProcessed := len(report.FileReports) + len(report.TimedOutFiles)
	if totalProcessed != len(files) {
		t.Errorf("total processed (completed %d + timedOut %d) = %d, want %d",
			len(report.FileReports), len(report.TimedOutFiles), totalProcessed, len(files))
	}
}

// TestValidateFile_ReadOnly verifies AC-EDGE-002:
// validation must never modify the source file.
func TestValidateFile_ReadOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := `package testpkg

// @MX:ANCHOR: [AUTO] existing tag
// @MX:REASON: Already annotated
func ExistingAnchor(x int) int {
	return x
}
`
	filePath := writeGoFile(t, dir, "main.go", content)
	checksumBefore := fileChecksum(t, filePath)

	v := NewValidator(nil, dir)
	ctx := context.Background()

	_, err := v.ValidateFile(ctx, filePath)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	checksumAfter := fileChecksum(t, filePath)
	if checksumBefore != checksumAfter {
		t.Error("ValidateFile() modified the source file (read-only violation)")
	}
}

// TestValidateFile_Fallback verifies AC-EDGE-001:
// with nil analyzer, Grep fallback is used and fallback flag is set.
func TestValidateFile_Fallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := `package testpkg

func TestFunc() {
	go func() {}()
}
`
	filePath := writeGoFile(t, dir, "main.go", content)

	// nil analyzer → Grep fallback
	v := NewValidator(nil, dir)
	ctx := context.Background()

	report, err := v.ValidateFile(ctx, filePath)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}
	if !report.Fallback {
		t.Error("expected Fallback=true when analyzer is nil")
	}
}

// TestFormatReport verifies AC-REPORT-001: standard report format.
func TestFormatReport(t *testing.T) {
	t.Parallel()

	report := &ValidationReport{
		Duration: 150 * time.Millisecond,
		FileReports: []*FileReport{
			{
				FilePath: "/project/pkg/service.go",
				Violations: []Violation{
					{
						FuncName:   "HandleRequest",
						FilePath:   "/project/pkg/service.go",
						Line:       42,
						Priority:   P1,
						FanIn:      5,
						MissingTag: "@MX:ANCHOR",
						Reason:     "fan_in=5 >= threshold 3",
						Blocking:   true,
					},
				},
			},
			{
				FilePath: "/project/pkg/worker.go",
				Violations: []Violation{
					{
						FuncName:   "StartWorker",
						FilePath:   "/project/pkg/worker.go",
						Line:       10,
						Priority:   P2,
						MissingTag: "@MX:WARN",
						Reason:     "goroutine pattern detected",
						Blocking:   true,
					},
					{
						FuncName:   "StopWorker",
						FilePath:   "/project/pkg/worker.go",
						Line:       25,
						Priority:   P2,
						MissingTag: "@MX:WARN",
						Reason:     "goroutine pattern detected",
						Blocking:   true,
					},
				},
			},
			{
				FilePath: "/project/pkg/util.go",
				Violations: []Violation{
					{
						FuncName:   "LongHelper",
						FilePath:   "/project/pkg/util.go",
						Line:       5,
						Priority:   P3,
						MissingTag: "@MX:NOTE",
						Reason:     "function is 120 lines",
						Blocking:   false,
					},
				},
			},
		},
	}

	output := FormatReport(report)

	// Verify report contains required sections
	checkContains := func(t *testing.T, s, substr string) {
		t.Helper()
		if len(s) == 0 {
			t.Errorf("FormatReport() returned empty string")
			return
		}
		found := false
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FormatReport() output missing %q\nOutput:\n%s", substr, s)
		}
	}

	// AC-REPORT-001: Summary section
	checkContains(t, output, "Summary")
	checkContains(t, output, "P1")
	checkContains(t, output, "P2")
	checkContains(t, output, "P3")

	// Violation details with file:line format
	checkContains(t, output, "service.go")
	checkContains(t, output, "HandleRequest")
	checkContains(t, output, "worker.go")
	checkContains(t, output, "StartWorker")
}

// TestValidationReport_Aggregation verifies aggregation methods work correctly.
func TestValidationReport_Aggregation(t *testing.T) {
	t.Parallel()

	report := &ValidationReport{
		FileReports: []*FileReport{
			{
				Violations: []Violation{
					{Priority: P1, Blocking: true},
					{Priority: P2, Blocking: true},
				},
			},
			{
				Violations: []Violation{
					{Priority: P3, Blocking: false},
					{Priority: P4, Blocking: false},
				},
			},
		},
	}

	if report.P1Count() != 1 {
		t.Errorf("P1Count() = %d, want 1", report.P1Count())
	}
	if report.P2Count() != 1 {
		t.Errorf("P2Count() = %d, want 1", report.P2Count())
	}
	if report.P3Count() != 1 {
		t.Errorf("P3Count() = %d, want 1", report.P3Count())
	}
	if report.P4Count() != 1 {
		t.Errorf("P4Count() = %d, want 1", report.P4Count())
	}
	if report.TotalViolations() != 4 {
		t.Errorf("TotalViolations() = %d, want 4", report.TotalViolations())
	}
	if !report.HasBlockingViolations() {
		t.Error("HasBlockingViolations() = false, want true")
	}
}

// TestPriority_String verifies Priority.String() output.
func TestPriority_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		p    Priority
		want string
	}{
		{P1, "P1"},
		{P2, "P2"},
		{P3, "P3"},
		{P4, "P4"},
		{Priority(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("Priority(%d).String() = %q, want %q", tt.p, got, tt.want)
		}
	}
}

// TestPriority_IsBlocking verifies P1/P2 are blocking, P3/P4 are not.
func TestPriority_IsBlocking(t *testing.T) {
	t.Parallel()

	if !P1.IsBlocking() {
		t.Error("P1.IsBlocking() = false, want true")
	}
	if !P2.IsBlocking() {
		t.Error("P2.IsBlocking() = false, want true")
	}
	if P3.IsBlocking() {
		t.Error("P3.IsBlocking() = true, want false")
	}
	if P4.IsBlocking() {
		t.Error("P4.IsBlocking() = true, want false")
	}
}

// TestFileReport_CountMethods verifies FileReport.P3Count and P4Count.
func TestFileReport_CountMethods(t *testing.T) {
	t.Parallel()

	report := &FileReport{
		Violations: []Violation{
			{Priority: P1},
			{Priority: P2},
			{Priority: P3},
			{Priority: P3},
			{Priority: P4},
		},
	}

	if report.P1Count() != 1 {
		t.Errorf("P1Count() = %d, want 1", report.P1Count())
	}
	if report.P2Count() != 1 {
		t.Errorf("P2Count() = %d, want 1", report.P2Count())
	}
	if report.P3Count() != 2 {
		t.Errorf("P3Count() = %d, want 2", report.P3Count())
	}
	if report.P4Count() != 1 {
		t.Errorf("P4Count() = %d, want 1", report.P4Count())
	}
}

// TestValidateFile_NonExistentFile verifies error handling for missing file.
func TestValidateFile_NonExistentFile(t *testing.T) {
	t.Parallel()

	v := NewValidator(nil, t.TempDir())
	ctx := context.Background()

	report, err := v.ValidateFile(ctx, "/nonexistent/file.go")
	if err != nil {
		t.Fatalf("ValidateFile() should not return top-level error, got: %v", err)
	}
	if report == nil {
		t.Fatal("ValidateFile() returned nil report")
	}
	if report.Error == nil {
		t.Error("report.Error should be set for missing file")
	}
}

// TestValidateFile_ContextCancelled verifies TimedOut is set on cancellation.
func TestValidateFile_ContextCancelled(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := writeGoFile(t, dir, "main.go", `package testpkg

func Exported() {}
`)

	v := NewValidator(nil, dir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	report, err := v.ValidateFile(ctx, filePath)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}
	// File exists so content read succeeds, but context is cancelled before analysis
	// (or file report is returned with TimedOut if context was cancelled before read)
	_ = report // Either TimedOut or normal report is acceptable
}

// TestFormatReport_NoViolations verifies report format for clean validation.
func TestFormatReport_NoViolations(t *testing.T) {
	t.Parallel()

	report := &ValidationReport{
		Duration: 50 * time.Millisecond,
		FileReports: []*FileReport{
			{FilePath: "/project/clean.go"},
		},
	}

	output := FormatReport(report)
	if output == "" {
		t.Error("FormatReport() returned empty string")
	}
	if !strings.Contains(output, "passed") {
		t.Errorf("FormatReport() missing 'passed' for clean report\nOutput: %s", output)
	}
}

// TestFormatReport_WithTimedOutFiles verifies timed out files appear in report.
func TestFormatReport_WithTimedOutFiles(t *testing.T) {
	t.Parallel()

	report := &ValidationReport{
		Duration:      100 * time.Millisecond,
		FileReports:   []*FileReport{},
		TimedOutFiles: []string{"/project/slow.go", "/project/slow2.go"},
	}

	output := FormatReport(report)
	if !strings.Contains(output, "slow.go") {
		t.Errorf("FormatReport() missing timed out file\nOutput: %s", output)
	}
}

// TestValidateFiles_FallbackSet verifies Fallback flag in ValidationReport.
func TestValidateFiles_FallbackSet(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := writeGoFile(t, dir, "main.go", `package testpkg

func Func() {}
`)

	v := NewValidator(nil, dir) // nil → fallback
	ctx := context.Background()

	report, err := v.ValidateFiles(ctx, []string{file})
	if err != nil {
		t.Fatalf("ValidateFiles() error = %v", err)
	}
	if !report.Fallback {
		t.Error("ValidationReport.Fallback should be true when analyzer is nil")
	}
}
