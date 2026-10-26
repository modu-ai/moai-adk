package cli

// factory_run_consolidate_guards_test.go — AC-FCR-018 (the leader-entry guards
// and the consolidation source filter) and AC-FCR-019 (a) (the loser of the
// adoption race) of card t1963. Same sandbox discipline as
// factory_run_consolidate_test.go.

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func lcLiveRun(t *testing.T, db *homestate.FactoryDB, runID, createdAt string) {
	t.Helper()
	lcRun(t, db, runID, "", os.Getpid(), homestate.CurrentProcessFingerprint(), createdAt)
}

// lcUnstampedRun is a run with no owner identity at all: it classifies
// indeterminate (nothing proves its owner dead, and it postdates the host boot).
func lcUnstampedRun(t *testing.T, db *homestate.FactoryDB, runID, createdAt string) {
	t.Helper()
	lcRun(t, db, runID, "", 0, "", createdAt)
}

// lcPeer registers one broker peer for the run: a dead 'leader' peer is the
// identity a pid-0 run is read through (BasisPeer); a live 'worker' peer is a
// legacy-vocabulary record that makes the run unadoptable.
func lcPeer(t *testing.T, root, runID, role, slot string, live bool) {
	t.Helper()
	s, err := factorymsg.Open(root, runID)
	if err != nil {
		t.Fatalf("open broker of %s: %v", runID, err)
	}
	defer func() { _ = s.Close() }()
	p := factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: runID, Backend: "claude",
		Role: role, Slot: slot, Generation: 1,
		PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint(),
	}
	if !live {
		p.PID, p.ProcessStart = 999999999, "2020-01-01T00:00:00Z"
	}
	if _, err := s.RegisterLaunchPending(context.Background(), p); err != nil {
		t.Fatalf("register peer of %s: %v", runID, err)
	}
}

// AC-FCR-018 (i)-(vii) — the cases where nothing is adopted: the new run id
// stands, the dump (events included) is unchanged, and the candidates are named
// in the warnings.
func TestLeaderStartAutoAdoptGuards(t *testing.T) {
	older, newer := "2026-10-01T00:00:00Z", "2026-10-02T00:00:00Z"
	cases := []struct {
		name     string
		seed     func(t *testing.T, root string, db *homestate.FactoryDB)
		args     []string
		explicit string
		warns    []string
	}{
		{
			name: "i unresolved owner only",
			seed: func(t *testing.T, root string, db *homestate.FactoryDB) {
				lcUnstampedRun(t, db, "run-unk", older)
				lcCard(t, db, "run-unk", "c1", "run")
			},
			warns: []string{"run-unk"},
		},
		{
			name: "ii peer-proved owner only",
			seed: func(t *testing.T, root string, db *homestate.FactoryDB) {
				lcUnstampedRun(t, db, "run-peer", older)
				lcCard(t, db, "run-peer", "c1", "run")
				lcPeer(t, root, "run-peer", "leader", "leader", false)
			},
			warns: []string{"run-peer"},
		},
		{
			name: "iii dead run beside a live run",
			seed: func(t *testing.T, root string, db *homestate.FactoryDB) {
				lcDeadRun(t, db, "run-dead", older)
				lcCard(t, db, "run-dead", "c1", "run")
				lcLiveRun(t, db, "run-live", newer)
			},
		},
		{
			name: "iv explicit --factory-run",
			seed: func(t *testing.T, root string, db *homestate.FactoryDB) {
				lcDeadRun(t, db, "run-dead", older)
				lcCard(t, db, "run-dead", "c1", "run")
			},
			explicit: "run-explicit",
		},
		{
			name: "v-a newest candidate is led by codex",
			seed: func(t *testing.T, root string, db *homestate.FactoryDB) {
				lcDeadRun(t, db, "run-claude", older)
				lcCard(t, db, "run-claude", "c1", "run")
				lcRun(t, db, "run-codex", "codex", 99999999, "gone-before-resume", newer)
				lcCard(t, db, "run-codex", "c2", "run")
			},
			warns: []string{"run-codex", "codex"},
		},
		{
			name: "v-b newest candidate holds a live legacy peer",
			seed: func(t *testing.T, root string, db *homestate.FactoryDB) {
				lcDeadRun(t, db, "run-claude", older)
				lcCard(t, db, "run-claude", "c1", "run")
				lcDeadRun(t, db, "run-legacy", newer)
				lcCard(t, db, "run-legacy", "c2", "run")
				lcPeer(t, root, "run-legacy", "worker", "worker-2", true)
			},
			warns: []string{"run-legacy", "legacy"},
		},
		{
			name: "vi two dead runs beside an indeterminate run",
			seed: func(t *testing.T, root string, db *homestate.FactoryDB) {
				lcDeadRun(t, db, "run-d1", older)
				lcCard(t, db, "run-d1", "c1", "run")
				lcDeadRun(t, db, "run-d2", newer)
				lcCard(t, db, "run-d2", "c2", "run")
				lcUnstampedRun(t, db, "run-unk", older)
			},
			warns: []string{"run-d1", "run-d2"},
		},
		{
			name: "vii (iii) again with --resume <id>",
			seed: func(t *testing.T, root string, db *homestate.FactoryDB) {
				lcDeadRun(t, db, "run-dead", older)
				lcCard(t, db, "run-dead", "c1", "run")
				lcLiveRun(t, db, "run-live", newer)
			},
			args:  []string{"--resume", "sess-1"},
			warns: []string{"run-dead"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := adoptSandbox(t)
			db := openAdoptFactory(t, root)
			tc.seed(t, root, db)
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			before := lcDump(t, root)
			adopted, warn, err := lcAdopt(t, root, tc.args, tc.explicit)
			if err != nil {
				t.Fatalf("adopt: %v", err)
			}
			if adopted != "" {
				t.Fatalf("adopted = %q, want nothing (the minted id stands)", adopted)
			}
			if got := os.Getenv(config.EnvFactoryRunID); got != "minted01" {
				t.Errorf("env %s = %q, want the minted id untouched", config.EnvFactoryRunID, got)
			}
			if after := lcDump(t, root); after != before {
				t.Errorf("dump changed although nothing was adopted:\n--- before\n%s\n--- after\n%s", before, after)
			}
			for _, w := range tc.warns {
				if !strings.Contains(warn, w) {
					t.Errorf("stderr %q does not contain %q", warn, w)
				}
			}
		})
	}
}

