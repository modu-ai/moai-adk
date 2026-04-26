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
	"strconv"
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

// maxMigrationFileSize is the single source of truth for the parseMigrationStub
// file size ceiling. Migration files larger than this are not read in full — the
// size-guard branch logs and returns (ParsedContent="", Truncated=true) so the
// pipeline never holds an arbitrary-sized file in memory.
// REQ-H1-001: 1 MiB (well above realistic Prisma schema / Alembic version sizes,
// small enough to prevent memory pressure from malformed or malicious input).
const maxMigrationFileSize = 1 << 20

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
// @MX:ANCHOR: Path traversal guard — cleaned paths that escape the project root (absolute or "../")
//
//	are rejected so matchGlob cannot be coerced into reading files outside the project tree.
//	@MX:REASON: matchGlob accepts unnormalized paths; without this guard a crafted file_path like
//	"migrations/../../../etc/passwd.sql" passes pattern matching and would be fed into parseMigrationStub.
func HandleDBSchemaSync(cfg Config) Result {
	// Empty path — exit silently (REQ-002 guard)
	if cfg.FilePath == "" {
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}

	// Path traversal guard: reject paths that escape to a parent directory after
	// filepath.Clean normalization. Absolute paths remain accepted because Claude Code's
	// tool_input.file_path is always absolute in production; matchGlob's prefix check
	// handles out-of-project absolute paths downstream.
	cleaned := filepath.Clean(cfg.FilePath)
	sep := string(filepath.Separator)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+sep) || strings.HasPrefix(cleaned, "../") {
		slog.Debug("db-schema-sync: rejected path traversal", "path", cfg.FilePath, "cleaned", cleaned)
		return Result{ExitCode: 0, Decision: DecisionSkip}
	}
	// Normalize to forward-slash form AFTER the traversal guard. Downstream
	// matchGlob / IsExcluded operate on forward-slash glob patterns from db.yaml
	// (e.g. `migrations/**/*.sql`), so Windows `migrations\001.sql` produced by
	// filepath.Clean would never match without this conversion.
	cfg.FilePath = filepath.ToSlash(cleaned)

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

	debounced, err := checkDebounceWithLog(cfg.StateFile, cfg.FilePath, cfg.DebounceWindow, cfg.ErrorLogFile)
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
	parsed, parseErr := parseMigrationStubWithLog(cfg.FilePath, cfg.ErrorLogFile)
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

	// REQ-H1-002: when the size guard rejected the file, parsed.ParsedContent == ""
	// and parsed.Truncated == true. BuildProposal still runs so the orchestrator can
	// surface the oversized event to the user (decision remains ask-user per REQ-009).
	proposal := BuildProposal(cfg.FilePath, parsed.ParsedContent)
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

// @MX:NOTE BuildProposal wraps the parsed migration content into a Proposal struct awaiting user approval.
// Maintains a flat JSON-marshalable layout so it can be serialized directly as proposal.json (REQ-009).
//
// Inputs:
//   - filePath: migration file path (the original tool_input.file_path from Claude Code)
//   - parsedContent: the parsing result string returned by parseMigrationStub
//     (empty string when the size guard was triggered; Truncated=true in that state)
//
// Output:
//   - Proposal: Decision is always "ask-user" (DecisionAskUser), Timestamp is UTC RFC3339
//
// Side effects: none (pure function).
func BuildProposal(filePath, parsedContent string) Proposal {
	return Proposal{
		FilePath:      filePath,
		ParsedContent: parsedContent,
		Decision:      DecisionAskUser,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// @MX:NOTE MatchesMigrationPattern checks whether file_path matches any of the DB migration globs.
// The `**` wildcard is handled by matchGlob using prefix/suffix splitting.
//
// Inputs:
//   - filePath: the file path to check (caller is responsible for applying filepath.Clean first)
//   - patterns: migration_patterns from db.yaml (e.g., ["prisma/schema.prisma", "migrations/**/*.sql"])
//
// Output:
//   - bool: true if any pattern matches, false if none match
//
// Side effects: none (pure function, read-only).
func MatchesMigrationPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchGlob(pattern, filePath) {
			return true
		}
	}
	return false
}

