// Package safety — Canary Veto: provisional apply + auto-rollback (M3 EXTEND).
// Implements the E5 Canary Veto Policy from Vision §6.5 and plan.md §3.3.
//
// Lifecycle:
//  1. L5 user approval arrives before Canary completes → ProvisionalApply (write + snapshot).
//  2. Canary PASS → call Confirm (set evolution_status=applied).
//  3. Canary FAIL → call VetoAndRollback (revert file + 48h cooldown entry).
//
// [HARD] This package does not call AskUserQuestion. It emits blocker reports only.
// Orchestrator handles user interaction per L3+L5 unified blocker-report pattern (plan.md §3.3a).
//
// @MX:ANCHOR: [AUTO] CanaryVeto is the Canary veto gateway for provisional evolution.
// @MX:REASON: [AUTO] fan_in >= 3: canary_veto_test.go, pipeline (M3 extend), integration_test.go
package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// canaryVetoCooldown is the cooldown duration after a Canary veto (48 hours per plan.md §3.3).
const canaryVetoCooldown = 48 * time.Hour

// CanaryVetoConfig holds configuration for CanaryVeto.
type CanaryVetoConfig struct {
	// RevertDir is the directory where pre-apply snapshots are stored.
	// Snapshots are written to RevertDir/<evolution-id>/<filename>.
	RevertDir string

	// RateLimitStatePath is the path to the rate-limit-state.json file
	// used to track the 48h cooldown entries from Canary veto.
	RateLimitStatePath string
}

// vetoRecord is the schema for .moai/harness/revert/<evo-id>/veto.log.
type vetoRecord struct {
	EvolutionID     string    `json:"evolution_id"`
	EvolutionStatus string    `json:"evolution_status"` // "vetoed_by_canary"
	VetoedAt        time.Time `json:"vetoed_at"`
	TargetPath      string    `json:"target_path"`
	Reason          string    `json:"reason"`
}

// cooldownEntry is persisted in rate-limit-state.json to track per-proposal veto cooldowns.
type cooldownEntry struct {
	ProposalID   string    `json:"proposal_id"`
	CooldownUntil time.Time `json:"cooldown_until"`
}

// CanaryVeto manages the provisional apply + veto + rollback lifecycle.
//
// [HARD] No AskUserQuestion calls. Blocker reports are emitted via structured error values.
type CanaryVeto struct {
	cfg   CanaryVetoConfig
	nowFn func() time.Time
}

// NewCanaryVeto creates a CanaryVeto with the given config.
func NewCanaryVeto(cfg CanaryVetoConfig) *CanaryVeto {
	return &CanaryVeto{cfg: cfg, nowFn: time.Now}
}

// ProvisionalApply snapshots the current target file content and writes the new value.
// evolution_status is set to "provisional" until Confirm or VetoAndRollback is called.
// Implements copy-before-write snapshot (plan.md §3.3: "snapshot to .moai/harness/revert/...").
func (v *CanaryVeto) ProvisionalApply(proposal harness.Proposal) error {
	if proposal.TargetPath == "" {
		return fmt.Errorf("canary_veto: ProvisionalApply: target_path is required")
	}

	// Step 1: Snapshot current content
	if err := v.snapshotFile(proposal); err != nil {
		return fmt.Errorf("canary_veto: snapshot: %w", err)
	}

	// Step 2: Write new content
	if err := writeFileContent(proposal.TargetPath, proposal.NewValue); err != nil {
		return fmt.Errorf("canary_veto: write new value: %w", err)
	}

	return nil
}

// Confirm marks evolution as applied (called when Canary PASS after provisional apply).
// Writes evolution_status=applied to veto.log.
func (v *CanaryVeto) Confirm(proposal harness.Proposal) error {
	return v.writeVetoLog(proposal.ID, vetoRecord{
		EvolutionID:     proposal.ID,
		EvolutionStatus: "applied",
		VetoedAt:        v.nowFn().UTC(),
		TargetPath:      proposal.TargetPath,
		Reason:          "canary PASS — provisional confirmed",
	})
}

// VetoAndRollback reverts the target file to its snapshot content and records a 48h
// cooldown entry in rate-limit-state.json (plan.md §3.3 E5 step 2(b)).
//
// After rollback, the orchestrator should present the B4 notification text to the user
// via AskUserQuestion (orchestrator responsibility — not this package).
func (v *CanaryVeto) VetoAndRollback(proposal harness.Proposal) error {
	// Step 1: Restore original content from snapshot
	if err := v.restoreSnapshot(proposal); err != nil {
		return fmt.Errorf("canary_veto: restore: %w", err)
	}

	// Step 2: Write veto.log with evolution_status=vetoed_by_canary
	if err := v.writeVetoLog(proposal.ID, vetoRecord{
		EvolutionID:     proposal.ID,
		EvolutionStatus: "vetoed_by_canary",
		VetoedAt:        v.nowFn().UTC(),
		TargetPath:      proposal.TargetPath,
		Reason:          "HARNESS_LEARNING_CANARY_VETO: canary regression detected, provisional change rolled back",
	}); err != nil {
		return fmt.Errorf("canary_veto: write veto log: %w", err)
	}

	// Step 3: Record 48h cooldown in rate-limiter
	if err := v.RecordCooldown(proposal); err != nil {
		return fmt.Errorf("canary_veto: record cooldown: %w", err)
	}

	return nil
}

