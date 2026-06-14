// Package harness — frontmatter modification applier.
// REQ-HL-003: description enrichment (Tier 2 heuristic).
// REQ-HL-004: trigger injection (Tier 3 rule, feature-gated).
// REQ-HL-005: Apply() — create snapshot first then modify files (Phase 4).
package harness

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// enableTriggerInjectionWrites is a feature flag that enables actual file writes in InjectTrigger.
// Phase 2 defaults to OFF — only verifies dedup logic without performing actual writes.
//
// @MX:TODO: [AUTO] Phase 4: wire learning.auto_apply config to enable writes
// @MX:SPEC: SPEC-V3R3-HARNESS-LEARNING-001 REQ-HL-004 (T-P2-05)
var enableTriggerInjectionWrites = false

// Applier is a component that modifies SKILL.md file frontmatter.
// All modifications target only description or triggers fields,
// other frontmatter fields and body are preserved byte-identically.
//
// @MX:ANCHOR: [AUTO] EnrichDescription, InjectTrigger are learning pipeline write paths.
// @MX:REASON: [AUTO] fan_in >= 3: applier_test.go, safety.go(Phase 3), CLI apply(Phase 4)
type Applier struct {
	// allowWrites is an instance-level flag that allows actual file writes in InjectTrigger.
	// Default value is enableTriggerInjectionWrites (package-level flag).
	// Can be set to true in tests via newApplierWithWritesEnabled().
	allowWrites bool

	// manifestPath is the M6 auditable lineage manifest path (REQ-HLC-001/002).
	// When non-empty, Apply() appends a LineageEntry on the approved/rejected transition.
	// When empty (the NewApplier() default), Apply() skips the lineage write — preserving
	// the pre-lineage Apply behavior for callers that do not opt into lineage logging.
	// The path is injectable so tests can point it at t.TempDir(); the production caller
	// passes <learning-history-dir>/manifest.jsonl.
	manifestPath string

	// measurer + baselineStore wire the M2-lite in-Apply non-regression gate
	// (SPEC-HARNESS-REGRESSION-GATE-001). The gate is ACTIVE only when BOTH are
	// non-nil. When either is nil (the NewApplier() / newApplierWithManifest()
	// defaults), Apply() skips the gate entirely — preserving the pre-gate Apply
	// behavior for callers that do not opt into the regression gate. Both are
	// injectable so tests inject a stub measurer + a t.TempDir() baseline path.
	//
	// HONEST FRAMING: for the current markdown-only write surface the measured
	// delta is typically Δ=0 (the gate is always-pass); the gate's value is a
	// measurement scaffold + a dormant defense-in-depth net (see regression_gate.go).
	measurer      Measurer
	baselineStore *BaselineStore

	// outcomeObserver wires the additive Apply-OUTCOME capture
	// (SPEC-HARNESS-OUTCOME-CAPTURE-001). When non-nil AND the regression gate is
	// active, applyWithRegressionGate emits one apply_outcome observer record at
	// each terminal branch AFTER the gate has decided (DD-2). When nil,
	// recordOutcome is a safe no-op — preserving callers that do not opt into
	// outcome capture (mirrors the manifestPath == "" → skip pattern). The capture
	// is purely additive; it never alters the gate's keep/rollback decision, the
	// returned error, the baseline-store update, or the lineage write (C10).
	outcomeObserver *Observer
}

// NewApplier creates a default Applier.
// Actual file writes in InjectTrigger follow package-level flag (enableTriggerInjectionWrites).
func NewApplier() *Applier {
	return &Applier{allowWrites: enableTriggerInjectionWrites}
}

