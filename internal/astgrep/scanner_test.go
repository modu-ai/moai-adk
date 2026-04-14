package astgrep_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// TestNewScanner_DefaultConfig: 기본 Config로 Scanner 생성이 성공하는지 검증
func TestNewScanner_DefaultConfig(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	s := astgrep.NewScanner(cfg)
	if s == nil {
		t.Fatal("NewScanner() returned nil")
	}
}

// TestNewScanner_NilConfig: nil Config 전달 시 패닉 없이 기본값으로 동작하는지 검증
func TestNewScanner_NilConfig(t *testing.T) {
	s := astgrep.NewScanner(nil)
	if s == nil {
		t.Fatal("NewScanner(nil) returned nil")
	}
}

// TestScanner_SGNotAvailable: sg CLI가 PATH에 없을 때 (nil, nil) 반환 검증 (AC3)
// 기본 허용 binary("sg")를 사용하되 실제 PATH에 없는 환경을 시뮬레이션한다.
// 참고: 비허용 binary name은 F2 보안 검사에서 에러로 차단된다.
func TestScanner_SGNotAvailable(t *testing.T) {
	// "sg"는 ValidateBinary를 통과하지만, PATH에 없으면 isSGAvailable이 false를 반환한다.
	// 이 테스트는 sg가 실제로 없는 환경(PATH에 "sg" 없음)에서만 유의미하다.
	// sg가 설치된 환경에서는 실제 스캔이 발생할 수 있으므로, 빈 rules 디렉토리로 테스트한다.
	tmpDir := t.TempDir()
	cfg := astgrep.DefaultScannerConfig()
	cfg.SGBinary = "sg" // 허용된 bare name
	cfg.RulesDir = tmpDir

	s := astgrep.NewScanner(cfg)
	findings, err := s.Scan(context.Background(), ".")
	if err != nil {
		t.Errorf("Scan() error = %v; sg 미존재 또는 빈 rules 시 nil 에러를 반환해야 함", err)
	}
	if len(findings) != 0 {
		t.Errorf("Scan() len(findings) = %d; 빈 rules 디렉토리에서 빈 슬라이스를 반환해야 함", len(findings))
	}
}

// TestScanner_EmptyRulesDir: 빈 rules 디렉토리일 때 오류 없이 빈 결과 반환 (AC3)
func TestScanner_EmptyRulesDir(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := astgrep.DefaultScannerConfig()
	cfg.RulesDir = tmpDir
	s := astgrep.NewScanner(cfg)

	findings, err := s.Scan(context.Background(), ".")
	if err != nil {
		t.Errorf("Scan() error = %v; 빈 rules 디렉토리에서 오류가 발생하면 안 됨", err)
	}
	if findings == nil {
		t.Error("Scan() returned nil findings; 빈 슬라이스를 반환해야 함")
	}
}

// TestScanner_RulesDirNotExist: rules 디렉토리가 존재하지 않을 때 오류 없이 동작 (AC3)
func TestScanner_RulesDirNotExist(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	cfg.RulesDir = "/path/that/does/not/exist/12345"
	s := astgrep.NewScanner(cfg)

	findings, err := s.Scan(context.Background(), ".")
	if err != nil {
		t.Errorf("Scan() error = %v; 존재하지 않는 rules 디렉토리에서 오류가 발생하면 안 됨", err)
	}
	if len(findings) != 0 {
		t.Errorf("Scan() len(findings) = %d; 규칙 없을 때 빈 슬라이스를 반환해야 함", len(findings))
	}
}

