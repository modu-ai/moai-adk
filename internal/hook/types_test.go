package hook

import (
	"testing"
)

func TestValidEventTypes(t *testing.T) {
	t.Parallel()

	events := ValidEventTypes()
	if len(events) != 6 {
		t.Errorf("ValidEventTypes() returned %d events, want 6", len(events))
	}

	expected := map[EventType]bool{
		EventSessionStart: true,
		EventPreToolUse:   true,
		EventPostToolUse:  true,
		EventSessionEnd:   true,
		EventStop:         true,
		EventPreCompact:   true,
	}

	for _, et := range events {
		if !expected[et] {
			t.Errorf("unexpected event type: %q", et)
		}
	}
}

func TestIsValidEventType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event EventType
		want  bool
	}{
		{"SessionStart is valid", EventSessionStart, true},
		{"PreToolUse is valid", EventPreToolUse, true},
		{"PostToolUse is valid", EventPostToolUse, true},
		{"SessionEnd is valid", EventSessionEnd, true},
		{"Stop is valid", EventStop, true},
		{"PreCompact is valid", EventPreCompact, true},
		{"empty string is invalid", EventType(""), false},
		{"unknown event is invalid", EventType("UnknownEvent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidEventType(tt.event); got != tt.want {
				t.Errorf("IsValidEventType(%q) = %v, want %v", tt.event, got, tt.want)
			}
		})
	}
}

func TestNewAllowOutput(t *testing.T) {
	t.Parallel()

	out := NewAllowOutput()
	if out.Decision != DecisionAllow {
		t.Errorf("Decision = %q, want %q", out.Decision, DecisionAllow)
	}
	if out.Reason != "" {
		t.Errorf("Reason = %q, want empty", out.Reason)
	}
}

func TestNewBlockOutput(t *testing.T) {
	t.Parallel()

	out := NewBlockOutput("test reason")
	if out.Decision != DecisionBlock {
		t.Errorf("Decision = %q, want %q", out.Decision, DecisionBlock)
	}
	if out.Reason != "test reason" {
		t.Errorf("Reason = %q, want %q", out.Reason, "test reason")
	}
}

func TestNewProtocol(t *testing.T) {
	t.Parallel()

	proto := NewProtocol()
	if proto == nil {
		t.Fatal("NewProtocol() returned nil")
	}
}
