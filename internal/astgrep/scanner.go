package astgrep

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ScannerConfig는 통합 Scanner의 설정을 담습니다.
// REQ-ASTG-UPG-010, REQ-ASTG-UPG-011, REQ-ASTG-UPG-012
type ScannerConfig struct {
	// RulesDir는 ast-grep 규칙 디렉토리 경로입니다 (재귀 탐색).
	// 기본값: ".moai/config/astgrep-rules"
	RulesDir string
	// SGBinary는 sg CLI 바이너리 이름 또는 경로입니다.
	// 기본값: "sg"
	SGBinary string
	// WarnOnlyMode가 true이면 error severity 발견 시에도 차단하지 않습니다.
	WarnOnlyMode bool
	// Timeout은 전체 스캔 타임아웃입니다.
	// 기본값: 30초
	Timeout time.Duration
}

// DefaultScannerConfig는 기본값이 설정된 ScannerConfig를 반환합니다.
func DefaultScannerConfig() *ScannerConfig {
	return &ScannerConfig{
		RulesDir: ".moai/config/astgrep-rules",
		SGBinary: "sg",
		Timeout:  30 * time.Second,
	}
}

// Finding은 ast-grep 스캔의 단일 결과를 나타냅니다.
// quality gate, post-tool hook, CLI 서브커맨드가 공유하는 표준 타입입니다.
//
// @MX:ANCHOR: [AUTO] Finding은 scanner, CLI, hook이 공유하는 표준 데이터 타입
// @MX:REASON: fan_in >= 3: Scanner.Scan, SARIF 변환기, hook 통합 모두 이 타입을 사용
type Finding struct {
	// RuleID는 규칙 ID입니다.
	RuleID string `json:"ruleId"`
	// Severity는 심각도입니다: "error", "warning", "info"
	Severity string `json:"severity"`
	// Message는 규칙 메시지입니다.
	Message string `json:"message"`
	// File은 발견된 파일 경로입니다.
	File string `json:"file"`
	// Line은 1-indexed 줄 번호입니다.
	Line int `json:"line"`
	// Column은 0-indexed 컬럼 번호입니다.
	Column int `json:"column,omitempty"`
	// EndLine은 1-indexed 종료 줄 번호입니다.
	EndLine int `json:"endLine,omitempty"`
	// EndColumn은 종료 컬럼 번호입니다.
	EndColumn int `json:"endColumn,omitempty"`
	// Note는 규칙의 추가 설명입니다.
	Note string `json:"note,omitempty"`
	// Metadata는 CWE/OWASP 등 추가 메타데이터입니다.
	Metadata map[string]string `json:"metadata,omitempty"`
	// Language는 이 finding을 생성한 규칙의 대상 언어입니다.
	// scanWithRules 경로에서 rule.Language가 주입됩니다.
	// scanWithConfig 경로에서는 per-finding 언어를 알 수 없으므로 빈 문자열입니다.
	// --lang 필터는 Language가 빈 finding을 항상 포함합니다 (언어-중립 규칙 허용).
	Language string `json:"language,omitempty"`
}

// IsError는 심각도가 error인지 반환합니다.
func (f Finding) IsError() bool {
	return strings.ToLower(f.Severity) == "error"
}

// IsWarning은 심각도가 warning인지 반환합니다.
func (f Finding) IsWarning() bool {
	return strings.ToLower(f.Severity) == "warning"
}

// IsInfo는 심각도가 info이거나 빈 문자열인지 반환합니다.
func (f Finding) IsInfo() bool {
	s := strings.ToLower(f.Severity)
	return s == "info" || s == ""
}

// String은 Finding을 사람이 읽기 좋은 형식으로 반환합니다.
func (f Finding) String() string {
	sev := f.Severity
	if sev == "" {
		sev = "info"
	}
	return fmt.Sprintf("%s:%d: [%s] %s (%s)", f.File, f.Line, f.RuleID, f.Message, sev)
}

