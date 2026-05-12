package cli

// @MX:NOTE: [AUTO] Doctor command runs comprehensive system diagnostics
// @MX:NOTE: [AUTO] Checks Go runtime, Git, MoAI/Claude config, binary freshness, MCP duplicates, constitution
// @MX:NOTE: [AUTO] Binary freshness check detects stale builds via commit hash comparison

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
	"github.com/modu-ai/moai-adk/internal/migration"
	"github.com/modu-ai/moai-adk/internal/tui"
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

// checkGroup holds a named group of DiagnosticCheck items for grouped rendering.
type checkGroup struct {
	title  string
	checks []DiagnosticCheck
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

// @MX:NOTE: [AUTO] doctor 명령어 출력 — tui.Section + 19+ CheckLine + 요약 Box/Pill로 구성.
// runDoctor executes the system diagnostics workflow.
func runDoctor(cmd *cobra.Command, _ []string) error {
	verbose := getBoolFlag(cmd, "verbose")
	fix := getBoolFlag(cmd, "fix")
	exportPath := getStringFlag(cmd, "export")
	checkName := getStringFlag(cmd, "check")

	out := cmd.OutOrStdout()
	th := resolveTheme()

	groups := runGroupedChecks(verbose, checkName)

	// Flatten for export / fix path.
	var allChecks []DiagnosticCheck
	for _, g := range groups {
		allChecks = append(allChecks, g.checks...)
	}

	okCount, warnCount, failCount := 0, 0, 0
	for _, c := range allChecks {
		switch c.Status {
		case CheckOK:
			okCount++
		case CheckWarn:
			warnCount++
		case CheckFail:
			failCount++
		}
	}

	// Render grouped output.
	var bodyLines []string
	for _, g := range groups {
		if len(g.checks) == 0 {
			continue
		}
		bodyLines = append(bodyLines, tui.Section(g.title, tui.SectionOpts{Theme: &th}))
		for _, c := range g.checks {
			status := checkStatusToTUI(c.Status)
			bodyLines = append(bodyLines, tui.CheckLine(status, c.Name, c.Message, "", &th))
			if verbose && c.Detail != "" {
				bodyLines = append(bodyLines, "    "+c.Detail)
			}
		}
		bodyLines = append(bodyLines, "")
	}

	// Summary pill row.
	pPass := tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("통과 %d", okCount), Theme: &th})
	pWarn := tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: fmt.Sprintf("주의 %d", warnCount), Theme: &th})
	pErr := tui.Pill(tui.PillOpts{Kind: tui.PillErr, Solid: false, Label: fmt.Sprintf("실패 %d", failCount), Theme: &th})
	summaryPills := pPass + "  " + pWarn + "  " + pErr

	bodyLines = append(bodyLines, summaryPills)

	box := tui.Box(tui.BoxOpts{
		Title: "System Diagnostics",
		Body:  strings.Join(bodyLines, "\n"),
		Theme: &th,
	})
	_, _ = fmt.Fprintln(out, box)

	if fix && failCount > 0 {
		var fixes []string
		for _, c := range allChecks {
			if c.Status == CheckFail {
				fixes = append(fixes, fmt.Sprintf("- %s: run 'moai init' to initialize project", c.Name))
			}
		}
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, renderInfoCard("Suggested Fixes", strings.Join(fixes, "\n")))
	}

	if exportPath != "" {
		if err := exportDiagnostics(exportPath, allChecks); err != nil {
			return fmt.Errorf("export diagnostics: %w", err)
		}
		_, _ = fmt.Fprintf(out, "\nDiagnostics exported to %s\n", exportPath)
	}

	return nil
}

// checkStatusToTUI converts a CheckStatus to the tui.CheckLine status string.
func checkStatusToTUI(s CheckStatus) string {
	switch s {
	case CheckOK:
		return "ok"
	case CheckWarn:
		return "warn"
	case CheckFail:
		return "err"
	default:
		return "info"
	}
}

