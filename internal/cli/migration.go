package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/migration"
	"github.com/spf13/cobra"
)

// migrationCmd는 'moai migration' cobra command 그룹입니다.
// REQ-V3R2-RT-007-015, REQ-V3R2-RT-007-040, REQ-V3R2-RT-007-041, REQ-V3R2-RT-007-024.
var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "마이그레이션을 관리합니다 (run, status, rollback)",
	Long: `마이그레이션 관리 도구입니다.

'run': pending migrations를 실행합니다.
'status': 현재 마이그레이션 상태를 표시합니다.
'rollback': 특정 버전으로 롤백합니다 (일부 마이그레이션은 롤백 불가능).

참고: 'moai migrate agency'와 혼동하지 마세요. 'migrate agency'는
일회성 agency 마이그레이션 명령이며, 이 명령은 버전 관리 마이그레이션
프레임워크입니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// migrationRunCmd는 pending migrations를 실행합니다.
// REQ-V3R2-RT-007-040: 'moai migration run' subcommand.
var migrationRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Pending migrations를 실행합니다",
	Long: `현재 프로젝트에서 아직 실행되지 않은 마이그레이션을 순서대로 실행합니다.

Session-start 훅에서도 자동으로 실행되지만, 이 명령으로 수동으로 실행할 수도 있습니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("현재 작업 디렉터리를 가져올 수 없습니다: %w", err)
		}

		runner := migration.NewRunner(cwd)
		ctx := context.Background()
		applied, err := runner.Apply(ctx)
		if err != nil {
			return fmt.Errorf("마이그레이션 실행 실패: %w", err)
		}

		if len(applied) == 0 {
			fmt.Println("실행할 pending 마이그레이션이 없습니다.")
			return nil
		}

		fmt.Printf("성공: %d개 마이그레이션 적용됨 (버전: %v)\n", len(applied), applied)
		return nil
	},
}

// migrationStatusCmd는 현재 마이그레이션 상태를 표시합니다.
// REQ-V3R2-RT-007-015: 'moai migration status [--json]' subcommand.
var migrationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "마이그레이션 상태를 표시합니다",
	Long: `현재 프로젝트의 마이그레이션 상태를 표시합니다.

출력 정보:
- 현재 버전 (가장 최근 적용된 마이그레이션)
- Pending 마이그레이션 목록
- 가장 최근 적용된 마이그레이션 상세 정보`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonFlag, _ := cmd.Flags().GetBool("json")

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("현재 작업 디렉터리를 가져올 수 없습니다: %w", err)
		}

		runner := migration.NewRunner(cwd)
		current, pending, lastApplied, err := runner.Status()
		if err != nil {
			return fmt.Errorf("마이그레이션 상태 조회 실패: %w", err)
		}

		if jsonFlag {
			// JSON 출력 (REQ-V3R2-RT-007-041)
			output := map[string]any{
				"current_version": current,
				"pending":         pending,
				"last_applied":    lastApplied,
			}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("JSON 마샬링 실패: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Human-readable 출력
		fmt.Printf("현재 버전: %d\n", current)
		if len(pending) > 0 {
			fmt.Printf("Pending 마이그레이션 (%d개): %v\n", len(pending), pending)
		} else {
			fmt.Println("Pending 마이그레이션 없음 (최신 상태)")
		}
		if lastApplied != nil {
			fmt.Printf("최근 적용: %s (버전 %d)\n", lastApplied.Name, lastApplied.Version)
		}

		return nil
	},
}

// migrationRollbackCmd는 특정 버전으로 롤백합니다.
// REQ-V3R2-RT-007-024: 'moai migration rollback <version>' subcommand.
var migrationRollbackCmd = &cobra.Command{
	Use:   "rollback <version>",
	Short: "특정 버전으로 롤백합니다",
	Long: `지정된 버전으로 롤백합니다.

주의: 일부 마이그레이션(특히 중요 버그 수정)은 롤백 불가능으로
표시될 수 있습니다. 롤백은 version-file을 이전 버전으로 되돌리고
해당 마이그레이션의 Rollback 함수를 실행합니다.

예시:
  moai migration rollback 0  # 모든 마이그레이션 롤백 (초기 상태)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetVersion int
		_, err := fmt.Sscanf(args[0], "%d", &targetVersion)
		if err != nil {
			return fmt.Errorf("잘못된 버전 번호: %s: %w", args[0], err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("현재 작업 디렉터리를 가져올 수 없습니다: %w", err)
		}

		runner := migration.NewRunner(cwd)
		if err := runner.Rollback(targetVersion); err != nil {
			return fmt.Errorf("롤백 실패: %w", err)
		}

		fmt.Printf("성공: 버전 %d로 롤백됨\n", targetVersion)
		return nil
	},
}

func init() {
	// migration 하위 명령 등록
	migrationCmd.AddCommand(migrationRunCmd)
	migrationCmd.AddCommand(migrationStatusCmd)
	migrationCmd.AddCommand(migrationRollbackCmd)

	// status 명령에 --json 플래그 추가
	migrationStatusCmd.Flags().Bool("json", false, "JSON 형식으로 출력")

	// rootCmd에 migration 그룹 등록은 root.go에서 수행
}