// NewApplierWithRegressionGate creates an Applier with the M2-lite in-Apply
// non-regression gate wired to production primitives (SPEC-HARNESS-REGRESSION-GATE-001):
// the real goMeasurer (runs `go test` / `go vet` via internal/measure) and a
// BaselineStore at baselinePath, plus the M6 lineage manifest at manifestPath.
//
// This is the production seam the harness apply CLI opts into when it wants the
// gate active. It is exported (not test-only) because the gate's primary present-day
// value is being an infrastructure seam (§A.2 honest framing). When the harness
// write surface is markdown-only the measured delta is typically Δ=0 (always-pass),
// so wiring the gate is a no-op safety net under the current narrow FROZEN allowlist.
func NewApplierWithRegressionGate(manifestPath, baselinePath string) *Applier {
	return &Applier{
		allowWrites:   enableTriggerInjectionWrites,
		manifestPath:  manifestPath,
		measurer:      newGoMeasurer(),
		baselineStore: NewBaselineStore(baselinePath),
	}
}

// WithOutcomeObserver wires an outcome observer so the gate-active Apply path
// emits an apply_outcome record at each terminal branch (SPEC-HARNESS-OUTCOME-CAPTURE-001).
// It returns the receiver for fluent production wiring:
//
//	a := NewApplierWithRegressionGate(manifestPath, baselinePath).
//	        WithOutcomeObserver(NewObserver(usageLogPath))
//
// A nil observer leaves the capture disabled (recordOutcome is a no-op). The
// capture is additive and never alters the gate's keep/rollback decision (C10).
func (a *Applier) WithOutcomeObserver(obs *Observer) *Applier {
	a.outcomeObserver = obs
	return a
}

// newApplierWithWritesEnabled creates an Applier with InjectTrigger actual writes enabled.
// Test-only function.
func newApplierWithWritesEnabled() *Applier {
	return &Applier{allowWrites: true}
}

// newApplierWithManifest creates an Applier wired to write M6 lineage entries to manifestPath.
// Test-only constructor: lets lineage tests inject a t.TempDir() manifest path.
func newApplierWithManifest(manifestPath string) *Applier {
	return &Applier{allowWrites: enableTriggerInjectionWrites, manifestPath: manifestPath}
}

// EnrichDescription adds heuristicNote to SKILL.md description field.
// REQ-HL-003: modifies only description field, preserves other frontmatter and body.
// Handles idempotently if same note already exists (no duplicate addition).
//
// heuristicNote is added to description in "# heuristic: <note>" format.
func (a *Applier) EnrichDescription(skillPath, heuristicNote string) error {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("applier: 파일 읽기 실패 %s: %w", skillPath, err)
	}

	fm, body, err := splitFrontmatterBody(string(content))
	if err != nil {
		return fmt.Errorf("applier: frontmatter 파싱 실패 %s: %w", skillPath, err)
	}

	// Find and modify description field
	newFM, changed := enrichDescriptionInFrontmatter(fm, heuristicNote)
	if !changed {
		// No change (already includes that note) — idempotent
		return nil
	}

	// Rejoin
	newContent := "---\n" + newFM + "---\n" + body

	if err := os.WriteFile(skillPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("applier: 파일 쓰기 실패 %s: %w", skillPath, err)
	}
	return nil
}

// InjectTrigger adds keyword to SKILL.md triggers list.
// REQ-HL-004: does not add duplicate keywords (dedup).
//
// @MX:WARN: [AUTO] If enableTriggerInjectionWrites is OFF (Phase 2), skip actual file write.
// @MX:REASON: [AUTO] Before Phase 4, verify only dedup logic without file changes.
func (a *Applier) InjectTrigger(skillPath, keyword string) error {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("applier: 파일 읽기 실패 %s: %w", skillPath, err)
	}

	fm, body, err := splitFrontmatterBody(string(content))
	if err != nil {
		return fmt.Errorf("applier: frontmatter 파싱 실패 %s: %w", skillPath, err)
	}

	// dedup: check if keyword already exists
	newFM, changed := injectTriggerInFrontmatter(fm, keyword)
	if !changed {
		// Already exists or no change
		return nil
	}

	// Check feature flag — if OFF, skip actual write (Phase 2 gate)
	if !a.allowWrites {
		return nil
	}

	// Actual file write (enabled via config in Phase 4)
	newContent := "---\n" + newFM + "---\n" + body
	if err := os.WriteFile(skillPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("applier: 파일 쓰기 실패 %s: %w", skillPath, err)
	}
	return nil
}

