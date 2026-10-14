package cli

import (
	"os/exec"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/discovery"
)

// t1628 review P2 — the codex child session inherits CLAUDE_PROJECT_DIR. Its
// own start prompt runs moai todo --auto, whose precondition reads that anchor
// before the child's working directory; a stale card-tree anchor would refuse
// it. The lane launch must export the root it selected as the child's anchor.
func TestCodexLaneChildAnchorFollowsLaunchRoot(t *testing.T) {
	primary, card := t1628LaneFixture(t)
	t.Setenv("MOAI_HOME", t.TempDir())
	clearFactoryTestEnv(t)
	t.Setenv(config.EnvClaudeProjectDir, card)
	t1628Enter(t, primary)
	t1628StubCodex(t)
	stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("runlead01")})

	var launched *exec.Cmd
	origLaunch := codexDirectLaunchFn
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		launched = c
		return nil
	}
	t.Cleanup(func() { codexDirectLaunchFn = origLaunch })

	if err := t1628RunLane(t); err != nil {
		t.Fatalf("codex -l with one verified leader: %v", err)
	}
	if launched == nil {
		t.Fatal("the lane launch never reached the child launch seam")
	}
	got := t1628EnvMap(launched.Env)[config.EnvClaudeProjectDir]
	if !t1628SameRealPath(got, primary) {
		t.Errorf("the child's %s = %q, want the launch root %q (a stale card anchor refuses the child's moai todo --auto)", config.EnvClaudeProjectDir, got, primary)
	}
}
