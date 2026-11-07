package template

import (
	"strings"
	"testing"
)

// TestT2013_LeaderCheckPassOrderIsStated pins the fixed order of one leader
// check pass under the leader's remaining role section.
func TestT2013_LeaderCheckPassOrderIsStated(t *testing.T) {
	markers := []string{
		"One leader check pass runs in this fixed order",
		"lane liveness",
		"the blocking cause",
		"whether a predecessor condition was released",
		"independent work that can be assigned meanwhile",
		"the same owner repairing its own failure",
		"ACK or action confirmation",
	}
	for copyName, root := range t2013HandoffRoots() {
		t.Run(copyName, func(t *testing.T) {
			section := t2013HandoffSection(t, root, "factory-dispatch-mechanics.md", "## The leader's remaining role")
			for _, m := range markers {
				if !strings.Contains(section, m) {
					t.Errorf("leader remaining role section lacks %q", m)
				}
			}
		})
	}
}

// TestT2013_WaitRecordTracksFourFields pins the wait command form and the
// sentence naming the four tracked fields.
func TestT2013_WaitRecordTracksFourFields(t *testing.T) {
	markers := []string{
		`[--next "<next check event>"]`,
		"reason, owner (the waker), release condition and next check event",
		"is read with `unknown` and is never rejected or blocked",
		"an identical re-record performs no write",
	}
	for copyName, root := range t2013HandoffRoots() {
		t.Run(copyName, func(t *testing.T) {
			section := t2013HandoffSection(t, root, "factory-dispatch.md", "## The dispatch cycle")
			for _, m := range markers {
				if !strings.Contains(section, m) {
					t.Errorf("dispatch cycle section lacks %q", m)
				}
			}
		})
	}
}

// TestT2013_ResumeStartsFromCheckpoint pins the sentence telling a resuming lane
// or agent to start from the checkpoint verb and read original evidence only for
// the fields it needs.
func TestT2013_ResumeStartsFromCheckpoint(t *testing.T) {
	markers := []string{
		"A resuming lane or agent starts from `moai factory checkpoint <card>`",
		"(state plus evidence paths)",
		"reads the original evidence only for the fields it needs",
	}
	for copyName, root := range t2013HandoffRoots() {
		t.Run(copyName, func(t *testing.T) {
			section := t2013HandoffSection(t, root, "factory-dispatch.md", "## The dispatch cycle")
			for _, m := range markers {
				if !strings.Contains(section, m) {
					t.Errorf("dispatch cycle section lacks %q", m)
				}
			}
		})
	}
}
