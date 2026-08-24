package statusline

import "testing"

// The branch half of the L3 repo+branch segment comes from local git and is
// available inside any repository. The forge half (📡 owner/name) comes from the
// optional workspace.repo sub-object on Claude Code's stdin payload, which is
// absent on payloads that do not carry it and on checkouts with no remote.
//
// Layout v3 CH3 merged the two into one segment. The segment inherited a
// hide-on-missing-repo rule written in 2026-05-22, when the segment carried only
// repo identity — so after the merge that rule silently took the branch down with
// it. These tests pin the two halves as independently renderable.

func repoBranchFixture() *StatusData {
	return &StatusData{
		Git: GitStatusData{
			Branch:    "main",
			Available: true,
			Modified:  7,
			Untracked: 21,
		},
	}
}

// TestRenderRepoBranchSegment_BranchSurvivesMissingForgeInfo is the regression
// guard: with no workspace.repo the forge half is correctly withheld, but the
// branch and its dirty count must still render.
func TestRenderRepoBranchSegment_BranchSurvivesMissingForgeInfo(t *testing.T) {
	t.Parallel()

	const want = "🅱️ main +28"

	t.Run("repo absent", func(t *testing.T) {
		t.Parallel()
		d := repoBranchFixture()
		if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
			t.Errorf("nil repo = %q, want %q", got, want)
		}
	})

	t.Run("repo present but incomplete", func(t *testing.T) {
		t.Parallel()
		for name, repo := range map[string]*RepoInfo{
			"no owner": {Host: "github.com", Name: "moai-adk"},
			"no name":  {Host: "github.com", Owner: "modu-ai"},
			"empty":    {},
		} {
			d := repoBranchFixture()
			d.Workspace = WorkspaceData{Repo: repo}
			if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
				t.Errorf("%s = %q, want %q", name, got, want)
			}
		}
	})
}

// TestRenderRepoBranchSegment_ForgeHalfStillJoinsWhenPresent pins the unchanged
// combined form, so restoring the branch does not cost the merged layout.
func TestRenderRepoBranchSegment_ForgeHalfStillJoinsWhenPresent(t *testing.T) {
	t.Parallel()

	d := repoBranchFixture()
	d.Workspace = WorkspaceData{Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"}}
	d.GitHub = GitHubCounts{OpenIssues: 9, OpenPRs: 4, Available: true}

	const want = "📡 modu-ai/moai-adk, 9/4 | 🅱️ main +28"
	if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
		t.Errorf("combined = %q, want %q", got, want)
	}
}

// TestRenderRepoBranchSegment_DirtySuffixOmittedWhenClean pins that a clean tree
// carries no "+N" tail, in both the forge-present and forge-absent forms.
func TestRenderRepoBranchSegment_DirtySuffixOmittedWhenClean(t *testing.T) {
	t.Parallel()

	clean := &StatusData{Git: GitStatusData{Branch: "main", Available: true}}
	if got := newTestRenderer().renderRepoBranchSegment(clean); got != "🅱️ main" {
		t.Errorf("clean, no repo = %q, want %q", got, "🅱️ main")
	}

	clean.Workspace = WorkspaceData{Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"}}
	clean.GitHub = GitHubCounts{Suppressed: true}
	if got := newTestRenderer().renderRepoBranchSegment(clean); got != "📡 modu-ai/moai-adk | 🅱️ main" {
		t.Errorf("clean, with repo = %q", got)
	}
}

// TestRenderRepoBranchSegment_DirtyCountSumsAllThreeCategories pins that the
// "+N" tail counts staged, modified, and untracked together — the number an
// operator reads as "uncommitted work".
func TestRenderRepoBranchSegment_DirtyCountSumsAllThreeCategories(t *testing.T) {
	t.Parallel()

	d := &StatusData{Git: GitStatusData{
		Branch:    "main",
		Available: true,
		Staged:    2,
		Modified:  3,
		Untracked: 5,
	}}
	if got := newTestRenderer().renderRepoBranchSegment(d); got != "🅱️ main +10" {
		t.Errorf("dirty sum = %q, want %q", got, "🅱️ main +10")
	}
}

// TestRenderRepoBranchSegment_NoGitContextStaysEmpty pins the two conditions that
// legitimately hide the whole segment: no repository, and no branch to name.
func TestRenderRepoBranchSegment_NoGitContextStaysEmpty(t *testing.T) {
	t.Parallel()

	cases := map[string]*StatusData{
		"nil data":          nil,
		"git unavailable":   {Git: GitStatusData{Branch: "main", Available: false}},
		"branch unresolved": {Git: GitStatusData{Branch: "", Available: true}},
	}
	for name, d := range cases {
		if got := newTestRenderer().renderRepoBranchSegment(d); got != "" {
			t.Errorf("%s = %q, want empty", name, got)
		}
	}
}

// TestRenderRepoBranchSegment_WorktreeMarkerRidesTheBranch pins that the [WT]
// marker stays attached to the branch when the forge half is absent — the
// worktree case is precisely where workspace.repo is most often missing.
func TestRenderRepoBranchSegment_WorktreeMarkerRidesTheBranch(t *testing.T) {
	t.Parallel()

	d := &StatusData{
		Git:      GitStatusData{Branch: "WT-mx-fanin-perf", Available: true, Modified: 1},
		Worktree: "/Users/x/proj/.claude/worktrees/perf",
	}
	if got := newTestRenderer().renderRepoBranchSegment(d); got != "🅱️ [WT] WT-mx-fanin-perf +1" {
		t.Errorf("worktree, no repo = %q", got)
	}

	off := NewRenderer("default", true, map[string]bool{SegmentWorktree: false})
	if got := off.renderRepoBranchSegment(d); got != "🅱️ WT-mx-fanin-perf +1" {
		t.Errorf("worktree segment off = %q", got)
	}
}
