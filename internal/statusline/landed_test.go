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
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mattn/go-runewidth"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedPicked writes the given ids as picked cards under root, each added at
// landedAddedAt.
func seedPicked(t *testing.T, root string, ids ...string) {
	t.Helper()
	seedPickedAt(t, root, landedAddedAt, ids...)
}

// seedPickedAt writes the given ids as picked cards under root with an
// explicit added_at — the generation boundary reads it.
func seedPickedAt(t *testing.T, root, addedAt string, ids ...string) {
	t.Helper()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRootAdopting(root))
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for _, id := range ids {
			rec.LastSeq++
			rec.Items = append(rec.Items, kanban.BacklogItem{
				ID:      id,
				Text:    "card " + id,
				AddedAt: addedAt,
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
		Criterion: landedCriterion,
	}))
	got := resolveLandedCounts(root)
	if !got.Known() || got.Landed != 0 {
		t.Fatalf("measured zero lost: %+v", got)
	}
}

// oldSchemaCache is a cache as the retired body-mention criterion wrote it:
// no criterion field at all.
func oldSchemaCache(fetchedAt int64) []byte {
	return []byte(`{"landed":9,"ref":"origin/develop","measured":true,"fetched_at":` +
		strconv.FormatInt(fetchedAt, 10) + `}`)
}

// TestResolveLandedCounts_OldCriterionIsUnknown is AC-SLL-005: a measured
// number written under the retired criterion is not a judgment under the
// current one, so it renders nothing — neither the new glyph nor the old.
func TestResolveLandedCounts_OldCriterionIsUnknown(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeLanded(t, root, oldSchemaCache(time.Now().Unix()))

	got := resolveLandedCounts(root)
	if got.Known() {
		t.Fatalf("old-criterion cache reported known: %+v", got)
	}
	line := NewRenderer("default", true, nil).renderSessionLine(landedData(got))
	if strings.Contains(line, landedGlyph) || strings.Contains(line, "✓") {
		t.Fatalf("old-criterion cache annotated the line: %q", line)
	}
	if !strings.Contains(line, "🔄 TODO: 76/4") || strings.Contains(line, "🔄 TODO: 76/4 ") {
		t.Fatalf("segment is not the unannotated pair: %q", line)
	}

	// A different, non-empty criterion is just as unknown as an absent one.
	other := t.TempDir()
	writeLanded(t, other, mustJSON(t, LandedCounts{
		Landed: 4, Measured: true, FetchedAt: time.Now().Unix(), Criterion: "body-mention/v0",
	}))
	if got := resolveLandedCounts(other); got.Known() {
		t.Fatalf("foreign-criterion cache reported known: %+v", got)
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
			landed:         LandedCounts{Landed: 48, Measured: true, Available: true, Criterion: landedCriterion},
			want:           "🔄 TODO: 76/4 ⚑48",
			wantAnnotation: true,
		},
		{
			name:           "observed zero annotates",
			landed:         LandedCounts{Landed: 0, Measured: true, Available: true, Criterion: landedCriterion},
			want:           "🔄 TODO: 76/4 ⚑0",
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
			// not print the glyph at all.
			if hasAnnotation := strings.Contains(got, "⚑"); hasAnnotation != tc.wantAnnotation {
				t.Fatalf("annotation present = %v, want %v (line: %q)",
					hasAnnotation, tc.wantAnnotation, got)
			}
			// The check mark reads "done"; it must not appear in any case.
			if strings.Contains(got, "✓") {
				t.Fatalf("line carries a check mark: %q", got)
			}
		})
	}
}

// TestRenderer_LandedNeverSubtracts is AC-SLL-009: the picked/queued pair is
// the same whether the landed count is known or not.
func TestRenderer_LandedNeverSubtracts(t *testing.T) {
	t.Parallel()
	r := NewRenderer("default", true, nil)
	known := r.renderSessionLine(landedData(LandedCounts{
		Landed: 48, Measured: true, Available: true, Criterion: landedCriterion,
	}))
	unknown := r.renderSessionLine(landedData(LandedCounts{}))
	const prefix = "🔄 TODO: 76/4"
	for name, line := range map[string]string{"known": known, "unknown": unknown} {
		if !strings.Contains(line, prefix) {
			t.Fatalf("%s line lost the unreduced pair %q: %q", name, prefix, line)
		}
	}
}

