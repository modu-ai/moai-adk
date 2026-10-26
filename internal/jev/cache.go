// cache.go — SPEC-FETCH-WEB-001 M5 (REQ-FW-011..REQ-FW-013). The client
// cache/retry surface: the disk cache keyed by request-body SHA combined
// with the RESOLVED interpretation model id (never the jev-latest alias),
// written by atomic replace; the Retry-After backoff composing with
// REQ-JEVC-008's repeatable/pre-delivery rules; the concurrency semaphore;
// and the new-vs-cache usage separation.
//
// Scope guard (REQ-FW-016): this file is CLIENT surface only. The
// ledger/consumers work of card t1762 lives in internal/jevledger and is
// untouched.
//
// The package stays standard-library-only (see display_only_test.go): the
// cache writes only under the cache directory it is given, never into a
// project, a repo, or any state another component reads.
package jev

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// maxRetryAfterDelay bounds a parsed Retry-After before the client will
// honor it: a header naming an hour must not stall a call for an hour.
const maxRetryAfterDelay = 30 * time.Second

// defaultMaxConcurrency is the in-flight request cap a client constructed
// by New carries (REQ-FW-012's semaphore limb). Zero on a hand-built Client
// struct means no cap; New sets the default.
const defaultMaxConcurrency = 4

// modelsMapName is the persisted alias→model map's file name inside the
// cache dir. The map is what makes the restart-then-offline hit work: a
// fresh process resolves the same id from disk and serves the cached answer
// with the network closed.
const modelsMapName = "models.json"

// DefaultCacheDir returns the production cache directory,
// <home>/.moai/cache/jev. The caller decides whether to use it; tests
// override CacheDir with a temp dir and NEVER touch this path.
func DefaultCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("jev: home directory unavailable for the cache path: %w", err)
	}
	return filepath.Join(home, ".moai", "cache", "jev"), nil
}

// cacheEntry is one cached answer on disk.
type cacheEntry struct {
	// BodySHA is the SHA-256 of the marshaled wire request body.
	BodySHA string `json:"body_sha"`
	// Model is the interpretation model id the answer is keyed under — the
	// resolved id (or explicit override), never the request alias.
	Model string `json:"model"`
	// Answers is the flattened caller-facing answer list.
	Answers []Answer `json:"answers"`
	// Usage is the recorded usage of the call that produced the entry.
	Usage Usage `json:"usage"`
	// CreatedAt orders prior-model entries for the keep-and-warn probe.
	CreatedAt time.Time `json:"created_at"`
}

// modelsMap is the persisted alias→model map.
type modelsMap struct {
	Models map[string]string `json:"models"`
}

// cacheKey hashes the body SHA and model id into the cache file name. The
// composite key is the whole point: two model ids over one body produce two
// files, and an alias move cannot silently re-attribute an old answer.
func cacheKey(bodySHA, model string) string {
	h := sha256.Sum256([]byte(bodySHA + "\x00" + model))
	return hex.EncodeToString(h[:])
}

// bodySHA hashes the marshaled request body.
func bodySHA(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

// resolveModel applies the lookup sequence (plan B5): (1) an explicit
// per-request override; (2) a configured default interpretation model (no
// such knob exists on this tree — the sequence falls through); (3) alias
// resolution from the persisted alias→model map. An explicit override
// returns with aliasResolved=false: its misses NEVER probe another model's
// entry — they go to the network, whose stored entry is attributed to the
// model id the response reported (the override's own exactly when the
// endpoint answers with it) — the N5 applicability condition. An
// unresolvable id (empty map entry, first-ever call) returns ok=false and
// the call simply proceeds to the network.
func (c *Client) resolveModel() (model string, aliasResolved bool, ok bool) {
	if c.ModelOverride != "" {
		return c.ModelOverride, false, true
	}
	if m, err := readModelsMap(c.CacheDir); err == nil {
		if id := m.Models[ModelID]; id != "" {
			return id, true, true
		}
	}
	return "", false, false
}

// readModelsMap loads the persisted alias→model map; any absence or damage
// reads as empty (fail-open to "no model known": the network call then
// re-establishes it).
func readModelsMap(dir string) (modelsMap, error) {
	var m modelsMap
	raw, err := os.ReadFile(filepath.Join(dir, modelsMapName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return modelsMap{Models: map[string]string{}}, nil
		}
		return m, err
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return modelsMap{Models: map[string]string{}}, nil
	}
	if m.Models == nil {
		m.Models = map[string]string{}
	}
	return m, nil
}

// writeModelsMap persists the map by atomic replace.
func writeModelsMap(dir string, m modelsMap) error {
	raw, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(dir, modelsMapName), raw)
}

