package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/constitution"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/pkg/version"
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
	Use:     "doctor",
	Short:   "Run system diagnostics",
	GroupID: "project",
	Long:    "Run comprehensive system health checks including Claude Code configuration, dependency verification, and environment diagnostics.",
	RunE:    runDoctor,
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

	checks := runDiagnosticChecks(verbose, checkName)

	// Compute max label width for alignment.
	maxLabel := 0
	for _, c := range checks {
		if len(c.Name) > maxLabel {
			maxLabel = len(c.Name)
		}
	}

	okCount, warnCount, failCount := 0, 0, 0
	var lines []string
	for _, c := range checks {
		lines = append(lines, renderStatusLine(c.Status, c.Name, c.Message, maxLabel))
		if verbose && c.Detail != "" {
			lines = append(lines, fmt.Sprintf("    %s", cliMuted.Render(c.Detail)))
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

	summary := renderSummaryLine(okCount, warnCount, failCount)
	content := strings.Join(lines, "\n") + "\n\n" + summary

	_, _ = fmt.Fprintln(out, renderCard("System Diagnostics", content))

	if fix && failCount > 0 {
		var fixes []string
		for _, c := range checks {
			if c.Status == CheckFail {
				fixes = append(fixes, fmt.Sprintf("- %s: run 'moai init' to initialize project", c.Name))
			}
		}
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, renderInfoCard("Suggested Fixes", strings.Join(fixes, "\n")))
	}

	if exportPath != "" {
		if err := exportDiagnostics(exportPath, checks); err != nil {
			return fmt.Errorf("export diagnostics: %w", err)
		}
		_, _ = fmt.Fprintf(out, "\nDiagnostics exported to %s\n", exportPath)
	}

	return nil
}

// runDiagnosticChecks runs all diagnostic checks and returns results.
func runDiagnosticChecks(verbose bool, filterCheck string) []DiagnosticCheck {
	type checkFunc struct {
		name string
		fn   func(bool) DiagnosticCheck
	}

	cwd, _ := os.Getwd()
	allChecks := []checkFunc{
		{"Go Runtime", checkGoRuntime},
		{"Git", checkGit},
		{"MoAI Config", checkMoAIConfig},
		{"Claude Config", checkClaudeConfig},
		{"MoAI Version", checkMoAIVersion},
		{"Binary Freshness", checkBinaryFreshness},
		{"MCP Scope Duplicates", func(v bool) DiagnosticCheck { return checkMCPScopeDuplicates(cwd, v) }},
		{"Constitution Registry", func(v bool) DiagnosticCheck {
			registryPath := resolveRegistryPath(cwd)
			strictMode := os.Getenv(constitutionStrictEnvKey) == "1"
			return checkConstitution(cwd, registryPath, v, strictMode)
		}},
		{"Skills Allowlist", func(_ bool) DiagnosticCheck { return runSkillsCheck(cwd) }},
		{"Harness 5-Layer", func(_ bool) DiagnosticCheck { return runHarnessCheck(cwd) }},
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
		check.Detail = fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH"))
	}
	return check
}

// GitInstallHint returns OS-specific git installation instructions.
func GitInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install git: run 'xcode-select --install' or 'brew install git'"
	case "windows":
		return "Install git: run 'winget install Git.Git' or download from https://git-scm.com"
	default: // linux and other unix
		return "Install git: run 'sudo apt install git' (Debian/Ubuntu) or 'sudo yum install git' (RHEL/Fedora)"
	}
}

