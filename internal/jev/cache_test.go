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
	b := New(true)
	b.LoadCredential = a.LoadCredential
	b.HTTP = doer
	b.CacheDir = a.CacheDir // separate clients, SAME dir
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

// N2 (card t1968 review round 2): with an explicit model override pinned,
// the wire still carries the request alias, so the RESPONSE names the model
// that actually answered. When the two differ, the entry must be keyed and
// attributed by the RESPONSE's id — never the override's — and a later
// override lookup must not serve that wrong-model answer as a clean hit.
func TestCache_OverrideMismatchStoresUnderResponseModel(t *testing.T) {
	// Override pins jev-1.13.0; the endpoint answers jev-1.14.0.
	doer := &cachedDoer{script: []doerStep{
		{status: 200, body: cacheOKBody("jev-1.14.0")},
		{status: 200, body: cacheOKBody("jev-1.14.0")},
	}}
	c := newCachedClient(t, doer)
	c.ModelOverride = "jev-1.13.0"
	first := c.Ask(context.Background(), cacheQuestion())
	if !first.OK() {
		t.Fatalf("override call failed: %s", first.NoticeLine())
	}
	for _, e := range cacheEntries(t, c.CacheDir) {
		if e.Model != "jev-1.14.0" {
			t.Errorf("entry keyed under %q; the response id jev-1.14.0 is the only authoritative key", e.Model)
		}
	}
	// The second override call must NOT be a clean hit on the mismatched
	// entry: the override probes its own id, misses (never another model's
	// entry), and goes to the network.
	second := c.Ask(context.Background(), cacheQuestion())
	if !second.OK() {
		t.Fatalf("second override call failed: %s", second.NoticeLine())
	}
	if got := atomic.LoadInt64(&doer.calls); got != 2 {
		t.Errorf("a mismatched override answer was served as a clean cache hit; network calls = %d, want 2", got)
	}
	if second.StaleCacheWarning != "" {
		t.Errorf("override lookups never serve stale cross-model entries, got warning %q", second.StaleCacheWarning)
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

// C1 (t1968 sync-audit iter6): the stale probe's candidate set is the
// alias's own move history — the model ids the alias→model map previously
// recorded. An explicit-override-authored entry is NOT one of them: a model
// the alias never resolved to must not be served as a later alias lookup's
// stale pick. With no prior alias model on record the probe admits nothing
// and the miss goes to the network.
func TestCache_StaleProbeIgnoresOverrideAuthoredEntries(t *testing.T) {
	// The alias's first recorded model is jev-1.14.0, primed over body2.
	prime := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.14.0")}}}
	c := newCachedClient(t, prime)
	other := Request{State: "a different state", Questions: cacheQuestion().Questions}
	if !c.Ask(context.Background(), other).OK() {
		t.Fatal("prime failed")
	}

	// An explicit override answers body1 under jev-1.15.0. Override calls
	// never update the alias map, so jev-1.15.0 is absent from the alias's
	// history — whatever entries they write.
	over := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.15.0")}}}
	oc := New(true)
	oc.LoadCredential = c.LoadCredential
	oc.CacheDir = c.CacheDir
	oc.ModelOverride = "jev-1.15.0"
	oc.HTTP = over
	if !oc.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("override call failed")
	}

	// A fresh process's alias lookup of body1 resolves jev-1.14.0, misses,
	// and must NOT serve the override-authored entry: no prior alias model
	// is recorded, so the probe admits nothing and the call goes to the
	// network for a fresh answer.
	fresh := New(true)
	fresh.LoadCredential = c.LoadCredential
	fresh.CacheDir = c.CacheDir
	net := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.14.0")}}}
	fresh.HTTP = net
	res := fresh.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("fresh call failed: %s", res.NoticeLine())
	}
	if got := atomic.LoadInt64(&net.calls); got != 1 {
		t.Errorf("stale probe served an override-authored entry (a model the alias never recorded); want a network call, saw %d", got)
	}
	if res.StaleCacheWarning != "" {
		t.Errorf("no stale serve is permitted against the alias's empty history, got warning %q", res.StaleCacheWarning)
	}
}

