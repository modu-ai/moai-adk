package quality

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk-go/internal/hook"
)

// Formatter handles automatic code formatting for PostToolUse hooks.
// It formats files after Write/Edit operations using language-specific tools.
type Formatter struct {
	registry *ToolRegistry
	detector *ChangeDetector
}

// NewFormatter creates a new Formatter with default registry and detector.
func NewFormatter() *Formatter {
	return &Formatter{
		registry: NewToolRegistry(),
		detector: NewChangeDetector(),
	}
}

// NewFormatterWithRegistry creates a Formatter using the provided registry.
func NewFormatterWithRegistry(registry *ToolRegistry) *Formatter {
	return &Formatter{
		registry: registry,
		detector: NewChangeDetector(),
	}
}

// FormatFile formats a single file using the appropriate formatter.
// Returns the tool result or an error if formatting fails catastrophically.
func (f *Formatter) FormatFile(ctx context.Context, filePath string, cwd string) (*ToolResult, error) {
	// Check if we should skip this file
	if !f.ShouldFormat(filePath) {
		return nil, nil
	}

	// Get formatters for this file
	formatters := f.registry.GetToolsForFile(filePath, ToolTypeFormatter)
	if len(formatters) == 0 {
		return nil, nil
	}

	// Try formatters in priority order
	for _, formatter := range formatters {
		// Check if tool is available
		if !f.registry.IsToolAvailable(formatter.Name) {
			continue
		}

		// Run the formatter
		result := f.registry.RunTool(ctx, formatter, filePath, cwd)
		result.ToolName = formatter.Name

		return &result, nil
	}

	// No formatter available
	return nil, nil
}

// ShouldFormat checks if a file should be formatted based on skip patterns.
func (f *Formatter) ShouldFormat(filePath string) bool {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check skip extensions
	if SkipExtensions[ext] {
		return false
	}

	// Check for minified files
	if strings.Contains(filepath.Base(filePath), ".min.") {
		return false
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return false
	}

	// Check skip directories
	for _, dir := range f.getParentDirectories(filePath) {
		if SkipDirectories[dir] {
			return false
		}
	}

	// Check if file is binary
	if f.isBinary(filePath) {
		return false
	}

	return true
}

// getParentDirectories returns all parent directory names for a file path.
func (f *Formatter) getParentDirectories(filePath string) []string {
	var dirs []string
	path := filePath

	for {
		dir := filepath.Base(path)
		if dir == "." || dir == "/" || dir == "" {
			break
		}
		dirs = append(dirs, dir)

		parent := filepath.Dir(path)
		if parent == path || parent == "." || parent == "/" {
			break
		}
		path = parent
	}

	return dirs
}

// isBinary checks if a file is binary by reading the first chunk.
func (f *Formatter) isBinary(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 8192)
	n, err := file.Read(buffer)
	if err != nil {
		return false
	}

	// Check for null byte (common in binary files)
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	return false
}

// FormatForHook formats a file and returns a HookOutput for PostToolUse.
// This is the main integration point with the hook system.
func (f *Formatter) FormatForHook(ctx context.Context, filePath string, cwd string) *hook.HookOutput {
	result, err := f.FormatFile(ctx, filePath, cwd)
	if err != nil {
		return &hook.HookOutput{
			HookSpecificOutput: &hook.HookSpecificOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: fmt.Sprintf("Formatter error: %v", err),
			},
		}
	}

	if result == nil {
		// File was skipped or no formatter available
		return hook.NewSuppressOutput()
	}

	if result.Success && result.FileModified {
		// File was formatted successfully
		return &hook.HookOutput{
			HookSpecificOutput: &hook.HookSpecificOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: fmt.Sprintf("Auto-formatted with %s", result.ToolName),
			},
		}
	}

	if !result.Success {
		// Formatter failed but don't block
		return &hook.HookOutput{
			HookSpecificOutput: &hook.HookSpecificOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: fmt.Sprintf("Format warning: %s", result.Error),
			},
		}
	}

	// Formatter ran but made no changes
	return hook.NewSuppressOutput()
}

// EventType implements hook.Handler interface.
func (f *Formatter) EventType() hook.EventType {
	return hook.EventPostToolUse
}

// Handle implements hook.Handler interface for PostToolUse events.
func (f *Formatter) Handle(ctx context.Context, input *hook.HookInput) (*hook.HookOutput, error) {
	// Only process Write and Edit tools
	if input.ToolName != "Write" && input.ToolName != "Edit" {
		return hook.NewAllowOutput(), nil
	}

	// Extract file path from tool input
	filePath, err := extractFilePath(input)
	if err != nil || filePath == "" {
		return hook.NewAllowOutput(), nil
	}

	// Format the file
	output := f.FormatForHook(ctx, filePath, input.CWD)

	return output, nil
}

// extractFilePath extracts the file_path from HookInput tool_input.
func extractFilePath(input *hook.HookInput) (string, error) {
	type toolInput struct {
		FilePath string `json:"file_path"`
	}

	var ti toolInput
	if err := json.Unmarshal(input.ToolInput, &ti); err != nil {
		return "", err
	}

	return ti.FilePath, nil
}
