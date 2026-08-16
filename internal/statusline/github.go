package statusline

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GitHubCountsTTL is how long a fetched count is served before a refresh is
// triggered. It trades staleness for calls: open issue and PR counts move on
// the scale of hours, and a shorter window would spend the API budget of every
// concurrent session on numbers that did not change.
const GitHubCountsTTL = 10 * time.Minute

// githubFetchBudget bounds the detached refresh. `gh` can hang on a captive
// portal or a stalled TLS handshake; without a ceiling that child would
// outlive the session that spawned it.
const githubFetchBudget = 20 * time.Second

// githubListLimit caps how many items `gh` enumerates per call. Repositories
// with more open items than this report the cap rather than the true count —
// an acceptable trade for a status bar, and far below any real backlog.
const githubListLimit = "1000"

// GitHubCounts is the repository's open work, as of the last successful fetch.
type GitHubCounts struct {
	OpenIssues int   `json:"open_issues"`
	OpenPRs    int   `json:"open_prs"`
	Available  bool  `json:"-"`          // false when no cache could be read
	FetchedAt  int64 `json:"fetched_at"` // unix seconds; 0 when never fetched
}

// githubCachePath returns where the counts cache lives for a board root.
func githubCachePath(boardRoot string) string {
	return filepath.Join(boardRoot, ".moai", "state", "github", "counts.json")
}

// resolveGitHubCounts reads the cached counts. Best-effort + fail-open: an
// absent, unreadable, or corrupt cache yields Available=false and renders
// nothing. Constant-cost — one read of one small file per statusline render.
//
// This NEVER calls the network. The render path only reads; refreshing is the
// detached child's job (see maybeRefreshGitHubCounts).
func resolveGitHubCounts(boardRoot string) GitHubCounts {
	if boardRoot == "" {
		return GitHubCounts{}
	}
	data, err := os.ReadFile(githubCachePath(boardRoot))
	if err != nil {
		return GitHubCounts{}
	}
	var c GitHubCounts
	if err := json.Unmarshal(data, &c); err != nil {
		return GitHubCounts{}
	}
	c.Available = true
	return c
}

// isSelfInvocable reports whether path names the moai binary — the only
// executable safe to re-invoke with `statusline --refresh-github`.
//
// Under `go test` the executable is the test binary, and handing it these
// arguments re-runs the whole suite — recursively, once per render, until the
// machine gives out. Any other host (a renamed binary, an embedding) degrades
// to a stale count, which is the right way to fail.
//
// Kept as a named predicate rather than an inline condition so the guard can be
// tested directly: exercising it through the spawn path would mean a broken
// guard turns its own regression test into the fork bomb it exists to prevent.
func isSelfInvocable(path string) bool {
	base := filepath.Base(path)
	return base == "moai" || base == "moai.exe"
}

// maybeRefreshGitHubCounts triggers a background refresh when the cache is
// missing or older than GitHubCountsTTL, and returns immediately either way.
// The caller keeps rendering the previous value — stale-while-revalidate, so a
// slow network degrades freshness rather than the status bar.
//
// The stampede guard is the cache file's own timestamp: the child rewrites it
// with a fresh FetchedAt before it starts fetching, so every render between the
// spawn and the result sees a fresh cache and spawns nothing.
func maybeRefreshGitHubCounts(boardRoot string) {
	if boardRoot == "" {
		return
	}
	cur := resolveGitHubCounts(boardRoot)
	if cur.Available && time.Since(time.Unix(cur.FetchedAt, 0)) < GitHubCountsTTL {
		return
	}

	self, err := os.Executable()
	if err != nil {
		return
	}
	if !isSelfInvocable(self) {
		return
	}
	cmd := exec.Command(self, "statusline", "--refresh-github", "--board-root", boardRoot)
	// Detach from this process's streams. A child holding the render's stdout
	// keeps the pipe open, which makes whatever is reading it wait for a
	// process it never knew about.
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	if err := cmd.Start(); err != nil {
		return
	}
	// Release rather than Wait: this is fire-and-forget, and the child bounds
	// itself with githubFetchBudget.
	_ = cmd.Process.Release()
}

// RefreshGitHubCounts fetches open issue and PR counts with `gh` and writes the
// cache under boardRoot. It is the detached child's entry point, never called
// on the render path.
//
// The timestamp is written BEFORE the fetch so concurrent renders stop spawning
// refreshes immediately, and the previous counts are preserved so a failed
// fetch degrades to a stale number rather than a blank segment.
func RefreshGitHubCounts(ctx context.Context, boardRoot string) error {
	if boardRoot == "" {
		return nil
	}

	prev := resolveGitHubCounts(boardRoot)
	prev.FetchedAt = time.Now().Unix()
	if err := writeGitHubCache(boardRoot, prev); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, githubFetchBudget)
	defer cancel()

	issues, errIssues := ghCount(ctx, boardRoot, "issue")
	prs, errPRs := ghCount(ctx, boardRoot, "pr")
	if errIssues != nil || errPRs != nil {
		return nil // keep the stale-but-timestamped cache; try again next TTL
	}

	return writeGitHubCache(boardRoot, GitHubCounts{
		OpenIssues: issues,
		OpenPRs:    prs,
		FetchedAt:  time.Now().Unix(),
	})
}

// ghCount runs `gh <kind> list` and returns how many open items it reported.
func ghCount(ctx context.Context, boardRoot, kind string) (int, error) {
	cmd := exec.CommandContext(ctx, "gh", kind, "list",
		"--state", "open", "--limit", githubListLimit, "--json", "number", "--jq", "length")
	cmd.Dir = boardRoot
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}

// writeGitHubCache writes the cache atomically so a render never reads a
// half-written file.
func writeGitHubCache(boardRoot string, c GitHubCounts) error {
	path := githubCachePath(boardRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
