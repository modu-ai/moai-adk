package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// newSpecLintCmd는 'moai spec lint' 서브커맨드를 생성한다.
// SPEC-V3R2-SPC-003 구현.
//
// @MX:NOTE: [AUTO] newSpecLintCmd는 cobra 패턴을 따르는 spec lint CLI 엔트리포인트이다.
func newSpecLintCmd() *cobra.Command {
	var (
		jsonOutput  bool
		sarifOutput bool
		strict      bool
		format      string
	)

	cmd := &cobra.Command{
		Use:   "lint [spec.md...]",
		Short: "Lint SPEC documents for EARS compliance and structural validity",
		Long: `Validate SPEC documents against:
  - EARS modality compliance (SHALL, WHEN, WHILE, WHERE, IF)
  - REQ ID uniqueness
  - AC→REQ coverage (100% required)
  - Frontmatter schema validation
  - Dependency DAG (no cycles, all deps exist)
  - Out of Scope section presence
  - Zone registry cross-references

Exit codes:
  0 = success (no errors)
  1 = errors found
  2 = linter crash
  3 = invalid arguments`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 인수 검증
			if jsonOutput && sarifOutput {
				return fmt.Errorf("--json과 --sarif는 함께 사용할 수 없음")
			}

			// BaseDir 결정: 현재 작업 디렉토리의 .moai/specs/ 우선
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("작업 디렉토리 확인 오류: %w", err)
			}

			baseDir := detectBaseDir(cwd)
			registryPath := detectRegistryPath(cwd)

			linterOpts := spec.LinterOptions{
				RegistryPath: registryPath,
				BaseDir:      baseDir,
				Strict:       strict,
			}

			linter := spec.NewLinter(linterOpts)
			report, lintErr := linter.Lint(args)
			if lintErr != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "linter 오류: %v\n", lintErr)
				os.Exit(2)
			}

			// 출력 형식 선택
			switch {
			case jsonOutput:
				data, marshalErr := report.ToJSON()
				if marshalErr != nil {
					return fmt.Errorf("JSON 직렬화 오류: %w", marshalErr)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))

			case sarifOutput:
				data, marshalErr := report.ToSARIF()
				if marshalErr != nil {
					return fmt.Errorf("SARIF 직렬화 오류: %w", marshalErr)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))

			default:
				// human-readable table (default 또는 --format table)
				_ = format // format=table은 기본값과 동일
				printTable(cmd, report)
			}

			if report.HasErrors() {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "JSON 형식으로 출력")
	cmd.Flags().BoolVar(&sarifOutput, "sarif", false, "SARIF 2.1.0 형식으로 출력")
	cmd.Flags().BoolVar(&strict, "strict", false, "경고를 오류로 처리")
	cmd.Flags().StringVar(&format, "format", "table", "출력 형식 (table)")

	return cmd
}

// printTable은 findings를 human-readable 테이블 형식으로 출력한다.
func printTable(cmd *cobra.Command, report *spec.Report) {
	out := cmd.OutOrStdout()
	if len(report.Findings) == 0 {
		_, _ = fmt.Fprintln(out, "✓ No findings — all SPEC documents are valid")
		return
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "SEVERITY\tCODE\tFILE\tLINE\tMESSAGE")
	_, _ = fmt.Fprintln(w, "--------\t----\t----\t----\t-------")

	for _, f := range report.Findings {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			strings.ToUpper(string(f.Severity)),
			f.Code,
			f.File,
			f.Line,
			f.Message,
		)
	}
	_ = w.Flush()

	// 요약 출력
	var errCount, warnCount int
	for _, f := range report.Findings {
		switch f.Severity {
		case spec.SeverityError:
			errCount++
		case spec.SeverityWarning:
			warnCount++
		}
	}
	_, _ = fmt.Fprintf(out, "\n%d error(s), %d warning(s)\n", errCount, warnCount)
}

// detectBaseDir는 프로젝트 기준 디렉토리를 결정한다.
// .moai/specs/ 디렉토리가 있으면 그것을 기준으로 사용한다.
func detectBaseDir(cwd string) string {
	specsDir := filepath.Join(cwd, ".moai", "specs")
	if _, err := os.Stat(specsDir); err == nil {
		return specsDir
	}
	return cwd
}

// detectRegistryPath는 zone registry 파일 경로를 탐지한다.
func detectRegistryPath(cwd string) string {
	candidate := filepath.Join(cwd, ".claude", "rules", "moai", "core", "zone-registry.md")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}
