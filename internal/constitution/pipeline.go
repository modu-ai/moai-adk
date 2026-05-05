package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Pipeline is the constitutional amendment pipeline that executes the 5-layer safety gate.
// Implements SPEC-V3R2-CON-002 REQ-CON-002-002.
type Pipeline struct {
	// FrozenGuard is Layer 1 gate.
	FrozenGuard FrozenGuard
	// Canary is Layer 2 gate.
	Canary Canary
	// ContradictionDetector is Layer 3 gate.
	ContradictionDetector ContradictionDetector
	// RateLimiter is Layer 4 gate.
	RateLimiter RateLimiter
	// HumanOversight is Layer 5 gate.
	HumanOversight HumanOversight

	// LockFilePath is the single-writer lock file path.
	LockFilePath string
}

// NewPipeline creates a Pipeline with default implementations.
func NewPipeline() *Pipeline {
	return &Pipeline{
		FrozenGuard:          NewFrozenGuard(),
		Canary:               NewCanary(),
		ContradictionDetector: NewContradictionDetector(),
		RateLimiter:          NewRateLimiter(),
		HumanOversight:       NewHumanOversight(),
	}
}

// Execute executes the 5-layer safety gate on a proposal and applies the amendment.
// Dry-run mode: executes all layers but does not modify files.
// Implements SPEC-V3R2-CON-002 AC-CON-002-001.
//
// Layer execution order (FROZEN):
// 1. FrozenGuard: Frozen zone check
// 2. Canary: Shadow evaluation
// 3. ContradictionDetector: Conflict scan
// 4. RateLimiter: Frequency check
// 5. HumanOversight: User approval
//
// On success:
// - Update registry file (modify source rule file)
// - Update zone registry (.claude/rules/moai/core/zone-registry.md)
// - Record in evolution-log.md
//
// On failure: returns error from the corresponding layer.
func (p *Pipeline) Execute(proposal *AmendmentProposal, projectDir string, dryRun bool) (*AmendmentLog, error) {
	// 0. Attempt to acquire single-writer lock
	if err := p.acquireLock(dryRun); err != nil {
		return nil, err
	}
	if !dryRun {
		defer p.releaseLock()
	}

	// Load registry
	registryPath := filepath.Join(projectDir, ".claude", "rules", "moai", "core", "zone-registry.md")
	registry, err := LoadRegistry(registryPath, projectDir)
	if err != nil {
		return nil, fmt.Errorf("registry load error: %w", err)
	}

	// Lookup current rule
	currentRule, exists := registry.Get(proposal.RuleID)
	if !exists {
		return nil, fmt.Errorf("rule %q not found", proposal.RuleID)
	}

	// Skip Canary for rules with canary_gate=false
	skipCanary := !currentRule.CanaryGate

	// ===== Layer 1: FrozenGuard =====
	if err := p.FrozenGuard.Check(proposal, currentRule.Zone); err != nil {
		return nil, fmt.Errorf("layer 1 (FrozenGuard) failed: %w", err)
	}

	// ===== Layer 2: Canary =====
	if !skipCanary {
		canaryResult, err := p.Canary.Evaluate(proposal, projectDir)
		proposal.CanaryResult = canaryResult
		if err != nil {
			// CanaryUnavailable is not fatal (similar to skip)
			if _, unavailable := err.(*ErrCanaryUnavailable); !unavailable {
				return nil, fmt.Errorf("layer 2 (Canary) failed: %w", err)
			}
			// CanaryUnavailable continues
		} else if !canaryResult.Passed {
			return nil, fmt.Errorf("layer 2 (Canary) failed: score drop %.2f > threshold %.2f",
				canaryResult.MaxDrop, canaryScoreDropThreshold)
		}
	} else {
		proposal.CanaryResult = &CanaryResult{
			Available: false,
			Reason:    fmt.Sprintf("Rule %q has canary_gate=false", proposal.RuleID),
		}
	}

	// ===== Layer 3: ContradictionDetector =====
	contradictionResult, err := p.ContradictionDetector.Scan(proposal, registry)
	proposal.Contradicts = contradictionResult
	if err != nil {
		return nil, fmt.Errorf("layer 3 (ContradictionDetector) failed: %w", err)
	}

	// ===== Layer 4: RateLimiter =====
	evolutionLogPath := filepath.Join(projectDir, ".moai", "research", "evolution-log.md")
	if err := p.RateLimiter.Admit(proposal, evolutionLogPath); err != nil {
		return nil, fmt.Errorf("layer 4 (RateLimiter) failed: %w", err)
	}

	// ===== Layer 5: HumanOversight =====
	approved, err := p.HumanOversight.Approve(proposal, dryRun)
	if err != nil {
		return nil, fmt.Errorf("layer 5 (HumanOversight) failed: %w", err)
	}
	if !approved {
		return nil, fmt.Errorf("user rejected the amendment")
	}
	proposal.Approved = true
	proposal.ApprovedBy = "human"
	proposal.ApprovedAt = time.Now()

	// ===== Apply Amendment =====
	if dryRun {
		// Dry-run: only return log creation
		log := p.createLogEntry(proposal, currentRule.Zone)
		return log, nil
	}

	// Actual application: update source file, registry, evolution-log
	if err := p.applyAmendment(proposal, currentRule, projectDir, registryPath); err != nil {
		return nil, fmt.Errorf("amendment application error: %w", err)
	}

	log := p.createLogEntry(proposal, currentRule.Zone)
	return log, nil
}

