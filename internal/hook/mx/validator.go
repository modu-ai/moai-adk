package mx

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// exportedFuncRe matches exported Go function declarations.
// Group 1: function name (starts with uppercase).
var exportedFuncRe = regexp.MustCompile(`^func\s+([A-Z]\w+)`)

// goroutineRe matches goroutine launch patterns.
var goroutineRe = regexp.MustCompile(`\bgo\s+func\s*\(`)

// mxValidator implements the Validator interface using Grep-based analysis.
// It is thread-safe: ValidateFile can be called concurrently.
// The projectRoot is used as the search base for fan_in counting.
type mxValidator struct {
	// analyzer is an optional AST-grep analyzer.
	// If nil, Grep-based fallback is used.
	analyzer any
	projectRoot string
	fanInThreshold int
}

// NewValidator creates a new MX Validator.
// analyzer may be nil (Grep fallback will be used).
// projectRoot is the base directory for fan_in reference counting.
func NewValidator(analyzer any, projectRoot string) Validator {
	return &mxValidator{
		analyzer:       analyzer,
		projectRoot:    projectRoot,
		fanInThreshold: 3,
	}
}

// ValidateFile validates a single Go source file for missing @MX tags.
// It is thread-safe and respects context cancellation.
func (v *mxValidator) ValidateFile(ctx context.Context, filePath string) (*FileReport, error) {
	start := time.Now()

	report := &FileReport{
		FilePath: filePath,
		Fallback: v.analyzer == nil, // always true for now (Grep fallback)
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		report.Error = fmt.Errorf("read file: %w", err)
		report.Duration = time.Since(start)
		return report, nil
	}

	// Check for context cancellation before expensive analysis
	select {
	case <-ctx.Done():
		report.TimedOut = true
		report.Duration = time.Since(start)
		return report, nil
	default:
	}

	content := string(data)
	violations := v.analyzeFile(ctx, filePath, content)
	report.Violations = violations
	report.Duration = time.Since(start)
	return report, nil
}

// analyzeFile performs the core analysis logic for a single file.
// Returns all detected violations.
func (v *mxValidator) analyzeFile(ctx context.Context, filePath, content string) []Violation {
	var violations []Violation

	lines := strings.Split(content, "\n")
	funcInfos := extractFunctions(lines)

	for _, fn := range funcInfos {
		// Check context cancellation between functions
		select {
		case <-ctx.Done():
			return violations
		default:
		}

		// P1: exported function with fan_in >= threshold missing @MX:ANCHOR
		if fn.exported && !fn.hasAnchor {
			fanIn := v.countFanIn(ctx, fn.name, filePath)
			if fanIn >= v.fanInThreshold {
				violations = append(violations, Violation{
					FuncName:   fn.name,
					FilePath:   filePath,
					Line:       fn.line,
					Priority:   P1,
					FanIn:      fanIn,
					MissingTag: "@MX:ANCHOR",
					Reason:     fmt.Sprintf("fan_in=%d >= threshold %d", fanIn, v.fanInThreshold),
					Blocking:   true,
				})
			}
		}

		// P2: function containing goroutine pattern missing @MX:WARN
		if fn.hasGoroutine && !fn.hasWarn {
			violations = append(violations, Violation{
				FuncName:   fn.name,
				FilePath:   filePath,
				Line:       fn.line,
				Priority:   P2,
				MissingTag: "@MX:WARN",
				Reason:     "goroutine pattern detected",
				Blocking:   true,
			})
		}

		// P3: exported function >= 100 lines missing @MX:NOTE
		if fn.exported && fn.lineCount >= 100 && !fn.hasNote {
			violations = append(violations, Violation{
				FuncName:   fn.name,
				FilePath:   filePath,
				Line:       fn.line,
				Priority:   P3,
				MissingTag: "@MX:NOTE",
				Reason:     fmt.Sprintf("function is %d lines", fn.lineCount),
				Blocking:   false,
			})
		}

		// P4: exported function with no corresponding test missing @MX:TODO
		if fn.exported && !fn.hasTodo {
			testFile := testFileFor(filePath)
			if !fileExists(testFile) {
				violations = append(violations, Violation{
					FuncName:   fn.name,
					FilePath:   filePath,
					Line:       fn.line,
					Priority:   P4,
					MissingTag: "@MX:TODO",
					Reason:     "no test file found",
					Blocking:   false,
				})
			}
		}
	}

	return violations
}

// funcInfo holds extracted information about a Go function.
type funcInfo struct {
	name         string
	line         int // 1-indexed
	exported     bool
	hasAnchor    bool
	hasWarn      bool
	hasNote      bool
	hasTodo      bool
	hasGoroutine bool
	lineCount    int // body line count
}

