package taxonomy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// staleWrapPrefix and staleWrapSuffix are the fixed tags used to wrap stale memory content.
// Format is intentionally stable — moai-memory.md documents the interpretation guide.
const (
	staleWrapPrefix = "<system-reminder>This memory may be stale; verify against current state before acting on it.\n"
	staleWrapSuffix = "\n</system-reminder>"
)

// StaleReport describes a single memory file whose mtime exceeds the staleness threshold.
type StaleReport struct {
	// Path is the absolute or relative path of the stale memory file.
	Path string
	// Age is the time elapsed since the file was last modified.
	Age time.Duration
	// Wrapped is the file's original content enclosed in <system-reminder> tags.
	Wrapped string
}

// DetectStale scans dir for memory files (.md) whose mtime exceeds thresholdHours
// relative to now. The now parameter is injected for deterministic testing
// (REQ-EXT001-006; avoids flakiness from wall-clock drift).
//
// Returns an empty slice (not an error) when dir does not exist or is empty.
// Symlinks are not followed.
func DetectStale(dir string, thresholdHours int, now time.Time) ([]StaleReport, error) {
	if thresholdHours <= 0 {
		thresholdHours = config.DefaultMemoryStalenessHours
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("taxonomy: DetectStale read dir %s: %w", dir, err)
	}

	var reports []StaleReport
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		// Skip symlinks (security: no path traversal).
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}

		fullPath := filepath.Join(dir, e.Name())
		fi, err := e.Info()
		if err != nil {
			continue
		}

		age := now.Sub(fi.ModTime())
		threshold := time.Duration(thresholdHours) * time.Hour

		if age >= threshold {
			// Read content for wrapping.
			data, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}
			wrapped := staleWrapPrefix + string(data) + staleWrapSuffix
			reports = append(reports, StaleReport{
				Path:    fullPath,
				Age:     age,
				Wrapped: wrapped,
			})
		}
	}

	return reports, nil
}

// AggregateWarning returns a warning message for the given stale reports.
// When the number of stale files is >= DefaultMemoryStaleAggregateThreshold (10),
// it emits a single aggregated warning rather than per-file entries
// (REQ-EXT001-017, AC-EXT001-08).
//
// Returns empty string when reports is empty.
func AggregateWarning(reports []StaleReport) string {
	if len(reports) == 0 {
		return ""
	}

	if len(reports) >= config.DefaultMemoryStaleAggregateThreshold {
		return fmt.Sprintf(
			"MEMORY_STALE_AGGREGATED: %d memory files are stale (>%dh). Verify all agent memories before acting on them.",
			len(reports),
			config.DefaultMemoryStalenessHours,
		)
	}

	// Per-file warnings for < threshold.
	var sb strings.Builder
	for _, r := range reports {
		fmt.Fprintf(&sb, "MEMORY_STALE: %s (age: %s) — verify before use\n", r.Path, r.Age.Round(time.Minute))
	}
	return strings.TrimRight(sb.String(), "\n")
}