// TestLandedGlyph_SingleCellLocaleNeutral is AC-SLL-010: one code point,
// U+2691, one terminal cell whether or not the terminal treats East Asian
// ambiguous characters as wide.
func TestLandedGlyph_SingleCellLocaleNeutral(t *testing.T) {
	t.Parallel()
	runes := []rune(landedGlyph)
	if len(runes) != 1 || runes[0] != '⚑' {
		t.Fatalf("landedGlyph = %q (%U), want exactly U+2691", landedGlyph, runes)
	}
	for _, eaw := range []bool{false, true} {
		cond := runewidth.NewCondition()
		cond.EastAsianWidth = eaw
		if w := cond.StringWidth(landedGlyph); w != 1 {
			t.Fatalf("EastAsianWidth=%v: width = %d, want 1", eaw, w)
		}
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
		Criterion: landedCriterion,
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
// assertion (AC-SLL-004): the child folds every picked card into a SINGLE git
// invocation, and that invocation is kanban's own subject-scan argv. A test
// that only checked the resulting number would pass for a per-card
// implementation, so the count is asserted AND asserted not to grow.
func TestRefreshLandedCounts_OneInvocationRegardlessOfCardCount(t *testing.T) {
	for _, n := range []int{1, 50} {
		var calls atomic.Int64
		var gotArgs []string
		ids := make([]string, 0, n)
		var body strings.Builder
		want := 0
		for i := 1; i <= n; i++ {
			id := "t" + itoa(1000+i)
			ids = append(ids, id)
			if i%2 == 1 { // odd ones carry an attributing subject
				body.WriteString(scanLine("sha"+id, landedFreshCT, "feat("+id+"): something landed"))
				want++
			}
		}

		root := t.TempDir()
		seedPicked(t, root, ids...)

		restore := landedGitRunner
		landedGitRunner = func(_ context.Context, _ string, args ...string) (string, error) {
			calls.Add(1)
			gotArgs = args
			return body.String(), nil
		}
		if err := RefreshLandedCounts(context.Background(), root); err != nil {
			t.Fatalf("refresh: %v", err)
		}
		landedGitRunner = restore

		if got := calls.Load(); got != 1 {
			t.Fatalf("cards=%d: git invocations = %d, want exactly 1", n, got)
		}
		wantArgs := kanban.LandedScanArgs(kanban.LandedRefFor(root))
		if strings.Join(gotArgs, "\x1f") != strings.Join(wantArgs, "\x1f") {
			t.Fatalf("cards=%d: argv = %q, want kanban's scan argv %q", n, gotArgs, wantArgs)
		}
		got := resolveLandedCounts(root)
		if !got.Known() || got.Landed != want {
			t.Fatalf("cards=%d: landed = %+v, want %d", n, got, want)
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
		return scanLine("sha900", landedFreshCT, "fix(t900): landed"), nil
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

// TestRefreshLandedCounts_FailedQueryKeepsThePriorMeasurement is AC-SLL-007: a
// failed or malformed query over a current-criterion cache degrades to the
// stale-but-observed number with an advanced timestamp, never to a
// fabricated zero.
func TestRefreshLandedCounts_FailedQueryKeepsThePriorMeasurement(t *testing.T) {
	cases := []struct {
		name string
		run  func(context.Context, string, ...string) (string, error)
	}{
		{
			name: "runner error",
			run: func(_ context.Context, _ string, _ ...string) (string, error) {
				return "", errors.New("no such ref")
			},
		},
		{
			name: "malformed stream (no separator)",
			run: func(_ context.Context, _ string, _ ...string) (string, error) {
				return "feat(t901): a line with no NUL separators\n", nil
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			seedPicked(t, root, "t901")
			writeLanded(t, root, mustJSON(t, LandedCounts{
				Landed: 3, Measured: true, FetchedAt: 1, Criterion: landedCriterion,
			}))

			restore := landedGitRunner
			landedGitRunner = tc.run
			defer func() { landedGitRunner = restore }()

			if err := RefreshLandedCounts(context.Background(), root); err != nil {
				t.Fatalf("refresh: %v", err)
			}
			got := resolveLandedCounts(root)
			if !got.Known() || got.Landed != 3 {
				t.Fatalf("stale measurement lost on a failed query: %+v", got)
			}
			if got.FetchedAt <= 1 {
				t.Fatalf("timestamp not advanced: %+v", got)
			}
		})
	}
}

// TestRefreshLandedCounts_OldCriterionNeverSurvivesAFailure is AC-SLL-008 /
// AC-SLL-006: a failed refresh over an old-criterion cache must not carry the
// old number forward as a current-criterion measurement.
func TestRefreshLandedCounts_OldCriterionNeverSurvivesAFailure(t *testing.T) {
	root := t.TempDir()
	seedPicked(t, root, "t905")
	writeLanded(t, root, oldSchemaCache(time.Now().Unix()))

	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("no such ref")
	}
	defer func() { landedGitRunner = restore }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	got := resolveLandedCounts(root)
	if got.Known() {
		t.Fatalf("old-criterion number became a current measurement: %+v", got)
	}
	if got.Criterion == landedCriterion && got.Landed == 9 {
		t.Fatalf("old number 9 relabelled under the current criterion: %+v", got)
	}
}

// TestRefreshLandedCounts_FailedRefreshOverOldCacheHoldsTheStampedeGuard pins
// the stampede guard's link to the criterion reset: a refresh that fails over
// an old-criterion cache must still leave a current-criterion timestamp behind,
// so renders inside the TTL spawn no further child. Without the reset, the
// cache keeps its old criterion, reads as stale, and every render spawns.
func TestRefreshLandedCounts_FailedRefreshOverOldCacheHoldsTheStampedeGuard(t *testing.T) {
	root := t.TempDir()
	seedPicked(t, root, "t906")
	writeLanded(t, root, oldSchemaCache(time.Now().Unix()))

	restoreRunner := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("no such ref")
	}
	defer func() { landedGitRunner = restoreRunner }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	var spawns atomic.Int64
	restoreProbe := landedSpawnProbe
	landedSpawnProbe = func(string) { spawns.Add(1) }
	defer func() { landedSpawnProbe = restoreProbe }()

	maybeRefreshLandedCounts(root)
	maybeRefreshLandedCounts(root)
	if got := spawns.Load(); got != 0 {
		t.Fatalf("spawn attempts after a failed refresh inside TTL = %d, want 0 (stampede guard)", got)
	}
}

// TestRefreshLandedCounts_AttributedSubjectCounts is AC-SLL-002: a subject
// the kanban predicate attributes, committed after the card was added,
// counts. The equal-second commit counts too (the boundary is `>=`).
func TestRefreshLandedCounts_AttributedSubjectCounts(t *testing.T) {
	for _, ct := range []int64{landedFreshCT, landedFreshCT - 60} {
		root := t.TempDir()
		seedPicked(t, root, "t101")

		restore := landedGitRunner
		landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
			return scanLine("bbbb", ct, "feat(statusline): x (t101)"), nil
		}
		if err := RefreshLandedCounts(context.Background(), root); err != nil {
			t.Fatalf("refresh: %v", err)
		}
		landedGitRunner = restore

		got := resolveLandedCounts(root)
		if !got.Known() || got.Landed != 1 {
			t.Fatalf("ct=%d: attributed subject: %+v, want known landed=1", ct, got)
		}
	}
}

// TestRefreshLandedCounts_GenerationBoundary is AC-SLL-003 / AC-SLL-003b: an
// attributing commit older than the card, or a card whose added_at cannot be
// parsed, does not count.
func TestRefreshLandedCounts_GenerationBoundary(t *testing.T) {
	cases := []struct {
		name    string
		addedAt string
		ct      int64
	}{
		{name: "commit predates the card", addedAt: landedAddedAt, ct: landedStaleCT},
		{name: "empty added_at fails closed", addedAt: "", ct: landedFreshCT},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			seedPickedAt(t, root, tc.addedAt, "t101")

			restore := landedGitRunner
			landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
				return scanLine("cccc", tc.ct, "feat(statusline): x (t101)"), nil
			}
			defer func() { landedGitRunner = restore }()

			if err := RefreshLandedCounts(context.Background(), root); err != nil {
				t.Fatalf("refresh: %v", err)
			}
			got := resolveLandedCounts(root)
			if !got.Known() || got.Landed != 0 {
				t.Fatalf("%s: %+v, want known landed=0", tc.name, got)
			}
		})
	}
}

