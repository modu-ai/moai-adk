// Package hook implements reflective learning analysis for the MoAI-ADK
// Write Phase (SPEC-REFLECT-001).
//
// AnalyzeSession is the primary entry point.  It reads today's telemetry
// records for a session, groups them by skill, and generates conservative
// LearningEntry proposals that a human can review before any file is changed.
package hook

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/evolution"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// AnalyzeSession examines today's telemetry for sessionID and proposes
// LearningEntry improvements.
//
// Safety contract:
//   - Only skills with identifiable evolvable zones are eligible targets.
//   - FrozenGuard and RateLimiter are applied before saving any proposal.
//   - Maximum MaxSessionProposals proposals per call.
//   - Sessions with fewer than MinToolInvocationsForAnalysis records are skipped.
//
// The returned entries have already been persisted to disk and are waiting
// for human review.  None of them will ever be auto-applied.
func AnalyzeSession(projectRoot, sessionID string) ([]*evolution.LearningEntry, error) {
	records, err := telemetry.LoadBySession(projectRoot, sessionID)
	if err != nil {
		return nil, fmt.Errorf("reflective_write: load telemetry: %w", err)
	}

	if len(records) < evolution.MinToolInvocationsForAnalysis {
		slog.Debug("reflective_write: skipping trivial session",
			"session_id", sessionID,
			"record_count", len(records),
		)
		return nil, nil
	}

	// Convert to pointer slice for groupBySkill.
	ptrs := make([]*telemetry.UsageRecord, len(records))
	for i := range records {
		ptrs[i] = &records[i]
	}

	// Group records by skill ID.
	bySkill := groupBySkill(ptrs)

	var proposals []*evolution.LearningEntry

	for skillID, skillRecords := range bySkill {
		if skillID == "" {
			continue // No skill identified — skip.
		}

		// Find the skill's SKILL.md file.
		skillFile := skillFilePath(projectRoot, skillID)
		if skillFile == "" {
			slog.Debug("reflective_write: skill file not found, skipping",
				"skill_id", skillID,
			)
			continue
		}

		// Apply frozen guard.
		relPath := strings.TrimPrefix(skillFile, projectRoot+string(os.PathSeparator))
		relPath = filepath.ToSlash(relPath)
		if err := evolution.CheckFrozenGuard(relPath); err != nil {
			slog.Debug("reflective_write: skill file is frozen, skipping",
				"skill_id", skillID,
				"path", relPath,
			)
			continue
		}

		// Apply rate limiter.
		if err := evolution.CheckRateLimit(projectRoot); err != nil {
			slog.Debug("reflective_write: rate limit reached, stopping",
				"session_id", sessionID,
			)
			break
		}
		if err := evolution.CheckFileCooldown(projectRoot, relPath); err != nil {
			slog.Debug("reflective_write: file in cooldown, skipping",
				"skill_id", skillID,
				"path", relPath,
			)
			continue
		}

		// Determine whether to propose a learning.
		entry := buildLearningProposal(skillID, relPath, sessionID, skillRecords)
		if entry == nil {
			continue
		}

		// Duplicate detection.
		dup, err := evolution.DetectDuplicate(projectRoot, entry.Observation)
		if err != nil {
			slog.Warn("reflective_write: duplicate detection error", "error", err)
		}
		if dup != nil {
			// Update the existing entry's observation count instead.
			_ = evolution.UpdateLearning(projectRoot, dup.ID, func(e *evolution.LearningEntry) {
				e.Observations++
				e.Evidence = append(e.Evidence, entry.Evidence...)
				e.Updated = time.Now()
				e.Confidence = evolution.CalculateConfidence(e)
				newStatus := evolution.EvaluateGraduation(e)
				if newStatus != e.Status {
					e.Status = newStatus
				}
			})
			slog.Debug("reflective_write: incremented existing learning",
				"id", dup.ID,
				"skill_id", skillID,
			)
			continue
		}

		// Anti-pattern detection.
		if evolution.DetectAntiPattern(entry) {
			entry.Status = evolution.StatusAntiPattern
		}

		if err := evolution.CreateLearning(projectRoot, entry); err != nil {
			slog.Warn("reflective_write: failed to save learning",
				"skill_id", skillID,
				"error", err,
			)
			continue
		}

		// Update rate-limit state.
		if err := evolution.UpdateRateLimit(projectRoot, relPath); err != nil {
			slog.Warn("reflective_write: failed to update rate limit", "error", err)
		}

		proposals = append(proposals, entry)
		if len(proposals) >= evolution.MaxSessionProposals {
			slog.Debug("reflective_write: proposal cap reached", "session_id", sessionID)
			break
		}
	}

	return proposals, nil
}

// groupBySkill groups telemetry records by their SkillID.
func groupBySkill(records []*telemetry.UsageRecord) map[string][]*telemetry.UsageRecord {
	out := make(map[string][]*telemetry.UsageRecord)
	for _, r := range records {
		out[r.SkillID] = append(out[r.SkillID], r)
	}
	return out
}

