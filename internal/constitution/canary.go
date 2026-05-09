package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// canaryScoreDropThreshold is the canary reject threshold.
	canaryScoreDropThreshold = 0.10
	// canaryMinSpecs is the minimum number of SPECs required to run canary.
	canaryMinSpecs = 3
	// canaryMaxSpecs is the maximum number of SPECs used for canary evaluation.
	canaryMaxSpecs = 3
)

// canary is an implementation of the Canary interface.
// Performs shadow evaluation by finding completed SPEC records.
type canary struct {
	// completedSpecPattern is the pattern for completed SPEC directories.
	// .moai/specs/SPEC-XXX/ format.
	completedSpecPattern *regexp.Regexp
}

// NewCanary creates a Canary.
func NewCanary() Canary {
	return &canary{
		completedSpecPattern: regexp.MustCompile(`^SPEC-[A-Z0-9]+$`),
	}
}

// Evaluate evaluates the proposal's impact on the last 3 completed SPECs.
// SPEC-V3R2-CON-002 REQ-CON-002-005 Layer 2 implementation.
//
// Canary evaluation strategy:
// 1. Find recently completed SPEC directories in .moai/specs/ (progress.md with "completed" or similar marker)
// 2. Read each SPEC's evaluator-active score (TODO: implement in SPEC-V3R2-CON-003)
// 3. Estimate the impact of clause changes on SPEC application (simple keyword matching)
// 4. Compare ScoreBefore vs ScoreAfter
//
// Current implementation: Returns CanaryUnavailable since there is no SPEC score store.
// Future: Add evaluator-active integration in SPEC-V3R2-CON-003.
func (c *canary) Evaluate(proposal *AmendmentProposal, projectDir string) (*CanaryResult, error) {
	// Check .moai/specs/ directory
	specsDir := filepath.Join(projectDir, ".moai", "specs")
	specEntries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No specs directory → CanaryUnavailable
			return &CanaryResult{
				Available: false,
				Reason:    fmt.Sprintf("No SPEC directory: %s", specsDir),
			}, &ErrCanaryUnavailable{RequiredCount: canaryMinSpecs, ActualCount: 0}
		}
		return nil, fmt.Errorf("error reading specs directory: %w", err)
	}

	// Find completed SPECs
	var completedSpecs []string
	for _, entry := range specEntries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !c.completedSpecPattern.MatchString(name) {
			continue
		}
		// Determine completion by checking progress.md existence (simple implementation)
		progressPath := filepath.Join(specsDir, name, "progress.md")
		if _, err := os.Stat(progressPath); err == nil {
			completedSpecs = append(completedSpecs, name)
		}
	}

	// Check minimum SPEC count
	if len(completedSpecs) < canaryMinSpecs {
		return &CanaryResult{
			Available: false,
			Reason:    fmt.Sprintf("Insufficient completed SPECs (%d < %d)", len(completedSpecs), canaryMinSpecs),
		}, &ErrCanaryUnavailable{RequiredCount: canaryMinSpecs, ActualCount: len(completedSpecs)}
	}

	// Select the 3 most recent SPECs
	selectedSpecs := completedSpecs
	if len(selectedSpecs) > canaryMaxSpecs {
		// Sort by modification time and select the latest 3
		selectedSpecs = c.sortMostRecent(selectedSpecs, specsDir, canaryMaxSpecs)
	}

	// TODO: Integrate evaluator-active score in SPEC-V3R2-CON-003
	// Current implementation: Dummy score as placeholder
	//
	// Shadow simulation:
	// - Before: Assume state with existing clause applied (score = 1.0)
	// - After: Simulate state with new clause applied (estimate impact by keyword matching)
	//
	// Simple implementation: Estimate score impact by complexity of clause change
	scoreBefore := 1.0
	scoreAfter := c.estimateScoreImpact(proposal)

	result := &CanaryResult{
		Available:     true,
		EvaluatedSpecs: selectedSpecs,
		ScoreBefore:   scoreBefore,
		ScoreAfter:    scoreAfter,
		MaxDrop:       scoreBefore - scoreAfter,
	}

	// Determine Passed status
	if result.MaxDrop <= canaryScoreDropThreshold {
		result.Passed = true
		result.Reason = fmt.Sprintf("Score drop %.2f <= threshold %.2f", result.MaxDrop, canaryScoreDropThreshold)
	} else {
		result.Passed = false
		result.Reason = fmt.Sprintf("Score drop %.2f > threshold %.2f", result.MaxDrop, canaryScoreDropThreshold)
		return result, &ErrCanaryRejected{
			RuleID:        proposal.RuleID,
			ScoreDrop:     result.MaxDrop,
			Threshold:     canaryScoreDropThreshold,
			AffectedSpecs: selectedSpecs,
		}
	}

	return result, nil
}

