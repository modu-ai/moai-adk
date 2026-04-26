package loop

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// GoplsBridge is an interface that exposes the GetDiagnostics method of gopls.Bridge.
// Does not directly depend on *gopls.Bridge, so it can be replaced with a mock in tests.
// GOPLS-BRIDGE-001: Phase 6 integration
type GoplsBridge interface {
	GetDiagnostics(ctx context.Context, path string) ([]gopls.Diagnostic, error)
}

// DiagnosticsAggregator abstracts the Aggregator facade (SPEC-LSP-AGG-003)
// so that GoFeedbackGenerator does not depend on the concrete aggregator type.
// REQ-LL-002: GoFeedbackGenerator wires Aggregator for Go files through this interface.
type DiagnosticsAggregator interface {
	GetDiagnostics(ctx context.Context, path string) ([]lsp.Diagnostic, error)
}

// GoFeedbackGenerator collects feedback by running Go toolchain commands
// (go test, go vet) and parsing their output. It implements FeedbackGenerator.
// GOPLS-BRIDGE-001: also collects LSP diagnostics when the bridge field is non-nil.
// REQ-LL-002: also collects Aggregator-based lsp.Diagnostic when the aggregator field is non-nil.
type GoFeedbackGenerator struct {
	projectRoot string
	bridge      GoplsBridge          // nil disables gopls diagnostic collection
	aggregator  DiagnosticsAggregator // nil disables Aggregator diagnostic collection (REQ-LL-002)
}

// NewGoFeedbackGenerator creates a FeedbackGenerator for Go projects.
// projectRoot is the directory where go test and go vet will be executed.
// Backward compatibility: bridge is set to nil (LSP diagnostic collection disabled).
func NewGoFeedbackGenerator(projectRoot string) FeedbackGenerator {
	return &GoFeedbackGenerator{projectRoot: projectRoot, bridge: nil}
}

// NewGoFeedbackGeneratorWithBridge creates a FeedbackGenerator for Go projects
// with an optional gopls bridge for LSP diagnostics.
// GOPLS-BRIDGE-001: when bridge is nil, only the existing behavior (go test + go vet) is performed.
// When bridge is non-nil, LSP diagnostics are also added to Feedback.Diagnostics.
func NewGoFeedbackGeneratorWithBridge(projectRoot string, bridge GoplsBridge) FeedbackGenerator {
	return &GoFeedbackGenerator{projectRoot: projectRoot, bridge: bridge}
}

// NewGoFeedbackGeneratorWithAggregator creates a FeedbackGenerator for Go projects
// with an optional Aggregator for LSP diagnostics via lsp.Diagnostic.
// REQ-LL-002: when aggregator is nil, only the existing behavior (go test + go vet) is performed.
// When aggregator is non-nil, lsp.Diagnostic is collected into Feedback.LSPDiagnostics.
func NewGoFeedbackGeneratorWithAggregator(projectRoot string, agg DiagnosticsAggregator) FeedbackGenerator {
	return &GoFeedbackGenerator{projectRoot: projectRoot, aggregator: agg}
}

// goTestEvent represents a single JSON event from `go test -json`.
type goTestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

