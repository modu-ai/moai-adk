package cli

// @MX:NOTE: [AUTO] Self-Research System for autonomous moai-adk component experimentation
// @MX:NOTE: [AUTO] Research data stored in .moai/research/ with experiments/ and evals/ subdirs
// @MX:NOTE: [AUTO] Eval suites defined as *.eval.yaml files for automated testing

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newResearchCmd creates the root of the research command tree.
func newResearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "research",
		Short:   "Manage the Self-Research System",
		Long:    "Self-Research System for autonomously experimenting and improving moai-adk components",
		GroupID: "tools",
	}
	cmd.AddCommand(newResearchStatusCmd())
	cmd.AddCommand(newResearchBaselineCmd())
	cmd.AddCommand(newResearchListCmd())
	return cmd
}

// newResearchStatusCmd creates the subcommand that displays the research dashboard.
func newResearchStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display research dashboard",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runResearchStatus(cmd.OutOrStdout(), cwd)
		},
	}
}

// runResearchStatus prints the research status relative to projectDir.
func runResearchStatus(w io.Writer, projectDir string) error {
	researchDir := filepath.Join(projectDir, ".moai", "research")
	if _, err := os.Stat(researchDir); os.IsNotExist(err) {
		_, _ = fmt.Fprintln(w, "No research data found. Run 'moai research baseline <target>' first.")
		return nil
	}

	// Print basic dashboard
	pairs := []kvPair{
		{"Directory", filepath.Join(".moai", "research")},
	}

	// Count experiment result directories
	experimentsDir := filepath.Join(researchDir, "experiments")
	expCount := countDirs(experimentsDir)
	pairs = append(pairs, kvPair{"Experiments", fmt.Sprintf("%d found", expCount)})

	// Count eval suites
	evalsDir := filepath.Join(researchDir, "evals")
	evalCount := countEvalFiles(evalsDir)
	pairs = append(pairs, kvPair{"Eval Suites", fmt.Sprintf("%d registered", evalCount)})

	_, _ = fmt.Fprintln(w, renderCard("Research Status", renderKeyValueLines(pairs)))
	return nil
}

// newResearchBaselineCmd creates the baseline measurement subcommand (not yet implemented).
func newResearchBaselineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "baseline [target]",
		Short: "Baseline measurement (coming soon)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Baseline measurement is not yet implemented. Coming in a future release.")
			return nil
		},
	}
}

// newResearchListCmd creates the subcommand that lists registered eval suites.
func newResearchListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered eval suites",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runResearchList(cmd.OutOrStdout(), cwd)
		},
	}
}

// runResearchList lists eval files under .moai/research/evals/ relative to projectDir.
func runResearchList(w io.Writer, projectDir string) error {
	evalsDir := filepath.Join(projectDir, ".moai", "research", "evals")
	files, err := findEvalFiles(evalsDir)
	if err != nil || len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No eval suites found. Create one at .moai/research/evals/")
		return nil
	}

	_, _ = fmt.Fprintf(w, "Eval suites (%d):\n", len(files))
	for _, f := range files {
		// Print as a path relative to the project root
		rel, relErr := filepath.Rel(projectDir, f)
		if relErr != nil {
			rel = f
		}
		_, _ = fmt.Fprintf(w, "  %s\n", rel)
	}
	return nil
}

// findEvalFiles recursively searches dir for *.eval.yaml files.
func findEvalFiles(dir string) ([]string, error) {
	var results []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".eval.yaml") {
			results = append(results, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// countEvalFiles returns the number of *.eval.yaml files under dir.
func countEvalFiles(dir string) int {
	files, err := findEvalFiles(dir)
	if err != nil {
		return 0
	}
	return len(files)
}

func init() {
	rootCmd.AddCommand(newResearchCmd())
}