// TestScanner_FindingSeverityClassification: Finding 구조체의 severity 분류 검증
func TestScanner_FindingSeverityClassification(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		wantErr  bool
		wantWarn bool
		wantInfo bool
	}{
		{"error severity", "error", true, false, false},
		{"warning severity", "warning", false, true, false},
		{"info severity", "info", false, false, true},
		{"empty defaults to info", "", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := astgrep.Finding{
				RuleID:   "test-rule",
				Severity: tt.severity,
				Message:  "테스트 메시지",
				File:     "test.go",
				Line:     1,
			}
			if got := f.IsError(); got != tt.wantErr {
				t.Errorf("Finding.IsError() = %v, want %v", got, tt.wantErr)
			}
			if got := f.IsWarning(); got != tt.wantWarn {
				t.Errorf("Finding.IsWarning() = %v, want %v", got, tt.wantWarn)
			}
			if got := f.IsInfo(); got != tt.wantInfo {
				t.Errorf("Finding.IsInfo() = %v, want %v", got, tt.wantInfo)
			}
		})
	}
}

// TestRuleLoader_LoadFromDirRecursive: 서브디렉토리를 재귀적으로 로딩하는지 검증 (AC2 - RuleLoader.LoadFromDir)
func TestRuleLoader_LoadFromDirRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	// go/ 서브디렉토리 생성
	goDir := filepath.Join(tmpDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("서브디렉토리 생성 실패: %v", err)
	}

	// security/ 서브디렉토리 생성
	secDir := filepath.Join(tmpDir, "security")
	if err := os.MkdirAll(secDir, 0o755); err != nil {
		t.Fatalf("security 서브디렉토리 생성 실패: %v", err)
	}

	// go/test.yml 생성
	goRuleYAML := `---
id: test-go-rule
language: go
severity: warning
message: "테스트 규칙"
pattern: "os.Getenv($X)"
`
	if err := os.WriteFile(filepath.Join(goDir, "test.yml"), []byte(goRuleYAML), 0o644); err != nil {
		t.Fatalf("규칙 파일 작성 실패: %v", err)
	}

	// security/test.yml 생성
	secRuleYAML := `---
id: test-sec-rule
language: go
severity: error
message: "보안 테스트 규칙"
pattern: "exec.Command($X)"
`
	if err := os.WriteFile(filepath.Join(secDir, "test.yml"), []byte(secRuleYAML), 0o644); err != nil {
		t.Fatalf("보안 규칙 파일 작성 실패: %v", err)
	}

	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("LoadFromDir() len(rules) = %d, want 2 (서브디렉토리 재귀 로딩 필요)", len(rules))
	}

	// 규칙 ID 검증
	ids := make(map[string]bool)
	for _, r := range rules {
		ids[r.ID] = true
	}
	if !ids["test-go-rule"] {
		t.Error("test-go-rule이 로딩되지 않았음")
	}
	if !ids["test-sec-rule"] {
		t.Error("test-sec-rule이 로딩되지 않았음")
	}
}

