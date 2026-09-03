// landed.go — how many picked cards are already named in the branch the
// project integrates on.
//
// Shaped exactly like github.go, and deliberately so: one small cache read on
// the render path, a detached child past a TTL, the cache file's own timestamp
// as the stampede guard, and isSelfInvocable as the fork-bomb guard. A second
// mechanism for the same job is how the same logic ends up written three times.
//
// The live form of this question is what must never reach a render:
// kanban.GitLandedQuerier.Landed asks git ONCE PER CARD (measured 0.174s per
// query — about fourteen seconds across eighty cards), and backlog.go's read is
// contracted to stay constant-cost per render. The child therefore folds every
// card into ONE `git log <ref> --format=%B` and intersects in memory.
package statusline

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// LandedCountsTTL is how long a measurement is served before a refresh is
// triggered. The integration branch moves on the scale of a card's lifetime,
// and the annotation is a prompt to reconcile rather than a live readout, so a
// shorter window would spend a subprocess per session on a number that did not
// change.
const LandedCountsTTL = 10 * time.Minute

// landedScanBudget bounds the detached refresh. A `git log` over a large
// history is fast (measured 0.347s across 5,674 commits) but a wedged object
// store is not, and an unbounded child would outlive the session that spawned
// it.
const landedScanBudget = 20 * time.Second

// landedCardToken bounds what may be interpolated into the matching pattern,
// mirroring kanban's own card-token rule.
var landedCardToken = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// LandedCounts is how many picked cards the integration branch already names,
// as of the last successful measurement.
//
// Available and Measured answer two different questions and only their
// combination is renderable. Available says a cache was read; Measured says
// that cache carries a NUMBER SOMEBODY OBSERVED rather than the timestamp the
// stampede guard writes before any work has happened. Rendering "0 landed" for
// either absence would assert a fact nobody measured — the same reason
// GitHubCounts renders "-/-" instead of "0/0", and the same reason
// kanban.LandingUnknown exists beside landed and not-landed.
type LandedCounts struct {
	// Landed is how many picked cards are named in Ref's history. It counts
	// MENTIONS, not authorship: the shipped criterion is `\b<id>\b` over the
	// full commit message, so another card's report commit naming this one
	// counts. That is why the render annotates and never subtracts.
	Landed int `json:"landed"`
	// Ref is the ref the count was measured against, recorded so a reader can
	// tell which branch answered.
	Ref string `json:"ref,omitempty"`
	// Measured is false for the timestamp-only placeholder the refresh writes
	// before it starts. Not a latch — every successful measurement sets it.
	Measured  bool  `json:"measured"`
	FetchedAt int64 `json:"fetched_at"` // unix seconds; 0 when never written
	Available bool  `json:"-"`          // false when no cache could be read
}

// Known reports whether a landed judgment actually exists. Everything else —
// absent cache, corrupt cache, un-measured placeholder — is UNKNOWN, and
// unknown renders nothing at all.
func (c LandedCounts) Known() bool { return c.Available && c.Measured }

// landedCachePath returns where the landed cache lives for a board root.
func landedCachePath(boardRoot string) string {
	return filepath.Join(boardRoot, ".moai", "state", "landed", "counts.json")
}

// resolveLandedCounts reads the cached measurement. Best-effort + fail-open:
// an absent, unreadable, or corrupt cache yields Available=false, which renders
// as no annotation — unknown, not zero. Constant-cost: one read of one small
// file, and NEVER a subprocess.
func resolveLandedCounts(boardRoot string) LandedCounts {
	if boardRoot == "" {
		return LandedCounts{}
	}
	data, err := os.ReadFile(landedCachePath(boardRoot))
	if err != nil {
		return LandedCounts{}
	}
	var c LandedCounts
	if err := json.Unmarshal(data, &c); err != nil {
		return LandedCounts{}
	}
	c.Available = true
	return c
}

// landedSpawnProbe is a test seam, set only from tests. It is invoked at the
// single point a refresh child would be spawned — after the TTL freshness
// check, before the self-invocation guard — so a test can count exactly the
// "would have spawned" attempts. Placement is load-bearing: under `go test` the
// isSelfInvocable guard always blocks the real exec, so a counter at function
// entry could not tell pre-spawn gating from spawn-blocking (the same reason
// githubSpawnProbe sits where it does).
var landedSpawnProbe func(boardRoot string)

