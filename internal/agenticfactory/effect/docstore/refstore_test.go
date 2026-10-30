package docstore

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/gitref"
)

// These tests run the cursor and queue documents over the real gitref authority: a local bare repository
// under t.TempDir() holding a genesis commit on a moai-control branch. Every document write is its own
// commit pushed to that bare repository, so a chunked cursor really is several commits. Nothing here
// touches a real remote or the live moai-control ref.

var hermeticGit = []string{
	"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
	"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
	"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid",
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), hermeticGit...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

type refEnv struct {
	t      *testing.T
	remote string
}

func newRefEnv(t *testing.T) *refEnv {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	git(t, root, "init", "--bare", "--initial-branch=main", remote)
	work := filepath.Join(root, "genesis")
	git(t, root, "init", "--initial-branch=main", work)
	if err := os.MkdirAll(filepath.Join(work, "control"), 0o755); err != nil {
		t.Fatal(err)
	}
	schema := []byte(`{"schema":"moai-control","schema_version":1}` + "\n")
	if err := os.WriteFile(filepath.Join(work, "control", "schema.json"), schema, 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, work, "add", "control/schema.json")
	git(t, work, "commit", "-m", "genesis")
	git(t, work, "push", remote, "HEAD:refs/heads/moai-control")
	if !strings.HasPrefix(remote, os.TempDir()) && !strings.Contains(remote, "/T/") && !strings.Contains(remote, "/tmp") {
		t.Fatalf("fixture remote %q is not a temporary path", remote)
	}
	return &refEnv{t: t, remote: remote}
}

// newRef is a fresh clone of the authority: what a restarted process opens.
func (e *refEnv) newRef() *gitref.Ref {
	e.t.Helper()
	dir := filepath.Join(e.t.TempDir(), "clone")
	git(e.t, filepath.Dir(dir), "init", "--initial-branch=main", dir)
	r := gitref.New(dir, e.remote)
	r.Env = hermeticGit
	return r
}

// store is a claim store with the given cache directory, attached to ref.
func (e *refEnv) store(cache string, ref claim.ControlRef) *claim.Store {
	e.t.Helper()
	st, err := claim.OpenWithLimit(cache, 100)
	if err != nil {
		e.t.Fatal(err)
	}
	st.Attach(ref)
	return st
}

// cold is a restart that lost its cache: an empty directory refilled from the authority.
func (e *refEnv) cold(ref claim.ControlRef) *claim.Store {
	e.t.Helper()
	st := e.store(e.t.TempDir(), ref)
	if _, err := st.Rebuild(context.Background()); err != nil {
		e.t.Fatalf("rebuild from the authority: %v", err)
	}
	return st
}

func (e *refEnv) commits() int {
	n, err := strconv.Atoi(git(e.t, e.remote, "rev-list", "--count", "refs/heads/moai-control"))
	if err != nil {
		e.t.Fatal(err)
	}
	return n
}

// faultRef counts the authority's compare-and-swap calls and breaks one: failAt refuses it before it
// happens, crashAt lets it happen and then reports a crash (so the authority is ahead of the cache).
type faultRef struct {
	claim.ControlRef
	cas, failAt, crashAt int
}

func (f *faultRef) arm(failAt, crashAt int) { f.cas, f.failAt, f.crashAt = 0, failAt, crashAt }

func (f *faultRef) CompareAndSwap(ctx context.Context, key string, expect int64, data []byte) (int64, error) {
	f.cas++
	if f.cas == f.failAt {
		return 0, errCrash
	}
	rev, err := f.ControlRef.CompareAndSwap(ctx, key, expect, data)
	if err == nil && f.cas == f.crashAt {
		return 0, errCrash
	}
	return rev, err
}

// smallDocs keeps the crash matrix quick: the same chunking logic with small documents.
func smallDocs(d Docs) Cursors {
	return Cursors{Docs: d, Kind: CollectorKind, MaxDocBytes: 4 << 10, MaxChunks: 16}
}

func TestAChunkedCursorOnTheRealRefRoundTripsWarmAndCold(t *testing.T) {
	e := newRefEnv(t)
	cache := t.TempDir()
	before := e.commits()
	body := bigBody(1, 600<<10) // over the real 256KiB cap: several chunks plus a manifest
	c := chunkedCursors(e.store(cache, e.newRef()))
	if err := c.SaveCursor("o/r/pulls", body); err != nil {
		t.Fatal(err)
	}
	if made := e.commits() - before; made < 4 {
		t.Errorf("the save made %d commits; each document should be its own commit (chunks + manifest)", made)
	}
	warm := chunkedCursors(e.store(cache, e.newRef())) // restart, cache kept
	if got, found, err := warm.LoadCursor("o/r/pulls"); err != nil || !found || !bytes.Equal(got, body) {
		t.Errorf("warm restart: found=%v err=%v equal=%v", found, err, bytes.Equal(got, body))
	}
	cold := chunkedCursors(e.cold(e.newRef())) // restart, cache lost
	if got, found, err := cold.LoadCursor("o/r/pulls"); err != nil || !found || !bytes.Equal(got, body) {
		t.Errorf("cold restart: found=%v err=%v equal=%v", found, err, bytes.Equal(got, body))
	}
}