// ─────────────────────────────────────────────
// Phase 4: Apply() — snapshot + safety pipeline integration
// ─────────────────────────────────────────────

// SafetyEvaluator is the Evaluate method interface of safety pipeline.
// Prevent circular import: cannot directly import harness → safety.
// safety.Pipeline implements this interface.
type SafetyEvaluator interface {
	Evaluate(proposal Proposal, sessions []Session) (Decision, error)
}

// ApplyPendingError is an error that occurs when safety pipeline returns pending_approval.
// orchestrator (moai-harness-learner skill) receives this error and presents
// OversightProposal to users via AskUserQuestion.
//
// @MX:ANCHOR: [AUTO] ApplyPendingError is the subagent→orchestrator boundary type.
// @MX:REASON: [AUTO] fan_in >= 3: applier.go, applier_test.go, harness CLI apply, moai-harness-learner skill
type ApplyPendingError struct {
	// OversightPayload is the payload orchestrator uses for AskUserQuestion.
	OversightPayload *OversightProposal
}

func (e *ApplyPendingError) Error() string {
	if e.OversightPayload != nil {
		return fmt.Sprintf("apply: awaiting user approval (proposal_id=%s)", e.OversightPayload.ProposalID)
	}
	return "apply: awaiting user approval"
}

// snapshotManifest is the manifest.json schema of snapshot directory.
type snapshotManifest struct {
	// ProposalID is the proposal ID that created this snapshot.
	ProposalID string `json:"proposal_id"`

	// CreatedAt is the snapshot creation time (UTC).
	CreatedAt time.Time `json:"created_at"`

	// Files is the list of backed up files.
	Files []snapshotFile `json:"files"`
}

// snapshotFile is the single backup file information.
type snapshotFile struct {
	// OriginalPath is the original file path.
	OriginalPath string `json:"original_path"`

	// BackupName is the backup filename within snapshot directory.
	BackupName string `json:"backup_name"`
}