// C1 control: when the alias DID previously record a model, its entry wins
// the stale probe even when a LATER override-authored entry for the same
// body exists — the pick is the alias's prior model, not the newest file.
func TestCache_StaleProbePicksAliasPriorModelNotNewestForeignEntry(t *testing.T) {
	// body1 primed under the alias's first model jev-1.13.0.
	doer := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.13.0")}}}
	c := newCachedClient(t, doer)
	if !c.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("prime failed")
	}
	// A later explicit override writes an entry for the same body under
	// jev-1.15.0 (the newest CreatedAt of the two).
	over := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.15.0")}}}
	oc := New(true)
	oc.LoadCredential = c.LoadCredential
	oc.CacheDir = c.CacheDir
	oc.ModelOverride = "jev-1.15.0"
	oc.HTTP = over
	if !oc.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("override call failed")
	}
	// The alias moves to jev-1.14.0 over a different body: jev-1.13.0 joins
	// the alias's recorded history.
	mover := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.14.0")}}}
	mc := New(true)
	mc.LoadCredential = c.LoadCredential
	mc.CacheDir = c.CacheDir
	mc.HTTP = mover
	moved := Request{State: "a different state", Questions: cacheQuestion().Questions}
	if !mc.Ask(context.Background(), moved).OK() {
		t.Fatal("alias-move call failed")
	}

	// Fresh process, network closed, same body1: the probe must serve the
	// alias's own prior model jev-1.13.0 — never the newer override-authored
	// jev-1.15.0 entry.
	fresh := New(true)
	fresh.LoadCredential = c.LoadCredential
	fresh.CacheDir = c.CacheDir
	fresh.HTTP = &refusingDoer{}
	res := fresh.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("stale serve failed: %s", res.NoticeLine())
	}
	if !strings.Contains(res.StaleCacheWarning, "jev-1.13.0") {
		t.Errorf("stale pick must be the alias's prior model jev-1.13.0, got %q", res.StaleCacheWarning)
	}
	if strings.Contains(res.StaleCacheWarning, "jev-1.15.0") {
		t.Errorf("an override-authored entry leaked into the alias's stale pick: %q", res.StaleCacheWarning)
	}
}

// N1 (t1968 sync-audit iter6): an OK response that answers only PART of the
// request is a transient endpoint fault, not a verdict. Caching it would
// persist the fault — every later identical request would hit the incomplete
// entry with from_cache=true. Only complete answers are stored.
func TestCache_IncompleteResponseNotCached(t *testing.T) {
	twoQ := Request{
		State: "the state under judgment",
		Questions: []Question{
			{ID: "q1", Text: "Is the premise alive?", Kind: KindNoul},
			{ID: "q2", Text: "Is the backup plan real?", Kind: KindNoul},
		},
		Repeatable: true,
	}
	doer := &cachedDoer{script: []doerStep{
		// Answers q1 only, then the complete pair.
		{status: 200, body: `{"model":"jev-1.13.0","answers":{"q1":{"type":"noul","noul":0.9}},"usage":{"input_tokens":7}}`},
		{status: 200, body: `{"model":"jev-1.13.0","answers":{"q1":{"type":"noul","noul":0.9},"q2":{"type":"noul","noul":0.4}},"usage":{"input_tokens":7}}`},
	}}
	c := newCachedClient(t, doer)
	first := c.Ask(context.Background(), twoQ)
	if !first.OK() {
		t.Fatalf("partial-but-OK call failed: %s", first.NoticeLine())
	}
	if len(first.Answers) != 1 {
		t.Fatalf("partial response carries 1 answer, got %d", len(first.Answers))
	}
	// The second identical call must NOT be served the incomplete entry: it
	// goes back to the network for the full answer.
	second := c.Ask(context.Background(), twoQ)
	if !second.OK() {
		t.Fatalf("second call failed: %s", second.NoticeLine())
	}
	if got := atomic.LoadInt64(&doer.calls); got != 2 {
		t.Errorf("the incomplete response was cached and re-served; network calls = %d, want 2", got)
	}
	if second.Usage.FromCache {
		t.Error("the second call must not read as a cache serve")
	}
}

