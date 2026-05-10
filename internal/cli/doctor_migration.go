package cli

import (
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/migration"
	"github.com/spf13/cobra"
)

// runDoctorMigration는 'moai doctor --check migration'을 처리합니다.
// REQ-V3R2-RT-007-015: doctor 명령에 마이그레이션 체크를 추가합니다.
func runDoctorMigration(cmd *cobra.Command, args []string) (string, int) {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("FAIL: 현재 작업 디렉터리를 가져올 수 없음: %v", err), 2 // FAIL status
	}

	runner := migration.NewRunner(cwd)
	current, pending, _, err := runner.Status()
	if err != nil {
		return fmt.Sprintf("FAIL: 마이그레이션 상태 조회 실패: %v", err), 2
	}

	// 상태 판단: ok/warn/fail
	status := "ok"
	message := fmt.Sprintf("현재 버전: %d", current)
	if len(pending) > 0 {
		status = "warn"
		message = fmt.Sprintf("%d개 pending 마이그레이션 (버전: %v)", len(pending), pending)
	}

	output := fmt.Sprintf("Migration: %s\n", status)
	if status != "ok" {
		output += fmt.Sprintf("  Message: %s\n", message)
	}
	output += fmt.Sprintf("  현재 버전: %d\n", current)
	if len(pending) > 0 {
		output += fmt.Sprintf("  Pending: %v\n", pending)
	}

	exitCode := 0
	if status == "warn" {
		exitCode = 1
	} else if status == "fail" {
		exitCode = 2
	}

	return output, exitCode
}
