package trace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionSummary aggregates TraceEntry records from a session into a report.
type SessionSummary struct {
	// SessionID is the session identifier.
	SessionID string
	// TotalHooks is the total number of hook invocations.
	TotalHooks int
	// Duration is the wall-clock duration from first to last entry.
	Duration time.Duration
	// EventBreakdown maps event type to invocation count.
	EventBreakdown map[string]int
	// DecisionBreakdown maps decision value to count.
	DecisionBreakdown map[string]int
	// Top5Slowest holds the five slowest hook executions (desc order).
	Top5Slowest []TraceEntry
	// ErrorCount is the number of entries with non-empty Error field.
	ErrorCount int
	// Errors contains the error messages encountered during the session.
	Errors []string
}

// GenerateSummary reads the trace JSONL file for the given session and returns
// an aggregated SessionSummary. Returns an error only for unrecoverable I/O
// failures; missing trace files produce an empty summary without error.
func GenerateSummary(logDir, sessionID string) (*SessionSummary, error) {
	path := filepath.Join(logDir, fmt.Sprintf("trace-%s.jsonl", sessionID))

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No trace file — return empty summary.
			return &SessionSummary{
				SessionID:         sessionID,
				EventBreakdown:    map[string]int{},
				DecisionBreakdown: map[string]int{},
			}, nil
		}
		return nil, fmt.Errorf("open trace file %q: %w", path, err)
	}
	defer f.Close()

	summary := &SessionSummary{
		SessionID:         sessionID,
		EventBreakdown:    map[string]int{},
		DecisionBreakdown: map[string]int{},
	}

	var entries []TraceEntry
	var first, last time.Time

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry TraceEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Skip malformed lines gracefully.
			continue
		}

		entries = append(entries, entry)
		summary.TotalHooks++
		summary.EventBreakdown[entry.Event]++

		if entry.Decision != "" {
			summary.DecisionBreakdown[entry.Decision]++
		}

		if entry.Error != "" {
			summary.ErrorCount++
			summary.Errors = append(summary.Errors, entry.Error)
		}

		// Track wall-clock range.
		if first.IsZero() || entry.Timestamp.Before(first) {
			first = entry.Timestamp
		}
		if entry.Timestamp.After(last) {
			last = entry.Timestamp
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan trace file: %w", err)
	}

	if !first.IsZero() && !last.IsZero() {
		summary.Duration = last.Sub(first)
	}

	summary.Top5Slowest = top5Slowest(entries)

	return summary, nil
}

// top5Slowest returns up to 5 entries with the highest DurationMs, descending.
func top5Slowest(entries []TraceEntry) []TraceEntry {
	if len(entries) == 0 {
		return nil
	}

	sorted := make([]TraceEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].DurationMs > sorted[j].DurationMs
	})

	if len(sorted) > 5 {
		sorted = sorted[:5]
	}
	return sorted
}

// FormatMarkdown formats the summary as a Markdown report string.
func (s *SessionSummary) FormatMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Session Summary: %s\n\n", s.SessionID))
	sb.WriteString(fmt.Sprintf("**Total Hook Invocations:** %d\n\n", s.TotalHooks))
	sb.WriteString(fmt.Sprintf("**Session Duration:** %s\n\n", s.Duration.Round(time.Millisecond)))

	// Event breakdown.
	sb.WriteString("## Event Breakdown\n\n")
	if len(s.EventBreakdown) == 0 {
		sb.WriteString("_No events recorded._\n\n")
	} else {
		events := sortedKeys(s.EventBreakdown)
		for _, ev := range events {
			sb.WriteString(fmt.Sprintf("- **%s**: %d\n", ev, s.EventBreakdown[ev]))
		}
		sb.WriteString("\n")
	}

	// Decision breakdown.
	sb.WriteString("## Decision Breakdown\n\n")
	if len(s.DecisionBreakdown) == 0 {
		sb.WriteString("_No decisions recorded._\n\n")
	} else {
		decisions := sortedKeys(s.DecisionBreakdown)
		for _, d := range decisions {
			sb.WriteString(fmt.Sprintf("- **%s**: %d\n", d, s.DecisionBreakdown[d]))
		}
		sb.WriteString("\n")
	}

	// Top 5 slowest.
	sb.WriteString("## Top 5 Slowest Hook Executions\n\n")
	if len(s.Top5Slowest) == 0 {
		sb.WriteString("_No hook executions recorded._\n\n")
	} else {
		sb.WriteString("| # | Event | Handler | Tool | Duration (ms) |\n")
		sb.WriteString("|---|-------|---------|------|---------------|\n")
		for i, e := range s.Top5Slowest {
			sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %d |\n",
				i+1, e.Event, e.Handler, e.Tool, e.DurationMs))
		}
		sb.WriteString("\n")
	}

	// Errors.
	sb.WriteString(fmt.Sprintf("## Errors (%d)\n\n", s.ErrorCount))
	if s.ErrorCount == 0 {
		sb.WriteString("_No errors recorded._\n\n")
	} else {
		for _, e := range s.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// sortedKeys returns the keys of m sorted alphabetically.
func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
