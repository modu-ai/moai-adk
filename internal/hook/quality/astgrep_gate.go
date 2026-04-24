package quality

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// RunAstGrepGateV2는 통합 Scanner를 사용하여 ast-grep 품질 게이트를 실행합니다.
// REQ-ASTG-UPG-030: quality gate hook이 통합 Scanner를 호출
// REQ-UTIL-002-012: 억제 정책 위반이 있으면 (false, formatted-list)를 반환합니다.
// RunAstGrepGate의 대체 구현으로, 점진적 마이그레이션을 지원합니다.
func RunAstGrepGateV2(ctx context.Context, projectDir string, cfg *AstGrepGateConfig) (bool, string) {
	if cfg == nil || !cfg.Enabled {
		return true, ""
	}

	// ── 1. 억제 정책 검사 (sg 독립적, pure-Go) ─────────────────────────────
	// REQ-UTIL-002-010/011/012: ast-grep-ignore 주석의 @MX:REASON 쌍 검사
	sourceFiles := walkSourceFiles(projectDir)
	var allViolations []SuppressionViolation
	for _, fp := range sourceFiles {
		allViolations = append(allViolations, checkSuppressionPairing(fp)...)
	}

	if len(allViolations) > 0 {
		var sb strings.Builder
		sb.WriteString("ast-grep suppression policy violations:\n\n")
		for _, v := range allViolations {
			fmt.Fprintf(&sb, "[%s] %s\n", v.Type, v.Message)
		}
		return false, strings.TrimSpace(sb.String())
	}

	// ── 2. ast-grep 스캔 (sg CLI 의존) ───────────────────────────────────────
	rulesDir := filepath.Join(projectDir, cfg.RulesDir)
	scannerCfg := &astgrep.ScannerConfig{
		RulesDir:     rulesDir,
		SGBinary:     "sg",
		WarnOnlyMode: cfg.WarnOnlyMode,
		Timeout:      astGrepScanTimeout,
	}

	scanner := astgrep.NewScanner(scannerCfg)
	findings, err := scanner.Scan(ctx, projectDir)
	if err != nil {
		// 스캔 오류 시 통과 (graceful degradation)
		return true, ""
	}

	if len(findings) == 0 {
		return true, ""
	}

	// 결과 포매팅
	var sb strings.Builder
	sb.WriteString("ast-grep domain rule scan results:\n\n")
	for _, f := range findings {
		sb.WriteString(f.String())
		sb.WriteString("\n")
	}
	output := strings.TrimSpace(sb.String())

	// error severity 발견 시 차단 (WarnOnlyMode가 아닌 경우)
	if astgrep.HasErrors(findings) && !cfg.WarnOnlyMode && cfg.BlockOnError {
		return false, fmt.Sprintf("quality gate failed: ast-grep domain rules\n\n%s", output)
	}

	return true, output
}

// AstGrepGateConfig holds configuration for ast-grep quality gate scanning.
type AstGrepGateConfig struct {
	// Enabled controls whether ast-grep scanning is performed.
	Enabled bool
	// RulesDir is the directory containing domain-specific ast-grep rule files.
	// Default: ".moai/config/astgrep-rules"
	RulesDir string
	// BlockOnError causes the gate to block a commit when error-severity matches are found.
	BlockOnError bool
	// WarnOnlyMode prevents blocking even when error-severity matches are found.
	// When true, matches are reported as warnings but never block.
	WarnOnlyMode bool
}

// DefaultAstGrepGateConfig returns default configuration.
func DefaultAstGrepGateConfig() *AstGrepGateConfig {
	return &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: true,
		WarnOnlyMode: false,
	}
}