// TestRuleLoader_SkipsInvalidYAML: 파싱 실패한 개별 규칙은 건너뛰고 나머지 로딩 (AC3)
func TestRuleLoader_SkipsInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// 유효한 규칙 파일
	validYAML := `---
id: valid-rule
language: go
severity: warning
message: "유효한 규칙"
pattern: "fmt.Println($X)"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "valid.yml"), []byte(validYAML), 0o644); err != nil {
		t.Fatalf("유효한 규칙 파일 작성 실패: %v", err)
	}

	// 유효하지 않은 YAML 파일 (파싱 불가)
	invalidYAML := `{invalid yaml: [`
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.yml"), []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("잘못된 규칙 파일 작성 실패: %v", err)
	}

	loader := astgrep.NewRuleLoader()
	// 에러를 반환하지 않고 유효한 규칙만 로딩해야 함
	rules, err := loader.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v; 잘못된 파일은 건너뛰어야 함", err)
	}

	if len(rules) < 1 {
		t.Error("유효한 규칙이 최소 1개 이상 로딩되어야 함")
	}
}

// TestScanner_ScanConfig: Config 필드 설정 검증
func TestScanner_ScanConfig(t *testing.T) {
	cfg := &astgrep.ScannerConfig{
		RulesDir:     "/custom/rules",
		SGBinary:     "sg",
		WarnOnlyMode: true,
	}
	s := astgrep.NewScanner(cfg)
	if s == nil {
		t.Fatal("NewScanner() with custom config returned nil")
	}
}

// TestScanner_ContextCancellation: context 취소 시 Scan이 즉시 반환하는지 검증
func TestScanner_ContextCancellation(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	s := astgrep.NewScanner(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	// context가 이미 취소된 상태에서도 패닉 없이 반환해야 함
	_, _ = s.Scan(ctx, ".")
}

// TestFinding_String: Finding.String() 메서드가 사람이 읽기 좋은 형식으로 출력하는지 검증
func TestFinding_String(t *testing.T) {
	f := astgrep.Finding{
		RuleID:   "go-no-raw-getenv",
		Severity: "warning",
		Message:  "환경변수를 직접 사용하지 마세요",
		File:     "main.go",
		Line:     42,
		Column:   10,
	}
	got := f.String()
	if got == "" {
		t.Error("Finding.String() returned empty string")
	}
	// 파일명과 줄 번호가 포함되어야 함
	if !containsSubstr(got, "main.go") {
		t.Errorf("Finding.String() = %q; 파일명을 포함해야 함", got)
	}
}

// TestScanner_HasErrors: HasErrors() 메서드가 error severity 여부를 올바르게 판단하는지 검증
func TestScanner_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		findings []astgrep.Finding
		want     bool
	}{
		{
			name:     "빈 findings",
			findings: nil,
			want:     false,
		},
		{
			name: "warning만 있는 경우",
			findings: []astgrep.Finding{
				{RuleID: "r1", Severity: "warning"},
			},
			want: false,
		},
		{
			name: "error가 포함된 경우",
			findings: []astgrep.Finding{
				{RuleID: "r1", Severity: "warning"},
				{RuleID: "r2", Severity: "error"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := astgrep.HasErrors(tt.findings); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestScanner_ScanWithSGAvailable: sg CLI가 있는 경우 실제 스캔 (통합 테스트, sg 없으면 skip)
func TestScanner_ScanWithSGAvailable(t *testing.T) {
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg CLI가 PATH에 없어 스킵합니다")
	}

	tmpDir := t.TempDir()

	goCode := `package main

import (
	"fmt"
	"os"
)

func main() {
	val := os.Getenv("MY_API_KEY")
	fmt.Println(val)
}
`
	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0o644); err != nil {
		t.Fatalf("Go 파일 작성 실패: %v", err)
	}

	rulesDir := filepath.Join(tmpDir, "rules", "go")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("rules 디렉토리 생성 실패: %v", err)
	}

	ruleYAML := "---\nid: test-getenv-rule\nlanguage: go\nseverity: warning\nmessage: \"raw os.Getenv 사용 금지\"\npattern: 'os.Getenv(\"$X\")'\n"
	if err := os.WriteFile(filepath.Join(rulesDir, "test.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatalf("규칙 파일 작성 실패: %v", err)
	}

	cfg := &astgrep.ScannerConfig{
		RulesDir: filepath.Join(tmpDir, "rules"),
		SGBinary: "sg",
	}
	s := astgrep.NewScanner(cfg)
	findings, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	// sg가 있으면 결과가 슬라이스여야 함 (nil이 아닌)
	if findings == nil {
		t.Error("Scan() returned nil with sg available")
	}
}

// TestParseSGFindings_EmptyOutput: 빈 출력에 대해 빈 슬라이스 반환 검증
func TestParseSGFindings_EmptyOutput(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	cfg.RulesDir = t.TempDir()
	s := astgrep.NewScanner(cfg)

	findings, err := s.Scan(context.Background(), ".")
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if findings == nil {
		t.Error("Scan() returned nil; 빈 슬라이스를 반환해야 함")
	}
}

// TestToFileURI_Paths: SARIF URI 변환 검증 (내부 함수 간접 테스트)
func TestToFileURI_Paths(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Severity: "warning", Message: "m1", File: "internal/pkg/file.go", Line: 1},
		{RuleID: "r2", Severity: "error", Message: "m2", File: "", Line: 1},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}
	if len(output) == 0 {
		t.Error("ToSARIF() returned empty output")
	}
}

// TestLoadFromDir_PlaceholderDirs: .gitkeep이 있는 플레이스홀더 디렉토리를 처리하는지 검증
func TestLoadFromDir_PlaceholderDirs(t *testing.T) {
	tmpDir := t.TempDir()

	for _, lang := range []string{"python", "typescript", "rust"} {
		dir := filepath.Join(tmpDir, lang)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("디렉토리 생성 실패: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte{}, 0o644); err != nil {
			t.Fatalf(".gitkeep 작성 실패: %v", err)
		}
	}

	goDir := filepath.Join(tmpDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("go 디렉토리 생성 실패: %v", err)
	}
	ruleYAML := "---\nid: go-test-rule\nlanguage: go\nseverity: warning\nmessage: \"테스트\"\npattern: \"fmt.Println($X)\"\n"
	if err := os.WriteFile(filepath.Join(goDir, "test.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatalf("규칙 파일 작성 실패: %v", err)
	}

	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("LoadFromDir() len(rules) = %d, want 1 (.gitkeep 파일은 무시되어야 함)", len(rules))
	}
}

// TestHasErrors_MixedSeverities: 다양한 severity가 섞인 슬라이스에서 HasErrors 동작 검증
func TestHasErrors_MixedSeverities(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Severity: "info"},
		{RuleID: "r2", Severity: "warning"},
		{RuleID: "r3", Severity: "ERROR"},
	}

	if !astgrep.HasErrors(findings) {
		t.Error("HasErrors() = false; 대소문자 무관하게 error severity를 감지해야 함")
	}
}

// TestFinding_IsInfoForHint: hint severity가 IsError/IsWarning에서 false인지 검증
func TestFinding_IsInfoForHint(t *testing.T) {
	f := astgrep.Finding{Severity: "hint"}
	if f.IsError() {
		t.Error("hint severity가 IsError() = true를 반환했음")
	}
	if f.IsWarning() {
		t.Error("hint severity가 IsWarning() = true를 반환했음")
	}
}

// TestScanner_ScanWithSGConfig: sgconfig.yml을 사용한 스캔 (sg 없으면 skip)
func TestScanner_ScanWithSGConfig(t *testing.T) {
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg CLI가 PATH에 없어 스킵합니다")
	}

	tmpDir := t.TempDir()

	// 대상 파일 생성
	goCode := `package main
import "fmt"
func main() { fmt.Println("hello") }
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goCode), 0o644); err != nil {
		t.Fatalf("파일 작성 실패: %v", err)
	}

	// sgconfig.yml과 rules 디렉토리 생성
	rulesDir := filepath.Join(tmpDir, "rules")
	goDir := filepath.Join(rulesDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("rules 디렉토리 생성 실패: %v", err)
	}

	ruleYAML := "---\nid: sgconfig-test-rule\nlanguage: go\nseverity: warning\nmessage: \"fmt.Println 사용\"\npattern: 'fmt.Println($X)'\n"
	if err := os.WriteFile(filepath.Join(goDir, "test.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatalf("규칙 파일 작성 실패: %v", err)
	}

	// sgconfig.yml 생성
	sgconfig := "ruleDirs:\n  - go\n"
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte(sgconfig), 0o644); err != nil {
		t.Fatalf("sgconfig.yml 작성 실패: %v", err)
	}

	cfg := &astgrep.ScannerConfig{
		RulesDir: rulesDir,
		SGBinary: "sg",
	}
	s := astgrep.NewScanner(cfg)
	findings, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// sgconfig.yml 경로를 통한 스캔이 성공해야 함 (findings는 sg 동작에 따라 달라짐)
	if findings == nil {
		t.Error("Scan() with sgconfig.yml returned nil")
	}
}

