package sessionmsg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestRegisterIdempotent verifies AC-CSM-002 (REQ-CSM-003): re-registering
// the same kind+name returns the same agentId, the record file exists under
// agents/<agentId>.json, and re-registration refreshes the heartbeat.
func TestRegisterIdempotent(t *testing.T) {
	root := t.TempDir()
	clk := &FakeClock{Current: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)

	rec1, err := s.Register(KindClaude, "lead", "orchestrator")
	if err != nil {
		t.Fatalf("first register: %v", err)
	}
	if rec1.AgentID == "" {
		t.Fatal("first register returned empty agentId")
	}
	if !strings.HasPrefix(rec1.AgentID, "claude-") {
		t.Errorf("agentId %q lacks kind prefix %q", rec1.AgentID, KindClaude)
	}
	if got := len(strings.TrimPrefix(rec1.AgentID, "claude-")); got != 8 {
		t.Errorf("agentId hex suffix length = %d, want 8 (<kind>-<hex8>)", got)
	}
	if _, err := os.Stat(filepath.Join(root, "agents", rec1.AgentID+".json")); err != nil {
		t.Fatalf("agent record file missing: %v", err)
	}
	if !rec1.Capabilities.Messaging {
		t.Errorf("new record capabilities.messaging = false, want true")
	}
	if rec1.Version != "1" {
		t.Errorf("new record version = %q, want \"1\"", rec1.Version)
	}
	if rec1.RegisteredAt.IsZero() || rec1.LastHeartbeat.IsZero() {
		t.Errorf("new record timestamps zero: registered=%v heartbeat=%v", rec1.RegisteredAt, rec1.LastHeartbeat)
	}

	// Re-register the same kind+name after the clock advanced: same id,
	// refreshed heartbeat.
	hb1 := rec1.LastHeartbeat
	clk.Current = hb1.Add(5 * time.Minute)
	rec2, err := s.Register(KindClaude, "lead", "")
	if err != nil {
		t.Fatalf("re-register: %v", err)
	}
	if rec2.AgentID != rec1.AgentID {
		t.Errorf("re-register returned different agentId: %q != %q", rec2.AgentID, rec1.AgentID)
	}
	if !rec2.LastHeartbeat.After(hb1) {
		t.Errorf("re-register did not refresh heartbeat: %v not after %v", rec2.LastHeartbeat, hb1)
	}
	if rec2.Description != "orchestrator" {
		t.Errorf("re-register with empty description dropped description: %q", rec2.Description)
	}

	// Distinct name or kind yields a distinct agent.
	rec3, err := s.Register(KindClaude, "worker", "")
	if err != nil {
		t.Fatalf("register other name: %v", err)
	}
	if rec3.AgentID == rec1.AgentID {
		t.Errorf("different name collided on agentId %q", rec1.AgentID)
	}
	rec4, err := s.Register(KindCodex, "lead", "")
	if err != nil {
		t.Fatalf("register other kind: %v", err)
	}
	if rec4.AgentID == rec1.AgentID {
		t.Errorf("different kind collided on agentId %q", rec1.AgentID)
	}
	if !strings.HasPrefix(rec4.AgentID, "codex-") {
		t.Errorf("codex agentId %q lacks kind prefix", rec4.AgentID)
	}

	// Invalid inputs are rejected before any state is written.
	if _, err := s.Register("gemini", "x", ""); err == nil {
		t.Errorf("invalid kind accepted")
	}
	if _, err := s.Register(KindClaude, "", ""); err == nil {
		t.Errorf("empty name accepted")
	}
}

// TestHeartbeatOnlineOffline verifies AC-CSM-014 (REQ-CSM-004): an agent past
// the offline threshold (config.DefaultSessionMsgAgentOfflineMinutes) reports
// online:false in ListAgents, and each of register, poll, and send
// (sender-side) refreshes the heartbeat back to online:true.
func TestHeartbeatOnlineOffline(t *testing.T) {
	root := t.TempDir()
	base := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	clk := &FakeClock{Current: base}
	s := NewStore(root, clk)

	recB, err := s.Register(KindCodex, "codex-worker", "")
	if err != nil {
		t.Fatalf("register B: %v", err)
	}
	recA, err := s.Register(KindClaude, "claude-lead", "")
	if err != nil {
		t.Fatalf("register A: %v", err)
	}

	advance := time.Duration(config.DefaultSessionMsgAgentOfflineMinutes+1) * time.Minute

	online := func(id string) bool {
		t.Helper()
		infos, err := s.ListAgents()
		if err != nil {
			t.Fatalf("list agents: %v", err)
		}
		for _, info := range infos {
			if info.AgentID == id {
				return info.Online
			}
		}
		t.Fatalf("agent %s not listed (got %d agents)", id, len(infos))
		return false
	}

	if !online(recB.AgentID) {
		t.Fatalf("freshly registered agent reported offline")
	}

	// 1) Cross the threshold: offline. Register restores online.
	clk.Current = base.Add(advance)
	if online(recB.AgentID) {
		t.Fatalf("agent past offline threshold reported online")
	}
	if _, err := s.Register(KindCodex, "codex-worker", ""); err != nil {
		t.Fatalf("re-register B: %v", err)
	}
	if !online(recB.AgentID) {
		t.Fatalf("register did not restore online")
	}

	// 2) Poll refreshes.
	clk.Current = clk.Current.Add(advance)
	if online(recB.AgentID) {
		t.Fatalf("expected offline before poll")
	}
	if _, err := s.Poll(recB.AgentID, nil); err != nil {
		t.Fatalf("poll B: %v", err)
	}
	if !online(recB.AgentID) {
		t.Fatalf("poll did not restore online")
	}

	// 3) Send (sender-side) refreshes.
	clk.Current = clk.Current.Add(advance)
	if online(recB.AgentID) {
		t.Fatalf("expected offline before send")
	}
	if _, err := s.Send(recB.AgentID, recA.AgentID, "ping", nil, "", ""); err != nil {
		t.Fatalf("send B->A: %v", err)
	}
	if !online(recB.AgentID) {
		t.Fatalf("send did not refresh sender heartbeat")
	}
}
