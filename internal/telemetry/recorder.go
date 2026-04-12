package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// recordMu guards all writes to telemetry files.
// A single mutex is sufficient because telemetry writes are infrequent and fast.
var recordMu sync.Mutex

// HashContext returns the first 8 characters of the SHA-256 hex digest of input.
// The output is irreversible and safe to store without PII risk.
func HashContext(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])[:8]
}

// RecordSkillUsage appends a usage record to the daily telemetry file.
//
// Storage layout:
//
//	<projectRoot>/.moai/evolution/telemetry/usage-YYYY-MM-DD.jsonl
//
// The directory is created if it does not exist. All errors are returned to
// the caller. The write is protected by a package-level mutex for concurrent
// safety.
//
// @MX:ANCHOR: [AUTO] RecordSkillUsage is the primary telemetry write path
// @MX:REASON: [AUTO] fan_in>=3, called from hook/post_tool.go, hook/stop.go, and tests
func RecordSkillUsage(projectRoot string, r UsageRecord) error {
	// Create telemetry directory (parent .moai/ must exist — validated by resolveProjectRoot).
	dir := filepath.Join(projectRoot, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("telemetry: mkdir: %w", err)
	}

	// Daily rotation: one file per UTC day.
	dayKey := r.Timestamp.UTC().Format("2006-01-02")
	path := filepath.Join(dir, "usage-"+dayKey+".jsonl")

	// Marshal the record.
	line, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("telemetry: marshal: %w", err)
	}

	// Atomic append with mutex.
	recordMu.Lock()
	defer recordMu.Unlock()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("telemetry: open: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("telemetry: write: %w", err)
	}
	return nil
}

// LoadBySession reads today's (and optionally yesterday's) telemetry files and
// returns only the records matching the given sessionID. Returns nil on any error.
func LoadBySession(projectRoot, sessionID string) ([]UsageRecord, error) {
	dir := filepath.Join(projectRoot, ".moai", "evolution", "telemetry")

	today := time.Now().UTC().Format("2006-01-02")
	candidates := []string{
		filepath.Join(dir, "usage-"+today+".jsonl"),
	}
	// Also check yesterday in case the session spans midnight.
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	candidates = append(candidates, filepath.Join(dir, "usage-"+yesterday+".jsonl"))

	var result []UsageRecord
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // file may not exist
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var rec UsageRecord
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				continue
			}
			if rec.SessionID == sessionID {
				result = append(result, rec)
			}
		}
	}
	return result, nil
}

// PruneOldFiles deletes telemetry JSONL files whose date (parsed from filename)
// is older than retentionDays from now. Non-telemetry files are ignored.
//
// If the telemetry directory does not exist, PruneOldFiles returns nil.
func PruneOldFiles(projectRoot string, retentionDays int) error {
	dir := filepath.Join(projectRoot, ".moai", "evolution", "telemetry")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("telemetry: prune readdir: %w", err)
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Only process files matching usage-YYYY-MM-DD.jsonl
		if !strings.HasPrefix(name, "usage-") || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		// Extract date string: "usage-YYYY-MM-DD.jsonl" → "YYYY-MM-DD"
		dateStr := strings.TrimPrefix(name, "usage-")
		dateStr = strings.TrimSuffix(dateStr, ".jsonl")

		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Not a valid date — skip
			continue
		}

		if fileDate.Before(cutoff) {
			path := filepath.Join(dir, name)
			if removeErr := os.Remove(path); removeErr != nil {
				return fmt.Errorf("telemetry: prune remove %s: %w", path, removeErr)
			}
		}
	}
	return nil
}
