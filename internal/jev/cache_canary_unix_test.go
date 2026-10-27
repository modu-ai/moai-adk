//go:build unix

package jev

// cache_canary_unix_test.go — leader codex-audit round 3, finding 2: the
// stale probe's miss path must not scan the cache directory. The canary is a
// FIFO named like a cache file: reading it blocks forever (no writer ever
// opens it), so any per-file read loop over the directory hangs on it. The
// probe reads only the KNOWN prior-model cacheKey(sha, model) files; with an
// empty prior history it reads nothing at all and the miss returns
// immediately. (Unix-only: the canary is a Mkfifo. The build for other
// platforms compiles the package without this file; the behavior it pins is
// platform-independent.)

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// mkCanaryFIFO plants the blocking canary in the cache dir.
func mkCanaryFIFO(t *testing.T, dir string) {
	t.Helper()
	if err := syscall.Mkfifo(filepath.Join(dir, "zz-canary.json"), 0o600); err != nil {
		t.Skipf("cannot create fifo canary: %v", err)
	}
}

// probeWithinDeadline runs fn and fails the test if it does not return
// before the deadline — the observable signature of the canary's blocking
// read.
func probeWithinDeadline(t *testing.T, name string, fn func() *cacheEntry) *cacheEntry {
	t.Helper()
	done := make(chan *cacheEntry, 1)
	go func() { done <- fn() }()
	select {
	case e := <-done:
		return e
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: the probe read a non-probed cache file (blocked on the fifo canary) — the miss path must not scan the directory", name)
		return nil
	}
}

// Empty prior history: the probe admits nothing AND reads no cache file —
// the miss is immediate, with zero directory reads (the round-3 finding:
// the old directory scan read and JSON-parsed every cache file even here,
// its cost growing with the cache size).
func TestPriorModelEntry_EmptyHistoryReadsNoCacheFile(t *testing.T) {
	dir := t.TempDir()
	if err := writeModelsMap(dir, modelsMap{Models: map[string]string{ModelID: "jev-1.14.0"}}); err != nil {
		t.Fatal(err)
	}
	mkCanaryFIFO(t, dir)
	c := &Client{CacheDir: dir}
	if e := probeWithinDeadline(t, "empty history", func() *cacheEntry {
		e, _ := c.priorModelEntry("bodiesha", "jev-1.14.0")
		return e
	}); e != nil {
		t.Errorf("empty prior history must admit nothing, served model %q", e.Model)
	}
}

// A recorded prior model: the probe serves its entry via the KNOWN
// cacheKey(sha, model) file directly — without reading anything else in the
// directory — and when that key is absent the probe misses without touching
// the directory either.
func TestPriorModelEntry_ProbesOnlyKnownPriorModelKeys(t *testing.T) {
	dir := t.TempDir()
	if err := writeModelsMap(dir, modelsMap{
		Models: map[string]string{ModelID: "jev-1.14.0"},
		Prior:  map[string][]string{ModelID: {"jev-1.13.0"}},
	}); err != nil {
		t.Fatal(err)
	}
	const sha = "sha256-of-body"
	entry := cacheEntry{BodySHA: sha, Model: "jev-1.13.0", CreatedAt: time.Now().UTC()}
	raw, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(dir, cacheKey(sha, "jev-1.13.0")+".json")
	if err := os.WriteFile(keyPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	mkCanaryFIFO(t, dir)
	c := &Client{CacheDir: dir}

	if e := probeWithinDeadline(t, "known-key serve", func() *cacheEntry {
		e, _ := c.priorModelEntry(sha, "jev-1.14.0")
		return e
	}); e == nil || e.Model != "jev-1.13.0" {
		t.Fatalf("probe must serve the prior-model entry via its known key, got %+v", e)
	}

	// Key absent (a different body under the same alias history): the probe
	// misses without reading anything else.
	if err := os.Remove(keyPath); err != nil {
		t.Fatal(err)
	}
	if e := probeWithinDeadline(t, "known-key miss", func() *cacheEntry {
		e, _ := c.priorModelEntry("sha256-of-other-body", "jev-1.14.0")
		return e
	}); e != nil {
		t.Errorf("absent prior-model key must miss, served model %q", e.Model)
	}
}
