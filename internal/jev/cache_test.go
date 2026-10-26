package jev

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// cache_test.go — SPEC-FETCH-WEB-001 M5 (AC-FW-009..011, REQ-FW-011..013).
// Disk cache (body-SHA + resolved-model-id key, atomic replace, restart
// hit, keep-and-warn stale serve), Retry-After backoff composing with
// REQ-JEVC-008, the concurrency semaphore, and new-vs-cache usage
// separation. Every test runs against a temp cache dir and a fake Doer or
// loopback server — never the real endpoint (REQ-FW-010).

// cachedDoer answers from a scripted list of responses, in order; calls
// counts every Do invocation. A script miss repeats the last response.
type cachedDoer struct {
	mu     sync.Mutex
	script []doerStep
	calls  int64
}

type doerStep struct {
	status  int
	body    string
	headers map[string]string
	delay   time.Duration
}

func (d *cachedDoer) Do(r *http.Request) (*http.Response, error) {
	atomic.AddInt64(&d.calls, 1)
	d.mu.Lock()
	var step doerStep
	if len(d.script) > 0 {
		step = d.script[0]
		if len(d.script) > 1 {
			d.script = d.script[1:]
		}
	}
	d.mu.Unlock()
	if step.delay > 0 {
		time.Sleep(step.delay)
	}
	resp := &http.Response{
		StatusCode: step.status,
		Body:       http.NoBody,
		Header:     http.Header{},
		Request:    r,
	}
	if step.body != "" {
		resp.Body = io.NopCloser(strings.NewReader(step.body))
	}
	for k, v := range step.headers {
		resp.Header.Set(k, v)
	}
	return resp, nil
}

// okBody builds a 200 wire response answered by model.
func cacheOKBody(model string) string {
	return fmt.Sprintf(`{"model":%q,"answers":{"q1":{"type":"noul","noul":0.9}},"usage":{"input_tokens":7}}`, model)
}

// newCachedClient builds an enabled client with the credential loader set
// and the cache rooted at a temp dir.
func newCachedClient(t *testing.T, doer Doer) *Client {
	t.Helper()
	c := New(true)
	c.LoadCredential = func() string { return "test-key" }
	c.HTTP = doer
	c.CacheDir = t.TempDir()
	return c
}

func cacheQuestion() Request {
	return Request{
		State: "the state under judgment",
		Questions: []Question{
			{ID: "q1", Text: "Is the premise alive?", Kind: KindNoul},
		},
		Repeatable: true, // a pure read: repetition is provably side-effect-free
	}
}

// AC-FW-009: a cache hit is served without a network request, and a fresh
// process (restart) resolves the same id from the persisted alias map and
// serves offline.
func TestCache_HitServedWithoutNetwork_RestartOffline(t *testing.T) {
	doer := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.13.0")}}}
	c := newCachedClient(t, doer)
	first := c.Ask(context.Background(), cacheQuestion())
	if !first.OK() {
		t.Fatalf("prime call failed: %s", first.NoticeLine())
	}

	// Same client: hit without network.
	before := atomic.LoadInt64(&doer.calls)
	second := c.Ask(context.Background(), cacheQuestion())
	if !second.OK() {
		t.Fatalf("cached call failed: %s", second.NoticeLine())
	}
	if got := atomic.LoadInt64(&doer.calls); got != before {
		t.Errorf("cache hit issued a network call (%d -> %d)", before, got)
	}

	// Fresh process: same cache dir, network refused.
	restart := New(true)
	restart.LoadCredential = c.LoadCredential
	restart.CacheDir = c.CacheDir
	restart.HTTP = &refusingDoer{}
	third := restart.Ask(context.Background(), cacheQuestion())
	if !third.OK() {
		t.Fatalf("restart offline hit failed: %s", third.NoticeLine())
	}
}

type refusingDoer struct{ calls int64 }

func (r *refusingDoer) Do(*http.Request) (*http.Response, error) {
	atomic.AddInt64(&r.calls, 1)
	return nil, fmt.Errorf("network is closed for this test")
}