// HasErrors는 findings 슬라이스에 error severity 항목이 있는지 반환합니다.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.IsError() {
			return true
		}
	}
	return false
}

// ErrUntrustedBinary는 신뢰할 수 없는 바이너리 경로가 지정된 경우 반환됩니다.
var ErrUntrustedBinary = errors.New("astgrep: 신뢰할 수 없는 바이너리 경로")

// trustedBinaryPrefixes는 허용된 절대 경로 prefix 목록을 반환한다.
func trustedBinaryPrefixes() []string {
	home, _ := os.UserHomeDir()
	sep := string(os.PathSeparator)
	prefixes := []string{
		"/usr/bin/",
		"/usr/local/bin/",
		"/opt/homebrew/bin/",
	}
	if home != "" {
		prefixes = append(prefixes,
			filepath.Join(home, "go", "bin")+sep,
			filepath.Join(home, ".local", "bin")+sep,
			filepath.Join(home, ".cargo", "bin")+sep,
		)
	}
	return prefixes
}

// ValidateBinary는 sg 바이너리 경로의 안전성을 검사한다.
// 빈 문자열은 기본값 "sg"로 폴백되므로 허용한다.
// bare name은 "sg" 또는 "ast-grep"만 허용한다.
// 절대 경로는 신뢰 prefix 목록에 있어야 한다.
// 셸 메타문자나 경로 트래버설(..)이 포함되면 ErrUntrustedBinary를 반환한다.
func ValidateBinary(binary string) error {
	if binary == "" {
		// 빈 값은 기본값 "sg"로 폴백됨
		return nil
	}
	// 셸 인젝션 방어: 메타문자 차단
	if strings.ContainsAny(binary, ";|&`$()<>\n\r") {
		return ErrUntrustedBinary
	}
	// 경로 트래버설 차단
	if strings.Contains(binary, "..") {
		return ErrUntrustedBinary
	}
	// bare name: 허용 목록의 고정값만
	// filepath.IsAbs는 Windows에서 Unix 스타일 "/usr/bin/sg"를 절대경로로 보지 않으므로
	// 슬래시/백슬래시 포함 여부로 경로성을 판정해 크로스플랫폼 일관성을 확보한다.
	looksLikePath := strings.ContainsAny(binary, "/\\") || filepath.IsAbs(binary)
	if !looksLikePath {
		if binary == "sg" || binary == "ast-grep" {
			return nil
		}
		return ErrUntrustedBinary
	}
	// 절대 경로: 신뢰 prefix 검사 (Clean으로 트래버설 정규화 후)
	// Windows와 Unix를 동일하게 다루기 위해 양쪽 모두 ToSlash로 정규화한다.
	cleaned := filepath.ToSlash(filepath.Clean(binary))
	for _, p := range trustedBinaryPrefixes() {
		prefixSlash := strings.TrimRight(filepath.ToSlash(p), "/")
		if strings.HasPrefix(cleaned, prefixSlash+"/") || cleaned == prefixSlash {
			return nil
		}
	}
	return ErrUntrustedBinary
}

// Scanner는 ast-grep 기반의 통합 코드 스캐너입니다.
// quality gate와 post-tool hook의 분리된 구현을 대체합니다.
//
// @MX:ANCHOR: [AUTO] Scanner.Scan은 모든 ast-grep 스캔의 단일 진입점
// @MX:REASON: fan_in >= 3: quality gate hook, PostToolUse hook, CLI subcommand 모두 이 메서드를 호출
type Scanner struct {
	cfg *ScannerConfig
}

// NewScanner는 주어진 설정으로 새 Scanner를 생성합니다.
// cfg가 nil이면 DefaultScannerConfig()를 사용합니다.
func NewScanner(cfg *ScannerConfig) *Scanner {
	if cfg == nil {
		cfg = DefaultScannerConfig()
	}
	return &Scanner{cfg: cfg}
}

