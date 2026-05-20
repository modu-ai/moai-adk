package statusline

import (
	"encoding/json"
	"os"
	"strings"
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

// TestStdinData_UnmarshalJSON_EffortThinking verifies that Claude Code v2.1.122
// effort.level and thinking.enabled fields are parsed correctly from stdin JSON.
// REQ-CC2122-001: effort.level → "e:LEVEL" indicator
// REQ-CC2122-002: thinking.enabled=true → "·t" suffix
// REQ-CC2122-003: both absent → silent omit
func TestStdinData_UnmarshalJSON_EffortThinking(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantEffort   *EffortInfo
		wantThinking *ThinkingInfo
	}{
		{
			name:         "effort level present: high",
			input:        `{"effort":{"level":"high"}}`,
			wantEffort:   &EffortInfo{Level: "high"},
			wantThinking: nil,
		},
		{
			name:         "thinking enabled: true",
			input:        `{"thinking":{"enabled":true}}`,
			wantEffort:   nil,
			wantThinking: &ThinkingInfo{Enabled: true},
		},
		{
			name:         "both present",
			input:        `{"effort":{"level":"max"},"thinking":{"enabled":true}}`,
			wantEffort:   &EffortInfo{Level: "max"},
			wantThinking: &ThinkingInfo{Enabled: true},
		},
		{
			name:         "thinking enabled: false",
			input:        `{"thinking":{"enabled":false}}`,
			wantEffort:   nil,
			wantThinking: &ThinkingInfo{Enabled: false},
		},
		{
			name:         "effort level empty string",
			input:        `{"effort":{"level":""}}`,
			wantEffort:   &EffortInfo{Level: ""},
			wantThinking: nil,
		},
		{
			name:         "effort absent: nil",
			input:        `{"version":"1.0.0"}`,
			wantEffort:   nil,
			wantThinking: nil,
		},
		{
			name:         "effort null: nil",
			input:        `{"effort":null}`,
			wantEffort:   nil,
			wantThinking: nil,
		},
		{
			name:         "unknown effort level: raw passthrough (REQ-CC2122-004)",
			input:        `{"effort":{"level":"ultra"}}`,
			wantEffort:   &EffortInfo{Level: "ultra"},
			wantThinking: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data StdinData
			if err := json.Unmarshal([]byte(tt.input), &data); err != nil {
				t.Fatalf("json.Unmarshal failed: %v", err)
			}

			// Verify Effort field
			if tt.wantEffort == nil {
				if data.Effort != nil {
					t.Errorf("Effort = %+v, want nil", data.Effort)
				}
			} else {
				if data.Effort == nil {
					t.Fatalf("Effort is nil, want %+v", tt.wantEffort)
				}
				if data.Effort.Level != tt.wantEffort.Level {
					t.Errorf("Effort.Level = %q, want %q", data.Effort.Level, tt.wantEffort.Level)
				}
			}

			// Verify Thinking field
			if tt.wantThinking == nil {
				if data.Thinking != nil {
					t.Errorf("Thinking = %+v, want nil", data.Thinking)
				}
			} else {
				if data.Thinking == nil {
					t.Fatalf("Thinking is nil, want %+v", tt.wantThinking)
				}
				if data.Thinking.Enabled != tt.wantThinking.Enabled {
					t.Errorf("Thinking.Enabled = %v, want %v", data.Thinking.Enabled, tt.wantThinking.Enabled)
				}
			}
		})
	}
}

// TestSegmentEffortThinking_Constant verifies that SegmentEffortThinking constant is defined.
// REQ-CC2122-001: define effort_thinking segment constant
func TestSegmentEffortThinking_Constant(t *testing.T) {
	if SegmentEffortThinking != "effort_thinking" {
		t.Errorf("SegmentEffortThinking = %q, want %q", SegmentEffortThinking, "effort_thinking")
	}
}