// extractFunctions parses Go source lines and extracts function metadata.
// It scans for exported function declarations and their preceding MX tags.
func extractFunctions(lines []string) []funcInfo {
	var funcs []funcInfo

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Check if this line is a function declaration
		if m := exportedFuncRe.FindStringSubmatch(strings.TrimSpace(line)); len(m) >= 2 {
			fn := funcInfo{
				name:     m[1],
				line:     i + 1, // 1-indexed
				exported: true,
			}

			// Scan preceding comment block for @MX tags (up to 10 lines back)
			for j := max(i-1, 0); j >= 0 && i-j <= 10; j-- {
				commentLine := strings.TrimSpace(lines[j])
				if !strings.HasPrefix(commentLine, "//") {
					break
				}
				if strings.Contains(commentLine, "@MX:ANCHOR") {
					fn.hasAnchor = true
				}
				if strings.Contains(commentLine, "@MX:WARN") {
					fn.hasWarn = true
				}
				if strings.Contains(commentLine, "@MX:NOTE") {
					fn.hasNote = true
				}
				if strings.Contains(commentLine, "@MX:TODO") {
					fn.hasTodo = true
				}
			}

			// Scan function body for goroutine patterns and count lines
			bodyStart := i + 1
			depth := 0
			// Find opening brace on function declaration line or next line
			if strings.Contains(line, "{") {
				depth = 1
			}
			j := bodyStart
			for j < len(lines) && (depth > 0 || j == bodyStart) {
				bodyLine := lines[j]
				// Count braces
				for _, ch := range bodyLine {
					switch ch {
					case '{':
						depth++
					case '}':
						depth--
					}
				}
				// Check for goroutine patterns
				if goroutineRe.MatchString(bodyLine) || strings.Contains(bodyLine, "\tgo ") || strings.Contains(bodyLine, " go func") {
					fn.hasGoroutine = true
				}
				if depth <= 0 {
					break
				}
				j++
			}
			fn.lineCount = j - i

			funcs = append(funcs, fn)
			i++
			continue
		}

		// Also detect goroutine patterns in unexported functions
		// (to check if the function containing the goroutine needs @MX:WARN)
		if strings.Contains(line, "func ") && !exportedFuncRe.MatchString(strings.TrimSpace(line)) {
			// unexported function - skip for now (P1/P2/P3/P4 are for exported only for P1/P3/P4)
			// But P2 checks all functions with goroutines that are exported
			// We already handle this above for exported functions
		}

		i++
	}

	return funcs
}

// countFanIn counts the number of references to funcName in the project directory.
// It uses grep to search for the function name and subtracts 1 for the declaration.
func (v *mxValidator) countFanIn(ctx context.Context, funcName, currentFile string) int {
	if v.projectRoot == "" {
		return 0
	}

	// Use grep to count references to the function name in .go files
	cmd := exec.CommandContext(ctx, "grep", "-r", "--include=*.go", "-l", funcName, v.projectRoot)
	out, err := cmd.Output()
	if err != nil {
		// grep returns exit 1 when no matches found
		return 0
	}

	// Count files that reference the function
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	callerCount := 0
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		// Count the number of times funcName appears in each file
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		count := strings.Count(string(data), funcName)
		if f == currentFile {
			// Subtract 1 for the declaration itself
			count--
		}
		callerCount += count
	}

	return callerCount
}

// validateResult holds the outcome of a single file validation goroutine.
type validateResult struct {
	report   *FileReport
	filePath string
	timedOut bool
}

// ValidateFiles validates multiple Go source files in parallel.
// Returns partial results if context is cancelled (AC-EDGE-005).
// Never returns an error for timeout: partial results are returned instead.
func (v *mxValidator) ValidateFiles(ctx context.Context, filePaths []string) (*ValidationReport, error) {
	start := time.Now()

	if len(filePaths) == 0 {
		return &ValidationReport{Duration: time.Since(start)}, nil
	}

	resultsCh := make(chan validateResult, len(filePaths))

	var wg sync.WaitGroup
	for _, path := range filePaths {
		wg.Add(1)
		go func(fp string) {
			defer wg.Done()

			// Fast path: context already cancelled
			select {
			case <-ctx.Done():
				resultsCh <- validateResult{filePath: fp, timedOut: true}
				return
			default:
			}

			report, err := v.ValidateFile(ctx, fp)
			if err != nil {
				slog.Warn("mx: ValidateFile error", "file", fp, "error", err)
				resultsCh <- validateResult{
					report: &FileReport{FilePath: fp, Error: err},
				}
				return
			}

			if report.TimedOut {
				resultsCh <- validateResult{filePath: fp, timedOut: true}
				return
			}

			resultsCh <- validateResult{report: report}
		}(path)
	}

	// Close channel once all goroutines finish
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect all results
	var fileReports []*FileReport
	var timedOutFiles []string
	usedFallback := false

	for r := range resultsCh {
		if r.timedOut {
			timedOutFiles = append(timedOutFiles, r.filePath)
			continue
		}
		if r.report != nil {
			fileReports = append(fileReports, r.report)
			if r.report.Fallback {
				usedFallback = true
			}
		}
	}

	return &ValidationReport{
		FileReports:   fileReports,
		TimedOutFiles: timedOutFiles,
		Duration:      time.Since(start),
		Fallback:      usedFallback,
	}, nil
}