// N1 serve leg: an incomplete entry already on disk (the shape a pre-repair
// client could have written) is not admitted as a hit either — the lookup
// misses and the call goes to the network.
func TestCache_IncompleteEntryNotServedAsHit(t *testing.T) {
	net := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.13.0")}}}
	c := newCachedClient(t, net)
	c.ModelOverride = "jev-1.13.0"
	body, err := json.Marshal(toWire(cacheQuestion()))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.cacheStore(cacheEntry{
		BodySHA:   bodySHA(body),
		Model:     "jev-1.13.0",
		Answers:   []Answer{{QuestionID: "unrelated", Kind: KindNoul, Probability: 0.5}},
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	res := c.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("call failed: %s", res.NoticeLine())
	}
	if got := atomic.LoadInt64(&net.calls); got != 1 {
		t.Errorf("an incomplete entry was served as a hit; network calls = %d, want 1", got)
	}
	if res.Usage.FromCache {
		t.Error("the serve must not read from_cache=true")
	}
}

// N2 (t1968 sync-audit iter6): the change notice (answered model differs
// from VerifiedInterpretedModel) applied only to the network path — a cache
// serve returned no notice, so consumers learned of the model change only
// when the network produced it. The comparison now runs on cache-served
// results too, against the ENTRY's model id.
func TestCache_HitCarriesChangeNotice(t *testing.T) {
	prime := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.14.0")}}}
	a := newCachedClient(t, prime)
	a.ModelOverride = "jev-1.14.0"
	if !a.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("prime failed")
	}
	// A FRESH client (its once-per-client notice has not fired) hits the
	// cache with the network closed: the served result must carry the
	// change notice naming the entry's model.
	b := New(true)
	b.LoadCredential = a.LoadCredential
	b.CacheDir = a.CacheDir
	b.ModelOverride = "jev-1.14.0"
	b.HTTP = &refusingDoer{}
	res := b.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("cache hit failed: %s", res.NoticeLine())
	}
	if res.Notice == "" || !strings.Contains(res.Notice, "jev-1.14.0") {
		t.Errorf("cache serve must carry the change notice for the entry's model, got %q", res.Notice)
	}
}

// AC-FW-011: usage aggregation separates new network requests from
// cache-served answers.
//
// FW-D2 deepening (t1968 debt-closure round, mutation-check finding): the
// test originally asserted the usage LOG only. The RESULT's Usage.FromCache —
// the surface an API consumer reads off the returned Result — was unasserted
// (a mutation disabling the cache-hit FromCache assignment escaped
// detection). Both surfaces are asserted now: log records and the per-call
// Result.
func TestUsage_FromCacheSeparation(t *testing.T) {
	doer := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.13.0")}}}
	c := newCachedClient(t, doer)
	net := c.Ask(context.Background(), cacheQuestion())
	if !net.OK() {
		t.Fatal("network call failed")
	}
	if net.Usage.FromCache {
		t.Error("network result Usage.FromCache must be false")
	}
	cached := c.Ask(context.Background(), cacheQuestion())
	if !cached.OK() {
		t.Fatal("cached call failed")
	}
	if !cached.Usage.FromCache {
		t.Error("cache-served result Usage.FromCache must be true (consumer surface, was unasserted)")
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

// R6 (card t1968 review): the in-flight slot is held until the response BODY
// has been read and closed, not just until the headers arrive — otherwise the
// concurrency a server observes (request + streaming body) exceeds
// MaxConcurrency while decodes are still reading. Observed inside the body
// read itself.
func TestSemaphore_HoldsSlotThroughBodyRead(t *testing.T) {
	const width = 2
	var reading, maxReading int64
	release := make(chan struct{})
	doer := DoerFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       gatingBody{release: release, reading: &reading, max: &maxReading},
			Header:     http.Header{},
			Request:    r,
		}, nil
	})
	c := newCachedClient(t, doer)
	c.MaxConcurrency = width
	c.CacheDir = "" // no cache: every call is network
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Ask(context.Background(), cacheQuestion()) // outcome is irrelevant; the probe is the read count
		}()
	}
	// Let every reader arrive at the blocked body, then release them all.
	time.Sleep(150 * time.Millisecond)
	close(release)
	wg.Wait()
	if got := atomic.LoadInt64(&maxReading); got > width {
		t.Errorf("observed %d concurrent body reads, semaphore width is %d (slot released before the body was read)", got, width)
	}
	if got := atomic.LoadInt64(&maxReading); got < 2 {
		t.Errorf("observed max body reads %d — the probe never exercised concurrency", got)
	}
}

// gatingBody counts a concurrent reader, blocks until release closes, then
// reports EOF.
type gatingBody struct {
	release <-chan struct{}
	reading *int64
	max     *int64
}

func (b gatingBody) Read([]byte) (int, error) {
	cur := atomic.AddInt64(b.reading, 1)
	for {
		old := atomic.LoadInt64(b.max)
		if cur <= old || atomic.CompareAndSwapInt64(b.max, old, cur) {
			break
		}
	}
	<-b.release
	atomic.AddInt64(b.reading, -1)
	return 0, io.EOF
}

func (b gatingBody) Close() error { return nil }

