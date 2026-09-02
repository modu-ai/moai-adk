package sessionmsg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestSendPollAck verifies AC-CSM-003 (REQ-CSM-005/006): send creates the
// pending envelope file, poll atomically claims it into claimed/ and returns
// it, and ack_ids deletes the claimed copy. Structured errors carry the
// known-agent list for unregistered counterparties.
func TestSendPollAck(t *testing.T) {
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
	if msgID == "" {
		t.Fatal("send returned empty messageId")
	}
	pendingPath := filepath.Join(root, "mailbox", receiver.AgentID, "pending", msgID+".json")
	if _, err := os.Stat(pendingPath); err != nil {
		t.Fatalf("pending envelope file missing: %v", err)
	}

	// Poll claims exactly the sent message.
	res, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("poll returned %d messages, want 1", len(res.Messages))
	}
	if res.Messages[0].Message.MessageID != msgID {
		t.Errorf("polled messageId %q != sent %q", res.Messages[0].Message.MessageID, msgID)
	}
	if res.Messages[0].Message.Parts[0].Kind != PartKindText || res.Messages[0].Message.Parts[0].Text != "hello" {
		t.Errorf("polled content mismatch: %+v", res.Messages[0].Message.Parts)
	}
	if res.Messages[0].Delivery.SenderID != sender.AgentID {
		t.Errorf("polled senderId %q != %q", res.Messages[0].Delivery.SenderID, sender.AgentID)
	}
	if res.Messages[0].Delivery.ClaimedAt == nil {
		t.Errorf("polled envelope has nil claimedAt")
	}
	if res.Remaining != 0 {
		t.Errorf("poll remaining = %d, want 0", res.Remaining)
	}
	if res.ExpiredCount != 0 {
		t.Errorf("poll expiredCount = %d, want 0", res.ExpiredCount)
	}
	claimedPath := filepath.Join(root, "mailbox", receiver.AgentID, "claimed", msgID+".json")
	if _, err := os.Stat(claimedPath); err != nil {
		t.Fatalf("claimed envelope file missing after poll: %v", err)
	}
	if _, err := os.Stat(pendingPath); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("pending file still present after claim (err=%v)", err)
	}

	// Ack deletes the claimed copy.
	res2, err := s.Poll(receiver.AgentID, []string{msgID})
	if err != nil {
		t.Fatalf("poll with ack: %v", err)
	}
	if res2.AckedCount != 1 {
		t.Errorf("ack count = %d, want 1", res2.AckedCount)
	}
	if _, err := os.Stat(claimedPath); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("claimed file still present after ack (err=%v)", err)
	}
	if len(res2.Messages) != 0 {
		t.Errorf("ack poll unexpectedly delivered %d messages", len(res2.Messages))
	}

	// Empty mailbox polls cleanly.
	res3, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("empty poll: %v", err)
	}
	if len(res3.Messages) != 0 || res3.Remaining != 0 || res3.AckedCount != 0 {
		t.Errorf("empty poll not clean: %+v", res3)
	}

	// Text + data + context/task ids round trip as a two-part message.
	msgID2, err := s.Send(sender.AgentID, receiver.AgentID, "body", json.RawMessage(`{"k":1}`), "ctx-7", "task-9")
	if err != nil {
		t.Fatalf("send with data: %v", err)
	}
	res4, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll data message: %v", err)
	}
	if len(res4.Messages) != 1 || res4.Messages[0].Message.MessageID != msgID2 {
		t.Fatalf("data message not delivered exactly once: %+v", res4.Messages)
	}
	m := res4.Messages[0].Message
	if len(m.Parts) != 2 || m.Parts[0].Kind != PartKindText || m.Parts[1].Kind != PartKindData {
		t.Errorf("two-part shape mismatch: %+v", m.Parts)
	}
	if m.ContextID != "ctx-7" || m.TaskID != "task-9" {
		t.Errorf("context/task ids lost: ctx=%q task=%q", m.ContextID, m.TaskID)
	}

	// Unknown recipient → structured error carrying the known-agent list.
	_, err = s.Send(sender.AgentID, "codex-00000000", "x", nil, "", "")
	if err == nil {
		t.Fatal("send to unknown recipient accepted")
	}
	var unknown *UnknownAgentError
	if !errors.As(err, &unknown) {
		t.Errorf("send error is not UnknownAgentError: %v", err)
	} else {
		found := false
		for _, a := range unknown.Known {
			if a.AgentID == receiver.AgentID || a.AgentID == sender.AgentID {
				found = true
			}
		}
		if !found {
			t.Errorf("known-agents list omits registered agents: %+v", unknown.Known)
		}
		if unknown.AgentID != "codex-00000000" {
			t.Errorf("unknown-agent id mismatch: %q", unknown.AgentID)
		}
	}

	// Unknown sender is rejected too.
	if _, err := s.Send("claude-00000000", receiver.AgentID, "x", nil, "", ""); err == nil {
		t.Error("send from unknown sender accepted")
	}
	// Polling as an unknown agent is rejected.
	if _, err := s.Poll("codex-00000000", nil); err == nil {
		t.Error("poll as unknown agent accepted")
	}

	// Oversize text is rejected at validation (threshold override proves the
	// bound is live, not vestigial). The override lives in a closure with a
	// deferred restore: an inline restore is skipped by any early t.Fatal
	// (which runs deferred functions but not the following statements), and
	// the shortened ceiling would then leak into every later test in the
	// package. The closure also restores immediately rather than at test end,
	// so the assertions below still run against the production ceiling.
	func() {
		origMax := config.DefaultSessionMsgMaxTextBytes
		defer func() { config.DefaultSessionMsgMaxTextBytes = origMax }()
		config.DefaultSessionMsgMaxTextBytes = 8
		if _, err := s.Send(sender.AgentID, receiver.AgentID, "123456789", nil, "", ""); err == nil {
			t.Error("oversize text accepted")
		}
	}()

	// Neither text nor data yields no parts → rejected.
	if _, err := s.Send(sender.AgentID, receiver.AgentID, "", nil, "", ""); err == nil {
		t.Error("empty message accepted")
	}

	// Invalid JSON data is rejected at validation.
	if _, err := s.Send(sender.AgentID, receiver.AgentID, "t", json.RawMessage(`{not-json`), "", ""); err == nil {
		t.Error("invalid JSON data accepted")
	}
}