// refSaves counts the authority writes one save of v2 makes after v1.
func refSaves(t *testing.T, v1, v2 []byte) int {
	t.Helper()
	e := newRefEnv(t)
	fr := &faultRef{ControlRef: e.newRef()}
	c := smallDocs(e.store(t.TempDir(), fr))
	if err := c.SaveCursor("k", v1); err != nil {
		t.Fatal(err)
	}
	fr.arm(0, 0)
	if err := c.SaveCursor("k", v2); err != nil {
		t.Fatal(err)
	}
	return fr.cas
}

// The crash matrix on the real authority. A crash before an authority write, or after it (so the
// authority is ahead of the cache), at every write of a chunked save: a warm restart and a cold restart
// each see a whole cursor — the previous one, or the new one once its manifest landed — never a mixture;
// and the interrupted save is then redone from the stale cache without conflict.
func TestACrashAtEveryAuthorityWriteLeavesAWholeCursorWarmAndCold(t *testing.T) {
	v1, v2 := bigBody(1, 6<<10), bigBody(2, 6<<10)
	total := refSaves(t, v1, v2)
	if total < 4 {
		t.Fatalf("expected several authority writes, got %d", total)
	}
	for _, mode := range []string{"refused-before-write", "crash-after-write"} {
		for k := 1; k <= total; k++ {
			e := newRefEnv(t)
			cache := t.TempDir()
			fr := &faultRef{ControlRef: e.newRef()}
			c := smallDocs(e.store(cache, fr))
			if err := c.SaveCursor("k", v1); err != nil {
				t.Fatal(err)
			}
			if mode == "refused-before-write" {
				fr.arm(k, 0)
			} else {
				fr.arm(0, k)
			}
			if err := c.SaveCursor("k", v2); !errors.Is(err, core.ErrStoreUnavailable) {
				t.Fatalf("%s at write %d/%d: err = %v, want the authority failure", mode, k, total, err)
			}
			authorityHasV2 := mode == "crash-after-write" && k == total // the manifest is the last write
			label := mode + " " + strconv.Itoa(k) + "/" + strconv.Itoa(total)

			warm := smallDocs(e.store(cache, e.newRef())) // same cache, new process
			if got, found, err := warm.LoadCursor("k"); err != nil || !found || !bytes.Equal(got, v1) {
				t.Errorf("%s warm: found=%v err=%v isV1=%v (a cache behind the authority still holds the previous whole cursor)", label, found, err, bytes.Equal(got, v1))
			}
			coldC := smallDocs(e.cold(e.newRef()))
			want := v1
			if authorityHasV2 {
				want = v2
			}
			if got, found, err := coldC.LoadCursor("k"); err != nil || !found || !bytes.Equal(got, want) {
				t.Errorf("%s cold: found=%v err=%v matches-expected=%v", label, found, err, bytes.Equal(got, want))
			}
			// Redo from the stale warm cache: the write must repair the cache, not conflict forever.
			if err := warm.SaveCursor("k", v2); err != nil {
				t.Errorf("%s: redo from the stale cache failed: %v", label, err)
				continue
			}
			if got, _, err := smallDocs(e.cold(e.newRef())).LoadCursor("k"); err != nil || !bytes.Equal(got, v2) {
				t.Errorf("%s: after the redo the authority does not hold the new cursor (err=%v)", label, err)
			}
		}
	}
}

