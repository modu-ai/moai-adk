package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// evolutionLogPath returns the default evolution-log.md path.
// To be used in SPEC-V3R2-CON-003.
var _ = filepath.Join // referenced by LoadEvolutionLogs path construction

// LoadEvolutionLogs loads the log list from the evolution-log.md file.
// Returns empty list and nil error if the file doesn't exist.
//
// Reads machine entries (---delimited, snake_case or legacy keys) and
// human-format fenced yaml entries alike; returns an error naming the file,
// line, key, and entry id when a recognized entry has no parseable approval
// timestamp (SPEC-CON-AMEND-APPLY-001 REQ-CAA-006 … REQ-CAA-009).
//
// @MX:ANCHOR: [AUTO] evolution-log read contract shared by the rate limiter, the apply step, and MarkRolledBack
// @MX:REASON: fan_in >= 3 (rateLimiter.Admit, applyAmendment, MarkRolledBack); a reader that drops entries silently blinds the Layer 4 gate
func LoadEvolutionLogs(path string) ([]AmendmentLog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []AmendmentLog{}, nil
		}
		return nil, fmt.Errorf("error reading evolution-log.md: %w", err)
	}
	return parseEvolutionLog(path, string(data))
}

// formatLogEntry renders one machine entry: --- delimiter + snake_case YAML + ---.
func formatLogEntry(log *AmendmentLog) (string, error) {
	yamlData, err := yaml.Marshal(log)
	if err != nil {
		return "", fmt.Errorf("YAML marshaling error: %w", err)
	}
	return fmt.Sprintf("---\n%s---\n", string(yamlData)), nil
}

// AppendEvolutionLog appends a new log to the evolution-log.md file.
// Append-only strategy: preserves existing content and adds to the end of the file.
func AppendEvolutionLog(path string, log *AmendmentLog) error {
	// Validation
	if err := log.Validate(); err != nil {
		return fmt.Errorf("amendment log validation error: %w", err)
	}

	entry, err := formatLogEntry(log)
	if err != nil {
		return err
	}

	// Open file (write-only, create, append)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// MarkRolledBack finds the log for the given rule in evolution-log.md and sets rolled_back to true.
// SPEC-V3R2-CON-002 REQ-CON-002-008 implementation.
//
// Rollback triggers:
// - After amendment, score drops by 0.10 or more in the next SPEC evaluation
// - sync-auditor detects regression
func MarkRolledBack(path, ruleID string, reason string) error {
	logs, err := LoadEvolutionLogs(path)
	if err != nil {
		return err
	}

	// Find the most recent log for the given rule
	var found bool
	for i := len(logs) - 1; i >= 0; i-- {
		if logs[i].RuleID == ruleID && !logs[i].RolledBack {
			now := time.Now()
			logs[i].RolledBack = true
			logs[i].RollbackReason = reason
			logs[i].RollbackAt = &now
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("active log for rule %s not found", ruleID)
	}

	// Rewrite entire file (append-only but rollback modifies existing entry)
	return rewriteEvolutionLog(path, logs)
}

// rewriteEvolutionLog rewrites the file with the given log list.
func rewriteEvolutionLog(path string, logs []AmendmentLog) error {
	// Check directory
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Write to temporary file
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Write header
	if _, err := f.WriteString("# Evolution Log\n\n"); err != nil {
		return err
	}

	// Write each log
	for _, log := range logs {
		yamlData, err := yaml.Marshal(log)
		if err != nil {
			return fmt.Errorf("YAML marshaling error: %w", err)
		}

		entry := fmt.Sprintf("---\n%s---\n", string(yamlData))
		if _, err := f.WriteString(entry); err != nil {
			return err
		}
	}

	// Atomic replacement (rename)
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("error replacing file: %w", err)
	}

	return nil
}

// GenerateLogID generates a new log ID.
// LEARN-YYYYMMDD-NNN format.
func GenerateLogID(now time.Time, lastLogs []AmendmentLog) string {
	dateStr := now.Format("20060102")

	// Find the last sequence number for the given date
	maxSeq := 0
	for _, log := range lastLogs {
		prefix := "LEARN-" + dateStr + "-"
		if strings.HasPrefix(log.ID, prefix) {
			seqStr := strings.TrimPrefix(log.ID, prefix)
			var seq int
			if _, err := fmt.Sscanf(seqStr, "%d", &seq); err == nil {
				if seq > maxSeq {
					maxSeq = seq
				}
			}
		}
	}

	// Next sequence number
	return fmt.Sprintf("LEARN-%s-%03d", dateStr, maxSeq+1)
}
