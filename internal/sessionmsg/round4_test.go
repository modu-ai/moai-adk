package sessionmsg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// skipIfDirModeIneffective skips a test that relies on a directory's mode
// bits actually gating file creation inside it: Windows does not enforce the
// POSIX bits os.Chmod sets, and root bypasses them everywhere.
func skipIfDirModeIneffective(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("directory mode bits do not gate file creation on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission bits")
	}
}

// freezeAgentsDir makes agents/ read+execute only: records stay readable (so
// the registration checks in Send/Poll still pass) but writeJSONAtomic's
// CreateTemp in that directory fails, which is what makes the heartbeat error.
func freezeAgentsDir(t *testing.T, root string) {
	t.Helper()
	agentsDir := filepath.Join(root, "agents")
	if err := os.Chmod(agentsDir, 0o500); err != nil {
		t.Fatalf("chmod agents dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(agentsDir, 0o755) })
}

// TestPollReturnsClaimWhenHeartbeatFails verifies that a heartbeat write
// failure does not destroy an already-committed claim. By the time the
// heartbeat runs, the pending→claimed move is on disk with a ClaimedAt stamp;
// returning an empty PollResult would hide those messages until the claim TTL
// redeems them.
//
// Mutant this test catches that a shallower one would not: an implementation
// that keeps `return PollResult{}, err` but merely *logs* the heartbeat
// failure still passes a test asserting only "Poll returns no error" — it is
// the returned batch, not the error, that carries the regression. So the
// assertion is on len(res.Messages), and the claimed file's existence pins
// that the claim really was committed before the heartbeat ran.
func TestPollReturnsClaimWhenHeartbeatFails(t *testing.T) {
	skipIfDirModeIneffective(t)

	root := t.TempDir()
	clk := &FakeClock{Current: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)

	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}
	msgID, err := s.Send(sender.AgentID, receiver.AgentID, "hello", nil, "", "")
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	freezeAgentsDir(t, root)

	// Precondition: the heartbeat really does fail under this mode, so a
	// passing assertion below cannot come from the failure never happening.
	if err := s.heartbeat(receiver.AgentID); err == nil {
		t.Fatal("heartbeat unexpectedly succeeded on a read-only agents dir")
	}

	res, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll failed on a heartbeat error: %v", err)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("poll delivered %d messages, want 1 (claim discarded by heartbeat failure)", len(res.Messages))
	}
	if res.Messages[0].Message.MessageID != msgID {
		t.Errorf("polled messageId %q, want %q", res.Messages[0].Message.MessageID, msgID)
	}
	if _, err := os.Stat(filepath.Join(root, "mailbox", receiver.AgentID, "claimed", msgID+".json")); err != nil {
		t.Errorf("claimed envelope missing after poll: %v", err)
	}
}

// TestSendReturnsMessageIDWhenHeartbeatFails is the Send-side twin: the
// envelope is already written to the recipient's pending mailbox when the
// sender-side heartbeat runs, so returning ("", err) loses the id of a message
// that WAS delivered — and a caller retrying on that error double-sends.
//
// Mutant this test catches: an implementation that returns the messageId but
// also a non-nil error still fails here, because the assertion requires
// err == nil (the MCP handler discards the result on any error).
func TestSendReturnsMessageIDWhenHeartbeatFails(t *testing.T) {
	skipIfDirModeIneffective(t)

	root := t.TempDir()
	clk := &FakeClock{Current: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)

	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}

	freezeAgentsDir(t, root)

	msgID, err := s.Send(sender.AgentID, receiver.AgentID, "hello", nil, "", "")
	if err != nil {
		t.Fatalf("send failed on a heartbeat error: %v", err)
	}
	if msgID == "" {
		t.Fatal("send returned an empty messageId for a delivered message")
	}
	if _, err := os.Stat(filepath.Join(root, "mailbox", receiver.AgentID, "pending", msgID+".json")); err != nil {
		t.Errorf("pending envelope missing after send: %v", err)
	}
}

