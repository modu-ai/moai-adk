package loop

import (
	"testing"
)

// TestFeedbackChannel_SendAndReceive verifies REQ-LL-009:
// NewFeedbackChannel creates a bounded channel that can send and receive Feedback.
func TestFeedbackChannel_SendAndReceive(t *testing.T) {
	t.Parallel()

	ch := NewFeedbackChannel(4)
	if ch == nil {
		t.Fatal("NewFeedbackChannel() returned nil")
	}

	fb := Feedback{TestsFailed: 1, LintErrors: 2}
	ch.Send(fb)

	got, ok := ch.TryReceive()
	if !ok {
		t.Fatal("TryReceive() returned false, expected a message")
	}
	if got.TestsFailed != fb.TestsFailed {
		t.Errorf("TestsFailed = %d, want %d", got.TestsFailed, fb.TestsFailed)
	}
}

// TestFeedbackChannel_BoundedDropOldest verifies REQ-LL-009:
// when the channel is full, Send drops the oldest event and warns.
func TestFeedbackChannel_BoundedDropOldest(t *testing.T) {
	t.Parallel()

	capacity := 2
	ch := NewFeedbackChannel(capacity)

	// Fill channel to capacity.
	ch.Send(Feedback{TestsFailed: 1})
	ch.Send(Feedback{TestsFailed: 2})

	// This send must drop the oldest (TestsFailed=1) and insert TestsFailed=3.
	ch.Send(Feedback{TestsFailed: 3})

	// First receive should be the second item (oldest dropped).
	first, ok := ch.TryReceive()
	if !ok {
		t.Fatal("TryReceive() returned false after overflow send")
	}
	if first.TestsFailed != 2 {
		t.Errorf("after overflow, first item TestsFailed = %d, want 2 (oldest dropped)", first.TestsFailed)
	}

	second, ok := ch.TryReceive()
	if !ok {
		t.Fatal("TryReceive() returned false for second item")
	}
	if second.TestsFailed != 3 {
		t.Errorf("second item TestsFailed = %d, want 3", second.TestsFailed)
	}
}

// TestFeedbackChannel_TryReceive_Empty verifies TryReceive on empty channel returns false.
func TestFeedbackChannel_TryReceive_Empty(t *testing.T) {
	t.Parallel()

	ch := NewFeedbackChannel(4)
	_, ok := ch.TryReceive()
	if ok {
		t.Error("TryReceive() on empty channel must return false")
	}
}

// TestFeedbackChannel_Chan returns the underlying channel for select statements.
func TestFeedbackChannel_Chan(t *testing.T) {
	t.Parallel()

	ch := NewFeedbackChannel(2)
	if ch.Chan() == nil {
		t.Error("Chan() must return a non-nil channel")
	}
}

// TestFeedbackChannel_Len verifies the Len() method returns correct count.
func TestFeedbackChannel_Len(t *testing.T) {
	t.Parallel()

	ch := NewFeedbackChannel(4)
	if ch.Len() != 0 {
		t.Errorf("Len() = %d, want 0 for empty channel", ch.Len())
	}

	ch.Send(Feedback{TestsFailed: 1})
	ch.Send(Feedback{TestsFailed: 2})
	if ch.Len() != 2 {
		t.Errorf("Len() = %d, want 2 after two sends", ch.Len())
	}
}
