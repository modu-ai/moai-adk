package cli

import (
	"fmt"

	"github.com/spf13/cobra"
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
			// RED 단계: 미구현 stub
			// GREEN 단계에서 실제 구현으로 교체됩니다
			_ = specID
			_ = kind
			_ = fanInMin
			_ = danger
			_ = filePrefix
			_ = since
			_ = limit
			_ = offset
			_ = format
			_ = includeTests
			return fmt.Errorf("not implemented")
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