// AC-FW-009: the model id is in the key — two identical bodies under two
// model ids produce two distinct cache files.
func TestCache_ModelIdInKey_TwoFiles(t *testing.T) {
	doer := &cachedDoer{script: []doerStep{
		{status: 200, body: cacheOKBody("jev-1.13.0")},
		{status: 200, body: cacheOKBody("jev-1.14.0")},
	}}
	a := newCachedClient(t, doer)
	b := newCachedClient(t, doer)
	a.CacheDir = a.CacheDir // separate clients, SAME dir
	b.CacheDir = a.CacheDir
	a.ModelOverride = "jev-1.13.0"
	b.ModelOverride = "jev-1.14.0"
	if !a.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("override call A failed")
	}
	if !b.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("override call B failed")
	}
	entries := cacheEntries(t, a.CacheDir)
	if len(entries) < 2 {
		t.Fatalf("want >= 2 cache files (one per model id), got %d", len(entries))
	}
	models := map[string]bool{}
	for _, e := range entries {
		models[e.Model] = true
	}
	if !models["jev-1.13.0"] || !models["jev-1.14.0"] {
		t.Errorf("distinct model keys expected, saw %v", models)
	}
}

// AC-FW-009's N5 staging: an explicit model override that misses against an
// alias-moved cache state goes to the network and writes its OWN entry —
// never another model's stale entry.
func TestCache_ExplicitOverrideMiss_GoesToNetworkOwnEntry(t *testing.T) {
	// body1 primed under the alias-resolved model jev-1.13.0.
	doer := &cachedDoer{script: []doerStep{
		{status: 200, body: cacheOKBody("jev-1.13.0")},
	}}
	c := newCachedClient(t, doer)
	if !c.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("prime failed")
	}

	// The alias moves to jev-1.14.0 via a DIFFERENT body.
	mover := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.14.0")}}}
	mc := New(true)
	mc.LoadCredential = c.LoadCredential
	mc.CacheDir = c.CacheDir
	mc.HTTP = mover
	moved := Request{State: "a different state", Questions: cacheQuestion().Questions}
	if !mc.Ask(context.Background(), moved).OK() {
		t.Fatal("alias-move call failed")
	}

	// Explicit override jev-1.15.0 over body1: miss under 1.15.0, and the
	// keep-and-warn probe MUST NOT apply (override lookups never serve
	// another model's entry) — this goes to the network and writes its own
	// entry.
	override := New(true)
	override.LoadCredential = c.LoadCredential
	override.CacheDir = c.CacheDir
	override.ModelOverride = "jev-1.15.0"
	net := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.15.0")}}}
	override.HTTP = net
	res := override.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("override call failed: %s", res.NoticeLine())
	}
	if got := atomic.LoadInt64(&net.calls); got != 1 {
		t.Errorf("override miss must hit the network exactly once, got %d calls", got)
	}
	if res.StaleCacheWarning != "" {
		t.Errorf("explicit override must never serve a stale cross-model entry, got warning %q", res.StaleCacheWarning)
	}
	// Two entries for body1 now: jev-1.13.0 (primed) and jev-1.15.0 (own).
	body1Count := 0
	for _, e := range cacheEntries(t, c.CacheDir) {
		if e.Model == "jev-1.13.0" || e.Model == "jev-1.15.0" {
			body1Count++
		}
	}
	if body1Count != 2 {
		t.Errorf("want the explicit model's OWN fresh entry beside the alias-resolved one, counted %d", body1Count)
	}
}