// Apply safely applies Proposal after safety pipeline evaluation.
// [HARD] Must call evaluator.Evaluate() first, return immediately if rejected.
// [HARD] Snapshot must be created before file write. Abort write on snapshot failure.
//
// evaluator is the SafetyEvaluator interface (implemented by safety.Pipeline).
// snapshotBase is the base path in ".moai/harness/learning-history/snapshots/" format.
// sessions is the list of recent sessions used for L2 canary check.
//
// @MX:ANCHOR: [AUTO] Apply is the single entry point of Phase 4 learning application pipeline.
// @MX:REASON: [AUTO] fan_in >= 3: applier_test.go, harness CLI apply, moai-harness-learner skill
func (a *Applier) Apply(proposal Proposal, evaluator SafetyEvaluator, snapshotBase string, sessions []Session) error {
	// ── Step 1: Safety Pipeline Evaluation ─────────────────────────────────────────
	// [HARD] Must pass all 5-Layers including Frozen Guard.
	decision, err := evaluator.Evaluate(proposal, sessions)
	if err != nil {
		return fmt.Errorf("applier: safety pipeline evaluation error: %w", err)
	}

	switch decision.Kind {
	case DecisionRejected:
		// [HARD] M6 auditable lineage (REQ-HLC-004/008): record the rejected transition
		// BEFORE returning, capturing the rejecting layer's reason. The active harness is
		// left unchanged (no snapshot, no SKILL.md write). The lineage write is an
		// after-effect record; a lineage write error is wrapped but the rejection still
		// stands (the rejection is the primary effect).
		if lerr := a.writeLineage(proposal, "rejected", "", rejectReason(decision)); lerr != nil {
			return fmt.Errorf("applier: proposal rejected (L%d, rejected); lineage write failed: %w", decision.RejectedBy, lerr)
		}
		return fmt.Errorf("applier: proposal rejected (L%d, rejected)", decision.RejectedBy)

	case DecisionPendingApproval:
		// [HARD] subagent must not call AskUserQuestion directly.
		// Return payload to orchestrator to delegate user approval.
		// Pending is NOT a transition — no lineage entry is written (REQ-HLC-005).
		return &ApplyPendingError{OversightPayload: decision.OversightProposal}

	case DecisionApproved:
		// approved — continue
	}

	// ── Step 2: Create Snapshot (before write) ───────────────────────────────
	// [HARD] Abort write on snapshot failure. snapshotDir is captured so the
	// in-Apply regression gate (DD-2) can roll back via RestoreSnapshot.
	snapshotDir, err := a.createSnapshot(proposal, snapshotBase)
	if err != nil {
		return fmt.Errorf("applier: snapshot creation failed — abort write: %w", err)
	}

	// ── M2-lite in-Apply non-regression gate (SPEC-HARNESS-REGRESSION-GATE-001) ──
	// When the gate is active (measurer + baselineStore both injected) it wraps the
	// file modification with a measure→apply→measure→compare→keep-or-rollback flow
	// (DD-2). When inactive (NewApplier default), the straight-line modify+lineage
	// path below runs unchanged, preserving pre-gate Apply behavior.
	if a.gateActive() {
		return a.applyWithRegressionGate(proposal, snapshotDir)
	}

	// ── Step 3: Actual File Modification ───────────────────────────────────────────────
	if err := a.applyFileModification(proposal); err != nil {
		return err
	}

	// ── Step 4: M6 auditable lineage (REQ-HLC-003): record the approved transition AFTER
	// the SKILL.md modification succeeds. applied_surface is the modified frontmatter field
	// key. A lineage write error is surfaced to the caller, but the file write (the primary
	// effect) has already happened.
	if lerr := a.writeLineage(proposal, "approved", proposal.FieldKey, "approved transition: "+proposal.FieldKey+" enriched"); lerr != nil {
		return fmt.Errorf("applier: file modified but lineage write failed: %w", lerr)
	}
	return nil
}

// writeLineage appends one M6 LineageEntry to the Applier's manifest path.
// When manifestPath is empty (the NewApplier() default), the lineage write is skipped
// (no-op) so callers that did not opt into lineage logging keep the pre-lineage behavior.
func (a *Applier) writeLineage(proposal Proposal, decision, appliedSurface, reason string) error {
	if a.manifestPath == "" {
		return nil
	}
	return WriteLineageEntry(a.manifestPath, LineageEntry{
		ProposalID:     proposal.ID,
		TargetPath:     proposal.TargetPath,
		AppliedSurface: appliedSurface,
		Decision:       decision,
		Reason:         reason,
	})
}

// recordOutcome emits one additive apply_outcome observer record composed from the
// values the regression gate already produced (verdict + baseline/candidate triples
// + regressed dimensions + proposal_id + decision). When no outcome observer is
// wired (the gate-active-but-no-observer case, or the gate-inactive default), it is
// a safe no-op — mirroring the manifestPath == "" → skip pattern for lineage
// (SPEC-HARNESS-OUTCOME-CAPTURE-001 REQ-OC-008, DD-4).
//
// The emit is purely additive: the caller invokes it AFTER the gate has decided, so
// the keep/rollback decision and the returned error are already fixed. An emit error
// is surfaced to the caller (REQ-OC-007) but the caller wraps it consistently with
// writeLineage's "primary effect already happened" semantics — it never flips the
// verdict.
func (a *Applier) recordOutcome(verdict, decision string, baseline, candidate MetricTriple, regressed []string, proposalID string) error {
	if a.outcomeObserver == nil {
		return nil
	}
	return a.outcomeObserver.RecordOutcome(OutcomeRecord{
		Verdict:    verdict,
		Decision:   decision,
		ProposalID: proposalID,
		Baseline:   baseline,
		Candidate:  candidate,
		Regressed:  regressed,
	})
}

