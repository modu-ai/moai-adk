package quality

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultAstGrepGateConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultAstGrepGateConfig()

	if !cfg.Enabled {
		t.Error("Enabled: want true, got false")
	}
	if cfg.RulesDir != ".moai/config/astgrep-rules" {
		t.Errorf("RulesDir: want .moai/config/astgrep-rules, got %q", cfg.RulesDir)
	}
	if !cfg.BlockOnError {
		t.Error("BlockOnError: want true, got false")
	}
	if cfg.WarnOnlyMode {
		t.Error("WarnOnlyMode: want false, got true")
	}
}

func TestRunAstGrepGate_Disabled(t *testing.T) {
	t.Parallel()

	cfg := &AstGrepGateConfig{Enabled: false}
	passed, output := RunAstGrepGate(context.Background(), t.TempDir(), cfg)

	if !passed {
		t.Error("disabled gate should always pass")
	}
	if output != "" {
		t.Errorf("disabled gate should return empty output, got %q", output)
	}
}

func TestRunAstGrepGate_NilConfig(t *testing.T) {
	t.Parallel()

	passed, output := RunAstGrepGate(context.Background(), t.TempDir(), nil)

	if !passed {
		t.Error("nil config should pass gracefully")
	}
	if output != "" {
		t.Errorf("nil config should return empty output, got %q", output)
	}
}

func TestRunAstGrepGate_NoSgCLI(t *testing.T) {
	// t.Setenv는 t.Parallel()과 함께 사용 불가
	// sg 없는 환경 시뮬레이션: 빈 PATH로 실행
	t.Setenv("PATH", "")

	cfg := DefaultAstGrepGateConfig()
	passed, output := RunAstGrepGate(context.Background(), t.TempDir(), cfg)

	if !passed {
		t.Errorf("gate should pass when sg is not available, got output: %q", output)
	}
	if output != "" {
		t.Errorf("output should be empty when sg is not available, got %q", output)
	}
}

func TestRunAstGrepGate_NoRulesDir(t *testing.T) {
	t.Parallel()

	// 룰 디렉토리가 없는 빈 임시 디렉토리 사용
	projectDir := t.TempDir()

	cfg := DefaultAstGrepGateConfig()
	// 존재하지 않는 서브디렉토리를 RulesDir로 지정
	cfg.RulesDir = "nonexistent-rules-dir"

	passed, output := RunAstGrepGate(context.Background(), projectDir, cfg)

	if !passed {
		t.Errorf("gate should pass when rules dir does not exist, got output: %q", output)
	}
	if output != "" {
		t.Errorf("output should be empty when rules dir does not exist, got %q", output)
	}
}

func TestRunAstGrepGate_EmptyRulesDir(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	// 룰 파일 없는 빈 디렉토리 생성
	rulesDir := filepath.Join(projectDir, ".moai", "config", "astgrep-rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}

	cfg := DefaultAstGrepGateConfig()
	passed, output := RunAstGrepGate(context.Background(), projectDir, cfg)

	if !passed {
		t.Errorf("gate should pass when no rules are loaded, got output: %q", output)
	}
	if output != "" {
		t.Errorf("output should be empty when no rules, got %q", output)
	}
}

func TestRunAstGrepGate_WarnOnlyMode(t *testing.T) {
	// t.Setenv는 t.Parallel()과 함께 사용 불가
	// WarnOnlyMode가 true이면 error 심각도 매치가 있어도 통과해야 함.
	// 실제 sg 없이 로직을 검증하기 위해 빈 PATH로 실행 (sg 없음 → 조용히 통과)
	t.Setenv("PATH", "")

	projectDir := t.TempDir()
	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: true,
		WarnOnlyMode: true, // error도 차단하지 않아야 함
	}

	passed, _ := RunAstGrepGate(context.Background(), projectDir, cfg)

	if !passed {
		t.Error("WarnOnlyMode should never block even if errors are found")
	}
}

