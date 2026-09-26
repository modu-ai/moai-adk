package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// runFactoryHandoffCommand runs `moai factory handoff <args>` through the
// production command tree and returns its output and error.
func runFactoryHandoffCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(append([]string{"factory", "handoff"}, args...))
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})
	err := rootCmd.Execute()
	return out.String(), err
}

// withAbandonLaneProbe swaps the one liveness seam the abandon command wires.
func withAbandonLaneProbe(t *testing.T, state homestate.ProcessIdentityState, start string) {
	t.Helper()
	prev := abandonLaneProbe
	abandonLaneProbe = func(int) (string, homestate.ProcessIdentityState) { return start, state }
	t.Cleanup(func() { abandonLaneProbe = prev })
}

// brokerSnapshot is every broker fact the abandon command may not touch on a
// refusal, rendered as one comparable string.
func (f *laneHandoffFixture) brokerSnapshot(t *testing.T) string {
	t.Helper()
	db := f.brokerDB(t)
	var parts []string
	for _, q := range []string{
		`SELECT id||'|'||state||'|'||reason||'|'||nonce||'|'||updated_at FROM lane_handoffs ORDER BY handoff_generation`,
		`SELECT slot||'|'||session_uuid||'|'||generation||'|'||pid||'|'||process_start||'|'||updated_at FROM peers ORDER BY slot`,
		`SELECT count(*) FROM lane_handoff_events`,
		`SELECT count(*) FROM lane_endpoint_tombstones`,
		`SELECT count(*) FROM lane_handoff_receipts`,
		`SELECT count(*) FROM lane_dispatch_releases`,
		`SELECT count(*) FROM lane_message_releases`,
		`SELECT id||'|'||recipient_session||'|'||recipient_generation||'|'||state FROM messages ORDER BY id`,
	} {
		rows, err := db.Query(q)
		if err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				t.Fatal(err)
			}
			parts = append(parts, s)
		}
		_ = rows.Close()
		parts = append(parts, "#")
	}
	return strings.Join(parts, ";")
}

// targetSnapshot is the target worktree's files, branch, and commits.
func targetSnapshot(t *testing.T, primary, target, branch string) string {
	t.Helper()
	var files []string
	_ = filepath.WalkDir(target, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.Contains(p, string(filepath.Separator)+".git") {
			return nil
		}
		b, _ := os.ReadFile(p)
		files = append(files, p+"="+string(b))
		return nil
	})
	return strings.Join(files, ";") + "#" + handoffGit(t, primary, "rev-parse", "refs/heads/"+branch) + "#" +
		handoffGit(t, target, "status", "--porcelain") + "#" + handoffGit(t, target, "log", "--format=%H", "-n", "5")
}

