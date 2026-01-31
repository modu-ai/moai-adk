package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/moai-adk-go/internal/output"
	"github.com/anthropics/moai-adk-go/pkg/version"
	"gopkg.in/yaml.v3"
)

// CheckResult represents the result of a single health check
type CheckResult struct {
	Name    string
	Status  string // "success", "warning", "error"
	Message string
	Value   string
}

// Doctor performs health checks on the MoAI installation
type Doctor struct {
	projectDir string
	results    []*CheckResult
}

// NewDoctor creates a new doctor instance
func NewDoctor(projectDir string) *Doctor {
	return &Doctor{
		projectDir: projectDir,
		results:    make([]*CheckResult, 0),
	}
}

// RunAllChecks runs all health checks
func (d *Doctor) RunAllChecks() error {
	fmt.Println(output.HeaderStyle.Render("MoAI Health Check"))
	fmt.Println()

	// Run checks in order
	d.checkBinaryVersion()
	d.checkClaudeCode()
	d.checkExternalTools()
	d.checkConfigFiles()
	d.checkHookRegistration()

	// Print results
	d.printResults()

	// Return appropriate exit code
	return d.getExitCode()
}

// checkBinaryVersion checks the Go binary version
func (d *Doctor) checkBinaryVersion() {
	// Get binary path
	binaryPath, err := os.Executable()
	if err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "Binary Version",
			Status:  "error",
			Message: "Failed to get binary path",
		})
		return
	}

	// Get version from build info (set via ldflags)
	ver := version.GetVersion()

	d.results = append(d.results, &CheckResult{
		Name:    "Binary Version",
		Status:  "success",
		Message: fmt.Sprintf("moai %s", ver),
		Value:   binaryPath,
	})
}

// checkClaudeCode checks if Claude Code is installed
func (d *Doctor) checkClaudeCode() {
	// Try to run claude --version
	cmd := exec.Command("claude", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "Claude Code",
			Status:  "error",
			Message: "not found",
		})
		return
	}

	version := strings.TrimSpace(string(output))
	d.results = append(d.results, &CheckResult{
		Name:    "Claude Code",
		Status:  "success",
		Message: version,
	})
}

// checkExternalTools checks for external tools
func (d *Doctor) checkExternalTools() {
	// Check git
	d.checkTool("git", "git", []string{"--version"}, true)

	// Check ruff (Python linter)
	d.checkTool("ruff", "ruff", []string{"--version"}, false)

	// Check eslint (JS/TS linter)
	d.checkTool("eslint", "eslint", []string{"--version"}, false)

	// Check ast-grep
	d.checkTool("ast-grep", "sg", []string{"--version"}, false)
}

// checkTool checks if a tool is installed
func (d *Doctor) checkTool(displayName, command string, args []string, required bool) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()

	status := "success"
	if err != nil {
		if required {
			status = "error"
		} else {
			status = "warning"
		}
	}

	version := strings.TrimSpace(string(output))
	if err != nil {
		version = "not found"
	}

	d.results = append(d.results, &CheckResult{
		Name:    displayName,
		Status:  status,
		Message: version,
	})
}

// checkConfigFiles validates configuration files
func (d *Doctor) checkConfigFiles() {
	// Check settings.json
	d.checkSettingsJSON()

	// Check .moai/config/ structure
	d.checkMoaiConfig()

	// Check language.yaml
	d.checkLanguageConfig()

	// Check user.yaml
	d.checkUserConfig()
}

// checkSettingsJSON validates settings.json
func (d *Doctor) checkSettingsJSON() {
	settingsPath := filepath.Join(d.projectDir, ".claude", "settings.json")

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		d.results = append(d.results, &CheckResult{
			Name:    "settings.json",
			Status:  "warning",
			Message: "not found (run 'moai init')",
		})
		return
	}

	// Try to parse JSON
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "settings.json",
			Status:  "error",
			Message: "failed to read",
		})
		return
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "settings.json",
			Status:  "error",
			Message: "invalid JSON",
		})
		return
	}

	d.results = append(d.results, &CheckResult{
		Name:    "settings.json",
		Status:  "success",
		Message: "valid",
	})
}

// checkMoaiConfig checks .moai/config/ structure
func (d *Doctor) checkMoaiConfig() {
	configDir := filepath.Join(d.projectDir, ".moai", "config")

	// Check if directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		d.results = append(d.results, &CheckResult{
			Name:    "config structure",
			Status:  "warning",
			Message: "not found (run 'moai init')",
		})
		return
	}

	// Check for sections directory
	sectionsDir := filepath.Join(configDir, "sections")
	if _, err := os.Stat(sectionsDir); os.IsNotExist(err) {
		d.results = append(d.results, &CheckResult{
			Name:    "config structure",
			Status:  "warning",
			Message: "incomplete",
		})
		return
	}

	d.results = append(d.results, &CheckResult{
		Name:    "config structure",
		Status:  "success",
		Message: "complete",
	})
}

