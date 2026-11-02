package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Card t2011, audit sync-audit-1 defect D3 — the frontmatter fallback reads the
// declaration from the tree the session RUNS in, not from CLAUDE_PROJECT_DIR.
//
// In a worktree session CLAUDE_PROJECT_DIR names the primary checkout. If the
// primary declares one model and the session's own tree declares another, the
// primary's declaration made a response served on the primary's model read `ok`
// when the session's tree expected something else. The audit LOG still lands
// under CLAUDE_PROJECT_DIR (servedAuditRoot): only the declaration lookup moves.

// stopServedRow runs one SubagentStop for plan-auditor served opus, with the
// hook input's cwd set to cwd, and returns the audit row it wrote under the
// primary checkout.
func stopServedRow(t *testing.T, primary, cwd string) map[string]any {
	t.Helper()
	tr := writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), "a9",
		repeatRows(assistantRow("claude-opus-5-5"), 2), map[string]string{"agentType": "plan-auditor"})
	t.Setenv(config.EnvClaudeProjectDir, primary)
	_, err := NewSubagentStopHandlerWithConfig(staticConfigProvider{cfg: &config.Config{}}).Handle(context.Background(), &HookInput{
		CWD: cwd, SessionID: "s-9", AgentID: "a9", AgentType: "plan-auditor",
		AgentTranscriptPath: tr, LastAssistantMessage: "x", HookEventName: string(EventSubagentStop),
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	rows, _ := readServedAuditRows(t, primary)
	if len(rows) == 0 {
		t.Fatalf("no audit row landed under the primary")
	}
	// Subtests share one primary, so the row this stop wrote is the last one.
	return rows[len(rows)-1]
}

func TestServedModel_DeclarationComesFromTheSessionTree(t *testing.T) {
	primary := frontmatterProject(t, "plan-auditor", "model: opus\n")
	wt := frontmatterProject(t, "plan-auditor", "model: sonnet\n")
	sub := filepath.Join(wt, "internal", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	for name, cwd := range map[string]string{"cwd at the tree root": wt, "cwd in a subdirectory": sub} {
		t.Run(name, func(t *testing.T) {
			row := stopServedRow(t, primary, cwd)
			if row["verdict"] != ServedVerdictDrift || row["expected_model"] != "sonnet" {
				t.Fatalf("row verdict=%v expected=%v, want drift against the session tree's sonnet "+
					"(the primary's opus declaration must not be read)", row["verdict"], row["expected_model"])
			}
		})
	}
}

func TestServedModel_SessionTreeWithoutADeclarationFallsBackToTheProjectRoot(t *testing.T) {
	primary := frontmatterProject(t, "plan-auditor", "model: opus\n")
	t.Run("cwd names no tree that carries agent definitions", func(t *testing.T) {
		bare := t.TempDir()
		if row := stopServedRow(t, primary, bare); row["verdict"] != ServedVerdictOK {
			t.Fatalf("verdict = %v, want ok from the project root (the pre-existing behavior)", row["verdict"])
		}
	})
	t.Run("cwd empty", func(t *testing.T) {
		if row := stopServedRow(t, primary, ""); row["verdict"] != ServedVerdictOK {
			t.Fatalf("verdict = %v, want ok from the project root (the pre-existing behavior)", row["verdict"])
		}
	})
	t.Run("cwd is the primary itself", func(t *testing.T) {
		if row := stopServedRow(t, primary, primary); row["verdict"] != ServedVerdictOK {
			t.Fatalf("verdict = %v, want ok (same tree, same declaration)", row["verdict"])
		}
	})
}
