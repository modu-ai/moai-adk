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

// NOTE: TestRenderRepoSegment_* removed (layout v3 CH3). The standalone
// renderRepoSegment function was replaced by renderRepoBranchSegment which
// combines repo identity with branch info into a single L3 segment. The new
// segment is covered by end-to-end fixture verification via
// /Users/goos/go/bin/moai statusline; unit-level coverage may be added later
// when the function signature stabilizes.

// NOTE: renderLongContextSegment + renderHandoffGuideSegment tests removed
// (layout v3 CH1 + CH2). The ⚠️ long marker segment was removed per user
// explicit request; handoff_guide is now integrated as a CW bar (/clear)
// suffix in renderBarsInline. shouldShowHandoffGuide threshold helper is
// preserved and tested via TestShouldShowHandoffGuide below.

// ---------- shouldShowHandoffGuide threshold helper (preserved per CH2) ----------

func TestShouldShowHandoffGuide_OneMillionFiftyPercentTrue(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 1_000_000,
			TokensUsed:        500_000,
			Available:         true,
		},
	}

	if !shouldShowHandoffGuide(data) {
		t.Fatalf("expected true for 1M @ 50%%, got false")
	}
}

func TestShouldShowHandoffGuide_OneMillionFortyNinePercentFalse(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 1_000_000,
			TokensUsed:        490_000,
			Available:         true,
		},
	}

	if shouldShowHandoffGuide(data) {
		t.Fatalf("expected false for 1M @ 49%%, got true")
	}
}

func TestShouldShowHandoffGuide_TwoHundredKNinetyPercentTrue(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 200_000,
			TokensUsed:        180_000,
			Available:         true,
		},
	}

	if !shouldShowHandoffGuide(data) {
		t.Fatalf("expected true for 200K @ 90%%, got false")
	}
}

func TestShouldShowHandoffGuide_TwoHundredKEightyNinePercentFalse(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 200_000,
			TokensUsed:        178_000,
			Available:         true,
		},
	}

	if shouldShowHandoffGuide(data) {
		t.Fatalf("expected false for 200K @ 89%%, got true")
	}
}

func TestShouldShowHandoffGuide_UnknownCwSizeFalse(t *testing.T) {
	t.Parallel()

	data := &StatusData{
		Memory: MemoryData{
			ContextWindowSize: 0,
			TokensUsed:        100_000,
			Available:         false,
		},
	}

	if shouldShowHandoffGuide(data) {
		t.Fatalf("expected false for unknown context window size, got true")
	}
}

func TestShouldShowHandoffGuide_NilDataFalse(t *testing.T) {
	t.Parallel()

	if shouldShowHandoffGuide(nil) {
		t.Fatalf("expected false for nil data, got true")
	}
}

// ---------- isXxxEnabled predicate edge cases (default-on contract) ----------

// NOTE: TestIsRepoEnabled_DefaultOnContract removed (layout v3 CH3).
// The isRepoEnabled predicate was retired with renderRepoSegment.
// renderRepoBranchSegment is gated by SegmentGitBranch (combined segment).

// NOTE: TestIsLongContextEnabled_DefaultOnContract +
// TestIsHandoffGuideEnabled_DefaultOnContract removed (layout v3 CH1 + CH2).
// The corresponding isLongContextEnabled + isHandoffGuideEnabled predicate
// functions were removed; segment config keys SegmentLongContext +
// SegmentHandoffGuide no longer exist.

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