// applyFileModification performs the actual SKILL.md frontmatter modification for
// an approved proposal (description enrichment or trigger injection). Extracted from
// Apply so the in-Apply regression gate can reuse the exact same write path.
func (a *Applier) applyFileModification(proposal Proposal) error {
	switch proposal.FieldKey {
	case "description":
		if err := a.EnrichDescription(proposal.TargetPath, proposal.NewValue); err != nil {
			return err
		}
	case "triggers":
		// Perform InjectTrigger with write-enabled Applier
		w := newApplierWithWritesEnabled()
		if err := w.InjectTrigger(proposal.TargetPath, proposal.NewValue); err != nil {
			return err
		}
	default:
		return fmt.Errorf("applier: unsupported fieldKey %q", proposal.FieldKey)
	}
	return nil
}

// gateActive reports whether the M2-lite in-Apply non-regression gate is wired.
// The gate runs only when BOTH the measurer and the baseline store are injected.
func (a *Applier) gateActive() bool {
	return a.measurer != nil && a.baselineStore != nil
}

// applyWithRegressionGate runs the M2-lite non-regression gate around the file
// modification (SPEC-HARNESS-REGRESSION-GATE-001, DD-2 ordering). The snapshot has
// already been created (snapshotDir captured by the caller). Order:
//
//	(1) measure baseline   — BEFORE apply (reflects pre-apply project state)
//	(2) apply modification  — the SKILL.md frontmatter write
//	(3) measure candidate   — AFTER apply
//	(4) compare deltas
//	(5) Δ ≥ 0 for all  → keep + update baseline + M6 "approved" lineage
//	    any regression → RestoreSnapshot rollback + "regression-blocked" lineage
//	                     + return ApplyRegressionError
//
// Fail-closed (REQ-RG-014): when either measurement cannot execute (build error,
// exec failure, timeout), the change is NOT kept — the snapshot is restored (if
// already applied) and a wrapped measurement error is returned.
//
// HONEST FRAMING: for the current markdown-only write surface baseline == candidate
// (Δ=0), so this path keeps the change every time. The genuine value is the
// measurement scaffold + the dormant defense-in-depth rollback (see §A.2 / DD-7).
func (a *Applier) applyWithRegressionGate(proposal Proposal, snapshotDir string) error {
	projectRoot := measurementRoot(snapshotDir)

	// firstRun is signalled by the ABSENCE of the baseline store file (REQ-RG-005):
	// on the very first gated Apply there is no prior baseline to regress against, so
	// the candidate is adopted as the new baseline and the Apply is NOT blocked.
	// The store-presence gate is read BEFORE the apply so a measurement-exec failure
	// during baseline measurement still fails closed (the file modification has not
	// happened yet, so no rollback is needed).
	_, hasPriorBaseline, lerr := a.baselineStore.Load()
	if lerr != nil {
		return fmt.Errorf("applier: regression gate baseline store read failed (fail-closed): %w", lerr)
	}

	// (1) baseline measure — BEFORE snapshot/apply (reflects pre-apply project state).
	// Fail-closed if it cannot execute (build error / exec failure / timeout).
	baseline, err := a.measurer.Measure(projectRoot)
	if err != nil {
		return fmt.Errorf("applier: regression gate baseline measurement failed (fail-closed): %w", err)
	}

	// (2)/(3) apply the file modification (snapshot already taken by the caller).
	if err := a.applyFileModification(proposal); err != nil {
		return err
	}

	// (4) candidate measure — AFTER apply. Fail-closed: roll back if it cannot execute.
	candidate, err := a.measurer.Measure(projectRoot)
	if err != nil {
		if rerr := RestoreSnapshot(snapshotDir); rerr != nil {
			return fmt.Errorf("applier: regression gate candidate measurement failed (fail-closed); rollback also failed: %w (rollback: %v)", err, rerr)
		}
		return fmt.Errorf("applier: regression gate candidate measurement failed (fail-closed), change rolled back: %w", err)
	}

	// (5) compare. First run (no prior baseline file) adopts the candidate without
	// blocking (REQ-RG-005); subsequent runs block on any regressed dimension.
	if hasPriorBaseline {
		if regressed := baseline.Regressions(candidate); len(regressed) > 0 {
			// (6a) Regression: roll back, audit, and return the typed error.
			summary := regressionSummary(baseline, candidate, regressed)
			if rerr := RestoreSnapshot(snapshotDir); rerr != nil {
				return fmt.Errorf("applier: non-regression gate blocked but rollback failed: %w", rerr)
			}
			if wlerr := a.writeLineage(proposal, "regression-blocked", "", summary); wlerr != nil {
				// Mirror existing writeLineage error semantics: the rollback + block are
				// the primary effects; a failed audit append is wrapped, not suppressed.
				return fmt.Errorf("applier: non-regression gate blocked (rolled back); lineage write failed: %w", wlerr)
			}
			// (NEW — additive, after the decision) emit the rolled-back outcome record.
			// The rollback + block decision is already fixed above; an emit error is
			// wrapped like the lineage-write error and never flips the verdict (REQ-OC-007).
			// On an emit error the typed *ApplyRegressionError signal is PRESERVED via
			// errors.Join (REQ-ERRJOIN-001/002): the joined value's Unwrap() []error is
			// walked by errors.As so the regression signal survives, while the
			// outcome-record error remains reachable via errors.Is/unwrap.
			regErr := &ApplyRegressionError{Baseline: baseline, Candidate: candidate, Regressed: regressed}
			if oerr := a.recordOutcome("rolled-back", "regression-blocked", baseline, candidate, regressed, proposal.ID); oerr != nil {
				return errors.Join(regErr, oerr)
			}
			return regErr
		}
	}

	// (6b) Non-regressing (or first run): keep, update baseline, write M6 "approved" lineage.
	if serr := a.baselineStore.Save(candidate); serr != nil {
		return fmt.Errorf("applier: file modified but baseline store update failed: %w", serr)
	}
	if wlerr := a.writeLineage(proposal, "approved", proposal.FieldKey, "approved transition: "+proposal.FieldKey+" enriched"); wlerr != nil {
		return fmt.Errorf("applier: file modified but lineage write failed: %w", wlerr)
	}
	// (NEW — additive, after the decision) emit the kept outcome record. The keep
	// decision (file modified + baseline saved + lineage written) is already fixed
	// above; an emit error is wrapped like the lineage-write error and never undoes
	// the kept change nor flips the verdict (REQ-OC-007). regressed is nil on a kept
	// outcome.
	if oerr := a.recordOutcome("kept", "approved", baseline, candidate, nil, proposal.ID); oerr != nil {
		return fmt.Errorf("applier: file modified but outcome record failed: %w", oerr)
	}
	return nil
}