// sgScanMatch는 sg scan --json 출력의 내부 파싱 구조체입니다.
type sgScanMatch struct {
	File     string `json:"file"`
	Lines    string `json:"lines,omitempty"`
	Text     string `json:"text,omitempty"`
	RuleID   string `json:"ruleId,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`
	Note     string `json:"note,omitempty"`
	Range    struct {
		Start struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"start"`
		End struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"end"`
	} `json:"range"`
}

// isSGAvailable은 sg CLI 바이너리가 PATH에 있는지 확인합니다.
func (s *Scanner) isSGAvailable() bool {
	binary := s.cfg.SGBinary
	if binary == "" {
		binary = "sg"
	}
	_, err := exec.LookPath(binary)
	return err == nil
}

// Scan은 주어진 경로에 대해 모든 규칙을 적용하여 스캔을 수행합니다.
// sg CLI가 없으면 ([]Finding{}, nil)을 반환합니다 (REQ-ASTG-UPG-012).
// rules 디렉토리가 비어있거나 존재하지 않으면 ([]Finding{}, nil)을 반환합니다 (REQ-ASTG-UPG-012).
// SGBinary가 신뢰할 수 없는 경로이면 에러를 반환합니다 (F2 보안 검사).
func (s *Scanner) Scan(ctx context.Context, path string) ([]Finding, error) {
	// 바이너리 경로 보안 검증 (F2): 신뢰할 수 없는 경로는 즉시 에러 반환
	if err := ValidateBinary(s.cfg.SGBinary); err != nil {
		return []Finding{}, fmt.Errorf("sg 바이너리 검증 실패 (SGBinary=%q): %w", s.cfg.SGBinary, err)
	}

	// sg CLI가 없으면 warn_and_skip (REQ-ASTG-UPG-012)
	if !s.isSGAvailable() {
		slog.Warn("ast-grep (sg) CLI를 찾을 수 없습니다. 스캔을 건너뜁니다.",
			"binary", s.cfg.SGBinary,
			"hint", "https://ast-grep.github.io/guide/quick-start.html 에서 설치하세요")
		return []Finding{}, nil
	}

	// rules 디렉토리가 없으면 빈 결과 반환
	if _, err := os.Stat(s.cfg.RulesDir); err != nil {
		if os.IsNotExist(err) {
			return []Finding{}, nil
		}
		return []Finding{}, nil
	}

	// sgconfig.yml이 있으면 config 기반 스캔 사용
	sgconfigPath := filepath.Join(s.cfg.RulesDir, "sgconfig.yml")
	if _, err := os.Stat(sgconfigPath); err == nil {
		return s.scanWithConfig(ctx, sgconfigPath, path)
	}

	// sgconfig.yml이 없으면 재귀적으로 규칙을 로딩하여 스캔
	loader := NewRuleLoader()
	rules, err := loader.LoadFromDir(s.cfg.RulesDir)
	if err != nil || len(rules) == 0 {
		return []Finding{}, nil
	}

	return s.scanWithRules(ctx, rules, path)
}

