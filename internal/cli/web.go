package cli

// @MX:NOTE: [AUTO] web 서브커맨드는 MoAI Web Console(브라우저 기반 설정 CRUD)을 띄우는 thin 진입점이다.
// 실제 HTTP 서버/핸들러/검증/영속화는 internal/web 패키지가 소유한다. CLI는 플래그 파싱 + 프로젝트 루트 해석 후
// web.Run 에 위임한다 (cc.go 의 thin-command 패턴). 사용자 상호작용 프롬프트 호출 금지(orchestrator-only HARD,
// C-HRA-008 / internal/cli/CLAUDE.md).

import (
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/web"
	"github.com/spf13/cobra"
)

// webPort / webNoOpen / webNoReuse back the --port / --no-open / --no-reuse flags.
var (
	webPort    int
	webNoOpen  bool
	webNoReuse bool
)

// webRunFn is the test-injection seam for web.Run (mirrors findProjectRootFn /
// findPortHolder in this package). web.Run blocks until SIGINT, so tests that
// exercise runWeb end-to-end substitute a no-op here. The default is the real
// implementation.
var webRunFn = web.Run

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

By default, if the target port is already held by a stale moai instance, the
Console terminates that instance and rebinds. A non-moai (foreign) process is
never terminated — the Console reports an error and suggests --port. Pass
--no-reuse to disable the reclaim and fail on any port conflict.

Flags:
  --port <int>   TCP port to bind on 127.0.0.1 (default 3041)
  --no-open      Do not auto-open the browser
  --no-reuse     Do not reclaim the port from a stale moai instance

Examples:
  moai web                 # bind 127.0.0.1:3041 and open the browser
  moai web --port 9000     # bind a different port
  moai web --no-open       # start without launching a browser
  moai web --no-reuse      # fail instead of reclaiming a busy port`,
		GroupID: "tools",
		RunE:    runWeb,
	}
	cmd.Flags().IntVar(&webPort, "port", 3041, "TCP port to bind on 127.0.0.1")
	cmd.Flags().BoolVar(&webNoOpen, "no-open", false, "do not auto-open the browser")
	cmd.Flags().BoolVar(&webNoReuse, "no-reuse", false, "do not reclaim the port from a stale moai instance")
	return cmd
}

// runWeb resolves the project root and starts the Console server, blocking until
// SIGINT/SIGTERM. Exit-code discipline: returns a wrapped error (cobra → exit 1)
// when the project root cannot be found or the server fails to bind.
func runWeb(cmd *cobra.Command, _ []string) (err error) {
	// SPEC-SESSION-WORKTREE-001 M3/M4: session-worktree auto-entry + exit
	// disposal. The auto-entry runs BEFORE findProjectRootFn /
	// ensurePortFree / emitWorktreeAdvisory so that, on success, the process
	// chdir's into the worktree and subsequent project-root resolution lands
	// inside it (BI-2 / R1 mitigation). When the feature is OFF (default) the
	// wrapper returns "" and runWeb is byte-identical to the baseline
	// (REQ-SW-001). On fail-back (REQ-SW-004) or already-in-worktree skip
	// (REQ-SW-012) the wrapper returns "" and web continues in the shared
	// checkout.
	//
	// wtMaterialized tracks whether the process actually entered the worktree.
	// Only then is the shared-checkout collision hazard avoided by construction
	// (REQ-SW-013). A chdir failure after a successful materialization is a
	// fail-back: the worktree exists but this process cannot use it, so the
	// hazard is NOT avoided and wtMaterialized stays false.
	//
	// M4 (REQ-SW-008/009/010): the deferred cleanup disposes the materialized
	// worktree at subcommand exit. It honors three cases (default-manual
	// persist / clean-exit remove / non-clean preserve) + the dirty guard.
	// cleanExit is derived from this function's named error return (err ==
	// nil means exit 0). wtPath == "" makes cleanup a no-op.
	swCfg := loadSessionWorktreeConfig(cmd)
	wtMaterialized := false
	wtPath := enterSessionWorktree(swCfg, "web", cmd.ErrOrStderr())
	if wtPath != "" {
		if cerr := os.Chdir(wtPath); cerr != nil {
			// Chdir failure is a fail-back (REQ-SW-004): continue in the shared
			// checkout. The worktree was materialized but unusable from this
			// process; the user can enter it manually later. The advisory
			// therefore still fires (hazard not avoided by construction).
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
				"moai: session-worktree chdir into %s failed (%v); continuing in shared checkout for %q\n",
				wtPath, cerr, "web")
		} else {
			wtMaterialized = true
		}
	}
	defer func() {
		cleanupSessionWorktree(swCfg, wtPath, err == nil, cmd.ErrOrStderr())
	}()

	projectRoot, perr := findProjectRootFn()
	if perr != nil {
		return fmt.Errorf("moai web must run inside a MoAI project: %w", perr)
	}


	// web.Run 위임 전 대상 포트를 확보한다: stale moai 인스턴스는 회수하고
	// 외부 프로세스는 보호(에러). --no-reuse면 회수를 건너뛴다.
	// SPEC-SESSION-WORKTREE-001 (REQ-SW-015): the worktree is filesystem
	// isolation only — it does NOT virtualize the loopback port, so
	// ensurePortFree runs identically whether or not a worktree was entered.
	if err := ensurePortFree(cmd.ErrOrStderr(), webPort, !webNoReuse); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		"MoAI Web Console starting on http://127.0.0.1:%d (Ctrl+C to stop)\n", webPort)

	// SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-009): the Web Console runs in the
	// shared primary checkout; branch-changing work belongs in a worktree.
	// SPEC-SESSION-WORKTREE-001 (REQ-SW-013): when the session-worktree was
	// materialized above, the collision hazard was avoided by construction and
	// the advisory is SUPPRESSED. When the feature is OFF OR materialization
	// fell back (REQ-SW-004) OR chdir failed, the advisory fires unchanged.
	if !wtMaterialized {
		emitWorktreeAdvisory(cmd.OutOrStdout(), projectRoot)
	}

	return webRunFn(cmd.Context(), web.Config{
		Port:        webPort,
		NoOpen:      webNoOpen,
		ProjectRoot: projectRoot,
		// Project-scoped read (SPEC-PROFILE-MEMORY-001 REQ-PM-024) to match the
		// project-scoped write wired in web.newApp. CLAUDE_CONFIG_DIR still wins
		// when set, so a console launched inside a `moai cc -p X` session is
		// unaffected; this only decides the bare-launch case.
		ProfileName: profile.GetCurrentNameForProject(projectRoot),
	})
}

func init() {
	rootCmd.AddCommand(newWebCmd())
}
