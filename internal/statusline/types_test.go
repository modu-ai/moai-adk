package statusline

import (
	"encoding/json"
	"testing"
)

// TestNormalizeMode verifies backward-compatible mode name normalization.
// REQ-V3-MODE-001: "minimal" → "default" conversion
// REQ-V3-MODE-002: "verbose" → "full" conversion
// REQ-V3-MODE-003: "compact" → "default" conversion
func TestNormalizeMode(t *testing.T) {
	tests := []struct {
		name  string
		input StatuslineMode
		want  StatuslineMode
	}{
		// Backward compatibility: old name → new name conversion
		{"minimal converts to default", "minimal", ModeDefault},
		{"compact converts to default", "compact", ModeDefault},
		{"verbose converts to full", "verbose", ModeFull},
		// Current names remain unchanged
		{"default unchanged", "default", ModeDefault},
		{"full unchanged", "full", ModeFull},
		// Edge cases
		{"empty unchanged", "", ""},
		{"unknown unchanged", "custom", "custom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeMode(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestStdinData_UnmarshalJSON_ResetsAtInteger verifies that the official Claude Code
// statusline JSON schema is parsed correctly. The schema sends resets_at as a Unix
// epoch integer, but our code previously declared it as string, causing a total
// parse failure.
//
// See Issue #549: https://github.com/modu-ai/moai-adk/issues/549
func TestStdinData_UnmarshalJSON_ResetsAtInteger(t *testing.T) {
	input := `{"rate_limits":{"five_hour":{"used_percentage":23.5,"resets_at":1738425600}}}`

	var data StdinData
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		t.Fatalf("json.Unmarshal failed (Bug 1: resets_at type mismatch): %v", err)
	}
	if data.RateLimits == nil {
		t.Fatal("data.RateLimits is nil, expected non-nil")
	}
	if data.RateLimits.FiveHour == nil {
		t.Fatal("data.RateLimits.FiveHour is nil, expected non-nil")
	}
	if got := data.RateLimits.FiveHour.UsedPercentage; got < 23.4 || got > 23.6 {
		t.Errorf("UsedPercentage = %v, want ~23.5", got)
	}
	if data.RateLimits.FiveHour.ResetsAt != 1738425600 {
		t.Errorf("ResetsAt = %v, want 1738425600", data.RateLimits.FiveHour.ResetsAt)
	}
}

// TestStdinData_UnmarshalJSON_SevenDayResetsAt verifies 7-day rate limit parsing.
func TestStdinData_UnmarshalJSON_SevenDayResetsAt(t *testing.T) {
	input := `{"rate_limits":{"seven_day":{"used_percentage":41.2,"resets_at":1738857600}}}`

	var data StdinData
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if data.RateLimits == nil || data.RateLimits.SevenDay == nil {
		t.Fatal("data.RateLimits.SevenDay is nil")
	}
	if data.RateLimits.SevenDay.ResetsAt != 1738857600 {
		t.Errorf("ResetsAt = %v, want 1738857600", data.RateLimits.SevenDay.ResetsAt)
	}
}

// TestStdinData_UnmarshalJSON_ModelAsString verifies that model field can be parsed
// as a plain string (not only as an object). Some Claude Code versions send
// "model": "claude-opus-4-6" instead of "model": {"id": "...", "display_name": "..."}.
//
// See Issue #549.
func TestStdinData_UnmarshalJSON_ModelAsString(t *testing.T) {
	input := `{"model":"claude-opus-4-6","context_window":{"used_percentage":25}}`

	var data StdinData
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		t.Fatalf("json.Unmarshal failed (Bug 2: model as string): %v", err)
	}
	if data.Model == nil {
		t.Fatal("data.Model is nil, expected non-nil")
	}
	if data.Model.ID != "claude-opus-4-6" {
		t.Errorf("Model.ID = %q, want %q", data.Model.ID, "claude-opus-4-6")
	}
}

// TestStdinData_UnmarshalJSON_ModelAsObject verifies that model field still works
// as an object (existing behavior must not break).
func TestStdinData_UnmarshalJSON_ModelAsObject(t *testing.T) {
	input := `{"model":{"id":"claude-opus-4-6","display_name":"Opus"}}`

	var data StdinData
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if data.Model == nil {
		t.Fatal("data.Model is nil")
	}
	if data.Model.ID != "claude-opus-4-6" {
		t.Errorf("Model.ID = %q, want %q", data.Model.ID, "claude-opus-4-6")
	}
	if data.Model.DisplayName != "Opus" {
		t.Errorf("Model.DisplayName = %q, want %q", data.Model.DisplayName, "Opus")
	}
}

// TestStdinData_UnmarshalJSON_OfficialSchema verifies that the exact JSON from
// the Claude Code official documentation parses correctly.
func TestStdinData_UnmarshalJSON_OfficialSchema(t *testing.T) {
	// Exact JSON from https://code.claude.com/docs/en/statusline
	input := `{
		"model": {"id": "claude-opus-4-6", "display_name": "Opus"},
		"context_window": {"used_percentage": 25, "context_window_size": 200000},
		"rate_limits": {
			"five_hour": {"used_percentage": 23.5, "resets_at": 1738425600},
			"seven_day": {"used_percentage": 41.2, "resets_at": 1738857600}
		},
		"workspace": {"current_dir": "/home/user/project", "project_dir": "/home/user/project"},
		"version": "1.0.80",
		"cost": {"total_cost_usd": 0.05}
	}`

	var data StdinData
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		t.Fatalf("official schema parse failed: %v", err)
	}

	// Verify all key fields parsed
	if data.Model == nil || data.Model.ID != "claude-opus-4-6" {
		t.Errorf("Model not parsed correctly: %+v", data.Model)
	}
	if data.ContextWindow == nil {
		t.Error("ContextWindow is nil")
	}
	if data.RateLimits == nil {
		t.Error("RateLimits is nil")
	} else {
		if data.RateLimits.FiveHour == nil {
			t.Error("RateLimits.FiveHour is nil")
		} else if data.RateLimits.FiveHour.ResetsAt != 1738425600 {
			t.Errorf("FiveHour.ResetsAt = %v, want 1738425600", data.RateLimits.FiveHour.ResetsAt)
		}
		if data.RateLimits.SevenDay == nil {
			t.Error("RateLimits.SevenDay is nil")
		} else if data.RateLimits.SevenDay.ResetsAt != 1738857600 {
			t.Errorf("SevenDay.ResetsAt = %v, want 1738857600", data.RateLimits.SevenDay.ResetsAt)
		}
	}
	if data.Version != "1.0.80" {
		t.Errorf("Version = %q, want %q", data.Version, "1.0.80")
	}
	if data.Cost == nil || data.Cost.TotalCostUSD != 0.05 {
		t.Errorf("Cost not parsed correctly: %+v", data.Cost)
	}
}

