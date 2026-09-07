package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// statuslineRenderBudget bounds the whole statusline build. It sits under Claude Code's
// render budget so a slow or rate-limited collector cannot stall the status bar.
const statuslineRenderBudget = 800 * time.Millisecond

// statuslineRefreshGitHub and statuslineBoardRoot back the detached refresh
// entry point. The statusline render path never calls the network; when its
// GitHub cache ages out it re-invokes this binary with these flags and returns
// immediately, so the fetch happens in a child that nothing waits on.
var (
	statuslineRefreshGitHub bool
	statuslineRefreshLanded bool
	statuslineBoardRoot     string
)

// StatuslineCmd is the statusline command.
var StatuslineCmd = &cobra.Command{
	Use:    "statusline",
	Short:  "Render statusline for Claude Code",
	Long:   "Generate a compact statusline string for display in Claude Code's status bar.",
	Hidden: true,
	RunE:   runStatusline,
}

func init() {
	StatuslineCmd.Flags().BoolVar(&statuslineRefreshGitHub, "refresh-github", false,
		"Refresh the cached GitHub issue/PR counts and exit (internal; spawned by the render path)")
	StatuslineCmd.Flags().StringVar(&statuslineBoardRoot, "board-root", "",
		"Project root holding .moai/state (internal; used with --refresh-github)")
	StatuslineCmd.Flags().BoolVar(&statuslineRefreshLanded, "refresh-landed", false,
		"Refresh the cached landed-card count and exit (internal; spawned by the render path)")
	_ = StatuslineCmd.Flags().MarkHidden("refresh-github")
	_ = StatuslineCmd.Flags().MarkHidden("refresh-landed")
	_ = StatuslineCmd.Flags().MarkHidden("board-root")
}

// runStatusline renders a statusline string suitable for Claude Code's status bar.
//
// @MX:ANCHOR: statusline CLI entry point (fan_in >= 3: shell wrapper, direct CLI, tests)
// @MX:REASON: public entry point for statusline rendering — cwd guard is critical for stability
// @MX:SPEC: SPEC-V3R3-STATUSLINE-FALLBACK-001
func runStatusline(cmd *cobra.Command, _ []string) error {
	// Detached refresh entry point: fetch the GitHub counts, write the cache,
	// and exit without rendering. Errors are swallowed — a failed refresh must
	// leave the status bar showing the previous value, not a failure.
	if statuslineRefreshGitHub {
		_ = statusline.RefreshGitHubCounts(cmd.Context(), statuslineBoardRoot)
		return nil
	}
	// Same contract for the landed count: measure, write the cache, exit
	// without rendering. A failed measurement leaves the previous value.
	if statuslineRefreshLanded {
		_ = statusline.RefreshLandedCounts(cmd.Context(), statuslineBoardRoot)
		return nil
	}

	// AC-SF-006: cwd guard — deleted directory fallback
	// os.Getwd() may succeed on macOS even with deleted cwd, so also check with os.Stat()
	if wd, err := os.Getwd(); err != nil || !dirExists(wd) {
		home, _ := os.UserHomeDir()
		if home != "" {
			_ = os.Chdir(home)
		}
	}

	out := cmd.OutOrStdout()

	// Render-budget guard (issue #646): Claude Code allots a statusline render only a
	// few hundred milliseconds. The usage collector's OAuth 429 retry honors Retry-After,
	// which can stall the render for minutes when the endpoint is rate limited. Every
	// collector swallows its own error, so an expired context drops just that segment
	// rather than failing the whole render.
	ctx, cancel := context.WithTimeout(context.Background(), statuslineRenderBudget)
	defer cancel()

	// Get project root for git and version detection (error ignored: empty root is valid)
	projectRoot, _ := findProjectRootFn() //nolint:errcheck // empty root is acceptable fallback

	// Load full statusline config from statusline.yaml
	statuslineCfg := loadStatuslineFileConfig(projectRoot)

	var segmentConfig map[string]bool
	var themeName string
	if statuslineCfg != nil {
		// Segments map is the only statusline configuration lever besides theme
		// (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 retired the preset shorthand).
		// When segments is absent the Builder falls back to all-enabled.
		segmentConfig = statuslineCfg.Segments
		themeName = statuslineCfg.Theme
	}

	// Build statusline options - git and version are auto-detected.
	// Mode is fixed to ModeDefault: the `mode:` config surface was removed
	// (SLR-1 — mode=full was inert), but the Builder Config Mode field is
	// preserved as part of the Builder API (HARD-1) and fed ModeDefault.
	// workflow.todo.enabled gates the backlog segment alongside the
	// statusline.yaml `segments.backlog` key (SPEC-TODO-ENABLE-FLAG-001
	// REQ-2). Either switch off means the segment is off; neither overrides
	// the other. TodoEnabledForRoot fails open, so an unreadable config keeps
	// the segment rather than hiding it for an invisible reason.
	todoEnabled := config.TodoEnabledForRoot(projectRoot)

	opts := statusline.Options{
		Mode:          statusline.ModeDefault,
		NoColor:       os.Getenv("NO_COLOR") != "" || os.Getenv(config.EnvNoColor) != "",
		RootDir:       projectRoot,
		SegmentConfig: segmentConfig,
		ThemeName:     themeName,
		TodoEnabled:   &todoEnabled,
	}

	// Create builder and render
	builder := statusline.New(opts)

	// Try to read stdin with TTY detection
	stdinData := readStdinWithTimeout()

	result, err := builder.Build(ctx, stdinData)
	if err != nil {
		// Fallback on error
		_, _ = fmt.Fprintln(out, renderSimpleFallback())
		return nil
	}

	_, _ = fmt.Fprintln(out, result)
	return nil
}

