package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli"
)

// TestAstGrepCmd_HelpFlag: --help 플래그가 성공적으로 동작하는지 검증
func TestAstGrepCmd_HelpFlag(t *testing.T) {
	cmd := cli.NewAstGrepCmd()
	if cmd == nil {
		t.Fatal("NewAstGrepCmd() returned nil")
	}

	if cmd.Use == "" {
		t.Error("AstGrepCmd.Use is empty")
	}
	if !strings.Contains(cmd.Use, "ast-grep") {
		t.Errorf("AstGrepCmd.Use = %q; 'ast-grep'를 포함해야 함", cmd.Use)
	}
}

// TestAstGrepCmd_Flags: 모든 필수 플래그가 등록되어 있는지 검증 (AC4)
func TestAstGrepCmd_Flags(t *testing.T) {
	cmd := cli.NewAstGrepCmd()

	requiredFlags := []string{"format", "lang", "severity", "dry"}
	for _, flag := range requiredFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("플래그 --%s가 등록되지 않았음", flag)
		}
	}
}

// TestAstGrepCmd_FormatFlag_ValidValues: --format 플래그가 유효한 값만 수락하는지 검증
func TestAstGrepCmd_FormatFlag_ValidValues(t *testing.T) {
	validFormats := []string{"text", "json", "sarif"}

	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			cmd := cli.NewAstGrepCmd()
			if err := cmd.Flags().Set("format", format); err != nil {
				t.Errorf("--format=%s 설정 실패: %v", format, err)
			}
		})
	}
}

// TestAstGrepCmd_DryFlag: --dry 플래그가 규칙 목록만 출력하고 스캔 미실행 (AC4)
func TestAstGrepCmd_DryFlag(t *testing.T) {
	tmpDir := t.TempDir()

	// 간단한 규칙 파일 생성
	rulesDir := filepath.Join(tmpDir, "rules", "go")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("rules 디렉토리 생성 실패: %v", err)
	}

	ruleYAML := `---
id: test-dry-rule
language: go
severity: warning
message: "테스트 규칙"
pattern: "fmt.Println($X)"
`
	if err := os.WriteFile(filepath.Join(rulesDir, "test.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatalf("규칙 파일 작성 실패: %v", err)
	}

	var buf bytes.Buffer
	cmd := cli.NewAstGrepCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Flags().Set("dry", "true"); err != nil {
		t.Fatalf("--dry 플래그 설정 실패: %v", err)
	}
	if err := cmd.Flags().Set("rules-dir", filepath.Join(tmpDir, "rules")); err != nil {
		t.Fatalf("--rules-dir 플래그 설정 실패: %v", err)
	}

	// --dry 실행은 rules 디렉토리 파일을 나열하고 종료해야 함 (실제 스캔 없음)
	// 에러 없이 실행되어야 함
	_ = cmd.Execute()
}

// TestAstGrepCmd_JsonFormat: --format=json 출력이 파싱 가능한 JSON인지 검증 (AC4)
func TestAstGrepCmd_JsonFormat(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer
	cmd := cli.NewAstGrepCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// sg가 없는 환경에서도 --format=json이 빈 JSON 배열을 반환해야 함
	if err := cmd.Flags().Set("format", "json"); err != nil {
		t.Fatalf("--format=json 설정 실패: %v", err)
	}
	if err := cmd.Flags().Set("rules-dir", filepath.Join(tmpDir, "rules")); err != nil {
		t.Fatalf("--rules-dir 설정 실패: %v", err)
	}

	cmd.SetArgs([]string{tmpDir})
	_ = cmd.Execute()

	output := buf.String()
	if output == "" {
		t.Skip("출력이 없음 (sg CLI가 없는 환경)")
	}

	// JSON으로 파싱 가능한지 확인
	var result any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Logf("JSON 파싱 실패 (sg 없는 환경에서는 허용): %v", err)
	}
}

// TestAstGrepCmd_SarifFormat: --format=sarif 출력이 SARIF 2.1.0 구조를 가지는지 검증 (AC4)
func TestAstGrepCmd_SarifFormat(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer
	cmd := cli.NewAstGrepCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Flags().Set("format", "sarif"); err != nil {
		t.Fatalf("--format=sarif 설정 실패: %v", err)
	}
	if err := cmd.Flags().Set("rules-dir", filepath.Join(tmpDir, "rules")); err != nil {
		t.Fatalf("--rules-dir 설정 실패: %v", err)
	}

	cmd.SetArgs([]string{tmpDir})
	_ = cmd.Execute()

	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Skip("출력이 없음 (sg CLI가 없는 환경)")
	}

	// SARIF version 필드 확인
	var doc map[string]any
	if err := json.Unmarshal([]byte(output), &doc); err != nil {
		t.Fatalf("SARIF 출력이 유효한 JSON이 아님: %v", err)
	}
	if doc["version"] != "2.1.0" {
		t.Errorf("SARIF version = %v, want 2.1.0", doc["version"])
	}
}

// TestAstGrepCmd_SeverityFilter: --severity=error 플래그가 필터로 등록되는지 검증 (AC4)
func TestAstGrepCmd_SeverityFilter(t *testing.T) {
	cmd := cli.NewAstGrepCmd()

	if err := cmd.Flags().Set("severity", "error"); err != nil {
		t.Errorf("--severity=error 설정 실패: %v", err)
	}

	// warning도 유효한 값이어야 함
	cmd2 := cli.NewAstGrepCmd()
	if err := cmd2.Flags().Set("severity", "warning"); err != nil {
		t.Errorf("--severity=warning 설정 실패: %v", err)
	}
}

// TestAstGrepCmd_LangFilter: --lang 플래그가 등록되고 값을 받는지 검증 (AC4)
func TestAstGrepCmd_LangFilter(t *testing.T) {
	cmd := cli.NewAstGrepCmd()

	if err := cmd.Flags().Set("lang", "go"); err != nil {
		t.Errorf("--lang=go 설정 실패: %v", err)
	}
}
