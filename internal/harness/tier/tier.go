// Package tier — M2 tier engine: 4-Tier state machine + observations.yaml write.
// REQ-HRA-003: Tier progression 1x→3x→5x→10x (thresholds per harness.yaml).
// REQ-HRA-004: Atomic status transitions via flock(2) advisory lock.
// REQ-HRA-006: Anti-pattern auto-flag on critical failure.
// REQ-HRA-017: Active count cap (default 50); oldest observation-status entries archived.
//
// @MX:ANCHOR: [AUTO] Engine.Increment is the tier engine entry point.
// @MX:REASON: [AUTO] fan_in >= 3: tier_test.go, capture dispatch (M1→M2), integration_test.go
package tier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Status represents the maturity status of an observation entry.
// Values match the YAML `status` field in observations.yaml (spec.md §1.7).
type Status string

const (
	// StatusObservation is the initial tier (count 1+).
	StatusObservation Status = "observation"

	// StatusHeuristic is count 3+.
	StatusHeuristic Status = "heuristic"

	// StatusRule is count 5+ (also used as seed inject starting point).
	StatusRule Status = "rule"

	// StatusHighConfidence is count 10+ (Tier 4, auto-propose candidate).
	StatusHighConfidence Status = "high-confidence"

	// StatusGraduated is set after L5 approve + Canary PASS.
	StatusGraduated Status = "graduated"

	// StatusAntiPattern is FROZEN — set by FlagAntiPattern (REQ-HRA-006).
	StatusAntiPattern Status = "anti-pattern"

	// StatusArchived is set when active count exceeds cap (REQ-HRA-017).
	StatusArchived Status = "archived"
)

// defaultMaxActive is the default active count cap (REQ-HRA-017).
const defaultMaxActive = 50

// tierThresholds are the canonical thresholds [1, 3, 5, 10] per REQ-HRA-007.
var tierThresholds = [4]int{1, 3, 5, 10}

// ClassifyStatus returns the Status corresponding to the given observation count.
// Uses the canonical fixed thresholds [1, 3, 5, 10] (REQ-HRA-007).
func ClassifyStatus(count int) Status {
	switch {
	case count >= tierThresholds[3]:
		return StatusHighConfidence
	case count >= tierThresholds[2]:
		return StatusRule
	case count >= tierThresholds[1]:
		return StatusHeuristic
	default:
		return StatusObservation
	}
}

// Entry is a single observation entry stored in observations.yaml.
// Field names follow spec.md §1.7 canonical policy (snake_case, `timestamp`).
type Entry struct {
	// AgentName is the subagent name (required).
	AgentName string `yaml:"agent_name"`

	// ContextHash is the context identifier (optional).
	ContextHash string `yaml:"context_hash,omitempty"`

	// Timestamp is the last observed time (canonical field name per §1.7).
	Timestamp time.Time `yaml:"timestamp"`

	// Count is the accumulated observation count.
	Count int `yaml:"count"`

	// Status is the current tier classification string.
	Status Status `yaml:"status"`
}

// AntiPatternEvidence carries evidence for an anti-pattern flag operation.
type AntiPatternEvidence struct {
	// AgentName is the related subagent name.
	AgentName string

	// ContextHash is the context identifier.
	ContextHash string

	// Reason describes why this is flagged as an anti-pattern.
	Reason string
}

// EngineConfig holds the Engine configuration.
type EngineConfig struct {
	// ObservationsPath is the path to .moai/harness/observations.yaml.
	ObservationsPath string

	// AntiPatternsPath is the path to .moai/harness/anti-patterns.yaml.
	AntiPatternsPath string

	// MaxActive is the active count cap (0 → defaultMaxActive=50).
	MaxActive int
}

// Engine manages the 4-Tier state machine and observations.yaml persistence.
//
// @MX:ANCHOR: [AUTO] Engine is the central tier state machine for the harness autonomy system.
// @MX:REASON: [AUTO] fan_in >= 3: tier_test.go, capture dispatch, integration pipeline
type Engine struct {
	cfg       EngineConfig
	maxActive int
}

// NewEngine creates an Engine with the given config.
func NewEngine(cfg EngineConfig) *Engine {
	max := cfg.MaxActive
	if max <= 0 {
		max = defaultMaxActive
	}
	return &Engine{cfg: cfg, maxActive: max}
}

// Increment loads observations.yaml, finds or creates an entry matching
// (AgentName, ContextHash), increments its count, updates status, and saves.
// Uses flock(2) advisory lock for concurrent SubagentStop safety (EC-HRA-002).
func (e *Engine) Increment(entry Entry) error {
	return e.withLock(func() error {
		entries, err := loadEntries(e.cfg.ObservationsPath)
		if err != nil {
			return fmt.Errorf("load: %w", err)
		}

		key := entryKey(entry.AgentName, entry.ContextHash)
		found := false
		for i := range entries {
			if entryKey(entries[i].AgentName, entries[i].ContextHash) == key {
				entries[i].Count++
				entries[i].Status = ClassifyStatus(entries[i].Count)
				entries[i].Timestamp = time.Now().UTC()
				found = true
				break
			}
		}
		if !found {
			newEntry := entry
			newEntry.Count = 1
			newEntry.Status = ClassifyStatus(1)
			if newEntry.Timestamp.IsZero() {
				newEntry.Timestamp = time.Now().UTC()
			}
			entries = append(entries, newEntry)
		}

		// Apply active count cap (REQ-HRA-017)
		entries = applyActiveCap(entries, e.maxActive)

		return saveEntries(e.cfg.ObservationsPath, entries)
	})
}