// checkGit verifies Git is installed and accessible.
func checkGit(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Git"}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		check.Status = CheckFail
		check.Message = "git not found in PATH"
		check.Detail = GitInstallHint()
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

	moaiDir := filepath.Join(cwd, defs.MoAIDir)
	info, err := os.Stat(moaiDir)
	if err != nil || !info.IsDir() {
		check.Status = CheckWarn
		check.Message = ".moai/ directory not found (run 'moai init')"
		return check
	}

	configDir := filepath.Join(moaiDir, defs.SectionsSubdir)
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

	claudeDir := filepath.Join(cwd, defs.ClaudeDir)
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

// checkBinaryFreshness warns when the installed binary was built from a
// commit older than the current source tree HEAD. Catches the class of
// regressions where a fix has been committed but the user has not rebuilt
// the binary, so hook handlers silently run the old code path.
//
// Skipped (reported OK) when:
//   - version.GetCommit() is unset ("", "none", "unknown") — dev build
//   - CWD is not inside a git working tree (git rev-parse walks upward)
//   - binary commit is not an ancestor of HEAD (release/branch build)
func checkBinaryFreshness(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Binary Freshness"}

	binCommit := strings.TrimSpace(version.GetCommit())
	if binCommit == "" || binCommit == "none" || binCommit == "unknown" {
		check.Status = CheckOK
		check.Message = "development build (no commit metadata)"
		return check
	}

	cwd, err := os.Getwd()
	if err != nil {
		check.Status = CheckOK
		check.Message = "cannot determine working directory"
		return check
	}

	headOut, err := exec.Command("git", "-C", cwd, "rev-parse", "HEAD").Output()
	if err != nil {
		check.Status = CheckOK
		check.Message = "not in a git source tree (skipped)"
		return check
	}
	sourceHead := strings.TrimSpace(string(headOut))

	if strings.HasPrefix(sourceHead, binCommit) {
		check.Status = CheckOK
		check.Message = fmt.Sprintf("binary matches source HEAD (%s)", binCommit)
		return check
	}

	// Binary differs from HEAD. Check whether it is an ancestor (= stale).
	ancestorErr := exec.Command("git", "-C", cwd, "merge-base", "--is-ancestor", binCommit, sourceHead).Run()
	if ancestorErr == nil {
		check.Status = CheckWarn
		check.Message = fmt.Sprintf("binary is behind source tree (binary: %s, HEAD: %s)", binCommit, shortCommit(sourceHead))
		check.Detail = "Run 'make build && make install' to rebuild with the latest fixes"
		return check
	}

	check.Status = CheckOK
	check.Message = fmt.Sprintf("binary from a different branch (binary: %s, HEAD: %s)", binCommit, shortCommit(sourceHead))
	if verbose {
		check.Detail = "binary commit is not an ancestor of source HEAD (release or branch build)"
	}
	return check
}

// shortCommit returns the first 9 characters of a git commit hash.
func shortCommit(hash string) string {
	if len(hash) < 9 {
		return hash
	}
	return hash[:9]
}

// findMCPDuplicates returns server names that appear more than once in the
// counts map. Used by checkMCPScopeDuplicates to detect same-name servers
// registered at multiple scopes (project + user).
func findMCPDuplicates(counts map[string]int) []string {
	var dups []string
	for name, count := range counts {
		if count > 1 {
			dups = append(dups, name)
		}
	}
	return dups
}

// mcpJSONServers holds the parsed structure of .mcp.json.
type mcpJSONServers struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// checkMCPScopeDuplicates detects MCP server names that appear in both the
// project .mcp.json and the global ~/.claude/.mcp.json, causing silent
// shadowing that can hide server configuration changes.
func checkMCPScopeDuplicates(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "MCP Scope Duplicates"}

	// Parse project .mcp.json
	projectServers := parseMCPJSON(filepath.Join(projectRoot, ".mcp.json"))

	// Parse global ~/.claude/.mcp.json
	homeDir, _ := os.UserHomeDir()
	globalServers := parseMCPJSON(filepath.Join(homeDir, ".claude", ".mcp.json"))

	if len(projectServers) == 0 && len(globalServers) == 0 {
		check.Status = CheckOK
		check.Message = "no MCP configuration found (project or global)"
		return check
	}

	// Tally server name occurrences across both scopes
	counts := make(map[string]int)
	for name := range projectServers {
		counts[name]++
	}
	for name := range globalServers {
		counts[name]++
	}

	dups := findMCPDuplicates(counts)
	if len(dups) == 0 {
		check.Status = CheckOK
		check.Message = fmt.Sprintf("%d project, %d global MCP servers — no duplicates", len(projectServers), len(globalServers))
		return check
	}

	check.Status = CheckWarn
	check.Message = fmt.Sprintf("duplicate MCP server names across scopes: %s", strings.Join(dups, ", "))
	if verbose {
		check.Detail = "Duplicate server names may shadow configuration. Remove duplicates from one scope."
	}
	return check
}

