package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/pkg/version"
)

// emitOffProtocolReminder writes a friendly notice when the user is on a
// branch that was created without going through MoAI's BODP entry points
// (skill plan-branch / plan-worktree / `moai worktree new`). RED phase stub —
// the GREEN phase (W7-T05) wires env opt-out, main-branch skip, audit-trail
// detection, and dir-absent false-positive suppression.
func emitOffProtocolReminder(repoRoot, currentBranch string, w io.Writer) {
	_, _, _ = repoRoot, currentBranch, w
}

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

// runStatus displays the current project status.
func runStatus(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectName := filepath.Base(cwd)

	pairs := []kvPair{
		{"Project", projectName},
		{"Path", cwd},
		{"ADK", "moai-adk " + version.GetVersion()},
	}

	// Check .moai/ directory
	moaiDir := filepath.Join(cwd, ".moai")
	if _, statErr := os.Stat(moaiDir); statErr != nil {
		pairs = append(pairs,
			kvPair{"Status", "Not initialized (run 'moai init')"},
		)
		_, _ = fmt.Fprintln(out, renderCard("Project Status", renderKeyValueLines(pairs)))
		return nil
	}
	pairs = append(pairs, kvPair{"Config", filepath.Join(".moai", "config", "sections")})

	// Count SPECs
	specsDir := filepath.Join(moaiDir, "specs")
	specCount := countDirs(specsDir)
	pairs = append(pairs, kvPair{"SPECs", fmt.Sprintf("%d found", specCount)})

	// Check config sections
	sectionsDir := filepath.Join(moaiDir, "config", "sections")
	sectionFiles := countFiles(sectionsDir, ".yaml")
	pairs = append(pairs, kvPair{"Configs", fmt.Sprintf("%d section files", sectionFiles)})

	pairs = append(pairs, kvPair{"Status", "Initialized"})

	_, _ = fmt.Fprintln(out, renderCard("Project Status", renderKeyValueLines(pairs)))

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
