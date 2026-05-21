// Package statusline tests for SPEC-V3R5-STATUSLINE-STDINFIELDS-001.
//
// Verifies the three new stdin-driven segments (repo, long_context,
// handoff_guide), the StdinData.ExceedsLong pass-through into StatusData,
// and the verification-only PR segment graceful no-output behavior.
//
// REQ-SSE-001..006 cover the new segments and field wiring.
// REQ-SSE-007 covers PR segment graceful no-output assertion.
package statusline

import (
	"strings"
	"testing"
)

// ---------- REQ-SSE-001/002: renderRepoSegment ----------

func TestRenderRepoSegment_PopulatedShowsOwnerName(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Workspace: WorkspaceData{
			Repo: &RepoInfo{
				Host:  "github.com",
				Owner: "modu-ai",
				Name:  "moai-adk",
			},
		},
	}

	out := renderRepoSegment(data)
	if !strings.Contains(out, "modu-ai/moai-adk") {
		t.Fatalf("expected output to contain owner/name marker %q, got %q", "modu-ai/moai-adk", out)
	}
}

func TestRenderRepoSegment_NilRepoReturnsEmpty(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Workspace: WorkspaceData{Repo: nil},
	}

	out := renderRepoSegment(data)
	if out != "" {
		t.Fatalf("expected empty output when Workspace.Repo is nil, got %q", out)
	}
}

func TestRenderRepoSegment_EmptyOwnerReturnsEmpty(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Workspace: WorkspaceData{Repo: &RepoInfo{Host: "github.com", Owner: "", Name: "moai-adk"}},
	}

	out := renderRepoSegment(data)
	if out != "" {
		t.Fatalf("expected empty output when Owner is empty, got %q", out)
	}
}

func TestRenderRepoSegment_EmptyNameReturnsEmpty(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Workspace: WorkspaceData{Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: ""}},
	}

	out := renderRepoSegment(data)
	if out != "" {
		t.Fatalf("expected empty output when Name is empty, got %q", out)
	}
}

func TestRenderRepoSegment_NilDataReturnsEmpty(t *testing.T) {
	t.Parallel()

	out := renderRepoSegment(nil)
	if out != "" {
		t.Fatalf("expected empty output when data is nil, got %q", out)
	}
}

// ---------- REQ-SSE-003/004: renderLongContextSegment + ExceedsLongTokens ----------

func TestRenderLongContextSegment_TrueShowsMarker(t *testing.T) {
	t.Parallel()

	data := &StatusData{ExceedsLongTokens: true}

	out := renderLongContextSegment(data)
	if !strings.Contains(out, "⚠️") || !strings.Contains(out, "long") {
		t.Fatalf("expected output to contain ⚠️ and 'long' markers, got %q", out)
	}
}

func TestRenderLongContextSegment_FalseReturnsEmpty(t *testing.T) {
	t.Parallel()

	data := &StatusData{ExceedsLongTokens: false}

	out := renderLongContextSegment(data)
	if out != "" {
		t.Fatalf("expected empty output when ExceedsLongTokens is false, got %q", out)
	}
}

func TestRenderLongContextSegment_NilDataReturnsEmpty(t *testing.T) {
	t.Parallel()

	out := renderLongContextSegment(nil)
	if out != "" {
		t.Fatalf("expected empty output when data is nil, got %q", out)
	}
}

// ---------- REQ-SSE-005/006: renderHandoffGuideSegment ----------

func TestRenderHandoffGuideSegment_OneMillionFiftyPercentShows(t *testing.T) {
	t.Parallel()

	// 1M budget, 50% usage = 500,000 tokens → should show
	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 1_000_000,
			TokensUsed:  500_000,
			Available:   true,
		},
	}

	out := renderHandoffGuideSegment(data)
	if out == "" {
		t.Fatalf("expected non-empty handoff guide for 1M @ 50%%, got empty")
	}
}

func TestRenderHandoffGuideSegment_OneMillionFortyNinePercentHidden(t *testing.T) {
	t.Parallel()

	// 1M budget, 49% usage = 490,000 tokens → should hide
	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 1_000_000,
			TokensUsed:  490_000,
			Available:   true,
		},
	}

	out := renderHandoffGuideSegment(data)
	if out != "" {
		t.Fatalf("expected empty handoff guide for 1M @ 49%%, got %q", out)
	}
}

func TestRenderHandoffGuideSegment_TwoHundredKNinetyPercentShows(t *testing.T) {
	t.Parallel()

	// 200K budget, 90% usage = 180,000 tokens → should show
	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 200_000,
			TokensUsed:  180_000,
			Available:   true,
		},
	}

	out := renderHandoffGuideSegment(data)
	if out == "" {
		t.Fatalf("expected non-empty handoff guide for 200K @ 90%%, got empty")
	}
}

func TestRenderHandoffGuideSegment_TwoHundredKEightyNinePercentHidden(t *testing.T) {
	t.Parallel()

	// 200K budget, 89% usage = 178,000 tokens → should hide
	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 200_000,
			TokensUsed:  178_000,
			Available:   true,
		},
	}

	out := renderHandoffGuideSegment(data)
	if out != "" {
		t.Fatalf("expected empty handoff guide for 200K @ 89%%, got %q", out)
	}
}

