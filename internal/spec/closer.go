// Package spec — atomic close orchestrator for `moai spec close` CLI.
//
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 design.md §B.1 + §A.2.
//
// M1 scope (this milestone): public API surface + precondition matrix validation
// + dry-run path + structured error reporting + lock acquisition. M3 milestone
// completes the atomic-commit transaction (file staging + git commit + SHA
// backfill mechanism per design §B.1 Option D).
//
// Public contract:
//
//	closer.Close(specID, opts) → (*CloseResult, error)
//
// Result semantics:
//   - error == nil + Result.CommitSHA non-empty → atomic close succeeded
//   - error == nil + Result.NoOp == true → backfill-only on already-completed SPEC
//   - error == ErrPreconditionMissing → preconditions not met; staging untouched
//   - error == ErrSpecCloseLockHeld → another process holds the lock
//   - error == ErrDryRun → --dry-run requested; no staging side-effects
package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CloseOptions configures Close() invocations.
type CloseOptions struct {
	// BaseDir is the project root (where .moai/specs/ lives). Default ".".
	BaseDir string

	// BackfillOnly transitions only the missing fields (sync_commit_sha /
	// mx_commit_sha / status) without requiring full preconditions. Per AC-LSG-022.
	// When the SPEC is already fully completed, this becomes a no-op (exit 0,
	// 0 commits) per AC-LSG-018 (v0.1.2 reframe).
	BackfillOnly bool

	// DryRun prints the diff that would be applied and exits without staging
	// or committing. Returns ErrDryRun.
	DryRun bool

	// Force bypasses precondition checks. Reserved for emergency recovery; not
	// part of normal close flow. Use only when L60 backfill rule applies.
	Force bool

	// LogPath overrides the audit-trail log destination. When empty, the log is
	// written to <BaseDir>/.moai/logs/lifecycle-close.log per NFR-LSG-004. Tests
	// inject a t.TempDir() path so they never touch the real project log.
	LogPath string
}

// CloseResult is the structured output of Close().
type CloseResult struct {
	SpecID    string            `json:"spec_id"`
	CommitSHA string            `json:"commit_sha,omitempty"`
	// Transitions records which fields were updated and to what values.
	// Empty map indicates no-op (AC-LSG-018 fully-completed-noop fixture).
	Transitions map[string]string `json:"transitions"`
	// NoOp is true when the SPEC is already completed and no changes are needed.
	NoOp bool `json:"noop"`
	// Mode is "full-close" or "backfill-only" per AC-LSG-020 log fields.
	Mode string `json:"mode"`
	// PreconditionsFailed lists the precondition names that prevented close.
	// Populated only when error == ErrPreconditionMissing.
	PreconditionsFailed []string `json:"preconditions_failed,omitempty"`
	// Result is the in-memory outcome: "success" | "failure" | "noop".
	// NOTE: the on-disk AC-LSG-020 audit log uses a narrower {success, failure}
	// enum — "noop" maps to "success" there via lifecycleCloseLogResult, because
	// a no-op close IS a successful close (see lifecycleCloseLogEntry).
	Result string `json:"result"`
	// DurationMs is total wall-clock time of the close invocation.
	DurationMs int64 `json:"duration_ms"`
	// AuditedAt is the close timestamp (RFC3339) per AC-LSG-020 log schema.
	AuditedAt time.Time `json:"audited_at"`
}

// Sentinel errors for the close path. The CLI layer maps these to exit codes:
//   - ErrPreconditionMissing → exit 1 (AC-LSG-006, AC-LSG-014)
//   - ErrSpecCloseLockHeld   → exit 1 (AC-LSG-010, AC-LSG-021)
//   - ErrDryRun              → exit 0 (informational, --dry-run requested)
var (
	// ErrPreconditionMissing is returned when one or more close preconditions
	// are not met (e.g., missing §E.5 Mx section in progress.md). The
	// PreconditionsFailed field on CloseResult names the specific failure.
	ErrPreconditionMissing = errors.New("close precondition not met")

	// ErrDryRun is returned when opts.DryRun is set. It signals "preview only".
	// Not a true error — CLI exits 0 and prints the would-apply diff.
	ErrDryRun = errors.New("dry-run requested; no staging performed")

	// ErrAlreadyCompleted indicates the SPEC is already at status: completed.
	// In backfill-only mode this is converted to a no-op success (NoOp=true).
	// In full close mode it is surfaced to the caller.
	ErrAlreadyCompleted = errors.New("SPEC already at status: completed")
)

