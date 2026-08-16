package statusline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The statusline is the only durable place a session's identity can be shown:
// Claude Code drops its own prompt-bar name chip after /clear even though the
// explicit name (--name / /rename) is retained, and the directory segment shows
// the worktree once a session enters one. The session line carries identity and
// board workload together; these tests pin what it renders and, just as
// importantly, what it leaves out.

func namedData() *StatusData {
	return &StatusData{
		SessionName: "Team-A-Lead",
		AgentName:   "manager-kanban",
		Directory:   "statusline-session-name",
		Backlog:     BacklogCounts{Picked: 12, Queued: 26, Available: true},
	}
}

func TestRenderSessionLine_OrdersIdentityThenWorkload(t *testing.T) {
	t.Parallel()

	got := NewRenderer("default", true, nil).renderSessionLine(namedData())

	for _, want := range []string{"🏷️ Team-A-Lead", "👤 manager-kanban", "🔄 12 / ⤵️ 26"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}

	// Order is load-bearing: identity is what the operator scans for when
	// several sessions are open, and the workload only means something once
	// you know whose it is.
	nameAt := strings.Index(got, "Team-A-Lead")
	agentAt := strings.Index(got, "manager-kanban")
	boardAt := strings.Index(got, "🔄 ")
	if !(nameAt < agentAt && agentAt < boardAt) {
		t.Errorf("want name < agent < backlog, got %d/%d/%d in %q", nameAt, agentAt, boardAt, got)
	}
}

func TestRenderSessionLine_EmptyWhenNothingToSay(t *testing.T) {
	t.Parallel()

	got := NewRenderer("default", true, nil).renderSessionLine(&StatusData{Directory: "moai-adk-go"})

	if got != "" {
		t.Errorf("a session with no identity and no backlog must render no line, got %q", got)
	}
}

func TestRenderSessionLine_UnreadableBacklogRendersNoCounts(t *testing.T) {
	t.Parallel()

	d := namedData()
	d.Backlog = BacklogCounts{} // Available=false — the fail-open outcome

	got := NewRenderer("default", true, nil).renderSessionLine(d)

	if strings.Contains(got, "🔄 ") {
		t.Errorf("unavailable backlog must render no counts, got %q", got)
	}
	if !strings.Contains(got, "🏷️ Team-A-Lead") {
		t.Errorf("identity lost when backlog unavailable: %q", got)
	}
}

func TestRenderSessionLine_SegmentsDisablable(t *testing.T) {
	t.Parallel()

	cfg := map[string]bool{SegmentSession: false}
	got := NewRenderer("default", true, cfg).renderSessionLine(namedData())
	if strings.Contains(got, "Team-A-Lead") || strings.Contains(got, "manager-kanban") {
		t.Errorf("disabled session segment still rendered: %q", got)
	}
	if !strings.Contains(got, "🔄 ") {
		t.Errorf("backlog must survive disabling the identity segment: %q", got)
	}

	cfg = map[string]bool{SegmentBacklog: false}
	got = NewRenderer("default", true, cfg).renderSessionLine(namedData())
	if strings.Contains(got, "🔄 ") {
		t.Errorf("disabled backlog segment still rendered: %q", got)
	}
	if !strings.Contains(got, "🏷️ Team-A-Lead") {
		t.Errorf("identity must survive disabling the backlog segment: %q", got)
	}
}

// The identity previously lived on the directory line and then on the info
// line. Leaving a copy behind would double it on every status update, so pin
// its absence from both rather than trusting the moves stayed clean.
func TestOtherLines_CarryNoSessionIdentity(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)

	dirLine := r.renderDirGitLine(namedData())
	if strings.Contains(dirLine, "🏷️") || strings.Contains(dirLine, "👤") {
		t.Errorf("identity leaked onto the directory line: %q", dirLine)
	}
	if !strings.Contains(dirLine, "📁 statusline-session-name") {
		t.Errorf("directory segment lost: %q", dirLine)
	}

	infoLine := r.renderInfoLine(namedData(), false)
	if strings.Contains(infoLine, "🏷️") || strings.Contains(infoLine, "👤") {
		t.Errorf("identity leaked onto the info line: %q", infoLine)
	}
}

func TestResolveBacklogCounts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string // empty means: write no file at all
		wantPicked int
		wantQueued int
		wantAvail  bool
	}{
		{
			name:       "counts picked and queued, ignores dropped",
			body:       `{"items":[{"state":"picked"},{"state":"queued"},{"state":"queued"},{"state":"dropped"}]}`,
			wantPicked: 1, wantQueued: 2, wantAvail: true,
		},
		{
			name: "empty board is available but zero",
			body: `{"items":[]}`, wantAvail: true,
		},
		{
			name:       "unknown states are not counted",
			body:       `{"items":[{"state":"archived"},{"state":"picked"}]}`,
			wantPicked: 1, wantAvail: true,
		},
		{name: "corrupt json fails open", body: `{"items":`},
		{name: "absent file fails open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tt.body != "" {
				dir := filepath.Join(root, ".moai", "state", "kanban")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "backlog.json"), []byte(tt.body), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			got := resolveBacklogCounts(root)
			if got.Available != tt.wantAvail {
				t.Errorf("Available = %v, want %v", got.Available, tt.wantAvail)
			}
			if got.Picked != tt.wantPicked || got.Queued != tt.wantQueued {
				t.Errorf("picked/queued = %d/%d, want %d/%d",
					got.Picked, got.Queued, tt.wantPicked, tt.wantQueued)
			}
		})
	}
}

// A worktree session must read the PRIMARY checkout's board: `.moai/state/` is
// gitignored, so it does not exist in the worktree at all. Getting this wrong
// shows an empty board to exactly the sessions doing the work.
func TestResolveBoardRoot_PrefersPrimaryCheckoutOverWorktree(t *testing.T) {
	t.Parallel()

	worktreeSession := &StdinData{
		Workspace: &WorkspaceInfo{CurrentDir: "/repo/.claude/worktrees/card-x"},
		Worktree:  &WorktreeInfo{Path: "/repo/.claude/worktrees/card-x", OriginalCwd: "/repo"},
	}
	if got := resolveBoardRoot(worktreeSession); got != "/repo" {
		t.Errorf("worktree session board root = %q, want /repo", got)
	}

	primarySession := &StdinData{Workspace: &WorkspaceInfo{CurrentDir: "/repo"}}
	if got := resolveBoardRoot(primarySession); got != "/repo" {
		t.Errorf("primary session board root = %q, want /repo", got)
	}
}
