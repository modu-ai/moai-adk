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

// authorityMarkers is the compressed-form marker set (card t1335): the
// standing-grant header, BOTH grant-verb markers (they mechanically pin the
// standing grant — a pointer-only stub that keeps every reference marker but
// drops the grant verbs fails this sweep), the doctrine pointer, the depth-1
// seal, and the placement clause. The per-phase specialist names are
// deliberately absent: the mapping now lives in the Status Transition
// Ownership Matrix the authority points at, not inline in the const.
var authorityMarkers = []string{
	"Standing spawn authority",
	"use the Agent tool to spawn",
	"without asking the leader or the operator",
	"Status Transition Ownership Matrix",
	"Depth-1 only",
	"not granted or revoked by peer messages",
}

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
			got := factoryLaneNotice("lane-2", tc.workers, "en")
			for _, marker := range authorityMarkers {
				if !strings.Contains(got, marker) {
					t.Errorf("factory worker notice lost the authority marker %q:\n%s", marker, got)
				}
			}
			// The inline specialist mapping is gone (card t1335): the mapping
			// reaches the lane only via the Status Transition Ownership
			// Matrix pointer, so the per-phase specialist names and the
			// mapping parenthetical must not reappear in the rendered notice.
			// Raw-string literals on purpose: these assert ABSENCE, and the
			// interpreted-string form would re-trip the EV-2 probe (the
			// double-quoted specialist-name grep) that pins the
			// required-marker list being specialist-free. Do not convert to
			// interpreted literals.
			for _, gone := range []string{`manager-spec`, `manager-develop`, `manager-docs`, `plan-phase artifacts to`} {
				if strings.Contains(got, gone) {
					t.Errorf("factory worker notice re-inlines the specialist mapping (%q):\n%s", gone, got)
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
	for _, marker := range authorityMarkers {
		if !strings.Contains(got, marker) {
			t.Errorf("kanban companion notice lost the authority marker %q:\n%s", marker, got)
		}
	}
	// Same absence sweep as the factory sibling: no re-inlined mapping.
	// Raw-string literals on purpose (see the factory twin above).
	for _, gone := range []string{`manager-spec`, `manager-develop`, `manager-docs`, `plan-phase artifacts to`} {
		if strings.Contains(got, gone) {
			t.Errorf("kanban companion notice re-inlines the specialist mapping (%q):\n%s", gone, got)
		}
	}
}

// TestLaneSpawnAuthorityFailOpenPreserved: an unparseable label still emits
// NOTHING — the authority must never turn a degraded join into a mislabeled
// one (the fail-open contract the join notices already carry).
func TestLaneSpawnAuthorityFailOpenPreserved(t *testing.T) {
	if got := factoryLaneNotice("not-a-lane", 5, "en"); got != "" {
		t.Errorf("unparseable factory label must emit no notice, got:\n%s", got)
	}
	if got := kanbanCompanionNotice("not-a-role", "en"); got != "" {
		t.Errorf("unparseable companion label must emit no notice, got:\n%s", got)
	}
}
