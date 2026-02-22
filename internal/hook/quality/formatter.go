package quality

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
)

// Formatter handles automatic code formatting per REQ-HOOK-070.
type Formatter struct {
	registry *toolRegistry
	detector *ChangeDetector
}

// NewFormatter creates a new Formatter with default registry and detector.
func NewFormatter(registry *toolRegistry) *Formatter {
	if registry == nil {
		registry = NewToolRegistry()
	}
	return &Formatter{
		registry: registry,
		detector: NewChangeDetector(),
	}
}

// NewFormatterWithRegistry creates a new Formatter with custom registry and detector.
func NewFormatterWithRegistry(registry *toolRegistry, detector *ChangeDetector) *Formatter {
	if registry == nil {
		registry = NewToolRegistry()
	}
	if detector == nil {
		detector = NewChangeDetector()
	}
	return &Formatter{
		registry: registry,
		detector: detector,
	}
}

// FormatFile formats a file using the appropriate formatter per REQ-HOOK-071.
// Returns nil result if file should not be formatted (skipped).
func (f *Formatter) FormatFile(ctx context.Context, filePath string) (*ToolResult, error) {
	// Check if file should be formatted
	if !f.ShouldFormat(filePath) {
		return nil, nil
	}

	// Get formatters for this file
	tools := f.registry.GetToolsForFile(filePath, ToolTypeFormatter)
	if len(tools) == 0 {
		return nil, nil
	}

	// Get directory for command execution
	cwd := filepath.Dir(filePath)
	if cwd == "" {
		cwd = "."
	}

	// Check if file exists first
	beforeHash, err := f.detector.ComputeHash(filePath)
	if err != nil {
		return nil, err
	}

	// Try each formatter in priority order
	for _, tool := range tools {
		// Check if tool is available
		if !f.registry.IsToolAvailable(tool.Command) {
			continue
		}

		// Run the formatter
		result := f.registry.RunTool(tool, filePath, cwd)

		// Check if file was modified
		if result.Success {
			_, hashErr := f.detector.ComputeHash(filePath)
			_ = hashErr // Ignore hash error, file may be locked
			changed, _ := f.detector.HasChanged(filePath, beforeHash)
			result.FileModified = changed
		}

		return &result, nil
	}

	return nil, nil
}

// ShouldFormat determines if a file should be formatted per REQ-HOOK-073, REQ-HOOK-074.
func (f *Formatter) ShouldFormat(filePath string) bool {
	// Convert to slashes for consistent path handling
	path := filepath.ToSlash(filePath)
	baseName := strings.ToLower(filepath.Base(filePath))

	// Check for skipped directories
	for _, skipDir := range skipDirectories {
		if strings.HasPrefix(path, skipDir+"/") || strings.Contains(path, "/"+skipDir+"/") {
			return false
		}
	}

	// Check for compound extensions like .min.js, .min.css
	for _, skipExt := range skipExtensions {
		if strings.HasSuffix(baseName, skipExt) {
			return false
		}
	}

	// Check file extension for simple cases
	ext := strings.ToLower(filepath.Ext(filePath))
	if slices.Contains(skipExtensions, ext) {
		return false
	}

	// Only format known code files
	tools := f.registry.GetToolsForFile(filePath, ToolTypeFormatter)
	return len(tools) > 0
}

// skipExtensions lists file extensions to skip per REQ-HOOK-073.
var skipExtensions = []string{
	".json", ".yaml", ".yml", ".toml", ".lock",
	".min.js", ".min.css",
	".map",
	".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".webp",
	".woff", ".woff2", ".ttf", ".eot", ".otf",
	".exe", ".dll", ".so", ".dylib", ".bin",
}

// skipDirectories lists directories to skip per REQ-HOOK-074.
var skipDirectories = []string{
	"node_modules", "vendor", ".venv", "venv",
	"__pycache__", ".cache", ".pytest_cache",
	"dist", "build", "target", ".next", ".nuxt", "out",
	".git", ".svn", ".hg",
	".idea", ".vscode", ".eclipse",
}