// A reader whose listing of the authority is older than its reads — documents are separate commits, so the
// listing can predate chunks the manifest it then reads already names — never adopts the incomplete newest
// generation: it falls back to the previous complete one.
func TestAReaderWithAnOlderListingFallsBackToThePreviousWholeCursor(t *testing.T) {
	v1, v2 := bigBody(1, 6<<10), bigBody(2, 6<<10)
	nChunks := refSaves(t, v1, v2) - 1 // the last authority write is the manifest
	for k := 0; k <= nChunks; k++ {    // k chunks of the new generation are in the listing
		e := newRefEnv(t)
		fr := &faultRef{ControlRef: e.newRef()}
		c := smallDocs(e.store(t.TempDir(), fr))
		if err := c.SaveCursor("k", v1); err != nil {
			t.Fatal(err)
		}
		if k > 0 {
			fr.arm(0, k) // the writer dies after its k-th chunk
			if err := c.SaveCursor("k", v2); !errors.Is(err, core.ErrStoreUnavailable) {
				t.Fatalf("k=%d: %v", k, err)
			}
		}
		listing, err := fr.Keys(context.Background(), "") // the reader lists now ...
		if err != nil {
			t.Fatal(err)
		}
		fr.arm(0, 0)
		if err := c.SaveCursor("k", v2); err != nil { // ... and the writer then finishes
			t.Fatalf("k=%d: finishing the save: %v", k, err)
		}
		reader := e.store(t.TempDir(), listedRef{ControlRef: e.newRef(), keys: listing}) // reads are live
		if _, err := reader.Rebuild(context.Background()); err != nil {
			t.Fatal(err)
		}
		got, found, err := smallDocs(reader).LoadCursor("k")
		want, name := v1, "previous"
		if k == nChunks { // every chunk the manifest names was in the listing
			want, name = v2, "new"
		}
		if err != nil || !found || !bytes.Equal(got, want) {
			t.Errorf("listing with %d/%d new chunks: found=%v err=%v: want the whole %s cursor, isV1=%v isV2=%v",
				k, nChunks, found, err, name, bytes.Equal(got, v1), bytes.Equal(got, v2))
		}
	}
}

// listedRef serves reads live but a listing fixed earlier.
type listedRef struct {
	claim.ControlRef
	keys []string
}

func (l listedRef) Keys(_ context.Context, prefix string) ([]string, error) {
	var out []string
	for _, k := range l.keys {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out, nil
}

// Queue on the real authority: the write lands but the process dies before the acknowledgement, so the
// cache is behind. The host redelivers; the redelivery repairs the cache, is acknowledged once and queues
// nothing twice. A cold restart sees the same single event.
func TestTheQueueOnTheRealRefAbsorbsACrashAfterTheAuthorityWrite(t *testing.T) {
	e := newRefEnv(t)
	cache := t.TempDir()
	fr := &faultRef{ControlRef: e.newRef()}
	q := newQueue(e.store(cache, fr))
	fr.arm(0, 1)
	acked := 0
	if err := q.Accept("d1", hook, func() { acked++ }); err == nil || acked != 0 {
		t.Fatalf("err=%v acked=%d, want the crash and no acknowledgement", err, acked)
	}
	if ids := pendingIDs(t, newQueue(e.cold(e.newRef()))); len(ids) != 1 {
		t.Fatalf("the authority write was lost: %v", ids)
	}
	warm := newQueue(e.store(cache, e.newRef())) // restart on the stale cache
	if ids := pendingIDs(t, warm); len(ids) != 0 {
		t.Fatalf("stale cache unexpectedly already has %v", ids)
	}
	if err := warm.Accept("d1", hook, func() { acked++ }); err != nil || acked != 1 {
		t.Fatalf("redelivery on a stale cache: err=%v acked=%d", err, acked)
	}
	if ids := pendingIDs(t, warm); len(ids) != 1 {
		t.Errorf("after the redelivery: %v", ids)
	}
	if ids := pendingIDs(t, newQueue(e.cold(e.newRef()))); len(ids) != 1 {
		t.Errorf("authority after the redelivery: %v", ids)
	}
}

// End to end on the real authority: a big snapshot is acknowledged, the cache is lost, and the restarted
// collector reports only the real change and no baseline.
func TestACollectorResumesFromTheRealRefAfterLosingItsCache(t *testing.T) {
	e := newRefEnv(t)
	now := time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)
	body1 := string(bigBody(7, 300<<10))
	body2 := strings.TrimSuffix(body1, "]") + `,{"number":2,"pad":"x"}]`
	mk := func(st *claim.Store, body string) *forge.Collector {
		return &forge.Collector{
			Client: &forge.Client{T: &oneShot{body: body}, B: &forge.Budget{}, Now: func() time.Time { return now },
				Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}},
			Path: "/p", Interval: forge.Adaptive{Min: time.Second, Max: time.Minute, Safety: time.Hour},
			Differ: forge.DiffByKey("number"), Cursors: chunkedCursors(st), CursorKey: "o/r/big",
		}
	}
	first := mk(e.store(t.TempDir(), e.newRef()), body1)
	rep := first.Tick(t.Context(), now)
	if err := first.Ack(rep); err != nil {
		t.Fatalf("ack on the real ref: %v", err)
	}
	r2 := mk(e.cold(e.newRef()), body2).Tick(t.Context(), now)
	if r2.Baseline || len(r2.Changes) != 1 || r2.Changes[0].Key != "2" || r2.Changes[0].Kind != forge.ChangeAdded {
		t.Errorf("after losing the cache: %+v, want only issue 2 added and no baseline", r2)
	}
}