// TestScanner_String_AllSeverities: Finding.String()이 다양한 severity를 출력하는지 검증
func TestScanner_String_AllSeverities(t *testing.T) {
	tests := []struct {
		severity string
	}{
		{"error"},
		{"warning"},
		{"info"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			f := astgrep.Finding{
				RuleID:   "test-rule",
				Severity: tt.severity,
				Message:  "테스트 메시지",
				File:     "test.go",
				Line:     1,
			}
			s := f.String()
			if s == "" {
				t.Error("Finding.String() returned empty string")
			}
			if !containsSubstr(s, "test.go") {
				t.Errorf("String() does not contain filename: %q", s)
			}
		})
	}
}

// TestDefaultScannerConfig_Fields: DefaultScannerConfig 기본값 검증
func TestDefaultScannerConfig_Fields(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	if cfg.SGBinary != "sg" {
		t.Errorf("DefaultScannerConfig().SGBinary = %q, want sg", cfg.SGBinary)
	}
	if cfg.RulesDir == "" {
		t.Error("DefaultScannerConfig().RulesDir is empty")
	}
	if cfg.Timeout == 0 {
		t.Error("DefaultScannerConfig().Timeout is zero")
	}
}

// containsSubstr는 문자열 s에 substr이 포함되어 있는지 확인하는 헬퍼 함수
func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// --- F2 재현 테스트: Binary allowlist 부재 ---

