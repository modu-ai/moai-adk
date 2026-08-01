package cli

// @MX:NOTE: [AUTO] Project status command showing SPEC progress, quality metrics, config
// @MX:NOTE: [AUTO] Counts SPECs in .moai/specs/ and config sections in .moai/config/sections/

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/pkg/version"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show project status",
	GroupID: "project",
	Long:    "Display project state overview showing SPEC progress, quality metrics, and configuration summary.",
	RunE:    runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

// @MX:NOTE: [AUTO] status command output — markdown payload through the
// glamour gateway (renderMarkdown): rich on TTY, plain markdown passthrough on
// non-TTY/NO_COLOR (SPEC-CLI-TUX-V3-004 REQ-TUX4-004/005). Data fields are
// unchanged from the legacy Box surface (render-layer-only swap, plan §D).
// runStatus displays the current project status as a glamour-rendered
// markdown document. No hex literals; the glamour style derives from
// internal/tui tokens (glamour_style.go).
func runStatus(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectName := filepath.Base(cwd)

	var b strings.Builder
	b.WriteString("# Project Status\n\n")
	b.WriteString("## Project\n\n")
	fmt.Fprintf(&b, "- **Project**: %s\n", projectName)
	fmt.Fprintf(&b, "- **ADK**: moai-adk %s\n", version.GetVersion())

	// Check .moai/ directory
	moaiDir := filepath.Join(cwd, ".moai")
	if _, statErr := os.Stat(moaiDir); statErr != nil {
		// Not initialized path.
		b.WriteString("\n**Status**: Not initialized — run 'moai init'\n")
		_, _ = fmt.Fprint(out, renderMarkdown(out, b.String()))
		return nil
	}

	b.WriteString("\n## Configuration\n\n")
	// Use forward-slash separator in display so the value is identical on
	// Windows (\) and macOS/Linux (/) golden tests.
	fmt.Fprintf(&b, "- **Config**: %s\n", filepath.ToSlash(filepath.Join(".moai", "config", "sections")))

	// Count SPECs
	specCount := countDirs(filepath.Join(moaiDir, "specs"))
	fmt.Fprintf(&b, "- **SPECs**: %d found\n", specCount)

	// Count config section files
	sectionFiles := countFiles(filepath.Join(moaiDir, "config", "sections"), ".yaml")
	fmt.Fprintf(&b, "- **Configs**: %d section files\n", sectionFiles)

	fmt.Fprintf(&b, "\n**Status**: Initialized (SPECs %d)\n", specCount)

	_, _ = fmt.Fprint(out, renderMarkdown(out, b.String()))

	return nil
}

// countDirs counts the number of subdirectories in a directory.
func countDirs(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			count++
		}
	}
	return count
}

// countFiles counts the number of files with a given extension in a directory.
func countFiles(dir, ext string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ext {
			count++
		}
	}
	return count
}
