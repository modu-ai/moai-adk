package harness

import "testing"

// TestIsEligibleForPromotion covers REQ-HRR-003 / REQ-HRR-004 (D4: filter at
// classification time). A pattern is NOT eligible for promotion when it carries
// an empty context hash AND an empty-or-"unknown" subject — the degenerate
// lifecycle-noise class (session_stop::, user_prompt::, subagent_stop:unknown:).
// Patterns with a non-empty context hash OR a meaningful subject remain
// eligible.
func TestIsEligibleForPromotion(t *testing.T) {
	cases := []struct {
		name string
		p    *Pattern
		want bool
	}{
		{
			name: "degenerate subagent_stop unknown (the 2026-05-24 regression)",
			p:    &Pattern{EventType: EventTypeSubagentStop, Subject: "unknown", ContextHash: ""},
			want: false,
		},
		{
			name: "degenerate session_stop empty subject empty context",
			p:    &Pattern{EventType: EventTypeSessionStop, Subject: "", ContextHash: ""},
			want: false,
		},
		{
			name: "degenerate user_prompt empty subject empty context",
			p:    &Pattern{EventType: EventTypeUserPrompt, Subject: "", ContextHash: ""},
			want: false,
		},
		{
			name: "eligible agent_invocation with non-empty context hash (control)",
			p:    &Pattern{EventType: EventTypeAgentInvocation, Subject: "Bash", ContextHash: "abc123"},
			want: true,
		},
		{
			name: "eligible tool_failure with error-class signature",
			p:    &Pattern{EventType: EventTypeToolFailure, Subject: "Bash", ContextHash: "ExitError"},
			want: true,
		},
		{
			name: "eligible non-empty context hash even with empty subject (real signal present)",
			p:    &Pattern{EventType: EventTypeAgentInvocation, Subject: "", ContextHash: "ctx-hash-1"},
			want: true,
		},
		{
			name: "eligible meaningful subject with empty context (subagent with real name)",
			p:    &Pattern{EventType: EventTypeSubagentStop, Subject: "manager-develop", ContextHash: ""},
			want: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := IsEligibleForPromotion(c.p)
			if got != c.want {
				t.Errorf("IsEligibleForPromotion(%s:%s:%s) = %v, want %v",
					c.p.EventType, c.p.Subject, c.p.ContextHash, got, c.want)
			}
		})
	}
}
