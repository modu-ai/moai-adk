// Package capture — M1 lesson capture pipeline.
// Emits observation candidates to .moai/harness/observations.yaml on SubagentStop events.
// REQ-HRA-001: SubagentStop hook trigger (heuristic-only, no LLM call).
// REQ-HRA-002: Completes within <500ms p95 (benchmark-verified).
//
// @MX:ANCHOR: [AUTO] Capturer.OnSubagentStop is the M1 lesson capture entry point.
// @MX:REASON: [AUTO] fan_in >= 3: capture_test.go, internal/hook/subagent_stop.go (M3 wire), integration_test.go
package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SubagentStopEvent carries the fields emitted by the Claude Code SubagentStop hook.
type SubagentStopEvent struct {
	// AgentName is the subagent display name (required).
	AgentName string

	// AgentType is the subagent type string (e.g., "subagent", "general-purpose").
	AgentType string

	// SessionID is the parent session identifier.
	SessionID string

	// Timestamp is the event occurrence time (UTC).
	Timestamp time.Time

	// ContextHash is a session context identifier used for pattern key building.
	// For large diffs, only a prefix/hash should be stored to avoid file bloat.
	ContextHash string
}

// Config holds the Capturer configuration.
type Config struct {
	// ObservationsPath is the path to .moai/harness/observations.yaml.
	ObservationsPath string
}

// Capturer emits observation candidates to observations.yaml on SubagentStop events.
type Capturer struct {
	cfg Config
}

// New creates a Capturer using the given Config.
func New(cfg Config) *Capturer {
	return &Capturer{cfg: cfg}
}

// OnSubagentStop processes a SubagentStop event and appends an observation entry to
// observations.yaml. Uses flock(2) advisory lock for concurrent SubagentStop safety
// (EC-HRA-002).
//
// Field Naming Policy (spec.md §1.7): snake_case for all YAML fields; `timestamp` is
// the canonical field (NOT `time` or `ts`).
func (c *Capturer) OnSubagentStop(event SubagentStopEvent) error {
	if event.AgentName == "" {
		return fmt.Errorf("capture: agent_name is required")
	}

	ts := event.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	entry := buildObservationEntry(event, ts)

	if err := appendObservation(c.cfg.ObservationsPath, entry); err != nil {
		return fmt.Errorf("capture: %w", err)
	}
	return nil
}

// observationEntry is the YAML representation of a single captured observation.
// Field naming follows spec.md §1.7 canonical policy (snake_case, `timestamp`).
type observationEntry struct {
	AgentName   string    `yaml:"agent_name"`
	AgentType   string    `yaml:"agent_type,omitempty"`
	SessionID   string    `yaml:"session_id,omitempty"`
	Timestamp   time.Time `yaml:"timestamp"`
	ContextHash string    `yaml:"context_hash,omitempty"`
	Status      string    `yaml:"status"`
	Count       int       `yaml:"count"`
}

// buildObservationEntry constructs an observationEntry from a SubagentStopEvent.
func buildObservationEntry(event SubagentStopEvent, ts time.Time) observationEntry {
	contextHash := event.ContextHash
	// Truncate ContextHash to 64 bytes for storage efficiency (large diffs).
	if len(contextHash) > 64 {
		contextHash = contextHash[:64]
	}
	return observationEntry{
		AgentName:   event.AgentName,
		AgentType:   event.AgentType,
		SessionID:   event.SessionID,
		Timestamp:   ts,
		ContextHash: contextHash,
		Status:      "observation",
		Count:       1,
	}
}

// appendObservation appends a YAML list entry to the observations file using
// flock(2) advisory locking to prevent concurrent SubagentStop write races.
func appendObservation(path string, entry observationEntry) error {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdirall %s: %w", dir, err)
		}
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	// Platform-specific advisory lock for exclusive write.
	// Unix: flock(2). Windows: no-op (single-process write semantics; concurrent
	// SubagentStop multi-process scenario is not supported on Windows in W3).
	acquireExclusiveLock(f)
	defer releaseLock(f)

	line := marshalEntry(entry)
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// marshalEntry serializes an observationEntry to a YAML list item string.
// Intentionally hand-rolled to avoid yaml.Marshal overhead and keep sub-500ms
// benchmark target (REQ-HRA-002). Uses RFC3339 for timestamp canonical form.
func marshalEntry(e observationEntry) string {
	ts := e.Timestamp.UTC().Format(time.RFC3339)
	s := fmt.Sprintf("- agent_name: %s\n  timestamp: %s\n  status: %s\n  count: %d\n",
		e.AgentName, ts, e.Status, e.Count)
	if e.AgentType != "" {
		s += fmt.Sprintf("  agent_type: %s\n", e.AgentType)
	}
	if e.SessionID != "" {
		s += fmt.Sprintf("  session_id: %s\n", e.SessionID)
	}
	if e.ContextHash != "" {
		s += fmt.Sprintf("  context_hash: %s\n", e.ContextHash)
	}
	return s
}
