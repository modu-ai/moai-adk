package sessionmsg

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// pollUntilDrained is the TestConcurrentSendPoll poller loop, extracted so
// the stop rule lives in exactly one place: the deterministic reproduction
// test below drives THIS loop through the fatal interleaving, so a regression
// of the rule fails the reproduction test too, not only the concurrent test
// that cannot open the window on a local scheduler.
//
// beforeStopCheck, when non-nil, runs after an empty-drain observation and
// before the sendersDone check — the exact span the CI race exploited.
// TestConcurrentSendPoll passes nil; the reproduction test holds the poller
// there while the final Sends land.
func pollUntilDrained(t *testing.T, s *Store, agentID string, sendersDone <-chan struct{}, beforeStopCheck func()) []string {
	t.Helper()
	var ids []string
	stop := false
	deadline := time.Now().Add(60 * time.Second)
	for !stop && time.Now().Before(deadline) {
		res, err := s.Poll(agentID, nil)
		if err != nil {
			t.Errorf("poll: %v", err)
			break
		}
		for _, m := range res.Messages {
			ids = append(ids, m.Message.MessageID)
		}
		if len(res.Messages) == 0 && res.Remaining == 0 {
			// Drain reached the send frontier: terminate only on an empty
			// observation made by a poll STARTED after every sender finished
			// (REQ-CFS-001). The observation above and the sendersDone check
			// below are two different instants: close(sendersDone) runs after
			// wgSend.Wait(), i.e. after every Send returned, but the empty
			// observation may predate the final Sends' pending renames — the
			// window behind the 97/100 CI loss (run 32774108273 a1: 3 lost,
			// zero poll errors, 0.07s). So a closed channel here triggers one
			// re-poll, and only ITS 0/0 terminates; anything it claims joins
			// the collected set.
			if beforeStopCheck != nil {
				beforeStopCheck()
			}
			select {
			case <-sendersDone:
				res2, err := s.Poll(agentID, nil)
				if err != nil {
					t.Errorf("poll: %v", err)
					break
				}
				for _, m := range res2.Messages {
					ids = append(ids, m.Message.MessageID)
				}
				if len(res2.Messages) == 0 && res2.Remaining == 0 {
					stop = true
				}
			default:
				time.Sleep(2 * time.Millisecond)
			}
		}
	}
	return ids
}

// TestPollerStopRule reproduces the fatal TestConcurrentSendPoll interleaving
// deterministically (REQ-CFS-002): the poller observes an empty drain, the
// final 3 Sends land and close sendersDone AFTER that observation but BEFORE
// the poller's channel check, and the stop rule must not treat the stale
// empty observation as terminal. CI hit this window by scheduling (run
// 32774108273 attempt 1: 97/100 loss in 0.07s with zero poll errors); the
// channel handshake here forces the same order without sleeps.
func TestPollerStopRule(t *testing.T) {
	root := t.TempDir()
	// Real clock, as in TestConcurrentSendPoll: time never advances here and
	// Now() stays race-free.
	s := NewStore(root, nil).WithLockTimeout(30 * time.Second)

	receiver, err := s.Register(KindCodex, "hub", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}
	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}

	const total = 100
	const heldBack = 3

	// The 97 messages the CI race straddled: fully sent and claimable BEFORE
	// the poller's empty observation.
	for i := 0; i < total-heldBack; i++ {
		if _, err := s.Send(sender.AgentID, receiver.AgentID, fmt.Sprintf("m-%d", i), nil, "", ""); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	sawEmpty := make(chan struct{})
	doneClosed := make(chan struct{})
	sendersDone := make(chan struct{})
	collected := make(chan []string, 1)

	// Poller: drain to the empty observation, then hold at the gate — after
	// 0/0, before the sendersDone check — until the sender side finishes the
	// last 3 Sends and closes the channel.
	var gateOnce sync.Once
	go func() {
		collected <- pollUntilDrained(t, s, receiver.AgentID, sendersDone, func() {
			gateOnce.Do(func() { close(sawEmpty) })
			<-doneClosed
		})
	}()

	select {
	case <-sawEmpty:
	case <-time.After(30 * time.Second):
		t.Fatal("poller never observed an empty drain — reproduction premise broken")
	}
	// Sender side: the final 3 Sends and the close, all completing inside the
	// observation-to-check window.
	for i := total - heldBack; i < total; i++ {
		if _, err := s.Send(sender.AgentID, receiver.AgentID, fmt.Sprintf("m-%d", i), nil, "", ""); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}
	close(sendersDone)
	close(doneClosed)

	ids := <-collected
	if len(ids) != total {
		t.Fatalf("received %d messages, want %d (stop rule treated a pre-completion empty observation as terminal)", len(ids), total)
	}
	uniq := make(map[string]bool, total)
	for _, id := range ids {
		if uniq[id] {
			t.Errorf("duplicate delivery of message %s", id)
		}
		uniq[id] = true
	}
	if len(uniq) != total {
		t.Fatalf("%d unique messages, want %d", len(uniq), total)
	}
}
