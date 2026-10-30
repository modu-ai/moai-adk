package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The lease holder must outlive the command. `moai af claim` is run from a shell wrapper (the Bash tool starts a
// fresh `sh -c` per command): its parent is that wrapper, which exits as soon as the command returns, so a lease
// that names it reads as a dead holder at the next reap and the card is taken back from a live session.
func holderSeams(t *testing.T, parent int, owner int, resolved bool) {
	t.Helper()
	origParent, origOwner, origProbe := afParentPID, afOwnerPID, afProbe
	t.Cleanup(func() { afParentPID, afOwnerPID, afProbe = origParent, origOwner, origProbe })
	afParentPID = func() int { return parent }
	afOwnerPID = func() (int, bool) { return owner, resolved }
	afProbe = func(pid int) (string, homestate.ProcessIdentityState) {
		return "start-of-" + string(rune('0'+pid%10)), homestate.ProcessIdentityLive
	}
}

func TestAFHolder_SkipsTheShortLivedWrapper(t *testing.T) {
	holderSeams(t, 9001 /* the sh -c wrapper */, 4242 /* the session */, true)
	pid, start := afHolder()
	if pid != 4242 {
		t.Fatalf("holder pid = %d, want the session's 4242 (the wrapper's 9001 dies with the command)", pid)
	}
	if start == "" {
		t.Fatal("the session's start fingerprint must be recorded with its pid")
	}
}

// When the session cannot be resolved the holder is recorded as unknown (pid 0): an unknown holder is treated as
// alive and never reclaimed (I5), where the dead-on-arrival wrapper pid would be reclaimed at once.
func TestAFHolder_UnresolvedSessionIsUnknownNotTheWrapper(t *testing.T) {
	holderSeams(t, 9001, 0, false)
	pid, start := afHolder()
	if pid != 0 || start != "" {
		t.Fatalf("holder = (%d, %q), want (0, \"\"): the wrapper must not be recorded", pid, start)
	}
}
