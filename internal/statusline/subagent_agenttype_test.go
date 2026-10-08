package statusline

import (
	"encoding/json"
	"strings"
	"testing"
)

// strPtr is a test helper for the pointer-nil agentType field.
func strPtr(s string) *string { return &s }

// TestAgentType verifies the subagentStatusLine tasks[] agentType badge
// (SPEC-CC-HAIKU55-STATUSLINE-001 M1, REQ-CC-HAIKU55-010 / REQ-CC-HAIKU55-011,
// AC-008). Four subcases: agentType present → badge rendered; key absent →
// no badge, no error; JSON null → no badge, no error; empty/absent tasks[]
// → render succeeds.
func TestAgentType(t *testing.T) {
	r := newTestRenderer()

	t.Run("badge_rendered_when_agentType_present", func(t *testing.T) {
		data := &StatusData{
			SubagentTasks: []SubagentTaskInfo{
				{AgentType: strPtr("manager-develop"), Name: "implement parser"},
			},
		}
		got := r.Render(data, ModeDefault)
		if !strings.Contains(got, "[manager-develop]") {
			t.Errorf("rendered output should carry the agentType badge [manager-develop], got %q", got)
		}
		if !strings.Contains(got, "implement parser") {
			t.Errorf("task row should still name the task, got %q", got)
		}
	})

	t.Run("no_badge_when_agentType_key_absent", func(t *testing.T) {
		// CC pre-2.1.293 payloads omit the agentType key entirely: the row
		// decodes with AgentType == nil, renders without the badge, and must
		// not error or drop the row (REQ-CC-HAIKU55-011).
		var input StdinData
		if err := json.Unmarshal([]byte(`{"tasks":[{"name":"implement parser","type":"task"}]}`), &input); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(input.SubagentTasks) != 1 {
			t.Fatalf("expected 1 decoded task, got %d", len(input.SubagentTasks))
		}
		if input.SubagentTasks[0].AgentType != nil {
			t.Fatalf("absent agentType key should decode to nil, got %v", *input.SubagentTasks[0].AgentType)
		}
		got := r.Render(&StatusData{SubagentTasks: input.SubagentTasks}, ModeDefault)
		if !strings.Contains(got, "implement parser") {
			t.Errorf("row must stay present without the badge, got %q", got)
		}
		if strings.Contains(got, "[]") {
			t.Errorf("empty badge brackets should never render, got %q", got)
		}
	})

	t.Run("no_badge_when_agentType_null", func(t *testing.T) {
		var input StdinData
		if err := json.Unmarshal([]byte(`{"tasks":[{"agentType":null,"name":"implement parser","type":"task"}]}`), &input); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if input.SubagentTasks[0].AgentType != nil {
			t.Fatalf("JSON null agentType should decode to nil, got %v", *input.SubagentTasks[0].AgentType)
		}
		got := r.Render(&StatusData{SubagentTasks: input.SubagentTasks}, ModeDefault)
		if !strings.Contains(got, "implement parser") {
			t.Errorf("row must stay present without the badge, got %q", got)
		}
		if strings.Contains(got, "[]") {
			t.Errorf("empty badge brackets should never render, got %q", got)
		}
	})

	t.Run("render_succeeds_with_empty_and_absent_tasks", func(t *testing.T) {
		gotAbsent := r.Render(&StatusData{}, ModeDefault)
		if strings.Contains(gotAbsent, "[]") {
			t.Errorf("absent tasks[] should render nothing, got %q", gotAbsent)
		}
		gotEmpty := r.Render(&StatusData{SubagentTasks: []SubagentTaskInfo{}}, ModeDefault)
		if strings.Contains(gotEmpty, "[]") {
			t.Errorf("empty tasks[] should render nothing, got %q", gotEmpty)
		}
	})
}
