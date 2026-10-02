package config

// factory_marker_values_m1_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1 (card
// t1399), AC-016: the six marker values the factory reads are frozen.
//
// A launcher of today's binary publishes these names to a leader or a lane;
// a lane of a newer binary discovers a leader started by an older binary
// through MOAI_KANBAN_ID. A rename of a value would make the two disagree,
// quietly. The Go constants may be renamed (M7 re-points this test at the new
// constant names mechanically); the string values below are written here as
// literals and never move.

import "testing"

func TestFactoryMarkerValuesFrozen(t *testing.T) {
	for _, tc := range []struct {
		constant string
		got      string
		want     string
	}{
		{"EnvMoaiKanbanID", EnvMoaiKanbanID, "MOAI_KANBAN_ID"},
		{"EnvMoaiKanbanLeadAddr", EnvMoaiKanbanLeadAddr, "MOAI_KANBAN_LEAD_ADDR"},
		{"EnvMoaiKanbanLeadName", EnvMoaiKanbanLeadName, "MOAI_KANBAN_LEAD_NAME"},
		{"EnvMoaiKanbanSettingsInjected", EnvMoaiKanbanSettingsInjected, "MOAI_KANBAN_SETTINGS_INJECTED"},
		{"EnvMoaiKanbanBackend", EnvMoaiKanbanBackend, "MOAI_KANBAN_BACKEND"},
		{"EnvMoaiKanbanCard", EnvMoaiKanbanCard, "MOAI_KANBAN_CARD"},
	} {
		if tc.got == "" {
			t.Errorf("%s is empty — the comparison would pass vacuously", tc.constant)
			continue
		}
		if tc.got != tc.want {
			t.Errorf("%s = %q, want the frozen wire value %q", tc.constant, tc.got, tc.want)
		}
	}
}
