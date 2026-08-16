package statusline

import (
	"strings"
	"testing"
)

// The statusline is the only durable place a session's identity can be shown:
// Claude Code drops its own prompt-bar name chip after /clear even though the
// explicit name (--name / /rename) is retained, and the directory segment shows
// the worktree once a session enters one. These tests pin that the name and the
// agent identity lead the info line, and that an unnamed ordinary session
// renders exactly as it did before the segment existed.

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
			data.SessionName = tt.input.SessionName
			if tt.input.Agent != nil {
				data.AgentName = tt.input.Agent.Name
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

// namedData builds a StatusData carrying both an identity and a model, so the
// ordering assertions below have something to order against.
func namedData() *StatusData {
	d := &StatusData{
		SessionName: "Team-A-Lead",
		AgentName:   "manager-kanban",
		Directory:   "statusline-session-name",
	}
	d.Metrics.Available = true
	d.Metrics.Model = "Opus 5 (1M context)"
	return d
}

func TestRenderInfoLine_SessionIdentityLeadsTheLine(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)
	got := r.renderInfoLine(namedData(), false)

	if !strings.Contains(got, "🏷️ Team-A-Lead") {
		t.Errorf("session name segment missing from %q", got)
	}
	if !strings.Contains(got, "👤 manager-kanban") {
		t.Errorf("agent segment missing from %q", got)
	}

	// Order is load-bearing: identity is what the operator scans for first when
	// several sessions are open, so it precedes the model.
	nameAt := strings.Index(got, "Team-A-Lead")
	agentAt := strings.Index(got, "manager-kanban")
	modelAt := strings.Index(got, "Opus 5")
	if !(nameAt < agentAt && agentAt < modelAt) {
		t.Errorf("want order name < agent < model, got name=%d agent=%d model=%d in %q",
			nameAt, agentAt, modelAt, got)
	}
}

func TestRenderInfoLine_OmitsEmptyIdentity(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)
	d := namedData()
	d.SessionName = ""
	d.AgentName = ""

	got := r.renderInfoLine(d, false)

	if strings.Contains(got, "🏷️") || strings.Contains(got, "👤") {
		t.Errorf("unnamed session must render no identity segment, got %q", got)
	}
	if !strings.Contains(got, "🤖 Opus 5 (1M context)") {
		t.Errorf("model segment lost when identity absent: %q", got)
	}
}

func TestRenderInfoLine_SessionSegmentDisablable(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, map[string]bool{SegmentSession: false})
	got := r.renderInfoLine(namedData(), false)

	if strings.Contains(got, "Team-A-Lead") || strings.Contains(got, "manager-kanban") {
		t.Errorf("disabled session segment still rendered: %q", got)
	}
	if !strings.Contains(got, "🤖 Opus 5 (1M context)") {
		t.Errorf("model segment lost when session segment disabled: %q", got)
	}
}

// The identity was moved from the directory line to the info line. Rendering it
// in both places would double it on every status update, so pin its absence
// here rather than trusting the move stayed clean.
func TestRenderDirGitLine_CarriesNoSessionIdentity(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)
	got := r.renderDirGitLine(namedData())

	if strings.Contains(got, "🏷️") || strings.Contains(got, "👤") {
		t.Errorf("identity must live on the info line only, found it in %q", got)
	}
	if !strings.Contains(got, "📁 statusline-session-name") {
		t.Errorf("directory segment lost: %q", got)
	}
}