// atomicWrite writes data to path via a temp file + rename, so a concurrent
// reader never observes a partial file (REQ-FW-011).
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }() // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// cacheLookup probes the cache for body under the resolved model id. On a
// direct hit it returns the entry. On a miss it returns aliasResolved=true
// only for alias-resolved lookups, where the keep-and-warn probe may then
// serve the most recent prior-model entry for the same body with a
// staleness warning naming the cached model id — the entry is KEPT, never
// deleted (plan B5's stale-selection policy).
func (c *Client) cacheLookup(sha, model string, aliasResolved bool) (entry cacheEntry, hit, stale bool, err error) {
	raw, readErr := os.ReadFile(filepath.Join(c.CacheDir, cacheKey(sha, model)+".json"))
	if readErr == nil {
		var e cacheEntry
		if jsonErr := json.Unmarshal(raw, &e); jsonErr == nil && e.Model == model {
			return e, true, false, nil
		}
	}
	if !aliasResolved {
		// An explicit override's miss goes to the network — never another
		// model's entry (the N5 applicability condition).
		return cacheEntry{}, false, false, nil
	}
	prior, perr := c.priorModelEntry(sha, model)
	if perr != nil || prior == nil {
		return cacheEntry{}, false, false, nil
	}
	return *prior, true, true, nil
}

// priorModelEntry finds the most recent cache entry for the same body SHA
// keyed under a DIFFERENT model id.
func (c *Client) priorModelEntry(sha, resolvedModel string) (*cacheEntry, error) {
	files, err := filepath.Glob(filepath.Join(c.CacheDir, "*.json"))
	if err != nil || len(files) == 0 {
		return nil, err
	}
	var prior *cacheEntry
	for _, f := range files {
		if filepath.Base(f) == modelsMapName {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var e cacheEntry
		if json.Unmarshal(raw, &e) != nil || e.BodySHA != sha || e.Model == resolvedModel || e.Model == "" {
			continue
		}
		if prior == nil || e.CreatedAt.After(prior.CreatedAt) {
			e := e
			prior = &e
		}
	}
	return prior, nil
}

// cacheStore writes the entry by atomic replace. Stale entries are never
// deleted — a prior model's answers stay valid for their own key.
func (c *Client) cacheStore(entry cacheEntry) error {
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(c.CacheDir, cacheKey(entry.BodySHA, entry.Model)+".json"), raw)
}

// serveCacheHit shapes a Result from a cached entry. The change notice is
// not re-derived: the notice belongs to the observation that produced the
// entry, and a serve is not an observation.
func serveCacheHit(entry cacheEntry, stale bool) Result {
	usage := entry.Usage
	usage.FromCache = true
	res := Result{
		Availability: Available,
		Answers:      entry.Answers,
		Usage:        usage,
	}
	if stale {
		res.StaleCacheWarning = fmt.Sprintf(
			"jev cache: served an entry answered by %s; the alias has since moved past it (entry kept)",
			entry.Model)
	}
	return res
}

// updateAliasMap records the interpreted model a network answer reported for
// the request alias, persisting the map. Only alias-resolved calls update
// it: an explicit override observes nothing about where the alias points.
func (c *Client) updateAliasMap(model string) error {
	m, err := readModelsMap(c.CacheDir)
	if err != nil {
		return err
	}
	if m.Models[ModelID] == model {
		return nil
	}
	m.Models[ModelID] = model
	return writeModelsMap(c.CacheDir, m)
}

// parseRetryAfter parses a Retry-After header value: delta-seconds or an
// HTTP-date, bounded by maxRetryAfterDelay. An unparseable value does not
// parse — the client then does not wait on it.
func parseRetryAfter(v string) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n >= 0 {
		d := time.Duration(n) * time.Second
		if d > maxRetryAfterDelay {
			d = maxRetryAfterDelay
		}
		return d, true
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d < 0 {
			d = 0
		}
		if d > maxRetryAfterDelay {
			d = maxRetryAfterDelay
		}
		return d, true
	}
	return 0, false
}

// retryableStatus reports whether a response status carries the Retry-After
// limb (REQ-FW-012): 429, 529, or another 5xx.
func retryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code == 529 || code >= 500
}

// semaphore acquires an in-flight slot when the client carries a
// concurrency cap, releasing via the returned function. The channel is
// built lazily so a MaxConcurrency set after construction is honored.
func (c *Client) semaphore(ctx context.Context) (release func(), err error) {
	c.mu.Lock()
	if c.sem == nil && c.MaxConcurrency > 0 {
		c.sem = make(chan struct{}, c.MaxConcurrency)
	}
	sem := c.sem
	c.mu.Unlock()
	if sem == nil {
		return func() {}, nil
	}
	select {
	case sem <- struct{}{}:
		return func() { <-sem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// retryAfterSleep waits d, honoring cancellation.
func retryAfterSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
