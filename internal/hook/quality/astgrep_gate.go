package quality

import (
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

// astGrepScanMatch는 sg scan --json 출력의 단일 매치 항목을 나타낸다.
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

// RunAstGrepGate scans the project using domain-specific ast-grep rules.
// Returns (passed, output) where passed=false blocks the commit.
// When sg CLI is not available, returns (true, "") gracefully.
// When the rules directory does not exist, returns (true, "") gracefully.
func RunAstGrepGate(ctx context.Context, projectDir string, cfg *AstGrepGateConfig) (bool, string) {
	if cfg == nil || !cfg.Enabled {
		return true, ""
	}

	// sg CLI 존재 여부 확인 — 없으면 조용히 통과
	if _, err := exec.LookPath("sg"); err != nil {
		return true, ""
	}

	rulesDir := filepath.Join(projectDir, cfg.RulesDir)
	if _, err := os.Stat(rulesDir); err != nil {
		// 룰 디렉토리가 없으면 조용히 통과
		return true, ""
	}

	// 룰 파일 로드
	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDirectory(rulesDir)
	if err != nil || len(rules) == 0 {
		return true, ""
	}

	// 전체 스캔에 30초 타임아웃 적용
	scanCtx, cancel := context.WithTimeout(ctx, astGrepScanTimeout)
	defer cancel()

	// sgconfig.yml 존재 시 config-based 스캔, 없으면 룰별 패턴 스캔
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

	// 결과 포맷팅 및 심각도별 분류
	var sb strings.Builder
	sb.WriteString("ast-grep domain rule scan results:\n\n")
	hasError := false
	for _, m := range allMatches {
		sev := strings.ToLower(m.Severity)
		if sev == "" {
			sev = "warning"
		}
		line := m.Range.Start.Line + 1 // 0-indexed → 1-indexed
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

	// error 심각도 매치가 있고 WarnOnlyMode가 아니며 BlockOnError가 활성화된 경우 차단
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

	// sg가 매치를 찾았을 때 non-zero 종료 코드를 반환할 수 있으므로 에러를 무시
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

	// sg가 매치를 찾았을 때 non-zero 종료 코드를 반환할 수 있으므로 에러를 무시
	_ = cmd.Run()

	matches, err := parseSGScanOutput(stdout.Bytes())
	if err != nil {
		return nil, err
	}

	// 룰 메타데이터를 매치에 주입
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