// measurementRoot derives the project root for measurement from the snapshot dir.
// snapshotDir is <snapshotBase>/<ISO-DATE>/; the regression coverage profile is
// written under <root>/.moai/harness/. The production caller passes a snapshotBase
// inside the project tree, so the project root is the snapshotBase's grandparent.
// Tests inject a stub measurer that ignores this argument.
func measurementRoot(snapshotDir string) string {
	// <root>/.moai/harness/learning-history/snapshots/<ISO-DATE>
	//   → walk up to <root>. We only need a writable dir for the cover profile;
	//     for the stub measurer the value is unused, so the snapshot's base dir
	//     (its parent) is a safe, always-present default.
	return filepath.Dir(snapshotDir)
}

// regressionSummary formats a human-readable regressed-dimension summary for the
// "regression-blocked" lineage Reason field (REQ-RG-010).
func regressionSummary(baseline, candidate MetricTriple, regressed []string) string {
	return fmt.Sprintf(
		"regression-blocked (regressed: %v); baseline{tests=%d cov=%.2f lint=%d} candidate{tests=%d cov=%.2f lint=%d}",
		regressed,
		baseline.TestsPassed, baseline.Coverage, baseline.LintCount,
		candidate.TestsPassed, candidate.Coverage, candidate.LintCount,
	)
}

