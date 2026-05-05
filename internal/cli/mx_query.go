package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// newMxQueryCmd는 'moai mx query' 서브커맨드를 생성합니다.
// SPEC-V3R2-SPC-004 REQ-SPC-004-001에 정의된 모든 필터 플래그를 지원합니다.
func newMxQueryCmd() *cobra.Command {
	var specID string
	var kind string
	var fanInMin int
	var danger string
	var filePrefix string
	var since string
	var limit int
	var offset int
	var format string
	var includeTests bool

	cmd := &cobra.Command{
		Use:   "query",
		Short: "@MX TAG 사이드카 인덱스 조회",
		Long: `@MX TAG 사이드카 인덱스에서 태그를 구조화된 형식으로 조회합니다.

여러 필터를 AND 조합으로 적용하며 JSON(기본), 테이블, 마크다운 형식으로 출력합니다.

예시:
  moai mx query --spec SPEC-AUTH-001 --kind anchor
  moai mx query --fan-in-min 3 --kind anchor
  moai mx query --danger concurrency
  moai mx query --file-prefix internal/auth/ --format table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 프로젝트 루트 확인
			projectRoot, err := findProjectRootFn()
			if err != nil {
				return fmt.Errorf("프로젝트 루트 탐색 실패: %w", err)
			}

			// KIND 유효성 검증 (REQ-SPC-004-041)
			if kind != "" {
				validKinds := map[string]bool{
					"note": true, "warn": true, "anchor": true,
					"todo": true, "legacy": true,
					"NOTE": true, "WARN": true, "ANCHOR": true,
					"TODO": true, "LEGACY": true,
				}
				if !validKinds[kind] {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "InvalidQuery: --kind 값 '%s'이 잘못되었습니다. 허용 값: note, warn, anchor, todo, legacy\n", kind)
					return &mx.InvalidQueryError{
						Field:   "kind",
						Value:   kind,
						Message: "허용 값: note, warn, anchor, todo, legacy",
					}
				}
			}

			// SINCE 파싱
			var sinceTime time.Time
			if since != "" {
				parsed, err := time.Parse(time.RFC3339, since)
				if err != nil {
					return &mx.InvalidQueryError{
						Field:   "since",
						Value:   since,
						Message: "RFC3339 형식 필요 (예: 2006-01-02T15:04:05Z)",
					}
				}
				sinceTime = parsed
			}

			// FORMAT 유효성 검증
			validFormats := map[string]bool{"json": true, "table": true, "markdown": true}
			if format != "" && !validFormats[format] {
				return &mx.InvalidQueryError{
					Field:   "format",
					Value:   format,
					Message: "허용 값: json, table, markdown",
				}
			}
			if format == "" {
				format = "json"
			}

			// 사이드카 경로 확인
			stateDir := filepath.Join(projectRoot, ".moai", "state")
			mgr := mx.NewManager(stateDir)

			// 사이드카 파일 존재 확인 (REQ-SPC-004-013)
			sidecarPath := filepath.Join(stateDir, mx.SidecarFileName)
			if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
					"SidecarUnavailable: 사이드카 인덱스가 없습니다 — '/moai mx --full' 를 실행하여 인덱스를 재빌드하세요\n")
				return fmt.Errorf("SidecarUnavailable: 사이드카 인덱스 없음")
			}

			// Resolver 생성
			resolver := mx.NewResolver(mgr)

			// KIND 문자열을 TagKind로 변환
			var tagKind mx.TagKind
			if kind != "" {
				switch kind {
				case "note", "NOTE":
					tagKind = mx.MXNote
				case "warn", "WARN":
					tagKind = mx.MXWarn
				case "anchor", "ANCHOR":
					tagKind = mx.MXAnchor
				case "todo", "TODO":
					tagKind = mx.MXTodo
				case "legacy", "LEGACY":
					tagKind = mx.MXLegacy
				}
			}

			// 쿼리 실행
			query := mx.Query{
				SpecID:       specID,
				Kind:         tagKind,
				FanInMin:     fanInMin,
				Danger:       danger,
				FilePrefix:   filePrefix,
				Since:        sinceTime,
				Limit:        limit,
				Offset:       offset,
				IncludeTests: includeTests,
			}

			result, err := resolver.Resolve(query)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
				return err
			}

			// 출력 형식별 렌더링
			switch format {
			case "table":
				_, _ = fmt.Fprint(cmd.OutOrStdout(), mx.FormatTable(result))

			case "markdown":
				_, _ = fmt.Fprint(cmd.OutOrStdout(), mx.FormatMarkdown(result))

			default: // json
				data, err := json.MarshalIndent(result.Tags, "", "  ")
				if err != nil {
					return fmt.Errorf("JSON 직렬화 실패: %w", err)
				}

				if result.TruncationNotice {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
						"TruncationNotice: 전체 %d개 중 %d개만 표시됩니다. --limit 플래그로 더 보기 가능.\n",
						result.TotalCount, len(result.Tags))
				}

				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			}

			return nil
		},
	}

	// REQ-SPC-004-001 플래그 등록
	cmd.Flags().StringVar(&specID, "spec", "", "SPEC ID 필터 (예: SPEC-AUTH-001)")
	cmd.Flags().StringVar(&kind, "kind", "", "TAG 종류 필터 (note|warn|anchor|todo|legacy)")
	cmd.Flags().IntVar(&fanInMin, "fan-in-min", 0, "최소 fan-in 수 필터 (ANCHOR 전용)")
	cmd.Flags().StringVar(&danger, "danger", "", "위험 카테고리 필터 (WARN 전용)")
	cmd.Flags().StringVar(&filePrefix, "file-prefix", "", "파일 경로 접두사 필터")
	cmd.Flags().StringVar(&since, "since", "", "최소 LastSeenAt 시간 필터 (RFC3339 형식)")
	cmd.Flags().IntVar(&limit, "limit", 0, "최대 반환 수 (기본 100, 최대 10000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "페이지네이션 오프셋 (기본 0)")
	cmd.Flags().StringVar(&format, "format", "json", "출력 형식 (json|table|markdown)")
	cmd.Flags().BoolVar(&includeTests, "include-tests", false, "fan-in 계산 시 테스트 파일 참조 포함")

	return cmd
}
