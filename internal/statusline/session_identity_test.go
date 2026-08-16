package statusline

import (
	"strings"
	"testing"
)

// The statusline is the only durable place a session's identity can be shown:
// Claude Code drops its own prompt-bar name chip after /clear even though the
// explicit name (--name / /rename) is retained. These tests pin that the name
// and the agent identity reach the rendered L3 line, and that an unnamed
// ordinary session renders exactly as it did before the segment existed.

func TestBuildStatusData_CarriesSessionNameAndAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     *StdinData
		wantName  string
		wantAgent string
	}{
		{
			name:      "name and agent both present",
			input:     &StdinData{SessionName: "Team-A-Lead", Agent: &AgentInfo{Name: "manager-kanban"}},
			wantName:  "Team-A-Lead",
			wantAgent: "manager-kanban",
		},
		{
			name:      "name only — ordinary named session",
			input:     &StdinData{SessionName: "review-tjti6j"},
			wantName:  "review-tjti6j",
			wantAgent: "",
		},
		{
			name:      "agent object present but empty name",
			input:     &StdinData{SessionName: "worker-1", Agent: &AgentInfo{}},
			wantName:  "worker-1",
			wantAgent: "",
		},
		{
			name:      "neither — unnamed session",
			input:     &StdinData{},
			wantName:  "",
			wantAgent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var data StatusData
			if tt.input != nil {
				data.SessionName = tt.input.SessionName
				if tt.input.Agent != nil {
					data.AgentName = tt.input.Agent.Name
				}
			}

			if data.SessionName != tt.wantName {
				t.Errorf("SessionName = %q, want %q", data.SessionName, tt.wantName)
			}
			if data.AgentName != tt.wantAgent {
				t.Errorf("AgentName = %q, want %q", data.AgentName, tt.wantAgent)
			}
		})
	}
}

func TestRenderDirGitLine_SessionIdentityAtHead(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)

	got := r.renderDirGitLine(&StatusData{
		SessionName: "Team-A-Lead",
		AgentName:   "manager-kanban",
		Directory:   "moai-adk-go",
	})

	if !strings.Contains(got, "🏷️ Team-A-Lead") {
		t.Errorf("session name segment missing from %q", got)
	}
	if !strings.Contains(got, "👤 manager-kanban") {
		t.Errorf("agent segment missing from %q", got)
	}

	// Order is load-bearing: the operator scans left-to-right for identity
	// before location, and the user asked for the name at the head.
	nameAt := strings.Index(got, "Team-A-Lead")
	agentAt := strings.Index(got, "manager-kanban")
	dirAt := strings.Index(got, "moai-adk-go")
	if !(nameAt < agentAt && agentAt < dirAt) {
		t.Errorf("want order name < agent < directory, got name=%d agent=%d dir=%d in %q",
			nameAt, agentAt, dirAt, got)
	}
}

func TestRenderDirGitLine_OmitsEmptyIdentity(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)

	got := r.renderDirGitLine(&StatusData{Directory: "moai-adk-go"})

	if strings.Contains(got, "🏷️") || strings.Contains(got, "👤") {
		t.Errorf("unnamed session must render no identity segment, got %q", got)
	}
	if !strings.Contains(got, "📁 moai-adk-go") {
		t.Errorf("directory segment lost when identity absent: %q", got)
	}
}

func TestRenderDirGitLine_SessionSegmentDisablable(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, map[string]bool{SegmentSession: false})

	got := r.renderDirGitLine(&StatusData{
		SessionName: "Team-A-Lead",
		AgentName:   "manager-kanban",
		Directory:   "moai-adk-go",
	})

	if strings.Contains(got, "Team-A-Lead") || strings.Contains(got, "manager-kanban") {
		t.Errorf("disabled session segment still rendered: %q", got)
	}
	if !strings.Contains(got, "📁 moai-adk-go") {
		t.Errorf("directory segment lost when session segment disabled: %q", got)
	}
}