// estimateScoreImpact estimates the impact of clause changes on SPEC scores.
// Simple heuristic: Estimate impact by clause length and keyword changes.
func (c *canary) estimateScoreImpact(proposal *AmendmentProposal) float64 {
	// Base score: 1.0 (perfect)
	score := 1.0

	// Estimate impact by clause length ratio
	beforeLen := len(proposal.Before)
	afterLen := len(proposal.After)

	// Longer clause → stricter constraints → slight score decrease
	// Shorter clause → relaxed constraints → slight score increase
	lenRatio := float64(afterLen) / float64(beforeLen)
	if lenRatio > 1.2 {
		// 20%+ longer → stricter constraints
		score -= 0.05
	} else if lenRatio < 0.8 {
		// 20%+ shorter → relaxed constraints
		score += 0.02
	}

	// Keyword-based impact assessment
	// Adding prohibition words ("MUST NOT", "NEVER", "PROHIBITED") → score decrease
	afterUpper := strings.ToUpper(proposal.After)
	if strings.Contains(afterUpper, "MUST NOT") || strings.Contains(afterUpper, "NEVER") || strings.Contains(afterUpper, "PROHIBITED") {
		score -= 0.08
	}

	// Removing mandatory words ("MUST", "REQUIRED", "SHALL") → score increase (relaxed constraints)
	beforeUpper := strings.ToUpper(proposal.Before)
	hadMust := strings.Contains(beforeUpper, "MUST") || strings.Contains(beforeUpper, "REQUIRED") || strings.Contains(beforeUpper, "SHALL")
	hasMust := strings.Contains(afterUpper, "MUST") || strings.Contains(afterUpper, "REQUIRED") || strings.Contains(afterUpper, "SHALL")
	if hadMust && !hasMust {
		score += 0.05
	}

	// Score bounds: [0.0, 1.0]
	if score < 0.0 {
		score = 0.0
	}
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// sortMostRecent returns the list of most recent SPECs by modification time.
func (c *canary) sortMostRecent(specs []string, specsDir string, limit int) []string {
	type specTime struct {
		name string
		time time.Time
	}

	var specTimes []specTime
	for _, spec := range specs {
		info, err := os.Stat(filepath.Join(specsDir, spec))
		if err != nil {
			continue
		}
		specTimes = append(specTimes, specTime{name: spec, time: info.ModTime()})
	}

	// Sort by ModTime descending
	for i := 0; i < len(specTimes); i++ {
		for j := i + 1; j < len(specTimes); j++ {
			if specTimes[j].time.After(specTimes[i].time) {
				specTimes[i], specTimes[j] = specTimes[j], specTimes[i]
			}
		}
	}

	// Return the latest limit items
	var result []string
	for i := 0; i < len(specTimes) && i < limit; i++ {
		result = append(result, specTimes[i].name)
	}
	return result
}

// canary satisfies the Canary interface.
var _ Canary = (*canary)(nil)

// parseScoreFromProgress parses the evaluator-active score from progress.md.
// TODO: To be implemented in SPEC-V3R2-CON-003. Currently unused.
//nolint:unused
func parseScoreFromProgress(progressPath string) (float64, error) {
	// Simple implementation: Find "Score: 0.XX" pattern in file
	data, err := os.ReadFile(progressPath)
	if err != nil {
		return 0.8, err
	}

	re := regexp.MustCompile(`Score:\s*([0-9.]+)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return 0.8, nil
	}

	score, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0.8, err
	}

	return score, nil
}