// scanWithConfig는 sgconfig.yml을 사용하여 스캔합니다.
func (s *Scanner) scanWithConfig(ctx context.Context, configPath, path string) ([]Finding, error) {
	timeout := s.cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	binary := s.cfg.SGBinary
	if binary == "" {
		binary = "sg"
	}

	cmd := exec.CommandContext(scanCtx, binary, "scan", "--config", configPath, "--json", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// sg는 발견 시 non-zero exit code를 반환할 수 있으므로 에러를 무시
	_ = cmd.Run()

	// F4: stderr 내용이 있으면 디버그 로깅 (config 기반 스캔에서는 언어를 알 수 없으므로 에러 전파 생략)
	if stderr.Len() > 0 {
		slog.Debug("sg scan stderr (config 기반)", "config", configPath, "stderr", stderr.String())
	}

	if stdout.Len() == 0 {
		return []Finding{}, nil
	}

	return parseSGFindings(stdout.Bytes())
}

// runSingleRule은 단일 규칙에 대해 sg run을 실행하고 finding을 반환합니다.
// defer cancel()로 context 누수를 방지합니다 (F3).
// stderr는 디버그 로깅하고, stdout이 비어있고 stderr가 있으면 에러를 반환합니다 (F4).
//
// @MX:WARN: [AUTO] context.WithTimeout을 defer cancel()로 보호 — 이전 구현에서 cancel() 누수 발생
// @MX:REASON: F3 버그 수정: loop 내 cancel() 미 defer 시 패닉/조기 반환 경로에서 context 누수
func (s *Scanner) runSingleRule(ctx context.Context, binary string, rule Rule, path string, timeout time.Duration) ([]Finding, error) {
	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(scanCtx, binary, "run",
		"--pattern", rule.Pattern,
		"--lang", rule.Language,
		"--json",
		path,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// F4: stderr 내용 로깅
		if stderr.Len() > 0 {
			slog.Debug("sg run stderr", "rule", rule.ID, "stderr", stderr.String())
		}
		// stdout이 비어있고 stderr에 내용이 있으면 실행 실패로 간주
		if stdout.Len() == 0 && stderr.Len() > 0 {
			return nil, fmt.Errorf("sg run 실패 (rule %s): %s", rule.ID, stderr.String())
		}
	}

	if stdout.Len() == 0 {
		return nil, nil
	}
	return parseSGFindings(stdout.Bytes())
}

// scanWithRules는 규칙 목록을 사용하여 개별 스캔합니다.
func (s *Scanner) scanWithRules(ctx context.Context, rules []Rule, path string) ([]Finding, error) {
	timeout := s.cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	binary := s.cfg.SGBinary
	if binary == "" {
		binary = "sg"
	}

	allFindings := make([]Finding, 0)

	for _, rule := range rules {
		if rule.Pattern == "" || rule.Language == "" {
			continue
		}

		// F3: runSingleRule에서 defer cancel()로 context 누수 방지
		findings, err := s.runSingleRule(ctx, binary, rule, path, timeout)
		if err != nil {
			slog.Debug("규칙 실행 실패, 건너뜀", "rule", rule.ID, "error", err)
			continue
		}

		// 규칙 메타데이터 주입 (F1: Language 필드 포함)
		for i := range findings {
			if findings[i].RuleID == "" {
				findings[i].RuleID = rule.ID
			}
			if findings[i].Severity == "" {
				findings[i].Severity = rule.Severity
			}
			if findings[i].Message == "" {
				findings[i].Message = rule.Message
			}
			// Language는 항상 rule에서 주입 (찾은 파일의 언어가 아닌 규칙 대상 언어)
			findings[i].Language = rule.Language
			// Note 전파: Finding에 Note가 없으면 Rule.Note에서 복사 (REQ-UTIL-002-003)
			if findings[i].Note == "" {
				findings[i].Note = rule.Note
			}
			// Metadata 전파: Finding에 Metadata가 없으면 Rule.Metadata에서 복사 (REQ-UTIL-002-004)
			if findings[i].Metadata == nil {
				findings[i].Metadata = rule.Metadata
			}
		}

		allFindings = append(allFindings, findings...)
	}

	return allFindings, nil
}

// parseSGFindings는 sg JSON 출력을 Finding 슬라이스로 변환합니다.
func parseSGFindings(output []byte) ([]Finding, error) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return []Finding{}, nil
	}

	var matches []sgScanMatch
	if err := json.Unmarshal(trimmed, &matches); err != nil {
		return nil, fmt.Errorf("sg 출력 파싱: %w", err)
	}

	findings := make([]Finding, 0, len(matches))
	for _, m := range matches {
		f := Finding{
			RuleID:    m.RuleID,
			Severity:  m.Severity,
			Message:   m.Message,
			File:      m.File,
			Line:      m.Range.Start.Line + 1, // sg는 0-indexed, Finding은 1-indexed
			Column:    m.Range.Start.Column,
			EndLine:   m.Range.End.Line + 1,
			EndColumn: m.Range.End.Column,
			Note:      m.Note,
		}
		findings = append(findings, f)
	}

	return findings, nil
}
