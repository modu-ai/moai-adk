package sessionmsg

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// pendingCount counts the envelope files in one agent's pending mailbox.
func pendingCount(t *testing.T, root, agentID string) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "mailbox", agentID, "pending"))
	if err != nil {
		t.Fatalf("read pending dir: %v", err)
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			n++
		}
	}
	return n
}

// TestSendRejectsAtPendingDepthCap covers the pending-mailbox depth ceiling
// (card t253, PR #1606 review): filling the recipient's mailbox to
// config.DefaultSessionMsgMaxPending succeeds, the next send is rejected
// with a structured MailboxFullError and writes nothing, and polling —
// which claims pending envelopes out of pending/ — frees depth so sends
// succeed again.
func TestSendRejectsAtPendingDepthCap(t *testing.T) {
	root := t.TempDir()
	clk := &FakeClock{Current: time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)

	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}

	// Fill to the ceiling: cap sends succeed.
	for i := 0; i < config.DefaultSessionMsgMaxPending; i++ {
		if _, err := s.Send(sender.AgentID, receiver.AgentID, "fill", nil, "", ""); err != nil {
			t.Fatalf("send %d/%d: %v", i+1, config.DefaultSessionMsgMaxPending, err)
		}
	}
	if got := pendingCount(t, root, receiver.AgentID); got != config.DefaultSessionMsgMaxPending {
		t.Fatalf("pending count at fill: got %d, want %d", got, config.DefaultSessionMsgMaxPending)
	}

	// The send at the ceiling is rejected with the structured error, and no
	// messageId comes back.
	overflowID, err := s.Send(sender.AgentID, receiver.AgentID, "overflow", nil, "", "")
	var full *MailboxFullError
	if !errors.As(err, &full) {
		t.Fatalf("send at cap: got %v, want *MailboxFullError", err)
	}
	if overflowID != "" {
		t.Errorf("rejected send returned messageId %q, want empty", overflowID)
	}
	if full.AgentID != receiver.AgentID {
		t.Errorf("MailboxFullError.AgentID: got %q, want %q", full.AgentID, receiver.AgentID)
	}
	if full.Limit != config.DefaultSessionMsgMaxPending {
		t.Errorf("MailboxFullError.Limit: got %d, want %d", full.Limit, config.DefaultSessionMsgMaxPending)
	}

	// The rejection wrote nothing: the mailbox stays exactly at the ceiling.
	if got := pendingCount(t, root, receiver.AgentID); got != config.DefaultSessionMsgMaxPending {
		t.Errorf("pending count after rejection: got %d, want %d (rejected send must not write)", got, config.DefaultSessionMsgMaxPending)
	}

	// Claiming frees depth: poll moves up to PollBatch envelopes out of
	// pending/, after which the mailbox accepts new sends.
	res, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(res.Messages) != config.DefaultSessionMsgPollBatch {
		t.Fatalf("poll claimed %d messages, want %d", len(res.Messages), config.DefaultSessionMsgPollBatch)
	}
	if _, err := s.Send(sender.AgentID, receiver.AgentID, "after-drain", nil, "", ""); err != nil {
		t.Fatalf("send after poll: %v", err)
	}
}

// TestMailboxFullErrorOtherRecipientUnaffected verifies the ceiling is
// per-recipient: filling one agent's mailbox does not block sends to a
// different agent whose mailbox is empty.
func TestMailboxFullErrorOtherRecipientUnaffected(t *testing.T) {
	root := t.TempDir()
	clk := &FakeClock{Current: time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)

	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	blocked, err := s.Register(KindCodex, "blocked", "")
	if err != nil {
		t.Fatalf("register blocked: %v", err)
	}
	other, err := s.Register(KindCodex, "other", "")
	if err != nil {
		t.Fatalf("register other: %v", err)
	}

	for i := 0; i < config.DefaultSessionMsgMaxPending; i++ {
		if _, err := s.Send(sender.AgentID, blocked.AgentID, "fill", nil, "", ""); err != nil {
			t.Fatalf("send %d: %v", i+1, err)
		}
	}
	if _, err := s.Send(sender.AgentID, other.AgentID, "unaffected", nil, "", ""); err != nil {
		t.Fatalf("send to other recipient: %v", err)
	}
	if _, err := s.Send(sender.AgentID, blocked.AgentID, "overflow", nil, "", ""); err == nil {
		t.Fatal("send to full mailbox: want error, got nil")
	}
}
