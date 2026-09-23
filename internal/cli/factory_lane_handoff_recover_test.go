package cli

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

// laneHandoffCrash is the panic value an injected crash point raises. A panic
// unwinds the controller the way a process death does for the broker: an open
// transaction rolls back, nothing after the crash point runs.
type laneHandoffCrash struct{ point string }

// crashAt makes the controller die at the named crash point.
func crashAt(t *testing.T, point string) {
	t.Helper()
	prev := laneHandoffFailpoint
	laneHandoffFailpoint = func(p string) {
		if p == point {
			panic(laneHandoffCrash{point: p})
		}
	}
	t.Cleanup(func() { laneHandoffFailpoint = prev })
}

// crashed runs fn and reports whether it died at an injected crash point.
func crashed(t *testing.T, fn func()) (died bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(laneHandoffCrash); !ok {
				panic(r)
			}
			died = true
		}
	}()
	fn()
	return false
}

// recoverRequest is the restarted controller's view of lane-1.
func (f *laneHandoffFixture) recoverRequest(t *testing.T) laneHandoffRecoverRequest {
	bind := headlessOwner(t)
	bind.ProjectRoot = f.primary
	return laneHandoffRecoverRequest{ProjectRoot: f.primary, Slot: handoffTestSlot, Owner: bind}
}

func (f *laneHandoffFixture) reconcile(t *testing.T) laneHandoffRecovery {
	t.Helper()
	laneHandoffFailpoint = func(string) {}
	got, err := recoverLaneHandoff(context.Background(), f.recoverRequest(t), f.deps())
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	return got
}

// driveHeadless runs the controller from reservation to BOUND.
func (f *laneHandoffFixture) driveHeadless(ctx context.Context, t *testing.T) {
	t.Helper()
	h, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeHeadless), f.deps())
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	h, err = switchLaneHandoffHeadless(ctx, h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{})
	if err != nil {
		t.Fatalf("switch: %v", err)
	}
	bind := headlessOwner(t)
	bind.ProjectRoot = f.primary
	if _, err := bindLaneHandoffHeadless(ctx, h, bind); err != nil {
		t.Fatalf("bind: %v", err)
	}
}

// latestHandoff is the lane's newest handoff.
func (f *laneHandoffFixture) latestHandoff(t *testing.T) factorymsg.Handoff {
	t.Helper()
	hs, err := f.store.HandoffsForLane(context.Background(), handoffTestSlot)
	if err != nil || len(hs) == 0 {
		t.Fatalf("handoffs = %v err=%v", hs, err)
	}
	return hs[len(hs)-1]
}

// requireRecoveryInvariants: one current endpoint, BOUND facts all-or-none,
// no duplicate release, no lost receipt, and the pre-handoff dispatch either
// released exactly once to the bound endpoint or still held for the source.
func (f *laneHandoffFixture) requireRecoveryInvariants(t *testing.T, d1 factorymsg.Envelope) {
	t.Helper()
	bound := f.count(t, `SELECT count(*) FROM lane_handoffs WHERE state='BOUND'`)
	for _, q := range []string{
		`SELECT count(*) FROM lane_endpoint_tombstones`,
		`SELECT count(*) FROM lane_handoff_receipts`,
		`SELECT count(*) FROM lane_dispatch_releases`,
		`SELECT count(*) FROM lane_handoff_events WHERE to_state='BOUND'`,
	} {
		if n := f.count(t, q); n != bound || n > 1 {
			t.Fatalf("%s = %d with %d BOUND handoff(s)", q, n, bound)
		}
	}
	if n := f.count(t, `SELECT count(*) FROM peers WHERE slot=?`, handoffTestSlot); n != 1 {
		t.Fatalf("lane rows = %d, want one current endpoint", n)
	}
	row := f.endpointRow(t)
	if bound == 0 {
		if row.Session != f.source.SessionUUID || row.Generation != f.source.Generation {
			t.Fatalf("endpoint moved without BOUND: %+v", row)
		}
		if n := f.count(t, `SELECT count(*) FROM messages WHERE id=? AND recipient_session=?`, d1.ID, f.source.SessionUUID); n != 1 {
			t.Fatalf("pre-handoff dispatch moved without BOUND")
		}
		return
	}
	if n := f.count(t, `SELECT count(*) FROM lane_handoff_receipts WHERE session_uuid=? AND generation=?`, row.Session, row.Generation); n != 1 {
		t.Fatalf("receipt does not name the current endpoint %+v", row)
	}
	if n := f.count(t, `SELECT count(*) FROM lane_message_releases WHERE message_id=?`, d1.ID); n != 1 {
		t.Fatalf("dispatch releases for %s = %d, want 1", d1.ID, n)
	}
	current := factorymsg.Peer{ProjectKey: "project", RunID: f.run, Backend: f.source.Backend, Role: f.source.Role, Slot: handoffTestSlot,
		SessionUUID: row.Session, Generation: row.Generation, PID: row.PID, ProcessStart: row.ProcessStart}
	ctx := context.Background()
	claims, err := f.store.Claim(ctx, current, factorymsg.MaxBatch, time.Hour)
	if err != nil || len(claims) != 1 || claims[0].ID != d1.ID {
		t.Fatalf("bound endpoint claims = %+v err=%v, want %s once", claims, err, d1.ID)
	}
	if more, err := f.store.Claim(ctx, current, factorymsg.MaxBatch, time.Hour); err != nil || len(more) != 0 {
		t.Fatalf("second claim = %+v err=%v, want none", more, err)
	}
}

