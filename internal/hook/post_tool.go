package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"

	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
)

// postToolHandler processes PostToolUse events.
// It collects tool execution metrics and prepares statusline data
// (REQ-HOOK-033). This handler is observation-only and always returns "allow".
// Optionally integrates with LSP diagnostics for Write/Edit operations.
type postToolHandler struct {
	diagnostics lsphook.LSPDiagnosticsCollector
}

// NewPostToolHandler creates a new PostToolUse event handler.
func NewPostToolHandler() Handler {
	return &postToolHandler{}
}

// NewPostToolHandlerWithDiagnostics creates a PostToolUse handler with LSP diagnostics.
// If diagnostics is nil, falls back to metrics-only collection.
func NewPostToolHandlerWithDiagnostics(diagnostics lsphook.LSPDiagnosticsCollector) Handler {
	return &postToolHandler{diagnostics: diagnostics}
}

// EventType returns EventPostToolUse.
func (h *postToolHandler) EventType() EventType {
	return EventPostToolUse
}

// Handle processes a PostToolUse event. It collects metrics about the tool
// execution (tool name, output size) and returns them in the Data field.
// For Write/Edit tools, also collects LSP diagnostics per REQ-HOOK-150.
// Always returns Decision "allow" (observation only).
func (h *postToolHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Debug("collecting post-tool metrics",
		"tool_name", input.ToolName,
		"session_id", input.SessionID,
	)

	metrics := map[string]any{
		"tool_name":  input.ToolName,
		"session_id": input.SessionID,
	}

	// Collect output size metric
	if len(input.ToolOutput) > 0 {
		metrics["output_size"] = len(input.ToolOutput)
	}

	// Collect input size metric
	if len(input.ToolInput) > 0 {
		metrics["input_size"] = len(input.ToolInput)
	}

	// Collect Task subagent metrics (SPEC-MONITOR-001).
	// Best-effort: errors are logged internally and never propagated.
	if input.ToolName == "Task" {
		logTaskMetrics(input)
	}

	// Collect LSP diagnostics for Write/Edit operations (REQ-HOOK-150, REQ-HOOK-153)
	if (input.ToolName == "Write" || input.ToolName == "Edit") && h.diagnostics != nil {
		h.collectDiagnostics(ctx, input, metrics)
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		slog.Error("failed to marshal post-tool metrics",
			"error", err.Error(),
		)
		return NewPostToolOutput(""), nil
	}

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName: "PostToolUse",
		},
		Data: jsonData,
	}, nil
}

// collectDiagnostics collects LSP diagnostics for the modified file.
// This is observation-only and MUST NOT block per REQ-HOOK-153.
func (h *postToolHandler) collectDiagnostics(ctx context.Context, input *HookInput, metrics map[string]any) {
	// Extract file path from tool input
	var parsed map[string]any
	if err := json.Unmarshal(input.ToolInput, &parsed); err != nil {
		slog.Debug("failed to parse tool input for diagnostics", "error", err)
		return
	}

	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		return
	}

	// Get diagnostics (observation only, never block)
	diagnostics, err := h.diagnostics.GetDiagnostics(ctx, filePath)
	if err != nil {
		slog.Debug("diagnostics collection failed (observation only)",
			"file_path", filePath,
			"error", err,
		)
		return
	}

	// Calculate severity counts
	counts := h.diagnostics.GetSeverityCounts(diagnostics)

	// Add diagnostic counts to metrics
	metrics["lsp_diagnostics"] = map[string]any{
		"file":        filepath.Base(filePath),
		"errors":      counts.Errors,
		"warnings":    counts.Warnings,
		"information": counts.Information,
		"hints":       counts.Hints,
		"total":       counts.Total(),
		"count":       len(diagnostics),
		"has_issues":  counts.Errors > 0 || counts.Warnings > 0,
	}

	// Log summary
	if counts.Errors > 0 || counts.Warnings > 0 {
		slog.Info("LSP diagnostics collected",
			"file_path", filepath.Base(filePath),
			"errors", counts.Errors,
			"warnings", counts.Warnings,
		)
	} else {
		slog.Debug("LSP diagnostics collected (clean)",
			"file_path", filepath.Base(filePath),
		)
	}
}