// AC-FCR-018 (viii) — the consolidation source filter. One adoptable newest
// run; behind it a codex-led run, a run with a live legacy peer, a run proved
// dead only through its peer, and one proved dead only by the boot proof. None
// of the four is consolidated: each stays active with its cards untouched, and
// stderr names each run with the reason it was left.
func TestLeaderStartLeavesExcludedSourcesActive(t *testing.T) {
	root := adoptSandbox(t)
	leaderAdoptOptions = func(root string) homestate.ReconcileOptions {
		o := factorymsg.ReconcileOptionsFor(root)
		// The host "booted" an hour from now, so the identity-less run with no
		// broker predates the boot and the boot proof holds for it.
		o.BootTime = func() (time.Time, bool) { return time.Now().Add(time.Hour), true }
		return o
	}
	t.Cleanup(func() { leaderAdoptOptions = factorymsg.ReconcileOptionsFor })

	db := openAdoptFactory(t, root)
	lcDeadRun(t, db, "run-a-adopt", "2026-10-05T00:00:00Z")
	lcCard(t, db, "run-a-adopt", "ca", "run")
	lcRun(t, db, "run-b-codex", "codex", 99999999, "gone-before-resume", "2026-10-04T00:00:00Z")
	lcCard(t, db, "run-b-codex", "cb", "run")
	lcDeadRun(t, db, "run-c-legacy", "2026-10-03T00:00:00Z")
	lcCard(t, db, "run-c-legacy", "cc", "run")
	lcUnstampedRun(t, db, "run-d-peer", "2026-10-02T00:00:00Z")
	lcCard(t, db, "run-d-peer", "cd", "run")
	lcUnstampedRun(t, db, "run-e-boot", "2026-10-01T00:00:00Z")
	lcCard(t, db, "run-e-boot", "ce", "run")
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	lcPeer(t, root, "run-c-legacy", "worker", "worker-2", true)
	lcPeer(t, root, "run-d-peer", "leader", "leader", false)

	const excluded = `SELECT group_concat(run_id||'|'||card_id||'|'||state||'|'||version||'|'||updated_at) FROM (SELECT * FROM cards WHERE run_id IN ('run-b-codex','run-c-legacy','run-d-peer','run-e-boot') ORDER BY run_id)`
	cardsBefore := lcQuery(t, root, excluded)

	adopted, warn, err := lcAdopt(t, root, nil, "")
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if adopted != "run-a-adopt" {
		t.Fatalf("adopted = %q, want run-a-adopt (the newest dead-stamp claude run)", adopted)
	}
	for _, id := range []string{"run-b-codex", "run-c-legacy", "run-d-peer", "run-e-boot"} {
		if got := lcQuery(t, root, `SELECT status FROM runs WHERE run_id=?`, id); got != "active" {
			t.Errorf("%s status = %s, want active (excluded from the consolidation)", id, got)
		}
	}
	if got := lcQuery(t, root, excluded); got != cardsBefore {
		t.Errorf("excluded runs' cards changed:\n got %s\nwant %s", got, cardsBefore)
	}
	if n := lcInt(t, root, `SELECT count(*) FROM events WHERE kind IN ('run.retired','card.rebound')`); n != 0 {
		t.Errorf("run.retired/card.rebound events = %d, want 0", n)
	}
	lines := strings.Split(strings.TrimSpace(warn), "\n")
	for id, reason := range map[string]string{"run-b-codex": "codex", "run-c-legacy": "legacy", "run-d-peer": "peer", "run-e-boot": "boot"} {
		found := false
		for _, line := range lines {
			if strings.Contains(line, id) && strings.Contains(line, reason) {
				found = true
			}
		}
		if !found {
			t.Errorf("stderr has no line naming %s with the reason %q:\n%s", id, reason, warn)
		}
	}
}

