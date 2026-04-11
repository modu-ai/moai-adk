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

	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// GoplsBridge는 gopls.Bridge의 GetDiagnostics 메서드를 노출하는 인터페이스다.
// *gopls.Bridge를 직접 의존하지 않아 테스트에서 mock으로 교체할 수 있다.
// GOPLS-BRIDGE-001: Phase 6 integration
type GoplsBridge interface {
	GetDiagnostics(ctx context.Context, path string) ([]gopls.Diagnostic, error)
}

// GoFeedbackGenerator collects feedback by running Go toolchain commands
// (go test, go vet) and parsing their output. It implements FeedbackGenerator.
// GOPLS-BRIDGE-001: bridge 필드가 nil이 아니면 LSP 진단도 수집한다.
type GoFeedbackGenerator struct {
	projectRoot string
	bridge      GoplsBridge // nil이면 LSP 진단 수집 비활성화
}

// NewGoFeedbackGenerator creates a FeedbackGenerator for Go projects.
// projectRoot is the directory where go test and go vet will be executed.
// 하위 호환성: bridge는 nil로 설정된다 (LSP 진단 비활성화).
func NewGoFeedbackGenerator(projectRoot string) FeedbackGenerator {
	return &GoFeedbackGenerator{projectRoot: projectRoot, bridge: nil}
}

// NewGoFeedbackGeneratorWithBridge creates a FeedbackGenerator for Go projects
// with an optional gopls bridge for LSP diagnostics.
// GOPLS-BRIDGE-001: bridge가 nil이면 기존 동작(go test + go vet)만 수행한다.
// bridge가 non-nil이면 LSP 진단도 Feedback.Diagnostics에 추가한다.
func NewGoFeedbackGeneratorWithBridge(projectRoot string, bridge GoplsBridge) FeedbackGenerator {
	return &GoFeedbackGenerator{projectRoot: projectRoot, bridge: bridge}
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

	// GOPLS-BRIDGE-001: bridge가 non-nil이면 LSP 진단을 수집한다.
	// GetDiagnostics 오류는 무시한다 — bridge 실패가 전체 피드백을 차단해선 안 된다.
	if g.bridge != nil {
		diags, err := g.bridge.GetDiagnostics(ctx, g.projectRoot)
		if err != nil {
			slog.Warn("gopls 진단 수집 실패, 건너뜀", "error", err)
		} else {
			fb.Diagnostics = diags
		}
	}

	return fb, nil
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