// checkLanguageConfig checks language.yaml
func (d *Doctor) checkLanguageConfig() {
	langPath := filepath.Join(d.projectDir, ".moai", "config", "sections", "language.yaml")

	if _, err := os.Stat(langPath); os.IsNotExist(err) {
		d.results = append(d.results, &CheckResult{
			Name:    "language.yaml",
			Status:  "warning",
			Message: "not found",
		})
		return
	}

	// Try to parse YAML
	data, err := os.ReadFile(langPath)
	if err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "language.yaml",
			Status:  "error",
			Message: "failed to read",
		})
		return
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "language.yaml",
			Status:  "error",
			Message: "invalid YAML",
		})
		return
	}

	d.results = append(d.results, &CheckResult{
		Name:    "language.yaml",
		Status:  "success",
		Message: "valid",
	})
}

// checkUserConfig checks user.yaml
func (d *Doctor) checkUserConfig() {
	userPath := filepath.Join(d.projectDir, ".moai", "config", "sections", "user.yaml")

	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		d.results = append(d.results, &CheckResult{
			Name:    "user.yaml",
			Status:  "warning",
			Message: "not found",
		})
		return
	}

	// Try to parse YAML
	data, err := os.ReadFile(userPath)
	if err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "user.yaml",
			Status:  "error",
			Message: "failed to read",
		})
		return
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "user.yaml",
			Status:  "error",
			Message: "invalid YAML",
		})
		return
	}

	d.results = append(d.results, &CheckResult{
		Name:    "user.yaml",
		Status:  "success",
		Message: "valid",
	})
}

// checkHookRegistration checks if hooks are registered
func (d *Doctor) checkHookRegistration() {
	settingsPath := filepath.Join(d.projectDir, ".claude", "settings.json")

	// Read settings.json
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "hooks",
			Status:  "warning",
			Message: "settings.json not found",
		})
		return
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		d.results = append(d.results, &CheckResult{
			Name:    "hooks",
			Status:  "warning",
			Message: "settings.json invalid",
		})
		return
	}

	// Check for hooks
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		d.results = append(d.results, &CheckResult{
			Name:    "hooks",
			Status:  "error",
			Message: "no hooks registered",
		})
		return
	}

	// Check for required hooks
	requiredHooks := []string{"SessionStart", "PreToolUse", "PostToolUse"}
	missingHooks := []string{}

	for _, hookName := range requiredHooks {
		if _, exists := hooks[hookName]; !exists {
			missingHooks = append(missingHooks, hookName)
		}
	}

	if len(missingHooks) > 0 {
		d.results = append(d.results, &CheckResult{
			Name:    "hooks",
			Status:  "error",
			Message: fmt.Sprintf("missing hooks: %s", strings.Join(missingHooks, ", ")),
		})
		return
	}

	d.results = append(d.results, &CheckResult{
		Name:    "hooks",
		Status:  "success",
		Message: "all registered",
	})
}

// printResults prints all check results
func (d *Doctor) printResults() {
	for _, result := range d.results {
		var prefix string

		switch result.Status {
		case "success":
			prefix = "✓ "
		case "warning":
			prefix = "⚠ "
		case "error":
			prefix = "✗ "
		}

		message := fmt.Sprintf("%s: %s", result.Name, result.Message)
		fmt.Println(prefix + message)
	}

	fmt.Println()

	// Print summary
	errorCount := 0
	warningCount := 0

	for _, result := range d.results {
		switch result.Status {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		}
	}

	if errorCount > 0 || warningCount > 0 {
		fmt.Println(output.WarningStyle.Render(fmt.Sprintf("Issues: %d errors, %d warnings", errorCount, warningCount)))
		fmt.Println()

		// Print details for errors and warnings
		for _, result := range d.results {
			if result.Status == "error" || result.Status == "warning" {
				if result.Value != "" {
					fmt.Println(output.MutedStyle.Render(fmt.Sprintf("- %s: %s", result.Name, result.Value)))
				}
			}
		}
	} else {
		fmt.Println(output.SuccessStyle.Render("All checks passed!"))
	}
}

// getExitCode returns the appropriate exit code based on results
func (d *Doctor) getExitCode() error {
	errorCount := 0
	warningCount := 0

	for _, result := range d.results {
		switch result.Status {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%d critical failures found", errorCount)
	}

	if warningCount > 0 {
		return fmt.Errorf("%d optional tools missing", warningCount)
	}

	return nil
}