// Load reads and returns all entries from observations.yaml.
// Returns empty slice if file does not exist.
func (e *Engine) Load() ([]Entry, error) {
	return loadEntries(e.cfg.ObservationsPath)
}

// FlagAntiPattern writes an anti-pattern entry to anti-patterns.yaml (REQ-HRA-006).
// Anti-patterns are FROZEN and cannot be changed by the tier engine.
func (e *Engine) FlagAntiPattern(ev AntiPatternEvidence) error {
	path := e.cfg.AntiPatternsPath
	if path == "" {
		path = filepath.Join(filepath.Dir(e.cfg.ObservationsPath), "anti-patterns.yaml")
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("tier/anti-pattern: mkdirall %s: %w", dir, err)
		}
	}
	line := fmt.Sprintf("- agent_name: %s\n  context_hash: %s\n  status: anti-pattern\n  reason: %q\n  timestamp: %s\n",
		ev.AgentName, ev.ContextHash, ev.Reason, time.Now().UTC().Format(time.RFC3339))
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("tier/anti-pattern: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("tier/anti-pattern: write: %w", err)
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────

// withLock acquires an exclusive flock on observations.yaml for the duration of fn.
func (e *Engine) withLock(fn func() error) error {
	path := e.cfg.ObservationsPath
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("tier: mkdirall %s: %w", dir, err)
		}
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("tier: open lock file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	// Platform-specific advisory exclusive lock (Unix flock; Windows no-op).
	acquireExclusiveLock(f)
	defer releaseLock(f)
	return fn()
}

// entryKey builds a deduplication key for an entry.
func entryKey(agentName, contextHash string) string {
	return agentName + ":" + contextHash
}

// applyActiveCap archives the oldest observation-status entries if active count
// exceeds max (REQ-HRA-017).
func applyActiveCap(entries []Entry, max int) []Entry {
	active := 0
	for _, e := range entries {
		if e.Status != StatusArchived {
			active++
		}
	}
	if active <= max {
		return entries
	}
	// Archive oldest observation-status entries first.
	toArchive := active - max
	for i := range entries {
		if toArchive <= 0 {
			break
		}
		if entries[i].Status == StatusObservation {
			entries[i].Status = StatusArchived
			toArchive--
		}
	}
	return entries
}

// ──────────────────────────────────────────────────────────────────
// YAML persistence (hand-rolled to avoid external deps)
// ──────────────────────────────────────────────────────────────────

// loadEntries reads observations.yaml and returns all entries.
// Returns empty slice if file does not exist.
func loadEntries(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("tier: readfile %s: %w", path, err)
	}
	return parseEntries(string(data))
}

// saveEntries writes all entries to observations.yaml using atomic write-tmp+rename.
func saveEntries(path string, entries []Entry) error {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("tier: mkdirall: %w", err)
		}
	}
	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(marshalEntry(e))
	}
	// Atomic write: write to tmp file then rename.
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(sb.String()), 0o644); err != nil {
		return fmt.Errorf("tier: write tmp %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("tier: rename to %s: %w", path, err)
	}
	return nil
}

// marshalEntry serializes a single Entry to a YAML list item.
// Field naming: timestamp (canonical per §1.7), context_hash (snake_case).
func marshalEntry(e Entry) string {
	ts := e.Timestamp.UTC().Format(time.RFC3339)
	s := fmt.Sprintf("- agent_name: %s\n  timestamp: %s\n  count: %d\n  status: %s\n",
		e.AgentName, ts, e.Count, e.Status)
	if e.ContextHash != "" {
		s += fmt.Sprintf("  context_hash: %s\n", e.ContextHash)
	}
	return s
}

// parseEntries parses the YAML list format produced by marshalEntry.
// This is a minimal line-based parser — not a full YAML parser.
func parseEntries(content string) ([]Entry, error) {
	var entries []Entry
	var current *Entry
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "- agent_name:") {
			if current != nil {
				entries = append(entries, *current)
			}
			current = &Entry{}
			current.AgentName = strings.TrimPrefix(line, "- agent_name:")
			current.AgentName = strings.TrimSpace(current.AgentName)
			continue
		}
		if current == nil {
			continue
		}
		switch {
		case strings.HasPrefix(line, "timestamp:"):
			val := strings.TrimSpace(strings.TrimPrefix(line, "timestamp:"))
			t, err := time.Parse(time.RFC3339, val)
			if err == nil {
				current.Timestamp = t
			}
		case strings.HasPrefix(line, "count:"):
			val := strings.TrimSpace(strings.TrimPrefix(line, "count:"))
			n := 0
			for _, ch := range val {
				if ch >= '0' && ch <= '9' {
					n = n*10 + int(ch-'0')
				}
			}
			current.Count = n
		case strings.HasPrefix(line, "status:"):
			current.Status = Status(strings.TrimSpace(strings.TrimPrefix(line, "status:")))
		case strings.HasPrefix(line, "context_hash:"):
			current.ContextHash = strings.TrimSpace(strings.TrimPrefix(line, "context_hash:"))
		}
	}
	if current != nil {
		entries = append(entries, *current)
	}
	return entries, nil
}
