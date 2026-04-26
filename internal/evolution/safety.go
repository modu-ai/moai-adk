package evolution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// FrozenPaths contains the set of project-root-relative file paths that the
// evolution engine is never permitted to modify.  These are constitutional
// files whose integrity is critical to system operation.
//
// The list is intentionally defined as a package-level variable (not a
// constant slice) so tests can extend it, but it is not exported for
// production mutation.
var frozenPaths = []string{
	"CLAUDE.md",
	".claude/rules/moai/core/moai-constitution.md",
	".claude/rules/moai/core/agent-common-protocol.md",
	// Agency fork manifest is immutably protected on par with the constitution.
	".agency/fork-manifest.yaml",
}

// frozenDirPrefixes contains directory prefixes whose contents are protected.
var frozenDirPrefixes = []string{
	".claude/rules/moai/core/",
	// Protect the full Agency constitution and related rules (Agency Constitution §2 FROZEN Zone).
	".claude/rules/agency/",
}

// CheckFrozenGuard reports whether targetFile is in the frozen zone.
//
// The check covers:
//   - Absolute paths (always rejected).
//   - Paths containing ".." components after filepath.Clean (path traversal rejected).
//   - Exact path matches against frozenPaths.
//   - Files inside frozen directory prefixes.
//   - Targets with a ":frontmatter" suffix (YAML frontmatter modification).
//
// Returns ErrFrozenPath if blocked, nil if allowed.
func CheckFrozenGuard(targetFile string) error {
	// Normalise the path separator.
	target := filepath.ToSlash(strings.TrimSpace(targetFile))

	// Absolute paths are always rejected (e.g. /etc/passwd, C:\Windows\...).
	// filepath.IsAbs does not treat "/foo" as absolute on Windows, so
	// Unix-style absolute paths (leading /) and Windows-style (leading \) are checked separately.
	if filepath.IsAbs(targetFile) ||
		strings.HasPrefix(targetFile, "/") ||
		strings.HasPrefix(targetFile, `\`) {
		return ErrFrozenPath
	}

	// Reject path traversal attempts containing ".." components.
	// If the original slash path contains /../ or ends with /.., treat it as a path manipulation attempt.
	// Check on the original slash path before applying filepath.Clean.
	if strings.Contains(target, "/../") ||
		strings.HasSuffix(target, "/..") ||
		strings.HasPrefix(target, "../") ||
		target == ".." {
		return ErrFrozenPath
	}

	// If the path still starts with ".." after filepath.Clean, it is an attempt to escape to a parent.
	cleaned := filepath.ToSlash(filepath.Clean(target))
	if strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return ErrFrozenPath
	}

	// Frontmatter modification is always blocked.
	if strings.HasSuffix(target, ":frontmatter") {
		return ErrFrozenPath
	}

	// Strip any zone suffix (e.g. ":zone-id") for path comparison.
	filePath := cleaned
	if idx := strings.LastIndex(cleaned, ":"); idx > 0 {
		// Only strip if what precedes ":" is a valid file-path character sequence.
		candidate := cleaned[:idx]
		if !strings.Contains(candidate, " ") {
			filePath = candidate
		}
	}

	// Exact match check.
	for _, fp := range frozenPaths {
		if filePath == fp {
			return ErrFrozenPath
		}
	}

	// Prefix directory check.
	for _, prefix := range frozenDirPrefixes {
		if strings.HasPrefix(filePath+"/", prefix) || strings.HasPrefix(filePath, prefix) {
			return ErrFrozenPath
		}
	}

	return nil
}

// rateMu guards the full Read→mutate→Write sequence in UpdateRateLimit.
// Without the lock, another goroutine could interleave between Read and Write,
// causing the counter to be lost. The entire sequence is protected by a single lock.
//
// @MX:WARN: [AUTO] Global mutex: all UpdateRateLimit calls are serialized
// @MX:REASON: [AUTO] Prevents TOCTOU race conditions; the rate state file requires a single writer guarantee
var rateMu sync.Mutex

// rateStatePath returns the path to the rate-limit state file.
func rateStatePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "evolution", "rate_state.yaml")
}

// ReadRateState reads the current rate-limit state from disk.
// Returns a zero-value RateState if the file does not exist.
func ReadRateState(projectRoot string) (*RateState, error) {
	path := rateStatePath(projectRoot)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &RateState{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("evolution: read rate state: %w", err)
	}

	var state RateState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("evolution: parse rate state: %w", err)
	}
	return &state, nil
}

// WriteRateState persists the rate-limit state atomically (write to temp file
// then rename to prevent partial writes).
func WriteRateState(projectRoot string, state *RateState) error {
	path := rateStatePath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("evolution: mkdir for rate state: %w", err)
	}

	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("evolution: marshal rate state: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("evolution: write temp rate state: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("evolution: rename rate state: %w", err)
	}
	return nil
}

// currentWeekStart returns the Monday of the current UTC week as YYYY-MM-DD.
func currentWeekStart() string {
	now := time.Now().UTC()
	wd := int(now.Weekday())
	if wd == 0 {
		wd = 7
	}
	monday := now.AddDate(0, 0, -(wd - 1)).Truncate(24 * time.Hour)
	return monday.Format("2006-01-02")
}

// CheckRateLimit checks whether a new proposal may be generated.
// Returns ErrRateLimit if the weekly quota (MaxProposalsPerWeek) has been
// exhausted.  Does NOT check per-file cooldown (see CheckFileCooldown).
func CheckRateLimit(projectRoot string) error {
	state, err := ReadRateState(projectRoot)
	if err != nil {
		// Conservative: on error, block rather than overflow.
		return ErrRateLimit
	}

	weekStart := currentWeekStart()
	if state.WeekStart != weekStart {
		// New week — reset is handled lazily on UpdateRateLimit.
		return nil
	}

	if state.ProposalsThisWeek >= MaxProposalsPerWeek {
		return ErrRateLimit
	}
	return nil
}

// CheckFileCooldown checks whether a new proposal targeting targetFile is
// allowed given the 24-hour per-file cooldown.
// Returns ErrRateLimit if the file is within the cooldown window.
func CheckFileCooldown(projectRoot, targetFile string) error {
	if targetFile == "" {
		return nil
	}

	state, err := ReadRateState(projectRoot)
	if err != nil {
		return ErrRateLimit
	}

	if state.LastProposalTimes == nil {
		return nil
	}

	lastStr, ok := state.LastProposalTimes[targetFile]
	if !ok {
		return nil
	}

	last, err := time.Parse(time.RFC3339, lastStr)
	if err != nil {
		return nil
	}

	cooldown := time.Duration(ProposalCooldownHours) * time.Hour
	if time.Since(last) < cooldown {
		return ErrRateLimit
	}
	return nil
}

// UpdateRateLimit increments the weekly proposal counter and records the
// last proposal time for targetFile.  If the state belongs to a past week,
// it is reset before incrementing.
//
// targetFile may be empty to update only the weekly counter.
// The entire Read→mutate→Write sequence is protected by rateMu to prevent race conditions.
func UpdateRateLimit(projectRoot, targetFile string) error {
	rateMu.Lock()
	defer rateMu.Unlock()

	state, err := ReadRateState(projectRoot)
	if err != nil {
		state = &RateState{}
	}

	weekStart := currentWeekStart()
	if state.WeekStart != weekStart {
		// New week — reset counter.
		state.WeekStart = weekStart
		state.ProposalsThisWeek = 0
		state.LastProposalTimes = nil
	}

	state.ProposalsThisWeek++

	if targetFile != "" {
		if state.LastProposalTimes == nil {
			state.LastProposalTimes = make(map[string]string)
		}
		state.LastProposalTimes[targetFile] = time.Now().UTC().Format(time.RFC3339)
	}

	return WriteRateState(projectRoot, state)
}
