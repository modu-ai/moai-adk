package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/moai-adk-go/internal/hooks/protocol"
	"github.com/anthropics/moai-adk-go/internal/hooks/tools"
)

// PostToolHandler handles post-tool events (formatting, linting)
type PostToolHandler struct {
	registry *tools.ToolRegistry
}

// NewPostToolHandler creates a new post-tool handler
func NewPostToolHandler() *PostToolHandler {
	return &PostToolHandler{
		registry: tools.GetGlobalRegistry(),
	}
}

// Handle executes the post-tool hook
func (h *PostToolHandler) Handle(ctx context.Context, input *protocol.HookInput) (*protocol.HookResponse, error) {
	// Only process Write and Edit tools
	if input.ToolName != "Write" && input.ToolName != "Edit" {
		return protocol.NewMessageResponse("", true), nil // Suppress output
	}

	filePath := input.GetToolInputPath()
	if filePath == "" {
		return protocol.NewMessageResponse("", true), nil // Suppress output
	}

	// Check if we should skip this file
	if shouldSkipFile(filePath) {
		return protocol.NewMessageResponse("", true), nil // Suppress output
	}

	// Format the file
	formatted, message, err := h.formatFile(ctx, filePath)
	if err != nil {
		return protocol.NewErrorResponse(fmt.Sprintf("Format error: %v", err)), nil
	}

	if formatted {
		return protocol.NewMessageResponse("Auto-formatted: "+message, false), nil
	}

	// No formatting applied, suppress output
	return protocol.NewMessageResponse("", true), nil
}

// shouldSkipFile checks if a file should be skipped for formatting
func shouldSkipFile(filePath string) bool {
	// Check extension
	skipExtensions := map[string]bool{
		".json":    true,
		".lock":    true,
		".min.js":  true,
		".min.css": true,
		".map":     true,
		".svg":     true,
		".png":     true,
		".jpg":     true,
		".gif":     true,
		".ico":     true,
		".woff":    true,
		".woff2":   true,
		".ttf":     true,
		".eot":     true,
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if skipExtensions[ext] {
		return true
	}

	// Check for minified files
	if strings.Contains(filepath.Base(filePath), ".min.") {
		return true
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true
	}

	// Check for binary files
	file, err := os.Open(filePath)
	if err != nil {
		return true
	}
	defer func() { _ = file.Close() }()

	buf := make([]byte, 8192)
	n, err := file.Read(buf)
	if err != nil && n == 0 {
		return true
	}

	// Check for null bytes (binary indicator)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}

	return false
}

// formatFile formats a file using the appropriate formatter
func (h *PostToolHandler) formatFile(ctx context.Context, filePath string) (bool, string, error) {
	// Get language for file
	language := h.registry.GetLanguageForFile(filePath)
	if language == "" {
		return false, "", nil // Unknown file type, skip
	}

	// Get formatters for this language
	formatters := h.registry.GetToolsForLanguage(language, tools.ToolTypeFormatter)
	if len(formatters) == 0 {
		return false, "", nil // No formatter available
	}

	// Try the highest priority formatter
	formatter := formatters[0]

	// Check if tool is available
	if !h.registry.IsToolAvailable(formatter.Name) {
		return false, "", nil // Tool not available, skip gracefully
	}

	// Run formatter
	result, err := h.registry.RunTool(ctx, formatter, filePath)
	if err != nil {
		return false, "", fmt.Errorf("%s: %s", formatter.Name, err)
	}

	if !result.Success {
		return false, "", fmt.Errorf("%s failed: %s", formatter.Name, result.Error)
	}

	return true, fmt.Sprintf("Formatted with %s", formatter.Name), nil
}