// @MX:ANCHOR: [AUTO] Close는 `moai spec close`의 단일 진입점 — CLI(spec_close.go) + 단위/통합 테스트가 호출(fan_in=3)
// @MX:REASON: [AUTO] no-op 판정 불변식(--backfill-only 모드에서 spec.md status=="completed"면 mx_commit_sha 백필 형태와 무관하게 무조건 no-op, 0 commit)이 AC-LSG-018/022 진실표에 묶여 있다. 이 술어를 변경하면 5개 이미-종료 SPEC의 dogfood가 깨진다.
//
// Close orchestrates the atomic 4-phase close transition for a SPEC.
//
// M1 implementation (this milestone): precondition matrix + lock + dry-run +
// no-op detection. M3 will add the atomic git commit transaction.
//
// The lock is held only for the duration of this call (released on return).
//
// Every invocation (success / no-op / failure / dry-run) appends one JSON line
// to the audit-trail log per NFR-LSG-004 (AC-LSG-020). The log destination is
// opts.LogPath, defaulting to <BaseDir>/.moai/logs/lifecycle-close.log. The
// empty-specID guard is the sole exception: it predates result construction and
// has no SPEC to attribute the log entry to.
func Close(specID string, opts CloseOptions) (result *CloseResult, err error) {
	startedAt := time.Now()

	if specID == "" {
		return nil, fmt.Errorf("Close: empty specID")
	}

	baseDir := opts.BaseDir
	if baseDir == "" {
		baseDir = "."
	}

	mode := "full-close"
	if opts.BackfillOnly {
		mode = "backfill-only"
	}

	result = &CloseResult{
		SpecID:      specID,
		Transitions: map[string]string{},
		Mode:        mode,
		AuditedAt:   startedAt.UTC(),
	}

	// Single audit-trail emission per invocation (NFR-LSG-004 / AC-LSG-020).
	// Runs on every return path below; DurationMs is finalized here so all paths
	// record consistent wall-clock time without per-return duplication. The log
	// `result` field maps the in-memory "noop" value to "success" (no-op IS a
	// success per AC-LSG-020 — the log schema's result enum is {success, failure}
	// only). Log I/O failure is non-fatal: it never overrides the close outcome.
	defer func() {
		result.DurationMs = time.Since(startedAt).Milliseconds()
		writeLifecycleCloseLog(baseDir, opts.LogPath, result)
	}()

	// AC-LSG-010 — acquire per-SPEC lock; ErrSpecCloseLockHeld on contention.
	lock, lockErr := AcquireSpecCloseLock(baseDir, specID)
	if lockErr != nil {
		result.Result = "failure"
		if IsLockHeldError(lockErr) {
			return result, lockErr
		}
		return result, fmt.Errorf("acquire lock: %w", lockErr)
	}
	defer func() { _ = lock.Release() }()

	// Read spec.md + progress.md + acceptance.md
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if _, statErr := os.Stat(specDir); statErr != nil {
		result.Result = "failure"
		return result, fmt.Errorf("spec directory not found: %s", specDir)
	}

	state, loadErr := loadSpecCloseState(specDir, specID)
	if loadErr != nil {
		result.Result = "failure"
		return result, fmt.Errorf("load spec state: %w", loadErr)
	}

	// Detect no-op (AC-LSG-018 / AC-LSG-022 fully-completed fixture state).
	//
	// M1/M2 remediation (Defect 1): the no-op predicate keys ONLY on
	// `spec.md status == "completed"`. A terminal-state SPEC is already closed
	// regardless of how its §E.5 mx_commit_sha was backfilled (the 5 already-
	// discharged target SPECs left mx_commit_sha empty / `null` / `(this commit)`
	// placeholder / absent). The earlier triple-AND gate additionally required
	// non-empty SyncCommitSHA AND MxCommitSHA, which let only the literal both-
	// SHA-present fixture through and caused 4/5 production SPECs to fall through
	// to the precondition matrix or compute-transitions — violating AC-LSG-018's
	// 0-commit no-op requirement.
	//
	// Truth-table safety (AC-LSG-022): the three transition fixtures
	// (Y_N_N_Y / Y_Y_N_Y / Y_Y_Y_Y_StatusDrift) all carry status: implemented,
	// so they never enter this branch — only `fully-completed-noop` (status:
	// completed) and the 5 production SPECs do.
	if state.SpecMDStatus == "completed" {
		if opts.BackfillOnly {
			// Already complete + backfill-only = no-op success path.
			result.NoOp = true
			result.Result = "noop"
			return result, nil
		}
		result.Result = "failure"
		return result, ErrAlreadyCompleted
	}

	// Precondition matrix validation per AC-LSG-006 + AC-LSG-014.
	preconditionsFailed := validatePreconditions(state, opts)
	if len(preconditionsFailed) > 0 && !opts.Force {
		result.PreconditionsFailed = preconditionsFailed
		result.Result = "failure"
		return result, fmt.Errorf("%w: %s", ErrPreconditionMissing, strings.Join(preconditionsFailed, ", "))
	}

	// Compute transitions (which fields would change)
	transitions := computeTransitions(state, opts)
	result.Transitions = transitions

	// Dry-run path — return ErrDryRun without staging
	if opts.DryRun {
		result.Result = "success"
		return result, ErrDryRun
	}

	// M1 stub: actual atomic commit transaction is M3 deliverable.
	// For now, M1 returns the computed transitions without performing the commit.
	// This allows AC-LSG-014 (precondition abort atomicity) to be verified — no
	// staging is performed because the commit phase is not yet implemented.
	//
	// M3 will replace this stub with: write changes to disk, git add, git commit,
	// then populate result.CommitSHA.
	result.Result = "success"
	return result, nil
}