// TestFactoryLaneHandoffCrashRecovery is AC-FLH-009 (REQ-FLH-011): a crash at
// each point of the handoff is recovered on restart by resume, idempotent
// finalize, or NACK — never two current endpoints, a duplicate release, or a
// lost receipt.
func TestFactoryLaneHandoffCrashRecovery(t *testing.T) {
	type crashCase struct {
		name, point, step, mode string
		want, wantState         string
		continueToBound         bool
	}
	cases := []crashCase{
		{name: "after_reserve_before_create", point: handoffPointReserved, mode: factorymsg.HandoffModeHeadless, want: laneRecoveryResume, wantState: factorymsg.HandoffWTReady, continueToBound: true},
		{name: "after_create_before_rename", point: handoffPointCreated, mode: factorymsg.HandoffModeHeadless, want: laneRecoveryResume, wantState: factorymsg.HandoffWTReady, continueToBound: true},
		{name: "after_rename_before_ready", point: handoffPointRenamed, mode: factorymsg.HandoffModeHeadless, want: laneRecoveryResume, wantState: factorymsg.HandoffWTReady, continueToBound: true},
		{name: "after_switch_pending_interactive", point: handoffPointSwitchPending, mode: factorymsg.HandoffModeInteractive, want: laneRecoveryResume, wantState: factorymsg.HandoffSwitchPendingInteractive},
		{name: "after_switch_pending_headless_before_rpc", point: handoffPointSwitchPending, mode: factorymsg.HandoffModeHeadless, want: laneRecoveryNack, wantState: factorymsg.HandoffNack},
		{name: "after_relocation_before_rebind", point: handoffPointRelocated, mode: factorymsg.HandoffModeHeadless, want: laneRecoveryResume, wantState: factorymsg.HandoffBound},
		{name: "after_commit_before_receipt_delivery", point: handoffPointBound, mode: factorymsg.HandoffModeHeadless, want: laneRecoveryFinalize, wantState: factorymsg.HandoffBound},
	}
	for _, step := range []string{"locked", "tombstone", "peer", "bound", "receipt", "release"} {
		cases = append(cases, crashCase{name: "rebind_write_" + step, step: step, mode: factorymsg.HandoffModeHeadless, want: laneRecoveryResume, wantState: factorymsg.HandoffBound})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-forked"})
			lead := abandonTestLead(t, f)
			ctx := context.Background()
			d1, err := f.store.Send(ctx, factorymsg.SendRequest{From: lead, To: f.source, Kind: factorymsg.KindDispatchNotice, IdempotencyKey: "d1", TaskRef: "t1082", CorrelationID: "c-d1", TTL: time.Hour, Payload: []byte("body-d1")})
			if err != nil {
				t.Fatal(err)
			}

			died := false
			switch {
			case tc.step != "":
				injected := errors.New("crash at rebind step " + tc.step)
				stepCtx := factorymsg.WithStepHook(ctx, func(s string) error {
					if s == tc.step {
						return injected
					}
					return nil
				})
				h, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, tc.mode), f.deps())
				if err != nil {
					t.Fatal(err)
				}
				if h, err = switchLaneHandoffHeadless(ctx, h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{}); err != nil {
					t.Fatal(err)
				}
				bind := headlessOwner(t)
				bind.ProjectRoot = f.primary
				_, err = bindLaneHandoffHeadless(stepCtx, h, bind)
				died = errors.Is(err, injected)
			case tc.mode == factorymsg.HandoffModeInteractive:
				crashAt(t, tc.point)
				died = crashed(t, func() {
					h, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, tc.mode), f.deps())
					if err != nil {
						t.Fatal(err)
					}
					_, _ = switchLaneHandoffInteractive(ctx, h, f.switchRequest(laneActivityIdle, ""), laneHandoffDeps{})
				})
			default:
				crashAt(t, tc.point)
				died = crashed(t, func() { f.driveHeadless(ctx, t) })
			}
			if !died {
				t.Fatalf("controller did not die at %s%s", tc.point, tc.step)
			}

			// Restart repeatedly: the first reconcile decides, later ones are
			// idempotent and write nothing.
			first := f.reconcile(t)
			if first.Decision != tc.want {
				t.Fatalf("decision = %s (%s), want %s", first.Decision, first.Reason, tc.want)
			}
			if st := f.latestHandoff(t); st.State != tc.wantState {
				t.Fatalf("state after recovery = %s/%s, want %s", st.State, st.Reason, tc.wantState)
			}
			settled := f.brokerSnapshot(t)
			for i := 0; i < 2; i++ {
				again := f.reconcile(t)
				switch again.Decision {
				case laneRecoveryResume, laneRecoveryFinalize, laneRecoveryNone:
				default:
					t.Fatalf("restart %d decision = %s, want an idempotent one", i+2, again.Decision)
				}
				if tc.want == laneRecoveryFinalize && (again.Binding == nil || first.Binding == nil || again.Binding.ReceiptID != first.Binding.ReceiptID) {
					t.Fatalf("finalize receipt changed: first=%+v again=%+v", first.Binding, again.Binding)
				}
				if got := f.brokerSnapshot(t); got != settled {
					t.Fatalf("restart %d wrote:\nbefore=%s\nafter=%s", i+2, settled, got)
				}
			}
			if tc.continueToBound {
				h := f.latestHandoff(t)
				h, err := switchLaneHandoffHeadless(ctx, h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{})
				if err != nil {
					t.Fatalf("resumed switch: %v", err)
				}
				bind := headlessOwner(t)
				bind.ProjectRoot = f.primary
				if _, err := bindLaneHandoffHeadless(ctx, h, bind); err != nil {
					t.Fatalf("resumed bind: %v", err)
				}
			}
			f.requireRecoveryInvariants(t, d1)
		})
	}
}

