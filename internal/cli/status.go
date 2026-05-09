package cli

// @MX:NOTE: [AUTO] Project status command showing SPEC progress, quality metrics, config
// @MX:NOTE: [AUTO] Counts SPECs in .moai/specs/ and config sections in .moai/config/sections/

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/bodp"
	"github.com/modu-ai/moai-adk/pkg/version"
)

const (
	envNoBODPReminder = "MOAI_NO_BODP_REMINDER"
	bodpAuditTrailDir = ".moai/branches/decisions"
)

// mainBranches enumerates the canonical default branches that BODP treats as
// "in-protocol" without an audit trail. Reminder must not fire on these.
var mainBranches = []string{"main", "master"}

const reminderMessage = `⚠️  Branch %q was created without going through MoAI BODP entry points.
Future branches: use ` + "`/moai plan --branch <SPEC-ID>`" + ` or ` + "`moai worktree new <SPEC-ID>`" + ` for relatedness check + audit trail.
Skip with %s=1 if intentional.
`

// emitOffProtocolReminder writes a notice to w when the user is on an
// off-protocol branch with no BODP audit trail. The reminder is purely
// informative (exit code unaffected by callers).
//
// Skip conditions (any one short-circuits to no-op):
//   - %s env var is "1"
//   - currentBranch is "main" or "master"
//   - bodp.HasAuditTrail returns true for currentBranch
//   - audit trail directory does not exist (fresh project — false-positive guard)
//
// @MX:NOTE Reminder는 status command 끝에 호출. Block 안 함 (REQ-CIAUT-050).
func emitOffProtocolReminder(repoRoot, currentBranch string, w io.Writer) {
	if os.Getenv(envNoBODPReminder) == "1" {
		return
	}
	if slices.Contains(mainBranches, currentBranch) {
		return
	}
	if bodp.HasAuditTrail(repoRoot, currentBranch) {
		return
	}
	dirPath := filepath.Join(repoRoot, bodpAuditTrailDir)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, reminderMessage, currentBranch, envNoBODPReminder)
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

	// W7-T05: BODP off-protocol branch reminder. Failures are silent — git
	// missing or non-repo cwd simply suppresses the reminder.
	if currentBranch, err := detectCurrentBranch(); err == nil {
		emitOffProtocolReminder(cwd, currentBranch, cmd.ErrOrStderr())
	}

	return nil
}

// detectCurrentBranch resolves the current git branch via `git rev-parse`.
// Returns an error when git is missing or cwd is not a git repository.
func detectCurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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
