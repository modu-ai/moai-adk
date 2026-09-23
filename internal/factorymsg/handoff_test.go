package factorymsg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validReservation(root string) HandoffReservation {
	return HandoffReservation{
		Slot: "lane-1", CardID: "t1082", SpecID: "SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001",
		Mode:         HandoffModeHeadless,
		DevelopPin:   strings.Repeat("a", 40),
		TargetPath:   filepath.Join(root, ".claude", "worktrees", "t1082"),
		TargetBranch: "WT-lane-handoff",
	}
}

// TestLaneHandoffStoreLifecycle pins the broker-level reservation contract the
// cli controller relies on: request validation, lane admission, one open
// handoff per lane, the nonce-and-state CAS, and the stored row shape.
func TestLaneHandoffStoreLifecycle(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	ctx := context.Background()
	root := filepath.Join(t.TempDir(), "project")
	s, err := Open(root, "run-lifecycle")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	requireNack := func(t *testing.T, err error, want string) {
		t.Helper()
		if got, ok := HandoffNackReason(err); !ok || got != want {
			t.Fatalf("err=%v, want NACK %s", err, want)
		}
	}

	invalid := map[string]func(*HandoffReservation){
		"slot":   func(r *HandoffReservation) { r.Slot = "../x" },
		"card":   func(r *HandoffReservation) { r.CardID = "t/1" },
		"spec":   func(r *HandoffReservation) { r.SpecID = "" },
		"mode":   func(r *HandoffReservation) { r.Mode = "desktop" },
		"pin":    func(r *HandoffReservation) { r.DevelopPin = "develop" },
		"path":   func(r *HandoffReservation) { r.TargetPath = "relative/t1082" },
		"branch": func(r *HandoffReservation) { r.TargetBranch = "t1082" },
	}
	for name, mutate := range invalid {
		r := validReservation(root)
		mutate(&r)
		_, err := s.ReserveHandoff(ctx, r)
		requireNack(t, err, NackInvalidRequest)
		if !strings.Contains(err.Error(), "handoff NACK "+NackInvalidRequest) {
			t.Fatalf("%s: error text %q", name, err)
		}
	}

	_, err = s.ReserveHandoff(ctx, validReservation(root))
	requireNack(t, err, NackLaneUnknown)

	pending, err := s.RegisterLaunchPending(ctx, Peer{ProjectKey: "project", RunID: "run-lifecycle", Backend: "codex", Role: "worker", Slot: "lane-1", PID: os.Getpid(), ProcessStart: "fake-start"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.ReserveHandoff(ctx, validReservation(root))
	requireNack(t, err, NackEndpointLaunchPending)

	bound := pending
	bound.SessionUUID = "src-uuid"
	source, ok, err := s.BindLaunchPending(ctx, bound)
	if err != nil || !ok {
		t.Fatalf("bind ok=%v err=%v", ok, err)
	}
	stale := validReservation(root)
	old := source
	old.Generation--
	stale.ExpectedSource = &old
	_, err = s.ReserveHandoff(ctx, stale)
	requireNack(t, err, NackStaleReservation)

	fresh := validReservation(root)
	fresh.ExpectedSource = &source
	h, err := s.ReserveHandoff(ctx, fresh)
	if err != nil {
		t.Fatal(err)
	}
	if h.State != HandoffReserved || h.Generation != 1 || h.Nonce == "" || h.Source.SessionUUID != "src-uuid" {
		t.Fatalf("reserved = %+v", h)
	}
	_, err = s.ReserveHandoff(ctx, validReservation(root))
	requireNack(t, err, NackHandoffInFlight)

	forged := h
	forged.Nonce = "not-the-nonce"
	if _, err := s.MarkHandoffWTReady(ctx, forged); err == nil {
		t.Fatal("transition accepted a forged nonce")
	}
	ready, err := s.MarkHandoffWTReady(ctx, h)
	if err != nil || ready.State != HandoffWTReady {
		t.Fatalf("wt ready = %+v err=%v", ready, err)
	}
	if _, err := s.MarkHandoffWTReady(ctx, ready); err == nil {
		t.Fatal("WT_READY accepted a second WT_READY transition")
	}
	if _, err := s.NackHandoff(ctx, ready, ""); err == nil {
		t.Fatal("NACK without a reason accepted")
	}
	nacked, err := s.NackHandoff(ctx, ready, "EVIDENCE_MISMATCH")
	if err != nil || nacked.State != HandoffNack || nacked.Reason != "EVIDENCE_MISMATCH" {
		t.Fatalf("nack = %+v err=%v", nacked, err)
	}
	if _, err := s.NackHandoff(ctx, nacked, "AGAIN"); err == nil {
		t.Fatal("a final handoff accepted another NACK")
	}

	// A NACK frees the lane: the next reservation takes the next generation.
	next, err := s.ReserveHandoff(ctx, validReservation(root))
	if err != nil || next.Generation != 2 {
		t.Fatalf("next = %+v err=%v", next, err)
	}
	hs, err := s.HandoffsForLane(ctx, "lane-1")
	if err != nil || len(hs) != 2 {
		t.Fatalf("handoffs = %+v err=%v", hs, err)
	}
	if hs[0].ID != h.ID || hs[0].State != HandoffNack || hs[0].Source != (Peer{
		ProjectKey: s.projectKey, RunID: "run-lifecycle", Backend: "codex", Role: "worker", Slot: "lane-1",
		SessionUUID: "src-uuid", Generation: source.Generation, PID: source.PID, ProcessStart: source.ProcessStart,
	}) || hs[1].State != HandoffReserved {
		t.Fatalf("stored handoffs = %+v", hs)
	}
	var events int
	if err := s.db.QueryRow(`SELECT count(*) FROM lane_handoff_events WHERE handoff_id=?`, h.ID).Scan(&events); err != nil || events != 3 {
		t.Fatalf("events for first handoff = %d err=%v, want 3 (RESERVED, WT_READY, NACK)", events, err)
	}
}
