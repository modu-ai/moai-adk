package hook

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// taskMetrics holds the metrics embedded in a Task tool response.
type taskMetrics struct {
	TokensUsed      int     `json:"tokensUsed"`
	ToolUses        int     `json:"toolUses"`
	DurationSeconds float64 `json:"durationSeconds"`
}

// taskToolResponse is used to parse the Task tool's JSON response.
// Only the metrics field is required; all others are ignored.
type taskToolResponse struct {
	Metrics *taskMetrics `json:"metrics"`
}

// taskMetricsRecord is the JSONL record written to .moai/logs/task-metrics.jsonl.
type taskMetricsRecord struct {
	Timestamp       string  `json:"timestamp"`
	SessionID       string  `json:"session_id"`
	ToolName        string  `json:"tool_name"`
	TokensUsed      int     `json:"tokens_used"`
	ToolUses        int     `json:"tool_uses"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// resolveProjectRoot returns the MoAI project root for task metrics logging.
// It prefers CLAUDE_PROJECT_DIR (set by the Claude Code hook system when invoking
// hook scripts) over input.CWD. Returns empty string when the resolved path does
// not already contain a .moai/ directory, preventing accidental creation of .moai/
// in subdirectories or unrelated directories.
func resolveProjectRoot(input *HookInput) string {
	root := os.Getenv("CLAUDE_PROJECT_DIR")
	if root == "" {
		root = input.CWD
	}
	if root == "" {
		return ""
	}
	// Guard: only write to directories that are already MoAI project roots.
	// This prevents creating .moai/ in non-project directories when CWD is a
	// subdirectory (e.g., internal/cli/) rather than the project root.
	if _, err := os.Stat(filepath.Join(root, ".moai")); err != nil {
		return ""
	}
	return root
}

// logTaskMetrics parses a Task tool response and appends a metrics record to
// .moai/logs/task-metrics.jsonl in the MoAI project root.
// The project root is resolved via resolveProjectRoot to prevent creating .moai/
// directories in non-project locations.
// All errors are best-effort: logged with slog.Warn, never returned.
func logTaskMetrics(input *HookInput) {
	if len(input.ToolResponse) == 0 {
		return
	}

	var resp taskToolResponse
	if err := json.Unmarshal(input.ToolResponse, &resp); err != nil {
		slog.Warn("task metrics: failed to parse ToolResponse",
			"session_id", input.SessionID,
			"error", err,
		)
		return
	}

	if resp.Metrics == nil {
		// No metrics field present – nothing to log.
		return
	}

	record := taskMetricsRecord{
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		SessionID:       input.SessionID,
		ToolName:        input.ToolName,
		TokensUsed:      resp.Metrics.TokensUsed,
		ToolUses:        resp.Metrics.ToolUses,
		DurationSeconds: resp.Metrics.DurationSeconds,
	}

	line, err := json.Marshal(record)
	if err != nil {
		slog.Warn("task metrics: failed to marshal record",
			"session_id", input.SessionID,
			"error", err,
		)
		return
	}
	line = append(line, '\n')

	projectRoot := resolveProjectRoot(input)
	if projectRoot == "" {
		slog.Debug("task metrics: skipping, no MoAI project root found",
			"session_id", input.SessionID,
			"cwd", input.CWD,
		)
		return
	}

	logDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		slog.Warn("task metrics: failed to create log directory",
			"path", logDir,
			"error", err,
		)
		return
	}

	logPath := filepath.Join(logDir, "task-metrics.jsonl")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Warn("task metrics: failed to open log file",
			"path", logPath,
			"error", err,
		)
		return
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(line); err != nil {
		slog.Warn("task metrics: failed to write log record",
			"path", logPath,
			"error", err,
		)
	}
}
