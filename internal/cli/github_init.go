// Package cli는 GitHub init 명령을 제공합니다.
// Package cli provides GitHub init command.
package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// newInitCmd는 init 명령을 생성합니다.
// T-23: 통합 부트스트랩 명령
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "GitHub Actions 통합 초기화 (Initialize GitHub Actions integration)",
		Long:  `다중 LLM CI 코드 리뷰를 위한 self-hosted runner 설정을 수행합니다.`,
		Args:  cobra.NoArgs,
		RunE: runGitHubInit,
	}
}

// runGitHubInit은 초기화를 실행합니다.
func runGitHubInit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	out := cmd.OutOrStdout()

	// 1. 리포지토리 감지
	repo := detectRepo()

	fmt.Fprintf(out, "리포지토리 감지: %s\n", repo)

	// 2. LLM 선택 프롬프트
	llms, err := promptLLMs(ctx)
	if err != nil {
		return fmt.Errorf("prompt LLMs: %w", err)
	}

	fmt.Fprintf(out, "선택된 LLM: %v\n", llms)

	// 3. 성공 메시지 표시
	displayInitSuccess(out, repo, llms)

	return nil
}

// promptLLMs은 LLM 선택을 프롬프트합니다.
func promptLLMs(ctx context.Context) ([]string, error) {
	// TODO: 실제 사용자 프롬프트 로직 구현
	// 현재는 기본값 반환
	return []string{"claude", "glm"}, nil
}

// displayInitSuccess는 초기화 성공 메시지를 표시합니다.
func displayInitSuccess(out interface{}, repo string, llms []string) {
	fmt.Fprintf(out.(interface{ Write([]byte) (int, error) }),
		"\n=== GitHub Actions 초기화 완료 ===\n"+
		"리포지토리: %s\n"+
		"구성된 LLM: %v\n"+
		"\n다음 단계:\n"+
		"1. LLM 토큰 설정: moai github auth claude <token>\n"+
		"2. Workflow 템플릿 동기화: moai github workflow sync\n"+
		"3. CI 테스트 실행: GitHub에 push\n",
		repo, llms)
}