// skillFilePath returns the absolute path to the SKILL.md file for skillID,
// or an empty string if the file does not exist.
func skillFilePath(projectRoot, skillID string) string {
	path := filepath.Join(projectRoot, defs.ClaudeDir, defs.SkillsSubdir, skillID, "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

// buildLearningProposal analyses the records for a single skill and returns
// a LearningEntry proposal, or nil if the patterns are too weak.
//
// Conservative heuristic:
//   - If the error rate is >= 50% AND there are at least 2 error records,
//     generate an error-based proposal.
//   - Otherwise return nil (do not propose anything).
func buildLearningProposal(skillID, relPath, sessionID string, records []*telemetry.UsageRecord) *evolution.LearningEntry {
	if len(records) < 2 {
		return nil // Not enough signal.
	}

	var errorCount int
	for _, r := range records {
		if r.Outcome == telemetry.OutcomeError {
			errorCount++
		}
	}

	errorRate := float64(errorCount) / float64(len(records))
	if errorRate < 0.50 || errorCount < 2 {
		return nil // Below threshold.
	}

	observation := fmt.Sprintf(
		"repeated tool errors detected for skill %s: error rate %.0f%% (%d/%d records)",
		skillID,
		errorRate*100,
		errorCount,
		len(records),
	)

	// Determine the best evolvable zone in the skill file.
	zoneID := pickEvolvableZone(filepath.Join(filepath.Dir(relPath), ".."))

	id := generateLearningID()

	entry := &evolution.LearningEntry{
		ID:           id,
		SkillID:      skillID,
		ZoneID:       zoneID,
		Observation:  observation,
		Status:       evolution.StatusObservation,
		Observations: 1,
		Created:      time.Now(),
		Updated:      time.Now(),
		Confidence:   0.0,
		Evidence: []evolution.EvidenceEntry{
			{
				SessionID: sessionID,
				Date:      time.Now().UTC().Format("2006-01-02"),
				Context:   fmt.Sprintf("error rate %.0f%% across %d tool calls", errorRate*100, len(records)),
			},
		},
		ProposedChange: &evolution.ProposedChange{
			TargetFile: relPath,
			ZoneID:     zoneID,
			Addition:   fmt.Sprintf("<!-- Observed: %s -->\n", observation),
		},
	}

	entry.Confidence = evolution.CalculateConfidence(entry)
	return entry
}

// pickEvolvableZone reads the skill file and returns the first evolvable zone
// ID found.  Falls back to "best-practices" if none are found.
func pickEvolvableZone(skillFilePath string) string {
	data, err := os.ReadFile(skillFilePath)
	if err != nil {
		return "best-practices"
	}
	zones, _ := merge.ParseEvolvableZones(string(data))
	if len(zones) > 0 {
		return zones[0].ID
	}
	return "best-practices"
}

// deduplicateSummaries is preserved as a helper for future Reflective Write extensions.
// Currently there is no call path, so it is commented out to suppress the linter unused warning.
// To restore: remove the block comment in this file and add a call site.
/*
func deduplicateSummaries(summaries []string) string {
	seen := make(map[string]bool)
	var unique []string
	for _, s := range summaries {
		key := strings.ToLower(strings.TrimSpace(s))
		if !seen[key] {
			seen[key] = true
			unique = append(unique, strings.TrimSpace(s))
		}
	}
	result := strings.Join(unique, "; ")
	if len(result) > 200 {
		result = result[:197] + "..."
	}
	return result
}
*/

// generateLearningID returns a new learning ID in the format LEARN-YYYYMMDD-NNN.
// The NNN component is derived from the current nanosecond to minimise
// collisions in the same second.
func generateLearningID() string {
	now := time.Now().UTC()
	// Use last 3 digits of nanoseconds as a simple collision avoider.
	ns := now.Nanosecond() % 1000
	return fmt.Sprintf("LEARN-%s-%03d", now.Format("20060102"), ns)
}

// PresentPendingProposals returns a formatted summary of learning entries that
// are ready for human review (StatusRule or StatusHighConfidence).
//
// The summary is intended to be included in the SessionStart additionalContext
// field.  At most 3 entries are returned.
func PresentPendingProposals(projectRoot string) string {
	entries, err := evolution.ListLearnings(projectRoot, evolution.LearningFilter{ExcludeArchived: true})
	if err != nil || len(entries) == 0 {
		return ""
	}

	var pending []*evolution.LearningEntry
	for _, e := range entries {
		if e.Status == evolution.StatusRule || e.Status == evolution.StatusHighConfidence {
			pending = append(pending, e)
		}
		if len(pending) >= 3 {
			break
		}
	}

	if len(pending) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Pending Skill Improvement Proposals\n\n")
	sb.WriteString("The following learning proposals are ready for your review:\n\n")

	for i, e := range pending {
		fmt.Fprintf(&sb, "### %d. %s (%s)\n\n", i+1, e.SkillID, e.ID)
		fmt.Fprintf(&sb, "**Observation**: %s\n\n", e.Observation)
		fmt.Fprintf(&sb, "**Observations**: %d | **Confidence**: %.0f%%\n\n",
			e.Observations, e.Confidence*100)
		if e.ProposedChange != nil {
			fmt.Fprintf(&sb, "**Target**: `%s` (zone: `%s`)\n\n",
				e.ProposedChange.TargetFile, e.ProposedChange.ZoneID)
		}
		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// AnalyzeSessionAndLog is a convenience wrapper that calls AnalyzeSession and
// logs the result at the appropriate slog level.  Errors are non-fatal.
func AnalyzeSessionAndLog(projectRoot, sessionID string) {
	proposals, err := AnalyzeSession(projectRoot, sessionID)
	if err != nil {
		slog.Warn("reflective_write: analysis failed",
			"session_id", sessionID,
			"error", err,
		)
		return
	}
	if len(proposals) == 0 {
		return
	}

	slog.Info("reflective_write: proposals generated",
		"session_id", sessionID,
		"count", len(proposals),
		"location", filepath.Join(projectRoot, defs.MoAIDir, defs.EvolutionSubdir, "learnings"),
	)
}
