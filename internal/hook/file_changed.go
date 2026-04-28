package hook

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// Supported language extensions for MX tag scanning.
var supportedExtensions = map[string]bool{
	".go":     true,
	".py":     true,
	".ts":     true,
	".js":     true,
	".rs":     true,
	".java":   true,
	".kt":     true,
	".cs":     true,
	".rb":     true,
	".php":    true,
	".ex":     true,
	".exs":    true,
	".cpp":    true,
	".cc":     true,
	".cxx":    true,
	".h":      true,
	".hpp":    true,
	".scala":  true,
	".r":      true,
	".dart":   true,
	".swift":  true,
}

// fileChangedHandler processes FileChanged events.
// It scans changed files for MX tag deltas.
type fileChangedHandler struct{}

// NewFileChangedHandler creates a new FileChanged event handler.
func NewFileChangedHandler() Handler {
	return &fileChangedHandler{}
}

// EventType returns EventFileChanged.
func (h *fileChangedHandler) EventType() EventType {
	return EventFileChanged
}

// Handle processes a FileChanged event. It checks if the changed file
// has a supported language extension, then scans for MX tag changes.
func (h *fileChangedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("file changed externally",
		"session_id", input.SessionID,
		"file_path", input.FilePath,
		"change_type", input.ChangeType,
	)

	// Skip deleted files
	if input.ChangeType == "deleted" {
		return &HookOutput{}, nil
	}

	// Check if file has supported extension
	ext := strings.ToLower(filepath.Ext(input.FilePath))
	if !supportedExtensions[ext] {
		slog.Debug("unsupported file extension for MX scan",
			"path", input.FilePath,
			"ext", ext,
		)
		return &HookOutput{}, nil
	}

	// Scan file for MX tags
	scanner := mx.NewScanner()
	tags, err := scanner.ScanFile(input.FilePath)
	if err != nil {
		slog.Warn("failed to scan file for MX tags",
			"path", input.FilePath,
			"error", err,
		)
		return &HookOutput{}, nil
	}

	// Compare with existing sidecar tags
	// For now, just report the count of tags found
	// TODO: Implement sidecar comparison when mx.Manager is available
	msg := h.formatTagDelta(input.FilePath, tags)
	if msg != "" {
		return &HookOutput{
			SystemMessage: msg,
		}, nil
	}

	return &HookOutput{}, nil
}

// formatTagDelta creates a summary message for MX tag changes.
func (h *fileChangedHandler) formatTagDelta(filePath string, tags []mx.Tag) string {
	if len(tags) == 0 {
		return ""
	}

	// Count tags by kind
	counts := make(map[mx.TagKind]int)
	for _, tag := range tags {
		counts[tag.Kind]++
	}

	var sb strings.Builder
	sb.WriteString(filePath)
	sb.WriteByte(':')
	for kind, count := range counts {
		fmt.Fprintf(&sb, " %s=%d", kind, count)
	}

	return "MX tag delta on " + sb.String()
}
