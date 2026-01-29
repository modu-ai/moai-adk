package initializer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/moai-adk-go/internal/output"
)

// ProjectDetector detects existing MoAI-ADK projects
type ProjectDetector struct{}

// NewProjectDetector creates a new project detector
func NewProjectDetector() *ProjectDetector {
	return &ProjectDetector{}
}

// DetectResult represents the result of project detection
type DetectResult struct {
	HasClaudeDir bool
	HasMoaiDir   bool
	IsMoaiADK    bool
}

// Detect detects if the current directory is a MoAI-ADK project
func (pd *ProjectDetector) Detect(dir string) (*DetectResult, error) {
	result := &DetectResult{}

	// Check for .claude directory
	claudeDir := filepath.Join(dir, ".claude")
	if _, err := os.Stat(claudeDir); err == nil {
		result.HasClaudeDir = true
	}

	// Check for .moai directory
	moaiDir := filepath.Join(dir, ".moai")
	if _, err := os.Stat(moaiDir); err == nil {
		result.HasMoaiDir = true
	}

	// Consider it a MoAI-ADK project if both directories exist
	result.IsMoaiADK = result.HasClaudeDir && result.HasMoaiDir

	return result, nil
}

// WarnExisting warns the user about existing project
func (pd *ProjectDetector) WarnExisting(result *DetectResult) {
	if !result.IsMoaiADK {
		return
	}

	fmt.Println()
	fmt.Println(output.WarningStyle.Render("⚠ Existing MoAI-ADK project detected"))
	fmt.Println(output.MutedStyle.Render("  This directory already contains .claude/ and .moai/ directories."))
	fmt.Println(output.MutedStyle.Render("  Running init again may overwrite existing configuration."))
	fmt.Println()
}

// CheckDirectoryEmpty checks if the directory is empty (excluding dotfiles)
func (pd *ProjectDetector) CheckDirectoryEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("error reading directory: %w", err)
	}

	// Filter out dotfiles and dotdirectories
	nonDotCount := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".") {
			nonDotCount++
		}
	}

	return nonDotCount == 0, nil
}

// ShouldInit determines if initialization should proceed
func (pd *ProjectDetector) ShouldInit(dir string, force bool) (bool, error) {
	// Detect existing project
	result, err := pd.Detect(dir)
	if err != nil {
		return false, err
	}

	// Warn if existing project found
	if result.IsMoaiADK && !force {
		pd.WarnExisting(result)
		return false, nil // User needs to confirm
	}

	// Check if directory is not empty (excluding dotfiles)
	isEmpty, err := pd.CheckDirectoryEmpty(dir)
	if err != nil {
		return false, err
	}

	if !isEmpty && !force {
		fmt.Println()
		fmt.Println(output.WarningStyle.Render("⚠ Directory is not empty"))
		fmt.Println(output.MutedStyle.Render("  The current directory contains files. Use --force to override."))
		fmt.Println()
		return false, nil
	}

	return true, nil
}
