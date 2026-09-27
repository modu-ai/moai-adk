//go:build !windows

package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

func TestClaudePOSIXExecFailureRollsBackPending(t *testing.T) {
	home, root, run := t.TempDir(), t.TempDir(), "run-claude-exec-failure"
	t.Setenv("MOAI_HOME", home)
	t.Setenv(config.EnvClaudeProjectDir, root)
	if err := recordFactoryRunStart(root, run, "claude", ""); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(),
		config.EnvMoaiKanbanID+"="+run,
		config.EnvMoaiKanbanBackend+"=claude",
		config.EnvMoaiFactoryWorkers+"=1",
		config.EnvMoaiFactoryWorker+"=",
		config.EnvClaudeProjectDir+"="+root,
	)
	missing := filepath.Join(root, "missing-claude")
	if err := execOrSpawnClaude(missing, []string{missing}, env); err == nil {
		t.Fatal("exec failure returned nil")
	}
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "factory message broker", s)
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Lanes) != 0 {
		t.Fatalf("exec failure left pending rows: %+v", status.Lanes)
	}
}
