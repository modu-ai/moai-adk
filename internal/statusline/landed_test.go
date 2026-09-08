// landed_test.go — the landed-annotation segment: an unavailable judgment must
// render nothing rather than a zero nobody observed, the render path must stay
// free of git subprocesses, and the refresh child must fold every card into ONE
// git invocation.
//
// The last property is the one a "does it work?" test would pass for the
// per-card implementation this segment exists to avoid, so it is asserted as a
// COUNT that does not grow with the card set.
package statusline

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedPicked writes the given ids as picked cards under root.
func seedPicked(t *testing.T, root string, ids ...string) {
	t.Helper()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRootAdopting(root))
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for _, id := range ids {
			rec.LastSeq++
			rec.Items = append(rec.Items, kanban.BacklogItem{
				ID:      id,
				Text:    "card " + id,
				AddedAt: "2026-01-02T03:04:05Z",
				State:   kanban.BacklogStatePicked,
			})
		}
		return nil
	}); err != nil {
		t.Fatalf("seed picked: %v", err)
	}
}

// writeLanded puts a cache file in place verbatim, so a test can express a
// corrupt or partially-written cache the writer would never produce.
func writeLanded(t *testing.T, root string, body []byte) {
	t.Helper()
	path := landedCachePath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write cache: %v", err)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

// TestResolveLandedCounts_UnknownStates pins the unknown-vs-zero boundary: an
// absent, unreadable, or corrupt cache is UNKNOWN, and unknown must never
// present as a landed count of zero.
func TestResolveLandedCounts_UnknownStates(t *testing.T) {
	t.Parallel()

	t.Run("absent cache is unknown", func(t *testing.T) {
		t.Parallel()
		if got := resolveLandedCounts(t.TempDir()); got.Known() {
			t.Fatalf("absent cache reported known: %+v", got)
		}
	})

	t.Run("corrupt cache is unknown", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeLanded(t, root, []byte("{not json"))
		if got := resolveLandedCounts(root); got.Known() {
			t.Fatalf("corrupt cache reported known: %+v", got)
		}
	})

	t.Run("timestamp-only placeholder is unknown", func(t *testing.T) {
		t.Parallel()
		// The stampede guard writes the timestamp BEFORE the measurement
		// exists. That write must not become a renderable zero.
		root := t.TempDir()
		writeLanded(t, root, mustJSON(t, LandedCounts{FetchedAt: time.Now().Unix()}))
		if got := resolveLandedCounts(root); got.Known() {
			t.Fatalf("un-measured placeholder reported known: %+v", got)
		}
	})

	t.Run("empty board root is unknown", func(t *testing.T) {
		t.Parallel()
		if got := resolveLandedCounts(""); got.Known() {
			t.Fatalf("empty root reported known: %+v", got)
		}
	})
}

// TestResolveLandedCounts_ObservedZeroIsAFact is the other half of the
// boundary: a measured zero is real and must survive the round trip.
func TestResolveLandedCounts_ObservedZeroIsAFact(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeLanded(t, root, mustJSON(t, LandedCounts{
		Landed: 0, Measured: true, Ref: "origin/develop", FetchedAt: time.Now().Unix(),
	}))
	got := resolveLandedCounts(root)
	if !got.Known() || got.Landed != 0 {
		t.Fatalf("measured zero lost: %+v", got)
	}
}

// landedData is a session line whose backlog segment has something to say.
func landedData(l LandedCounts) *StatusData {
	return &StatusData{
		SessionName: "lane-1",
		Backlog:     BacklogCounts{Picked: 76, Queued: 4, Available: true},
		Landed:      l,
	}
}