// TestRefreshLandedCounts_WritesTheCriterion is AC-SLL-004b: the cache on
// disk names the criterion that produced its number.
func TestRefreshLandedCounts_WritesTheCriterion(t *testing.T) {
	root := t.TempDir()
	seedPicked(t, root, "t101")

	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		return scanLine("dddd", landedFreshCT, "feat(statusline): x (t101)"), nil
	}
	defer func() { landedGitRunner = restore }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	raw, err := os.ReadFile(landedCachePath(root))
	if err != nil {
		t.Fatalf("read cache: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode cache: %v", err)
	}
	if landedCriterion == "" || m["criterion"] != landedCriterion {
		t.Fatalf("cache criterion = %v, want %q (raw: %s)", m["criterion"], landedCriterion, raw)
	}
}

// TestLandedScanRunner_RefusesNonGit pins the adapter's one guard: kanban's
// runner contract carries a command name, and the adapter only ever runs git.
func TestLandedScanRunner_RefusesNonGit(t *testing.T) {
	var calls atomic.Int64
	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		calls.Add(1)
		return "", nil
	}
	defer func() { landedGitRunner = restore }()

	run := landedScanRunner(context.Background(), t.TempDir())
	if _, err := run("sh", "-c", "true"); err == nil {
		t.Fatalf("non-git command accepted")
	}
	if _, err := run("git", "log"); err != nil {
		t.Fatalf("git refused: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("runner calls = %d, want 1 (only the git call)", got)
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

	// Simulate the child's first act: the timestamp lands before the work,
	// under the current criterion.
	writeLanded(t, root, mustJSON(t, LandedCounts{FetchedAt: time.Now().Unix(), Criterion: landedCriterion}))
	maybeRefreshLandedCounts(root)
	maybeRefreshLandedCounts(root)
	if got := spawns.Load(); got != 1 {
		t.Fatalf("inside TTL: spawn attempts = %d, want no further attempt", got)
	}

	// Past the TTL the question is asked again.
	writeLanded(t, root, mustJSON(t, LandedCounts{
		Landed: 1, Measured: true, FetchedAt: time.Now().Add(-2 * LandedCountsTTL).Unix(),
		Criterion: landedCriterion,
	}))
	maybeRefreshLandedCounts(root)
	if got := spawns.Load(); got != 2 {
		t.Fatalf("past TTL: spawn attempts = %d, want 2", got)
	}
}

// TestMaybeRefreshLandedCounts_OldCriterionIsStale is AC-SLL-005b: a cache
// written under a different criterion is refreshed at once, whatever its
// timestamp says.
func TestMaybeRefreshLandedCounts_OldCriterionIsStale(t *testing.T) {
	var spawns atomic.Int64
	restore := landedSpawnProbe
	landedSpawnProbe = func(string) { spawns.Add(1) }
	defer func() { landedSpawnProbe = restore }()

	root := t.TempDir()
	writeLanded(t, root, oldSchemaCache(time.Now().Unix()))
	maybeRefreshLandedCounts(root)
	if got := spawns.Load(); got != 1 {
		t.Fatalf("fresh old-criterion cache: spawn attempts = %d, want 1", got)
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
	if !got.Known() || got.Landed != 0 || got.Criterion != landedCriterion {
		t.Fatalf("empty queue: %+v, want an observed zero under the current criterion", got)
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

// landedAddedAt is the added_at seedPicked writes; landedFreshCT is a
// committer time one minute after it, and landedStaleCT one day before it.
const (
	landedAddedAt = "2026-01-02T03:04:05Z"
	landedFreshCT = int64(1767323045 + 60)
	landedStaleCT = int64(1767323045 - 86400)
)

// scanLine renders one row of the kanban subject scan's stream
// (`%H%x00%ct%x00%s`), the shape the refresh child now reads.
func scanLine(sha string, ct int64, subject string) string {
	return sha + "\x00" + strconv.FormatInt(ct, 10) + "\x00" + subject + "\n"
}

// TestRefreshLandedCounts_NonAttributingMentionDoesNotCount is AC-SLL-001.
// The subject NAMES t101, but not in any position kanban's subject
// attribution accepts, so the new criterion counts nothing. The fixture is
// chosen to DISCRIMINATE: the retired body-mention criterion (`\bt101\b`
// over the same text) counts it, which the control below states outright.
func TestRefreshLandedCounts_NonAttributingMentionDoesNotCount(t *testing.T) {
	const subject = "chore: follow-up review notes for t101"

	// Discrimination control: the retired criterion would have counted 1.
	if !regexp.MustCompile(`\bt101\b`).MatchString(subject) {
		t.Fatalf("control broken: the old body-mention criterion must match %q", subject)
	}

	root := t.TempDir()
	seedPicked(t, root, "t101")

	restore := landedGitRunner
	landedGitRunner = func(_ context.Context, _ string, _ ...string) (string, error) {
		return scanLine("aaaa", landedFreshCT, subject), nil
	}
	defer func() { landedGitRunner = restore }()

	if err := RefreshLandedCounts(context.Background(), root); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	got := resolveLandedCounts(root)
	if !got.Measured || got.Landed != 0 {
		t.Fatalf("non-attributing mention: %+v, want measured landed=0 (old criterion would give 1)", got)
	}
}