// TestPollClaimsOldestFirst verifies that the claim batch is the OLDEST N by
// Delivery.SentAt, in ascending order. Envelope filenames are msg-<random
// hex>, so os.ReadDir's lexical order is effectively random: a mailbox holding
// more than the batch ceiling would otherwise return a random subset.
// REQ-CSM-006 fixes the ceiling but is SILENT on ordering — FIFO is a
// defensible interpretation, not a stated requirement.
//
// Mutant this test catches that a shallower one would not: an implementation
// that sorts DESCENDING by SentAt still returns exactly `batch` messages and
// still drains the mailbox over repeated polls, so a test asserting only the
// count passes. The assertion pins both the identity of the claimed subset
// (oldest N) and its order.
func TestPollClaimsOldestFirst(t *testing.T) {
	root := t.TempDir()
	base := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	clk := &FakeClock{Current: base}
	s := NewStore(root, clk)

	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}

	origBatch := config.DefaultSessionMsgPollBatch
	config.DefaultSessionMsgPollBatch = 3
	defer func() { config.DefaultSessionMsgPollBatch = origBatch }()

	const total = 8
	sentOrder := make([]string, 0, total)
	for i := 0; i < total; i++ {
		clk.Current = base.Add(time.Duration(i) * time.Minute)
		id, err := s.Send(sender.AgentID, receiver.AgentID, fmt.Sprintf("body-%d", i), nil, "", "")
		if err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
		sentOrder = append(sentOrder, id)
	}
	clk.Current = base.Add(time.Hour)

	res, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(res.Messages) != 3 {
		t.Fatalf("poll claimed %d messages, want 3", len(res.Messages))
	}
	for i, env := range res.Messages {
		if env.Message.MessageID != sentOrder[i] {
			t.Errorf("claim[%d] = %q, want %q (oldest-first)", i, env.Message.MessageID, sentOrder[i])
		}
		if i > 0 && env.Delivery.SentAt.Before(res.Messages[i-1].Delivery.SentAt) {
			t.Errorf("claim[%d] sentAt %v precedes claim[%d] %v — not ascending",
				i, env.Delivery.SentAt, i-1, res.Messages[i-1].Delivery.SentAt)
		}
	}
	if res.Remaining != total-3 {
		t.Errorf("remaining = %d, want %d", res.Remaining, total-3)
	}
}

// TestDataPartSizeCeiling verifies that data parts are bounded by
// config.DefaultSessionMsgMaxDataBytes. REQ-CSM-005 requires the broker to
// validate the body size ceiling, and a data part IS body content — the text
// ceiling alone leaves an arbitrarily large JSON payload validating and being
// persisted.
//
// Mutant this test catches that a shallower one would not: an implementation
// that bounds data with `>=` instead of `>` rejects an exactly-at-ceiling
// payload, so the under-ceiling case is asserted alongside the over-ceiling
// one and the boundary is pinned at one byte on each side.
func TestDataPartSizeCeiling(t *testing.T) {
	orig := config.DefaultSessionMsgMaxDataBytes
	config.DefaultSessionMsgMaxDataBytes = 64
	defer func() { config.DefaultSessionMsgMaxDataBytes = orig }()

	// A JSON string literal of exactly n bytes: 2 quotes + (n-2) filler.
	payload := func(n int) json.RawMessage {
		return json.RawMessage(`"` + strings.Repeat("x", n-2) + `"`)
	}

	atCeiling := Message{
		MessageID: "msg-0000000000000001",
		Role:      RoleAgent,
		Parts:     []Part{{Kind: PartKindData, Data: payload(64)}},
	}
	if err := atCeiling.Validate(); err != nil {
		t.Errorf("data payload exactly at the ceiling rejected: %v", err)
	}

	overCeiling := Message{
		MessageID: "msg-0000000000000002",
		Role:      RoleAgent,
		Parts:     []Part{{Kind: PartKindData, Data: payload(65)}},
	}
	err := overCeiling.Validate()
	if err == nil {
		t.Fatal("data payload one byte over the ceiling accepted")
	}
	if !strings.Contains(err.Error(), "data size") {
		t.Errorf("unexpected rejection reason: %v", err)
	}
}