// parseMCPJSON reads .mcp.json and returns the set of server names.
// Returns empty map on error or missing file.
func parseMCPJSON(path string) map[string]struct{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var parsed mcpJSONServers
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil
	}
	result := make(map[string]struct{}, len(parsed.MCPServers))
	for name := range parsed.MCPServers {
		result[name] = struct{}{}
	}
	return result
}

// statusIcon returns a colored Unicode icon for the check status.
func statusIcon(s CheckStatus) string {
	switch s {
	case CheckOK:
		return cliSuccess.Render("\u2713")
	case CheckWarn:
		return cliWarn.Render("\u26A0")
	case CheckFail:
		return cliError.Render("\u2717")
	default:
		return "?"
	}
}

// constitutionStrictEnvKey is the environment variable name that activates strict mode.
const constitutionStrictEnvKey = "MOAI_CONSTITUTION_STRICT"

// checkConstitution checks the zone registry state.
// - registry file missing: Warn (optional feature)
// - load error (duplicate ID, invalid YAML, etc.): Fail
// - 0 Frozen entries: Warn
// - orphan warnings present + strictMode: Fail; otherwise: Warn
// - healthy: OK
func checkConstitution(projectDir, registryPath string, verbose, strictMode bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Constitution Registry"}

	// Check whether the registry file exists.
	if _, err := os.Stat(registryPath); err != nil {
		check.Status = CheckWarn
		check.Message = fmt.Sprintf("zone-registry.md not found at %q — run `moai constitution list` to verify", registryPath)
		return check
	}

	reg, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		check.Status = CheckFail
		check.Message = fmt.Sprintf("registry load error: %v", err)
		return check
	}

	// Check for orphan warnings.
	if len(reg.Warnings) > 0 && strictMode {
		check.Status = CheckFail
		check.Message = fmt.Sprintf("%d orphan/overflow warning(s) detected (strict mode)", len(reg.Warnings))
		if verbose {
			check.Detail = strings.Join(reg.Warnings, "\n")
		}
		return check
	}

	// Check the number of Frozen entries.
	frozen := reg.FilterByZone(constitution.ZoneFrozen)
	if len(frozen) == 0 {
		check.Status = CheckWarn
		check.Message = "no Frozen entries found in registry — expected at least 1"
		return check
	}

	// Only orphan warnings present (non-strict).
	if len(reg.Warnings) > 0 {
		check.Status = CheckWarn
		check.Message = fmt.Sprintf("registry OK (%d entries, %d Frozen), %d orphan/overflow warning(s)",
			len(reg.Entries), len(frozen), len(reg.Warnings))
		if verbose {
			check.Detail = strings.Join(reg.Warnings, "\n")
		}
		return check
	}

	check.Status = CheckOK
	check.Message = fmt.Sprintf("registry OK — %d entries (%d Frozen, %d Evolvable)",
		len(reg.Entries), len(frozen), len(reg.Entries)-len(frozen))
	return check
}

// exportDiagnostics writes check results to a JSON file.
func exportDiagnostics(path string, checks []DiagnosticCheck) error {
	data, err := json.MarshalIndent(checks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal diagnostics: %w", err)
	}
	return os.WriteFile(path, data, defs.FilePerm)
}