// AC-FW-011's alias-moved case: a fresh process re-queries the same body
// after the alias moved; the miss probes the prior-model entry, serves it
// with a staleness warning naming the cached model id, and KEEPS the entry.
func TestCache_StaleServeKeepAndWarn(t *testing.T) {
	doer := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.13.0")}}}
	c := newCachedClient(t, doer)
	if !c.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("prime failed")
	}
	// Alias moves to jev-1.14.0 via a different body.
	mover := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.14.0")}}}
	mc := New(true)
	mc.LoadCredential = c.LoadCredential
	mc.CacheDir = c.CacheDir
	mc.HTTP = mover
	moved := Request{State: "a different state", Questions: cacheQuestion().Questions}
	if !mc.Ask(context.Background(), moved).OK() {
		t.Fatal("alias-move call failed")
	}

	// Fresh process, network closed, same body: probe the prior-model entry.
	fresh := New(true)
	fresh.LoadCredential = c.LoadCredential
	fresh.CacheDir = c.CacheDir
	refuse := &refusingDoer{}
	fresh.HTTP = refuse
	res := fresh.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("stale serve failed: %s", res.NoticeLine())
	}
	if !strings.Contains(res.StaleCacheWarning, "jev-1.13.0") {
		t.Errorf("staleness warning must name the cached model id, got %q", res.StaleCacheWarning)
	}
	if res.NoticeLine() == "" || !strings.Contains(res.NoticeLine(), "jev-1.13.0") {
		t.Errorf("the consumer notice line must surface the staleness warning: %q", res.NoticeLine())
	}
	kept := false
	for _, e := range cacheEntries(t, c.CacheDir) {
		if e.Model == "jev-1.13.0" {
			kept = true
		}
	}
	if !kept {
		t.Error("the stale entry must be KEPT, not deleted (keep-and-warn)")
	}
	if refuse.calls != 0 {
		t.Errorf("stale probe must serve without network, saw %d calls", refuse.calls)
	}
}

// AC-FW-011: usage aggregation separates new network requests from
// cache-served answers.
func TestUsage_FromCacheSeparation(t *testing.T) {
	doer := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.13.0")}}}
	c := newCachedClient(t, doer)
	if !c.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("network call failed")
	}
	if !c.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("cached call failed")
	}
	log := c.UsageLog()
	if len(log) != 2 {
		t.Fatalf("usage log holds %d records, want 2", len(log))
	}
	if log[0].FromCache {
		t.Error("first record is a network call; FromCache must be false")
	}
	if !log[1].FromCache {
		t.Error("second record is cache-served; FromCache must be true")
	}
}

// Cache writes are atomic: no temp residue, every file parses.
func TestCache_AtomicReplaceNoResidue(t *testing.T) {
	doer := &cachedDoer{}
	c := newCachedClient(t, doer)
	for i := 0; i < 5; i++ {
		c.ModelOverride = fmt.Sprintf("jev-1.1%d.0", i)
		doer.mu.Lock()
		doer.script = []doerStep{{status: 200, body: cacheOKBody(fmt.Sprintf("jev-1.1%d.0", i))}}
		doer.mu.Unlock()
		if !c.Ask(context.Background(), cacheQuestion()).OK() {
			t.Fatalf("write %d failed", i)
		}
	}
	entries, err := os.ReadDir(c.CacheDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp residue after writes: %s", e.Name())
		}
	}
	if got := len(cacheEntries(t, c.CacheDir)); got != 5 {
		t.Errorf("5 writes produced %d parseable entries", got)
	}
}

// AC-FW-010 retry limb: a repeatable request facing 429 + Retry-After waits
// the parsed delay and retries; the retry count is bounded.
func TestRetryAfter_WaitsParsedDelayAndRetries(t *testing.T) {
	doer := &cachedDoer{script: []doerStep{
		{status: 429, headers: map[string]string{"Retry-After": "1"}},
		{status: 200, body: cacheOKBody("jev-1.13.0")},
	}}
	c := newCachedClient(t, doer)
	start := time.Now()
	res := c.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("retry after 429 failed: %s", res.NoticeLine())
	}
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Errorf("client retried after only %v; the parsed Retry-After: 1 delay was not honored", elapsed)
	}
	if got := atomic.LoadInt64(&doer.calls); got != 2 {
		t.Errorf("want exactly 2 server calls, got %d", got)
	}
}

// The parsed delay is bounded by a maximum.
func TestRetryAfter_DelayBounded(t *testing.T) {
	if d, ok := parseRetryAfter("3600"); !ok || d > maxRetryAfterDelay {
		t.Errorf("parseRetryAfter(3600) = %v (ok=%v); must cap at %v", d, ok, maxRetryAfterDelay)
	}
	if d, ok := parseRetryAfter("2"); !ok || d != 2*time.Second {
		t.Errorf("parseRetryAfter(2) = %v (ok=%v), want 2s", d, ok)
	}
	if _, ok := parseRetryAfter("not-a-date"); ok {
		t.Error("unparseable Retry-After must not parse")
	}
}

