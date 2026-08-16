package statusline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The GitHub counts are the one segment backed by a network call, so what
// matters is the behaviour when that call has not happened, has failed, or has
// left a damaged file behind. The render path must never be what notices.

func TestResolveGitHubCounts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string // empty means: write no file at all
		wantIssues int
		wantPRs    int
		wantAvail  bool
	}{
		{
			name:       "valid cache",
			body:       `{"open_issues":7,"open_prs":3,"fetched_at":1700000000}`,
			wantIssues: 7, wantPRs: 3, wantAvail: true,
		},
		{
			name:      "zero counts are still a real answer",
			body:      `{"open_issues":0,"open_prs":0,"fetched_at":1700000000}`,
			wantAvail: true,
		},
		{name: "corrupt json fails open", body: `{"open_issues":`},
		{name: "never fetched fails open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tt.body != "" {
				dir := filepath.Join(root, ".moai", "state", "github")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "counts.json"), []byte(tt.body), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			got := resolveGitHubCounts(root)
			if got.Available != tt.wantAvail {
				t.Errorf("Available = %v, want %v", got.Available, tt.wantAvail)
			}
			if got.OpenIssues != tt.wantIssues || got.OpenPRs != tt.wantPRs {
				t.Errorf("issues/prs = %d/%d, want %d/%d",
					got.OpenIssues, got.OpenPRs, tt.wantIssues, tt.wantPRs)
			}
		})
	}
}

func TestWriteGitHubCache_RoundTrips(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := GitHubCounts{OpenIssues: 12, OpenPRs: 4, FetchedAt: 1700000000}
	if err := writeGitHubCache(root, want); err != nil {
		t.Fatal(err)
	}

	got := resolveGitHubCounts(root)
	if !got.Available || got.OpenIssues != want.OpenIssues ||
		got.OpenPRs != want.OpenPRs || got.FetchedAt != want.FetchedAt {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
}

// A fresh cache must not trigger a refresh. That guard is what keeps a render
// running every few hundred milliseconds from spawning a process each time —
// without it this feature becomes a fork bomb with a status bar attached.
func TestMaybeRefreshGitHubCounts_FreshCacheSpawnsNothing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := writeGitHubCache(root, GitHubCounts{OpenIssues: 1, FetchedAt: time.Now().Unix()}); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(githubCachePath(root))
	if err != nil {
		t.Fatal(err)
	}

	maybeRefreshGitHubCounts(root)

	after, err := os.Stat(githubCachePath(root))
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Error("fresh cache was rewritten; the TTL guard did not hold")
	}
}

// The binary-name guard is the one that keeps a `go test` run from re-invoking
// its own test binary with these arguments — which re-runs the whole suite,
// once per render, recursively. It is exercised directly rather than through
// maybeRefreshGitHubCounts because a regression there would make this test the
// fork bomb instead of catching it.
func TestIsSelfInvocable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{"/Users/x/go/bin/moai", true},
		{"moai", true},
		// Bare, because filepath.Base is separator-aware per OS: a
		// backslash path only splits on Windows, so asserting one here
		// would test the host rather than the guard.
		{"moai.exe", true},
		{"/tmp/go-build123/b001/statusline.test", false}, // the `go test` case
		{"/usr/local/bin/moai-adk", false},               // a renamed host
		{"/usr/local/bin/claude", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			if got := isSelfInvocable(tt.path); got != tt.want {
				t.Errorf("isSelfInvocable(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// The guard must hold for the binary this suite actually runs as, otherwise
// every other test in this file is passing for the wrong reason.
func TestIsSelfInvocable_RejectsThisTestBinary(t *testing.T) {
	t.Parallel()

	self, err := os.Executable()
	if err != nil {
		t.Skipf("executable path unavailable: %v", err)
	}
	if isSelfInvocable(self) {
		t.Fatalf("test binary %q classified as self-invocable; a refresh would re-run this suite", self)
	}
}

// An empty board root means we could not work out where state lives. Spawning
// a refresh against "" would run `gh` in whatever directory the process happens
// to be in, so the guard is a real one rather than defensive noise.
func TestMaybeRefreshGitHubCounts_EmptyRootIsANoOp(t *testing.T) {
	t.Parallel()

	maybeRefreshGitHubCounts("") // must not panic and must not spawn

	if got := resolveGitHubCounts(""); got.Available {
		t.Errorf("empty root must yield no counts, got %+v", got)
	}
}

func TestRenderSessionLine_GitHubCounts(t *testing.T) {
	t.Parallel()

	d := namedData()
	d.GitHub = GitHubCounts{OpenIssues: 7, OpenPRs: 3, Available: true}

	got := NewRenderer("default", true, nil).renderSessionLine(d)
	if !strings.Contains(got, "🐛 7 / 📥 3") {
		t.Errorf("github segment missing from %q", got)
	}

	d.GitHub = GitHubCounts{} // never fetched
	got = NewRenderer("default", true, nil).renderSessionLine(d)
	if strings.Contains(got, "🐛") {
		t.Errorf("unfetched counts must render nothing, got %q", got)
	}
}

func TestRenderSessionLine_GitHubSegmentDisablable(t *testing.T) {
	t.Parallel()

	d := namedData()
	d.GitHub = GitHubCounts{OpenIssues: 7, OpenPRs: 3, Available: true}

	got := NewRenderer("default", true, map[string]bool{SegmentGitHub: false}).renderSessionLine(d)
	if strings.Contains(got, "🐛") {
		t.Errorf("disabled github segment still rendered: %q", got)
	}
	if !strings.Contains(got, "🔄 12 / ⤵️ 26") {
		t.Errorf("backlog must survive disabling the github segment: %q", got)
	}
}