// TestFactoryLaneHandoffOperatorAbandon is AC-FLH-020 (REQ-FLH-011): the
// operator command terminates a stuck non-final handoff only when its source
// owner is not current, and preserves the target worktree.
func TestFactoryLaneHandoffOperatorAbandon(t *testing.T) {
	states := []struct {
		name string
		to   func(t *testing.T, f *laneHandoffFixture) factorymsg.Handoff
	}{
		{factorymsg.HandoffReserved, func(t *testing.T, f *laneHandoffFixture) factorymsg.Handoff {
			h, err := f.store.ReserveHandoff(context.Background(), factorymsg.HandoffReservation{
				Slot: handoffTestSlot, CardID: handoffTestCard, SpecID: handoffTestSpec, Mode: factorymsg.HandoffModeInteractive,
				DevelopPin: f.developPin, TargetPath: f.target(handoffTestCard), TargetBranch: handoffTestBranch,
			})
			if err != nil {
				t.Fatal(err)
			}
			return h
		}},
		{factorymsg.HandoffWTReady, func(t *testing.T, f *laneHandoffFixture) factorymsg.Handoff {
			return f.wtReady(t, factorymsg.HandoffModeInteractive)
		}},
		{factorymsg.HandoffSwitchPendingInteractive, func(t *testing.T, f *laneHandoffFixture) factorymsg.Handoff {
			h, err := f.store.MarkHandoffSwitchPendingInteractive(context.Background(), f.wtReady(t, factorymsg.HandoffModeInteractive))
			if err != nil {
				t.Fatal(err)
			}
			return h
		}},
		{factorymsg.HandoffSwitchPendingHeadless, func(t *testing.T, f *laneHandoffFixture) factorymsg.Handoff {
			h, err := f.store.MarkHandoffSwitchPendingHeadless(context.Background(), f.wtReady(t, factorymsg.HandoffModeHeadless))
			if err != nil {
				t.Fatal(err)
			}
			return h
		}},
	}
	for _, tc := range states {
		t.Run(tc.name, func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			ctx := context.Background()
			lead := abandonTestLead(t, f)
			d1, err := f.store.Send(ctx, factorymsg.SendRequest{From: lead, To: f.source, Kind: factorymsg.KindDispatchNotice, IdempotencyKey: "d1", TaskRef: "t1082", CorrelationID: "c-d1", TTL: time.Hour, Payload: []byte("body-d1")})
			if err != nil {
				t.Fatal(err)
			}
			h := tc.to(t, f)
			if h.State != tc.name {
				t.Fatalf("fixture state = %s, want %s", h.State, tc.name)
			}
			target := f.target(handoffTestCard)
			if pathExists(target) {
				// A dirty target with an unmerged commit on its WT branch.
				if err := os.WriteFile(filepath.Join(target, "card-work.txt"), []byte("committed\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				handoffGit(t, target, "add", "card-work.txt")
				handoffGit(t, target, "commit", "-qm", "t1082 unmerged card work")
				if err := os.WriteFile(filepath.Join(target, "dirty.txt"), []byte("uncommitted\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			restarted := factorymsg.Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: handoffTestSlot, SessionUUID: "restarted-uuid", Generation: 1, PID: 999_983, ProcessStart: "restarted-start"}
			if _, err := f.store.RegisterPeer(ctx, restarted); err == nil {
				t.Fatal("UserPromptSubmit registration accepted during the non-final handoff")
			} else {
				requireHandoffNack(t, err, factorymsg.NackEndpointHandoffPending)
			}

			var targetBefore string
			if pathExists(target) {
				targetBefore = targetSnapshot(t, f.primary, target, handoffTestBranch)
			}
			refusals := []struct {
				name  string
				state homestate.ProcessIdentityState
				start string
			}{
				{"owner_current", homestate.ProcessIdentityLive, f.source.ProcessStart},
				{"owner_liveness_not_established", homestate.ProcessIdentityIndeterminate, ""},
			}
			for _, r := range refusals {
				before := f.brokerSnapshot(t)
				withAbandonLaneProbe(t, r.state, r.start)
				out, err := runFactoryHandoffCommand(t, "abandon-lane", "--slot", handoffTestSlot)
				requireHandoffNack(t, err, factorymsg.NackSourceOwnerLive)
				if !strings.Contains(out, factorymsg.NackSourceOwnerLive) {
					t.Fatalf("%s: output %q does not name %s", r.name, out, factorymsg.NackSourceOwnerLive)
				}
				if after := f.brokerSnapshot(t); after != before {
					t.Fatalf("%s: refusal wrote:\nbefore=%s\nafter=%s", r.name, before, after)
				}
			}

			eventsBefore := f.count(t, `SELECT count(*) FROM lane_handoff_events WHERE handoff_id=?`, h.ID)
			rowBefore := f.endpointRow(t)
			withAbandonLaneProbe(t, homestate.ProcessIdentityDead, "")
			out, err := runFactoryHandoffCommand(t, "abandon-lane", "--slot", handoffTestSlot)
			if err != nil {
				t.Fatalf("abandon with the source owner not current: %v (output %q)", err, out)
			}
			if !strings.Contains(out, factorymsg.HandoffAbandoned) || !strings.Contains(out, factorymsg.AbandonOperator) {
				t.Fatalf("abandon output %q", out)
			}
			if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffAbandoned || st.Reason != factorymsg.AbandonOperator {
				t.Fatalf("handoff = %s/%s, want ABANDONED/%s", st.State, st.Reason, factorymsg.AbandonOperator)
			}
			if n := f.count(t, `SELECT count(*) FROM lane_handoff_events WHERE handoff_id=?`, h.ID); n != eventsBefore+1 {
				t.Fatalf("events = %d, want %d (one ABANDONED event)", n, eventsBefore+1)
			}
			for q, want := range map[string]int{
				`SELECT count(*) FROM lane_handoffs WHERE state='BOUND'`:          0,
				`SELECT count(*) FROM lane_handoff_events WHERE to_state='BOUND'`: 0,
				`SELECT count(*) FROM lane_endpoint_tombstones`:                   0,
				`SELECT count(*) FROM lane_handoff_receipts`:                      0,
				`SELECT count(*) FROM lane_dispatch_releases`:                     0,
				`SELECT count(*) FROM lane_message_releases`:                      0,
			} {
				if n := f.count(t, q); n != want {
					t.Fatalf("%s = %d, want %d", q, n, want)
				}
			}
			if after := f.endpointRow(t); after != rowBefore {
				t.Fatalf("abandon moved the endpoint: before=%+v after=%+v", rowBefore, after)
			}
			if targetBefore != "" {
				if after := targetSnapshot(t, f.primary, target, handoffTestBranch); after != targetBefore {
					t.Fatalf("target worktree changed:\nbefore=%s\nafter=%s", targetBefore, after)
				}
				if list := handoffGit(t, f.primary, "worktree", "list", "--porcelain"); !strings.Contains(list, target) {
					t.Fatalf("target worktree removed: %s", list)
				}
			}

			// t1074 semantics return: the restarted session registers, and the
			// dispatch body stays withheld because no BOUND exists.
			bound, err := f.store.RegisterPeer(ctx, restarted)
			if err != nil {
				t.Fatalf("UserPromptSubmit registration after abandon: %v", err)
			}
			if claims, err := f.store.Claim(ctx, bound, factorymsg.MaxBatch, time.Hour); err != nil || len(claims) != 0 {
				t.Fatalf("restarted session claims = %+v err=%v, want the pre-handoff dispatch %s withheld", claims, err, d1.ID)
			}

			before := f.brokerSnapshot(t)
			out, err = runFactoryHandoffCommand(t, "abandon-lane", "--slot", handoffTestSlot)
			requireHandoffNack(t, err, factorymsg.NackHandoffNotPending)
			if !strings.Contains(out, factorymsg.NackHandoffNotPending) {
				t.Fatalf("second run output %q", out)
			}
			if after := f.brokerSnapshot(t); after != before {
				t.Fatal("second run wrote")
			}
		})
	}

	t.Run("no_handoff_on_slot", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		withAbandonLaneProbe(t, homestate.ProcessIdentityDead, "")
		before := f.brokerSnapshot(t)
		out, err := runFactoryHandoffCommand(t, "abandon-lane", "--slot", handoffTestSlot)
		requireHandoffNack(t, err, factorymsg.NackHandoffNotPending)
		if !strings.Contains(out, factorymsg.NackHandoffNotPending) {
			t.Fatalf("output %q", out)
		}
		if after := f.brokerSnapshot(t); after != before {
			t.Fatal("run on a slot with no handoff wrote")
		}
	})
}

// abandonTestLead binds a lead endpoint through the production launcher path.
func abandonTestLead(t *testing.T, f *laneHandoffFixture) factorymsg.Peer {
	t.Helper()
	ctx := context.Background()
	pending, err := f.store.RegisterLaunchPending(ctx, factorymsg.Peer{ProjectKey: "project", RunID: f.run, Backend: "claude", Role: "lead", Slot: "lead", PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint()})
	if err != nil {
		t.Fatal(err)
	}
	pending.SessionUUID = "lead-uuid"
	lead, ok, err := f.store.BindLaunchPending(ctx, pending)
	if err != nil || !ok {
		t.Fatalf("bind lead ok=%v err=%v", ok, err)
	}
	return lead
}