// TestRenderer_LandedAnnotation pins the rendered form on both sides of the
// boundary. The unknown case must be BYTE-IDENTICAL to the pre-annotation
// segment — an unavailable judgment changes nothing on screen.
func TestRenderer_LandedAnnotation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		landed LandedCounts
		want   string
		// wantAnnotation is stated literally rather than read back from
		// Known(): a test whose expectation is computed by the predicate under
		// test asserts nothing about that predicate.
		wantAnnotation bool
	}{
		{
			name:   "unknown renders the segment unchanged",
			landed: LandedCounts{},
			want:   "🔄 TODO: 76/4",
		},
		{
			name:           "cache read but never measured renders unchanged",
			landed:         LandedCounts{Available: true, FetchedAt: 1},
			want:           "🔄 TODO: 76/4",
			wantAnnotation: false,
		},
		{
			name:           "measured count annotates",
			landed:         LandedCounts{Landed: 48, Measured: true, Available: true},
			want:           "🔄 TODO: 76/4 ✓48",
			wantAnnotation: true,
		},
		{
			name:           "observed zero annotates",
			landed:         LandedCounts{Landed: 0, Measured: true, Available: true},
			want:           "🔄 TODO: 76/4 ✓0",
			wantAnnotation: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := NewRenderer("default", true, nil).renderSessionLine(landedData(tc.landed))
			if !strings.Contains(got, tc.want) {
				t.Fatalf("segment = %q, want it to contain %q", got, tc.want)
			}
			// An unknown judgment must not merely omit the number — it must
			// not print a check mark at all.
			if hasAnnotation := strings.Contains(got, "✓"); hasAnnotation != tc.wantAnnotation {
				t.Fatalf("annotation present = %v, want %v (line: %q)",
					hasAnnotation, tc.wantAnnotation, got)
			}
		})
	}
}

// fakeGitOnPath installs a `git` on PATH that records every invocation, and
// returns the path of the record file.
func fakeGitOnPath(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	record := filepath.Join(binDir, "invocations")
	script := "#!/bin/sh\necho \"$@\" >> " + record + "\n"
	if err := os.WriteFile(filepath.Join(binDir, "git"), []byte(script), 0o755); err != nil { //nolint:gosec // deliberately executable test shim
		t.Fatalf("write shim: %v", err)
	}
	t.Setenv("PATH", binDir)
	return record
}

// TestLandedRenderPath_SpawnsNoGit asserts the render-path property
// MECHANICALLY rather than by reading the code: with a recording `git` as the
// only one on PATH, a resolve plus a render plus a refresh check on a fresh
// cache must leave no record behind.
func TestLandedRenderPath_SpawnsNoGit(t *testing.T) {
	record := fakeGitOnPath(t)

	root := t.TempDir()
	writeLanded(t, root, mustJSON(t, LandedCounts{
		Landed: 3, Measured: true, Ref: "origin/develop", FetchedAt: time.Now().Unix(),
	}))

	counts := resolveLandedCounts(root)
	_ = NewRenderer("default", true, nil).renderSessionLine(landedData(counts))
	maybeRefreshLandedCounts(root) // fresh cache: nothing to do

	if _, err := os.Stat(record); !errors.Is(err, os.ErrNotExist) {
		body, _ := os.ReadFile(record)
		t.Fatalf("render path invoked git:\n%s", body)
	}
}

// TestRefreshLandedCounts_OneInvocationRegardlessOfCardCount is the load-bearing
// assertion: the child folds every picked card into a SINGLE git invocation. A
// test that only checked the resulting number would pass for a per-card
// implementation, so the count is asserted AND asserted not to grow.
func TestRefreshLandedCounts_OneInvocationRegardlessOfCardCount(t *testing.T) {
	for _, n := range []int{3, 30} {
		var calls atomic.Int64
		ids := make([]string, 0, n)
		var body strings.Builder
		for i := 1; i <= n; i++ {
			id := "t" + itoa(1000+i)
			ids = append(ids, id)
			if i%2 == 0 { // half of them are named in history
				body.WriteString("feat(" + id + "): something landed\n\n")
			}
		}

		root := t.TempDir()
		seedPicked(t, root, ids...)

		restore := landedGitRunner
		landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
			calls.Add(1)
			return body.String(), nil
		}
		if err := RefreshLandedCounts(context.Background(), root); err != nil {
			t.Fatalf("refresh: %v", err)
		}
		landedGitRunner = restore

		if got := calls.Load(); got != 1 {
			t.Fatalf("cards=%d: git invocations = %d, want exactly 1", n, got)
		}
		got := resolveLandedCounts(root)
		if !got.Known() || got.Landed != n/2 {
			t.Fatalf("cards=%d: landed = %+v, want %d", n, got, n/2)
		}
	}
}

