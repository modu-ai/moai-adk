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
	"github.com/modu-ai/moai-adk/internal/tui"
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

// @MX:NOTE: [AUTO] status 명령어 출력 — tui.Box + tui.Section + tui.KV + tui.Pill로 구성.
// runStatus displays the current project status using the internal/tui design system.
// All colours are sourced from resolveTheme(); no hex literals appear in this function.
func runStatus(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	th := resolveTheme()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectName := filepath.Base(cwd)

	// Build body lines using tui primitives.
	var bodyLines []string

	// Section: 프로젝트 (Project information)
	bodyLines = append(bodyLines, tui.Section("Project", tui.SectionOpts{Theme: &th}))
	bodyLines = append(bodyLines, tui.KV("Project", projectName, tui.KVOpts{Theme: &th, KeyWidth: 8}))
	bodyLines = append(bodyLines, tui.KV("ADK", "moai-adk "+version.GetVersion(), tui.KVOpts{Theme: &th, KeyWidth: 8}))
	bodyLines = append(bodyLines, "")

	// Check .moai/ directory
	moaiDir := filepath.Join(cwd, ".moai")
	if _, statErr := os.Stat(moaiDir); statErr != nil {
		// Not initialized path: show single status pill.
		bodyLines = append(bodyLines, tui.Section("Status", tui.SectionOpts{Theme: &th}))
		pill := tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "Not initialized", Theme: &th})
		bodyLines = append(bodyLines, pill+" run 'moai init'")

		box := tui.Box(tui.BoxOpts{
			Title: "Project Status",
			Body:  strings.Join(bodyLines, "\n"),
			Theme: &th,
		})
		_, _ = fmt.Fprintln(out, box)
		return nil
	}

	// Section: Configuration
	bodyLines = append(bodyLines, tui.Section("Configuration", tui.SectionOpts{Theme: &th}))
	// Use forward-slash separator in display so the value is identical on
	// Windows (\) and macOS/Linux (/) golden tests.
	bodyLines = append(bodyLines, tui.KV("Config", filepath.ToSlash(filepath.Join(".moai", "config", "sections")), tui.KVOpts{Theme: &th, KeyWidth: 8}))

	// Count SPECs
	specsDir := filepath.Join(moaiDir, "specs")
	specCount := countDirs(specsDir)
	bodyLines = append(bodyLines, tui.KV("SPECs", fmt.Sprintf("%d found", specCount), tui.KVOpts{Theme: &th, KeyWidth: 8}))

	// Count config section files
	sectionsDir := filepath.Join(moaiDir, "config", "sections")
	sectionFiles := countFiles(sectionsDir, ".yaml")
	bodyLines = append(bodyLines, tui.KV("Configs", fmt.Sprintf("%d section files", sectionFiles), tui.KVOpts{Theme: &th, KeyWidth: 8}))
	bodyLines = append(bodyLines, "")

	// Summary pill row: status indicator
	pillStatus := tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Initialized", Theme: &th})
	pillSpecs := tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Solid: false, Label: fmt.Sprintf("SPECs %d", specCount), Theme: &th})
	bodyLines = append(bodyLines, pillStatus+"  "+pillSpecs)

	box := tui.Box(tui.BoxOpts{
		Title: "Project Status",
		Body:  strings.Join(bodyLines, "\n"),
		Theme: &th,
	})
	_, _ = fmt.Fprintln(out, box)

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
