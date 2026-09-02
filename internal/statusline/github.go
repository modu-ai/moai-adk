package statusline

import (
	"context"
	"encoding/json"
	"fmt"
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
//
// Available and Suppressed answer two different questions and only their
// combination is renderable: Available says whether a number was read,
// Suppressed says whether a number was ever going to exist. A checkout with no
// forge is not a checkout whose fetch failed, and rendering the two alike is
// the defect this pair exists to prevent.
type GitHubCounts struct {
	OpenIssues int  `json:"open_issues"`
	OpenPRs    int  `json:"open_prs"`
	Available  bool `json:"-"` // false when no cache could be read
	// Suppressed marks a checkout no forge will answer for: the operator set
	// `statusline.forge: none` (or a value naming no forge), the remote is on
	// no recognised host, or the forge's CLI is not installed. The pair is then
	// dropped entirely rather than shown as "-/-", which would promise an
	// answer that is not coming.
	//
	// Serialized: the refresh child pays the git and PATH lookups that decide
	// it, so the render path reads the verdict instead of re-deriving it. The
	// explicit config opt-out is the one case the read side decides for itself,
	// since it must hold before any child has run.
	//
	// Not a latch — every refresh rewrites it, so installing the CLI or naming
	// a forge brings the pair back within one TTL.
	Suppressed bool  `json:"suppressed,omitempty"`
	FetchedAt  int64 `json:"fetched_at"` // unix seconds; 0 when never fetched
}

// githubCachePath returns where the counts cache lives for a board root.
func githubCachePath(boardRoot string) string {
	return filepath.Join(boardRoot, ".moai", "state", "github", "counts.json")
}

// resolveGitHubCounts reads the cached counts. Best-effort + fail-open: an
// absent, unreadable, or corrupt cache yields Available=false, which renders as
// "-/-" — unknown, not zero, and not absent. Constant-cost: one read of one
// small cache file, plus one read of the small config file below.
//
// The config read is what lets an explicit opt-out take effect immediately. The
// refresh child decides every other suppression case and writes its verdict
// into the cache, but `statusline.forge: none` must hold on the very first
// render of a checkout that has never had a child run — otherwise the operator
// who just switched the counts off watches "-/-" until a TTL elapses. Only the
// override is consulted here; the remote host and the CLI's presence cost a
// subprocess each and stay on the child's side of the line.
//
// This NEVER calls the network. The render path only reads; refreshing is the
// detached child's job (see maybeRefreshGitHubCounts).
func resolveGitHubCounts(boardRoot string) GitHubCounts {
	if boardRoot == "" {
		return GitHubCounts{}
	}
	var c GitHubCounts
	if data, err := os.ReadFile(githubCachePath(boardRoot)); err == nil {
		if err := json.Unmarshal(data, &c); err == nil {
			c.Available = true
		} else {
			c = GitHubCounts{}
		}
	}
	if override := forgeOverride(boardRoot); override != "" {
		if _, ok := resolveForge("", override); !ok {
			// The value names no forge — "none", "off", or a typo. All three
			// mean the operator will not get counts here, and an absent pair is
			// the symptom a typo should show (see resolveForge).
			c.Suppressed = true
		}
	}
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

// githubSpawnProbe is a test seam, set only from tests. When non-nil it is
// invoked at the single point a refresh child would be spawned — after the TTL
// freshness check (and the explicit-override check), before the self-invocation
// guard — so a test can count exactly the "would have spawned" attempts.
// Placement is load-bearing: under `go test` the isSelfInvocable guard always
// blocks the real exec, so a counter at function entry cannot tell pre-spawn
// gating from spawn-blocking — the distinction the opt-out gates exist to make
// (SPEC-STATUSLINE-PROFILE-RESPECT-001 acceptance §D, kickoff decision D3).
var githubSpawnProbe func(boardRoot string)

// maybeRefreshGitHubCounts triggers a background refresh when the cache is
// missing or older than GitHubCountsTTL, and returns immediately either way.
// The caller keeps rendering the previous value — stale-while-revalidate, so a
// slow network degrades freshness rather than the status bar.
//
// The stampede guard is the cache file's own timestamp: the child rewrites it
// with a fresh FetchedAt before it starts fetching, so every render between the
// spawn and the result sees a fresh cache and spawns nothing.
//
// @MX:NOTE: [AUTO] two opt-out gates reach this spawn: the segment gate at the
// builder call site (REQ-001) and the explicit no-forge override below
// (REQ-002) — both must keep the not-a-latch contract of GitHubCounts.Suppressed.
func maybeRefreshGitHubCounts(boardRoot string) {
	if boardRoot == "" {
		return
	}
	cur := resolveGitHubCounts(boardRoot)
	if cur.Available && time.Since(time.Unix(cur.FetchedAt, 0)) < GitHubCountsTTL {
		return
	}

	// An explicit override naming no forge is a decision the config read above
	// already carries (cur.Suppressed). A child here could only re-derive it —
	// at the cost of a spawn per TTL across every concurrent session, which is
	// the polling the operator's opt-out was ordering stopped
	// (SPEC-STATUSLINE-PROFILE-RESPECT-001 REQ-002). An UNSET override must not
	// take this exit: detection belongs to the child, and a child-written
	// Suppressed verdict is not a latch — gating on it would stop the very
	// refresh that notices the CLI being installed.
	if override := forgeOverride(boardRoot); override != "" {
		if _, ok := resolveForge("", override); !ok {
			return
		}
	}

	if githubSpawnProbe != nil {
		githubSpawnProbe(boardRoot)
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

// RefreshGitHubCounts fetches open issue and change-request counts from the
// checkout's forge and writes the cache under boardRoot. It is the detached
// child's entry point, never called on the render path.
//
// The timestamp is written BEFORE the fetch so concurrent renders stop spawning
// refreshes immediately, and the previous counts are preserved so a failed
// fetch degrades to a stale number rather than a blank segment.
//
// Forge resolution happens here rather than in the caller because it costs a
// `git remote` call, and the render path is the one place that must stay a
// single file read.
//
// @MX:DEBT: this entry point, its cache path, and GitHubCounts still say
// "github" now that GitLab is served by the same code
// @MX:CEILING: two forges — the shape is identical, so the name misleads a
// reader without misleading the code
// @MX:UPGRADE: rename to Forge* when a third forge lands, or when the cache
// file is next versioned for another reason (renaming it alone would orphan
// every existing cache to buy nothing)
func RefreshGitHubCounts(ctx context.Context, boardRoot string) error {
	if boardRoot == "" {
		return nil
	}

	prev := resolveGitHubCounts(boardRoot)
	prev.FetchedAt = time.Now().Unix()
	// The timestamp goes down first, but the suppression verdict rides through
	// this write unchanged. Clearing it here would open a window the width of a
	// `git remote` call in which a render reads "not suppressed" for a checkout
	// that has no forge at all, and draws the stale counts underneath it. The
	// verdict this run reaches is written below, once it has one.
	if err := writeGitHubCache(boardRoot, prev); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, githubFetchBudget)
	defer cancel()

	forge, ok := resolveForge(originRemoteURL(ctx, boardRoot), forgeOverride(boardRoot))
	if !ok {
		// No forge for this checkout — opted out, or a remote on a host nothing
		// recognises. Record it so the pair is dropped rather than shown as
		// "-/-": nothing about waiting will produce a count here.
		prev.Suppressed = true
		return writeGitHubCache(boardRoot, prev)
	}
	if _, err := exec.LookPath(forge.bin); err != nil {
		// The forge exists but its CLI does not. Same verdict, same reason: the
		// next refresh cannot succeed either, and "-/-" would keep offering an
		// answer that will not arrive. Installing the CLI clears this on the
		// next TTL, since every run re-decides the flag from these two checks.
		prev.Suppressed = true
		return writeGitHubCache(boardRoot, prev)
	}

	// Both checks answered: a forge serves this checkout and its CLI is
	// installed. Any suppression verdict from an earlier run is stale now — a
	// forge that appeared, or a CLI installed since, must not stay hidden — so
	// clear it here, where the clearing is backed by evidence.
	prev.Suppressed = false
	if err := writeGitHubCache(boardRoot, prev); err != nil {
		return err
	}

	if !forgeBudgetAllows(ctx, boardRoot, forge) {
		return nil // near the API ceiling; keep the stale count rather than spend the last of it
	}

	issues, prs, err := forgeCounts(ctx, boardRoot, forge)
	if err != nil {
		return nil // keep the stale-but-timestamped cache; try again next TTL
	}

	return writeGitHubCache(boardRoot, GitHubCounts{
		OpenIssues: issues,
		OpenPRs:    prs,
		FetchedAt:  time.Now().Unix(),
	})
}

// forgeRateFloor is how much API budget must remain before a refresh is
// allowed to spend any. A refresh costs a single point on the forge that
// reports a budget at all, so this is not self-protection: it keeps a status
// bar from being the thing that consumes the last of a budget the operator's
// own tooling is about to need. Well below any level the TTL could reach on
// its own — at ten minutes a repository spends about six points an hour
// against a five-thousand ceiling.
const forgeRateFloor = 100

// forgeBudgetAllows reports whether the forge has enough API budget left for a
// refresh.
//
// Fail-open in both directions that matter: a forge with no rateArgs (nothing
// free to ask) and a query that errors or answers unparseably both return
// true. An unanswerable question about the budget is not evidence the budget
// is gone, and treating it as such would silence the segment on every forge
// that cannot be asked.
//
// The query itself must be free — on GitHub `gh api rate_limit` consumes
// neither the core nor the graphql bucket (measured 2026-08-20: three
// consecutive reads, zero delta in both) — otherwise the guard would spend
// exactly what it exists to conserve.
func forgeBudgetAllows(ctx context.Context, boardRoot string, forge forgeSpec) bool {
	if len(forge.rateArgs) == 0 {
		return true
	}
	cmd := exec.CommandContext(ctx, forge.bin, forge.rateArgs...)
	cmd.Dir = boardRoot
	out, err := cmd.Output()
	if err != nil {
		return true
	}
	remaining, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return true
	}
	return remaining >= forgeRateFloor
}

// forgeCounts returns the open issue and change-request counts, preferring the
// forge's single totals call and falling back to counting two listings.
//
// The totals path parses two lines from one call — issues first — which is the
// order forgeSpec.countArgs documents. A forge answering fewer than two lines
// is treated as a failed fetch rather than as a zero, since a zero would
// overwrite good cached numbers with a parse accident.
func forgeCounts(ctx context.Context, boardRoot string, forge forgeSpec) (int, int, error) {
	if len(forge.countArgs) == 0 {
		issues, err := forgeCount(ctx, boardRoot, forge, "issue")
		if err != nil {
			return 0, 0, err
		}
		prs, err := forgeCount(ctx, boardRoot, forge, "pr")
		if err != nil {
			return 0, 0, err
		}
		return issues, prs, nil
	}

	cmd := exec.CommandContext(ctx, forge.bin, forge.countArgs...)
	cmd.Dir = boardRoot
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(string(out))
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("forge totals: want two numbers, got %q", strings.TrimSpace(string(out)))
	}
	issues, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, err
	}
	prs, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, err
	}
	return issues, prs, nil
}

// forgeCount runs one listing and returns how many open items it reported.
// Every forge's argument list ends in a filter that collapses the listing to a
// bare integer, so this parses one number without knowing which CLI answered.
// It is the enumerating fallback — see forgeCounts.
func forgeCount(ctx context.Context, boardRoot string, forge forgeSpec, kind string) (int, error) {
	cmd := exec.CommandContext(ctx, forge.bin, forge.argsFor(kind)...)
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
