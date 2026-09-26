package cli

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Opt-in test against a MoAI-owned Codex TUI in an isolated project. The
// sender is a real, process-backed peer; no direct App Server call is made.
func TestManagedCodexFactoryBrokerLive(t *testing.T) {
	root, run := os.Getenv("MOAI_FACTORY_LIVE_ROOT"), os.Getenv("MOAI_FACTORY_LIVE_RUN")
	if root == "" || run == "" {
		t.Skip("set MOAI_FACTORY_LIVE_ROOT and MOAI_FACTORY_LIVE_RUN")
	}
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	target := os.Getenv("MOAI_FACTORY_LIVE_TO_SLOT")
	if target == "" {
		target = "lead"
	}
	to, err := s.ResolveLane(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	before, err := s.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sender, err := s.RegisterPeer(ctx, factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "codex", Role: "agent", Slot: "agent-99",
		SessionUUID: "live-probe-" + strconv.FormatInt(time.Now().UnixNano(), 10), Generation: 1,
		PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint(),
	})
	if err != nil {
		t.Fatal(err)
	}
	key := "managed-codex-live-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	msg, err := s.Send(ctx, factorymsg.SendRequest{From: sender, To: to, Kind: factorymsg.KindStatusRequest,
		IdempotencyKey: key, TaskRef: "t-live", CorrelationID: key, TTL: 2 * time.Minute,
		Payload: []byte("LIVE_MANAGED_CODEX_DELIVERY. Read this message via factory_msg_body and record an accepted factory_msg_receipt. Do not edit files.")})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("sent message_id=%s to %s", msg.ID, to.SessionUUID)
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		status, err := s.Status(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if status.Acknowledged > before.Acknowledged {
			t.Logf("acknowledged=%d pending=%d claimed=%d", status.Acknowledged, status.Pending, status.Claimed)
			return
		}
		time.Sleep(time.Second)
	}
	status, _ := s.Status(ctx)
	t.Fatalf("receipt not observed: %+v", status)
}

func TestManagedCodexFactoryStatusLive(t *testing.T) {
	root, run := os.Getenv("MOAI_FACTORY_LIVE_ROOT"), os.Getenv("MOAI_FACTORY_LIVE_RUN")
	if root == "" || run == "" {
		t.Skip("set MOAI_FACTORY_LIVE_ROOT and MOAI_FACTORY_LIVE_RUN")
	}
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("capability=%s acknowledged=%d pending=%d claimed=%d lanes=%d", status.Capability, status.Acknowledged, status.Pending, status.Claimed, len(status.Lanes))
}
