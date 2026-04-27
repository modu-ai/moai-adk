// Package cli는 GitHub auth 명령을 제공합니다.
// Package cli provides GitHub auth command.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/github/auth"
)

// newAuthCmd는 auth 명령을 생성합니다.
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "LLM 제공업체 인증 관리 (Authenticate with LLM providers)",
		Long:  `Claude, Codex, Gemini, GLM 인증 토큰을 설정합니다.`,
	}

	// 서브커먼드 등록
	cmd.AddCommand(newAuthClaudeCmd())
	cmd.AddCommand(newAuthCodexCmd())
	cmd.AddCommand(newAuthGeminiCmd())
	cmd.AddCommand(newAuthGLMCmd())

	return cmd
}

// newAuthClaudeCmd는 claude 인증 서브커맨드를 생성합니다.
func newAuthClaudeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claude <token>",
		Short: "Claude API 토큰 설정 (Set Claude API token)",
		Long:  `Anthropic Claude API 토큰을 설정하여 GitHub Actions에서 Claude를 사용합니다.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := detectRepo()
			token := args[0]
			handler := auth.NewClaudeAuthHandler(nil)
			return handler.Setup(cmd.Context(), repo, token)
		},
	}
}

// newAuthCodexCmd는 codex 인증 서브커맨드를 생성합니다.
func newAuthCodexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "codex <token>",
		Short: "Codex API 토큰 설정 (Set Codex API token)",
		Long:  `OpenAI Codex API 토큰을 설정하여 GitHub Actions에서 Codex를 사용합니다.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := detectRepo()
			token := args[0]
			handler := auth.NewCodexAuthHandler(nil)
			return handler.Setup(cmd.Context(), repo, token, true)
		},
	}
}

// newAuthGeminiCmd는 gemini 인증 서브커맨드를 생성합니다.
func newAuthGeminiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gemini <token>",
		Short: "Gemini API 토큰 설정 (Set Gemini API token)",
		Long:  `Google Gemini API 토큰을 설정하여 GitHub Actions에서 Gemini를 사용합니다.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := detectRepo()
			token := args[0]
			handler := auth.NewGeminiAuthHandler(nil)
			return handler.Setup(cmd.Context(), repo, token)
		},
	}
}

// newAuthGLMCmd는 glm 인증 서브커맨드를 생성합니다.
func newAuthGLMCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "glm <token>",
		Short: "GLM API 토큰 설정 (Set GLM API token)",
		Long:  `Zhipu AI GLM API 토큰을 설정하여 GitHub Actions에서 GLM을 사용합니다.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := detectRepo()
			token := args[0]
			handler := auth.NewGLMAuthHandler(nil)
			return handler.Setup(cmd.Context(), repo, token)
		},
	}
}

// detectRepo는 현재 Git 리포지토리를 감지합니다.
func detectRepo() string {
	// TODO: 실제 Git 리포지토리 감지 로직 구현
	// 현재는 placeholder 반환
	return "owner/repo"
}
