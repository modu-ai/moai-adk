package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk-go/pkg/version"
)

// CheckStatus represents the result of a single diagnostic check.
type CheckStatus string

const (
	// CheckOK indicates the check passed.
	CheckOK CheckStatus = "ok"
	// CheckWarn indicates a non-fatal issue.
	CheckWarn CheckStatus = "warn"
	// CheckFail indicates a critical failure.
	CheckFail CheckStatus = "fail"
)

// DiagnosticCheck holds the result of a single health check.
type DiagnosticCheck struct {
	Name    string      `json:"name"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message"`
	Detail  string      `json:"detail,omitempty"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run system diagnostics",
	Long:  "Run comprehensive system health checks including Claude Code configuration, dependency verification, and environment diagnostics.",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	doctorCmd.Flags().BoolP("verbose", "v", false, "Show detailed diagnostic information")
	doctorCmd.Flags().Bool("fix", false, "Suggest fixes for detected issues")
	doctorCmd.Flags().String("export", "", "Export diagnostics to JSON file")
	doctorCmd.Flags().String("check", "", "Run a specific check only (e.g., git, go, config)")
}

// runDoctor executes the system diagnostics workflow.
func runDoctor(cmd *cobra.Command, _ []string) error {
	verbose := getBoolFlag(cmd, "verbose")
	fix := getBoolFlag(cmd, "fix")
	exportPath := getStringFlag(cmd, "export")
	checkName := getStringFlag(cmd, "check")

	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "System Diagnostics")
	fmt.Fprintln(out, "==================")
	fmt.Fprintln(out)

	checks := runDiagnosticChecks(verbose, checkName)

	okCount, warnCount, failCount := 0, 0, 0
	for _, c := range checks {
		icon := statusIcon(c.Status)
		fmt.Fprintf(out, "  %s %s: %s\n", icon, c.Name, c.Message)
		if verbose && c.Detail != "" {
			fmt.Fprintf(out, "      %s\n", c.Detail)
		}
		switch c.Status {
		case CheckOK:
			okCount++
		case CheckWarn:
			warnCount++
		case CheckFail:
			failCount++
		}
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Results: %d passed, %d warnings, %d failed\n", okCount, warnCount, failCount)

	if fix && failCount > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Suggested fixes:")
		for _, c := range checks {
			if c.Status == CheckFail {
				fmt.Fprintf(out, "  - %s: run 'moai init' to initialize project\n", c.Name)
			}
		}
	}

	if exportPath != "" {
		if err := exportDiagnostics(exportPath, checks); err != nil {
			return fmt.Errorf("export diagnostics: %w", err)
		}
		fmt.Fprintf(out, "\nDiagnostics exported to %s\n", exportPath)
	}

	return nil
}

// runDiagnosticChecks runs all diagnostic checks and returns results.
func runDiagnosticChecks(verbose bool, filterCheck string) []DiagnosticCheck {
	type checkFunc struct {
		name string
		fn   func(bool) DiagnosticCheck
	}

	allChecks := []checkFunc{
		{"Go Runtime", checkGoRuntime},
		{"Git", checkGit},
		{"MoAI Config", checkMoAIConfig},
		{"Claude Config", checkClaudeConfig},
		{"MoAI Version", checkMoAIVersion},
	}

	var results []DiagnosticCheck
	for _, c := range allChecks {
		if filterCheck != "" && c.name != filterCheck {
			continue
		}
		results = append(results, c.fn(verbose))
	}
	return results
}

// checkGoRuntime verifies the Go runtime is available.
func checkGoRuntime(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Go Runtime"}
	goVersion := runtime.Version()
	check.Status = CheckOK
	check.Message = fmt.Sprintf("%s (%s/%s)", goVersion, runtime.GOOS, runtime.GOARCH)
	if verbose {
		check.Detail = fmt.Sprintf("GOROOT=%s GOPATH=%s", runtime.GOROOT(), os.Getenv("GOPATH"))
	}
	return check
}

// checkGit verifies Git is installed and accessible.
func checkGit(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Git"}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		check.Status = CheckFail
		check.Message = "git not found in PATH"
		return check
	}

	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		check.Status = CheckWarn
		check.Message = "git found but version check failed"
		return check
	}

	check.Status = CheckOK
	check.Message = string(out[:len(out)-1]) // trim newline
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", gitPath)
	}
	return check
}

// checkMoAIConfig verifies .moai/ directory exists and contains valid config.
func checkMoAIConfig(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "MoAI Config"}

	cwd, err := os.Getwd()
	if err != nil {
		check.Status = CheckFail
		check.Message = "cannot determine working directory"
		return check
	}

	moaiDir := filepath.Join(cwd, ".moai")
	info, err := os.Stat(moaiDir)
	if err != nil || !info.IsDir() {
		check.Status = CheckWarn
		check.Message = ".moai/ directory not found (run 'moai init')"
		return check
	}

	configDir := filepath.Join(moaiDir, "config", "sections")
	if _, statErr := os.Stat(configDir); statErr != nil {
		check.Status = CheckWarn
		check.Message = ".moai/config/sections/ not found"
		return check
	}

	check.Status = CheckOK
	check.Message = "configuration directory found"
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", moaiDir)
	}
	return check
}

// checkClaudeConfig verifies .claude/ directory exists.
func checkClaudeConfig(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Claude Config"}

	cwd, err := os.Getwd()
	if err != nil {
		check.Status = CheckFail
		check.Message = "cannot determine working directory"
		return check
	}

	claudeDir := filepath.Join(cwd, ".claude")
	info, err := os.Stat(claudeDir)
	if err != nil || !info.IsDir() {
		check.Status = CheckWarn
		check.Message = ".claude/ directory not found"
		return check
	}

	check.Status = CheckOK
	check.Message = "Claude Code configuration found"
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", claudeDir)
	}
	return check
}

// checkMoAIVersion reports the current MoAI-ADK version.
func checkMoAIVersion(_ bool) DiagnosticCheck {
	return DiagnosticCheck{
		Name:    "MoAI Version",
		Status:  CheckOK,
		Message: fmt.Sprintf("moai-adk %s", version.GetVersion()),
	}
}

// statusIcon returns a text icon for the check status.
func statusIcon(s CheckStatus) string {
	switch s {
	case CheckOK:
		return "[OK]"
	case CheckWarn:
		return "[WARN]"
	case CheckFail:
		return "[FAIL]"
	default:
		return "[??]"
	}
}

// exportDiagnostics writes check results to a JSON file.
func exportDiagnostics(path string, checks []DiagnosticCheck) error {
	data, err := json.MarshalIndent(checks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal diagnostics: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
