package statusline

import (
	"regexp"
	"strings"
	"testing"
)

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

// Ahead/behind — how many commits the local branch holds that the remote does
// not, and vice versa — is collected on GitStatusData and was rendered on the
// branch as ↑N/↓N until a layout change moved it to a slash pair on the repo,
// and a later one reassigned that pair to open issues / open PRs, leaving the
// counts collected but rendered nowhere.
//
// The constraint that removal was protecting is real and is preserved here: only
// ONE number/number pair may appear on the line, because two slash pairs side by
// side were misread. Arrows are not a slash pair, so ↑3 ↓1 restores the counts
// without reintroducing the ambiguity.

// TestRenderRepoBranchSegment_AheadBehindRendersAsArrows pins the restored form.
func TestRenderRepoBranchSegment_AheadBehindRendersAsArrows(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		ahead, behind int
		want          string
	}{
		{"ahead only", 3, 0, "🅱️ main ↑3"},
		{"behind only", 0, 259, "🅱️ main ↓259"},
		{"both", 1, 259, "🅱️ main ↑1 ↓259"},
		{"synced — no arrows", 0, 0, "🅱️ main"},
	}
	for _, tc := range cases {
		d := &StatusData{Git: GitStatusData{
			Branch:    "main",
			Available: true,
			Ahead:     tc.ahead,
			Behind:    tc.behind,
		}}
		if got := newTestRenderer().renderRepoBranchSegment(d); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestRenderRepoBranchSegment_ArrowsPrecedeDirtyCount pins the ordering: the
// arrows describe committed history, the "+N" tail describes uncommitted work, and
// they read left to right in that order.
func TestRenderRepoBranchSegment_ArrowsPrecedeDirtyCount(t *testing.T) {
	t.Parallel()

	d := repoBranchFixture() // 7 modified + 21 untracked
	d.Git.Ahead = 1
	d.Git.Behind = 259

	const want = "🅱️ main ↑1 ↓259 +28"
	if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
		t.Errorf("no repo = %q, want %q", got, want)
	}

	d.Workspace = WorkspaceData{Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"}}
	d.GitHub = GitHubCounts{OpenIssues: 9, OpenPRs: 4, Available: true}
	const wantFull = "📡 modu-ai/moai-adk, 9/4 | 🅱️ main ↑1 ↓259 +28"
	if got := newTestRenderer().renderRepoBranchSegment(d); got != wantFull {
		t.Errorf("with repo = %q, want %q", got, wantFull)
	}
}

// TestRenderRepoBranchSegment_OnlyOneSlashPair is the guard for the constraint the
// earlier removal was protecting: the forge counts are the line's only "N/N", so
// restoring ahead/behind must not add a second one.
func TestRenderRepoBranchSegment_OnlyOneSlashPair(t *testing.T) {
	t.Parallel()

	d := repoBranchFixture()
	d.Git.Ahead = 59
	d.Git.Behind = 0
	d.Workspace = WorkspaceData{Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"}}
	d.GitHub = GitHubCounts{OpenIssues: 9, OpenPRs: 4, Available: true}

	got := newTestRenderer().renderRepoBranchSegment(d)
	// The owner/name slash is not a number pair; only "N/N" reads as counts.
	pairs := regexp.MustCompile(`\d+/\d+`).FindAllString(got, -1)
	if len(pairs) != 1 {
		t.Errorf("segment %q carries number pairs %v, want exactly 1 (the forge counts)", got, pairs)
	}
	if strings.Contains(got, "59/0") {
		t.Errorf("segment %q renders ahead/behind as a slash pair — the misread this guards against", got)
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

// TestRenderRepoBranchSegment_WorktreeSessionKeepsItsBranch covers the case the
// branch-survival fix matters most for: a worktree session, where workspace.repo
// is most often absent from the payload. The branch reads as the WT-<slug>
// convention alone — the "[WT] " marker was retired, and this stays green either
// way because it asserts the branch is present rather than how it is decorated.
func TestRenderRepoBranchSegment_WorktreeSessionKeepsItsBranch(t *testing.T) {
	t.Parallel()

	d := &StatusData{
		Git:      GitStatusData{Branch: "WT-mx-fanin-perf", Available: true, Modified: 1, Ahead: 6},
		Worktree: "/Users/x/proj/.claude/worktrees/perf",
	}
	const want = "🅱️ WT-mx-fanin-perf ↑6 +1"
	if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
		t.Errorf("worktree, no repo = %q, want %q", got, want)
	}
	if got := NewRenderer("default", true, map[string]bool{SegmentWorktree: false}).renderRepoBranchSegment(d); got != want {
		t.Errorf("worktree segment off = %q, want %q (the segment toggle no longer decorates the branch)", got, want)
	}
}