// maybeRefreshLandedCounts triggers a background refresh when the cache is
// missing or older than LandedCountsTTL, and returns immediately either way.
// The caller keeps rendering the previous value — stale-while-revalidate.
//
// The stampede guard is the cache file's own timestamp: the child writes a
// fresh FetchedAt before it starts, so every render between the spawn and the
// result sees a fresh cache and spawns nothing.
func maybeRefreshLandedCounts(boardRoot string) {
	if boardRoot == "" {
		return
	}
	cur := resolveLandedCounts(boardRoot)
	if cur.Available && time.Since(time.Unix(cur.FetchedAt, 0)) < LandedCountsTTL {
		return
	}

	if landedSpawnProbe != nil {
		landedSpawnProbe(boardRoot)
	}

	self, err := os.Executable()
	if err != nil || !isSelfInvocable(self) {
		return
	}
	cmd := exec.Command(self, "statusline", "--refresh-landed", "--board-root", boardRoot)
	// Detach from this process's streams: a child holding the render's stdout
	// keeps the pipe open, which makes whatever is reading it wait for a
	// process it never knew about.
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	if err := cmd.Start(); err != nil {
		return
	}
	_ = cmd.Process.Release() // fire-and-forget; the child bounds itself
}

// landedGitRunner runs the single git query. A package variable so the
// subprocess census can be counted in tests against the implementation's own
// call rather than a transcription of it.
var landedGitRunner = func(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}

// RefreshLandedCounts measures how many picked cards the project's integration
// branch already names, and writes the cache under boardRoot. It is the
// detached child's entry point, never called on the render path.
//
// ONE git invocation, whatever the card count. The per-card form
// (kanban.GitLandedQuerier.Landed) is correct and is what `moai todo pr` uses;
// it is simply the wrong shape behind a status bar.
func RefreshLandedCounts(ctx context.Context, boardRoot string) error {
	if boardRoot == "" {
		return nil
	}

	// The timestamp goes down first so concurrent renders stop spawning
	// immediately. The previous measurement rides through unchanged: a failed
	// query below must degrade to a stale-but-observed number, and a first-ever
	// failure must stay UNKNOWN rather than become a zero.
	prev := resolveLandedCounts(boardRoot)
	prev.FetchedAt = time.Now().Unix()
	if err := writeLandedCache(boardRoot, prev); err != nil {
		return err
	}

	picked := pickedCardIDs(boardRoot)
	ref := kanban.LandedRefFor(boardRoot)
	if len(picked) == 0 {
		// Nothing in flight is an OBSERVED zero, reached without asking git
		// anything — renderable, unlike the unknowns above.
		return writeLandedCache(boardRoot, LandedCounts{
			Landed: 0, Ref: ref, Measured: true, FetchedAt: time.Now().Unix(),
		})
	}

	ctx, cancel := context.WithTimeout(ctx, landedScanBudget)
	defer cancel()

	out, err := landedGitRunner(ctx, boardRoot, "log", ref, "--format=%B")
	if err != nil {
		// Keep the stale-but-timestamped cache; try again next TTL. Writing a
		// zero here is exactly the fabricated fact this segment refuses.
		return nil
	}

	return writeLandedCache(boardRoot, LandedCounts{
		Landed:    countNamed(out, picked),
		Ref:       ref,
		Measured:  true,
		FetchedAt: time.Now().Unix(),
	})
}

// pickedCardIDs returns the ids of the cards in flight, read PURELY: a status
// refresh must never perform the queue's one-time storage cutover, which is why
// this uses LoadPure rather than Load.
func pickedCardIDs(boardRoot string) []string {
	rec, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(boardRoot)).LoadPure()
	if err != nil || rec == nil {
		return nil
	}
	var ids []string
	for _, it := range rec.Items {
		if it.State == kanban.BacklogStatePicked && landedCardToken.MatchString(it.ID) {
			ids = append(ids, it.ID)
		}
	}
	return ids
}

// countNamed reports how many of ids the log body names, under the same
// word-boundary criterion `moai todo pr` ships (`\b<id>\b`). One alternation
// over one pass: the cost stays in the child either way, but a per-id scan
// would re-introduce the shape this file exists to avoid.
func countNamed(logBody string, ids []string) int {
	if len(ids) == 0 {
		return 0
	}
	quoted := make([]string, 0, len(ids))
	for _, id := range ids {
		quoted = append(quoted, regexp.QuoteMeta(id))
	}
	re, err := regexp.Compile(`\b(?:` + strings.Join(quoted, "|") + `)\b`)
	if err != nil {
		return 0
	}
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	seen := make(map[string]bool, len(ids))
	for _, m := range re.FindAllString(logBody, -1) {
		if wanted[m] {
			seen[m] = true
		}
	}
	return len(seen)
}

// writeLandedCache writes the cache atomically so a render never reads a
// half-written file.
func writeLandedCache(boardRoot string, c LandedCounts) error {
	path := landedCachePath(boardRoot)
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