// lifecycleCloseLogEntry is the on-disk NFR-LSG-004 audit-trail schema. One
// such entry is appended (as a single JSON line) per Close() invocation.
//
// Distinct from CloseResult: the log `result` enum is {success, failure} ONLY
// (AC-LSG-020). A no-op close — whose in-memory CloseResult.Result is "noop" —
// serializes here as result: "success" with transitions: {} (empty object).
// This reconciliation lets BOTH AC-LSG-018's jq filter
// (`.result == "success" and .transitions == {}`) AND AC-LSG-020's
// `.result == "success"` filter match the 5 no-op dogfood closes.
type lifecycleCloseLogEntry struct {
	Timestamp   string            `json:"timestamp"`   // RFC3339
	SpecID      string            `json:"spec_id"`
	Mode        string            `json:"mode"`        // full-close | backfill-only
	Transitions map[string]string `json:"transitions"` // changed fields; {} when none
	CommitSHA   string            `json:"commit_sha"`
	Result      string            `json:"result"`      // success | failure
	DurationMs  int64             `json:"duration_ms"`
}

// lifecycleCloseLogResult maps the in-memory CloseResult.Result to the log
// schema's {success, failure} enum. "noop" maps to "success" (a no-op IS a
// successful close); any non-"failure" value is treated as success.
func lifecycleCloseLogResult(in string) string {
	if in == "failure" {
		return "failure"
	}
	return "success"
}