// dirExists checks whether a directory exists at the given path.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// readStdinWithTimeout reads stdin with TTY detection.
// Returns an empty reader if stdin is a terminal (to prevent blocking).
// Returns os.Stdin if stdin is piped or redirected (for Claude Code context).
func readStdinWithTimeout() io.Reader {
	stdinFile, err := os.Stdin.Stat()
	if err != nil {
		return io.MultiReader()
	}

	// Check if stdin is a terminal (character device)
	// If not a terminal (pipe/redirect), read normally
	if stdinFile.Mode()&os.ModeCharDevice == 0 {
		return os.Stdin
	}

	// stdin is a terminal - use empty reader to prevent blocking
	return io.MultiReader()
}

// renderSimpleFallback returns a simple fallback statusline.
func renderSimpleFallback() string {
	return "moai"
}

// statuslineFileConfig holds all statusline configuration read from YAML. It
// mirrors the canonical models.StatuslineConfig shape {Theme, Segments}. The
// `mode:` surface was removed (SLM-1/SLR-2 — inert) and the `preset:` surface
// was retired (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001). A legacy `preset:` key
// in an existing statusline.yaml is silently ignored (unknown YAML keys do not
// error on unmarshal).
type statuslineFileConfig struct {
	Theme    string
	Segments map[string]bool
}

// loadStatuslineFileConfig reads the full statusline configuration from
// .moai/config/sections/statusline.yaml. Returns nil if the file is missing,
// unreadable, or unparseable.
func loadStatuslineFileConfig(projectRoot string) *statuslineFileConfig {
	if projectRoot == "" {
		return nil
	}

	configPath := filepath.Join(projectRoot, ".moai", "config", "sections", "statusline.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var raw struct {
		Statusline struct {
			Theme    string          `yaml:"theme"`
			Segments map[string]bool `yaml:"segments"`
		} `yaml:"statusline"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil
	}

	return &statuslineFileConfig{
		Theme:    raw.Statusline.Theme,
		Segments: raw.Statusline.Segments,
	}
}