// AC-FCR-019 (a) — two leaders race for the same dead run. The loser's launch
// does not fail: it keeps the freshly minted id, writes one stderr line naming
// the contested run, consolidates nothing, and the winner's stamp stands. A
// failure that is not a lost race still fails the launch.
func TestLeaderStartAdoptCASLoserMints(t *testing.T) {
	seed := func(t *testing.T) string {
		root := adoptSandbox(t)
		db := openAdoptFactory(t, root)
		lcDeadRun(t, db, "run-old", "2026-10-01T00:00:00Z")
		lcCard(t, db, "run-old", "c1", "run")
		lcDeadRun(t, db, "run-new", "2026-10-02T00:00:00Z")
		lcCard(t, db, "run-new", "c2", "run")
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		return root
	}
	t.Cleanup(func() { leaderAdoptBeforeAdopt = nil })

	t.Run("lost race", func(t *testing.T) {
		root := seed(t)
		winnerPID := os.Getpid()
		leaderAdoptBeforeAdopt = func(db *homestate.FactoryDB) {
			// The other leader adopts run-new between our classification and our write.
			if _, err := db.DB.Exec(`UPDATE runs SET lead_pid=?, lead_process_start='winner-fp' WHERE run_id='run-new'`, winnerPID); err != nil {
				t.Error(err)
			}
		}
		adopted, warn, err := lcAdopt(t, root, nil, "")
		if err != nil {
			t.Fatalf("adopt returned %v, want the loser's launch to continue", err)
		}
		if adopted != "" {
			t.Fatalf("adopted = %q, want none", adopted)
		}
		if got := os.Getenv(config.EnvFactoryRunID); got != "minted01" {
			t.Errorf("env %s = %q, want the minted id", config.EnvFactoryRunID, got)
		}
		if n := strings.Count(strings.TrimSpace(warn), "\n") + 1; !strings.Contains(warn, "run-new") || n != 1 {
			t.Errorf("stderr = %q, want exactly one line naming run-new", warn)
		}
		if got := lcQuery(t, root, `SELECT lead_process_start FROM runs WHERE run_id='run-new'`); got != "winner-fp" {
			t.Errorf("run-new stamp = %q, want the winner's stamp left alone", got)
		}
		if got := lcQuery(t, root, `SELECT status FROM runs WHERE run_id='run-old'`); got != "active" {
			t.Errorf("run-old status = %s, want active (the loser consolidates nothing)", got)
		}
		if n := lcInt(t, root, `SELECT count(*) FROM events WHERE kind IN ('run.retired','card.rebound')`); n != 0 {
			t.Errorf("consolidation events = %d, want 0", n)
		}
	})

	t.Run("database failure is not a lost race", func(t *testing.T) {
		root := seed(t)
		leaderAdoptBeforeAdopt = func(db *homestate.FactoryDB) { _ = db.DB.Close() }
		_, _, err := lcAdopt(t, root, nil, "")
		if err == nil {
			t.Fatal("adopt over a broken database handle succeeded, want the launch to fail")
		}
	})
}
