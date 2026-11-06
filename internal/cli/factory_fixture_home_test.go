package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox — card t2047. fcFixture
// promises that the card landing directory resolves in a sandbox, never the
// real ~/.moai. todoFixture clears MOAI_HOME so queue state stays project-local,
// and with MOAI_HOME empty paths.MoaiHome falls back to $HOME/.moai; without a
// HOME sandbox every factory test then created its project folder under the
// real ~/.moai/worktrees/projects (21.8k leaked folders measured on 2026-11-05),
// and every later worktreeroot.Root call scanned that directory.
func TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox(t *testing.T) {
	fcFixture(t)

	home, err := paths.MoaiHome()
	if err != nil {
		t.Fatalf("paths.MoaiHome: %v", err)
	}
	tmp, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Dir(home))
	if err != nil {
		// The sandbox .moai may not exist yet; its parent (the sandbox HOME) does.
		t.Fatalf("resolve sandbox home %q: %v", filepath.Dir(home), err)
	}
	if !strings.HasPrefix(resolved+string(filepath.Separator), tmp+string(filepath.Separator)) {
		t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real ~/.moai", home, tmp)
	}
}
