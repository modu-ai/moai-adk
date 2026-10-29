// Package forge is the neutral boundary for talking to a code host (GitHub today; GitLab behind the
// same Transport later). It adds no network code of its own: a Transport is injected, so the package
// is fully testable with fakes and makes no remote call by itself.
//
// What it provides: conditional GET with a response cache (ETag / Last-Modified), one shared call
// Budget that follows the host's rate-limit headers, a verdict that tells a rate limit (primary or
// secondary, 429 or 403) from a permission denial, and wait times computed from Retry-After and the
// reset header. No limit value is hardcoded: the budget is learned from response headers, and the
// backoff numbers come from configuration or the documented defaults below.
//
// Documentation relied on (fetched 2026-10-29):
//   - https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api
//   - https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api
package forge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// Request is a host-neutral request. Header keys are lower-case.
type Request struct {
	Method string
	Path   string
	Header map[string]string
	Body   []byte // request body for writes; empty for GET
}

// Response is a host-neutral response. Header keys are lower-case.
type Response struct {
	Status int
	Header map[string]string
	Body   []byte
}

// Transport sends one request. The production implementation (gh api, net/http, or a GitLab client)
// lives outside this package.
type Transport interface {
	Do(ctx context.Context, r Request) (Response, error)
}

// ErrDenied marks a permission or authentication refusal. It is never retried; the caller parks.
var ErrDenied = errors.New("forge: permission denied (not retried; park the work)")

// Backoff holds the wait parameters. There are no hidden defaults: Validate rejects a zero value.
type Backoff struct {
	MinSecondaryWait time.Duration // wait when a limit is signalled with no Retry-After and no reset
	Base             time.Duration // first transient wait; doubles per attempt
	Max              time.Duration // cap for computed (not header-given) waits
	MaxAttempts      int           // attempts before the caller must give up
}

func (b Backoff) Validate() error {
	if b.MinSecondaryWait <= 0 || b.Base <= 0 || b.Max <= 0 || b.MaxAttempts <= 0 {
		return errors.New("forge: backoff needs MinSecondaryWait, Base, Max and MaxAttempts from configuration")
	}
	return nil
}

// Documented carries the two values the GitHub documentation states outright, with where they come from.
type Documented struct {
	Source           string
	MinSecondaryWait time.Duration // "wait at least one minute" when a secondary limit gives no hint
	MinWriteGap      time.Duration // "wait at least one second between each" write
}

// DocDefaults returns the documented values. Base, Max and MaxAttempts are NOT in the documentation
// ("a set number of retries") and must be configured.
func DocDefaults() Documented {
	return Documented{
		Source:           "https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api",
		MinSecondaryWait: time.Minute,
		MinWriteGap:      time.Second,
	}
}

// Kind is the outcome class of a response.
type Kind string

const (
	KindOK           Kind = "ok"
	KindNotModified  Kind = "not-modified"
	KindLimited      Kind = "rate-limited"
	KindDenied       Kind = "permission-denied"
	KindUnclassified Kind = "unclassified-403" // 403 with no rate-limit signal and no permission wording: stop, do not guess
	KindTransient    Kind = "transient"
	KindOther        Kind = "other"
)

// Verdict is Classify's answer. RetryAt is set for limited and transient responses.
type Verdict struct {
	Kind    Kind
	RetryAt time.Time
	Reason  string
}

func lower(h map[string]string) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		out[strings.ToLower(k)] = v
	}
	return out
}

func firstOf(h map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := h[k]; ok {
			return v
		}
	}
	return ""
}

func parseRetryAfter(v string, now time.Time) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	if n, err := strconv.Atoi(v); err == nil && n >= 0 {
		return now.Add(time.Duration(n) * time.Second), true
	}
	if t, err := http.ParseTime(v); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// parseReset reads a reset header: epoch seconds (GitHub x-ratelimit-reset, GitLab RateLimit-Reset)
