// Package migration provides dual-mode logging for tracking
// which implementation handles each hook execution
package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HookLogEntry represents a single hook execution log entry
type HookLogEntry struct {
	Hook           string `json:"hook"`
	Implementation string `json:"implementation"`
	BinaryPath     string `json:"binary_path,omitempty"`
	Version        string `json:"version,omitempty"`
	Timestamp      string `json:"timestamp"`
	Duration       string `json:"duration,omitempty"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
}

// HookLogger manages hook execution logging
type HookLogger struct {
	logDir  string
	logFile string
}

// NewHookLogger creates a new hook logger instance
func NewHookLogger(projectDir string) (*HookLogger, error) {
	logDir := filepath.Join(projectDir, ".moai", "logs")

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create dated log file
	dateStr := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("hook-implementation-%s.log", dateStr))

	return &HookLogger{
		logDir:  logDir,
		logFile: logFile,
	}, nil
}

// LogHookExecution logs a hook execution with implementation details
func (l *HookLogger) LogHookExecution(entry *HookLogEntry) error {
	// Set timestamp if not provided
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Open log file in append mode
	f, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Write log entry as JSON line
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// LogHookExecutionFromResult logs a hook execution based on detection result
func (l *HookLogger) LogHookExecutionFromResult(hookName string, result *DetectionResult, duration time.Duration, err error) error {
	entry := &HookLogEntry{
		Hook:           hookName,
		Implementation: result.Type.String(),
		Timestamp:      time.Now().Format(time.RFC3339),
		Duration:       duration.String(),
		Success:        err == nil,
	}

	if result.Type == ImplementationGo {
		entry.BinaryPath = result.BinaryPath
		entry.Version = result.Version
	}

	if err != nil {
		entry.Error = err.Error()
	}

	return l.LogHookExecution(entry)
}

// RotateLogs performs log rotation, keeping logs for 7 days
func (l *HookLogger) RotateLogs() error {
	// Find all log files in log directory
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	// Calculate cutoff date (7 days ago)
	cutoff := time.Now().AddDate(0, 0, -7)

	// Remove old log files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Parse date from filename
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			oldFile := filepath.Join(l.logDir, entry.Name())
			if err := os.Remove(oldFile); err != nil {
				// Log error but continue
				fmt.Fprintf(os.Stderr, "Warning: failed to remove old log file %s: %v\n", oldFile, err)
			}
		}
	}

	return nil
}

// GetLogFile returns the current log file path
func (l *HookLogger) GetLogFile() string {
	return l.logFile
}

// ReadLogs reads and returns all log entries from the current log file
func (l *HookLogger) ReadLogs() ([]HookLogEntry, error) {
	data, err := os.ReadFile(l.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []HookLogEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	var entries []HookLogEntry
	lines := splitLines(data)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var entry HookLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Skip invalid entries
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// splitLines splits byte data into lines
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0

	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}

	// Add last line if not empty
	if start < len(data) {
		lines = append(lines, data[start:])
	}

	return lines
}
