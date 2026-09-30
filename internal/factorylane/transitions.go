package factorylane

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Trigger names why a mode switch happened. The two fallback kinds are the
// REQ-FLA-003 enumeration; channel-restored records the switch back so a
// second real fallback switch stays separately countable.
type Trigger string

const (
	// TriggerChannelUnavailable is the REQ-FLA-001 probe verdict fallback.
	TriggerChannelUnavailable Trigger = "channel-unavailable"
	// TriggerNoResponse is the REQ-FLA-002 expired directed-request fallback.
	TriggerNoResponse Trigger = "no-response"
	// TriggerChannelRestored marks the return to messaging (switch-back).
	TriggerChannelRestored Trigger = "channel-restored"
)

var (
	// ErrFallbackActive refuses a second declare while the lane is already in
	// fallback — one mode switch, one event (REQ-FLA-003).
	ErrFallbackActive = errors.New("factorylane: fallback already active for lane")
	// ErrUnknownTrigger refuses a trigger kind outside the enumeration.
	ErrUnknownTrigger = errors.New("factorylane: unknown fallback trigger")
)

// validFallbackTriggers is the REQ-FLA-003 activation enumeration. Restore
// markers are deliberately outside it: they end a fallback episode rather
// than begin one.
var validFallbackTriggers = map[Trigger]bool{
	TriggerChannelUnavailable: true,
	TriggerNoResponse:         true,
}

// Transition is one recorded mode switch: lane id, trigger kind, the
// in-progress card id (empty before pickup), and the switch timestamp
// (REQ-FLA-003). Persisted one file per event under transitions/<lane>/,
// append-only — a restore never rewrites or removes an activation.
type Transition struct {
	Lane    string    `json:"lane"`
	Trigger Trigger   `json:"trigger"`
	Card    string    `json:"card"`
	At      time.Time `json:"at"`
}

// transitionDir is the directory holding one lane's transition events.
func (s *Store) transitionDir(lane string) string {
	return filepath.Join(s.root, "transitions", lane)
}

// writeTransition appends one event file. The filename stamp is bumped on a
// collision (coarse clocks can mint two events inside one tick); the event's
// At field always carries the true switch instant.
func (s *Store) writeTransition(ev Transition) error {
	data, err := marshalRecord(ev)
	if err != nil {
		return err
	}
	dir := s.transitionDir(ev.Lane)
	stamp := ev.At
	for attempt := 0; attempt < 8; attempt++ {
		path := filepath.Join(dir, fmt.Sprintf("evt-%019d.json", stamp.UnixNano()))
		if _, err := os.Stat(path); err == nil {
			stamp = stamp.Add(time.Nanosecond)
			continue
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("factorylane: stat transition %s: %w", path, err)
		}
		return writeFileAtomic(path, data)
	}
	return fmt.Errorf("factorylane: transition filename collisions exhausted for lane %s", ev.Lane)
}

// DeclareFallback records one fallback activation — exactly one event per
// mode switch (REQ-FLA-003). While the lane's latest event is an unrestored
// fallback activation the lane is already in fallback, so a second declare is
// refused with ErrFallbackActive and writes nothing.
//
// @MX:ANCHOR: [AUTO] fallback mode-switch recorder — the exactly-once transition guarantee funnels through it
// @MX:REASON: fan_in >= 3 (factory fallback declare verb, no-response sweep wiring, M5 observability polish); a double-record here corrupts the count-by-lane audit (AC-FLA-003).
// @MX:SPEC: SPEC-FACTORY-LANE-AUTONOMY-001
func (s *Store) DeclareFallback(lane string, trigger Trigger, card string) (Transition, error) {
	if !validFallbackTriggers[trigger] {
		return Transition{}, fmt.Errorf("%w: %q", ErrUnknownTrigger, trigger)
	}
	events, err := s.Transitions(lane)
	if err != nil {
		return Transition{}, err
	}
	if len(events) > 0 {
		if last := events[len(events)-1]; validFallbackTriggers[last.Trigger] {
			return Transition{}, fmt.Errorf("%w: %s since %s", ErrFallbackActive, last.Trigger, last.At.Format(time.RFC3339))
		}
	}
	ev := Transition{Lane: lane, Trigger: trigger, Card: card, At: s.clock.Now().UTC()}
	if err := s.writeTransition(ev); err != nil {
		return Transition{}, err
	}
	return ev, nil
}

// Restore records the switch back to messaging. It always appends — restores
// are their own record, never an edit of the activation they close.
func (s *Store) Restore(lane string) (Transition, error) {
	ev := Transition{Lane: lane, Trigger: TriggerChannelRestored, At: s.clock.Now().UTC()}
	if err := s.writeTransition(ev); err != nil {
		return Transition{}, err
	}
	return ev, nil
}

// Transitions returns the lane's events in recorded order.
func (s *Store) Transitions(lane string) ([]Transition, error) {
	dir := s.transitionDir(lane)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("factorylane: read transitions dir %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "evt-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	events := make([]Transition, 0, len(names))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("factorylane: read transition %s: %w", name, err)
		}
		var ev Transition
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, fmt.Errorf("factorylane: parse transition %s: %w", name, err)
		}
		events = append(events, ev)
	}
	return events, nil
}