// TestClaimExpiryReturn verifies AC-CSM-004 (REQ-CSM-007): a claimed message
// whose claim exceeded the claim TTL is returned to pending by the next
// sweep and re-delivered by the same poll (at-least-once).
func TestClaimExpiryReturn(t *testing.T) {
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

	msgID, err := s.Send(sender.AgentID, receiver.AgentID, "hello", nil, "", "")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	res, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("first poll: %v", err)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("first poll delivered %d messages, want 1", len(res.Messages))
	}

	// ClaimedAt = base; advance 11 minutes (claim TTL is 10m by default).
	clk.Current = base.Add(config.DefaultSessionMsgClaimTTL + time.Minute)
	res2, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("second poll: %v", err)
	}
	if len(res2.Messages) != 1 {
		t.Fatalf("claim-expired message not re-delivered: got %d messages", len(res2.Messages))
	}
	if res2.Messages[0].Message.MessageID != msgID {
		t.Errorf("re-delivered messageId %q != %q", res2.Messages[0].Message.MessageID, msgID)
	}
	if res2.Messages[0].Message.Parts[0].Text != "hello" {
		t.Errorf("re-delivered content mismatch: %q", res2.Messages[0].Message.Parts[0].Text)
	}
	// The re-claim stamps a fresh ClaimedAt at the advanced time.
	if res2.Messages[0].Delivery.ClaimedAt == nil || !res2.Messages[0].Delivery.ClaimedAt.Equal(clk.Current) {
		t.Errorf("re-claim did not refresh claimedAt: %+v", res2.Messages[0].Delivery.ClaimedAt)
	}
}

// TestMessageExpirySweep verifies AC-CSM-005 (REQ-CSM-008): messages past
// their message TTL are lazily deleted by the next broker call, from both
// pending and claimed, and poll reports them via ExpiredCount.
func TestMessageExpirySweep(t *testing.T) {
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

	// Phase A: a PENDING message past the message TTL is deleted and counted.
	msgID1, err := s.Send(sender.AgentID, receiver.AgentID, "expiring-pending", nil, "", "")
	if err != nil {
		t.Fatalf("send 1: %v", err)
	}
	clk.Current = base.Add(config.DefaultSessionMsgMessageTTL + time.Minute)
	res, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll A: %v", err)
	}
	if len(res.Messages) != 0 {
		t.Errorf("expired pending message was delivered: %+v", res.Messages)
	}
	if res.ExpiredCount != 1 {
		t.Errorf("poll A expiredCount = %d, want 1", res.ExpiredCount)
	}
	if _, err := os.Stat(filepath.Join(root, "mailbox", receiver.AgentID, "pending", msgID1+".json")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expired pending file not deleted (err=%v)", err)
	}

	// Phase B: a CLAIMED message past the message TTL is deleted and counted.
	t1 := clk.Current
	msgID2, err := s.Send(sender.AgentID, receiver.AgentID, "expiring-claimed", nil, "", "")
	if err != nil {
		t.Fatalf("send 2: %v", err)
	}
	res2, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll B claim: %v", err)
	}
	if len(res2.Messages) != 1 || res2.Messages[0].Message.MessageID != msgID2 {
		t.Fatalf("claim before expiry failed: %+v", res2.Messages)
	}
	clk.Current = t1.Add(config.DefaultSessionMsgMessageTTL + time.Minute)
	res3, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll B sweep: %v", err)
	}
	if len(res3.Messages) != 0 {
		t.Errorf("expired claimed message was re-delivered: %+v", res3.Messages)
	}
	if res3.ExpiredCount != 1 {
		t.Errorf("poll B expiredCount = %d, want 1 (this call's deletions only)", res3.ExpiredCount)
	}
	if _, err := os.Stat(filepath.Join(root, "mailbox", receiver.AgentID, "claimed", msgID2+".json")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expired claimed file not deleted (err=%v)", err)
	}
}