// runGroupedChecks runs all diagnostic checks grouped into sections.
// It returns groups in a deterministic order: System, MoAI-ADK, Workspace.
func runGroupedChecks(verbose bool, filterCheck string) []checkGroup {
	cwd, _ := os.Getwd()

	type checkFunc struct {
		name string
		fn   func(bool) DiagnosticCheck
	}

	systemChecks := []checkFunc{
		{"Go Runtime", checkGoRuntime},
		{"Git", checkGit},
		{"Claude Code", checkClaudeCode},
		{"GitHub CLI", checkGitHubCLI},
	}

	moaiChecks := []checkFunc{
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
		{"Harness 5-Layer", func(v bool) DiagnosticCheck { return runHarnessCheck(cwd) }},
		{"Migration", func(v bool) DiagnosticCheck { return checkMigration(cwd, v) }},
	}

	workspaceChecks := []checkFunc{
		{"Hooks Config", func(v bool) DiagnosticCheck { return checkHooksConfig(cwd, v) }},
		{"Slash Commands", func(v bool) DiagnosticCheck { return checkSlashCommands(cwd, v) }},
		{"Skills Allowlist", func(v bool) DiagnosticCheck { return checkSkillsAllowlist(cwd, v) }},
		{"MX Tag Config", func(v bool) DiagnosticCheck { return checkMXTagConfig(cwd, v) }},
		{"Worktree State", func(v bool) DiagnosticCheck { return checkWorktreeState(cwd, v) }},
		{"BODP Config", func(v bool) DiagnosticCheck { return checkBODPConfig(cwd, v) }},
		{"Telemetry Config", func(v bool) DiagnosticCheck { return checkTelemetryConfig(cwd, v) }},
		{"Glamour Cache", checkGlamourCache},
	}

	run := func(items []checkFunc) []DiagnosticCheck {
		var results []DiagnosticCheck
		for _, c := range items {
			if filterCheck != "" && c.name != filterCheck {
				continue
			}
			results = append(results, c.fn(verbose))
		}
		return results
	}

	return []checkGroup{
		{title: "System", checks: run(systemChecks)},
		{title: "MoAI-ADK", checks: run(moaiChecks)},
		{title: "Workspace", checks: run(workspaceChecks)},
	}
}

// runDiagnosticChecks runs all diagnostic checks and returns a flat list.
// Kept for backward compatibility with JSON export and existing tests.
func runDiagnosticChecks(verbose bool, filterCheck string) []DiagnosticCheck {
	groups := runGroupedChecks(verbose, filterCheck)
	var results []DiagnosticCheck
	for _, g := range groups {
		results = append(results, g.checks...)
	}
	return results
}

// checkGoRuntime verifies the Go runtime is available.
func checkGoRuntime(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Go Runtime"}
	// goVersion() is defined in banner.go (same package). It reads
	// MOAI_GO_VERSION_OVERRIDE env first for deterministic test output.
	goVer := goVersion()
	check.Status = CheckOK
	// goosArch() reads MOAI_GOOS_OVERRIDE/MOAI_GOARCH_OVERRIDE for golden-test
	// determinism across CI runners (linux/amd64, darwin/arm64, windows/amd64).
	check.Message = fmt.Sprintf("go %s (%s)", goVer, goosArch())
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
// MOAI_GIT_VERSION_OVERRIDE env short-circuits exec for golden-test determinism
// across CI runners (Apple Git 2.50.x on macOS-latest vs git 2.53.x on ubuntu-latest).
func checkGit(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Git"}
	if v := gitVersionOverride(); v != "" {
		check.Status = CheckOK
		check.Message = v
		return check
	}
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

// checkClaudeCode verifies the Claude Code CLI version.
// Reads CLAUDE_CODE_VERSION env first (deterministic for tests), then exec fallback.
func checkClaudeCode(_ bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Claude Code"}
	cv := claudeVersion()
	if cv == "claude" {
		// env not set — try exec
		out, err := exec.Command("claude", "--version").Output()
		if err != nil {
			check.Status = CheckWarn
			check.Message = "claude CLI not found (CLAUDE_CODE_VERSION unset)"
			return check
		}
		cv = strings.TrimSpace(string(out))
	}
	check.Status = CheckOK
	check.Message = cv
	return check
}

// checkGitHubCLI verifies the GitHub CLI (gh) is installed.
// MOAI_GH_VERSION_OVERRIDE env short-circuits exec for golden-test determinism
// across CI runners (gh release lag differs between ubuntu-latest and macOS).
func checkGitHubCLI(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "GitHub CLI"}
	if v := ghVersionOverride(); v != "" {
		check.Status = CheckOK
		check.Message = v
		return check
	}
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		check.Status = CheckWarn
		check.Message = "gh not found — GitHub Actions and PR workflows unavailable"
		return check
	}
	out, err := exec.Command("gh", "--version").Output()
	if err != nil {
		check.Status = CheckWarn
		check.Message = "gh found but version check failed"
		return check
	}
	// gh --version prints "gh version X.Y.Z (YYYY-MM-DD)\n..."
	first := strings.SplitN(string(out), "\n", 2)[0]
	check.Status = CheckOK
	check.Message = strings.TrimSpace(first)
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", ghPath)
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
		return cliSuccess.Render("✓")
	case CheckWarn:
		return cliWarn.Render("!")
	case CheckFail:
		return cliError.Render("✗")
	default:
		return "?"
	}
}

