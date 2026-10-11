package statusline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The four-state display contract of the forge pair (REQ-005 / F5,
// SPEC-STATUSLINE-PROFILE-RESPECT-001), characterized end-to-end: cache
// fixture → resolveGitHubCounts → renderRepoBranchSegment. No production
// change is expected here — this pins the contract the issue asked about
// ("0/0 seen during failed polling"), so a future render change that blurs
// unknown and zero fails a named test instead of an operator's eye.
func TestRender_ForgePairFourStateContract(t *testing.T) {
	t.Parallel()

	base := func() *StatusData {
		return &StatusData{
			Git:       GitStatusData{Branch: "main", Available: true},
			Workspace: WorkspaceData{Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"}},
		}
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T, root string)
		wantSub string // substring the repo segment must carry; "" = no pair
	}{
		{
			name: "fetched counts render as issues/PRs",
			setup: func(t *testing.T, root string) {
				t.Helper()
				if err := writeGitHubCache(root, GitHubCounts{OpenIssues: 7, OpenPRs: 3, FetchedAt: time.Now().Unix()}); err != nil {
					t.Fatal(err)
				}
			},
			wantSub: ", 7/3",
		},
		{
			name: "zeros a successful fetch produced are honest zeros",
			setup: func(t *testing.T, root string) {
				t.Helper()
				if err := writeGitHubCache(root, GitHubCounts{OpenIssues: 0, OpenPRs: 0, FetchedAt: time.Now().Unix()}); err != nil {
					t.Fatal(err)
				}
			},
			wantSub: ", 0/0",
		},
		{
			name: "stale zeros stay honest — a later rate-limited refresh must not rewrite them as unknown",
			setup: func(t *testing.T, root string) {
				t.Helper()
				if err := writeGitHubCache(root, GitHubCounts{OpenIssues: 0, OpenPRs: 0, FetchedAt: time.Now().Add(-time.Hour).Unix()}); err != nil {
					t.Fatal(err)
				}
			},
			wantSub: ", 0/0",
		},
		{
			name: "absent cache is unknown, not zero",
			setup: func(t *testing.T, root string) {
				t.Helper()
			},
			wantSub: ", -/-",
		},
		{
			name: "corrupt cache is unknown too",
			setup: func(t *testing.T, root string) {
				t.Helper()
				dir := filepath.Join(root, ".moai", "state", "github")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "counts.json"), []byte(`{"open_issues":`), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantSub: ", -/-",
		},
		{
			name: "suppressed from an explicit no-forge override renders no pair",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeForgeOverride(t, root, "none")
			},
			wantSub: "",
		},
		{
			name: "a child-written suppression verdict keeps the counts it rides with hidden",
			setup: func(t *testing.T, root string) {
				t.Helper()
				// The refresh child carries the previous counts through its
				// suppression write, so the cache can serialize counts AND
				// suppressed: true together (RefreshGitHubCounts writes prev
				// on a no-forge verdict). A valid override must not
				// resurrect them — Suppressed governs the render (card
				// t1583 ②): the contradiction state stays invisible.
				writeForgeOverride(t, root, "github")
				if err := writeGitHubCache(root, GitHubCounts{OpenIssues: 2, OpenPRs: 5, Suppressed: true, FetchedAt: time.Now().Unix()}); err != nil {
					t.Fatal(err)
				}
			},
			wantSub: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			tt.setup(t, root)

			d := base()
			d.GitHub = resolveGitHubCounts(root)
			got := newTestRenderer().renderRepoBranchSegment(d)

			if tt.wantSub == "" {
				if want := "📡 modu-ai/moai-adk | 🅱️ main"; got != want {
					t.Errorf("segment = %q, want %q (no pair)", got, want)
				}
				return
			}
			if !strings.Contains(got, tt.wantSub) {
				t.Errorf("segment = %q, want it to contain %q", got, tt.wantSub)
			}
		})
	}
}
