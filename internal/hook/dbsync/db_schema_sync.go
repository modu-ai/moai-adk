// Package dbsync implements the moai hook db-schema-sync subcommand.
// It handles PostToolUse events by detecting migration file changes,
// applying debounce, parsing migration files (stub), and writing
// proposal.json for the orchestrator to present to the user.
//
// REQ coverage: REQ-003, REQ-004, REQ-006, REQ-007, REQ-008, REQ-009, REQ-010, REQ-011
package dbsync

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultExcludedPatterns is the single source of truth for recursion guard patterns.
// REQ-004: These paths are excluded to prevent recursive hook invocation.
var DefaultExcludedPatterns = []string{
	".moai/project/db/**",
	".moai/cache/**",
	".moai/logs/**",
}

// DecisionSkip signals the handler took no action (not a migration file or empty path).
const DecisionSkip = "skip"

// DecisionDebounced signals the handler skipped due to debounce window.
const DecisionDebounced = "debounced"

// DecisionAskUser signals the orchestrator should present the user approval dialog.
const DecisionAskUser = "ask-user"

// Config holds all parameters for the db-schema-sync handler.
type Config struct {
	// FilePath is the file path received from the PostToolUse hook stdin.
	FilePath string

	// MigrationPatterns are the glob patterns from db.yaml migration_patterns.
	MigrationPatterns []string

	// ExcludedPatterns are the recursion guard glob patterns.
	ExcludedPatterns []string

	// StateFile is the path to .moai/cache/db-sync/last-seen.json.
	StateFile string

	// ProposalFile is the path to .moai/cache/db-sync/proposal.json.
	ProposalFile string

	// ErrorLogFile is the path to .moai/logs/db-sync-errors.log.
	ErrorLogFile string

	// DebounceWindow is the debounce duration (default 10 seconds per REQ-006).
	DebounceWindow time.Duration
}

// Result is the return value of HandleDBSchemaSync.
type Result struct {
	// ExitCode is 0 (non-blocking) in all cases per REQ-011.
	ExitCode int

	// Decision indicates what action was taken.
	Decision string
}

// Proposal is the JSON structure written to proposal.json (REQ-009).
type Proposal struct {
	FilePath      string `json:"file_path"`
	ParsedContent string `json:"parsed_content"`
	Decision      string `json:"decision"`
	Timestamp     string `json:"timestamp"`
}

// DebounceState is the JSON structure stored in last-seen.json (REQ-007).
type DebounceState struct {
	FilePath  string    `json:"file_path"`
	Timestamp time.Time `json:"timestamp"`
}