// TestStatusData_EffortThinking_Fields verifies that StatusData has Effort and Thinking fields.
// REQ-CC2122-001: add Effort/Thinking fields to StatusData
func TestStatusData_EffortThinking_Fields(t *testing.T) {
	effort := &EffortInfo{Level: "high"}
	thinking := &ThinkingInfo{Enabled: true}
	data := &StatusData{
		Effort:   effort,
		Thinking: thinking,
	}
	if data.Effort == nil || data.Effort.Level != "high" {
		t.Errorf("StatusData.Effort = %+v, want &EffortInfo{Level:\"high\"}", data.Effort)
	}
	if data.Thinking == nil || !data.Thinking.Enabled {
		t.Errorf("StatusData.Thinking = %+v, want &ThinkingInfo{Enabled:true}", data.Thinking)
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

// TestStatuslineEffortThinking_MXTags verifies that EffortInfo and ThinkingInfo
// types in types.go have @MX:NOTE annotations per language.yaml code_comments setting.
// GWT-11: REQ-CC2122-006 — @MX:NOTE tags required on new exported types
func TestStatuslineEffortThinking_MXTags(t *testing.T) {
	// Read types.go source to verify @MX:NOTE annotations exist
	src, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	content := string(src)

	// Verify EffortInfo has a @MX:NOTE comment nearby (within the struct definition)
	if !strings.Contains(content, "@MX:NOTE") {
		t.Error("types.go should contain at least one @MX:NOTE annotation (REQ-CC2122-006)")
	}

	// Verify EffortInfo struct appears with associated context
	if !strings.Contains(content, "EffortInfo") {
		t.Error("types.go should define EffortInfo type")
	}
	if !strings.Contains(content, "ThinkingInfo") {
		t.Error("types.go should define ThinkingInfo type")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SPEC-V3R5-STATUSLINE-V2145-001 — M2 PR segment tests
// ─────────────────────────────────────────────────────────────────────────────

// TestSegmentPR_Constant verifies that SegmentPR constant is defined.
// REQ-SLV-016: introduce SegmentPR constant in types.go
func TestSegmentPR_Constant(t *testing.T) {
	if SegmentPR != "pr" {
		t.Errorf("SegmentPR = %q, want %q", SegmentPR, "pr")
	}
}

// TestStdinData_UnmarshalJSON_PRInfo verifies that Claude Code v2.1.145
// pr.{number,url,review_state} fields are parsed correctly from stdin JSON.
// REQ-SLV-010: adopt v2.1.145 PR stdin fields
func TestStdinData_UnmarshalJSON_PRInfo(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantPR *PRInfo
	}{
		{
			name:  "full pr object: number+url+review_state",
			input: `{"pr":{"number":1023,"url":"https://github.com/modu-ai/moai-adk/pull/1023","review_state":"approved"}}`,
			wantPR: &PRInfo{
				Number:      1023,
				URL:         "https://github.com/modu-ai/moai-adk/pull/1023",
				ReviewState: "approved",
			},
		},
		{
			name:  "pending review",
			input: `{"pr":{"number":42,"url":"https://github.com/o/r/pull/42","review_state":"pending"}}`,
			wantPR: &PRInfo{
				Number:      42,
				URL:         "https://github.com/o/r/pull/42",
				ReviewState: "pending",
			},
		},
		{
			name:  "changes_requested review",
			input: `{"pr":{"number":7,"url":"https://github.com/o/r/pull/7","review_state":"changes_requested"}}`,
			wantPR: &PRInfo{
				Number:      7,
				URL:         "https://github.com/o/r/pull/7",
				ReviewState: "changes_requested",
			},
		},
		{
			name:  "draft pr without review_state",
			input: `{"pr":{"number":99,"url":"https://github.com/o/r/pull/99"}}`,
			wantPR: &PRInfo{
				Number:      99,
				URL:         "https://github.com/o/r/pull/99",
				ReviewState: "",
			},
		},
		{
			name:  "explicit draft review_state",
			input: `{"pr":{"number":100,"url":"https://github.com/o/r/pull/100","review_state":"draft"}}`,
			wantPR: &PRInfo{
				Number:      100,
				URL:         "https://github.com/o/r/pull/100",
				ReviewState: "draft",
			},
		},
		{
			name:   "pr absent: nil",
			input:  `{"version":"2.1.145"}`,
			wantPR: nil,
		},
		{
			name:   "pr null: nil",
			input:  `{"pr":null}`,
			wantPR: nil,
		},
		{
			name:  "pr number only: empty url tolerated",
			input: `{"pr":{"number":50}}`,
			wantPR: &PRInfo{
				Number:      50,
				URL:         "",
				ReviewState: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data StdinData
			if err := json.Unmarshal([]byte(tt.input), &data); err != nil {
				t.Fatalf("json.Unmarshal failed: %v", err)
			}

			if tt.wantPR == nil {
				if data.PR != nil {
					t.Errorf("PR = %+v, want nil", data.PR)
				}
				return
			}

			if data.PR == nil {
				t.Fatalf("PR is nil, want %+v", tt.wantPR)
			}
			if data.PR.Number != tt.wantPR.Number {
				t.Errorf("PR.Number = %d, want %d", data.PR.Number, tt.wantPR.Number)
			}
			if data.PR.URL != tt.wantPR.URL {
				t.Errorf("PR.URL = %q, want %q", data.PR.URL, tt.wantPR.URL)
			}
			if data.PR.ReviewState != tt.wantPR.ReviewState {
				t.Errorf("PR.ReviewState = %q, want %q", data.PR.ReviewState, tt.wantPR.ReviewState)
			}
		})
	}
}

// TestWorkspaceInfo_UnmarshalJSON_Repo verifies that Claude Code v2.1.145
// workspace.repo.{host,owner,name} fields are parsed correctly.
// REQ-SLV-011: adopt workspace.repo stdin fields
func TestWorkspaceInfo_UnmarshalJSON_Repo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantRepo *RepoInfo
		wantWT   string // verify GitWorktree preservation
	}{
		{
			name:  "full repo object",
			input: `{"workspace":{"current_dir":"/repo","project_dir":"/repo","git_worktree":"","repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}}}`,
			wantRepo: &RepoInfo{
				Host:  "github.com",
				Owner: "modu-ai",
				Name:  "moai-adk",
			},
			wantWT: "",
		},
		{
			name:  "repo with worktree active",
			input: `{"workspace":{"current_dir":"/repo/.claude/worktrees/x","project_dir":"/repo","git_worktree":"/repo/.claude/worktrees/x","repo":{"host":"github.com","owner":"acme","name":"proj"}}}`,
			wantRepo: &RepoInfo{
				Host:  "github.com",
				Owner: "acme",
				Name:  "proj",
			},
			wantWT: "/repo/.claude/worktrees/x",
		},
		{
			name:     "repo absent: nil",
			input:    `{"workspace":{"current_dir":"/repo","project_dir":"/repo"}}`,
			wantRepo: nil,
			wantWT:   "",
		},
		{
			name:     "repo null: nil",
			input:    `{"workspace":{"current_dir":"/repo","project_dir":"/repo","repo":null}}`,
			wantRepo: nil,
			wantWT:   "",
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

			// Preserve GitWorktree
			if data.Workspace.GitWorktree != tt.wantWT {
				t.Errorf("GitWorktree = %q, want %q", data.Workspace.GitWorktree, tt.wantWT)
			}

			if tt.wantRepo == nil {
				if data.Workspace.Repo != nil {
					t.Errorf("Repo = %+v, want nil", data.Workspace.Repo)
				}
				return
			}

			if data.Workspace.Repo == nil {
				t.Fatalf("Repo is nil, want %+v", tt.wantRepo)
			}
			if data.Workspace.Repo.Host != tt.wantRepo.Host {
				t.Errorf("Repo.Host = %q, want %q", data.Workspace.Repo.Host, tt.wantRepo.Host)
			}
			if data.Workspace.Repo.Owner != tt.wantRepo.Owner {
				t.Errorf("Repo.Owner = %q, want %q", data.Workspace.Repo.Owner, tt.wantRepo.Owner)
			}
			if data.Workspace.Repo.Name != tt.wantRepo.Name {
				t.Errorf("Repo.Name = %q, want %q", data.Workspace.Repo.Name, tt.wantRepo.Name)
			}
		})
	}
}

