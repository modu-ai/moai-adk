package statusline

import (
	"encoding/json"
	"strings"
	"testing"
)

// strPtr is a test helper for the pointer-nil optional fields.
func strPtr(s string) *string { return &s }

// subagentRow is the JSONL output shape the subagentStatusLine contract
// requires ({"id","content"} per stdout line).
type subagentRow struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// parseSubagentRows splits a renderSubagentOutput result into rows, failing
// the test on any line that does not parse as the contract shape.
func parseSubagentRows(t *testing.T, out string) []subagentRow {
	t.Helper()
	var rows []subagentRow
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		var row subagentRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("output line is not contract JSON %q: %v", line, err)
		}
		rows = append(rows, row)
	}
	return rows
}

// TestAgentType verifies the subagentStatusLine tasks[] agentType badge
// (SPEC-CC-HAIKU55-STATUSLINE-001 M1, REQ-CC-HAIKU55-010 / REQ-CC-HAIKU55-011,
// AC-008). Four subcases: agentType present → badge rendered; key absent →
// no badge, no error; JSON null → no badge, no error; empty/absent tasks[]
// → render succeeds. A fifth subcase covers the unnamed-two-tasks
// disambiguation (the name field is optional upstream).
func TestAgentType(t *testing.T) {
	t.Run("badge_rendered_when_agentType_present", func(t *testing.T) {
		out := renderSubagentOutput([]SubagentTaskInfo{
			{ID: "task-1", AgentType: strPtr("manager-develop"), Name: strPtr("implement parser"), Status: "running", TokenCount: 12300},
		})
		rows := parseSubagentRows(t, out)
		if len(rows) != 1 {
			t.Fatalf("expected 1 JSONL row, got %d (%q)", len(rows), out)
		}
		if rows[0].ID != "task-1" {
			t.Errorf("task id must be echoed verbatim, got %q", rows[0].ID)
		}
		if !strings.Contains(rows[0].Content, "[manager-develop]") {
			t.Errorf("row content should carry the agentType badge [manager-develop], got %q", rows[0].Content)
		}
		if !strings.Contains(rows[0].Content, "12.3k tokens") {
			t.Errorf("row content should carry the running token count, got %q", rows[0].Content)
		}
	})

	t.Run("no_badge_when_agentType_key_absent", func(t *testing.T) {
		// CC pre-2.1.293 payloads omit the agentType key entirely: the row
		// decodes with AgentType == nil, renders without the badge, and must
		// not error or drop the row (REQ-CC-HAIKU55-011).
		var input StdinData
		if err := json.Unmarshal([]byte(`{"tasks":[{"id":"task-2","type":"local_agent"}]}`), &input); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(input.SubagentTasks) != 1 {
			t.Fatalf("expected 1 decoded task, got %d", len(input.SubagentTasks))
		}
		if input.SubagentTasks[0].AgentType != nil {
			t.Fatalf("absent agentType key should decode to nil, got %v", *input.SubagentTasks[0].AgentType)
		}
		rows := parseSubagentRows(t, renderSubagentOutput(input.SubagentTasks))
		if len(rows) != 1 {
			t.Fatalf("row must stay present without the badge, got %d rows", len(rows))
		}
		if strings.Contains(rows[0].Content, "[]") {
			t.Errorf("empty badge brackets should never render, got %q", rows[0].Content)
		}
	})

	t.Run("no_badge_when_agentType_null", func(t *testing.T) {
		var input StdinData
		if err := json.Unmarshal([]byte(`{"tasks":[{"id":"task-3","agentType":null,"type":"local_agent"}]}`), &input); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if input.SubagentTasks[0].AgentType != nil {
			t.Fatalf("JSON null agentType should decode to nil, got %v", *input.SubagentTasks[0].AgentType)
		}
		rows := parseSubagentRows(t, renderSubagentOutput(input.SubagentTasks))
		if len(rows) != 1 {
			t.Fatalf("row must stay present without the badge, got %d rows", len(rows))
		}
		if strings.Contains(rows[0].Content, "[]") {
			t.Errorf("empty badge brackets should never render, got %q", rows[0].Content)
		}
	})

	t.Run("render_succeeds_with_empty_and_absent_tasks", func(t *testing.T) {
		// Absent key: an ordinary statusline payload — the bar path, not the
		// JSONL contract.
		var absent StdinData
		if err := json.Unmarshal([]byte(`{"session_id":"s"}`), &absent); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if absent.SubagentTasks != nil {
			t.Fatalf("absent tasks key should decode to nil slice, got %v", absent.SubagentTasks)
		}
		r := newTestRenderer()
		got := r.Render(&StatusData{}, ModeDefault)
		if strings.Contains(got, "⚙") {
			t.Errorf("absent tasks[] should render nothing new, got %q", got)
		}
		// Present-but-empty `[]`: the JSONL contract with zero rows → empty
		// output (no rows to override).
		if out := renderSubagentOutput([]SubagentTaskInfo{}); out != "" {
			t.Errorf("empty tasks[] should produce empty output, got %q", out)
		}
	})

	t.Run("unnamed_tasks_disambiguate_through_fallback_chain", func(t *testing.T) {
		// The name field is optional upstream; two unnamed tasks must never
		// render identically. Fallback: name → label → description → #id-prefix.
		out := renderSubagentOutput([]SubagentTaskInfo{
			{ID: "11111111-aaaa-bbbb-cccc-dddddddddddd", AgentType: strPtr("Explore"), Label: "scanning imports"},
			{ID: "22222222-aaaa-bbbb-cccc-dddddddddddd", AgentType: strPtr("Explore"), Description: "reading test files"},
		})
		rows := parseSubagentRows(t, out)
		if len(rows) != 2 {
			t.Fatalf("expected 2 JSONL rows, got %d", len(rows))
		}
		if rows[0].Content == rows[1].Content {
			t.Fatalf("two unnamed tasks must not render identically, both %q", rows[0].Content)
		}
		if !strings.Contains(rows[0].Content, "scanning imports") {
			t.Errorf("first row should fall back to label, got %q", rows[0].Content)
		}
		if !strings.Contains(rows[1].Content, "reading test files") {
			t.Errorf("second row should fall back to description, got %q", rows[1].Content)
		}
		// With no label, no description, and no name: the truncated id prefix.
		out2 := renderSubagentOutput([]SubagentTaskInfo{
			{ID: "abcdefgh-1234", AgentType: strPtr("Explore")},
			{ID: "task-x", AgentType: strPtr("Explore"), Name: strPtr("big runner"), TokenCount: 1_234_567},
			{ID: "task-y", AgentType: strPtr("Explore"), Name: strPtr("small runner"), TokenCount: 850},
		})
		rows2 := parseSubagentRows(t, out2)
		if !strings.Contains(rows2[0].Content, "#abcdefgh") {
			t.Errorf("fully unnamed row should fall back to the id prefix, got %q", rows2[0].Content)
		}
		if !strings.Contains(rows2[1].Content, "1.2M tokens") {
			t.Errorf("M-scale token count should format as 1.2M, got %q", rows2[1].Content)
		}
		if !strings.Contains(rows2[2].Content, "850 tokens") {
			t.Errorf("sub-k token count should render bare, got %q", rows2[2].Content)
		}
	})

	t.Run("row_with_empty_id_is_skipped_default_rendering_kept", func(t *testing.T) {
		// Contract: omitting a task's id keeps Claude Code's default rendering
		// for that row — so an empty id writes no JSON line at all.
		out := renderSubagentOutput([]SubagentTaskInfo{
			{ID: "", AgentType: strPtr("Explore"), Name: strPtr("no id")},
			{ID: "task-9", AgentType: strPtr("Explore"), Name: strPtr("real row")},
		})
		rows := parseSubagentRows(t, out)
		if len(rows) != 1 || rows[0].ID != "task-9" {
			t.Fatalf("empty-id row must be skipped, got %v", rows)
		}
	})
}
