package cli

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/github"
	"github.com/modu-ai/moai-adk/internal/github/auth"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "LLM 제공업체 인증 관리 (Authenticate with LLM providers)",
		Long:  `Claude, OpenAI, Gemini, GLM 인증 토큰을 설정합니다.`,
	}

	cmd.AddCommand(newAuthClaudeCmd())
	cmd.AddCommand(newAuthCodexCmd())
	cmd.AddCommand(newAuthGeminiCmd())
	cmd.AddCommand(newAuthGLMCmd())

	return cmd
}

func newAuthClaudeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claude <token>",
		Short: "Claude API 토큰 설정 (Set Claude API token)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := detectRepo()
			if err != nil {
				return err
			}
			return auth.NewClaudeAuthHandler(newSecretSetter(cmd)).Setup(cmd.Context(), repo, args[0])
		},
	}
}

func newAuthCodexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "codex <token>",
		Short: "OpenAI (Codex) API 토큰 설정 (Set OpenAI API token)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := detectRepo()
			if err != nil {
				return err
			}
			return auth.NewCodexAuthHandler(newSecretSetter(cmd)).Setup(cmd.Context(), repo, args[0], true)
		},
	}
}

func newAuthGeminiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gemini <token>",
		Short: "Gemini API 토큰 설정 (Set Gemini API token)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := detectRepo()
			if err != nil {
				return err
			}
			return auth.NewGeminiAuthHandler(newSecretSetter(cmd)).Setup(cmd.Context(), repo, args[0])
		},
	}
}

func newAuthGLMCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "glm <token>",
		Short: "GLM API 토큰 설정 (Set GLM API token)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := detectRepo()
			if err != nil {
				return err
			}
			return auth.NewGLMAuthHandler(newSecretSetter(cmd)).Setup(cmd.Context(), repo, args[0])
		},
	}
}

func newSecretSetter(cmd *cobra.Command) auth.SecretSetter {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		return &dryRunSecretSetter{}
	}
	return github.NewRealGHSecretManager()
}

type dryRunSecretSetter struct{}

func (d *dryRunSecretSetter) SetSecret(_ context.Context, repo, name, value string) error {
	fmt.Printf("[DRY-RUN] gh secret set %s -R %s (value: %s)\n", name, repo, github.MaskSecret(value))
	return nil
}

// remoteRepos matches SSH (git@github.com:owner/repo) and HTTPS GitHub remote URLs.
var remoteRepos = regexp.MustCompile(`github\.com[:/]([^/]+)/([^/.]+)`)

func detectRepo() (string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("git remote: %w (are you in a git repository?)", err)
	}
	m := remoteRepos.FindStringSubmatch(strings.TrimSpace(string(out)))
	if len(m) < 3 {
		return "", fmt.Errorf("cannot parse GitHub owner/repo from remote: %s", strings.TrimSpace(string(out)))
	}
	return m[1] + "/" + m[2], nil
}
