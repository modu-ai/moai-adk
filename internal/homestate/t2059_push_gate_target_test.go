// t2059_push_gate_target_test.go — card t2059 (root cause): the push gate's
// default destination is not decided by the repository's remote.
package homestate

import (
	"context"
	"testing"
)

// TestT2059_PushGateTargetIgnoresTheRemote — PushGateTarget used to answer
// pushed for a repository with a remote and done for one without, the old
// premise that a remote means push and CI come before done. The operator's
// Done standard (local verification + local main merge + the leader's
// receipt) makes done the default whatever the remote.
func TestT2059_PushGateTargetIgnoresTheRemote(t *testing.T) {
	for _, tc := range []struct {
		name       string
		withRemote bool
	}{{"with a remote", true}, {"without a remote", false}} {
		repo := frNewRepo(t, tc.withRemote)
		got, err := PushGateTarget(context.Background(), Card{WorktreePath: repo.Dir})
		if err != nil || got != CardDone {
			t.Errorf("%s: target = %q err=%v, want done", tc.name, got, err)
		}
	}
}