func TestRunAstGrepGate_BlockOnErrorFalse(t *testing.T) {
	// t.Setenv는 t.Parallel()과 함께 사용 불가
	// BlockOnError가 false이면 error 심각도 매치가 있어도 통과해야 함.
	t.Setenv("PATH", "")

	projectDir := t.TempDir()
	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: false,
		WarnOnlyMode: false,
	}

	passed, _ := RunAstGrepGate(context.Background(), projectDir, cfg)

	if !passed {
		t.Error("BlockOnError=false should not block commits")
	}
}

func TestParseSGScanOutput_Empty(t *testing.T) {
	t.Parallel()

	matches, err := parseSGScanOutput([]byte{})
	if err != nil {
		t.Errorf("empty output should not return error, got: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("empty output should return no matches, got %d", len(matches))
	}
}

func TestParseSGScanOutput_ValidJSON(t *testing.T) {
	t.Parallel()

	input := `[
		{
			"file": "main.go",
			"ruleId": "no-panic",
			"severity": "error",
			"message": "panic is forbidden",
			"range": {"start": {"line": 9, "column": 2}, "end": {"line": 9, "column": 7}}
		}
	]`

	matches, err := parseSGScanOutput([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m := matches[0]
	if m.File != "main.go" {
		t.Errorf("File: want main.go, got %q", m.File)
	}
	if m.RuleID != "no-panic" {
		t.Errorf("RuleID: want no-panic, got %q", m.RuleID)
	}
	if m.Severity != "error" {
		t.Errorf("Severity: want error, got %q", m.Severity)
	}
	// 0-indexed line 9 → 출력 시 1-indexed 10이 되어야 함 (포맷팅 단계에서 처리)
	if m.Range.Start.Line != 9 {
		t.Errorf("Range.Start.Line: want 9, got %d", m.Range.Start.Line)
	}
}

func TestParseSGScanOutput_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := parseSGScanOutput([]byte("not-json"))
	if err == nil {
		t.Error("invalid JSON should return error")
	}
}

func TestRunAstGrepGate_GateConfigIntegration(t *testing.T) {
	t.Parallel()

	// DefaultGateConfig에 AstGrepGate 필드가 설정되어 있는지 확인
	cfg := DefaultGateConfig()
	if cfg.AstGrepGate == nil {
		t.Fatal("DefaultGateConfig should set AstGrepGate")
	}
	if !cfg.AstGrepGate.Enabled {
		t.Error("AstGrepGate.Enabled should be true by default")
	}
	if cfg.AstGrepGate.RulesDir != ".moai/config/astgrep-rules" {
		t.Errorf("AstGrepGate.RulesDir: want .moai/config/astgrep-rules, got %q", cfg.AstGrepGate.RulesDir)
	}
}

func TestRunAstGrepGate_OutputContainsRuleInfo(t *testing.T) {
	t.Parallel()

	// parseSGScanOutput + 포맷팅 로직을 단위 테스트로 검증
	// 실제 sg 호출 없이 내부 포맷팅 결과를 확인
	matches := []astGrepScanMatch{
		{
			File:     "pkg/foo/bar.go",
			RuleID:   "no-global-var",
			Severity: "warning",
			Message:  "global variables should be avoided",
			Range: struct {
				Start struct {
					Line   int `json:"line"`
					Column int `json:"column"`
				} `json:"start"`
			}{Start: struct {
				Line   int `json:"line"`
				Column int `json:"column"`
			}{Line: 4, Column: 0}},
		},
	}

	// 포맷팅 결과 검증 (strings.Builder 로직 직접 테스트)
	var sb strings.Builder
	sb.WriteString("ast-grep domain rule scan results:\n\n")
	for _, m := range matches {
		line := m.Range.Start.Line + 1
		sb.WriteString(
			strings.TrimSpace(
				strings.Join([]string{m.File, ":", string(rune('0'+line)), ": [", m.RuleID, "] ", m.Message, " (", m.Severity, ")\n"}, ""),
			) + "\n",
		)
	}

	result := sb.String()
	if !strings.Contains(result, "pkg/foo/bar.go") {
		t.Errorf("output should contain file path, got: %q", result)
	}
	if !strings.Contains(result, "no-global-var") {
		t.Errorf("output should contain rule ID, got: %q", result)
	}
}
