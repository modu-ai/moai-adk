// Package cli는 GitHub status 명령을 제공합니다.
// Package cli provides GitHub status command.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/github/runner"
)

// newStatusCmd는 status 명령을 생성합니다.
func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "GitHub Actions 통합 상태 확인 (Check integration status)",
		Long:  `Runner 버전, 인증 토큰, 워크플로우 상태를 표시합니다.`,
		Args:  cobra.NoArgs,
		RunE: runGitHubStatus,
	}
}

// runGitHubStatus는 상태 확인을 실행합니다.
func runGitHubStatus(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	out := cmd.OutOrStdout()

	// 1. Runner 버전 확인 (T-05)
	runnerStatus, err := checkRunnerVersion(ctx)
	if err != nil {
		return fmt.Errorf("check runner version: %w", err)
	}

	// 2. 포맷된 상태 카드 표시
	displayStatusCard(out, runnerStatus)

	return nil
}

// checkRunnerVersion은 runner 버전을 확인합니다.
func checkRunnerVersion(ctx context.Context) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	ghRunnerDir := filepath.Join(homeDir, "actions-runner")

	ghClient := runner.NewFileSystemGitHubClient()
	checker := runner.NewVersionChecker(ghRunnerDir, ghClient)
	result, err := checker.CheckVersion(ctx)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("설치 버전: %s\n", result.InstalledVersion))
	sb.WriteString(fmt.Sprintf("최신 버전: %s\n", result.LatestVersion))
	sb.WriteString(fmt.Sprintf("경과 일수: %d일\n", result.DaysOld))
	sb.WriteString(fmt.Sprintf("상태: %s - %s", result.Status, result.Message))

	return sb.String(), nil
}

// displayStatusCard는 상태 카드를 표시합니다.
func displayStatusCard(out interface{}, runnerStatus string) {
	fmt.Fprintf(out.(interface{ Write([]byte) (int, error) }), 
		"=== GitHub Actions 상태 ===\n\n"+
		"[Runner]\n%s\n\n"+
		"[Auth]\n토큰 확인 기능: moai github auth <llm> <token>로 설정\n\n"+
		"[Workflow]\n템플릿 확인 기능: 구현 예정\n",
		runnerStatus)
}
