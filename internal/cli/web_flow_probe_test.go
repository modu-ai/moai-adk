package cli

// Defect answer: a board built without the real process probe reads every lease holder as Unknown, so a dead
// lane never shows as stalled. These tests catch runWeb handing web.Run no probe, and a probe that cannot tell a
// live holder (or a reused pid) from an unknown one.

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/web"
)

func TestRunWeb_InjectsTheRealFlowProbe(t *testing.T) {
	var got web.Config
	root := t.TempDir()
	runWebHarness(t, materializeSeams(t, root), withRootDir(root), func(c *runWebHarnessCfg) { c.captured = &got })
	if got.FlowProbe == nil {
		t.Fatal("runWeb gave web.Run no FlowProbe: every holder would read Unknown")
	}
}

func TestWebFlowProbe_ALiveHolderIsNotUnknown(t *testing.T) {
	probe := webFlowProbe()
	start, state := homestate.ProbeProcessIdentity(os.Getpid())
	if state != homestate.ProcessIdentityLive {
		t.Skipf("this platform cannot fingerprint the test process (state %v)", state)
	}
	if got := probe(os.Getpid(), start); got != ledger.LivenessAlive {
		t.Errorf("the running test process read %v, want Alive", got)
	}
	// Control: the same pid with a start fingerprint it never had is a reused pid, which reads Dead, so the
	// probe tells holders apart rather than answering one value for all.
	if got := probe(os.Getpid(), start+"-other"); got != ledger.LivenessDead {
		t.Errorf("a reused pid read %v, want Dead", got)
	}
}