// @MX:NOTE IsExcluded checks whether file_path matches any of the recursion guard (REQ-004) patterns.
// DefaultExcludedPatterns (.moai/project/db/**, .moai/cache/**, .moai/logs/**) are the default blocklist;
// if the hook processes files at these paths, the automatic schema.md rewrite would trigger the hook again,
// causing an infinite loop — these paths must always be blocked.
//
// Inputs:
//   - filePath: the file path to check
//   - excluded: list of exclusion glob patterns (typically DefaultExcludedPatterns)
//
// Output:
//   - bool: true if any exclusion pattern matches (hook handles as exit 0 skip)
//
// Side effects: none (pure function, read-only).
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

// isWithinDebounceWindow reports whether the persisted state file records the
// given filePath with a timestamp inside the debounce window. Any I/O or decode
// error is treated as "no fresh state" (returns false) since the caller will
// proceed to establish a new window under a lock. This helper collapses the
// two read-then-check sites in checkDebounceWithLog (fast path + under-lock
// re-check) into a single implementation.
func isWithinDebounceWindow(stateFile, filePath string, window time.Duration) bool {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return false
	}
	var state DebounceState
	if err := json.Unmarshal(data, &state); err != nil {
		return false
	}
	return state.FilePath == filePath && time.Since(state.Timestamp) < window
}

// @MX:NOTE CheckDebounce는 동일 filePath가 window 내에 이미 관측되었는지 확인하고,
// 관측되지 않았다면 stateFile을 원자적으로 갱신한다(임시 파일 + os.Rename, POSIX rename
// 원자성에 의존). 동일 stateFile을 대상으로 동시에 호출되더라도 정확히 한 호출만
// debounced=false를 반환한다(REQ-H2-002).
//
// 입력:
//   - stateFile: 디바운스 상태 JSON 파일의 절대 경로 (.moai/cache/db-sync/last-seen.json)
//   - filePath: PostToolUse 훅이 보고한 마이그레이션 파일 경로
//   - window: 디바운스 윈도우 (기본 10초; 테스트에서는 50ms 등 짧은 값 사용 가능)
//
// 출력:
//   - debounced: true이면 window 내 중복 이벤트 (호출자는 조용히 종료), false면 신규 이벤트
//   - error: 현재 구현은 항상 nil을 반환한다 (I/O 실패도 안전 기본값 (true, nil)로 흡수;
//     REQ-H2-003). 미래 시그니처 호환성을 위해 error를 유지한다.
//
// 부작용:
//   - stateFile을 temp-file + os.Rename으로 원자적 교체 (정상 경로)
//   - I/O 실패 시 내부 로그 없음 (공개 API는 logFile 없이 호출됨;
//     HandleDBSchemaSync 경유 시 checkDebounceWithLog가 ErrorLogFile에 기록)
func CheckDebounce(stateFile, filePath string, window time.Duration) (bool, error) {
	return checkDebounceWithLog(stateFile, filePath, window, "")
}