// TestRefreshLandedCounts_TimestampPrecedesTheWork pins the stampede guard:
// by the time the git call runs, the cache already carries a fresh timestamp
// and is still un-measured.
func TestRefreshLandedCounts_TimestampPrecedesTheWork(t *testing.T) {
	root := t.TempDir()
	seedPicked(t, root, "t900")

	var seen LandedCounts
	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		seen = resolveLandedCounts(root)
		return "fix(t900): landed\n", nil
	}
	defer func() { landedGitRunner = restore }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if seen.FetchedAt == 0 || time.Since(time.Unix(seen.FetchedAt, 0)) > time.Minute {
		t.Fatalf("timestamp was not written before the work: %+v", seen)
	}
	if seen.Known() {
		t.Fatalf("placeholder was renderable mid-refresh: %+v", seen)
	}
}

// TestRefreshLandedCounts_FailedQueryKeepsThePriorMeasurement: a git failure
// degrades to the stale-but-observed number, never to a fabricated zero.
func TestRefreshLandedCounts_FailedQueryKeepsThePriorMeasurement(t *testing.T) {
	root := t.TempDir()
	seedPicked(t, root, "t901")
	writeLanded(t, root, mustJSON(t, LandedCounts{Landed: 7, Measured: true, FetchedAt: 1}))

	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("no such ref")
	}
	defer func() { landedGitRunner = restore }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	got := resolveLandedCounts(root)
	if !got.Known() || got.Landed != 7 {
		t.Fatalf("stale measurement lost on a failed query: %+v", got)
	}
}

// TestRefreshLandedCounts_NeverMeasuredStaysUnknownOnFailure is the same case
// without a prior measurement — the one that must NOT become "✓0".
func TestRefreshLandedCounts_NeverMeasuredStaysUnknownOnFailure(t *testing.T) {
	root := t.TempDir()
	seedPicked(t, root, "t902")

	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("no git here")
	}
	defer func() { landedGitRunner = restore }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if got := resolveLandedCounts(root); got.Known() {
		t.Fatalf("a failed first measurement became renderable: %+v", got)
	}
}

// TestMaybeRefreshLandedCounts_StampedeGuard: a second attempt inside the TTL
// does not spawn. The probe sits where the github one does — after the TTL
// check, before the self-invocation guard — because under `go test` the real
// exec is always blocked and a naive counter cannot tell gating from blocking.
func TestMaybeRefreshLandedCounts_StampedeGuard(t *testing.T) {
	var spawns atomic.Int64
	restore := landedSpawnProbe
	landedSpawnProbe = func(string) { spawns.Add(1) }
	defer func() { landedSpawnProbe = restore }()

	root := t.TempDir()

	maybeRefreshLandedCounts(root) // no cache at all: must attempt
	if got := spawns.Load(); got != 1 {
		t.Fatalf("cold cache: spawn attempts = %d, want 1", got)
	}

	// Simulate the child's first act: the timestamp lands before the work.
	writeLanded(t, root, mustJSON(t, LandedCounts{FetchedAt: time.Now().Unix()}))
	maybeRefreshLandedCounts(root)
	maybeRefreshLandedCounts(root)
	if got := spawns.Load(); got != 1 {
		t.Fatalf("inside TTL: spawn attempts = %d, want no further attempt", got)
	}

	// Past the TTL the question is asked again.
	writeLanded(t, root, mustJSON(t, LandedCounts{
		Landed: 1, Measured: true, FetchedAt: time.Now().Add(-2 * LandedCountsTTL).Unix(),
	}))
	maybeRefreshLandedCounts(root)
	if got := spawns.Load(); got != 2 {
		t.Fatalf("past TTL: spawn attempts = %d, want 2", got)
	}
}

