package cli

import (
	"fmt"
	"testing"
)

// Issue #1948 §8, fix 5: today's shape — one live run and five dead-owner runs,
// each holding an unfinished card and a dead worker row. A lane join resolves
// to the live run without opening the chooser, and the dead runs end retired
// with their cards under the live run.
func TestLaneJoinZombieShapeJoinsTheLiveRun(t *testing.T) {
	root := laneGateRoot(t)
	seedLiveRun(t, root, "rx-live")
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("rz-dead-%d", i)
		selSeedDead(t, root, id, false, 1)
		selExec(t, root, `INSERT INTO workers(label,pid,run_id,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`,
			"w-"+id, selDeadPID, id, "2026-10-18T00:00:00Z", "2026-10-18T00:00:00Z")
	}
	spy := newSpy(root)
	installLaneSeams(t, spy)

	restore, err := enterFactoryLaneRun(root, "", "", nil)
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	defer restore()
	requireEnv(t, "rx-live")
	if len(spy.entries) != 0 {
		t.Fatalf("the chooser opened %d times, want 0 (one live run joins without a prompt)", len(spy.entries))
	}
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("rz-dead-%d", i)
		if got := selRunStatus(t, root, id); got != "retired" {
			t.Errorf("%s status = %q, want retired", id, got)
		}
		if n := selQueryInt(t, root, `SELECT COUNT(*) FROM cards WHERE run_id='rx-live' AND card_id=?`, fmt.Sprintf("card-%s-0", id)); n != 1 {
			t.Errorf("card of %s is not under rx-live", id)
		}
	}
	if n := selQueryInt(t, root, `SELECT COUNT(*) FROM workers WHERE pid=?`, selDeadPID); n != 0 {
		t.Errorf("%d dead worker rows remain", n)
	}
}
