package config

import (
	"testing"
)

func TestSourceString(t *testing.T) {
	tests := []struct {
		source   Source
		expected string
	}{
		{SrcPolicy, "policy"},
		{SrcUser, "user"},
		{SrcProject, "project"},
		{SrcLocal, "local"},
		{SrcPlugin, "plugin"},
		{SrcSkill, "skill"},
		{SrcSession, "session"},
		{SrcBuiltin, "builtin"},
		{Source(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.source.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSourcePriority(t *testing.T) {
	tests := []struct {
		source   Source
		expected int
	}{
		{SrcPolicy, 0},
		{SrcUser, 1},
		{SrcProject, 2},
		{SrcLocal, 3},
		{SrcPlugin, 4},
		{SrcSkill, 5},
		{SrcSession, 6},
		{SrcBuiltin, 7},
	}

	for _, tt := range tests {
		t.Run(tt.source.String(), func(t *testing.T) {
			if got := tt.source.Priority(); got != tt.expected {
				t.Errorf("Priority() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseSource(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  Source
		wantError bool
	}{
		{"policy", "policy", SrcPolicy, false},
		{"user", "user", SrcUser, false},
		{"project", "project", SrcProject, false},
		{"local", "local", SrcLocal, false},
		{"plugin", "plugin", SrcPlugin, false},
		{"skill", "skill", SrcSkill, false},
		{"session", "session", SrcSession, false},
		{"builtin", "builtin", SrcBuiltin, false},
		{"invalid", "invalid", SrcBuiltin, true},
		{"empty", "", SrcBuiltin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSource(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseSource() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got != tt.expected {
				t.Errorf("ParseSource() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAllSources(t *testing.T) {
	sources := AllSources()

	if len(sources) != 8 {
		t.Errorf("AllSources() returned %d sources, want 8", len(sources))
	}

	// Verify priority order
	for i := 0; i < len(sources); i++ {
		if sources[i].Priority() != i {
			t.Errorf("AllSources()[%d].Priority() = %d, want %d", i, sources[i].Priority(), i)
		}
	}
}

func TestSourceOrdering(t *testing.T) {
	// This test verifies that the 8-tier ordering matches the SPEC requirement
	// REQ-V3R2-RT-005-001: "SrcPolicy, SrcUser, SrcProject, SrcLocal, SrcPlugin, SrcSkill, SrcSession, SrcBuiltin"
	sources := AllSources()

	expectedOrder := []Source{
		SrcPolicy, SrcUser, SrcProject, SrcLocal,
		SrcPlugin, SrcSkill, SrcSession, SrcBuiltin,
	}

	for i, expected := range expectedOrder {
		if sources[i] != expected {
			t.Errorf("AllSources()[%d] = %v, want %v", i, sources[i], expected)
		}
	}
}
