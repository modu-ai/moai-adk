// landed.go — how many picked cards already have a landing commit on the
// branch the project integrates on.
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
// card into ONE subject-stream query — kanban's own landed scan
// (kanban.LandedScanArgs) — and attributes in memory through kanban's own
// predicate (kanban.LandedAttributions) and generation boundary
// (kanban.AutoDoneSubjectFresh). The count is the same subject criterion
// `moai todo auto-done` evaluates; this file owns no matcher of its own.
//
// The count answers "a landing commit exists", not "this card may close": a
// plan-only landing counts, and the auto-done close guards (reissued-id,
// SPEC-status) are not applied. That is why the render annotates with a
// verify-before-done glyph and never subtracts.
package statusline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// landedCriterion names the counting criterion a cache's number was produced
// under. A cache carrying any other value — including none, which is what the
// retired body-mention criterion wrote — is not a judgment under this one:
// it reads as unknown and is refreshed at once. Change the value whenever the
// criterion changes, so an old number can never render under a new meaning.
const landedCriterion = "subject-attribution/v1"

// landedGlyph prefixes the landed count in the TODO segment: U+2691 BLACK
// FLAG, "flagged — verify before done". Chosen to carry no completion
// connotation, to render as text in one cell (no Emoji property, East Asian
// Width N), and to read the same in every locale because it has no words.
const landedGlyph = "⚑"

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

// landedCardToken bounds which picked ids are considered at all, mirroring
// kanban's own card-token rule.
var landedCardToken = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// LandedCounts is how many picked cards the integration branch already carries
// a landing commit for, as of the last successful measurement.
//
// Available and Measured answer two different questions and only their
// combination is renderable. Available says a cache was read; Measured says
// that cache carries a NUMBER SOMEBODY OBSERVED rather than the timestamp the
// stampede guard writes before any work has happened. Rendering "0 landed" for
// either absence would assert a fact nobody measured — the same reason
// GitHubCounts renders "-/-" instead of "0/0", and the same reason
// kanban.LandingUnknown exists beside landed and not-landed.
type LandedCounts struct {
	// Landed is how many picked cards Ref's subject stream attributes, through
	// kanban's subject-attribution predicate, to a commit no older than the
	// card itself. A body mention does not count. A landing is not a close —
	// a plan-only landing counts too — which is why the render annotates with
	// a verify-before-done glyph and never subtracts.
	Landed int `json:"landed"`
	// Ref is the ref the count was measured against, recorded so a reader can
	// tell which branch answered.
	Ref string `json:"ref,omitempty"`
	// Measured is false for the timestamp-only placeholder the refresh writes
	// before it starts. Not a latch — every successful measurement sets it.
	Measured  bool  `json:"measured"`
	FetchedAt int64 `json:"fetched_at"` // unix seconds; 0 when never written
	// Criterion names the counting criterion that produced Landed (see
	// landedCriterion). Empty in caches written before it existed.
	Criterion string `json:"criterion,omitempty"`
	Available bool   `json:"-"` // false when no cache could be read
}

