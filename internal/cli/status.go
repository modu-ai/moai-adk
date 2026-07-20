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

const reminderMessage = `[!] Branch %q was created without going through MoAI BODP entry points.
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
// @MX:NOTE Reminder is invoked at the end of the status command. Does not block (REQ-CIAUT-050).
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