// constitutionStrictEnvKey is the environment variable name to enable strict mode.
const constitutionStrictEnvKey = "MOAI_CONSTITUTION_STRICT"

// checkConstitution checks the zone registry status.
// - registry file not found: Warn (optional feature)
// - load error (duplicate ID, invalid YAML, etc.): Fail
// - zero Frozen entries: Warn
// - orphan warnings present + strictMode: Fail; otherwise Warn
// - normal/OK: OK
func checkConstitution(projectDir, registryPath string, verbose, strictMode bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Constitution Registry"}

	// Check if registry file exists.
	// Use a project-relative path in the message so doctor output is stable
	// across CI runners (otherwise /home/runner/work/... vs /Users/goos/...
	// drift breaks golden tests).
	if _, err := os.Stat(registryPath); err != nil {
		displayPath := registryPath
		if rel, relErr := filepath.Rel(projectDir, registryPath); relErr == nil {
			// Force forward-slash separators so the message is identical on
			// Windows (\) and macOS/Linux (/) golden tests.
			displayPath = filepath.ToSlash(rel)
		}
		check.Status = CheckWarn
		check.Message = fmt.Sprintf("zone-registry.md not found at %q — run `moai constitution list` to verify", displayPath)
		return check
	}

	reg, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		check.Status = CheckFail
		check.Message = fmt.Sprintf("registry load error: %v", err)
		return check
	}

	// Check orphan warnings
	if len(reg.Warnings) > 0 && strictMode {
		check.Status = CheckFail
		check.Message = fmt.Sprintf("%d orphan/overflow warning(s) detected (strict mode)", len(reg.Warnings))
		if verbose {
			check.Detail = strings.Join(reg.Warnings, "\n")
		}
		return check
	}

	// Check Frozen entry count
	frozen := reg.FilterByZone(constitution.ZoneFrozen)
	if len(frozen) == 0 {
		check.Status = CheckWarn
		check.Message = "no Frozen entries found in registry — expected at least 1"
		return check
	}

	// Only orphan warnings case (non-strict)
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

// checkHooksConfig verifies the Claude Code hooks configuration exists.
func checkHooksConfig(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Hooks Config"}
	hooksDir := filepath.Join(projectRoot, ".claude", "hooks")
	info, err := os.Stat(hooksDir)
	if err != nil || !info.IsDir() {
		check.Status = CheckWarn
		check.Message = ".claude/hooks/ not found — hook handlers unavailable"
		return check
	}
	check.Status = CheckOK
	check.Message = "hook handlers directory found"
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", hooksDir)
	}
	return check
}

// checkSlashCommands verifies .claude/commands/ directory exists.
func checkSlashCommands(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Slash Commands"}
	cmdsDir := filepath.Join(projectRoot, ".claude", "commands")
	info, err := os.Stat(cmdsDir)
	if err != nil || !info.IsDir() {
		check.Status = CheckWarn
		check.Message = ".claude/commands/ not found — slash commands unavailable"
		return check
	}
	entries, _ := os.ReadDir(cmdsDir)
	check.Status = CheckOK
	check.Message = fmt.Sprintf("%d command file(s) registered", len(entries))
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", cmdsDir)
	}
	return check
}