// Known reports whether a landed judgment actually exists under the current
// criterion. Everything else — absent cache, corrupt cache, un-measured
// placeholder, a number produced by a different criterion — is UNKNOWN, and
// unknown renders nothing at all.
func (c LandedCounts) Known() bool {
	return c.Available && c.Measured && c.Criterion == landedCriterion
}

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
// missing, older than LandedCountsTTL, or written under a different criterion,
// and returns immediately either way. The caller keeps rendering the previous
// value — stale-while-revalidate.
//
// The stampede guard is the cache file's own timestamp: the child writes a
// fresh FetchedAt (under the current criterion) before it starts, so every
// render between the spawn and the result sees a fresh cache and spawns
// nothing.
func maybeRefreshLandedCounts(boardRoot string) {
	if boardRoot == "" {
		return
	}
	cur := resolveLandedCounts(boardRoot)
	if cur.Available && cur.Criterion == landedCriterion &&
		time.Since(time.Unix(cur.FetchedAt, 0)) < LandedCountsTTL {
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

// landedScanRunner adapts landedGitRunner to kanban's CommandRunner contract,
// so kanban.ScanLandedSubjects runs its one query through the same seam the
// tests count. kanban's contract carries a command name; the adapter only
// ever runs git and refuses anything else rather than silently running git in
// its place.
func landedScanRunner(ctx context.Context, dir string) kanban.CommandRunner {
	return func(name string, args ...string) (string, error) {
		if name != "git" {
			return "", fmt.Errorf("statusline: landed scan runs git only, got %q", name)
		}
		return landedGitRunner(ctx, dir, args...)
	}
}

// RefreshLandedCounts measures how many picked cards the project's integration
// branch already carries a subject-attributed landing commit for, and writes
// the cache under boardRoot. It is the detached child's entry point, never
// called on the render path.
//
// ONE git query, whatever the card count. The per-card form
// (kanban.GitLandedQuerier.Landed) is correct and is what `moai todo pr` uses;
// it is simply the wrong shape behind a status bar.
func RefreshLandedCounts(ctx context.Context, boardRoot string) error {
	if boardRoot == "" {
		return nil
	}

	// The timestamp goes down first so concurrent renders stop spawning
	// immediately. A previous measurement under the CURRENT criterion rides
	// through unchanged: a failed query below must degrade to a
	// stale-but-observed number, and a first-ever failure must stay UNKNOWN
	// rather than become a zero. A previous measurement under any OTHER
	// criterion is dropped here, so a failure below cannot relabel the old
	// number as a current one.
	prev := resolveLandedCounts(boardRoot)
	if prev.Criterion != landedCriterion {
		prev = LandedCounts{Criterion: landedCriterion}
	}
	prev.FetchedAt = time.Now().Unix()
	if err := writeLandedCache(boardRoot, prev); err != nil {
		return err
	}

	picked := pickedCards(boardRoot)
	ref := kanban.LandedRefFor(boardRoot)
	if len(picked) == 0 {
		// Nothing in flight is an OBSERVED zero, reached without asking git
		// anything — renderable, unlike the unknowns above.
		return writeLandedCache(boardRoot, LandedCounts{
			Landed: 0, Ref: ref, Measured: true, FetchedAt: time.Now().Unix(),
			Criterion: landedCriterion,
		})
	}

	ctx, cancel := context.WithTimeout(ctx, landedScanBudget)
	defer cancel()

	commits, err := kanban.ScanLandedSubjects(landedScanRunner(ctx, boardRoot), ref)
	if err != nil {
		// A failed or malformed query: keep the stale-but-timestamped cache and
		// try again next TTL. Writing a zero here is exactly the fabricated
		// fact this segment refuses.
		return nil
	}

	attributed := kanban.LandedAttributions(commits, kanban.LandedBranchFromRef(ref))
	landed := 0
	for _, c := range picked {
		if hit, ok := attributed[c.ID]; ok && kanban.AutoDoneSubjectFresh(hit, c.AddedAt) {
			landed++
		}
	}

	return writeLandedCache(boardRoot, LandedCounts{
		Landed:    landed,
		Ref:       ref,
		Measured:  true,
		FetchedAt: time.Now().Unix(),
		Criterion: landedCriterion,
	})
}

// pickedCard is the part of a picked backlog item the landed count reads: its
// id, and the added_at the generation boundary judges against.
type pickedCard struct {
	ID      string
	AddedAt string
}

// pickedCards returns the cards in flight, read PURELY: a status refresh must
// never perform the queue's one-time storage cutover, which is why this uses
// LoadPure rather than Load.
func pickedCards(boardRoot string) []pickedCard {
	rec, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(boardRoot)).LoadPure()
	if err != nil || rec == nil {
		return nil
	}
	var cards []pickedCard
	for _, it := range rec.Items {
		if it.State == kanban.BacklogStatePicked && landedCardToken.MatchString(it.ID) {
			cards = append(cards, pickedCard{ID: it.ID, AddedAt: it.AddedAt})
		}
	}
	return cards
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