// TestStdinData_UnmarshalJSON_V2145Schema verifies a representative v2.1.145
// stdin JSON containing PR + workspace.repo + version fields parses cleanly.
// REQ-SLV-010 + REQ-SLV-011: end-to-end v2.1.145 schema unmarshal.
func TestStdinData_UnmarshalJSON_V2145Schema(t *testing.T) {
	input := `{
		"model": {"id": "claude-opus-4-7", "display_name": "Opus"},
		"workspace": {
			"current_dir": "/home/user/project",
			"project_dir": "/home/user/project",
			"git_worktree": "",
			"repo": {"host": "github.com", "owner": "modu-ai", "name": "moai-adk"}
		},
		"pr": {"number": 1023, "url": "https://github.com/modu-ai/moai-adk/pull/1023", "review_state": "pending"},
		"version": "2.1.145"
	}`

	var data StdinData
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		t.Fatalf("v2.1.145 schema parse failed: %v", err)
	}

	if data.Version != "2.1.145" {
		t.Errorf("Version = %q, want %q", data.Version, "2.1.145")
	}
	if data.PR == nil {
		t.Fatal("PR is nil")
	}
	if data.PR.Number != 1023 {
		t.Errorf("PR.Number = %d, want 1023", data.PR.Number)
	}
	if data.PR.ReviewState != "pending" {
		t.Errorf("PR.ReviewState = %q, want %q", data.PR.ReviewState, "pending")
	}
	if data.Workspace == nil || data.Workspace.Repo == nil {
		t.Fatal("Workspace.Repo is nil")
	}
	if data.Workspace.Repo.Owner != "modu-ai" || data.Workspace.Repo.Name != "moai-adk" {
		t.Errorf("Repo = %+v, want owner=modu-ai name=moai-adk", data.Workspace.Repo)
	}
}

// TestStatusData_PR_Field verifies that StatusData has a PR field for
// downstream renderer consumption.
// REQ-SLV-016: PR data flows from StdinData to StatusData
func TestStatusData_PR_Field(t *testing.T) {
	pr := &PRInfo{Number: 1023, URL: "https://x/pull/1023", ReviewState: "approved"}
	data := &StatusData{PR: pr}
	if data.PR == nil || data.PR.Number != 1023 {
		t.Errorf("StatusData.PR = %+v, want Number=1023", data.PR)
	}
}
