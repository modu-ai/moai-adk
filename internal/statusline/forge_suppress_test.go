package statusline

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Suppression is the state the four-state model was missing: a checkout that
// no forge will ever answer for. It must render like a switched-off segment,
// not like "-/-", which promises an answer that is still coming.

// TestResolveGitHubCounts_ExplicitOptOutSuppresses: `statusline.forge: none`
// is a decision, not a failure. The data layer learns it from the same small
// config read the refresh already uses, so the pair disappears on the very
// first render rather than only after a child has run.
func TestResolveGitHubCounts_ExplicitOptOutSuppresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		forge         string // "" means: write no statusline.yaml at all
		cache         string // "" means: write no cache at all
		wantSuppress  bool
		wantAvailable bool
	}{
		{name: "forge none, no cache", forge: "none", wantSuppress: true},
		{name: "forge off, no cache", forge: "off", wantSuppress: true},
		{
			name: "forge none beats a populated cache", forge: "none",
			cache:        `{"open_issues":7,"open_prs":3,"fetched_at":1700000000}`,
			wantSuppress: true, wantAvailable: true,
		},
		{name: "a typo names no forge either", forge: "githbu", wantSuppress: true},
		{
			name: "a real forge does not suppress", forge: "github",
			cache:         `{"open_issues":7,"open_prs":3,"fetched_at":1700000000}`,
			wantAvailable: true,
		},
		{
			name:          "no config at all leaves detection to the child",
			cache:         `{"open_issues":7,"open_prs":3,"fetched_at":1700000000}`,
			wantAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tt.forge != "" {
				writeForgeOverride(t, root, tt.forge)
			}
			if tt.cache != "" {
				dir := filepath.Join(root, ".moai", "state", "github")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "counts.json"), []byte(tt.cache), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			got := resolveGitHubCounts(root)
			if got.Suppressed != tt.wantSuppress {
				t.Errorf("Suppressed = %v, want %v", got.Suppressed, tt.wantSuppress)
			}
			if got.Available != tt.wantAvailable {
				t.Errorf("Available = %v, want %v", got.Available, tt.wantAvailable)
			}
		})
	}
}

// TestRender_ForgePairSuppressedRendersNoPair: a suppressed checkout renders
// exactly like a switched-off segment. "-/-" is reserved for a state the next
// refresh could change on its own.
func TestRender_ForgePairSuppressedRendersNoPair(t *testing.T) {
	t.Parallel()

	base := func() *StatusData {
		return &StatusData{
			Git:       GitStatusData{Branch: "main", Available: true},
			Workspace: WorkspaceData{Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"}},
		}
	}
	const want = "📡 modu-ai/moai-adk | 🅱️ main"

	// Suppressed with nothing cached — the opted-out user's steady state.
	d := base()
	d.GitHub = GitHubCounts{Suppressed: true}
	if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
		t.Errorf("suppressed, no cache = %q, want %q", got, want)
	}

	// Suppressed with counts still cached from before the opt-out. The user
	// asked for no forge data; stale numbers are still forge data.
	d = base()
	d.GitHub = GitHubCounts{OpenIssues: 7, OpenPRs: 3, Available: true, Suppressed: true}
	if got := newTestRenderer().renderRepoBranchSegment(d); got != want {
		t.Errorf("suppressed, stale cache = %q, want %q", got, want)
	}

	// It renders identically to the segment being switched off — the two are
	// the same fact arrived at two ways.
	off := base()
	off.GitHub = GitHubCounts{OpenIssues: 7, OpenPRs: 3, Available: true}
	gotOff := NewRenderer("default", true, map[string]bool{SegmentGitHub: false}).renderRepoBranchSegment(off)
	if gotOff != want {
		t.Errorf("segment off = %q, want %q", gotOff, want)
	}
}

// TestRefreshGitHubCounts_NoForgeWritesSuppressed: with no forge for this
// checkout the child writes the fact down, so later renders read it from the
// cache rather than re-deriving it — and never print "-/-".
func TestRefreshGitHubCounts_NoForgeWritesSuppressed(t *testing.T) {
	root := t.TempDir()
	writeForgeOverride(t, root, "none")

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}

	got := resolveGitHubCounts(root)
	if !got.Suppressed {
		t.Errorf("counts = %+v, want Suppressed — no forge can ever answer here", got)
	}
}

// TestRefreshGitHubCounts_MissingCLIWritesSuppressed: a repository on a
// recognised forge whose CLI is not installed cannot be counted by waiting, so
// the pair is dropped rather than left as "-/-" forever. Suppression is data
// the child rewrites every TTL, so installing the CLI restores the pair on its
// own.
func TestRefreshGitHubCounts_MissingCLIWritesSuppressed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "github")
	seedStaleCache(t, root, 7, 3)
	t.Setenv("PATH", t.TempDir()) // an empty PATH: no gh, no git

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}

	got := resolveGitHubCounts(root)
	if !got.Suppressed {
		t.Errorf("counts = %+v, want Suppressed — the forge CLI is not installed", got)
	}
}

// TestRefreshGitHubCounts_SuccessClearsSuppressed: suppression records the
// last refresh, it does not latch. Once the forge answers, the pair returns.
func TestRefreshGitHubCounts_SuccessClearsSuppressed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "github")
	stale := GitHubCounts{Suppressed: true, FetchedAt: time.Now().Add(-time.Hour).Unix()}
	if err := writeGitHubCache(root, stale); err != nil {
		t.Fatal(err)
	}

	stubDir := writeForgeStub(t, "gh", 42, 17, false)
	t.Setenv("PATH", stubDir+":"+os.Getenv("PATH"))

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}

	got := resolveGitHubCounts(root)
	if got.Suppressed {
		t.Errorf("counts = %+v, want Suppressed cleared after a successful fetch", got)
	}
	if got.OpenIssues != 42 || got.OpenPRs != 17 {
		t.Errorf("counts = %d/%d, want 42/17", got.OpenIssues, got.OpenPRs)
	}
}

// TestRefreshGitHubCounts_FailedFetchDoesNotSuppress: a forge that is present
// but errors is exactly the "-/-" case — the next TTL may succeed, so the pair
// must stay, carrying the stale numbers.
func TestRefreshGitHubCounts_FailedFetchDoesNotSuppress(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}

	root := t.TempDir()
	writeForgeOverride(t, root, "github")
	seedStaleCache(t, root, 7, 3)

	stubDir := writeForgeStub(t, "gh", 0, 0, true) // every call exits 1
	t.Setenv("PATH", stubDir+":"+os.Getenv("PATH"))

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts: %v", err)
	}

	got := resolveGitHubCounts(root)
	if got.Suppressed {
		t.Errorf("counts = %+v, want Suppressed unset — a failing forge is still a forge", got)
	}
	if got.OpenIssues != 7 || got.OpenPRs != 3 {
		t.Errorf("counts = %d/%d, want the stale 7/3 preserved", got.OpenIssues, got.OpenPRs)
	}
}
