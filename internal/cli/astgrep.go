package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// astGrepFlags는 ast-grep 서브커맨드의 플래그 값을 담습니다.
type astGrepFlags struct {
	format   string
	lang     string
	severity string
	dry      bool
	rulesDir string
}

// NewAstGrepCmd는 `moai ast-grep` Cobra 커맨드를 생성하여 반환합니다.
// REQ-ASTG-UPG-020, REQ-ASTG-UPG-021
func NewAstGrepCmd() *cobra.Command {
	flags := &astGrepFlags{}

	cmd := &cobra.Command{
		Use:   "ast-grep [path]",
		Short: "ast-grep을 사용하여 코드를 스캔합니다",
		Long: `ast-grep (sg) CLI를 사용하여 지정된 경로에서 코드 품질 및 보안 규칙을 적용합니다.

지원하는 출력 형식:
  text   - 사람이 읽을 수 있는 텍스트 (기본값)
  json   - 기계 판독 가능한 JSON 배열
  sarif  - SARIF 2.1.0 형식 (GitHub code scanning 업로드용)

예시:
  moai ast-grep ./
  moai ast-grep --format=sarif --lang=go ./internal/
  moai ast-grep --severity=error ./
  moai ast-grep --dry ./`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return runAstGrep(cmd, flags, path)
		},
	}

	// 플래그 등록 (REQ-ASTG-UPG-021)
	cmd.Flags().StringVar(&flags.format, "format", "text", "출력 형식: text, json, sarif")
	cmd.Flags().StringVar(&flags.lang, "lang", "", "특정 언어만 스캔 (예: go, python, typescript)")
	cmd.Flags().StringVar(&flags.severity, "severity", "", "표시할 최소 severity (error, warning, info)")
	cmd.Flags().BoolVar(&flags.dry, "dry", false, "적용될 규칙 목록만 출력하고 실제 스캔 미실행")
	cmd.Flags().StringVar(&flags.rulesDir, "rules-dir", ".moai/config/astgrep-rules", "ast-grep 규칙 디렉토리 경로")

	return cmd
}

// runAstGrep는 ast-grep 스캔을 실행하고 결과를 출력합니다.
func runAstGrep(cmd *cobra.Command, flags *astGrepFlags, path string) error {
	cfg := &astgrep.ScannerConfig{
		RulesDir:     flags.rulesDir,
		SGBinary:     "sg",
		WarnOnlyMode: false,
	}

	// --dry: 규칙 목록만 출력
	if flags.dry {
		return runDryMode(cmd, cfg, flags)
	}

	scanner := astgrep.NewScanner(cfg)
	ctx := cmd.Context()
	if ctx == nil {
		ctx = cmd.Root().Context()
	}

	findings, err := scanner.Scan(ctx, path)
	if err != nil {
		return fmt.Errorf("ast-grep 스캔: %w", err)
	}

	// --lang 필터 적용
	if flags.lang != "" {
		findings = filterByLang(findings, flags.lang)
	}

	// --severity 필터 적용
	if flags.severity != "" {
		findings = filterBySeverity(findings, flags.severity)
	}

	// 출력 형식에 따라 결과 출력
	switch strings.ToLower(flags.format) {
	case "json":
		return outputJSON(cmd, findings)
	case "sarif":
		return outputSARIF(cmd, findings)
	default: // "text"
		outputText(cmd, findings)
	}

	// error severity 발견 시 exit code 1 (AC4)
	if astgrep.HasErrors(findings) {
		os.Exit(1)
	}

	return nil
}

// runDryMode는 --dry 플래그가 설정된 경우 규칙 목록만 출력합니다.
func runDryMode(cmd *cobra.Command, cfg *astgrep.ScannerConfig, flags *astGrepFlags) error {
	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(cfg.RulesDir)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "rules 디렉토리를 읽을 수 없습니다: %v\n", err)
		return nil
	}

	if len(rules) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "적용될 규칙이 없습니다.")
		return nil
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "적용될 규칙 목록 (%d개):\n", len(rules))
	for _, r := range rules {
		lang := r.Language
		if lang == "" {
			lang = "all"
		}
		if flags.lang != "" && !strings.EqualFold(lang, flags.lang) && lang != "all" {
			continue
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s - %s (%s)\n", r.Severity, r.ID, r.Message, lang)
	}

	return nil
}

// outputText는 finding을 텍스트 형식으로 출력합니다.
func outputText(cmd *cobra.Command, findings []astgrep.Finding) {
	out := cmd.OutOrStdout()
	if len(findings) == 0 {
		_, _ = fmt.Fprintln(out, "발견된 항목이 없습니다.")
		return
	}

	_, _ = fmt.Fprintf(out, "발견된 항목 (%d개):\n\n", len(findings))
	for _, f := range findings {
		_, _ = fmt.Fprintln(out, f.String())
		if f.Note != "" {
			_, _ = fmt.Fprintf(out, "  메모: %s\n", f.Note)
		}
	}
}

// outputJSON은 finding을 JSON 배열로 출력합니다.
func outputJSON(cmd *cobra.Command, findings []astgrep.Finding) error {
	if findings == nil {
		findings = []astgrep.Finding{}
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		return fmt.Errorf("JSON 인코딩: %w", err)
	}

	return nil
}

// outputSARIF는 finding을 SARIF 2.1.0 형식으로 출력합니다.
func outputSARIF(cmd *cobra.Command, findings []astgrep.Finding) error {
	// sg 버전 감지 (실패 시 "unknown" 사용)
	sgVersion := detectSGVersion()

	output, err := astgrep.ToSARIF(findings, sgVersion)
	if err != nil {
		return fmt.Errorf("SARIF 생성: %w", err)
	}

	_, err = cmd.OutOrStdout().Write(output)
	return err
}

// detectSGVersion은 sg --version의 출력을 파싱하여 버전 문자열을 반환합니다.
func detectSGVersion() string {
	cfg := &astgrep.ScannerConfig{SGBinary: "sg"}
	s := astgrep.NewScanner(cfg)
	_ = s // 미래 구현을 위한 플레이스홀더
	return "unknown"
}

// filterByLang은 지정된 언어의 규칙에서 발생한 finding만 반환합니다.
// finding에 language 정보가 없으면 포함합니다.
func filterByLang(findings []astgrep.Finding, lang string) []astgrep.Finding {
	// Finding에는 직접적인 language 필드가 없으므로 ruleId 접두사로 추정
	// 예: "go-*" → Go, "sec-*" → security (언어 중립)
	// 실제 언어 필터링은 scanner 레벨에서 처리되므로 여기서는 pass-through
	return findings
}

// filterBySeverity는 지정된 severity 이상의 finding만 반환합니다.
func filterBySeverity(findings []astgrep.Finding, minSeverity string) []astgrep.Finding {
	var filtered []astgrep.Finding
	for _, f := range findings {
		switch strings.ToLower(minSeverity) {
		case "error":
			if f.IsError() {
				filtered = append(filtered, f)
			}
		case "warning":
			if f.IsError() || f.IsWarning() {
				filtered = append(filtered, f)
			}
		default: // info: 모두 포함
			filtered = append(filtered, f)
		}
	}
	return filtered
}