// or an HTTP date (GitLab RateLimit-ResetTime).
func parseReset(h map[string]string) (time.Time, bool) {
	if t, ok := parseEpoch(firstOf(h, "x-ratelimit-reset", "ratelimit-reset")); ok {
		return t, true
	}
	if v := h["ratelimit-resettime"]; v != "" {
		if t, err := http.ParseTime(strings.TrimSpace(v)); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseEpoch(v string) (time.Time, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil || n <= 0 {
		return time.Time{}, false
	}
	return time.Unix(n, 0).UTC(), true
}

func grow(base time.Duration, attempt int, max time.Duration) time.Duration {
	d := base
	for i := 0; i < attempt && d < max; i++ {
		d *= 2
	}
	if d > max {
		d = max
	}
	return d
}

// Classify separates a rate limit from a permission denial. A 403 is a limit when it carries
// Retry-After, an exhausted remaining count, or the host's rate-limit wording; otherwise it is a
// denial and is never retried.
func Classify(r Response, now time.Time, b Backoff, attempt int) Verdict {
	h := lower(r.Header)
	switch {
	case r.Status >= 200 && r.Status < 300:
		return Verdict{Kind: KindOK}
	case r.Status == http.StatusNotModified:
		return Verdict{Kind: KindNotModified}
	case r.Status == http.StatusUnauthorized:
		return Verdict{Kind: KindDenied, Reason: "authentication failed"}
	case r.Status == http.StatusTooManyRequests || r.Status == http.StatusForbidden:
		_, hasRA := parseRetryAfter(h["retry-after"], now)
		exhausted := firstOf(h, "x-ratelimit-remaining", "ratelimit-remaining") == "0"
		worded := strings.Contains(strings.ToLower(string(r.Body)), "rate limit")
		if r.Status == http.StatusForbidden && !hasRA && !exhausted && !worded {
			if strings.Contains(strings.ToLower(string(r.Body)), "not accessible") {
				return Verdict{Kind: KindDenied, Reason: "resource not accessible (permission)"}
			}
			return Verdict{Kind: KindUnclassified, Reason: "403 with neither a rate-limit signal nor permission wording"}
		}
		if t, ok := parseRetryAfter(h["retry-after"], now); ok {
			return Verdict{Kind: KindLimited, RetryAt: t, Reason: "retry-after"}
		}
		if t, ok := parseReset(h); ok && exhausted {
			return Verdict{Kind: KindLimited, RetryAt: t, Reason: "reset time"}
		}
		return Verdict{Kind: KindLimited, RetryAt: now.Add(grow(b.MinSecondaryWait, attempt, b.Max)), Reason: "no hint; configured minimum"}
	case r.Status >= 500:
		return Verdict{Kind: KindTransient, RetryAt: now.Add(grow(b.Base, attempt, b.Max)), Reason: fmt.Sprintf("status %d", r.Status)}
	}
	return Verdict{Kind: KindOther, Reason: fmt.Sprintf("status %d", r.Status)}
}

// Budget is one call budget shared by every caller using the same credential. It learns the limit
// from response headers; it never assumes one.
type Budget struct {
	Reserve int // calls to keep unspent; from configuration

	mu           sync.Mutex
	remaining    int
	known        bool
	resetAt      time.Time
	blockedUntil time.Time
}

// Observe reads the rate-limit headers of a response.
func (b *Budget) Observe(r Response, now time.Time) {
	h := lower(r.Header)
	b.mu.Lock()
	defer b.mu.Unlock()
	if v := firstOf(h, "x-ratelimit-remaining", "ratelimit-remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			b.remaining, b.known = n, true
		}
	}
	if t, ok := parseReset(h); ok {
		b.resetAt = t
	}
}

// Block forbids calls until the given time.
func (b *Budget) Block(until time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if until.After(b.blockedUntil) {
		b.blockedUntil = until
	}
}

// Allow says whether a call may go out now, and otherwise when to look again.
func (b *Budget) Allow(now time.Time) (bool, time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if now.Before(b.blockedUntil) {
		return false, b.blockedUntil
	}
	if b.known && b.remaining <= b.Reserve && now.Before(b.resetAt) {
		return false, b.resetAt
	}
	if b.known && !now.Before(b.resetAt) && !b.resetAt.IsZero() {
		b.known = false // the window rolled over; forget the stale count
	}
	return true, time.Time{}
}

type entry struct {
	etag, lastMod string
	body          []byte
}

// Client does conditional GETs through a Transport and a shared Budget.
type Client struct {
	T       Transport
	B       *Budget
	Backoff Backoff
	Now     func() time.Time

	mu       sync.Mutex
	cache    map[string]entry
	attempts map[string]int
}

// Result is the outcome of one Get. When Deferred is set nothing usable came back and the caller
// should try again at NotBefore; the client itself never sleeps or retries.
type Result struct {
	Body         []byte
	Changed      bool // new content from the host
	FromCache    bool // 304: the cached body is still current
	Deferred     bool
	NotBefore    time.Time
	PollInterval time.Duration // from x-poll-interval, 0 when absent
}

// Get fetches a path with the stored validators. A permission denial returns ErrDenied.
func (c *Client) Get(ctx context.Context, path string) (Result, error) {
	now := c.Now()
	if ok, until := c.B.Allow(now); !ok {
		return Result{Deferred: true, NotBefore: until}, nil
	}
	c.mu.Lock()
	if c.cache == nil {
		c.cache, c.attempts = map[string]entry{}, map[string]int{}
	}
	e, have := c.cache[path]
	attempt := c.attempts[path]
	c.mu.Unlock()

	req := Request{Method: http.MethodGet, Path: path, Header: map[string]string{}}
	if have && e.etag != "" {
		req.Header["if-none-match"] = e.etag
	}
	if have && e.lastMod != "" {
		req.Header["if-modified-since"] = e.lastMod
	}
	resp, err := c.T.Do(ctx, req)
	if err != nil {
		c.bump(path)
		return Result{Deferred: true, NotBefore: now.Add(grow(c.Backoff.Base, attempt, c.Backoff.Max))}, nil
	}
	c.B.Observe(resp, now)
	v := Classify(resp, now, c.Backoff, attempt)
	h := lower(resp.Header)
	res := Result{}
	if n, err := strconv.Atoi(h["x-poll-interval"]); err == nil && n > 0 {
		res.PollInterval = time.Duration(n) * time.Second
	}
	switch v.Kind {
	case KindOK:
		c.mu.Lock()
		c.cache[path] = entry{etag: h["etag"], lastMod: h["last-modified"], body: resp.Body}
		c.attempts[path] = 0
		c.mu.Unlock()
		res.Body, res.Changed = resp.Body, true
		return res, nil
	case KindNotModified:
		if !have {
			return Result{Deferred: true, NotBefore: now.Add(grow(c.Backoff.Base, attempt, c.Backoff.Max))}, nil
		}
		c.mu.Lock()
		c.attempts[path] = 0
		c.mu.Unlock()
		res.Body, res.FromCache = e.body, true
		return res, nil
	case KindLimited:
		c.B.Block(v.RetryAt)
		c.bump(path)
		res.Deferred, res.NotBefore = true, v.RetryAt
		return res, nil
	case KindTransient:
		c.bump(path)
		res.Deferred, res.NotBefore = true, v.RetryAt
		return res, nil
	case KindDenied, KindUnclassified:
		return Result{}, fmt.Errorf("%w: %s", ErrDenied, v.Reason)
	}
	return Result{}, fmt.Errorf("forge: unexpected response: %s", v.Reason)
}

func (c *Client) bump(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.attempts == nil {
		c.cache, c.attempts = map[string]entry{}, map[string]int{}
	}
	c.attempts[path]++
}

// EffectError maps a verdict to the error the effect executor understands, so a rate limit waits for
// its RetryAt and a permission denial parks the intent. Other kinds map to nil.
func EffectError(v Verdict) error {
	switch v.Kind {
	case KindLimited:
		return &effect.LimitedError{RetryAt: v.RetryAt}
	case KindDenied, KindUnclassified:
		return fmt.Errorf("%w: %s", effect.ErrDenied, v.Reason)
	}
	return nil
}

// StatusError is a typed HTTP failure. A Transport or Runner returns it so callers decide on the
// status code and headers, never on error wording.
type StatusError struct {
	Status int
	Header map[string]string
	Body   []byte
}

func (e *StatusError) Error() string { return fmt.Sprintf("forge: http status %d", e.Status) }

// ClassifyError turns a *StatusError (anywhere in err's chain) into a verdict; ok is false for any other error.
func ClassifyError(err error, now time.Time, b Backoff, attempt int) (Verdict, bool) {
	var se *StatusError
	if !errors.As(err, &se) {
		return Verdict{}, false
	}
	return Classify(Response{Status: se.Status, Header: se.Header, Body: se.Body}, now, b, attempt), true
}

// ForEffect maps a failed call to what the effect executor understands: a rate limit waits for its
// RetryAt (when a valid Backoff is configured), a permission or missing-repository answer parks the
// intent, and anything else is returned unchanged. The decision uses the status code only.
func ForEffect(err error, now time.Time, b Backoff) error {
	var se *StatusError
	if !errors.As(err, &se) {
		return err
	}
	if b.Validate() == nil {
		if v, ok := ClassifyError(err, now, b, 0); ok {
			if mapped := EffectError(v); mapped != nil {
				return mapped
			}
		}
	}
	switch se.Status {
	case 401, 403, 404:
		return fmt.Errorf("%w: http %d", effect.ErrDenied, se.Status)
	}
	return err
}
