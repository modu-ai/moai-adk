package factorymsg

// legacy_slot_m1_test.go — SPEC-ROLE-NAMING-CODE-001 M1: the broker recognizes
// only the new vocabulary (`leader`, `lane`, `lane-<n>`); legacy slot inputs
// are refused with the canonical name; the run-retire identity fallback reads
// a legacy `lead` peer as identity evidence; a live legacy peer blocks a
// leader re-entry into its run (REQ-RNC-009, REQ-RNC-013, REQ-RNC-022,
// REQ-RNC-024).

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func m1Peer(run, session, role, slot string, generation int64) Peer {
	p := testPeer(run, session, generation)
	p.Role = role
	p.Slot = slot
	return p
}

func TestResolveLaneRefusesLegacySlots(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	leader := m1Peer(s.runID, "lead-s", "leader", "leader", 1)
	lane := m1Peer(s.runID, "lane-s", "lane", "lane-1", 1)
	if _, err := s.RegisterPeer(ctx, leader); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegisterPeer(ctx, lane); err != nil {
		t.Fatal(err)
	}

	// New vocabulary delivers (resolves to the peer).
	for _, slot := range []string{"leader", "lane-1"} {
		if _, err := s.ResolveLane(ctx, slot); err != nil {
			t.Errorf("ResolveLane(%q) = %v, want the peer", slot, err)
		}
	}
	// Legacy slots are refused with the canonical name.
	for slot, canonical := range map[string]string{
		"lead":     "leader",
		"worker":   "lane",
		"agent":    "lane",
		"worker-1": "lane-1",
		"agent-1":  "lane-1",
	} {
		_, err := s.ResolveLane(ctx, slot)
		if err == nil {
			t.Errorf("ResolveLane(%q) = nil error, want refusal", slot)
			continue
		}
		if !strings.Contains(err.Error(), canonical) {
			t.Errorf("ResolveLane(%q) error %q does not name canonical %q", slot, err.Error(), canonical)
		}
	}
}

func TestRegisterPeerBareLaneNumbersLaneOnly(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	// A pre-existing legacy row does not participate in lane numbering.
	legacy := m1Peer(s.runID, "legacy-s", "worker", "worker-1", 1)
	if _, err := s.RegisterPeer(ctx, legacy); err != nil {
		t.Fatalf("register legacy fixture peer: %v", err)
	}
	p1 := m1Peer(s.runID, "lane-1-s", "lane", "lane", 1)
	got, err := s.RegisterPeer(ctx, p1)
	if err != nil {
		t.Fatalf("RegisterPeer(lane): %v", err)
	}
	if got.Slot != "lane-1" {
		t.Errorf("bare lane slot resolved to %q, want lane-1 (legacy worker-1 not in the number space)", got.Slot)
	}
	p2 := m1Peer(s.runID, "lane-2-s", "lane", "lane", 1)
	got, err = s.RegisterPeer(ctx, p2)
	if err != nil {
		t.Fatalf("RegisterPeer(lane) second: %v", err)
	}
	if got.Slot != "lane-2" {
		t.Errorf("second bare lane slot resolved to %q, want lane-2", got.Slot)
	}
	// Bare legacy role inputs are refused, never auto-numbered.
	for _, slot := range []string{"worker", "agent"} {
		p := m1Peer(s.runID, "x-"+slot, "lane", slot, 1)
		if _, err := s.RegisterPeer(ctx, p); err == nil {
			t.Errorf("RegisterPeer(slot %q) = nil error, want refusal naming lane", slot)
		}
	}
}

func TestLeadPeerIdentityReadsLeaderAndLegacyLead(t *testing.T) {
	root := t.TempDir()

	// A legacy-only run: the leader peer is recorded under role='lead'
	// (pre-rename binary). The identity is still readable — as evidence only.
	s, err := Open(root, "run-legacy")
	if err != nil {
		t.Fatal(err)
	}
	start := homestate.CurrentProcessFingerprint()
	legacy := m1Peer("run-legacy", "lead-s", "lead", "lead", 1)
	legacy.PID = os.Getpid()
	legacy.ProcessStart = start
	if _, err := s.RegisterPeer(context.Background(), legacy); err != nil {
		t.Fatal(err)
	}
	_ = s.Close()
	pid, pstart, ok := LeaderPeerIdentity(root, "run-legacy")
	if !ok {
		t.Fatalf("LeaderPeerIdentity(legacy-only run) ok=false, want the lead peer identity")
	}
	if pid != os.Getpid() || pstart != start {
		t.Errorf("LeaderPeerIdentity = (%d, %q), want (%d, %q)", pid, pstart, os.Getpid(), start)
	}

	// A post-change run: role='leader' wins.
	s2, err := Open(root, "run-new")
	if err != nil {
		t.Fatal(err)
	}
	leader := m1Peer("run-new", "leader-s", "leader", "leader", 1)
	leader.PID = os.Getpid()
	leader.ProcessStart = start
	if _, err := s2.RegisterPeer(context.Background(), leader); err != nil {
		t.Fatal(err)
	}
	_ = s2.Close()
	pid, _, ok = LeaderPeerIdentity(root, "run-new")
	if !ok || pid != os.Getpid() {
		t.Errorf("LeaderPeerIdentity(leader run) = (%d, %v), want the leader peer", pid, ok)
	}
}

