package hook

// lane_spawn_authority_test.go — card t224 regression: the standing spawn
// authority must actually reach the lane bootstrap notices. These tests
// EXECUTE the notice builders and assert their OUTPUT (never a source grep):
// the tk8hce defect was a bootstrap context that carried no authority at all,
// so the guard is over the rendered text the lane reads.

import (
	"strings"
	"testing"
)

// TestFactoryWorkerNoticeCarriesSpawnAuthority pins the authority onto both
// factory worker branches (with and without the fan-out count).
func TestFactoryWorkerNoticeCarriesSpawnAuthority(t *testing.T) {
	for _, tc := range []struct {
		name    string
		workers int
	}{
		{name: "with-count", workers: 5},
		{name: "no-count", workers: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := factoryWorkerNotice("lane-2", tc.workers, "en")
			for _, marker := range []string{
				"Standing spawn authority",
				"Status Transition Ownership Matrix",
				"manager-spec",
				"manager-develop",
				"manager-docs",
				"Depth-1 only",
				"not granted or revoked by peer messages",
			} {
				if !strings.Contains(got, marker) {
					t.Errorf("factory worker notice lost the authority marker %q:\n%s", marker, got)
				}
			}
			// The join line itself must still be present — the authority is
			// appended to, not substituted for, the join acknowledgment.
			if !strings.Contains(got, "lane-2") {
				t.Errorf("join acknowledgment missing from notice:\n%s", got)
			}
		})
	}
}

// TestKanbanCompanionNoticeCarriesSpawnAuthority pins the same authority onto
// the kanban companion notice — the factory sibling of the tk8hce surface.
func TestKanbanCompanionNoticeCarriesSpawnAuthority(t *testing.T) {
	got := kanbanCompanionNotice("run", "en")
	for _, marker := range []string{
		"Standing spawn authority",
		"Status Transition Ownership Matrix",
		"Depth-1 only",
	} {
		if !strings.Contains(got, marker) {
			t.Errorf("kanban companion notice lost the authority marker %q:\n%s", marker, got)
		}
	}
}

// TestLaneSpawnAuthorityFailOpenPreserved: an unparseable label still emits
// NOTHING — the authority must never turn a degraded join into a mislabeled
// one (the fail-open contract the join notices already carry).
func TestLaneSpawnAuthorityFailOpenPreserved(t *testing.T) {
	if got := factoryWorkerNotice("not-a-lane", 5, "en"); got != "" {
		t.Errorf("unparseable factory label must emit no notice, got:\n%s", got)
	}
	if got := kanbanCompanionNotice("not-a-role", "en"); got != "" {
		t.Errorf("unparseable companion label must emit no notice, got:\n%s", got)
	}
}