// The retry count stays bounded: an always-429 server ends the call after
// the attempt budget, not forever.
func TestRetryAfter_BoundedRetries(t *testing.T) {
	doer := &cachedDoer{script: []doerStep{
		{status: 429, headers: map[string]string{"Retry-After": "0"}},
	}}
	c := newCachedClient(t, doer) // MaxRetries 1 -> 2 attempts
	res := c.Ask(context.Background(), cacheQuestion())
	if res.OK() {
		t.Fatal("always-429 must not produce an answer")
	}
	if res.Availability != RateLimited {
		t.Errorf("availability = %s, want rate-limited", res.Availability)
	}
	if got := atomic.LoadInt64(&doer.calls); got != 2 {
		t.Errorf("attempt budget is 2, server saw %d calls", got)
	}
}

// Negative case (REQ-FW-012 composing with REQ-JEVC-008): a request whose
// repetition is not declared side-effect-free is NOT retried on 429.
func TestRetryAfter_NonRepeatableNotRetried(t *testing.T) {
	doer := &cachedDoer{script: []doerStep{
		{status: 429, headers: map[string]string{"Retry-After": "0"}},
	}}
	c := newCachedClient(t, doer)
	req := cacheQuestion()
	req.Repeatable = false
	res := c.Ask(context.Background(), req)
	if res.OK() {
		t.Fatal("429 must not produce an answer")
	}
	if got := atomic.LoadInt64(&doer.calls); got != 1 {
		t.Errorf("non-repeatable request was retried: %d calls", got)
	}
}

// Negative case: a cache-served answer issues no retry at all.
func TestRetryAfter_CacheHitIssuesNoRetry(t *testing.T) {
	prime := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.13.0")}}}
	c := newCachedClient(t, prime)
	if !c.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("prime failed")
	}
	hazard := &cachedDoer{script: []doerStep{
		{status: 429, headers: map[string]string{"Retry-After": "0"}},
	}}
	c.HTTP = hazard
	res := c.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("cache hit must serve without contacting the retry path: %s", res.NoticeLine())
	}
	if got := atomic.LoadInt64(&hazard.calls); got != 0 {
		t.Errorf("cache hit issued %d network calls", got)
	}
}

// AC-FW-010 concurrency limb: a semaphore of width W bounds in-flight
// requests, observed server-side.
func TestSemaphore_BoundsInFlight(t *testing.T) {
	const width = 3
	var live, maxLive int64
	release := make(chan struct{})
	doer := DoerFunc(func(r *http.Request) (*http.Response, error) {
		cur := atomic.AddInt64(&live, 1)
		for {
			old := atomic.LoadInt64(&maxLive)
			if cur <= old || atomic.CompareAndSwapInt64(&maxLive, old, cur) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		atomic.AddInt64(&live, -1)
		<-release
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(cacheOKBody("jev-1.13.0"))), Header: http.Header{}, Request: r}, nil
	})
	c := newCachedClient(t, doer)
	c.MaxConcurrency = width
	c.CacheDir = "" // no cache: every call is network
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !c.Ask(context.Background(), cacheQuestion()).OK() {
				t.Error("concurrent call failed")
			}
		}()
	}
	// Let all requests arrive, then release them.
	time.Sleep(150 * time.Millisecond)
	close(release)
	wg.Wait()
	if got := atomic.LoadInt64(&maxLive); got > width {
		t.Errorf("observed %d concurrent in-flight requests, semaphore width is %d", got, width)
	}
	if got := atomic.LoadInt64(&maxLive); got < 2 {
		t.Errorf("observed max in-flight %d — the probe never exercised concurrency", got)
	}
}

// DefaultCacheDir points under the home directory, never into a project.
func TestDefaultCacheDir_UnderHome(t *testing.T) {
	dir, err := DefaultCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(dir), ".moai/cache/jev") {
		t.Errorf("DefaultCacheDir = %q, want under ~/.moai/cache/jev", dir)
	}
}

// helpers

// cacheEntries parses every cache file in dir.
func cacheEntries(t *testing.T, dir string) []cacheEntry {
	t.Helper()
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []cacheEntry
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var e cacheEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatalf("entry %s does not parse: %v", f.Name(), err)
		}
		out = append(out, e)
	}
	return out
}

// DoerFunc adapts a function to the Doer seam.
type DoerFunc func(*http.Request) (*http.Response, error)

func (f DoerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