// DefaultCacheDir points under the home directory, never into a project.
func TestDefaultCacheDir_UnderHome(t *testing.T) {
	dir, err := DefaultCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(dir), ".moai/cache/jev") {
		// home-guard:allow string literal in an error message, not a home access
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

// L1 (t1968 leader audit): the wire questions travel as a MAP (toWire), so
// two requests differing only in question order marshal to one body, share
// one cache key, and the second is served as a hit. The stored entry carries
// the STORING request's answer order; the hit must reassemble the answers
// into the CURRENT request's question order — positional reuse of the stored
// slice would answer [q2,q1] with the first request's [a1,a2].
func TestCache_HitReassemblesAnswersToCurrentRequestOrder(t *testing.T) {
	twoQ := func(first, second string) Request {
		return Request{
			State: "the state under judgment",
			Questions: []Question{
				{ID: first, Text: "Is " + first + " alive?", Kind: KindNoul},
				{ID: second, Text: "Is " + second + " alive?", Kind: KindNoul},
			},
			Repeatable: true,
		}
	}
	body := `{"model":"jev-1.13.0","answers":{"q1":{"type":"noul","noul":0.9},"q2":{"type":"noul","noul":0.1}},"usage":{"input_tokens":7}}`
	prime := &cachedDoer{script: []doerStep{{status: 200, body: body}}}
	c := newCachedClient(t, prime)
	if first := c.Ask(context.Background(), twoQ("q1", "q2")); !first.OK() {
		t.Fatalf("prime failed: %s", first.NoticeLine())
	}

	// Same question set, opposite order: the wire body is identical (a map),
	// so a fresh process must be served from the cache with the network closed.
	fresh := New(true)
	fresh.LoadCredential = c.LoadCredential
	fresh.CacheDir = c.CacheDir
	refuse := &refusingDoer{}
	fresh.HTTP = refuse
	second := fresh.Ask(context.Background(), twoQ("q2", "q1"))
	if !second.OK() {
		t.Fatalf("cache hit failed: %s", second.NoticeLine())
	}
	if refuse.calls != 0 {
		t.Fatalf("identical wire body must be a cache hit, saw %d network calls", refuse.calls)
	}
	if len(second.Answers) != 2 {
		t.Fatalf("answers = %d, want 2", len(second.Answers))
	}
	if second.Answers[0].QuestionID != "q2" || second.Answers[1].QuestionID != "q1" {
		t.Errorf("a hit must answer in the CURRENT request's order (q2 first), got [%s %s]",
			second.Answers[0].QuestionID, second.Answers[1].QuestionID)
	}
	if second.Answers[0].Probability != 0.1 || second.Answers[1].Probability != 0.9 {
		t.Errorf("answers must keep their own values (identity, not position), got p(first)=%v p(second)=%v",
			second.Answers[0].Probability, second.Answers[1].Probability)
	}
}

// L2 (t1968 leader audit): a stale cache serve on a fresh client whose
// change notice has not yet fired sets BOTH Notice (entry model differs from
// VerifiedInterpretedModel) and StaleCacheWarning (prior-model entry).
// NoticeLine() surfaced only Notice, so the staleness fact was silently
// lost; the one line consumers emit must carry both facts.
func TestCache_StaleServeNoticeLineCarriesBothFacts(t *testing.T) {
	// body1 primed under jev-1.12.0 — differs from verified jev-1.13.0, so
	// the change notice fires on a fresh client.
	doer := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.12.0")}}}
	c := newCachedClient(t, doer)
	if !c.Ask(context.Background(), cacheQuestion()).OK() {
		t.Fatal("prime failed")
	}
	// The alias moves to jev-1.14.0 over a different body: jev-1.12.0 joins
	// the alias's recorded history as a prior model.
	mover := &cachedDoer{script: []doerStep{{status: 200, body: cacheOKBody("jev-1.14.0")}}}
	mc := New(true)
	mc.LoadCredential = c.LoadCredential
	mc.CacheDir = c.CacheDir
	mc.HTTP = mover
	moved := Request{State: "a different state", Questions: cacheQuestion().Questions}
	if !mc.Ask(context.Background(), moved).OK() {
		t.Fatal("alias-move call failed")
	}

	// Fresh process, network closed: the stale serve sets BOTH fields.
	fresh := New(true)
	fresh.LoadCredential = c.LoadCredential
	fresh.CacheDir = c.CacheDir
	refuse := &refusingDoer{}
	fresh.HTTP = refuse
	res := fresh.Ask(context.Background(), cacheQuestion())
	if !res.OK() {
		t.Fatalf("stale serve failed: %s", res.NoticeLine())
	}
	if res.Notice == "" || res.StaleCacheWarning == "" {
		t.Fatalf("scenario must set both fields, got notice %q warning %q", res.Notice, res.StaleCacheWarning)
	}
	line := res.NoticeLine()
	if !strings.Contains(line, res.StaleCacheWarning) {
		t.Errorf("the surfaced line must carry the staleness warning too, got %q", line)
	}
	if strings.Contains(line, "\n") {
		t.Errorf("the surfaced line must stay ONE line: %q", line)
	}
}
