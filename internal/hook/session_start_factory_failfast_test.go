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

// TestSessionStartCodexLaneFailFastRidesContext wires the stop line through
// the SessionStart handler: the broken-launch codex session reads it in
// additionalContext, and the owned-card rule is withheld.
func TestSessionStartCodexLaneFailFastRidesContext(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvFactoryBackend, "gpt")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvFactoryRunID, "")
	t.Setenv(config.EnvFactoryCard, "t1776")

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	h := NewSessionStartHandler(nil)
	out, err := h.Handle(t.Context(), &HookInput{
		SessionID:  "codex-failfast",
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
	if !strings.Contains(ctx, "MOAI_FACTORY_WORKER") {
		t.Fatalf("broken-launch session got no fail-fast line in additionalContext:\n%s", ctx)
	}
	if strings.Contains(ctx, "owned card") {
		t.Fatalf("the owned-card rule must be withheld on an incomplete launch:\n%s", ctx)
	}
}