// writeLifecycleCloseLog appends one JSON line describing the close outcome to
// the audit-trail log. logPath overrides the destination; when empty it
// defaults to <baseDir>/.moai/logs/lifecycle-close.log per NFR-LSG-004. The
// parent directory is created if absent. All I/O errors are swallowed — the
// audit trail is best-effort and MUST NOT alter the close outcome.
func writeLifecycleCloseLog(baseDir, logPath string, result *CloseResult) {
	if result == nil {
		return
	}

	path := logPath
	if path == "" {
		path = filepath.Join(baseDir, ".moai", "logs", "lifecycle-close.log")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}

	transitions := result.Transitions
	if transitions == nil {
		transitions = map[string]string{}
	}

	entry := lifecycleCloseLogEntry{
		Timestamp:   result.AuditedAt.UTC().Format(time.RFC3339),
		SpecID:      result.SpecID,
		Mode:        result.Mode,
		Transitions: transitions,
		CommitSHA:   result.CommitSHA,
		Result:      lifecycleCloseLogResult(result.Result),
		DurationMs:  result.DurationMs,
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.Write(append(line, '\n'))
}

// closeState captures the parsed state of spec.md + progress.md needed for
// precondition validation.
type closeState struct {
	SpecMDStatus      string // current spec.md frontmatter status
	HasSyncSection    bool   // §E.2 section present in progress.md
	HasMxSection      bool   // §E.5 section present in progress.md
	SyncCommitSHA     string // extracted §E.2 sync_commit_sha
	MxCommitSHA       string // extracted §E.5 mx_commit_sha
	ProgressMDStatus  string // progress.md §E.3 status field
	ACAllPass         bool   // all MUST-PASS acceptance criteria PASS
	HasPassWithDebt   bool   // any AC marked PASS-WITH-DEBT
}

// loadSpecCloseState reads spec.md + progress.md + acceptance.md and populates
// closeState. Missing optional files are tolerated (corresponding bool fields
// set to false).
func loadSpecCloseState(specDir, specID string) (*closeState, error) {
	state := &closeState{}

	// spec.md
	specMDPath := filepath.Join(specDir, "spec.md")
	specContent, err := os.ReadFile(specMDPath)
	if err != nil {
		return nil, fmt.Errorf("read spec.md: %w", err)
	}
	if m := specStatusPattern.FindStringSubmatch(string(specContent)); len(m) > 1 {
		state.SpecMDStatus = strings.TrimSpace(m[1])
	}

	// progress.md (optional — V2.x SPECs lack this)
	progressMDPath := filepath.Join(specDir, "progress.md")
	if progressContent, perr := os.ReadFile(progressMDPath); perr == nil {
		body := string(progressContent)
		state.HasSyncSection = hasProgressMarker(body, "§E.2")
		state.HasMxSection = hasProgressMarker(body, "§E.5")
		state.SyncCommitSHA = extractProgressField(body, "sync_commit_sha")
		state.MxCommitSHA = extractProgressField(body, "mx_commit_sha")
		// progress.md §E.3 status is read as `status:` line within §E.3 block.
		// For M1 we accept that if a top-level `status:` line exists, that is the value.
		// Refinement: M3 may need a per-section parser.
		state.ProgressMDStatus = extractProgressField(body, "status")
	}

	// acceptance.md AC PASS check (M1: simple grep-based)
	acMDPath := filepath.Join(specDir, "acceptance.md")
	if acContent, aerr := os.ReadFile(acMDPath); aerr == nil {
		acBody := string(acContent)
		// Look for PASS-WITH-DEBT marker (case-insensitive)
		state.HasPassWithDebt = strings.Contains(strings.ToUpper(acBody), "PASS-WITH-DEBT")
		// M1: assume all PASS unless we can detect failures; M3 will parse the AC table.
		// For now we treat presence of FAIL markers as failure.
		hasFail := strings.Contains(strings.ToUpper(acBody), "**FAIL**") ||
			strings.Contains(strings.ToUpper(acBody), "| FAIL |") ||
			strings.Contains(strings.ToUpper(acBody), "| FAILED |")
		state.ACAllPass = !hasFail
	} else {
		// Missing acceptance.md is acceptable for Tier S SPECs (LEAN workflow).
		state.ACAllPass = true
	}

	return state, nil
}

// validatePreconditions checks the 4-phase precondition matrix per AC-LSG-006.
// Returns a slice of failing precondition names (empty slice = all pass).
//
// Backfill-only mode relaxes the spec.md status requirement (allows implemented
// or in-progress to backfill); other preconditions still apply.
func validatePreconditions(state *closeState, opts CloseOptions) []string {
	var failed []string

	// Precondition 1: §E.2 sync section present
	if !state.HasSyncSection {
		failed = append(failed, "missing §E.2 sync-phase audit-ready signal in progress.md")
	}

	// Precondition 2: §E.5 mx section present
	if !state.HasMxSection {
		failed = append(failed, "missing §E.5 Mx-phase audit-ready signal in progress.md")
	}

	// Precondition 3: all MUST-PASS AC PASS
	if !state.ACAllPass {
		failed = append(failed, "one or more acceptance criteria are not PASS")
	}

	// Precondition 4: no PASS-WITH-DEBT
	if state.HasPassWithDebt && !opts.Force {
		failed = append(failed, "PASS-WITH-DEBT marker present; use --force to override")
	}

	// In backfill-only mode the spec.md status requirement is relaxed
	// (the whole point of backfill is to bring an `implemented`/`in-progress`
	// SPEC up to `completed`). In full-close mode the spec.md status must be
	// `implemented` to proceed.
	if !opts.BackfillOnly {
		if state.SpecMDStatus != "implemented" && state.SpecMDStatus != "completed" {
			failed = append(failed,
				fmt.Sprintf("spec.md status is %q; full close requires status=implemented "+
					"(use --backfill-only for in-progress SPECs)", state.SpecMDStatus))
		}
	}

	return failed
}

// computeTransitions returns the field changes that the atomic close would apply.
// The returned map describes intent, not actual applied changes (M3 applies them).
func computeTransitions(state *closeState, opts CloseOptions) map[string]string {
	transitions := map[string]string{}

	if state.SpecMDStatus != "completed" {
		transitions["spec.md:frontmatter.status"] = "completed"
	}
	if state.ProgressMDStatus != "completed" {
		transitions["progress.md:§E.3.status"] = "completed"
	}
	if state.SyncCommitSHA == "" {
		transitions["progress.md:§E.2.sync_commit_sha"] = "<derived-from-recent-sync-commit>"
	}
	if state.MxCommitSHA == "" {
		transitions["progress.md:§E.5.mx_commit_sha"] = "<derived-from-recent-mx-commit>"
	}

	return transitions
}