// checkSkillsAllowlist verifies skill directories against the static allowlist.
func checkSkillsAllowlist(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Skills Allowlist"}
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			check.Status = CheckWarn
			check.Message = ".claude/skills/ not found"
			return check
		}
		check.Status = CheckFail
		check.Message = fmt.Sprintf("cannot read skills directory: %v", err)
		return check
	}
	warnCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if classifySkill(e.Name()) == "WARN" {
			warnCount++
		}
	}
	if warnCount > 0 {
		check.Status = CheckWarn
		check.Message = fmt.Sprintf("%d unknown moai- skill(s) detected (run 'moai update' to sync)", warnCount)
		if verbose {
			check.Detail = "Unknown skills may be outdated or removed. Verify with 'moai update'."
		}
		return check
	}
	check.Status = CheckOK
	check.Message = fmt.Sprintf("%d skill(s) verified", len(entries))
	return check
}

// checkMXTagConfig verifies the MX tag configuration file exists.
func checkMXTagConfig(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "MX Tag Config"}
	mxPath := filepath.Join(projectRoot, ".moai", "config", "sections", "mx.yaml")
	if _, err := os.Stat(mxPath); err != nil {
		check.Status = CheckWarn
		check.Message = ".moai/config/sections/mx.yaml not found — MX annotation thresholds use defaults"
		return check
	}
	check.Status = CheckOK
	check.Message = "MX tag configuration present"
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", mxPath)
	}
	return check
}

// checkWorktreeState verifies the .moai/state/ directory exists.
func checkWorktreeState(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Worktree State"}
	statePath := filepath.Join(projectRoot, ".moai", "state")
	info, err := os.Stat(statePath)
	if err != nil || !info.IsDir() {
		check.Status = CheckWarn
		check.Message = ".moai/state/ not found — worktree checkpoint storage unavailable"
		return check
	}
	check.Status = CheckOK
	check.Message = "worktree state directory found"
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", statePath)
	}
	return check
}

// checkBODPConfig verifies the BODP audit trail directory exists.
func checkBODPConfig(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "BODP Config"}
	bodpPath := filepath.Join(projectRoot, ".moai", "branches")
	info, err := os.Stat(bodpPath)
	if err != nil || !info.IsDir() {
		check.Status = CheckWarn
		check.Message = ".moai/branches/ not found — BODP audit trail disabled"
		return check
	}
	check.Status = CheckOK
	check.Message = "BODP audit trail directory found"
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", bodpPath)
	}
	return check
}

// checkTelemetryConfig verifies telemetry configuration section exists.
func checkTelemetryConfig(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Telemetry Config"}
	telPath := filepath.Join(projectRoot, ".moai", "config", "sections", "telemetry.yaml")
	if _, err := os.Stat(telPath); err != nil {
		check.Status = CheckWarn
		check.Message = ".moai/config/sections/telemetry.yaml not found — telemetry uses defaults"
		return check
	}
	check.Status = CheckOK
	check.Message = "telemetry configuration present"
	if verbose {
		check.Detail = fmt.Sprintf("path: %s", telPath)
	}
	return check
}

// checkGlamourCache is a D8 Placeholder check for Glamour cache health.
// The actual Glamour integration is scheduled for a follow-up SPEC.
func checkGlamourCache(_ bool) DiagnosticCheck {
	return DiagnosticCheck{
		Name:    "Glamour Cache",
		Status:  CheckWarn,
		Message: "glamour 미도입",
		Detail:  "후속 SPEC에서 실제 검사로 교체 예정",
	}
}

// checkMigration verifies migration framework status.
// REQ-V3R2-RT-007-015: doctor 명령에 마이그레이션 체크 추가.
func checkMigration(projectDir string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Migration"}

	runner := migration.NewRunner(projectDir)
	current, pending, lastApplied, err := runner.Status()
	if err != nil {
		check.Status = CheckFail
		check.Message = "마이그레이션 상태 조회 실패"
		check.Detail = err.Error()
		return check
	}

	check.Status = CheckOK
	check.Message = fmt.Sprintf("현재 버전 %d", current)

	if len(pending) > 0 {
		check.Status = CheckWarn
		check.Message = fmt.Sprintf("%d개 pending 마이그레이션", len(pending))
		if verbose {
			check.Detail = fmt.Sprintf("Pending 버전: %v", pending)
		}
	}

	if lastApplied != nil && verbose {
		if check.Detail != "" {
			check.Detail += "\n"
		}
		check.Detail += fmt.Sprintf("최근 적용: %s (버전 %d)", lastApplied.Name, lastApplied.Version)
	}

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