// rejectReason derives a non-empty reason string from a rejection Decision.
// It prefers the layer-supplied Reason; if absent, it synthesizes one from RejectedBy.
func rejectReason(decision Decision) string {
	if decision.Reason != "" {
		return decision.Reason
	}
	return fmt.Sprintf("rejected by safety layer L%d", decision.RejectedBy)
}

// createSnapshot backs up current content of proposal.TargetPath to snapshotBase/<ISO-DATE>/.
// Creates manifest.json then performs file copy. Returns the dated snapshotDir so the
// in-Apply regression gate can pass it to RestoreSnapshot for rollback (DD-2). An error
// returns an empty snapshotDir.
func (a *Applier) createSnapshot(proposal Proposal, snapshotBase string) (string, error) {
	// Generate ISO-DATE format directory name (date + nano collision prevention)
	now := time.Now().UTC()
	dirName := now.Format("2006-01-02T15-04-05.000000000Z")
	snapshotDir := filepath.Join(snapshotBase, dirName)

	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return "", fmt.Errorf("createSnapshot: directory creation failed %s: %w", snapshotDir, err)
	}

	// Read original file
	originalData, err := os.ReadFile(proposal.TargetPath)
	if err != nil {
		return "", fmt.Errorf("createSnapshot: original file read failed %s: %w", proposal.TargetPath, err)
	}

	// Backup filename: use original filename as-is
	backupName := filepath.Base(proposal.TargetPath)
	backupPath := filepath.Join(snapshotDir, backupName)

	if err := os.WriteFile(backupPath, originalData, 0o644); err != nil {
		return "", fmt.Errorf("createSnapshot: backup file write failed %s: %w", backupPath, err)
	}

	// Create manifest.json
	manifest := snapshotManifest{
		ProposalID: proposal.ID,
		CreatedAt:  now,
		Files: []snapshotFile{
			{
				OriginalPath: proposal.TargetPath,
				BackupName:   backupName,
			},
		},
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("createSnapshot: manifest serialization failed: %w", err)
	}

	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return "", fmt.Errorf("createSnapshot: manifest write failed: %w", err)
	}

	return snapshotDir, nil
}

