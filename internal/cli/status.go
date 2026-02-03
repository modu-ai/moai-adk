package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk-go/pkg/version"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status",
	Long:  "Display project state overview showing SPEC progress, quality metrics, and configuration summary.",
	RunE:  runStatus,
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

	fmt.Fprintln(out, "Project Status")
	fmt.Fprintln(out, "==============")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "  Project:  %s\n", projectName)
	fmt.Fprintf(out, "  Path:     %s\n", cwd)
	fmt.Fprintf(out, "  ADK:      moai-adk %s\n", version.GetVersion())

	// Check .moai/ directory
	moaiDir := filepath.Join(cwd, ".moai")
	if _, statErr := os.Stat(moaiDir); statErr != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "  Status: Not initialized (run 'moai init')")
		return nil
	}
	fmt.Fprintf(out, "  Config:   %s\n", filepath.Join(".moai", "config", "sections"))

	// Count SPECs
	specsDir := filepath.Join(moaiDir, "specs")
	specCount := countDirs(specsDir)
	fmt.Fprintln(out)
	fmt.Fprintf(out, "  SPECs:    %d found\n", specCount)

	// Check config sections
	sectionsDir := filepath.Join(moaiDir, "config", "sections")
	sectionFiles := countFiles(sectionsDir, ".yaml")
	fmt.Fprintf(out, "  Configs:  %d section files\n", sectionFiles)

	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Status: Initialized")

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