func TestRenderHandoffGuideSegment_UnknownBudgetHidden(t *testing.T) {
	t.Parallel()

	// Budget not in {200000, 1000000} (e.g., zero / unknown) → should hide
	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 0,
			TokensUsed:  100_000,
			Available:   false,
		},
	}

	out := renderHandoffGuideSegment(data)
	if out != "" {
		t.Fatalf("expected empty handoff guide for unknown budget, got %q", out)
	}
}

func TestRenderHandoffGuideSegment_NilDataReturnsEmpty(t *testing.T) {
	t.Parallel()

	out := renderHandoffGuideSegment(nil)
	if out != "" {
		t.Fatalf("expected empty handoff guide for nil data, got %q", out)
	}
}

// ---------- isXxxEnabled predicate edge cases (default-on contract) ----------

// TestIsRepoEnabled_DefaultOnContract verifies REQ-SSE-002 default-on
// activation across the three documented config states.
func TestIsRepoEnabled_DefaultOnContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  map[string]bool
		want bool
	}{
		{"nil config → enabled", nil, true},
		{"unset key → enabled (default-on)", map[string]bool{SegmentPR: true}, true},
		{"explicit false → disabled", map[string]bool{SegmentRepo: false}, false},
		{"explicit true → enabled", map[string]bool{SegmentRepo: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRenderer("default", true, tc.cfg)
			if got := r.isRepoEnabled(); got != tc.want {
				t.Errorf("isRepoEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestIsLongContextEnabled_DefaultOnContract verifies REQ-SSE-004 default-on.
func TestIsLongContextEnabled_DefaultOnContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  map[string]bool
		want bool
	}{
		{"nil config → enabled", nil, true},
		{"unset key → enabled", map[string]bool{SegmentPR: true}, true},
		{"explicit false → disabled", map[string]bool{SegmentLongContext: false}, false},
		{"explicit true → enabled", map[string]bool{SegmentLongContext: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRenderer("default", true, tc.cfg)
			if got := r.isLongContextEnabled(); got != tc.want {
				t.Errorf("isLongContextEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestIsHandoffGuideEnabled_DefaultOnContract verifies REQ-SSE-006 default-on.
func TestIsHandoffGuideEnabled_DefaultOnContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  map[string]bool
		want bool
	}{
		{"nil config → enabled", nil, true},
		{"unset key → enabled", map[string]bool{SegmentPR: true}, true},
		{"explicit false → disabled", map[string]bool{SegmentHandoffGuide: false}, false},
		{"explicit true → enabled", map[string]bool{SegmentHandoffGuide: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRenderer("default", true, tc.cfg)
			if got := r.isHandoffGuideEnabled(); got != tc.want {
				t.Errorf("isHandoffGuideEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestCollectAll_RepoAndExceedsLongPropagation verifies REQ-SSE-001/003 wiring
// from StdinData → StatusData via builder.collectAll.
func TestCollectAll_RepoAndExceedsLongPropagation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        *StdinData
		wantRepo     *RepoInfo
		wantExceeds  bool
	}{
		{
			name: "repo + exceeds_200k both present",
			input: &StdinData{
				Workspace: &WorkspaceInfo{Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"}},
				ExceedsLong: true,
			},
			wantRepo:    &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"},
			wantExceeds: true,
		},
		{
			name:        "both absent (backward compat)",
			input:       &StdinData{},
			wantRepo:    nil,
			wantExceeds: false,
		},
		{
			name:        "nil input safe",
			input:       nil,
			wantRepo:    nil,
			wantExceeds: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &defaultBuilder{
				renderer: NewRenderer("default", true, nil),
				mode:     ModeDefault,
			}
			data := b.collectAll(t.Context(), tt.input)
			// Repo
			if tt.wantRepo == nil {
				if data.Workspace.Repo != nil {
					t.Errorf("Workspace.Repo = %+v, want nil", data.Workspace.Repo)
				}
			} else {
				if data.Workspace.Repo == nil {
					t.Fatalf("Workspace.Repo is nil, want %+v", tt.wantRepo)
				}
				if data.Workspace.Repo.Owner != tt.wantRepo.Owner ||
					data.Workspace.Repo.Name != tt.wantRepo.Name {
					t.Errorf("Workspace.Repo = %+v, want %+v", data.Workspace.Repo, tt.wantRepo)
				}
			}
			// ExceedsLongTokens
			if data.ExceedsLongTokens != tt.wantExceeds {
				t.Errorf("ExceedsLongTokens = %v, want %v", data.ExceedsLongTokens, tt.wantExceeds)
			}
		})
	}
}

// ---------- REQ-SSE-007: PR segment graceful no-output ----------

// TestPRSegment_NilPRGracefulNoOutput verifies the AC-SSE-007 contract:
// when StatusData.PR is nil, no PR-specific markers appear in the L3 line.
// This is a regression guard against changes to the PR segment introduced
// by this SPEC.
func TestPRSegment_NilPRGracefulNoOutput(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)
	data := &StatusData{
		PR:        nil,
		Git:       GitStatusData{Branch: "main", Available: true},
		Directory: "moai-adk-go",
	}

	out := r.renderDirGitLine(data)

	// With PR nil, the rendered output must contain no PR review markers
	// ("⌥approved", "⌥pending", etc.) and no PR number anchor "#<digit>".
	if strings.Contains(out, "⌥approved") ||
		strings.Contains(out, "⌥pending") ||
		strings.Contains(out, "⌥changes_requested") ||
		strings.Contains(out, "⌥draft") {
		t.Fatalf("expected no PR review markers when data.PR is nil, got %q", out)
	}
}
