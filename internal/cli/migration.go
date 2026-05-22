package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/migration"
	"github.com/spf13/cobra"
)

// migrationCmd is the 'moai migration' cobra command group.
// REQ-V3R2-RT-007-015, REQ-V3R2-RT-007-040, REQ-V3R2-RT-007-041, REQ-V3R2-RT-007-024.
var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Manage migrations (run, status, rollback)",
	Long: `Migration management tool.

'run': execute pending migrations.
'status': display the current migration status.
'rollback': roll back to a specific version (some migrations are not rollback-capable).

Note: do not confuse this with 'moai migrate agency'. 'migrate agency' is a
one-off agency migration command; this command is the version-controlled
migration framework.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// migrationRunCmd executes pending migrations.
// REQ-V3R2-RT-007-040: 'moai migration run' subcommand.
var migrationRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute pending migrations",
	Long: `Run, in order, every migration that has not yet been applied in the current project.

Migrations also run automatically via the session-start hook; this command lets you trigger them manually.`,
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

// migrationStatusCmd displays the current migration status.
// REQ-V3R2-RT-007-015: 'moai migration status [--json]' subcommand.
var migrationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the migration status",
	Long: `Display the current migration status for the project.

Output fields:
- Current version (most recently applied migration)
- List of pending migrations
- Details of the most recently applied migration`,
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
			// JSON output (REQ-V3R2-RT-007-041)
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

		// Human-readable output
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

// migrationRollbackCmd rolls back to a specific version.
// REQ-V3R2-RT-007-024: 'moai migration rollback <version>' subcommand.
var migrationRollbackCmd = &cobra.Command{
	Use:   "rollback <version>",
	Short: "Roll back to a specific version",
	Long: `Roll back to the specified version.

Caution: some migrations (especially critical bug fixes) may be marked as
non-rollback-capable. Rollback reverts the version-file to the previous
version and executes the Rollback function of the affected migration.

Example:
  moai migration rollback 0  # Roll back every migration (initial state)`,
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
	// Register migration subcommands
	migrationCmd.AddCommand(migrationRunCmd)
	migrationCmd.AddCommand(migrationStatusCmd)
	migrationCmd.AddCommand(migrationRollbackCmd)

	// Add the --json flag to the status command
	migrationStatusCmd.Flags().Bool("json", false, "JSON 형식으로 출력")

	// Registration of the migration group on rootCmd is performed by root.go
}
