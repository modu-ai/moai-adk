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
	"github.com/modu-ai/moai-adk/internal/hook/quality"
	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
	"github.com/modu-ai/moai-adk/internal/loop"
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
	// cfg provides access to ralph configuration for lint_as_instruction (REQ-LAI-003).
	// If nil, lint_as_instruction defaults to true.
	cfg ConfigProvider
	// feedbackCh is an optional bounded channel for emitting loop.Feedback events.
	// REQ-LL-003: PostTool hook emits diagnostics to both systemMessage and this channel.
	// If nil, channel emission is skipped (no-op).
	feedbackCh *loop.FeedbackChannel
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

// NewPostToolHandlerWithConfig creates a PostToolUse handler with full configuration
// including a ConfigProvider for reading ralph settings (lint_as_instruction, etc.).
// Use this constructor when systemMessage injection for LSP diagnostics is required.
func NewPostToolHandlerWithConfig(diagnostics lsphook.LSPDiagnosticsCollector, analyzer FileAnalyzer, projectRoot string, timeout time.Duration, cfg ConfigProvider) Handler {
	var validator mx.Validator
	if projectRoot != "" {
		validator = mx.NewValidator(nil, projectRoot)
	}
	return &postToolHandler{
		diagnostics: diagnostics,
		analyzer:    analyzer,
		mxValidator: validator,
		mxTimeout:   timeout,
		cfg:         cfg,
	}
}

// NewPostToolHandlerWithFeedbackChannel creates a PostToolUse handler that emits
// diagnostics to both the agent systemMessage and the provided loop.FeedbackChannel.
// REQ-LL-003: PostTool hook connects to LoopController via the feedback channel.
// If feedbackCh is nil, channel emission is a no-op (backwards compatible).
func NewPostToolHandlerWithFeedbackChannel(
	diagnostics lsphook.LSPDiagnosticsCollector,
	analyzer FileAnalyzer,
	projectRoot string,
	timeout time.Duration,
	cfg ConfigProvider,
	feedbackCh *loop.FeedbackChannel,
) Handler {
	var validator mx.Validator
	if projectRoot != "" {
		validator = mx.NewValidator(nil, projectRoot)
	}
	return &postToolHandler{
		diagnostics: diagnostics,
		analyzer:    analyzer,
		mxValidator: validator,
		mxTimeout:   timeout,
		cfg:         cfg,
		feedbackCh:  feedbackCh,
	}
}

// EventType returns EventPostToolUse.
func (h *postToolHandler) EventType() EventType {
	return EventPostToolUse
}

// Handle processes a PostToolUse event. It collects metrics about the tool
// execution (tool name, output size) and returns them in the Data field.
// For Write/Edit tools, also collects LSP diagnostics per REQ-HOOK-150 and
// injects a systemMessage when lint_as_instruction is enabled (REQ-LAI-001).
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

	// Collect Agent (formerly Task) subagent metrics (SPEC-MONITOR-001).
	// Best-effort: errors are logged internally and never propagated.
	// Since v2.1.63 Claude Code renamed Task → Agent; accept both for backward compatibility.
	if input.ToolName == "Agent" || input.ToolName == "Task" {
		logTaskMetrics(input)
	}

	var systemMessage string
	var collectedDiags []lsphook.Diagnostic

	// Collect LSP diagnostics for Write/Edit operations (REQ-HOOK-150, REQ-HOOK-153).
	// Also generates systemMessage if lint_as_instruction is enabled (REQ-LAI-001).
	if (input.ToolName == "Write" || input.ToolName == "Edit") && h.diagnostics != nil {
		systemMessage, collectedDiags = h.collectDiagnosticsWithInstructionAndReturn(ctx, input, metrics)
	}

	// REQ-LL-003: emit diagnostics to FeedbackChannel for LoopController consumption.
	if (input.ToolName == "Write" || input.ToolName == "Edit") && h.feedbackCh != nil {
		h.emitToFeedbackChannel(collectedDiags)
	}

	// Perform AST file scan after Write/Edit operations (observation-only, never blocks).
	// When lint_as_instruction is enabled, security findings are also appended to
	// systemMessage alongside LSP errors (REQ-LAI-008).
	if (input.ToolName == "Write" || input.ToolName == "Edit") && h.analyzer != nil {
		if astResult := h.runAstScan(ctx, input, metrics); astResult != nil && h.lintAsInstructionEnabled() {
			// Extract file path from tool input for the security message header.
			var parsed map[string]any
			if err := json.Unmarshal(input.ToolInput, &parsed); err == nil {
				if fp, ok := parsed["file_path"].(string); ok {
					systemMessage = quality.AppendAstSecurityFindings(systemMessage, fp, astResult)
				}
			}
		}
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
		SystemMessage: systemMessage,
		Data:          jsonData,
	}, nil
}