// checkDebounceWithLog is the implementation body with an explicit ErrorLogFile
// path for testability. When logFile is empty, the I/O failure branch writes
// nothing to disk and still returns the safe default (true, nil) per REQ-H2-003.
//
// Concurrency contract (REQ-H2-001/002): exactly one concurrent caller targeting
// the same (stateFile, filePath, window) returns debounced=false. The rest return
// debounced=true. Mutual exclusion is established via a companion lock file
// opened with O_CREATE|O_EXCL — the OS serializes the create-or-fail call so a
// single winner is selected atomically. The winner additionally persists the
// new state via temp-file + os.Rename so readers never observe a torn write.
func checkDebounceWithLog(stateFile, filePath string, window time.Duration, logFile string) (bool, error) {
	// Fast path: if the already-persisted state indicates we are inside the
	// window, skip the lock acquisition entirely. This is safe because the
	// fast-path result is re-validated after acquiring the lock.
	if isWithinDebounceWindow(stateFile, filePath, window) {
		return true, nil
	}

	// Ensure the state directory exists so the lock file can be created.
	stateDir := filepath.Dir(stateFile)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		logError(logFile, "CheckDebounce: mkdir state dir: "+err.Error())
		return true, nil
	}

	// Acquire the advisory lock. Exactly one concurrent caller succeeds here;
	// the others see os.IsExist and return the safe debounced=true result.
	// REQ-H2-002 — "exactly one winner" is enforced by O_EXCL's atomicity.
	//
	// Note: the companion ".lock" file lives beside stateFile, which is
	// conventionally under .moai/cache/db-sync/ — covered by
	// DefaultExcludedPatterns' ".moai/cache/**" entry, so the lock itself
	// cannot re-trigger the hook even if a user's editor briefly Write/Edits it.
	lockPath := stateFile + ".lock"
	lockFD, lockErr := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if lockErr != nil {
		if os.IsExist(lockErr) {
			// Another caller holds the lock — they will establish the window.
			return true, nil
		}
		// REQ-H2-003: any other lock acquisition failure → safe default.
		logError(logFile, "CheckDebounce: acquire lock: "+lockErr.Error())
		return true, nil
	}
	_ = lockFD.Close()
	defer func() { _ = os.Remove(lockPath) }()

	// Re-check under the lock: a process that released the lock just before
	// our acquisition may have established a fresh window.
	if isWithinDebounceWindow(stateFile, filePath, window) {
		return true, nil
	}

	// Winner path — persist new state atomically via temp-file + os.Rename.
	newState := DebounceState{
		FilePath:  filePath,
		Timestamp: time.Now(),
	}
	stateJSON, err := json.Marshal(newState)
	if err != nil {
		logError(logFile, "CheckDebounce: marshal state: "+err.Error())
		return true, nil
	}

	tmpFile, err := os.CreateTemp(stateDir, ".last-seen-*.json.tmp")
	if err != nil {
		logError(logFile, "CheckDebounce: create temp: "+err.Error())
		return true, nil
	}
	tmpName := tmpFile.Name()
	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(tmpName)
		}
	}()

	if _, werr := tmpFile.Write(stateJSON); werr != nil {
		_ = tmpFile.Close()
		logError(logFile, "CheckDebounce: write temp: "+werr.Error())
		return true, nil
	}
	if cerr := tmpFile.Close(); cerr != nil {
		logError(logFile, "CheckDebounce: close temp: "+cerr.Error())
		return true, nil
	}

	// os.Rename is atomic within the same filesystem (POSIX guarantees).
	// os.CreateTemp(stateDir, ...) keeps both files on the same filesystem.
	if rerr := os.Rename(tmpName, stateFile); rerr != nil {
		logError(logFile, "CheckDebounce: rename: "+rerr.Error())
		return true, nil
	}
	renamed = true
	return false, nil
}

// parseMigrationResult captures the outcome of parseMigrationStub: the parsed
// content string and whether the input was rejected by the size guard. Both
// fields are always set so callers can make decisions without either-or
// ambiguity (REQ-H1-002). The actual parser in internal/db/parser/ may adopt
// a richer shape later; this struct is the minimal contract parseMigrationStub
// honors today.
type parseMigrationResult struct {
	ParsedContent string
	Truncated     bool
}

// parseMigrationStub is a placeholder parser that reads the file content.
// The actual parser implementation is in internal/db/parser/ (SPEC scope exclusion).
// This stub ensures REQ-008's behavior contract (input=migration file, output=normalized schema)
// without implementing the full parser. Files exceeding maxMigrationFileSize are
// rejected by os.Stat pre-check BEFORE any whole-file read (REQ-H1-002, REQ-H1-003).
func parseMigrationStub(filePath string) (parseMigrationResult, error) {
	return parseMigrationStubWithLog(filePath, "")
}

// parseMigrationStubWithLog is the implementation body of parseMigrationStub with an
// explicit ErrorLogFile path for testability. When logFile is empty, the size-guard
// branch writes nothing to disk (used by callers that do not track a log path).
func parseMigrationStubWithLog(filePath, logFile string) (parseMigrationResult, error) {
	// REQ-H1-003: size judgment via os.Stat BEFORE any full read.
	info, statErr := os.Stat(filePath)
	if statErr != nil {
		return parseMigrationResult{}, statErr
	}
	if info.Size() > maxMigrationFileSize {
		// REQ-H1-002: log + return (parsed_content="", truncated=true), no full read.
		logError(logFile, "parseMigrationStub: file exceeds maxMigrationFileSize="+strconv.FormatInt(info.Size(), 10)+" path="+filePath)
		return parseMigrationResult{ParsedContent: "", Truncated: true}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return parseMigrationResult{}, err
	}
	return parseMigrationResult{ParsedContent: string(data), Truncated: false}, nil
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
	defer func() { _ = f.Close() }()
	entry := time.Now().UTC().Format(time.RFC3339) + " " + message + "\n"
	_, _ = f.WriteString(entry)
}
