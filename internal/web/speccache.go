package web

import (
	"sync"
	"time"
)

// specCacheMaxAge bounds how long a cached SPEC scan is served without an
// invalidation. The hub's "spec" signal is the primary invalidation, but the
// watcher is not recursive: an in-place edit of a file inside a SPEC directory
// produces no event (a SPEC directory created, removed or renamed does, and on
// darwin so does an atomic rename-into-place). This age is the safety net that
// keeps a missed event from serving stale rows forever. It counts from when a
// scan completed.
const specCacheMaxAge = 60 * time.Second

// specLoader computes the SPEC rows and per-SPEC drift findings for a project
// root. loadSpecRows is the production loader; tests inject counting loaders.
type specLoader func(root string) ([]SpecRowVM, map[string][]FindingVM, error)

// specCall is one in-flight scan that concurrent callers wait on.
type specCall struct {
	done     chan struct{}
	rows     []SpecRowVM
	findings map[string][]FindingVM
	err      error
}

// specCache keeps the result of the full SPEC scan (spec.ListDocs + spec.Audit
// over every SPEC — seconds on a large catalog) in process memory and reuses it
// until invalidate is called or specCacheMaxAge passes (card t1460).
//
// Concurrent misses share one scan. An invalidation that lands while a scan is
// in flight bumps gen, so that scan's possibly pre-change result is handed to
// the callers already waiting on it but never stored as fresh.
//
// Errors are not cached: the next request retries the scan.
//
// The returned rows slice is a per-caller copy; the findings map is shared and
// MUST be treated as read-only by callers.
type specCache struct {
	load   specLoader
	maxAge time.Duration
	now    func() time.Time

	mu       sync.Mutex
	gen      uint64
	have     bool
	root     string
	at       time.Time
	rows     []SpecRowVM
	findings map[string][]FindingVM
	inflight *specCall
}

func newSpecCache(load specLoader, maxAge time.Duration) *specCache {
	return &specCache{load: load, maxAge: maxAge, now: time.Now}
}

// get returns the cached scan for root, computing it at most once per
// invalidation across concurrent callers.
func (c *specCache) get(root string) ([]SpecRowVM, map[string][]FindingVM, error) {
	c.mu.Lock()
	if c.have && c.root == root && c.now().Sub(c.at) < c.maxAge {
		rows, findings := copyRows(c.rows), c.findings
		c.mu.Unlock()
		return rows, findings, nil
	}
	if call := c.inflight; call != nil {
		c.mu.Unlock()
		<-call.done
		return copyRows(call.rows), call.findings, call.err
	}
	call := &specCall{done: make(chan struct{})}
	c.inflight = call
	gen := c.gen
	c.mu.Unlock()

	call.rows, call.findings, call.err = c.load(root)

	c.mu.Lock()
	c.inflight = nil
	if call.err == nil && gen == c.gen {
		// Age counts from completion: a scan taking tens of seconds on a loaded
		// machine must not spend its own max age. Changes during the scan are
		// covered by gen, not by the age.
		c.have, c.root, c.at = true, root, c.now()
		c.rows, c.findings = call.rows, call.findings
	}
	c.mu.Unlock()
	close(call.done)
	return copyRows(call.rows), call.findings, call.err
}

// invalidate drops the cached scan; the next get recomputes it.
func (c *specCache) invalidate() {
	c.mu.Lock()
	c.gen++
	c.have = false
	c.rows, c.findings = nil, nil
	c.mu.Unlock()
}

func copyRows(rows []SpecRowVM) []SpecRowVM {
	if rows == nil {
		return nil
	}
	return append([]SpecRowVM(nil), rows...)
}
