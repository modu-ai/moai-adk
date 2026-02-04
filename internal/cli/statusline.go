package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/spf13/cobra"
)

// StatuslineCmd is the statusline command.
var StatuslineCmd = &cobra.Command{
	Use:    "statusline",
	Short:  "Render statusline for Claude Code",
	Long:   "Generate a compact statusline string for display in Claude Code's status bar.",
	Hidden: true,
	RunE:   runStatusline,
}

// runStatusline renders a statusline string suitable for Claude Code's status bar.
func runStatusline(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	ctx := context.Background()

	// Determine display mode from environment or use default
	mode := statusline.StatuslineMode(os.Getenv("MOAI_STATUSLINE_MODE"))
	if mode == "" {
		mode = statusline.ModeDefault
	}

	// Get project root for git and version detection (error ignored: empty root is valid)
	projectRoot, _ := findProjectRoot() //nolint:errcheck // empty root is acceptable fallback

	// Build statusline options - git and version are auto-detected
	opts := statusline.Options{
		Mode:    mode,
		NoColor: os.Getenv("NO_COLOR") != "" || os.Getenv("MOAI_NO_COLOR") != "",
		RootDir: projectRoot,
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
