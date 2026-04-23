package mx

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/mx/complexity"
)

// exportedFuncRe matches exported top-level Go function declarations.
// Group 1: function name (starts with uppercase).
var exportedFuncRe = regexp.MustCompile(`^func\s+([A-Z]\w+)`)

// exportedMethodRe matches exported Go method receiver declarations.
// Group 1: method name (starts with uppercase).
// Matches: func (r *T) Method(...) and func (v T) Method(...)
var exportedMethodRe = regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+([A-Z]\w+)`)

// goroutineRe matches goroutine launch patterns.
var goroutineRe = regexp.MustCompile(`\bgo\s+func\s*\(`)

// defaultExcludeGlobs are the default file/directory patterns excluded from fan-in counting.
// These mirror the mx.yaml default exclude list.
var defaultExcludeGlobs = []string{
	"vendor",
	"*_generated.go",
	"mock_*.go",
}

// mxValidator implements the Validator interface using Go-native file analysis.
// It is thread-safe: ValidateFile can be called concurrently.
// The projectRoot is used as the search base for fan_in counting.
type mxValidator struct {
	// analyzer is an optional AST-grep analyzer (unused in current implementation).
	analyzer       any
	projectRoot    string
	fanInThreshold int
	transitionMode bool

	// testOnWorkerStart is an optional hook called when a ValidateFiles worker goroutine
	// begins body execution. Used in tests to measure observed concurrency.
	testOnWorkerStart func()
	// testOnWorkerDone is an optional hook called when a ValidateFiles worker goroutine
	// exits body execution.
	testOnWorkerDone func()
}

// @MX:ANCHOR: [AUTO] MX tag validator factory. Entry point for the ValidateFile/ValidateFiles call chain.
// @MX:REASON: fan_in=8, called frequently from mx hook handlers and CLI, interface changes have wide propagation scope
// NewValidator creates a new MX Validator with default configuration.
// analyzer may be nil (Go-native fallback is always used).
// projectRoot is the base directory for fan_in reference counting.
func NewValidator(analyzer any, projectRoot string) Validator {
	return newValidatorWithConfig(analyzer, projectRoot, DefaultValidationConfig())
}

// newValidatorWithConfig creates a new MX Validator with the given configuration.
// This is an unexported factory used by tests to inject transition_mode and other settings.
func newValidatorWithConfig(analyzer any, projectRoot string, cfg *ValidationConfig) Validator {
	tm := false
	if cfg != nil {
		tm = cfg.TransitionMode
	}
	return &mxValidator{
		analyzer:       analyzer,
		projectRoot:    projectRoot,
		fanInThreshold: 3,
		transitionMode: tm,
	}
}

// ValidateFile validates a single Go source file for missing @MX tags.
// It is thread-safe and respects context cancellation.
func (v *mxValidator) ValidateFile(ctx context.Context, filePath string) (*FileReport, error) {
	start := time.Now()

	report := &FileReport{
		FilePath: filePath,
		Fallback: v.analyzer == nil, // always true for now (Go-native fallback)
	}

	// Read file content.
	data, err := os.ReadFile(filePath)
	if err != nil {
		report.Error = fmt.Errorf("read file: %w", err)
		report.Duration = time.Since(start)
		return report, nil
	}

	// Check for context cancellation before expensive analysis.
	select {
	case <-ctx.Done():
		report.TimedOut = true
		report.Duration = time.Since(start)
		return report, nil
	default:
	}

	content := string(data)
	// Detect language from file extension for complexity analysis.
	lang := langFromExt(filePath)
	violations := v.analyzeFile(ctx, filePath, content, lang, data)
	report.Violations = violations
	report.Duration = time.Since(start)
	return report, nil
}

// @MX:WARN: [AUTO] Called in parallel via goroutines from ValidateFiles. Can be interrupted via ctx.Done(), but individual countFanIn calls may block on file I/O.
// @MX:REASON: parallel validation via goroutines, file I/O sections that do not stop immediately on context cancellation exist
// analyzeFile performs the core analysis logic for a single file.
// Returns all detected violations.
func (v *mxValidator) analyzeFile(ctx context.Context, filePath, content, lang string, rawContent []byte) []Violation {
	var violations []Violation

	lines := strings.Split(content, "\n")
	funcInfos := extractFunctions(lines)

	for _, fn := range funcInfos {
		// Check context cancellation between functions.
		select {
		case <-ctx.Done():
			return violations
		default:
		}

		// P1: exported function with fan_in >= threshold missing @MX:ANCHOR.
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

		// P1: @MX:ANCHOR present but @MX:REASON sub-line missing (D6 enforcement).
		if fn.hasAnchor && !fn.hasAnchorReason {
			violations = append(violations, Violation{
				FuncName:   fn.name,
				FilePath:   filePath,
				Line:       fn.line,
				Priority:   P1,
				MissingTag: "@MX:REASON",
				Reason:     "@MX:ANCHOR present but @MX:REASON sub-line missing within 1 line",
				Blocking:   !v.transitionMode,
			})
		}

		// P2: function containing goroutine pattern missing @MX:WARN.
		if fn.hasGoroutine && !fn.hasWarn {
			blocking := true
			if fn.fromMethodReceiver && v.transitionMode {
				blocking = false
			}
			violations = append(violations, Violation{
				FuncName:   fn.name,
				FilePath:   filePath,
				Line:       fn.line,
				Priority:   P2,
				MissingTag: "@MX:WARN",
				Reason:     "goroutine pattern detected",
				Blocking:   blocking,
			})
		}

		// P2: @MX:WARN present but @MX:REASON sub-line missing (D6 enforcement).
		if fn.hasWarn && !fn.hasWarnReason {
			violations = append(violations, Violation{
				FuncName:   fn.name,
				FilePath:   filePath,
				Line:       fn.line,
				Priority:   P2,
				MissingTag: "@MX:REASON",
				Reason:     "@MX:WARN present but @MX:REASON sub-line missing within 1 line",
				Blocking:   !v.transitionMode,
			})
		}

		// P2: high cyclomatic complexity missing @MX:WARN (D5 / REQ-UTIL-001-040,041).
		if !fn.hasWarn && rawContent != nil {
			cmplx, _ := complexity.Measure(lang, rawContent, fn.name, fn.line)
			if cmplx.Supported {
				if cmplx.Cyclomatic >= 15 {
					violations = append(violations, Violation{
						FuncName:   fn.name,
						FilePath:   filePath,
						Line:       fn.line,
						Priority:   P2,
						MissingTag: "@MX:WARN",
						Reason:     fmt.Sprintf("cyclomatic complexity %d >= 15", cmplx.Cyclomatic),
						Blocking:   true,
					})
				} else if cmplx.IfBranches >= 8 {
					violations = append(violations, Violation{
						FuncName:   fn.name,
						FilePath:   filePath,
						Line:       fn.line,
						Priority:   P2,
						MissingTag: "@MX:WARN",
						Reason:     fmt.Sprintf("if-branches %d >= 8", cmplx.IfBranches),
						Blocking:   true,
					})
				}
			}
		}

		// P3: exported function >= 100 lines missing @MX:NOTE.
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

		// P4: exported function with no corresponding test missing @MX:TODO.
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

// funcInfo holds extracted information about a Go function or method.
type funcInfo struct {
	name               string
	line               int // 1-indexed
	exported           bool
	fromMethodReceiver bool // true if this is a method receiver (not a top-level function)
	hasAnchor          bool
	hasAnchorReason    bool // @MX:REASON found within ±1 line of @MX:ANCHOR
	hasWarn            bool
	hasWarnReason      bool // @MX:REASON found within ±1 line of @MX:WARN
	hasNote            bool
	hasTodo            bool
	hasGoroutine       bool
	lineCount          int // body line count
}

// extractFunctions parses Go source lines and extracts function metadata.
// It handles both top-level functions and method receivers.
func extractFunctions(lines []string) []funcInfo {
	var funcs []funcInfo

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		var name string
		var isMethod bool

		// Try top-level function first, then method receiver.
		if m := exportedFuncRe.FindStringSubmatch(trimmed); len(m) >= 2 {
			name = m[1]
			isMethod = false
		} else if m := exportedMethodRe.FindStringSubmatch(trimmed); len(m) >= 2 {
			name = m[1]
			isMethod = true
		}

		if name == "" {
			i++
			continue
		}

		fn := funcInfo{
			name:               name,
			line:               i + 1, // 1-indexed
			exported:           true,
			fromMethodReceiver: isMethod,
		}

		// Scan preceding comment block for @MX tags (up to 10 lines back).
		// Track which lines have which tags for ±1 REASON pairing.
		type tagLine struct {
			lineIdx int
			tag     string
		}
		var anchorLines, warnLines []int

		for j := max(i-1, 0); j >= 0 && i-j <= 10; j-- {
			commentLine := strings.TrimSpace(lines[j])
			if !strings.HasPrefix(commentLine, "//") {
				break
			}
			if strings.Contains(commentLine, "@MX:ANCHOR") {
				fn.hasAnchor = true
				anchorLines = append(anchorLines, j)
			}
			if strings.Contains(commentLine, "@MX:WARN") {
				fn.hasWarn = true
				warnLines = append(warnLines, j)
			}
			if strings.Contains(commentLine, "@MX:NOTE") {
				fn.hasNote = true
			}
			if strings.Contains(commentLine, "@MX:TODO") {
				fn.hasTodo = true
			}
		}

		// Check for @MX:REASON within ±1 line of each ANCHOR/WARN line.
		if fn.hasAnchor {
			fn.hasAnchorReason = hasReasonNearLines(lines, anchorLines)
		}
		if fn.hasWarn {
			fn.hasWarnReason = hasReasonNearLines(lines, warnLines)
		}

		// Scan function body for goroutine patterns and count lines.
		bodyStart := i + 1
		depth := 0
		if strings.Contains(line, "{") {
			depth = 1
		}
		j := bodyStart
		for j < len(lines) && (depth > 0 || j == bodyStart) {
			bodyLine := lines[j]
			for _, ch := range bodyLine {
				switch ch {
				case '{':
					depth++
				case '}':
					depth--
				}
			}
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
	}

	return funcs
}

// hasReasonNearLines returns true if any comment line containing @MX:REASON exists
// within ±1 line of any of the given tag line indices.
func hasReasonNearLines(lines []string, tagLines []int) bool {
	for _, tagIdx := range tagLines {
		// Check tagIdx-1, tagIdx, tagIdx+1.
		for delta := -1; delta <= 1; delta++ {
			checkIdx := tagIdx + delta
			if checkIdx < 0 || checkIdx >= len(lines) {
				continue
			}
			commentLine := strings.TrimSpace(lines[checkIdx])
			if strings.HasPrefix(commentLine, "//") && strings.Contains(commentLine, "@MX:REASON") {
				return true
			}
		}
	}
	return false
}

// scanProjectForIdentifier walks projectRoot and counts word-boundary occurrences
// of funcName in each .go file, skipping excluded patterns.
// Returns a map of filePath → match count.
func scanProjectForIdentifier(ctx context.Context, projectRoot, funcName string) (map[string]int, error) {
	if projectRoot == "" {
		return nil, nil
	}

	// Compile word-boundary regex once.
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(funcName) + `\b`)

	results := make(map[string]int)

	err := filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		// Check context cancellation.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip excluded directories.
		if d.IsDir() {
			base := d.Name()
			if base == "vendor" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files.
		if filepath.Ext(path) != ".go" {
			return nil
		}

		// Skip excluded file patterns.
		base := filepath.Base(path)
		for _, glob := range defaultExcludeGlobs {
			if matched, _ := filepath.Match(glob, base); matched {
				return nil
			}
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		count := len(re.FindAllIndex(data, -1))
		if count > 0 {
			results[path] = count
		}
		return nil
	})

	if err != nil && err == ctx.Err() {
		return results, nil // return partial results on cancellation
	}
	return results, nil
}

// countFanIn counts the number of word-boundary references to funcName in the project.
// It uses a Go-native file walker with no external subprocess calls.
func (v *mxValidator) countFanIn(ctx context.Context, funcName, currentFile string) int {
	if v.projectRoot == "" {
		return 0
	}

	fileCounts, err := scanProjectForIdentifier(ctx, v.projectRoot, funcName)
	if err != nil {
		return 0
	}

	callerCount := 0
	for path, count := range fileCounts {
		if path == currentFile {
			// Subtract the declaration itself (one occurrence in the current file).
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

// @MX:WARN: [AUTO] Spawns bounded goroutine pool; semaphore caps concurrency at runtime.NumCPU()*2.
// @MX:REASON: goroutine lifecycle — semaphore of capacity NumCPU*2 limits max in-flight workers; bounded pool prevents goroutine exhaustion on large file sets
// ValidateFiles validates multiple Go source files in parallel.
// Returns partial results if context is cancelled (AC-EDGE-005).
// Never returns an error for timeout: partial results are returned instead.
func (v *mxValidator) ValidateFiles(ctx context.Context, filePaths []string) (*ValidationReport, error) {
	start := time.Now()

	if len(filePaths) == 0 {
		return &ValidationReport{Duration: time.Since(start)}, nil
	}

	resultsCh := make(chan validateResult, len(filePaths))

	// Semaphore bounds concurrent in-flight workers to runtime.NumCPU()*2 (REQ-UTIL-001-003).
	sem := make(chan struct{}, runtime.NumCPU()*2)

	var wg sync.WaitGroup
	for _, path := range filePaths {
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore BEFORE launching goroutine

		go func(fp string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore when done

			if v.testOnWorkerStart != nil {
				v.testOnWorkerStart()
			}
			if v.testOnWorkerDone != nil {
				defer v.testOnWorkerDone()
			}

			// Fast path: context already cancelled.
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

	// @MX:WARN: [AUTO] Goroutine that closes resultsCh after all workers finish; must not be cancelled externally
	// @MX:REASON: goroutine lifecycle — if parent function returns early this goroutine orphans until wg reaches zero, then close panics on already-drained channel
	// Close channel once all goroutines finish.
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect all results.
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

// langFromExt returns the moai language identifier for a file based on its extension.
// Returns an empty string for unrecognized extensions (complexity analysis skipped).
func langFromExt(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".kt", ".kts":
		return "kotlin"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".ex", ".exs":
		return "elixir"
	case ".cpp", ".cc", ".cxx", ".c++", ".h", ".hpp":
		return "cpp"
	case ".scala":
		return "scala"
	case ".r":
		return "r"
	case ".dart":
		return "flutter"
	case ".swift":
		return "swift"
	default:
		return ""
	}
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

	totalFiles := len(report.FileReports)
	timedOut := len(report.TimedOutFiles)

	fmt.Fprintf(&buf, "## MX Validation Report - Summary\n\n")
	fmt.Fprintf(&buf, "- Files validated: %d\n", totalFiles)
	if timedOut > 0 {
		fmt.Fprintf(&buf, "- Files timed out: %d\n", timedOut)
	}
	fmt.Fprintf(&buf, "- Duration: %s\n", report.Duration.Round(time.Millisecond))
	fmt.Fprintf(&buf, "- P1 violations: %d\n", report.P1Count())
	fmt.Fprintf(&buf, "- P2 violations: %d\n", report.P2Count())
	fmt.Fprintf(&buf, "- P3 violations: %d\n", report.P3Count())
	fmt.Fprintf(&buf, "- P4 violations: %d\n", report.P4Count())

	if report.TotalViolations() == 0 && timedOut == 0 {
		fmt.Fprintf(&buf, "\nAll files passed MX validation.\n")
		return buf.String()
	}

	// P1 Violations (blocking)
	p1s := collectByPriority(report.FileReports, P1)
	if len(p1s) > 0 {
		fmt.Fprintf(&buf, "\n### P1 Violations (Blocking - Missing @MX:ANCHOR)\n\n")
		for _, v := range p1s {
			fmt.Fprintf(&buf, "- `%s:%d` `%s` — %s\n",
				filepath.Base(v.FilePath), v.Line, v.FuncName, v.Reason)
		}
	}

	// P2 Violations (blocking)
	p2s := collectByPriority(report.FileReports, P2)
	if len(p2s) > 0 {
		fmt.Fprintf(&buf, "\n### P2 Violations (Blocking - Missing @MX:WARN)\n\n")
		for _, v := range p2s {
			fmt.Fprintf(&buf, "- `%s:%d` `%s` — %s\n",
				filepath.Base(v.FilePath), v.Line, v.FuncName, v.Reason)
		}
	}

	// P3 Violations (advisory)
	p3s := collectByPriority(report.FileReports, P3)
	if len(p3s) > 0 {
		fmt.Fprintf(&buf, "\n### P3 Violations (Advisory - Missing @MX:NOTE)\n\n")
		for _, v := range p3s {
			fmt.Fprintf(&buf, "- `%s:%d` `%s` — %s\n",
				filepath.Base(v.FilePath), v.Line, v.FuncName, v.Reason)
		}
	}

	// P4 Violations (advisory)
	p4s := collectByPriority(report.FileReports, P4)
	if len(p4s) > 0 {
		fmt.Fprintf(&buf, "\n### P4 Violations (Advisory - Missing @MX:TODO)\n\n")
		for _, v := range p4s {
			fmt.Fprintf(&buf, "- `%s:%d` `%s` — %s\n",
				filepath.Base(v.FilePath), v.Line, v.FuncName, v.Reason)
		}
	}

	if len(report.TimedOutFiles) > 0 {
		fmt.Fprintf(&buf, "\n### Timed Out Files\n\n")
		for _, f := range report.TimedOutFiles {
			fmt.Fprintf(&buf, "- %s\n", filepath.Base(f))
		}
	}

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
