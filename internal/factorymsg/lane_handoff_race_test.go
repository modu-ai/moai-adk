package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// endpointRow is the byte-level snapshot of one peers row that the race
// criteria compare before and after a racer returns.
type endpointRow struct {
	Session, ProcessStart, UpdatedAt string
	Generation                       int64
	PID                              int
}

func readEndpointRow(t *testing.T, db *sql.DB, slot string) endpointRow {
	t.Helper()
	var r endpointRow
	if err := db.QueryRow(`SELECT session_uuid,generation,pid,process_start,updated_at FROM peers WHERE slot=?`, slot).
		Scan(&r.Session, &r.Generation, &r.PID, &r.ProcessStart, &r.UpdatedAt); err != nil {
		t.Fatalf("read endpoint row %q: %v", slot, err)
	}
	return r
}

func countLaneHandoffs(t *testing.T, db *sql.DB, slot string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM lane_handoffs WHERE slot=?`, slot).Scan(&n); err != nil {
		t.Fatalf("count lane handoffs: %v", err)
	}
	return n
}

// requireDistinctRacerHandles is the separate-handle guard of AC-FLH-018/019:
// two racers sharing one Store (and so one SetMaxOpenConns(1) pool) serialize
// in the Go pool and never reach the SQLite BEGIN IMMEDIATE boundary.
func requireDistinctRacerHandles(a, b *Store) error {
	if a == nil || b == nil {
		return errors.New("racer handle is nil")
	}
	if a == b {
		return errors.New("racers share one *Store")
	}
	if a.db == b.db {
		return errors.New("racers share one *sql.DB")
	}
	return nil
}

// TestFactoryLaneHandoffRebindVsUserPromptRegisterRace is the AC-FLH-019
// named test. Run-phase M1 implements order (vii) only; orders (i)-(vi)
// depend on the M3 rebind and RegisterPeer transaction changes and are added
// by M3. Until then a green run of this selector is NOT an AC-FLH-019 PASS.
func TestFactoryLaneHandoffRebindVsUserPromptRegisterRace(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Log("AC_FLH_019_ORDERS_COVERED=vii")

	t.Run("order_vii_reservation_reads_source_inside_its_transaction", func(t *testing.T) {
		ctx := context.Background()
		root := filepath.Join(t.TempDir(), "project")
		run := "run-handoff-vii"
		const slot = "lane-1"

		seed, err := Open(root, run)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = seed.Close() })

		// Not-current source variant: seeded through the production launcher
		// path with a fake process-start, so the owner is never current.
		pending, err := seed.RegisterLaunchPending(ctx, Peer{
			ProjectKey: "project", RunID: run, Backend: "codex", Role: "worker", Slot: slot,
			PID: os.Getpid(), ProcessStart: "fake-start-vii",
		})
		if err != nil {
			t.Fatal(err)
		}
		bound := pending
		bound.SessionUUID = "src-uuid"
		source, ok, err := seed.BindLaunchPending(ctx, bound)
		if err != nil || !ok {
			t.Fatalf("seed bind ok=%v err=%v", ok, err)
		}

		aStore, err := Open(root, run)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = aStore.Close() })
		vStore, err := Open(root, run)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = vStore.Close() })

		// Guard probe: the distinctness guard must go red on a shared pair
		// before it is trusted on the real racers.
		if err := requireDistinctRacerHandles(aStore, aStore); err == nil {
			t.Fatal("distinct-handle guard accepted a shared pair")
		}
		if err := requireDistinctRacerHandles(aStore, vStore); err != nil {
			t.Fatalf("racer handles not distinct: %v", err)
		}
		t.Log("RACER_HANDLES_DISTINCT=2")

		// Adversary A (acceptance-allowed fixture, because the REQ-FLH-017
		// finalize step inside RegisterPeer is an M3 item): a raw BEGIN
		// IMMEDIATE on A's own production-Open handle rewrites the slot row
		// exactly as a launcher provisional registration would, and holds it.
		conn, err := aStore.db.Conn(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = conn.Close() }()
		if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
			t.Fatal(err)
		}
		relaunchStart := homestate.CurrentProcessFingerprint()
		if relaunchStart == "" {
			relaunchStart = "relaunch-start-vii"
		}
		res, err := conn.ExecContext(ctx, `UPDATE peers SET session_uuid=?,generation=generation+1,pid=?,process_start=?,updated_at=? WHERE slot=?`,
			launchPendingSessionPrefix+newID(), os.Getpid(), relaunchStart, time.Now().UTC().Format(time.RFC3339Nano), slot)
		if err != nil {
			t.Fatal(err)
		}
		if n, _ := res.RowsAffected(); n != 1 {
			t.Fatalf("adversary rewrote %d rows", n)
		}

		// Subject V: the production reservation, pinned by its caller to the
		// source tuple it resolved before A began.
		type result struct {
			h   Handoff
			err error
			at  time.Time
		}
		done := make(chan result, 1)
		go func() {
			h, err := vStore.ReserveHandoff(ctx, HandoffReservation{
				Slot: slot, CardID: "t1082", SpecID: "SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001",
				Mode:           HandoffModeInteractive,
				DevelopPin:     "3f3ffbb57aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				TargetPath:     filepath.Join(root, ".claude", "worktrees", "t1082"),
				TargetBranch:   "WT-lane-handoff",
				ExpectedSource: &source,
			})
			done <- result{h: h, err: err, at: time.Now()}
		}()

		// V must be observed blocked on the SQLite lock while holding its own
		// connection: InUse==1 and WaitCount==0 on V's handle, call not returned.
		blocked := false
		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			st := vStore.db.Stats()
			if st.InUse == 1 && st.WaitCount == 0 {
				time.Sleep(150 * time.Millisecond)
				st = vStore.db.Stats()
				select {
				case r := <-done:
					t.Fatalf("subject returned while adversary held the lock: %+v err=%v", r.h, r.err)
				default:
				}
				blocked = st.InUse == 1 && st.WaitCount == 0
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if !blocked {
			t.Fatal("V_blocked_observed=false: subject was not observed waiting on the SQLite lock")
		}
		t.Log("V_blocked_observed=true")

		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			t.Fatal(err)
		}
		aCommit := time.Now()
		committed := readEndpointRow(t, seed.db, slot)

		var r result
		select {
		case r = <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("subject did not return after adversary commit")
		}
		t.Logf("A_commit_at=%s V_return_at=%s", aCommit.Format(time.RFC3339Nano), r.at.Format(time.RFC3339Nano))
		if !aCommit.Before(r.at) {
			t.Fatalf("timestamp order reversed: A commit %s is not earlier than V return %s", aCommit, r.at)
		}

		reason, isNack := HandoffNackReason(r.err)
		if !isNack || reason != NackEndpointLaunchPending {
			t.Fatalf("subject result = %+v err=%v, want NACK %s", r.h, r.err, NackEndpointLaunchPending)
		}
		if n := countLaneHandoffs(t, seed.db, slot); n != 0 {
			t.Fatalf("reservation rows for lane = %d, want 0 (stale-tuple reservation committed)", n)
		}
		if after := readEndpointRow(t, seed.db, slot); after != committed {
			t.Fatalf("slot row changed after subject: committed=%+v after=%+v", committed, after)
		}
	})
}