// TestWorkspaceInfo_UnmarshalJSON_GitWorktree verifies that git_worktree field
// from Claude Code 2.1.97+ is correctly parsed from workspace JSON.
// REQ-CC297-003: support workspace.git_worktree field
func TestWorkspaceInfo_UnmarshalJSON_GitWorktree(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantWT  string
		wantDir string
	}{
		{
			name:    "worktree path present: parsed correctly",
			input:   `{"workspace": {"current_dir": "/repo/.claude/worktrees/abc123", "project_dir": "/repo", "git_worktree": "/repo/.claude/worktrees/abc123"}}`,
			wantWT:  "/repo/.claude/worktrees/abc123",
			wantDir: "/repo",
		},
		{
			name:    "no worktree: empty string",
			input:   `{"workspace": {"current_dir": "/repo", "project_dir": "/repo"}}`,
			wantWT:  "",
			wantDir: "/repo",
		},
		{
			name:    "git_worktree empty string",
			input:   `{"workspace": {"current_dir": "/repo", "project_dir": "/repo", "git_worktree": ""}}`,
			wantWT:  "",
			wantDir: "/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data StdinData
			if err := json.Unmarshal([]byte(tt.input), &data); err != nil {
				t.Fatalf("json.Unmarshal failed: %v", err)
			}
			if data.Workspace == nil {
				t.Fatal("Workspace is nil")
			}
			if data.Workspace.GitWorktree != tt.wantWT {
				t.Errorf("GitWorktree = %q, want %q", data.Workspace.GitWorktree, tt.wantWT)
			}
			if data.Workspace.ProjectDir != tt.wantDir {
				t.Errorf("ProjectDir = %q, want %q", data.Workspace.ProjectDir, tt.wantDir)
			}
		})
	}
}

// TestSegmentWorktree_Constant verifies that SegmentWorktree constant is defined.
// REQ-CC297-003: define worktree segment constant
func TestSegmentWorktree_Constant(t *testing.T) {
	// Verify constant value is "worktree"
	if SegmentWorktree != "worktree" {
		t.Errorf("SegmentWorktree = %q, want %q", SegmentWorktree, "worktree")
	}
}

// TestStatusData_Worktree_Field verifies that StatusData has a Worktree field.
// REQ-CC297-003: add Worktree field to StatusData
func TestStatusData_Worktree_Field(t *testing.T) {
	data := &StatusData{
		Worktree: "/repo/.claude/worktrees/abc123",
	}
	if data.Worktree != "/repo/.claude/worktrees/abc123" {
		t.Errorf("Worktree = %q, want %q", data.Worktree, "/repo/.claude/worktrees/abc123")
	}
}