// testFileFor returns the expected test file path for a given Go source file.
// e.g., "/pkg/service.go" → "/pkg/service_test.go"
func testFileFor(filePath string) string {
	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filePath, ext)
	return base + "_test" + ext
}

// fileExists returns true if the file exists on disk.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// formatReport formats a ValidationReport as a human-readable string.
func formatReport(report *ValidationReport) string {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	totalFiles := len(report.FileReports)
	timedOut := len(report.TimedOutFiles)

	fmt.Fprintf(w, "## MX Validation Report - Summary\n\n")
	fmt.Fprintf(w, "- Files validated: %d\n", totalFiles)
	if timedOut > 0 {
		fmt.Fprintf(w, "- Files timed out: %d\n", timedOut)
	}
	fmt.Fprintf(w, "- Duration: %s\n", report.Duration.Round(time.Millisecond))
	fmt.Fprintf(w, "- P1 violations: %d\n", report.P1Count())
	fmt.Fprintf(w, "- P2 violations: %d\n", report.P2Count())
	fmt.Fprintf(w, "- P3 violations: %d\n", report.P3Count())
	fmt.Fprintf(w, "- P4 violations: %d\n", report.P4Count())

	if report.TotalViolations() == 0 && timedOut == 0 {
		fmt.Fprintf(w, "\nAll files passed MX validation.\n")
		_ = w.Flush()
		return buf.String()
	}

	// P1 Violations (blocking)
	p1s := collectByPriority(report.FileReports, P1)
	if len(p1s) > 0 {
		fmt.Fprintf(w, "\n### P1 Violations (Blocking - Missing @MX:ANCHOR)\n\n")
		for _, v := range p1s {
			fmt.Fprintf(w, "- `%s:%d` `%s` — %s\n",
				filepath.Base(v.FilePath), v.Line, v.FuncName, v.Reason)
		}
	}

	// P2 Violations (blocking)
	p2s := collectByPriority(report.FileReports, P2)
	if len(p2s) > 0 {
		fmt.Fprintf(w, "\n### P2 Violations (Blocking - Missing @MX:WARN)\n\n")
		for _, v := range p2s {
			fmt.Fprintf(w, "- `%s:%d` `%s` — %s\n",
				filepath.Base(v.FilePath), v.Line, v.FuncName, v.Reason)
		}
	}

	// P3 Violations (advisory)
	p3s := collectByPriority(report.FileReports, P3)
	if len(p3s) > 0 {
		fmt.Fprintf(w, "\n### P3 Violations (Advisory - Missing @MX:NOTE)\n\n")
		for _, v := range p3s {
			fmt.Fprintf(w, "- `%s:%d` `%s` — %s\n",
				filepath.Base(v.FilePath), v.Line, v.FuncName, v.Reason)
		}
	}

	// P4 Violations (advisory)
	p4s := collectByPriority(report.FileReports, P4)
	if len(p4s) > 0 {
		fmt.Fprintf(w, "\n### P4 Violations (Advisory - Missing @MX:TODO)\n\n")
		for _, v := range p4s {
			fmt.Fprintf(w, "- `%s:%d` `%s` — %s\n",
				filepath.Base(v.FilePath), v.Line, v.FuncName, v.Reason)
		}
	}

	if len(report.TimedOutFiles) > 0 {
		fmt.Fprintf(w, "\n### Timed Out Files\n\n")
		for _, f := range report.TimedOutFiles {
			fmt.Fprintf(w, "- %s\n", filepath.Base(f))
		}
	}

	_ = w.Flush()
	return buf.String()
}

// collectByPriority collects all violations of a given priority from file reports.
func collectByPriority(fileReports []*FileReport, p Priority) []Violation {
	var result []Violation
	for _, fr := range fileReports {
		for _, v := range fr.Violations {
			if v.Priority == p {
				result = append(result, v)
			}
		}
	}
	return result
}