// TestConcurrentSendPoll verifies AC-CSM-006 (REQ-CSM-009): 10 concurrent
// senders each send 10 messages to one receiver while 2 pollers consume —
// exactly 100 unique deliveries, zero lost, zero duplicated. Run under
// -race by the acceptance command.
func TestConcurrentSendPoll(t *testing.T) {
	root := t.TempDir()
	// Real clock: a FakeClock shared across goroutines must not be mutated,
	// and this test never advances time, so the real clock keeps Now()
	// race-free while timestamps stay realistic.
	s := NewStore(root, nil).WithLockTimeout(30 * time.Second)

	receiver, err := s.Register(KindCodex, "hub", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}
	const senderCount = 10
	const perSender = 10
	const total = senderCount * perSender
	senders := make([]AgentRecord, senderCount)
	for i := range senders {
		senders[i], err = s.Register(KindClaude, fmt.Sprintf("sender-%d", i), "")
		if err != nil {
			t.Fatalf("register sender %d: %v", i, err)
		}
	}

	var wgSend, wgPoll sync.WaitGroup
	sendersDone := make(chan struct{})
	collected := make(chan []string, 2)

	for i, snd := range senders {
		wgSend.Add(1)
		go func(idx int, from AgentRecord) {
			defer wgSend.Done()
			for j := 0; j < perSender; j++ {
				text := fmt.Sprintf("m-%d-%d", idx, j)
				if _, err := s.Send(from.AgentID, receiver.AgentID, text, nil, "", ""); err != nil {
					t.Errorf("send %s: %v", text, err)
					return
				}
			}
		}(i, snd)
	}
	for p := 0; p < 2; p++ {
		wgPoll.Add(1)
		go func() {
			defer wgPoll.Done()
			// The stop rule lives in pollUntilDrained (stoprule_test.go) so
			// TestPollerStopRule polices the same code this test runs.
			collected <- pollUntilDrained(t, s, receiver.AgentID, sendersDone, nil)
		}()
	}
	wgSend.Wait()
	close(sendersDone)
	wgPoll.Wait()
	close(collected)

	var all []string
	for ids := range collected {
		all = append(all, ids...)
	}
	if len(all) != total {
		t.Fatalf("received %d messages, want %d (loss or duplication)", len(all), total)
	}
	uniq := make(map[string]bool, total)
	for _, id := range all {
		if uniq[id] {
			t.Errorf("duplicate delivery of message %s", id)
		}
		uniq[id] = true
	}
	if len(uniq) != total {
		t.Fatalf("%d unique messages, want %d", len(uniq), total)
	}
}

// TestPollBatchCeiling pins config.DefaultSessionMsgPollBatch as the
// per-poll claim ceiling (REQ-CSM-006): one poll claims exactly the ceiling
// and reports the rest as Remaining; the next poll drains the remainder.
//
// Mutant this test catches that a shallower one would not: an implementation
// that claims everything and merely REPORTS Remaining correctly (or reports
// 0) passes any assertion made only on len(res.Messages) <= batch or only on
// Remaining. Both the claimed count AND the exact Remaining are pinned on
// both polls, and the second poll's batch is checked for the exact overflow.
func TestPollBatchCeiling(t *testing.T) {
	root := t.TempDir()
	clk := &FakeClock{Current: time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)

	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}

	batch := config.DefaultSessionMsgPollBatch
	overflow := 3
	total := batch + overflow
	sent := make(map[string]bool, total)
	for i := 0; i < total; i++ {
		id, err := s.Send(sender.AgentID, receiver.AgentID, fmt.Sprintf("m%d", i), nil, "", "")
		if err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
		sent[id] = true
	}

	first, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("first poll: %v", err)
	}
	if len(first.Messages) != batch {
		t.Errorf("first poll claimed %d messages, want exactly the ceiling %d", len(first.Messages), batch)
	}
	if first.Remaining != overflow {
		t.Errorf("first poll Remaining = %d, want %d", first.Remaining, overflow)
	}

	// Everything claimed must actually have left pending: the second poll
	// sees only the overflow.
	second, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("second poll: %v", err)
	}
	if len(second.Messages) != overflow {
		t.Errorf("second poll claimed %d messages, want the %d-message overflow", len(second.Messages), overflow)
	}
	if second.Remaining != 0 {
		t.Errorf("second poll Remaining = %d, want 0", second.Remaining)
	}

	seen := map[string]bool{}
	for _, env := range append(append([]Envelope{}, first.Messages...), second.Messages...) {
		id := env.Message.MessageID
		if seen[id] {
			t.Errorf("message %s delivered twice across the two polls", id)
		}
		seen[id] = true
		if !sent[id] {
			t.Errorf("polled unknown messageId %s", id)
		}
	}
	if len(seen) != total {
		t.Errorf("drained %d distinct messages, want %d", len(seen), total)
	}
}