// TestRefreshLandedCounts_EmptyQueueNeedsNoQuery: zero picked cards is an
// OBSERVED zero — renderable, and reached without asking git anything.
func TestRefreshLandedCounts_EmptyQueueNeedsNoQuery(t *testing.T) {
	root := t.TempDir()

	var calls atomic.Int64
	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		calls.Add(1)
		return "", nil
	}
	defer func() { landedGitRunner = restore }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("empty queue asked git %d time(s), want 0", got)
	}
	got := resolveLandedCounts(root)
	if !got.Known() || got.Landed != 0 {
		t.Fatalf("empty queue: %+v, want an observed zero", got)
	}
}

// TestRefreshLandedCounts_AsksTheConfiguredRef: the ref is resolved, never
// hardcoded — asking origin/main in a project that integrates on develop
// answers a question nobody posed.
func TestRefreshLandedCounts_AsksTheConfiguredRef(t *testing.T) {
	root := t.TempDir()
	seedPicked(t, root, "t903")

	var gotArgs []string
	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, args ...string) (string, error) {
		gotArgs = args
		return "", nil
	}
	defer func() { landedGitRunner = restore }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	want := kanban.LandedRefFor(root)
	if !namesArg(gotArgs, want) {
		t.Fatalf("git args %v do not name the resolved ref %q", gotArgs, want)
	}
}

func namesArg(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// TestCountNamed_WordBoundaryCriterion pins the matching criterion itself.
//
// countNamed IS the feature: everything else is caching around it. Its failure
// mode is the silent one internal/kanban/prlink_landed.go's file header
// documents — a matcher that matches nothing returns a clean, error-free count
// byte-identical to "nothing landed". A prefix collision is where the
// alternation `\b(?:id1|id2|...)\b` would go wrong if the engine picked the
// shorter branch, failed the trailing boundary, and gave up instead of trying
// the longer one, so that case is pinned rather than assumed.
//
// The boundary-adjacency and repeat cases are the positive control: without
// them a matcher that always returned 0 would satisfy every negative case.
func TestCountNamed_WordBoundaryCriterion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		ids  []string
		body string
		want int
	}{
		{
			name: "prefix collision, both named",
			ids:  []string{"t45", "t456"},
			body: "fix(t45): one thing\n\nfeat(t456): another\n",
			want: 2,
		},
		{
			name: "prefix collision, only the longer named",
			ids:  []string{"t45", "t456"},
			body: "feat(t456): only the longer card is here\n",
			want: 1, // and specifically NOT 2 — t45 must not match inside t456
		},
		{
			name: "a substring is not a mention",
			ids:  []string{"t45"},
			body: "feat(t456): a different card entirely\n",
			want: 0,
		},
		{
			name: "start of body",
			ids:  []string{"t336"},
			body: "t336 landed",
			want: 1,
		},
		{
			name: "end of body",
			ids:  []string{"t336"},
			body: "landed as t336",
			want: 1,
		},
		{
			name: "parenthesised scope",
			ids:  []string{"t336"},
			body: "fix(t336): repair the thing\n",
			want: 1,
		},
		{
			name: "trailing colon",
			ids:  []string{"t336"},
			body: "t336: repair the thing\n",
			want: 1,
		},
		{
			name: "sentence punctuation",
			ids:  []string{"t336"},
			body: "docs: record the verdict for card t336.\n",
			want: 1,
		},
		{
			name: "hyphen is a boundary, not part of the token",
			ids:  []string{"t336"},
			body: "chore: revert t336-followup\n",
			want: 1,
		},
		{
			name: "many mentions of one card count once",
			ids:  []string{"t336"},
			body: "feat(t336): a\n\nfix(t336): b\n\ndocs(t336): c\n",
			want: 1,
		},
		{
			name: "no ids is zero without touching the body",
			ids:  nil,
			body: "feat(t336): a\n",
			want: 0,
		},
		{
			name: "an empty history names nobody",
			ids:  []string{"t336", "t337"},
			body: "",
			want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := countNamed(tc.body, tc.ids); got != tc.want {
				t.Fatalf("countNamed(ids=%v) = %d, want %d\nbody:\n%s", tc.ids, got, tc.want, tc.body)
			}
		})
	}
}
