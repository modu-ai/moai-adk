package factorylane

import (
	"errors"
	"testing"
	"time"
)

// REQ-FLA-003 / AC-FLA-003: one mode switch records exactly one transition
// event carrying lane id, trigger kind, timestamp, and the in-progress card
// id.
func TestDeclareFallbackRecordsExactlyOneEventWithAllFields(t *testing.T) {
	store, clock := newTestStore(t)
	clock.Current = base.Add(time.Minute)
	ev, err := store.DeclareFallback("lane-1", TriggerChannelUnavailable, "t12")
	if err != nil {
		t.Fatalf("DeclareFallback: %v", err)
	}
	if ev.Lane != "lane-1" || ev.Trigger != TriggerChannelUnavailable || ev.Card != "t12" {
		t.Fatalf("event fields = %+v, want lane-1/channel-unavailable/t12", ev)
	}
	if !ev.At.Equal(clock.Current) {
		t.Fatalf("event At = %s, want %s", ev.At, clock.Current)
	}
	events, err := store.Transitions("lane-1")
	if err != nil {
		t.Fatalf("Transitions: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("transition count = %d, want exactly 1 for one mode switch", len(events))
	}
	if events[0] != ev {
		t.Fatalf("persisted event %+v != returned %+v", events[0], ev)
	}
}

// REQ-FLA-003: while the lane is already in fallback (latest event is a
// fallback activation without a restore), a second declare is not a new mode
// switch — it is refused and the count stays at exactly one.
func TestSecondDeclareWhileFallbackActiveIsRefused(t *testing.T) {
	store, clock := newTestStore(t)
	clock.Current = base.Add(time.Minute)
	if _, err := store.DeclareFallback("lane-1", TriggerChannelUnavailable, "t12"); err != nil {
		t.Fatalf("first DeclareFallback: %v", err)
	}
	clock.Current = base.Add(2 * time.Minute)
	_, err := store.DeclareFallback("lane-1", TriggerNoResponse, "t12")
	if !errors.Is(err, ErrFallbackActive) {
		t.Fatalf("second DeclareFallback error = %v, want ErrFallbackActive", err)
	}
	events, err := store.Transitions("lane-1")
	if err != nil {
		t.Fatalf("Transitions: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("transition count = %d, want 1 (the refused declare must write nothing)", len(events))
	}
}

// REQ-FLA-003: restore records the switch back, and a later real switch
// appends a second counted event.
func TestRestoreEnablesSecondCountedSwitch(t *testing.T) {
	store, clock := newTestStore(t)
	clock.Current = base.Add(time.Minute)
	if _, err := store.DeclareFallback("lane-1", TriggerChannelUnavailable, "t12"); err != nil {
		t.Fatalf("declare 1: %v", err)
	}
	clock.Current = base.Add(2 * time.Minute)
	if _, err := store.Restore("lane-1"); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	clock.Current = base.Add(3 * time.Minute)
	if _, err := store.DeclareFallback("lane-1", TriggerNoResponse, "t13"); err != nil {
		t.Fatalf("declare 2: %v", err)
	}
	events, err := store.Transitions("lane-1")
	if err != nil {
		t.Fatalf("Transitions: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("transition count = %d, want 3 (two switches + one restore)", len(events))
	}
	if events[1].Trigger != TriggerChannelRestored {
		t.Fatalf("middle event trigger = %q, want %q", events[1].Trigger, TriggerChannelRestored)
	}
}

// REQ-FLA-003: an unknown trigger kind is refused before anything is written.
func TestDeclareRefusesUnknownTrigger(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.DeclareFallback("lane-1", Trigger("banana"), ""); !errors.Is(err, ErrUnknownTrigger) {
		t.Fatalf("error = %v, want ErrUnknownTrigger", err)
	}
	events, err := store.Transitions("lane-1")
	if err != nil {
		t.Fatalf("Transitions: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("transition count = %d, want 0 (invalid trigger wrote nothing)", len(events))
	}
}