// runAstScan performs an AST-based scan on the modified file.
// Observation-only: never blocks even if an error occurs.
// Returns the ScanResult so callers can use it for systemMessage injection (REQ-LAI-008).
func (h *postToolHandler) runAstScan(ctx context.Context, input *HookInput, metrics map[string]any) *astgrep.ScanResult {
	// Extract file path from tool input
	var parsed map[string]any
	if err := json.Unmarshal(input.ToolInput, &parsed); err != nil {
		slog.Debug("failed to parse tool input for AST scan", "error", err)
		return nil
	}

	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		return nil
	}

	// Perform AST scan (observation-only, errors are only logged)
	result, err := h.analyzer.ScanFile(ctx, filePath, nil)
	if err != nil {
		slog.Debug("AST scan failed (observation-only)",
			"file_path", filePath,
			"error", err,
		)
		return nil
	}
	if result == nil {
		return nil
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
	return result
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

// lintAsInstructionEnabled returns true when the lint_as_instruction feature
// should be active. If no config is set, it defaults to true (opt-out model).
func (h *postToolHandler) lintAsInstructionEnabled() bool {
	if h.cfg == nil {
		return true
	}
	cfg := h.cfg.Get()
	if cfg == nil {
		return true
	}
	return cfg.Ralph.LintAsInstruction
}

// warnAsInstructionEnabled returns true when warnings should also be injected
// as systemMessage instructions (REQ-LAI-006). Defaults to false.
func (h *postToolHandler) warnAsInstructionEnabled() bool {
	if h.cfg == nil {
		return false
	}
	cfg := h.cfg.Get()
	if cfg == nil {
		return false
	}
	return cfg.Ralph.WarnAsInstruction
}

// collectDiagnosticsWithInstruction collects LSP diagnostics for the modified file,
// populates the metrics map (REQ-LAI-005, backward compatible), and returns a
// formatted systemMessage string when lint_as_instruction is enabled (REQ-LAI-001).
// This is observation-only and MUST NOT block per REQ-HOOK-153.
func (h *postToolHandler) collectDiagnosticsWithInstruction(ctx context.Context, input *HookInput, metrics map[string]any) string {
	msg, _ := h.collectDiagnosticsWithInstructionAndReturn(ctx, input, metrics)
	return msg
}

// collectDiagnosticsWithInstructionAndReturn is the internal implementation that also
// returns collected diagnostics for FeedbackChannel emission (REQ-LL-003).
func (h *postToolHandler) collectDiagnosticsWithInstructionAndReturn(ctx context.Context, input *HookInput, metrics map[string]any) (string, []lsphook.Diagnostic) {
	// Extract file path from tool input
	var parsed map[string]any
	if err := json.Unmarshal(input.ToolInput, &parsed); err != nil {
		slog.Debug("failed to parse tool input for diagnostics", "error", err)
		return "", nil
	}

	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		return "", nil
	}

	// Get diagnostics (observation only, never block)
	diagnostics, err := h.diagnostics.GetDiagnostics(ctx, filePath)
	if err != nil {
		slog.Debug("diagnostics collection failed (observation only)",
			"file_path", filePath,
			"error", err,
		)
		return "", nil
	}

	// Calculate severity counts
	counts := h.diagnostics.GetSeverityCounts(diagnostics)

	// REQ-LAI-005: Add diagnostic counts to metrics (always, regardless of lint_as_instruction).
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

	// REQ-LAI-003: skip systemMessage injection when lint_as_instruction is false.
	if !h.lintAsInstructionEnabled() {
		return "", diagnostics
	}

	// REQ-LAI-001 / REQ-LAI-002 / REQ-LAI-004 / REQ-LAI-006 / REQ-LAI-007:
	// Format diagnostics as an instruction for the AI.
	msg := quality.FormatDiagnosticsAsInstructionWithFile(filePath, diagnostics, counts, h.warnAsInstructionEnabled())
	return msg, diagnostics
}

// emitToFeedbackChannel converts collected lsphook.Diagnostic entries to
// lsp.Diagnostic and emits a Feedback event to the feedback channel.
// REQ-LL-003: observation-only, never blocks.
func (h *postToolHandler) emitToFeedbackChannel(diags []lsphook.Diagnostic) {
	if h.feedbackCh == nil || len(diags) == 0 {
		return
	}
	lspDiags := convertHookDiagsToLSP(diags)
	fb := loop.Feedback{
		LSPDiagnostics: lspDiags,
		Phase:          loop.PhaseImplement, // PostTool fires during implement phase
	}
	h.feedbackCh.Send(fb)
}

// convertHookDiagsToLSP converts lsphook.Diagnostic to lsp.Diagnostic.
// lsphook uses string severity; lsp uses integer severity per LSP 3.17 spec.
func convertHookDiagsToLSP(diags []lsphook.Diagnostic) []lsp.Diagnostic {
	if len(diags) == 0 {
		return nil
	}
	result := make([]lsp.Diagnostic, 0, len(diags))
	for _, d := range diags {
		lspDiag := lsp.Diagnostic{
			Code:    d.Code,
			Source:  d.Source,
			Message: d.Message,
		}
		switch d.Severity {
		case lsphook.SeverityError:
			lspDiag.Severity = lsp.SeverityError
		case lsphook.SeverityWarning:
			lspDiag.Severity = lsp.SeverityWarning
		case lsphook.SeverityInformation:
			lspDiag.Severity = lsp.SeverityInfo
		case lsphook.SeverityHint:
			lspDiag.Severity = lsp.SeverityHint
		default:
			lspDiag.Severity = lsp.SeverityInfo
		}
		result = append(result, lspDiag)
	}
	return result
}