// TestValidateBinary_AllowsSg: "sg", "ast-grep" bare name이 허용되는지 검증
func TestValidateBinary_AllowsSg(t *testing.T) {
	for _, name := range []string{"sg", "ast-grep"} {
		if err := astgrep.ValidateBinary(name); err != nil {
			t.Errorf("ValidateBinary(%q) = %v; 허용된 bare name이어야 함", name, err)
		}
	}
}

// TestValidateBinary_AllowsTrustedPrefix: /usr/local/bin/sg 같은 신뢰 경로가 허용되는지 검증
func TestValidateBinary_AllowsTrustedPrefix(t *testing.T) {
	paths := []string{
		"/usr/bin/sg",
		"/usr/local/bin/sg",
		"/opt/homebrew/bin/sg",
	}
	for _, p := range paths {
		if err := astgrep.ValidateBinary(p); err != nil {
			t.Errorf("ValidateBinary(%q) = %v; 신뢰 경로여야 함", p, err)
		}
	}
}

// TestValidateBinary_RejectsUntrustedPath: /tmp 같은 비신뢰 절대 경로가 거부되는지 검증
func TestValidateBinary_RejectsUntrustedPath(t *testing.T) {
	if err := astgrep.ValidateBinary("/tmp/evil/sg"); err == nil {
		t.Error("ValidateBinary(/tmp/evil/sg) = nil; 비신뢰 경로는 에러를 반환해야 함")
	}
}

// TestValidateBinary_RejectsShellMetachars: 셸 메타문자가 포함된 바이너리 경로가 거부되는지 검증
func TestValidateBinary_RejectsShellMetachars(t *testing.T) {
	malicious := []string{
		"sg; rm -rf /",
		"sg|cat /etc/passwd",
		"`sg`",
	}
	for _, bin := range malicious {
		if err := astgrep.ValidateBinary(bin); err == nil {
			t.Errorf("ValidateBinary(%q) = nil; 셸 메타문자는 에러를 반환해야 함", bin)
		}
	}
}

// TestValidateBinary_RejectsTraversal: ..을 포함한 경로 트래버설이 거부되는지 검증
func TestValidateBinary_RejectsTraversal(t *testing.T) {
	if err := astgrep.ValidateBinary("/usr/local/bin/../../tmp/sg"); err == nil {
		t.Error("ValidateBinary(경로 트래버설) = nil; 에러를 반환해야 함")
	}
}

// TestScan_RejectsUntrustedBinary: Scan이 신뢰할 수 없는 바이너리 경로로 에러를 반환하는지 검증
func TestScan_RejectsUntrustedBinary(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	cfg.SGBinary = "/tmp/evil/sg"
	s := astgrep.NewScanner(cfg)

	_, err := s.Scan(context.Background(), ".")
	if err == nil {
		t.Error("Scan() with untrusted binary = nil error; 에러를 반환해야 함")
	}
}
