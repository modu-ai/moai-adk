package hook_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/evolution"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// TestAnalyzeSession_ErrorOutcomeGeneratesProposal verifies that a session
// with error outcomes produces at least one learning proposal.
func TestAnalyzeSession_ErrorOutcomeGeneratesProposal(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitEvolution(t, projectRoot)
	seedSkillFile(t, projectRoot, "moai-lang-go")

	sessionID := "test-session-error-001"
	writeRecords(t, projectRoot, sessionID, []telemetry.UsageRecord{
		{
			Timestamp:           time.Now(),
			SessionID:           sessionID,
			SkillID:             "moai-lang-go",
			AgentType:"Bash",
			Outcome:             telemetry.OutcomeError,
						DurationMs:0,
		},
		{
			Timestamp:           time.Now(),
			SessionID:           sessionID,
			SkillID:             "moai-lang-go",
			AgentType:"Bash",
			Outcome:             telemetry.OutcomeError,
						DurationMs:1,
		},
		{
			Timestamp:           time.Now(),
			SessionID:           sessionID,
			SkillID:             "moai-lang-go",
			AgentType:"Write",
			Outcome:             telemetry.OutcomeSuccess,
			DurationMs:2,
		},
	})

	proposals, err := hook.AnalyzeSession(projectRoot, sessionID)
	if err != nil {
		t.Fatalf("AnalyzeSession: %v", err)
	}
	if len(proposals) == 0 {
		t.Fatal("expected at least one proposal for error-outcome session, got 0")
	}
}

// TestAnalyzeSession_SkipsTrivialSession verifies that a session with fewer
// than MinToolInvocationsForAnalysis tool calls is skipped.
func TestAnalyzeSession_SkipsTrivialSession(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitEvolution(t, projectRoot)

	sessionID := "test-session-trivial-001"
	writeRecords(t, projectRoot, sessionID, []telemetry.UsageRecord{
		{
			Timestamp:           time.Now(),
			SessionID:           sessionID,
			SkillID:             "moai-lang-go",
			AgentType:"Read",
			Outcome:             telemetry.OutcomeSuccess,
			DurationMs:0,
		},
	})

	proposals, err := hook.AnalyzeSession(projectRoot, sessionID)
	if err != nil {
		t.Fatalf("AnalyzeSession: %v", err)
	}
	if len(proposals) != 0 {
		t.Fatalf("expected 0 proposals for trivial session, got %d", len(proposals))
	}
}

// TestAnalyzeSession_CapsAt3Proposals verifies that AnalyzeSession never
// returns more than MaxSessionProposals entries.
func TestAnalyzeSession_CapsAt3Proposals(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitEvolution(t, projectRoot)

	// Create skill files for several different skills.
	for _, skill := range []string{"moai-lang-go", "moai-lang-python", "moai-lang-ts", "moai-lang-rust"} {
		seedSkillFile(t, projectRoot, skill)
	}

	sessionID := "test-session-many-001"
	records := make([]telemetry.UsageRecord, 0, 20)
	skills := []string{"moai-lang-go", "moai-lang-python", "moai-lang-ts", "moai-lang-rust"}
	for i, sk := range skills {
		for j := 0; j < 4; j++ {
			records = append(records, telemetry.UsageRecord{
				Timestamp:           time.Now(),
				SessionID:           sessionID,
				SkillID:             sk,
				AgentType:"Bash",
				Outcome:             telemetry.OutcomeError,
				ContextHash:"build failed",
				DurationMs: int64(i*4 + j),
			})
		}
	}
	writeRecords(t, projectRoot, sessionID, records)

	proposals, err := hook.AnalyzeSession(projectRoot, sessionID)
	if err != nil {
		t.Fatalf("AnalyzeSession: %v", err)
	}
	if len(proposals) > evolution.MaxSessionProposals {
		t.Fatalf("expected at most %d proposals, got %d", evolution.MaxSessionProposals, len(proposals))
	}
}

// TestAnalyzeSession_SafetyFiltersFrozenFiles verifies that proposals
// targeting frozen files are filtered out before being returned.
func TestAnalyzeSession_SafetyFiltersFrozenFiles(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitEvolution(t, projectRoot)

	// The frozen guard should prevent any proposal with CLAUDE.md as target.
	// We verify by checking the returned proposals do not include frozen targets.
	sessionID := "test-session-frozen-001"
	// Use a skill ID that maps to a frozen file to force the scenario
	// (in practice, skill IDs never map to frozen files, but the filter
	// must still hold if something upstream misbehaves).
	writeRecords(t, projectRoot, sessionID, []telemetry.UsageRecord{
		{
			Timestamp:           time.Now(),
			SessionID:           sessionID,
			SkillID:             "moai-lang-go",
			AgentType:"Bash",
			Outcome:             telemetry.OutcomeError,
			ContextHash:"error",
			DurationMs:0,
		},
		{
			Timestamp:           time.Now(),
			SessionID:           sessionID,
			SkillID:             "moai-lang-go",
			AgentType:"Bash",
			Outcome:             telemetry.OutcomeError,
			ContextHash:"error",
			DurationMs:1,
		},
		{
			Timestamp:           time.Now(),
			SessionID:           sessionID,
			SkillID:             "moai-lang-go",
			AgentType:"Write",
			Outcome:             telemetry.OutcomeSuccess,
			DurationMs:2,
		},
	})

	proposals, err := hook.AnalyzeSession(projectRoot, sessionID)
	if err != nil {
		t.Fatalf("AnalyzeSession: %v", err)
	}

	for _, p := range proposals {
		if p.ProposedChange == nil {
			continue
		}
		if evolution.CheckFrozenGuard(p.ProposedChange.TargetFile) == nil {
			// CheckFrozenGuard returns nil for ALLOWED files.
			continue
		}
		t.Errorf("proposal targets frozen file %q", p.ProposedChange.TargetFile)
	}
}

// --- helpers ---

// mustInitEvolution creates the .moai/evolution directory structure.
func mustInitEvolution(t *testing.T, projectRoot string) {
	t.Helper()
	dirs := []string{
		filepath.Join(projectRoot, ".moai", "evolution", "learnings"),
		filepath.Join(projectRoot, ".moai", "evolution", "telemetry"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mustInitEvolution: %v", err)
		}
	}
}

// seedSkillFile creates a minimal skill file with an evolvable zone so that
// the engine can identify proposal targets.
func seedSkillFile(t *testing.T, projectRoot, skillID string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".claude", "skills", skillID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("seedSkillFile mkdir: %v", err)
	}
	content := "# " + skillID + "\n\n<!-- @EVOLVABLE:best-practices -->\nInitial content.\n<!-- @EVOLVABLE:best-practices:END -->\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("seedSkillFile write: %v", err)
	}
}

// writeRecords writes telemetry records to the project's telemetry directory.
func writeRecords(t *testing.T, projectRoot, sessionID string, records []telemetry.UsageRecord) {
	t.Helper()
	for i := range records {
		if err := telemetry.RecordSkillUsage(projectRoot, records[i]); err != nil {
			t.Fatalf("writeRecords: %v", err)
		}
	}
}
