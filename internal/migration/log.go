package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogEntry represents a migration log entry.
// REQ-V3R2-RT-007-014: recorded in structured JSONL format.
type LogEntry struct {
	Version     int        `json:"version"`
	Name        string     `json:"name"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      string     `json:"result"` // success, failed, rolled-back, skipped
	Details     string     `json:"details"`
}

const logFileName = "migrations.log"

// Append appends a migration log entry.
// REQ-V3R2-RT-007-014: uses JSONL append-only format.
func Append(projectRoot string, entry LogEntry) error {
	logsDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("logs 디렉터리 생성 실패: %w", err)
	}

	logFile := filepath.Join(logsDir, logFileName)

	// Serialize as a JSONL row.
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("로그 엔트리 직렬화 실패: %w", err)
	}

	// File append (open write-only).
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("로그 파일 열기 실패: %w", err)
	}
	defer func() { _ = f.Close() }()

	// JSONL row + newline character.
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("로그 파일 쓰기 실패: %w", err)
	}

	return nil
}

// LastApplied returns the most recently applied (successful) migration log entry.
// REQ-V3R2-RT-007-015: looks up the last applied log entry.
func LastApplied(projectRoot string) (*LogEntry, error) {
	logFile := filepath.Join(projectRoot, ".moai", "logs", logFileName)

	data, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Log file absent -> return nil.
			return nil, nil
		}
		return nil, fmt.Errorf("로그 파일 읽기 실패: %w", err)
	}

	// Parse JSONL line by line.
	lines := parseJSONL(data)
	var lastSuccess *LogEntry
	for _, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Skip on parse failure.
			continue
		}
		if entry.Result == "success" || entry.Result == "rolled-back" {
			lastSuccess = &entry
		}
	}

	return lastSuccess, nil
}

// parseJSONL parses JSONL data line by line.
func parseJSONL(data []byte) [][]byte {
	lines := [][]byte{}
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
