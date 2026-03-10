package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	astgrep "github.com/modu-ai/moai-adk/internal/astgrep"
	"github.com/modu-ai/moai-adk/internal/hook/mx"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
)

// FileAnalyzer is the interface for performing AST-based code scanning on a single file.
// astgrep.SGAnalyzer implements this interface.
type FileAnalyzer interface {
	ScanFile(ctx context.Context, filePath string, config *astgrep.ScanConfig) (*astgrep.ScanResult, error)
}

// postToolHandler processes PostToolUse events.
// It collects tool execution metrics and prepares statusline data
// (REQ-HOOK-033). This handler is observation-only and always returns "allow".
// Optionally integrates with LSP diagnostics for Write/Edit operations.
type postToolHandler struct {
	diagnostics lsphook.LSPDiagnosticsCollector
	// analyzer is an optional analyzer that performs AST-based scanning after Write/Edit operations.
	// If nil, AST scanning is skipped.
	analyzer FileAnalyzer
	// mxValidator is an optional validator for @MX tag checks after Write/Edit operations.
	// If nil, MX validation is skipped.
	mxValidator mx.Validator
	// mxTimeout is the timeout for MX validation. Default is 500ms.
	mxTimeout time.Duration
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

// NewPostToolHandlerWithAstgrep creates a PostToolUse handler that uses both LSP diagnostics and AST scanning.
// If diagnostics or analyzer is nil, the corresponding feature is skipped.
func NewPostToolHandlerWithAstgrep(diagnostics lsphook.LSPDiagnosticsCollector, analyzer FileAnalyzer) Handler {
	return &postToolHandler{diagnostics: diagnostics, analyzer: analyzer}
}

// NewPostToolHandlerWithMxValidator creates a PostToolUse handler with MX tag validation.
// projectRoot is the directory used for fan_in reference counting.
// If projectRoot is empty, MX validation is skipped.
// Uses default 500ms timeout for MX validation.
func NewPostToolHandlerWithMxValidator(diagnostics lsphook.LSPDiagnosticsCollector, analyzer FileAnalyzer, projectRoot string) Handler {
	return NewPostToolHandlerWithMxValidatorAndTimeout(diagnostics, analyzer, projectRoot, 500*time.Millisecond)
}

// NewPostToolHandlerWithMxValidatorAndTimeout creates a PostToolUse handler with MX tag validation
// and a custom timeout for validation.
func NewPostToolHandlerWithMxValidatorAndTimeout(diagnostics lsphook.LSPDiagnosticsCollector, analyzer FileAnalyzer, projectRoot string, timeout time.Duration) Handler {
	var validator mx.Validator
	if projectRoot != "" {
		validator = mx.NewValidator(nil, projectRoot)
	}
	return &postToolHandler{
		diagnostics: diagnostics,
		analyzer:    analyzer,
		mxValidator: validator,
		mxTimeout:   timeout,
	}
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

	// Perform AST file scan after Write/Edit operations (observation-only, never blocks)
	if (input.ToolName == "Write" || input.ToolName == "Edit") && h.analyzer != nil {
		h.runAstScan(ctx, input, metrics)
	}

	// Perform MX tag validation after Write/Edit operations (observation-only, never blocks)
	if (input.ToolName == "Write" || input.ToolName == "Edit") && h.mxValidator != nil {
		h.runMxValidation(ctx, input, metrics)
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

// runAstScan performs an AST-based scan on the modified file.
// Observation-only: never blocks even if an error occurs.
func (h *postToolHandler) runAstScan(ctx context.Context, input *HookInput, metrics map[string]any) {
	// Extract file path from tool input
	var parsed map[string]any
	if err := json.Unmarshal(input.ToolInput, &parsed); err != nil {
		slog.Debug("failed to parse tool input for AST scan", "error", err)
		return
	}

	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		return
	}

	// Perform AST scan (observation-only, errors are only logged)
	result, err := h.analyzer.ScanFile(ctx, filePath, nil)
	if err != nil {
		slog.Debug("AST scan failed (observation-only)",
			"file_path", filePath,
			"error", err,
		)
		return
	}
	if result == nil {
		return
	}

	// Add scan results to metrics
	metrics["ast_scan"] = map[string]any{
		"file":        filepath.Base(filePath),
		"matches":     len(result.Matches),
		"lang":        result.Language,
		"duration_ms": result.Duration.Milliseconds(),
	}

	slog.Debug("AST scan complete",
		"file_path", filepath.Base(filePath),
		"matches", len(result.Matches),
		"lang", result.Language,
	)
}

// runMxValidation performs MX tag validation on the modified file.
// Observation-only: never blocks even if an error occurs.
// AC-POST-001: adds mx_validation metrics to the output.
// AC-POST-002: respects 500ms timeout via context.
func (h *postToolHandler) runMxValidation(ctx context.Context, input *HookInput, metrics map[string]any) {
	// Extract file path from tool input
	var parsed map[string]any
	if err := json.Unmarshal(input.ToolInput, &parsed); err != nil {
		slog.Debug("mx: failed to parse tool input", "error", err)
		return
	}

	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		return
	}

	// Only validate Go files
	if !strings.HasSuffix(filePath, ".go") {
		return
	}

	// Use configured timeout (AC-POST-002: 500ms budget)
	timeout := h.mxTimeout
	if timeout <= 0 {
		timeout = 500 * time.Millisecond
	}

	mxCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	report, err := h.mxValidator.ValidateFile(mxCtx, filePath)
	duration := time.Since(start)

	if err != nil {
		slog.Debug("mx: validation failed (observation-only)",
			"file_path", filePath,
			"error", err,
		)
		metrics["mx_validation"] = map[string]any{
			"status":      "skipped",
			"violations":  []any{},
			"duration_ms": duration.Milliseconds(),
		}
		return
	}

	if report == nil || report.TimedOut {
		// AC-POST-002: return "skipped" on timeout
		metrics["mx_validation"] = map[string]any{
			"status":      "skipped",
			"violations":  []any{},
			"duration_ms": duration.Milliseconds(),
		}
		return
	}

	// Classify status
	status := "pass"
	if report.P1Count() > 0 || report.P2Count() > 0 {
		status = "fail"
	} else if report.P3Count() > 0 || report.P4Count() > 0 {
		status = "warn"
	}

	// Build violations list
	violations := make([]map[string]any, 0, len(report.Violations))
	for _, v := range report.Violations {
		violations = append(violations, map[string]any{
			"func":     v.FuncName,
			"line":     v.Line,
			"priority": v.Priority.String(),
			"tag":      v.MissingTag,
			"blocking": v.Blocking,
		})
	}

	metrics["mx_validation"] = map[string]any{
		"status":      status,
		"violations":  violations,
		"duration_ms": duration.Milliseconds(),
		"file":        filepath.Base(filePath),
		"fallback":    report.Fallback,
	}

	if len(report.Violations) > 0 {
		slog.Info("mx: validation complete",
			"file_path", filepath.Base(filePath),
			"status", status,
			"p1", report.P1Count(),
			"p2", report.P2Count(),
			"p3", report.P3Count(),
			"p4", report.P4Count(),
		)
	} else {
		slog.Debug("mx: validation complete (no violations)",
			"file_path", filepath.Base(filePath),
		)
	}
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
