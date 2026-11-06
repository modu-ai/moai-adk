package intake

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
)

// t2078 (follow-up of t2072): register sent RegisterBranch with nothing recorded
// beforehand. A process that dies after the branch was registered but before the
// request row says "unknown" restarts at StageIssueConfirmed with no Blocked row
// and registers the branch a second time.
func TestAProcessDyingDuringRegisterDoesNotRegisterTwice(t *testing.T) {
	r := newRig(t, Config{})
	advanceTo(t, r, "crash-reg", claim.StageIssueConfirmed)

	r.br.mu.Lock()
	r.br.crash = true
	r.br.mu.Unlock()
	func() {
		defer func() { _ = recover() }() // the process died here
		_, _ = r.in.Advance(ctx(), "crash-reg")
	}()
	r.br.mu.Lock()
	r.br.crash = false
	calls := r.br.calls
	r.br.mu.Unlock()
	if calls != 1 {
		t.Fatalf("setup: %d RegisterBranch calls before the restart, want 1", calls)
	}

	// restart
	got, err := r.in.Advance(ctx(), "crash-reg")
	if err != nil {
		t.Fatalf("restart: %v", err)
	}
	r.br.mu.Lock()
	calls = r.br.calls
	r.br.mu.Unlock()
	if calls != 1 {
		t.Errorf("RegisterBranch ran %d times for one request, want 1", calls)
	}
	if got.Stage != claim.StageBranchRegistered {
		t.Errorf("the restart should resolve through the lookup: stage %s", got.Stage)
	}
}
