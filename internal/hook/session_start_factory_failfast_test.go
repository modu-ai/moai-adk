package hook

// session_start_factory_failfast_test.go — card t1776: a session stamped with
// the Codex factory backend (MOAI_FACTORY_BACKEND=gpt) but missing its lane
// stamps (MOAI_FACTORY_WORKER / MOAI_KANBAN_ID — the launcher writes them
// together, so a missing one is a broken launch) gets ONE agent-facing stop
// line on SessionStart. This is the operator-measured failure's fail-fast:
// the unstamped session inferred lane behavior from source reading and burned
// the dispatch context; the hook must say STOP instead of staying silent.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestCodexLaneEnvFailFast(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvFactoryBackend, "gpt")

	t.Run("stamps_complete_no_notice", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-7")
		t.Setenv(config.EnvFactoryRunID, "run-abc")
		if got := codexLaneEnvFailFast(); got != "" {
			t.Errorf("a fully stamped codex lane must not trip the fail-fast, got: %s", got)
		}
	})

	t.Run("worker_missing_one_line_stop", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorker, "")
		t.Setenv(config.EnvFactoryRunID, "run-abc")
		got := codexLaneEnvFailFast()
		if got == "" {
			t.Fatal("missing worker stamp must produce the one-line stop")
		}
		if strings.Count(got, "\n") > 1 {
			t.Errorf("the stop must be ONE line, got %d newlines: %s", strings.Count(got, "\n"), got)
		}
		if !strings.Contains(got, config.EnvMoaiFactoryWorker) || !strings.Contains(got, config.EnvFactoryRunID) {
			t.Errorf("stop line does not name the missing stamps: %s", got)
		}
		if !strings.Contains(got, "moai codex -l") {
			t.Errorf("stop line does not name the relaunch door: %s", got)
		}
		if !strings.Contains(got, "stop") || !strings.Contains(got, "Do not read") {
			t.Errorf("stop line does not forbid the infer-from-source behavior: %s", got)
		}
	})

	t.Run("run_id_missing_one_line_stop", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-7")
		t.Setenv(config.EnvFactoryRunID, "")
		if got := codexLaneEnvFailFast(); got == "" {
			t.Fatal("missing run-id stamp must produce the one-line stop")
		}
	})

	t.Run("non_gpt_backend_silent", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorker, "")
		t.Setenv(config.EnvFactoryRunID, "")
		t.Setenv(config.EnvFactoryBackend, "claude")
		if got := codexLaneEnvFailFast(); got != "" {
			t.Errorf("non-codex backends stay silent, got: %s", got)
		}
	})
}

// TestSessionStartCodexLaneFailFastSuppressesRoleRules wires the round-3
// precedence fix: on a broken codex lane launch the fail-fast line is the
// session's LAST word — the role-rule injection (whose core carries the
// read-the-full-rule directive that conflicts with the boot directive) is
// suppressed behind it.
func TestSessionStartCodexLaneFailFastSuppressesRoleRules(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvFactoryBackend, "gpt")
	// The worker stamp is PRESENT and the run id missing — the fail-fast
	// arm that leaves the worker in place, so the role-marker registry
	// (EnvMoaiFactoryWorker) still detects the role and the role-rule
	// injection WOULD fire without the suppression. A worker-empty fixture
	// would pass vacuously (no role detected, nothing to suppress).
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-7")
	t.Setenv(config.EnvFactoryRunID, "")
	t.Setenv(config.EnvFactoryCard, "t1776")

	root := t.TempDir()
	writeDeployedRoleRules(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	h := NewSessionStartHandler(nil)
	out, err := h.Handle(t.Context(), &HookInput{
		SessionID:  "codex-failfast-role-rules",
		ProjectDir: root,
		CWD:        root,
		Source:     "startup",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	ctx := ""
	if out.HookSpecificOutput != nil {
		ctx = out.HookSpecificOutput.AdditionalContext
	}
	if !strings.Contains(ctx, "MOAI factory lane launch is incomplete") {
		t.Fatalf("fail-fast line missing:\n%s", ctx)
	}
	if strings.Contains(ctx, "Before any work, read") {
		t.Fatalf("the role core's read-first directive must not ride behind the fail-fast:\n%s", ctx)
	}
}

// TestRoleCoreCarriesNoQueueStart pins the round-4/6 shape: the role core
// carries NO queue-start verb (it reaches card-assigned sessions via
// startup/clear/compact), and it carries NO carve-out either — the
// TestRoleInjectionAssemblyBudget cap leaves the core 7 UTF-16 units of
// headroom (audit round-6 P1), so the Codex precedence resolution lives in
// the hook context (codexLanePrecedenceNote), not in the core text.
func TestRoleCoreCarriesNoQueueStart(t *testing.T) {
	for _, rel := range []string{
		filepath.Join("..", "..", "internal", "template", "templates", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
		filepath.Join("..", "..", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
	} {
		data, err := os.ReadFile(rel)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		content := string(data)
		start := strings.Index(content, config.RoleCoreMarkerStart)
		end := strings.Index(content, config.RoleCoreMarkerEnd)
		if start < 0 || end < 0 || end < start {
			t.Fatalf("%s: role-core region not found", rel)
		}
		core := content[start:end]
		if strings.Contains(core, "todo --auto") {
			t.Errorf("%s: the role core carries a queue-start command that reaches card-assigned sessions:\n%s", rel, core)
		}
		if strings.Contains(core, "Codex") {
			t.Errorf("%s: the role core grew a Codex carve-out past the injection budget — the precedence lives in codexLanePrecedenceNote", rel)
		}
	}
}

// TestFactoryLaneRuleGptCarriesPrecedence pins the healthy codex lane's
// precedence note: the owned-card rule resolves the role core's read-first
// duty toward the leased card's work (audit round-3), with no queue-start
// directive (audit round-4 — an assigned session stays in its card worktree).
func TestFactoryLaneRuleGptCarriesPrecedence(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvFactoryBackend, "gpt")
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-7")
	t.Setenv(config.EnvFactoryCard, "t1776")

	rule := factoryLaneRuleForSource("startup", "en")
	if rule == "" {
		t.Fatal("a stamped gpt lane receives the owned-card rule")
	}
	if !strings.Contains(rule, "Codex auto-lane precedence") {
		t.Fatalf("owned-card rule does not carry the boot-directive precedence note:\n%s", rule)
	}
	if strings.Contains(rule, "moai todo --auto") {
		t.Fatalf("owned-card rule must not direct a card-assigned session at the queue:\n%s", rule)
	}
}