// astGrepScanMatch represents a single match entry from sg scan --json output.
type astGrepScanMatch struct {
	File     string `json:"file"`
	Lines    string `json:"lines,omitempty"`
	Text     string `json:"text,omitempty"`
	RuleID   string `json:"ruleId,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`
	Range    struct {
		Start struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"start"`
	} `json:"range"`
}

const astGrepScanTimeout = 30 * time.Second

// SuppressionViolation은 ast-grep-ignore 주석이 @MX:REASON과 쌍을 이루지 않는 경우를 나타냅니다.
// REQ-UTIL-002-010/011
type SuppressionViolation struct {
	// Type은 항상 "SUPPRESSION_WITHOUT_REASON"입니다.
	Type    string `json:"type"`
	// File은 위반이 발생한 파일 경로입니다.
	File    string `json:"file"`
	// Line은 1-indexed 줄 번호입니다.
	Line    int    `json:"line"`
	// Message는 사람이 읽을 수 있는 오류 메시지입니다.
	Message string `json:"message"`
}

// commentPrefix는 파일 확장자에 따라 line comment 접두사를 반환합니다.
// 지원하지 않는 확장자는 빈 문자열을 반환합니다.
func commentPrefix(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".rb", ".py", ".pyi", ".ex", ".exs":
		return "#"
	case ".hs":
		return "--"
	case ".go", ".java", ".ts", ".tsx", ".js", ".jsx", ".rs", ".swift",
		".kt", ".kts", ".cs", ".scala", ".cpp", ".cc", ".cxx", ".hpp",
		".c", ".h", ".dart", ".zig":
		return "//"
	default:
		return ""
	}
}

// checkSuppressionPairing는 파일을 한 줄씩 검사하여
// ast-grep-ignore 주석이 인접한 @MX:REASON 주석과 쌍을 이루지 않는 경우를 찾습니다.
//
// 허용 거리: ast-grep-ignore 다음 빈 줄 0~1개 후 @MX:REASON (REQ-UTIL-002-010).
// 허용 안됨: 비어 있지 않은 다른 주석 또는 코드 행이 사이에 있는 경우 (REQ-UTIL-002-011).
//
// REQ-UTIL-002-010/011
func checkSuppressionPairing(filePath string) []SuppressionViolation {
	prefix := commentPrefix(filePath)
	if prefix == "" {
		return nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	ignoreMarker := prefix + " ast-grep-ignore"
	reasonPrefix := prefix + " @MX:REASON"

	// 파일을 행 단위로 읽기
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	var violations []SuppressionViolation
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != ignoreMarker {
			continue
		}

		// i번째 줄에서 ast-grep-ignore 발견 (1-indexed: i+1)
		found := false
		blankCount := 0

		for j := i + 1; j < len(lines); j++ {
			nextTrimmed := strings.TrimSpace(lines[j])

			if nextTrimmed == "" {
				blankCount++
				if blankCount > 1 {
					// 빈 줄 2개 이상: 너무 멀다
					break
				}
				continue
			}

			// 비어 있지 않은 첫 번째 행 도달
			// @MX:REASON이 존재하고 이후 텍스트가 비어 있지 않으면 OK
			if strings.HasPrefix(nextTrimmed, reasonPrefix) {
				rest := strings.TrimSpace(strings.TrimPrefix(nextTrimmed, reasonPrefix))
				if rest != "" {
					found = true
				}
			}
			break
		}

		if !found {
			violations = append(violations, SuppressionViolation{
				Type: "SUPPRESSION_WITHOUT_REASON",
				File: filePath,
				Line: i + 1,
				Message: fmt.Sprintf(
					"ast-grep suppression at %s:%d requires adjacent '// @MX:REASON <rationale>' on next line",
					filePath, i+1,
				),
			})
		}
	}

	return violations
}

// walkSourceFiles는 projectDir 하위의 소스 파일을 재귀 탐색합니다.
// commentPrefix가 비어 있지 않은 파일만 반환합니다.
// 억제 정책 검사는 프로덕션 코드에 적용되므로 *_test.go 및 일반적인 제외 경로는 건너뜁니다.
func walkSourceFiles(projectDir string) []string {
	var files []string
	_ = filepath.WalkDir(projectDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		// vendor, .git, node_modules 등 제외
		parts := strings.Split(filepath.ToSlash(path), "/")
		for _, part := range parts {
			if part == "vendor" || part == ".git" || part == "node_modules" || part == "__pycache__" {
				return nil
			}
		}
		// *_test.go 파일 제외: 테스트 파일은 ast-grep-ignore를 테스트 픽스처 데이터로
		// 사용하는 경우가 많으므로 억제 정책 검사에서 제외합니다.
		if strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		if commentPrefix(path) != "" {
			files = append(files, path)
		}
		return nil
	})
	return files
}

// RunAstGrepGate scans the project using domain-specific ast-grep rules.
// Returns (passed, output) where passed=false blocks the commit.
// When sg CLI is not available, returns (true, "") gracefully.
// When the rules directory does not exist, returns (true, "") gracefully.
func RunAstGrepGate(ctx context.Context, projectDir string, cfg *AstGrepGateConfig) (bool, string) {
	if cfg == nil || !cfg.Enabled {
		return true, ""
	}

	// Check if sg CLI is available — pass silently if not found
	if _, err := exec.LookPath("sg"); err != nil {
		return true, ""
	}

	rulesDir := filepath.Join(projectDir, cfg.RulesDir)
	if _, err := os.Stat(rulesDir); err != nil {
		// Pass silently when the rules directory does not exist
		return true, ""
	}

	// Load rule files
	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDirectory(rulesDir)
	if err != nil || len(rules) == 0 {
		return true, ""
	}

	// Apply 30-second timeout for the full scan
	scanCtx, cancel := context.WithTimeout(ctx, astGrepScanTimeout)
	defer cancel()

	// Use config-based scan if sgconfig.yml exists, otherwise scan per-rule pattern
	var allMatches []astGrepScanMatch
	sgconfigPath := filepath.Join(rulesDir, "sgconfig.yml")
	if _, err := os.Stat(sgconfigPath); err == nil {
		matches, scanErr := runSGConfig(scanCtx, sgconfigPath, projectDir)
		if scanErr == nil {
			allMatches = matches
		}
	} else {
		for _, rule := range rules {
			if rule.Pattern == "" || rule.Language == "" {
				continue
			}
			matches, scanErr := runSGRule(scanCtx, rule, projectDir)
			if scanErr != nil {
				continue
			}
			allMatches = append(allMatches, matches...)
		}
	}

	if len(allMatches) == 0 {
		return true, ""
	}

	// Format results and classify by severity
	var sb strings.Builder
	sb.WriteString("ast-grep domain rule scan results:\n\n")
	hasError := false
	for _, m := range allMatches {
		sev := strings.ToLower(m.Severity)
		if sev == "" {
			sev = "warning"
		}
		line := m.Range.Start.Line + 1 // 0-indexed → 1-indexed (convert to human-readable line number)
		ruleID := m.RuleID
		if ruleID == "" {
			ruleID = "unknown"
		}
		msg := m.Message
		if msg == "" {
			msg = m.Lines
		}
		fmt.Fprintf(&sb, "%s:%d: [%s] %s (%s)\n", m.File, line, ruleID, msg, sev)
		if sev == "error" {
			hasError = true
		}
	}

	output := strings.TrimSpace(sb.String())

	// Block when error-severity matches exist, WarnOnlyMode is off, and BlockOnError is enabled
	if hasError && !cfg.WarnOnlyMode && cfg.BlockOnError {
		return false, fmt.Sprintf("quality gate failed: ast-grep domain rules\n\n%s", output)
	}

	return true, output
}

// runSGConfig runs sg scan using an sgconfig.yml file.
func runSGConfig(ctx context.Context, configPath, projectDir string) ([]astGrepScanMatch, error) {
	cmd := exec.CommandContext(ctx, "sg", "scan", "--config", configPath, "--json", projectDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Ignore the error: sg may return a non-zero exit code when matches are found
	_ = cmd.Run()

	return parseSGScanOutput(stdout.Bytes())
}

// runSGRule runs sg for a single rule using its pattern and language.
func runSGRule(ctx context.Context, rule astgrep.Rule, projectDir string) ([]astGrepScanMatch, error) {
	cmd := exec.CommandContext(ctx, "sg", "run",
		"--pattern", rule.Pattern,
		"--lang", rule.Language,
		"--json",
		projectDir,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Ignore the error: sg may return a non-zero exit code when matches are found
	_ = cmd.Run()

	matches, err := parseSGScanOutput(stdout.Bytes())
	if err != nil {
		return nil, err
	}

	// Inject rule metadata into matches
	for i := range matches {
		if matches[i].RuleID == "" {
			matches[i].RuleID = rule.ID
		}
		if matches[i].Severity == "" {
			matches[i].Severity = rule.Severity
		}
		if matches[i].Message == "" {
			matches[i].Message = rule.Message
		}
	}

	return matches, nil
}

// parseSGScanOutput parses the JSON array output from sg scan/run --json.
func parseSGScanOutput(output []byte) ([]astGrepScanMatch, error) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return nil, nil
	}

	var matches []astGrepScanMatch
	if err := json.Unmarshal(trimmed, &matches); err != nil {
		return nil, fmt.Errorf("parse sg output: %w", err)
	}
	return matches, nil
}