// Collect runs go test and go vet, parsing results into a Feedback struct.
// The context controls timeout — callers should set appropriate deadlines.
func (g *GoFeedbackGenerator) Collect(ctx context.Context) (*Feedback, error) {
	start := time.Now()
	fb := &Feedback{
		Phase:        PhaseTest,
		BuildSuccess: true,
	}

	// Run go test with JSON output and coverage.
	coverFile := filepath.Join(g.projectRoot, ".moai", "state", "loop", "coverage.out")
	testCmd := exec.CommandContext(ctx, "go", "test", "-count=1", "-json",
		"-coverprofile="+coverFile, "./...")
	testCmd.Dir = g.projectRoot

	var testOut bytes.Buffer
	testCmd.Stdout = &testOut
	testCmd.Stderr = &bytes.Buffer{}

	testErr := testCmd.Run()
	if testErr != nil {
		// Build failure or test failure — differentiate by exit code.
		fb.BuildSuccess = false
	}

	// Parse test JSON output for pass/fail counts.
	passed, failed := parseGoTestJSON(testOut.Bytes())
	fb.TestsPassed = passed
	fb.TestsFailed = failed
	if failed == 0 && passed > 0 {
		fb.BuildSuccess = true
	}

	// Parse coverage from the profile file.
	fb.Coverage = parseCoverageFile(coverFile)

	// Run go vet for lint errors.
	vetCmd := exec.CommandContext(ctx, "go", "vet", "./...")
	vetCmd.Dir = g.projectRoot
	var vetStderr bytes.Buffer
	vetCmd.Stdout = &bytes.Buffer{}
	vetCmd.Stderr = &vetStderr

	_ = vetCmd.Run()
	fb.LintErrors = countNonEmptyLines(vetStderr.Bytes())

	fb.Duration = time.Since(start)

	// GOPLS-BRIDGE-001: when bridge is non-nil, collect gopls.Diagnostic.
	// Errors from GetDiagnostics are ignored — bridge failures must not block the entire feedback.
	if g.bridge != nil {
		diags, err := g.bridge.GetDiagnostics(ctx, g.projectRoot)
		if err != nil {
			slog.Warn("gopls diagnostic collection failed, skipping", "error", err)
		} else {
			fb.Diagnostics = diags
		}
	}

	// REQ-LL-002: when aggregator is non-nil, collect lsp.Diagnostic into LSPDiagnostics.
	// Aggregator errors are ignored — diagnostic failures must not block the entire feedback.
	if g.aggregator != nil {
		lspDiags, err := g.aggregator.GetDiagnostics(ctx, g.projectRoot)
		if err != nil {
			slog.Warn("aggregator diagnostic collection failed, skipping", "error", err)
		} else {
			// Filter to Go-only results: only include diagnostics for .go files.
			fb.LSPDiagnostics = filterGoOnlyDiagnostics(lspDiags)
		}
	}

	return fb, nil
}

// filterGoOnlyDiagnostics filters diagnostics to include only Go-language entries.
// REQ-LL-002: GoFeedbackGenerator is Go-specific; non-Go diagnostics are excluded.
// Since the aggregator is queried with the projectRoot path, all returned diagnostics
// are considered Go-relevant when the aggregator is the Go-specific Aggregator instance.
// This function is a no-op pass-through for now but provides an extension point.
func filterGoOnlyDiagnostics(diags []lsp.Diagnostic) []lsp.Diagnostic {
	if len(diags) == 0 {
		return nil
	}
	// All diagnostics from the Go-project-root query are Go-relevant.
	// Future: if multi-language aggregator is used, filter by file extension here.
	result := make([]lsp.Diagnostic, len(diags))
	copy(result, diags)
	return result
}

// parseGoTestJSON parses go test -json output and returns (passed, failed) counts.
func parseGoTestJSON(data []byte) (passed, failed int) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var ev goTestEvent
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}
		// Only count top-level test results (Test field non-empty, package-level events have empty Test).
		if ev.Test == "" {
			continue
		}
		switch ev.Action {
		case "pass":
			passed++
		case "fail":
			failed++
		}
	}
	return
}

// parseCoverageFile reads a Go coverage profile and returns the total coverage percentage.
func parseCoverageFile(path string) float64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var totalStatements, coveredStatements int
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		// Format: file:startLine.startCol,endLine.endCol numStatements count
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		// parts[1] = numStatements, parts[2] = count
		var stmts, count int
		if _, err := parseIntFromString(parts[1]); err == nil {
			stmts = mustParseInt(parts[1])
		}
		if _, err := parseIntFromString(parts[2]); err == nil {
			count = mustParseInt(parts[2])
		}
		totalStatements += stmts
		if count > 0 {
			coveredStatements += stmts
		}
	}

	if totalStatements == 0 {
		return 0
	}
	return float64(coveredStatements) / float64(totalStatements) * 100.0
}

// countNonEmptyLines counts the number of non-empty lines in byte data.
func countNonEmptyLines(data []byte) int {
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	return count
}

// parseIntFromString is a helper that validates a string is a valid integer.
func parseIntFromString(s string) (int, error) {
	return mustParseIntErr(s)
}

func mustParseInt(s string) int {
	v, _ := mustParseIntErr(s)
	return v
}

func mustParseIntErr(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &json.InvalidUnmarshalError{}
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// Compile-time interface compliance check.
var _ FeedbackGenerator = (*GoFeedbackGenerator)(nil)