// createLogEntry creates an AmendmentLog from a proposal.
func (p *Pipeline) createLogEntry(proposal *AmendmentProposal, originalZone Zone) *AmendmentLog {
	// Canary verdict
	canaryVerdict := "skipped"
	if proposal.CanaryResult != nil {
		if proposal.CanaryResult.Available {
			if proposal.CanaryResult.Passed {
				canaryVerdict = "passed"
			} else {
				canaryVerdict = "rejected"
			}
		} else {
			canaryVerdict = "unavailable"
		}
	}

	// Contradictions
	var contradictions []string
	if proposal.Contradicts != nil {
		for _, c := range proposal.Contradicts.Conflicts {
			contradictions = append(contradictions,
				fmt.Sprintf("%s: %s", c.ConflictingRuleID, c.Description))
		}
	}

	return &AmendmentLog{
		ID:            "", // Generated later in Execute
		RuleID:        proposal.RuleID,
		ZoneBefore:    originalZone,
		ZoneAfter:     originalZone, // Zone changes only allowed with demotion evidence
		ClauseBefore:  proposal.Before,
		ClauseAfter:   proposal.After,
		CanaryVerdict: canaryVerdict,
		Contradictions: contradictions,
		ApprovedBy:    proposal.ApprovedBy,
		ApprovedAt:    proposal.ApprovedAt,
		RolledBack:    false,
	}
}

// applyAmendment applies the amendment to the file system.
// Modifies 3 files: source rule file, zone registry, evolution-log.
func (p *Pipeline) applyAmendment(proposal *AmendmentProposal, rule Rule, projectDir, registryPath string) error {
	// 1. Update source rule file
	sourceFilePath := rule.File
	if !filepath.IsAbs(sourceFilePath) {
		sourceFilePath = filepath.Join(projectDir, sourceFilePath)
	}
	if err := updateSourceFile(sourceFilePath, rule.Anchor, proposal.After); err != nil {
		return fmt.Errorf("source file update error: %w", err)
	}

	// 2. Update zone registry (Clause only)
	if err := updateRegistryClause(registryPath, proposal.RuleID, proposal.After); err != nil {
		return fmt.Errorf("registry update error: %w", err)
	}

	// 3. Record in evolution-log
	evolutionLogPath := filepath.Join(projectDir, ".moai", "research", "evolution-log.md")
	logs, _ := LoadEvolutionLogs(evolutionLogPath)
	log := p.createLogEntry(proposal, rule.Zone)
	log.ID = GenerateLogID(time.Now(), logs)

	if err := AppendEvolutionLog(evolutionLogPath, log); err != nil {
		return fmt.Errorf("evolution-log recording error: %w", err)
	}

	return nil
}

// acquireLock acquires the single-writer lock.
func (p *Pipeline) acquireLock(dryRun bool) error {
	if dryRun {
		return nil // Dry-run does not require lock
	}

	lockPath := p.LockFilePath
	if lockPath == "" {
		// Use default path
		lockPath = ".moai/research/.amendment.lock"
	}

	if _, err := os.Stat(lockPath); err == nil {
		return &ErrAmendmentInProgress{LockFilePath: lockPath}
	}

	// Create lock file
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return fmt.Errorf("lock directory creation error: %w", err)
	}
	if err := os.WriteFile(lockPath, []byte(time.Now().Format(time.RFC3339)), 0644); err != nil {
		return fmt.Errorf("lock file creation error: %w", err)
	}

	p.LockFilePath = lockPath
	return nil
}

// releaseLock releases the single-writer lock.
func (p *Pipeline) releaseLock() {
	if p.LockFilePath != "" {
		_ = os.Remove(p.LockFilePath)
		p.LockFilePath = ""
	}
}

// updateSourceFile updates the clause for the corresponding anchor in the source rule file.
// TODO: Implementation needed - anchor search + replacement logic.
func updateSourceFile(filePath, anchor, newClause string) error {
	// Simple implementation: read entire file and replace line after anchor
	// Actual implementation requires markdown section search
	return fmt.Errorf("updateSourceFile: not yet implemented (path=%s, anchor=%s)", filePath, anchor)
}

// updateRegistryClause updates the clause for the corresponding rule in the zone registry.
// TODO: Implementation needed - YAML parsing + replacement logic.
func updateRegistryClause(registryPath, ruleID, newClause string) error {
	// Simple implementation: parse YAML, find by rule ID, replace clause
	return fmt.Errorf("updateRegistryClause: not yet implemented (path=%s, rule=%s)", registryPath, ruleID)
}