// RestoreSnapshot reads manifest.json from snapshotDir and restores original files.
// REQ-HL-009: used in rollback <date> verb.
//
// @MX:ANCHOR: [AUTO] RestoreSnapshot is the core function of rollback functionality.
// @MX:REASON: [AUTO] fan_in >= 3: applier_test.go, harness CLI rollback, Phase 5 IT
func RestoreSnapshot(snapshotDir string) error {
	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("RestoreSnapshot: manifest.json 읽기 실패 %s: %w", manifestPath, err)
	}

	var manifest snapshotManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("RestoreSnapshot: manifest 파싱 실패: %w", err)
	}

	for _, f := range manifest.Files {
		backupPath := filepath.Join(snapshotDir, f.BackupName)
		backupData, err := os.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("RestoreSnapshot: 백업 파일 읽기 실패 %s: %w", backupPath, err)
		}

		// Restore to original path
		if err := os.WriteFile(f.OriginalPath, backupData, 0o644); err != nil {
			return fmt.Errorf("RestoreSnapshot: 원본 파일 복원 실패 %s: %w", f.OriginalPath, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────
// Internal helpers: frontmatter parsing and modification
// ─────────────────────────────────────────────

// splitFrontmatterBody splits SKILL.md content into frontmatter and body.
// frontmatter is content within --- delimiters, body is after the second ---.
// Returns error if frontmatter does not exist.
func splitFrontmatterBody(content string) (fm, body string, err error) {
	// Normalize newlines (CRLF → LF)
	content = strings.ReplaceAll(content, "\r\n", "\n")

	const sep = "---"

	// First line must start with ---
	if !strings.HasPrefix(content, sep+"\n") && content != sep {
		return "", "", fmt.Errorf("frontmatter 시작 구분자 없음")
	}

	// Remove first --- then find second ---
	rest := content[len(sep)+1:] // "---\n" after

	idx := strings.Index(rest, "\n"+sep+"\n")
	if idx == -1 {
		// Case with only --- at end (no "\n---" at file end)
		idx = strings.Index(rest, "\n"+sep)
		if idx == -1 {
			return "", "", fmt.Errorf("frontmatter 종료 구분자 없음")
		}
		fm = rest[:idx+1] // '\n' included
		body = ""
		return fm, body, nil
	}

	fm = rest[:idx+1]              // '\n' included
	body = rest[idx+1+len(sep)+1:] // "---\n" body after
	return fm, body, nil
}

// enrichDescriptionInFrontmatter adds "# heuristic: <note>" to description field in frontmatter YAML text.
// Returns changed=false if already exists.
// Uses line-based parsing to preserve other fields.
func enrichDescriptionInFrontmatter(fm, heuristicNote string) (newFM string, changed bool) {
	targetLine := "# heuristic: " + heuristicNote
	lines := strings.Split(fm, "\n")

	// Check if already exists (idempotent)
	for _, line := range lines {
		if strings.Contains(line, targetLine) {
			return fm, false
		}
	}

	// Find and modify description field
	var result []string
	inDescription := false
	descModified := false

	for i, line := range lines {
		// Detect line starting with description:
		if !descModified && strings.HasPrefix(strings.TrimLeft(line, " \t"), "description:") {
			trimmed := strings.TrimLeft(line, " \t")
			indent := line[:len(line)-len(trimmed)]

			// description: value (single line)
			after := strings.TrimPrefix(trimmed, "description:")
			after = strings.TrimLeft(after, " ")

			if after == "" || after == "|" || after == "|-" || after == "|+" {
				// Block scalar — cannot handle simply, add after line
				result = append(result, line)
				inDescription = true
			} else {
				// Inline value: description: original value
				result = append(result, line)
				// Insert heuristic note on next line (same indent, line continuation)
				// Simply append "\n# heuristic: ..." to description value
				// However, single-line YAML requires conversion to multi-line block
				// Simpler approach: insert immediately after description line
				_ = i
				result = append(result, indent+"# heuristic: "+heuristicNote)
				descModified = true
			}
			continue
		}

		if inDescription {
			// Inside description block
			if line == "" || (!strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t")) {
				// Block end — insert heuristic note then add current line
				result = append(result, "# heuristic: "+heuristicNote)
				result = append(result, line)
				inDescription = false
				descModified = true
				continue
			}
		}
		result = append(result, line)
	}

	if !descModified {
		return fm, false
	}

	return strings.Join(result, "\n"), true
}

// injectTriggerInFrontmatter adds keyword to triggers list in frontmatter.
// Returns changed=false if already exists.
// Returns changed=false without adding if triggers field does not exist.
func injectTriggerInFrontmatter(fm, keyword string) (newFM string, changed bool) {
	targetEntry := `keyword: "` + keyword + `"`

	// Check if already exists
	if strings.Contains(fm, targetEntry) {
		return fm, false
	}

	lines := strings.Split(fm, "\n")
	var result []string
	triggersFound := false
	lastTriggerIdx := -1

	// Find triggers: section and last trigger item position
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "triggers:") {
			triggersFound = true
		}
		if triggersFound && strings.Contains(line, `keyword:`) {
			lastTriggerIdx = i
		}
	}

	if !triggersFound || lastTriggerIdx == -1 {
		// No triggers section — return without changes
		return fm, false
	}

	// Insert new item referencing indentation level of lastTriggerIdx line
	lastLine := lines[lastTriggerIdx]
	lastTrimmed := strings.TrimLeft(lastLine, " \t")
	indent := lastLine[:len(lastLine)-len(lastTrimmed)]

	// Insert new item immediately after last trigger item
	for i, line := range lines {
		result = append(result, line)
		if i == lastTriggerIdx {
			result = append(result, indent+`keyword: "`+keyword+`"`)
		}
	}

	return strings.Join(result, "\n"), true
}
