package agenthost

import (
	"strings"
	"testing"
)

func TestMatrixFor_CodexHookParity(t *testing.T) {
	matrix, err := MatrixFor(HostCodex)
	if err != nil {
		t.Fatalf("MatrixFor(codex): %v", err)
	}

	for _, event := range []Event{
		EventSessionStart,
		EventPreToolUse,
		EventPermissionRequest,
		EventPostToolUse,
		EventUserPromptSubmit,
		EventStop,
		EventSubagentStart,
		EventSubagentStop,
		EventPreCompact,
		EventPostCompact,
	} {
		mapping, ok := matrix.Find(event)
		if !ok {
			t.Fatalf("codex matrix missing %s", event)
		}
		if mapping.Support != SupportNative {
			t.Errorf("codex %s support = %s, want native", event, mapping.Support)
		}
		if mapping.HostEvent == "" {
			t.Errorf("codex %s host event should not be empty", event)
		}
	}
}

func TestMatrixFor_OpenCodeAdapterBoundaries(t *testing.T) {
	matrix, err := MatrixFor(HostOpenCode)
	if err != nil {
		t.Fatalf("MatrixFor(opencode): %v", err)
	}

	cases := []struct {
		event    Event
		want     SupportLevel
		hostPart string
	}{
		{EventPreToolUse, SupportAdapter, "tool.execute.before"},
		{EventPermissionRequest, SupportAdapter, "permission.asked"},
		{EventPostToolUse, SupportAdapter, "tool.execute.after"},
		{EventUserPromptSubmit, SupportFallback, "tui.prompt.append"},
		{EventStop, SupportFallback, "session.idle"},
		{EventSubagentStop, SupportFallback, "session.idle"},
		{EventPreCompact, SupportAdapter, "experimental.session.compacting"},
	}

	for _, tc := range cases {
		mapping, ok := matrix.Find(tc.event)
		if !ok {
			t.Fatalf("opencode matrix missing %s", tc.event)
		}
		if mapping.Support != tc.want {
			t.Errorf("opencode %s support = %s, want %s", tc.event, mapping.Support, tc.want)
		}
		if !strings.Contains(mapping.HostEvent, tc.hostPart) {
			t.Errorf("opencode %s host event = %q, want to contain %q", tc.event, mapping.HostEvent, tc.hostPart)
		}
		if mapping.Support != SupportNative && mapping.Degradation == "" {
			t.Errorf("opencode %s should explain degradation", tc.event)
		}
	}
}

func TestParseHost(t *testing.T) {
	host, err := ParseHost(" Codex ")
	if err != nil {
		t.Fatalf("ParseHost: %v", err)
	}
	if host != HostCodex {
		t.Fatalf("host = %q, want %q", host, HostCodex)
	}

	if _, err := ParseHost("unknown"); err == nil {
		t.Fatal("ParseHost should reject unknown host")
	}
}