// dirListing renders every file under dir, excluding Git metadata.
func dirListing(t *testing.T, dir string) string {
	t.Helper()
	var files []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.Contains(p, string(filepath.Separator)+".git") {
			return nil
		}
		b, _ := os.ReadFile(p)
		files = append(files, p+"="+string(b))
		return nil
	})
	sort.Strings(files)
	return strings.Join(files, ";")
}

// TestFactoryLaneHandoffAbandonedWorktreeRecovery is AC-FLH-010 (REQ-FLH-011):
// a target recovery cannot prove safe is recorded ABANDONED with its reason
// and preserved byte for byte; it is never removed, bound, or dispatched.
func TestFactoryLaneHandoffAbandonedWorktreeRecovery(t *testing.T) {
	cases := []struct {
		name, reason string
		breakTarget  func(t *testing.T, f *laneHandoffFixture, target string)
		gitTarget    bool
	}{
		{"dirty", factorymsg.NackTargetDirty, func(t *testing.T, f *laneHandoffFixture, target string) {
			if err := os.WriteFile(filepath.Join(target, "wip.txt"), []byte("uncommitted\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}, true},
		{"unmerged", factorymsg.NackTargetUnmerged, func(t *testing.T, f *laneHandoffFixture, target string) {
			if err := os.WriteFile(filepath.Join(target, "card.txt"), []byte("card work\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			handoffGit(t, target, "add", "card.txt")
			handoffGit(t, target, "commit", "-qm", "t1082 unmerged")
		}, true},
		{"base_drifted", factorymsg.NackBaseDrift, func(t *testing.T, f *laneHandoffFixture, target string) {
			handoffGit(t, target, "reset", "-q", "--hard", f.mainSHA)
		}, true},
		{"branch_collided", factorymsg.NackBranchCollision, func(t *testing.T, f *laneHandoffFixture, target string) {
			handoffGit(t, target, "switch", "-q", "-c", "WT-other-card")
		}, true},
		{"unknown_owner", factorymsg.NackOwnerUnknown, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			lead := abandonTestLead(t, f)
			ctx := context.Background()
			d1, err := f.store.Send(ctx, factorymsg.SendRequest{From: lead, To: f.source, Kind: factorymsg.KindDispatchNotice, IdempotencyKey: "d1", TaskRef: "t1082", CorrelationID: "c-d1", TTL: time.Hour, Payload: []byte("body-d1")})
			if err != nil {
				t.Fatal(err)
			}
			target := f.target(handoffTestCard)
			var h factorymsg.Handoff
			if tc.gitTarget {
				h = f.wtReady(t, factorymsg.HandoffModeHeadless)
				tc.breakTarget(t, f, target)
			} else {
				// A directory MoAI did not create sits at the reserved target
				// and the controller died before it could look.
				crashAt(t, handoffPointReserved)
				if !crashed(t, func() {
					_, _ = prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeHeadless), f.deps())
				}) {
					t.Fatal("controller did not die after the reservation")
				}
				if err := os.MkdirAll(target, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(target, "someone-elses.txt"), []byte("not ours\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				h = f.latestHandoff(t)
			}
			files := dirListing(t, target)
			var gitState string
			if tc.gitTarget {
				gitState = targetSnapshot(t, f.primary, target, handoffGit(t, target, "branch", "--show-current"))
			}
			row := f.endpointRow(t)

			got := f.reconcile(t)
			if got.Decision != laneRecoveryAbandoned || got.Reason != tc.reason {
				t.Fatalf("decision = %s/%s, want %s/%s", got.Decision, got.Reason, laneRecoveryAbandoned, tc.reason)
			}
			if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffAbandoned || st.Reason != tc.reason {
				t.Fatalf("stored = %s/%s, want ABANDONED/%s", st.State, st.Reason, tc.reason)
			}
			if after := dirListing(t, target); after != files {
				t.Fatalf("target files changed:\nbefore=%s\nafter=%s", files, after)
			}
			if tc.gitTarget {
				if after := targetSnapshot(t, f.primary, target, handoffGit(t, target, "branch", "--show-current")); after != gitState {
					t.Fatalf("target branch or commits changed:\nbefore=%s\nafter=%s", gitState, after)
				}
				if list := handoffGit(t, f.primary, "worktree", "list", "--porcelain"); !strings.Contains(list, target) {
					t.Fatalf("worktree removed: %s", list)
				}
			}
			if after := f.endpointRow(t); after != row {
				t.Fatalf("endpoint moved: before=%+v after=%+v", row, after)
			}
			f.requireRecoveryInvariants(t, d1)
			if n := f.count(t, `SELECT count(*) FROM lane_handoffs WHERE state='BOUND'`); n != 0 {
				t.Fatalf("BOUND handoffs = %d", n)
			}
			// A later restart leaves the abandoned handoff alone.
			settled := f.brokerSnapshot(t)
			if again := f.reconcile(t); again.Decision != laneRecoveryNone {
				t.Fatalf("restart decision = %s, want none", again.Decision)
			}
			if after := f.brokerSnapshot(t); after != settled {
				t.Fatal("restart after ABANDONED wrote")
			}
		})
	}
}

// TestLaneHandoffRecoveryEdges pins the reconciler's remaining decisions: no
// handoff, a created target that vanished, and a resumed headless rebind the
// broker NACKs; plus the abandon command's argument refusals.
func TestLaneHandoffRecoveryEdges(t *testing.T) {
	ctx := context.Background()
	t.Run("no_handoff", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		if got := f.reconcile(t); got.Decision != laneRecoveryNone {
			t.Fatalf("decision = %s, want none", got.Decision)
		}
	})
	t.Run("target_missing", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		h := f.wtReady(t, factorymsg.HandoffModeHeadless)
		handoffGit(t, f.primary, "worktree", "remove", "--force", h.TargetPath)
		got := f.reconcile(t)
		if got.Decision != laneRecoveryAbandoned || got.Reason != factorymsg.NackTargetMissing {
			t.Fatalf("decision = %s/%s, want abandoned/%s", got.Decision, got.Reason, factorymsg.NackTargetMissing)
		}
	})
	t.Run("resumed_rebind_nacked", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-forked"})
		h := f.wtReady(t, factorymsg.HandoffModeHeadless)
		if _, err := switchLaneHandoffHeadless(ctx, h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{}); err != nil {
			t.Fatal(err)
		}
		req := f.recoverRequest(t)
		req.Owner.OwnerProcessStart = "not-the-owner-start" // the relocated thread's owner is not current
		got, err := recoverLaneHandoff(ctx, req, f.deps())
		if err != nil || got.Decision != laneRecoveryNack || got.Reason != factorymsg.NackBindingEvidenceInvalid {
			t.Fatalf("decision = %+v err=%v, want nack/%s", got, err, factorymsg.NackBindingEvidenceInvalid)
		}
		if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffNack {
			t.Fatalf("stored = %s, want NACK", st.State)
		}
	})
	t.Run("no_active_run", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		req := f.recoverRequest(t)
		req.RunID = "no-such-run"
		if _, err := recoverLaneHandoff(ctx, req, f.deps()); err == nil {
			t.Fatal("recovery on an inactive run accepted")
		}
	})
	t.Run("develop_moved_after_rename", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		crashAt(t, handoffPointRenamed)
		if !crashed(t, func() {
			_, _ = prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeHeadless), f.deps())
		}) {
			t.Fatal("controller did not die after the rename")
		}
		handoffGit(t, f.primary, "update-ref", "refs/heads/develop", f.mainSHA)
		got := f.reconcile(t)
		if got.Decision != laneRecoveryAbandoned || got.Reason != factorymsg.NackBaseDrift {
			t.Fatalf("decision = %s/%s, want abandoned/%s", got.Decision, got.Reason, factorymsg.NackBaseDrift)
		}
	})
	t.Run("rename_collides_on_resume", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		crashAt(t, handoffPointCreated)
		if !crashed(t, func() {
			_, _ = prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeHeadless), f.deps())
		}) {
			t.Fatal("controller did not die after creation")
		}
		handoffGit(t, f.primary, "branch", handoffTestBranch, f.developPin)
		got := f.reconcile(t)
		if got.Decision != laneRecoveryAbandoned || got.Reason != factorymsg.NackBranchCollision {
			t.Fatalf("decision = %s/%s, want abandoned/%s", got.Decision, got.Reason, factorymsg.NackBranchCollision)
		}
	})
	t.Run("bound_without_receipt_is_corruption", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-forked"})
		f.driveHeadless(ctx, t)
		path, err := factorymsg.BrokerPath(f.primary, f.run)
		if err != nil {
			t.Fatal(err)
		}
		db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
		if err != nil {
			t.Fatal(err)
		}
		// Test-only corruption the one-transaction rebind cannot produce.
		if _, err := db.Exec(`DELETE FROM lane_handoff_receipts`); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
		laneHandoffFailpoint = func(string) {}
		if _, err := recoverLaneHandoff(ctx, f.recoverRequest(t), f.deps()); err == nil || !strings.Contains(err.Error(), "refusing to guess") {
			t.Fatalf("BOUND without receipt: err=%v, want a corruption report", err)
		}
	})
	t.Run("abandon_command_arguments", func(t *testing.T) {
		newLaneHandoffFixture(t, "develop", true) // cwd and active run for the command
		if _, err := runFactoryHandoffCommand(t, "abandon-lane"); err == nil {
			t.Fatal("abandon-lane without --slot accepted")
		}
		if _, err := runFactoryHandoffCommand(t, "abandon-lane", "--slot", handoffTestSlot, "--run", "no-such-run"); err == nil {
			t.Fatal("abandon-lane on an inactive run accepted")
		}
	})
}
