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

	// rename is the forward-rename operation of the apply step; nil means
	// os.Rename. restore writes one file's pre-apply bytes back, or removes a
	// file that did not exist; nil means restoreFile. Per-pipeline fields, not
	// package variables, so a test can fail the Nth rename or a restore without
	// touching process-global state (SPEC-CON-AMEND-APPLY-001 REQ-CAA-011).
	rename  func(oldpath, newpath string) error
	restore func(path string, data []byte, existed bool) error
}

// NewPipeline creates a Pipeline with default implementations.
func NewPipeline() *Pipeline {
	return &Pipeline{
		FrozenGuard:           NewFrozenGuard(),
		Canary:                NewCanary(),
		ContradictionDetector: NewContradictionDetector(),
		RateLimiter:           NewRateLimiter(),
		HumanOversight:        NewHumanOversight(),
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

	// Resolve the registry path with the resolver the CLI uses (REQ-CAA-019)
	// and admit it, every entry's file:, and the evolution log only when they
	// lie inside projectDir (REQ-CAA-020, REQ-CAA-021). The registry-path
	// check runs before the registry is read.
	registryPath := ResolveRegistryPath(projectDir)
	registry, err := LoadAmendRegistry(registryPath, projectDir)
	if err != nil {
		return nil, fmt.Errorf("registry load error: %w", err)
	}
	evolutionLogPath := filepath.Join(projectDir, ".moai", "research", "evolution-log.md")
	if err := checkContained(projectDir, evolutionLogPath); err != nil {
		return nil, fmt.Errorf("evolution log load error: %w", err)
	}

	// Lookup current rule
	currentRule, exists := registry.Get(proposal.RuleID)
	if !exists {
		return nil, fmt.Errorf("rule %q not found", proposal.RuleID)
	}

	// The proposal must amend the clause the registry holds now, byte for
	// byte, in both modes, before any gate runs (REQ-CAA-017).
	if proposal.Before != currentRule.Clause {
		return nil, fmt.Errorf("proposal for rule %s: Before differs from the current registry clause", proposal.RuleID)
	}
	if proposal.After == "" {
		return nil, fmt.Errorf("proposal for rule %s: After is empty", proposal.RuleID)
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
	// Validate against the real bytes in both modes, so a dry-run fails
	// exactly where a real apply would (REQ-CAA-012); a dry-run writes nothing.
	changes, log, err := p.prepareApply(proposal, currentRule, projectDir, registryPath, evolutionLogPath)
	if err != nil {
		return nil, fmt.Errorf("amendment validation error: %w", err)
	}
	if dryRun {
		return log, nil
	}
	if err := p.commitChanges(changes); err != nil {
		return nil, fmt.Errorf("amendment application error: %w", err)
	}
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
		ID:             "", // Generated later in Execute
		RuleID:         proposal.RuleID,
		ZoneBefore:     originalZone,
		ZoneAfter:      originalZone, // Zone changes only allowed with demotion evidence
		ClauseBefore:   proposal.Before,
		ClauseAfter:    proposal.After,
		CanaryVerdict:  canaryVerdict,
		Contradictions: contradictions,
		ApprovedBy:     proposal.ApprovedBy,
		ApprovedAt:     proposal.ApprovedAt,
		RolledBack:     false,
	}
}

// prepareApply runs every pre-apply validation against the real bytes of the
// three files and returns their new contents plus the new log entry, writing
// nothing (REQ-CAA-001 … REQ-CAA-004, REQ-CAA-009, REQ-CAA-016). The evolution
// log content is its pre-apply bytes followed by the new entry (REQ-CAA-010).
func (p *Pipeline) prepareApply(proposal *AmendmentProposal, rule Rule, projectDir, registryPath, logPath string) ([]*fileChange, *AmendmentLog, error) {
	source, err := readForChange("rule file", ruleFilePath(projectDir, rule.File), false)
	if err != nil {
		return nil, nil, err
	}
	registry, err := readForChange("registry", registryPath, false)
	if err != nil {
		return nil, nil, err
	}
	logFile, err := readForChange("evolution log", logPath, true)
	if err != nil {
		return nil, nil, err
	}
	if sameFile(source.path, registry.path) || sameFile(source.path, logFile.path) {
		return nil, nil, fmt.Errorf("rule file %s: is also the registry or the evolution log", source.path)
	}

	if source.newData, err = replaceSourceClause(source.path, source.oldData, rule.Clause, proposal.After); err != nil {
		return nil, nil, err
	}
	if registry.newData, err = rewriteRegistryClause(registry.path, registry.oldData, proposal.RuleID, proposal.After); err != nil {
		return nil, nil, err
	}

	logs, err := parseEvolutionLog(logFile.path, string(logFile.oldData))
	if err != nil {
		return nil, nil, err
	}
	log := p.createLogEntry(proposal, rule.Zone)
	log.ID = GenerateLogID(time.Now(), logs)
	if err := log.Validate(); err != nil {
		return nil, nil, fmt.Errorf("evolution log %s: %w", logFile.path, err)
	}
	entry, err := formatLogEntry(log)
	if err != nil {
		return nil, nil, err
	}
	prefix := logFile.oldData
	if len(prefix) > 0 && prefix[len(prefix)-1] != '\n' {
		prefix = append(append([]byte{}, prefix...), '\n')
	}
	logFile.newData = append(append([]byte{}, prefix...), entry...)

	return []*fileChange{source, registry, logFile}, log, nil
}

// sameFile reports whether two paths name the same file after cleaning.
func sameFile(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	return errA == nil && errB == nil && absA == absB
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
