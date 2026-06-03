package cli

// @MX:NOTE: [AUTO] web 서브커맨드는 MoAI Web Console(브라우저 기반 설정 CRUD)을 띄우는 thin 진입점이다.
// 실제 HTTP 서버/핸들러/검증/영속화는 internal/web 패키지가 소유한다. CLI는 플래그 파싱 + 프로젝트 루트 해석 후
// web.Run 에 위임한다 (cc.go / brain.go 의 thin-command 패턴). 사용자 상호작용 프롬프트 호출 금지(orchestrator-only HARD,
// C-HRA-008 / internal/cli/CLAUDE.md).

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/web"
	"github.com/spf13/cobra"
)

// webPort / webNoOpen back the --port / --no-open flags.
var (
	webPort   int
	webNoOpen bool
)

// newWebCmd constructs the `web` subcommand. Factory form (mirrors
// newNewCmd/newCleanCmd) so tests can build an isolated command instance.
func newWebCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web [--port N] [--no-open]",
		Short: "Launch the MoAI Web Console (browser-based settings editor)",
		Long: `Launch the MoAI Web Console — a local, browser-based editor for your
MoAI settings (profile preferences plus the project user / language /
statusline sections).

The Console binds to loopback only (127.0.0.1) and reuses the same validation
and persistence logic as the terminal profile wizard. There is no external
database, no auth, and no network exposure.

Flags:
  --port <int>   TCP port to bind on 127.0.0.1 (default 8080)
  --no-open      Do not auto-open the browser

Examples:
  moai web                 # bind 127.0.0.1:8080 and open the browser
  moai web --port 9000     # bind a different port
  moai web --no-open       # start without launching a browser`,
		GroupID: "tools",
		RunE:    runWeb,
	}
	cmd.Flags().IntVar(&webPort, "port", 8080, "TCP port to bind on 127.0.0.1")
	cmd.Flags().BoolVar(&webNoOpen, "no-open", false, "do not auto-open the browser")
	return cmd
}

// runWeb resolves the project root and starts the Console server, blocking until
// SIGINT/SIGTERM. Exit-code discipline: returns a wrapped error (cobra → exit 1)
// when the project root cannot be found or the server fails to bind.
func runWeb(cmd *cobra.Command, _ []string) error {
	projectRoot, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("moai web must run inside a MoAI project: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		"MoAI Web Console starting on http://127.0.0.1:%d (Ctrl+C to stop)\n", webPort)

	return web.Run(cmd.Context(), web.Config{
		Port:        webPort,
		NoOpen:      webNoOpen,
		ProjectRoot: projectRoot,
		ProfileName: profile.GetCurrentName(),
	})
}

func init() {
	rootCmd.AddCommand(newWebCmd())
}