// RecordCooldown writes a 48h cooldown entry for the proposal to rate-limit-state.json.
func (v *CanaryVeto) RecordCooldown(proposal harness.Proposal) error {
	entries, err := v.loadCooldowns()
	if err != nil {
		return fmt.Errorf("canary_veto: load cooldowns: %w", err)
	}

	now := v.nowFn()
	// Replace or append
	found := false
	for i := range entries {
		if entries[i].ProposalID == proposal.ID {
			entries[i].CooldownUntil = now.Add(canaryVetoCooldown)
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, cooldownEntry{
			ProposalID:    proposal.ID,
			CooldownUntil: now.Add(canaryVetoCooldown),
		})
	}

	return v.saveCooldowns(entries)
}

// CheckCooldown returns an error if the proposal is still within its 48h veto cooldown.
// The error contains HARNESS_LEARNING_RATELIMIT_EXCEEDED (AC-HRA-008b).
func (v *CanaryVeto) CheckCooldown(proposal harness.Proposal) error {
	entries, err := v.loadCooldowns()
	if err != nil {
		return fmt.Errorf("canary_veto: load cooldowns: %w", err)
	}

	now := v.nowFn()
	for _, e := range entries {
		if e.ProposalID == proposal.ID && now.Before(e.CooldownUntil) {
			remaining := e.CooldownUntil.Sub(now).Round(time.Minute)
			return fmt.Errorf("HARNESS_LEARNING_RATELIMIT_EXCEEDED: proposal %s is in 48h veto cooldown (%v remaining)",
				proposal.ID, remaining)
		}
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────

// snapshotFile copies current target file to RevertDir/<evo-id>/<basename>.
func (v *CanaryVeto) snapshotFile(proposal harness.Proposal) error {
	snapDir := filepath.Join(v.cfg.RevertDir, proposal.ID)
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return fmt.Errorf("mkdirall %s: %w", snapDir, err)
	}

	data, err := os.ReadFile(proposal.TargetPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read target %s: %w", proposal.TargetPath, err)
	}
	// Snapshot file
	basename := filepath.Base(proposal.TargetPath)
	snapPath := filepath.Join(snapDir, basename)
	if err := os.WriteFile(snapPath, data, 0o644); err != nil {
		return fmt.Errorf("write snapshot %s: %w", snapPath, err)
	}
	return nil
}

// restoreSnapshot copies the snapshot back to the original target path.
func (v *CanaryVeto) restoreSnapshot(proposal harness.Proposal) error {
	snapDir := filepath.Join(v.cfg.RevertDir, proposal.ID)
	basename := filepath.Base(proposal.TargetPath)
	snapPath := filepath.Join(snapDir, basename)

	data, err := os.ReadFile(snapPath)
	if err != nil {
		return fmt.Errorf("read snapshot %s: %w", snapPath, err)
	}
	return writeFileContent(proposal.TargetPath, string(data))
}

// writeVetoLog serializes and appends a vetoRecord to RevertDir/<evo-id>/veto.log.
func (v *CanaryVeto) writeVetoLog(evolutionID string, record vetoRecord) error {
	logDir := filepath.Join(v.cfg.RevertDir, evolutionID)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("mkdirall %s: %w", logDir, err)
	}
	logPath := filepath.Join(logDir, "veto.log")

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal veto record: %w", err)
	}
	data = append(data, '\n')

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open veto.log: %w", err)
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(data)
	return err
}

// loadCooldowns reads cooldown entries from rate-limit-state.json.
func (v *CanaryVeto) loadCooldowns() ([]cooldownEntry, error) {
	data, err := os.ReadFile(v.cfg.RateLimitStatePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", v.cfg.RateLimitStatePath, err)
	}

	// Try to unmarshal as a slice of cooldownEntry
	var entries []cooldownEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		// Might be in old format (rateLimitState) — treat as empty cooldown list
		return nil, nil
	}
	return entries, nil
}

// saveCooldowns writes cooldown entries to rate-limit-state.json (atomic write).
func (v *CanaryVeto) saveCooldowns(entries []cooldownEntry) error {
	if dir := filepath.Dir(v.cfg.RateLimitStatePath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdirall %s: %w", dir, err)
		}
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal cooldowns: %w", err)
	}
	tmpPath := v.cfg.RateLimitStatePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	return os.Rename(tmpPath, v.cfg.RateLimitStatePath)
}

// writeFileContent writes content to path (creates parent dirs if needed).
func writeFileContent(path, content string) error {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdirall %s: %w", dir, err)
		}
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmpPath, err)
	}
	return os.Rename(tmpPath, path)
}
