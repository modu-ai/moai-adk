// Package harness — frontmatter modification applier.
// REQ-HL-003: description enrichment (Tier 2 heuristic).
// REQ-HL-004: trigger injection (Tier 3 rule, feature-gated).
// REQ-HL-005: Apply() — create snapshot first then modify files (Phase 4).
package harness

import (
	"encoding/json"
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
}

// NewApplier creates a default Applier.
// Actual file writes in InjectTrigger follow package-level flag (enableTriggerInjectionWrites).
func NewApplier() *Applier {
	return &Applier{allowWrites: enableTriggerInjectionWrites}
}

// newApplierWithWritesEnabled creates an Applier with InjectTrigger actual writes enabled.
// Test-only function.
func newApplierWithWritesEnabled() *Applier {
	return &Applier{allowWrites: true}
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
		return fmt.Errorf("applier: proposal rejected (L%d, rejected)", decision.RejectedBy)

	case DecisionPendingApproval:
		// [HARD] subagent must not call AskUserQuestion directly.
		// Return payload to orchestrator to delegate user approval.
		return &ApplyPendingError{OversightPayload: decision.OversightProposal}

	case DecisionApproved:
		// approved — continue
	}

	// ── Step 2: Create Snapshot (before write) ───────────────────────────────
	// [HARD] Abort write on snapshot failure.
	if err := a.createSnapshot(proposal, snapshotBase); err != nil {
		return fmt.Errorf("applier: snapshot creation failed — abort write: %w", err)
	}

	// ── Step 3: Actual File Modification ───────────────────────────────────────────────
	switch proposal.FieldKey {
	case "description":
		return a.EnrichDescription(proposal.TargetPath, proposal.NewValue)
	case "triggers":
		// Perform InjectTrigger with write-enabled Applier
		w := newApplierWithWritesEnabled()
		return w.InjectTrigger(proposal.TargetPath, proposal.NewValue)
	default:
		return fmt.Errorf("applier: unsupported fieldKey %q", proposal.FieldKey)
	}
}

// createSnapshot backs up current content of proposal.TargetPath to snapshotBase/<ISO-DATE>/.
// Creates manifest.json then performs file copy.
func (a *Applier) createSnapshot(proposal Proposal, snapshotBase string) error {
	// Generate ISO-DATE format directory name (date + nano collision prevention)
	now := time.Now().UTC()
	dirName := now.Format("2006-01-02T15-04-05.000000000Z")
	snapshotDir := filepath.Join(snapshotBase, dirName)

	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return fmt.Errorf("createSnapshot: directory creation failed %s: %w", snapshotDir, err)
	}

	// Read original file
	originalData, err := os.ReadFile(proposal.TargetPath)
	if err != nil {
		return fmt.Errorf("createSnapshot: original file read failed %s: %w", proposal.TargetPath, err)
	}

	// Backup filename: use original filename as-is
	backupName := filepath.Base(proposal.TargetPath)
	backupPath := filepath.Join(snapshotDir, backupName)

	if err := os.WriteFile(backupPath, originalData, 0o644); err != nil {
		return fmt.Errorf("createSnapshot: backup file write failed %s: %w", backupPath, err)
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
		return fmt.Errorf("createSnapshot: manifest serialization failed: %w", err)
	}

	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return fmt.Errorf("createSnapshot: manifest write failed: %w", err)
	}

	return nil
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
		fm = rest[:idx+1]  // '\n' included
		body = ""
		return fm, body, nil
	}

	fm = rest[:idx+1]               // '\n' included
	body = rest[idx+1+len(sep)+1:]  // "---\n" body after
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
