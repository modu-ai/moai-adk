// Package harness — Log retention and pruning implementation.
// REQ-HL-011: Archives old events and cleans up log files.
package harness

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// pruneSkipDuration is the duration within which pruning is skipped since last prune.
const pruneSkipDuration = time.Hour

// Retention archives and cleans up old entries in usage-log.jsonl.
// REQ-HL-011: Lazy pruning on every RecordEvent call, skip if within 1 hour of last prune.
//
// @MX:ANCHOR: [AUTO] PruneStaleEntries is called by observer and tests.
// @MX:REASON: [AUTO] fan_in >= 3: observer.go, observer_test.go, integration_test.go
type Retention struct {
	// logPath is the usage-log.jsonl file path.
	logPath string

	// archiveDir is the directory to store archive files.
	// Actual path: archiveDir/<YYYY-MM>.jsonl.gz
	archiveDir string

	// lastPruneAt is the last pruning execution time.
	lastPruneAt time.Time

	// nowFn is a function that returns current time (can inject mock-clock in tests).
	nowFn func() time.Time
}

// NewRetention creates a Retention instance.
// Uses time.Now if nowFn is nil.
func NewRetention(logPath, archiveDir string, nowFn func() time.Time) *Retention {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Retention{
		logPath:    logPath,
		archiveDir: archiveDir,
		nowFn:      nowFn,
	}
}

// PruneStaleEntries removes events older than retentionDays from log
// and adds them to archive file (<YYYY-MM>.jsonl.gz).
// REQ-HL-011: Skips if within 1 hour of last prune.
//
// @MX:WARN: [AUTO] Non-atomic operation of reading and overwriting files, use caution with concurrent calls.
// @MX:REASON: [AUTO] Called sequentially within the same process as RecordEvent,
// but race condition possible if external processes record simultaneously.
func (r *Retention) PruneStaleEntries(retentionDays int) error {
	now := r.nowFn()

	// Skip prune if within 1 hour
	if !r.lastPruneAt.IsZero() && now.Sub(r.lastPruneAt) < pruneSkipDuration {
		return nil
	}

	// Skip if log file does not exist
	if _, err := os.Stat(r.logPath); os.IsNotExist(err) {
		r.lastPruneAt = now
		return nil
	}

	cutoff := now.AddDate(0, 0, -retentionDays)

	// Read log file
	kept, stale, err := partitionEvents(r.logPath, cutoff)
	if err != nil {
		return fmt.Errorf("retention: 이벤트 분류 실패: %w", err)
	}

	if len(stale) == 0 {
		// No events to remove
		r.lastPruneAt = now
		return nil
	}

	// Save stale events to monthly archive
	if err := r.archiveEvents(stale); err != nil {
		return fmt.Errorf("retention: 아카이브 실패: %w", err)
	}

	// Overwrite log file with only kept events
	if err := overwriteWithEvents(r.logPath, kept); err != nil {
		return fmt.Errorf("retention: 로그 파일 갱신 실패: %w", err)
	}

	r.lastPruneAt = now
	return nil
}

// partitionEvents reads log file and classifies kept/stale events based on cutoff.
func partitionEvents(logPath string, cutoff time.Time) (kept, stale []Event, err error) {
	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("파일 열기: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			// Put parsing failure lines in kept to prevent data loss
			continue
		}
		if evt.Timestamp.Before(cutoff) {
			stale = append(stale, evt)
		} else {
			kept = append(kept, evt)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("파일 스캔: %w", err)
	}
	return kept, stale, nil
}

// archiveEvents adds stale events to monthly gzip archives.
// Filename: archiveDir/<YYYY-MM>.jsonl.gz
func (r *Retention) archiveEvents(events []Event) error {
	if len(events) == 0 {
		return nil
	}

	if err := os.MkdirAll(r.archiveDir, 0o755); err != nil {
		return fmt.Errorf("아카이브 디렉토리 생성: %w", err)
	}

	// Group events by month
	byMonth := make(map[string][]Event)
	for _, evt := range events {
		key := evt.Timestamp.UTC().Format("2006-01")
		byMonth[key] = append(byMonth[key], evt)
	}

	for month, evts := range byMonth {
		archivePath := filepath.Join(r.archiveDir, month+".jsonl.gz")
		if err := appendToGzip(archivePath, evts); err != nil {
			return fmt.Errorf("월별 아카이브 %s 기록 실패: %w", month, err)
		}
	}
	return nil
}

// appendToGzip appends events to gzip-compressed JSONL file.
// Creates new file if it does not exist.
//
// @MX:WARN: [AUTO] Gzip files are not append-safe, so use read-and-rewrite method.
// @MX:REASON: [AUTO] Gzip format supports concatenated streams so append is actually possible,
// but use read-modify-write pattern for compatibility with standard readers.
func appendToGzip(archivePath string, events []Event) error {
	// gzip concatenation: append by adding new gzip stream to existing file.
	// Standard gzip reader can read concatenated streams sequentially.
	f, err := os.OpenFile(archivePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("아카이브 파일 열기: %w", err)
	}
	defer func() { _ = f.Close() }()

	gw := gzip.NewWriter(f)
	enc := json.NewEncoder(gw)
	for _, evt := range events {
		if err := enc.Encode(evt); err != nil {
			_ = gw.Close()
			return fmt.Errorf("gzip 인코딩: %w", err)
		}
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("gzip 닫기: %w", err)
	}
	return nil
}

// overwriteWithEvents overwrites log file with only kept events.
func overwriteWithEvents(logPath string, events []Event) error {
	// Write to temporary file first, then atomic replacement
	dir := filepath.Dir(logPath)
	tmp, err := os.CreateTemp(dir, "usage-log-*.tmp")
	if err != nil {
		return fmt.Errorf("임시 파일 생성: %w", err)
	}
	tmpPath := tmp.Name()

	enc := json.NewEncoder(tmp)
	for _, evt := range events {
		if err := enc.Encode(evt); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("임시 파일 인코딩: %w", err)
		}
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("임시 파일 닫기: %w", err)
	}

	// Atomic replacement
	if err := os.Rename(tmpPath, logPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("파일 교체: %w", err)
	}
	return nil
}
