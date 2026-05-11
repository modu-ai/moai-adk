package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogEntry는 마이그레이션 로그 엔트리입니다.
// REQ-V3R2-RT-007-014: structured JSONL format으로 기록됩니다.
type LogEntry struct {
	Version     int        `json:"version"`
	Name        string     `json:"name"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      string     `json:"result"` // success, failed, rolled-back, skipped
	Details     string     `json:"details"`
}

const logFileName = "migrations.log"

// Append는 마이그레이션 로그를 추가합니다.
// REQ-V3R2-RT-007-014: JSONL append-only format입니다.
func Append(projectRoot string, entry LogEntry) error {
	logsDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("logs 디렉터리 생성 실패: %w", err)
	}

	logFile := filepath.Join(logsDir, logFileName)

	// JSONL 행으로 직렬화
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("로그 엔트리 직렬화 실패: %w", err)
	}

	// File append (쓰기 전용으로 열기)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("로그 파일 열기 실패: %w", err)
	}
	defer func() { _ = f.Close() }()

	// JSONL 행 + 개행 문자
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("로그 파일 쓰기 실패: %w", err)
	}

	return nil
}

// LastApplied는 마지막으로 성공적으로 적용된 마이그레이션 로그를 반환합니다.
// REQ-V3R2-RT-007-015: last applied log entry를 조회합니다.
func LastApplied(projectRoot string) (*LogEntry, error) {
	logFile := filepath.Join(projectRoot, ".moai", "logs", logFileName)

	data, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 로그 파일 부재 = nil 반환
			return nil, nil
		}
		return nil, fmt.Errorf("로그 파일 읽기 실패: %w", err)
	}

	// JSONL 라인별 파싱
	lines := parseJSONL(data)
	var lastSuccess *LogEntry
	for _, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// 파싱 실패 시 건너뜀
			continue
		}
		if entry.Result == "success" || entry.Result == "rolled-back" {
			lastSuccess = &entry
		}
	}

	return lastSuccess, nil
}

// parseJSONL은 JSONL 데이터를 라인별로 파싱합니다.
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