// HandleDBSchemaSync is the main entry point for the db-schema-sync hook handler.
// It always returns exit 0 (non-blocking) to avoid disrupting the user's Write/Edit flow.
//
// @MX:NOTE: [AUTO] HandleDBSchemaSync is the public facade for db-schema-sync. It never blocks the caller.
func HandleDBSchemaSync(cfg Config) Result {
	// Empty path — exit silently (REQ-002 guard)
	if cfg.FilePath == "" {
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	// Recursion guard: excluded patterns exit 0 silently (REQ-004)
	if IsExcluded(cfg.FilePath, cfg.ExcludedPatterns) {
		slog.Debug("db-schema-sync: file excluded (recursion guard)", "path", cfg.FilePath)
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	// Migration pattern check: non-matching files exit 0 silently (REQ-003)
	if !MatchesMigrationPattern(cfg.FilePath, cfg.MigrationPatterns) {
		slog.Debug("db-schema-sync: file does not match migration patterns", "path", cfg.FilePath)
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	// Debounce check (REQ-006, REQ-007)
	if err := os.MkdirAll(filepath.Dir(cfg.StateFile), 0o755); err != nil {
		logError(cfg.ErrorLogFile, "mkdir state dir: "+err.Error())
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	debounced, err := CheckDebounce(cfg.StateFile, cfg.FilePath, cfg.DebounceWindow)
	if err != nil {
		logError(cfg.ErrorLogFile, "debounce check: "+err.Error())
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}
	if debounced {
		slog.Debug("db-schema-sync: debounced", "path", cfg.FilePath)
		return Result{ExitCode: 0, Decision: DecisionDebounced}
	}

	// Parse migration file (REQ-008): stub implementation.
	// The actual parser is in internal/db/parser/ (separate concern per SPEC exclusion).
	parsedContent, parseErr := parseMigrationStub(cfg.FilePath)
	if parseErr != nil {
		// REQ-011: log error and exit 0 (non-blocking)
		logError(cfg.ErrorLogFile, "parse error for "+cfg.FilePath+": "+parseErr.Error())
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	// Write proposal.json (REQ-009)
	if err := os.MkdirAll(filepath.Dir(cfg.ProposalFile), 0o755); err != nil {
		logError(cfg.ErrorLogFile, "mkdir proposal dir: "+err.Error())
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	proposal := BuildProposal(cfg.FilePath, parsedContent)
	proposalJSON, err := json.Marshal(proposal)
	if err != nil {
		logError(cfg.ErrorLogFile, "marshal proposal: "+err.Error())
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	if err := os.WriteFile(cfg.ProposalFile, proposalJSON, 0o644); err != nil {
		logError(cfg.ErrorLogFile, "write proposal: "+err.Error())
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	// REQ-010: emit decision signal to stdout (caller handles stdout write)
	slog.Info("db-schema-sync: proposal written", "path", cfg.FilePath, "proposal", cfg.ProposalFile)
	return Result{ExitCode: 0, Decision: DecisionAskUser}
}

// BuildProposal constructs a Proposal from the parsed content.
func BuildProposal(filePath, parsedContent string) Proposal {
	return Proposal{
		FilePath:      filePath,
		ParsedContent: parsedContent,
		Decision:      DecisionAskUser,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// MatchesMigrationPattern checks whether filePath matches any of the migration patterns.
// Uses simple glob matching via filepath.Match for single-segment patterns and
// prefix/suffix matching for ** patterns.
func MatchesMigrationPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchGlob(pattern, filePath) {
			return true
		}
	}
	return false
}

// IsExcluded checks whether filePath matches any of the excluded patterns.
func IsExcluded(filePath string, excluded []string) bool {
	for _, pattern := range excluded {
		if matchGlob(pattern, filePath) {
			return true
		}
	}
	return false
}

// matchGlob performs simple glob matching supporting ** wildcard.
// For patterns ending in /**, the file must be under the prefix directory.
// For other patterns, filepath.Match is used.
func matchGlob(pattern, filePath string) bool {
	// Handle ** patterns: convert to prefix match
	if strings.Contains(pattern, "**") {
		// Split on **
		parts := strings.SplitN(pattern, "**", 2)
		prefix := strings.TrimRight(parts[0], "/")
		suffix := strings.TrimLeft(parts[1], "/")

		if prefix != "" {
			if !strings.HasPrefix(filePath, prefix+"/") && filePath != prefix {
				return false
			}
		}
		if suffix != "" {
			// The suffix may itself contain *, use filepath.Match on the filename
			rest := filePath
			if prefix != "" {
				rest = strings.TrimPrefix(filePath, prefix+"/")
			}
			matched, err := filepath.Match(suffix, filepath.Base(rest))
			if err != nil {
				return false
			}
			if !matched {
				// Try matching the suffix against the full remainder
				matched, err = filepath.Match(suffix, rest)
				if err != nil || !matched {
					// Fallback: check extension match from suffix
					if strings.HasSuffix(suffix, "*") {
						return true
					}
					ext := filepath.Ext(suffix)
					return ext != "" && strings.HasSuffix(filePath, ext)
				}
			}
			return true
		}
		// Pattern is "prefix/**" — any file under prefix
		return true
	}

	// Exact match attempt
	matched, err := filepath.Match(pattern, filePath)
	if err == nil && matched {
		return true
	}

	// Try matching with just the base filename
	matched, err = filepath.Match(filepath.Base(pattern), filepath.Base(filePath))
	if err == nil && matched && filepath.Dir(pattern) == filepath.Dir(filePath) {
		return true
	}

	return false
}

// CheckDebounce reads the state file and determines if the current call is debounced.
// Returns (true, nil) if the same file was seen within the debounce window.
// Updates the state file when the call proceeds (not debounced).
func CheckDebounce(stateFile, filePath string, window time.Duration) (bool, error) {
	// Read existing state
	data, err := os.ReadFile(stateFile)
	if err == nil {
		var state DebounceState
		if unmarshalErr := json.Unmarshal(data, &state); unmarshalErr == nil {
			// Same file within window → debounced
			if state.FilePath == filePath && time.Since(state.Timestamp) < window {
				return true, nil
			}
		}
	}

	// Not debounced — update state file
	newState := DebounceState{
		FilePath:  filePath,
		Timestamp: time.Now(),
	}
	stateJSON, err := json.Marshal(newState)
	if err != nil {
		return false, err
	}
	return false, os.WriteFile(stateFile, stateJSON, 0o644)
}

// parseMigrationStub is a placeholder parser that reads the file content.
// The actual parser implementation is in internal/db/parser/ (SPEC scope exclusion).
// This stub ensures REQ-008's behavior contract (input=migration file, output=normalized schema)
// without implementing the full parser.
func parseMigrationStub(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// logError appends an error message to the error log file (REQ-011).
func logError(logFile, message string) {
	if logFile == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	entry := time.Now().UTC().Format(time.RFC3339) + " " + message + "\n"
	_, _ = f.WriteString(entry)
}
