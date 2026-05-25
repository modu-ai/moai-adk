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
	// Result is "success" | "failure" | "noop" per AC-LSG-020 log schema.
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

// Close orchestrates the atomic 4-phase close transition for a SPEC.
//
// M1 implementation (this milestone): precondition matrix + lock + dry-run +
// no-op detection. M3 will add the atomic git commit transaction.
//
// The lock is held only for the duration of this call (released on return).
func Close(specID string, opts CloseOptions) (*CloseResult, error) {
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

	result := &CloseResult{
		SpecID:      specID,
		Transitions: map[string]string{},
		Mode:        mode,
		AuditedAt:   startedAt.UTC(),
	}

	// AC-LSG-010 — acquire per-SPEC lock; ErrSpecCloseLockHeld on contention.
	lock, err := AcquireSpecCloseLock(baseDir, specID)
	if err != nil {
		if IsLockHeldError(err) {
			result.Result = "failure"
			result.DurationMs = time.Since(startedAt).Milliseconds()
			return result, err
		}
		result.Result = "failure"
		result.DurationMs = time.Since(startedAt).Milliseconds()
		return result, fmt.Errorf("acquire lock: %w", err)
	}
	defer func() { _ = lock.Release() }()

	// Read spec.md + progress.md + acceptance.md
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if _, err := os.Stat(specDir); err != nil {
		result.Result = "failure"
		result.DurationMs = time.Since(startedAt).Milliseconds()
		return result, fmt.Errorf("spec directory not found: %s", specDir)
	}

	state, err := loadSpecCloseState(specDir, specID)
	if err != nil {
		result.Result = "failure"
		result.DurationMs = time.Since(startedAt).Milliseconds()
		return result, fmt.Errorf("load spec state: %w", err)
	}

	// Detect no-op (AC-LSG-018 fully-completed fixture)
	if state.SpecMDStatus == "completed" && state.SyncCommitSHA != "" && state.MxCommitSHA != "" {
		if opts.BackfillOnly {
			// Already complete + backfill-only = no-op success path
			result.NoOp = true
			result.Result = "noop"
			result.DurationMs = time.Since(startedAt).Milliseconds()
			return result, nil
		}
		result.Result = "failure"
		result.DurationMs = time.Since(startedAt).Milliseconds()
		return result, ErrAlreadyCompleted
	}

	// Precondition matrix validation per AC-LSG-006 + AC-LSG-014.
	preconditionsFailed := validatePreconditions(state, opts)
	if len(preconditionsFailed) > 0 && !opts.Force {
		result.PreconditionsFailed = preconditionsFailed
		result.Result = "failure"
		result.DurationMs = time.Since(startedAt).Milliseconds()
		return result, fmt.Errorf("%w: %s", ErrPreconditionMissing, strings.Join(preconditionsFailed, ", "))
	}

	// Compute transitions (which fields would change)
	transitions := computeTransitions(state, opts)
	result.Transitions = transitions

	// Dry-run path — return ErrDryRun without staging
	if opts.DryRun {
		result.Result = "success"
		result.DurationMs = time.Since(startedAt).Milliseconds()
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
	result.DurationMs = time.Since(startedAt).Milliseconds()
	return result, nil
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
