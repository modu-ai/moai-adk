// audit_transition.go — status-transition-audit.log consumer.
//
// SPEC-OBSERVE-HYGIENE-001 M1 (REQ-OBH-001, AC-OBH-001).
//
// The hook .claude/hooks/moai/status-transition-ownership.sh appends one line
// per Write/Edit/MultiEdit on a SPEC artifact to
// .moai/logs/status-transition-audit.log. The line carries the status value
// captured at WRITE TIME — a signal the git-history lint
// (OwnershipTransitionRule in lint_ownership.go) structurally cannot see,
// because git only records the final committed state, not the intermediate
// write-site values.
//
// This consumer cross-checks each logged status value against the canonical
// 8-value Status enum (the domain vocabulary owned by the Status Transition
// Ownership Matrix, spec-frontmatter-schema.md). A value that is non-empty,
// not the recognized "<file absent — Write creating new>" sentinel, and not a
// canonical status is surfaced as an INFO-severity finding. INFO severity keeps
// the check non-breaking: a write-site anomaly is observational, not a gate.
//
// Graceful contract (EC-1): an absent, corrupt, or unparseable log degrades to
// zero findings and NEVER returns an error. Unknown line shapes are skipped and
// counted (parseStats.SkippedUnparseable), never fatal.
package spec

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FindingStatusTransitionLogAnomaly is the INFO finding emitted when a parsed
// status-transition-audit.log line carries a status value that is non-empty,
// not the new-file sentinel, and not in the canonical Status enum. JSON
// consumers depend on this string staying stable.
const FindingStatusTransitionLogAnomaly = "StatusTransitionLogAnomaly"

// fileAbsentSentinel is the status value the hook records when a Write creates
// a brand-new SPEC artifact (the file did not exist pre-write). It represents
// the "(none) → draft" creation transition (owned by manager-spec) and is an
// EXPECTED write-site value, not an anomaly.
const fileAbsentSentinel = "<file absent — Write creating new>"

// transitionLogEntry is one parsed line of status-transition-audit.log.
type transitionLogEntry struct {
	Timestamp string // ISO-8601 UTC, e.g. "2026-07-09T08:00:00Z"
	Tool      string // Write | Edit | MultiEdit
	FilePath  string // absolute path to the SPEC artifact
	RawStatus string // status value as captured at write time (untrimmed)
}

// transitionLogParseStats summarizes a parse run for observability.
type transitionLogParseStats struct {
	LogAbsent         bool // true when the log file does not exist (graceful no-op)
	ReadError         bool // true when the log exists but could not be read (graceful no-op)
	Parsed            int  // count of successfully parsed lines
	SkippedUnparseable int // count of lines that did not match the expected shape (EC-1)
}

// transitionLogLinePattern captures the fixed prefix of each log line:
// "<ts> [status-transition-ownership] <Tool> ". The remainder (path +
// " status=" + status) is split separately because both path and status may
// contain spaces (e.g. the sentinel "<file absent — Write creating new>").
var transitionLogLinePattern = regexp.MustCompile(
	`^(\S+)\s+\[status-transition-ownership\]\s+(\S+)\s+(.*)$`,
)

// specIDExtractPattern finds a SPEC-ID substring within a file path. The path
// is not required to be the full canonical form (it carries a directory prefix
// and a filename suffix), so the pattern is non-anchored.
var specIDExtractPattern = regexp.MustCompile(
	`SPEC-[A-Z0-9]+(-[A-Z0-9]+)*-[0-9]{3}`,
)

// parseStatusTransitionLog reads and parses the status-transition-audit.log at
// logPath. Graceful contract (EC-1): an absent log returns (nil, stats{LogAbsent:
// true}, nil) — never an error. A read failure returns (nil, stats{ReadError:
// true}, nil) — never an error. Lines that do not match the expected shape are
// skipped and counted in SkippedUnparseable; parsing never aborts on a bad line.
func parseStatusTransitionLog(logPath string) ([]transitionLogEntry, transitionLogParseStats, error) {
	var stats transitionLogParseStats
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			stats.LogAbsent = true
			return nil, stats, nil
		}
		// Corrupt / permission error — degrade gracefully, never error (EC-1).
		stats.ReadError = true
		return nil, stats, nil
	}

	lines := strings.Split(string(data), "\n")
	var entries []transitionLogEntry
	for _, line := range lines {
		// Skip blank / trailing-newline artifacts.
		if strings.TrimSpace(line) == "" {
			continue
		}
		m := transitionLogLinePattern.FindStringSubmatch(line)
		if m == nil {
			stats.SkippedUnparseable++
			continue
		}
		ts, tool, rest := m[1], m[2], m[3]
		// rest = "<path> status=<status>". Split on the LAST " status=" so a
		// path containing " status=" (extremely unlikely) does not mis-split.
		idx := strings.LastIndex(rest, " status=")
		if idx < 0 {
			// No status marker — malformed; skip + count.
			stats.SkippedUnparseable++
			continue
		}
		filePath := strings.TrimSpace(rest[:idx])
		rawStatus := rest[idx+len(" status="):]
		entries = append(entries, transitionLogEntry{
			Timestamp: ts,
			Tool:      tool,
			FilePath:  filePath,
			RawStatus: rawStatus,
		})
		stats.Parsed++
	}
	return entries, stats, nil
}

// crossCheckTransitionLog is the project-level cross-check wired into Audit().
// It parses <baseDir>/.moai/logs/status-transition-audit.log and emits an INFO
// finding for each logged status value that is non-empty, not the new-file
// sentinel, and not in the canonical Status enum. When opts.FilterSpec is set,
// only findings whose extracted SPEC-ID matches the filter are surfaced.
//
// Absent / corrupt log → zero findings (graceful no-op per EC-1).
func crossCheckTransitionLog(baseDir string, opts AuditOptions) []DriftFinding {
	logPath := filepath.Join(baseDir, ".moai", "logs", "status-transition-audit.log")
	entries, _, err := parseStatusTransitionLog(logPath)
	if err != nil {
		// parseStatusTransitionLog never returns a non-nil error today, but
		// defend against a future change so the cross-check stays graceful.
		return nil
	}
	if len(entries) == 0 {
		return nil
	}

	var findings []DriftFinding
	for _, e := range entries {
		specID := specIDExtractPattern.FindString(e.FilePath)
		if specID == "" {
			// Path did not carry a SPEC-ID (not a SPEC artifact) — skip.
			continue
		}
		status := strings.TrimSpace(e.RawStatus)
		if status == "" {
			// Empty status = transient capture (hook grepped mid-keystroke) — skip.
			continue
		}
		if status == fileAbsentSentinel {
			// New-file Write sentinel — expected, not an anomaly.
			continue
		}
		if IsValidStatus(status) {
			// Canonical status value — no anomaly.
			continue
		}
		// Non-canonical status at a write-site — a signal git history cannot see.
		if opts.FilterSpec != "" && specID != opts.FilterSpec {
			continue
		}
		findings = append(findings, DriftFinding{
			SpecID:      specID,
			FindingType: FindingStatusTransitionLogAnomaly,
			Severity:    "INFO",
			Details: map[string]any{
				"raw_status": e.RawStatus,
				"tool":       e.Tool,
				"file_path":  e.FilePath,
				"timestamp":  e.Timestamp,
				"reason": "logged status value is not in the canonical Status enum " +
					"(write-site capture; may be a transient mid-edit value)",
			},
		})
	}
	return findings
}
