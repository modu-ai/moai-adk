package factorymsg

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// wtReadyHandoff returns a store and a WT_READY handoff in the given mode, on a
// lane bound through the production launcher + SessionStart path.
func wtReadyHandoff(t *testing.T, mode string) (*Store, Handoff) {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	ctx := context.Background()
	root := filepath.Join(t.TempDir(), "project")
	s, err := Open(root, "run-switch")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	pending, err := s.RegisterLaunchPending(ctx, Peer{ProjectKey: "project", RunID: "run-switch", Backend: "codex", Role: "worker", Slot: "lane-1", PID: os.Getpid(), ProcessStart: "fake-start"})
	if err != nil {
		t.Fatal(err)
	}
	pending.SessionUUID = "src-uuid"
	if _, ok, err := s.BindLaunchPending(ctx, pending); err != nil || !ok {
		t.Fatalf("bind ok=%v err=%v", ok, err)
	}
	r := validReservation(root)
	r.Mode = mode
	h, err := s.ReserveHandoff(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	h, err = s.MarkHandoffWTReady(ctx, h)
	if err != nil {
		t.Fatal(err)
	}
	return s, h
}

func (s *Store) countRows(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := s.db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

// validRelocation is official headless evidence for h: a history-preserving
// fork into the reserved target, observed thread/started, and a controller
// readback that equals the reservation.
func validRelocation(h Handoff) HeadlessRelocation {
	return HeadlessRelocation{
		Method: RelocationMethodThreadFork, SourceThreadID: "thr-source", ThreadID: "thr-forked",
		ForkedFromID: "thr-source", ThreadStarted: true,
		RequestCwd: h.TargetPath, ResponseCwd: h.TargetPath,
		ReadbackCwd: h.TargetPath, ReadbackBranch: h.TargetBranch, ReadbackHead: h.DevelopPin,
	}
}

// TestHandoffSwitchPendingIsModeBound pins REQ-FLH-002: interactive and headless
// each reach only their own SWITCH_PENDING state, only from WT_READY, and never
// write BOUND, a tombstone, or a peer change.
func TestHandoffSwitchPendingIsModeBound(t *testing.T) {
	ctx := context.Background()

	s, h := wtReadyHandoff(t, HandoffModeInteractive)
	if _, err := s.MarkHandoffSwitchPendingHeadless(ctx, h); err == nil {
		t.Fatal("interactive handoff entered SWITCH_PENDING_HEADLESS")
	}
	peer := s.countRows(t, `SELECT generation FROM peers WHERE slot='lane-1'`)
	got, err := s.MarkHandoffSwitchPendingInteractive(ctx, h)
	if err != nil || got.State != HandoffSwitchPendingInteractive {
		t.Fatalf("interactive switch = %+v err=%v", got, err)
	}
	if _, err := s.MarkHandoffSwitchPendingInteractive(ctx, got); err == nil {
		t.Fatal("SWITCH_PENDING_INTERACTIVE accepted a second switch")
	}
	if n := s.countRows(t, `SELECT count(*) FROM lane_handoffs WHERE state='BOUND'`); n != 0 {
		t.Fatalf("BOUND rows = %d, want 0 in M2", n)
	}
	if n := s.countRows(t, `SELECT count(*) FROM lane_endpoint_tombstones`); n != 0 {
		t.Fatalf("tombstones = %d, want 0", n)
	}
	if after := s.countRows(t, `SELECT generation FROM peers WHERE slot='lane-1'`); after != peer {
		t.Fatalf("peer generation %d -> %d", peer, after)
	}

	hs, hh := wtReadyHandoff(t, HandoffModeHeadless)
	if _, err := hs.MarkHandoffSwitchPendingInteractive(ctx, hh); err == nil {
		t.Fatal("headless handoff entered SWITCH_PENDING_INTERACTIVE")
	}
	got, err = hs.MarkHandoffSwitchPendingHeadless(ctx, hh)
	if err != nil || got.State != HandoffSwitchPendingHeadless {
		t.Fatalf("headless switch = %+v err=%v", got, err)
	}
	// A SWITCH_PENDING handoff is still unfinished: it may be NACKed.
	if nacked, err := hs.NackHandoff(ctx, got, NackRelocationRPCFailed); err != nil || nacked.State != HandoffNack {
		t.Fatalf("nack from SWITCH_PENDING_HEADLESS = %+v err=%v", nacked, err)
	}
}

// TestHeadlessRelocationEvidenceRejectsNonEvidence pins REQ-FLH-007: only the
// official fork/start result with thread/started and a matching controller
// readback is relocation evidence; SessionStart, an empty turn/start, and
// turn/steer never are.
func TestHeadlessRelocationEvidenceRejectsNonEvidence(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name, detail string
		mutate       func(h Handoff, r *HeadlessRelocation)
	}{
		{"wrong_method_thread_start", "wrong_method_thread_start", func(_ Handoff, r *HeadlessRelocation) {
			r.Method, r.ForkedFromID = RelocationMethodThreadStart, ""
		}},
		{"fork_without_history", "wrong_method_thread_fork", func(_ Handoff, r *HeadlessRelocation) { r.SourceThreadID = "" }},
		{"session_start", "session_start_not_evidence", func(_ Handoff, r *HeadlessRelocation) { r.Method = "SessionStart" }},
		{"empty_turn_start", "empty_turn_not_evidence", func(_ Handoff, r *HeadlessRelocation) { r.Method = "turn/start" }},
		{"turn_steer", "turn_steer_not_evidence", func(_ Handoff, r *HeadlessRelocation) { r.Method = "turn/steer" }},
		{"unknown_method", "unknown_method", func(_ Handoff, r *HeadlessRelocation) { r.Method = "thread/resume" }},
		{"thread_started_missing", "thread_started_not_observed", func(_ Handoff, r *HeadlessRelocation) { r.ThreadStarted = false }},
		{"empty_thread_id", "empty_thread_id", func(_ Handoff, r *HeadlessRelocation) { r.ThreadID = "" }},
		{"thread_id_not_new", "thread_id_not_new", func(_ Handoff, r *HeadlessRelocation) { r.ThreadID = r.SourceThreadID }},
		{"lineage_missing", "forked_from_mismatch", func(_ Handoff, r *HeadlessRelocation) { r.ForkedFromID = "" }},
		{"lineage_foreign", "forked_from_mismatch", func(_ Handoff, r *HeadlessRelocation) { r.ForkedFromID = "thr-other" }},
		{"request_cwd", "cwd_mismatch", func(_ Handoff, r *HeadlessRelocation) { r.RequestCwd += "-x" }},
		{"response_cwd", "cwd_mismatch", func(h Handoff, r *HeadlessRelocation) { r.ResponseCwd = filepath.Dir(h.TargetPath) }},
		{"readback_cwd", "cwd_mismatch", func(h Handoff, r *HeadlessRelocation) { r.ReadbackCwd = filepath.Join(h.TargetPath, "sub") }},
		{"readback_branch", "branch_mismatch", func(_ Handoff, r *HeadlessRelocation) { r.ReadbackBranch = "WT-elsewhere" }},
		{"readback_head", "head_mismatch", func(_ Handoff, r *HeadlessRelocation) { r.ReadbackHead = strings.Repeat("b", 40) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, h := wtReadyHandoff(t, HandoffModeHeadless)
			h, err := s.MarkHandoffSwitchPendingHeadless(ctx, h)
			if err != nil {
				t.Fatal(err)
			}
			rel := validRelocation(h)
			tc.mutate(h, &rel)
			err = s.RecordHeadlessRelocation(ctx, h, rel)
			var nack *HandoffNackError
			if !errors.As(err, &nack) || nack.Reason != NackRelocationEvidenceInvalid {
				t.Fatalf("err=%v, want NACK %s", err, NackRelocationEvidenceInvalid)
			}
			if nack.Detail != tc.detail {
				t.Fatalf("detail = %q, want %q", nack.Detail, tc.detail)
			}
			if n := s.countRows(t, `SELECT count(*) FROM lane_handoff_relocations`); n != 0 {
				t.Fatalf("relocation rows = %d after rejected evidence", n)
			}
		})
	}

	t.Run("valid_fork_and_start_are_recorded", func(t *testing.T) {
		for _, history := range []bool{true, false} {
			s, h := wtReadyHandoff(t, HandoffModeHeadless)
			rel := validRelocation(h)
			if !history {
				rel.Method, rel.SourceThreadID, rel.ForkedFromID = RelocationMethodThreadStart, "", ""
			}
			if err := s.RecordHeadlessRelocation(ctx, h, rel); err == nil {
				t.Fatal("relocation recorded outside SWITCH_PENDING_HEADLESS")
			}
			h, err := s.MarkHandoffSwitchPendingHeadless(ctx, h)
			if err != nil {
				t.Fatal(err)
			}
			forged := h
			forged.Nonce = "not-the-nonce"
			if err := s.RecordHeadlessRelocation(ctx, forged, rel); err == nil {
				t.Fatal("relocation recorded with a forged nonce")
			}
			if err := s.RecordHeadlessRelocation(ctx, h, rel); err != nil {
				t.Fatalf("history=%v: %v", history, err)
			}
			if err := s.RecordHeadlessRelocation(ctx, h, rel); err == nil {
				t.Fatal("second relocation recorded for one handoff")
			}
			got, ok, err := s.HeadlessRelocationFor(ctx, h.ID)
			if err != nil || !ok {
				t.Fatalf("read back ok=%v err=%v", ok, err)
			}
			want := rel
			want.HandoffID, want.Nonce, want.RecordedAt = h.ID, h.Nonce, got.RecordedAt
			if got != want || got.RecordedAt.IsZero() {
				t.Fatalf("stored relocation = %+v, want %+v", got, want)
			}
			// Evidence is recorded, never bound: M3 owns BOUND.
			if n := s.countRows(t, `SELECT count(*) FROM lane_handoffs WHERE state!='SWITCH_PENDING_HEADLESS'`); n != 0 {
				t.Fatalf("handoff left SWITCH_PENDING_HEADLESS: %d rows", n)
			}
			if n := s.countRows(t, `SELECT count(*) FROM lane_endpoint_tombstones`); n != 0 {
				t.Fatalf("tombstones = %d", n)
			}
		}
		s, _ := wtReadyHandoff(t, HandoffModeHeadless)
		if _, ok, err := s.HeadlessRelocationFor(ctx, "no-such-handoff"); err != nil || ok {
			t.Fatalf("missing relocation ok=%v err=%v", ok, err)
		}
	})
}