func TestLiveLegacyPeerDetectsLiveLegacyRow(t *testing.T) {
	root := t.TempDir()
	s, err := Open(root, "runR")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	start := homestate.CurrentProcessFingerprint()
	live := m1Peer("runR", "legacy-live", "worker", "worker-2", 1)
	live.PID = os.Getpid()
	live.ProcessStart = start
	if _, err := s.RegisterPeer(ctx, live); err != nil {
		t.Fatal(err)
	}

	value, blocked, err := LiveLegacyPeer(ctx, root, "runR")
	if err != nil {
		t.Fatal(err)
	}
	if !blocked || value != "worker-2" {
		t.Errorf("LiveLegacyPeer = (%q, %v), want (worker-2, true)", value, blocked)
	}
	_ = s.Close()

	// Dead legacy peer: stale, not blocking.
	root2 := t.TempDir()
	s2, err := Open(root2, "runD")
	if err != nil {
		t.Fatal(err)
	}
	dead := m1Peer("runD", "legacy-dead", "worker", "worker-2", 1)
	dead.PID = 999999999
	dead.ProcessStart = "2020-01-01T00:00:00Z"
	if _, err := s2.RegisterPeer(ctx, dead); err != nil {
		t.Fatal(err)
	}
	_ = s2.Close()
	value, blocked, err = LiveLegacyPeer(ctx, root2, "runD")
	if err != nil {
		t.Fatal(err)
	}
	if blocked {
		t.Errorf("LiveLegacyPeer(dead row) = (%q, true), want not blocking", value)
	}

	// A current-vocabulary run blocks nothing.
	root3 := t.TempDir()
	s3, err := Open(root3, "runN")
	if err != nil {
		t.Fatal(err)
	}
	cur := m1Peer("runN", "lane-s", "lane", "lane-1", 1)
	cur.PID = os.Getpid()
	cur.ProcessStart = start
	if _, err := s3.RegisterPeer(ctx, cur); err != nil {
		t.Fatal(err)
	}
	_ = s3.Close()
	if _, blocked, err := LiveLegacyPeer(ctx, root3, "runN"); err != nil || blocked {
		t.Errorf("LiveLegacyPeer(current vocab) = (_, %v, %v), want false, nil", blocked, err)
	}
}

// AC-RNC-009: delivery. A message addressed to leader / lane-1 is stored; a
// message addressed through a legacy slot is refused and stores nothing.
func TestSendDeliversNewVocabularyRefusesLegacy(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	from := m1Peer(s.runID, "from-s", "leader", "leader", 1)
	to := m1Peer(s.runID, "to-s", "lane", "lane-1", 1)
	if _, err := s.RegisterPeer(ctx, from); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegisterPeer(ctx, to); err != nil {
		t.Fatal(err)
	}
	send := func(dest Peer) (int, error) {
		_, err := s.Send(ctx, SendRequest{
			From: from, To: dest, Kind: KindStatusRequest,
			IdempotencyKey: "idem-" + dest.Slot, TaskRef: "t1256", CorrelationID: "c-" + dest.Slot, TTL: time.Minute,
			ExpectedTaskRevision: 0, CurrentTaskRevision: 0,
			Payload: []byte("ping"),
		})
		var n int
		if e := s.db.QueryRowContext(ctx, `SELECT count(*) FROM messages`).Scan(&n); e != nil {
			t.Fatal(e)
		}
		return n, err
	}

	n0, err := send(to)
	if err != nil {
		t.Fatalf("Send to lane-1: %v", err)
	}
	if n0 != 1 {
		t.Fatalf("messages after lane-1 send = %d, want 1", n0)
	}

	for _, slot := range []string{"lead", "worker-1", "agent-1", "worker", "agent"} {
		p := to
		p.Slot = slot
		_, err := send(p)
		if err == nil {
			t.Errorf("Send to legacy slot %q = nil error, want refusal", slot)
		}
		var n int
		if e := s.db.QueryRowContext(ctx, `SELECT count(*) FROM messages`).Scan(&n); e != nil {
			t.Fatal(e)
		}
		if n != n0 {
			t.Errorf("Send to legacy slot %q changed message count %d → %d", slot, n0, n)
		}
	}
}
