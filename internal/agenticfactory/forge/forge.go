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
// Consumer contract per Kind (what a caller of Client/Classify must do). Adding or changing a Kind means
// updating this table in the same change and telling the consumers (lane 18's mirror and claim code,
// lane 22's tests); TestEveryKindHasAConsumerContract fails when a Kind is missing from the table.
//
//	ok                 use the body.
//	not-modified       use the cached body; nothing changed.
//	rate-limited       STOP sending on that host resource until RetryAt; keep the work and retry later,
//	                   bounded by Backoff.MaxAttempts; no block record until the attempts run out.
//	transient          retry at RetryAt, bounded by Backoff.MaxAttempts, then give up.
//	permission-denied  STOP, no retry; record a block (PARKED, reason permission-denied).
//	unclassified-403   STOP, no retry; record a block (PARKED, reason unclassified-403); do not guess.
//	other              caller decides (a 404 is not evidence of deletion or of missing permission).
//
// After MaxAttempts consecutive rate-limited/transient answers for one path the client returns
// ErrGiveUp: treat it like a denial (STOP, record a block).
//
// Documentation relied on (fetched 2026-10-29):
//   - https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api
//   - https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api
//   - https://docs.github.com/en/rest/using-the-rest-api/troubleshooting-the-rest-api
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

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
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
	// Redirected is set by the Transport when it followed a redirect (a renamed or moved repository)
	// to produce this response. The caller must not treat the answer as being for the repository it asked.
	Redirected bool
}

// Transport sends one request. The production implementation (gh api, net/http, or a GitLab client)
// lives outside this package.
type Transport interface {
	Do(ctx context.Context, r Request) (Response, error)
}

// ErrDenied marks a permission or authentication refusal. It is never retried; the caller parks.
var ErrDenied = errors.New("forge: permission denied (not retried; park the work)")

// ErrGiveUp means the same path was rate-limited or failing transiently Backoff.MaxAttempts times in a
// row. The caller stops and records a block; the client never loops on its own.
var ErrGiveUp = errors.New("forge: retry attempts exhausted (stop and record a block)")

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
	if b.Max < b.MinSecondaryWait || b.Max < b.Base {
		return errors.New("forge: backoff Max must not be below MinSecondaryWait or Base (a cap may not shorten the documented minimum)")
	}
	return nil
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
	// Signal says what decided a rate-limit verdict: "header" (Retry-After, an exhausted remaining count)
	// or "wording" (the host's message text only). It is empty for other kinds.
	Signal string
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
	if max < base {
		max = base // a cap never shortens the base wait, even when the config was not validated
	}
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
		exhausted := firstOf(h, "x-ratelimit-remaining", "ratelimit-remaining") == "0"
		if t, ok := parseRetryAfter(h["retry-after"], now); ok {
			return Verdict{Kind: KindLimited, RetryAt: t, Reason: "retry-after", Signal: "header"}
		}
		if t, ok := parseReset(h); ok && exhausted {
			return Verdict{Kind: KindLimited, RetryAt: t, Reason: "reset time", Signal: "header"}
		}
		if exhausted || r.Status == http.StatusTooManyRequests {
			return Verdict{Kind: KindLimited, RetryAt: now.Add(grow(b.MinSecondaryWait, attempt, b.Max)), Reason: "no hint; configured minimum", Signal: "header"}
		}
		// A 403 with no usable header. GitHub's documentation separates a secondary rate limit from a
		// permission error by the message text only (rate-limits page: the response carries "an error
		// message that indicates that you exceeded a secondary rate limit"; troubleshooting page, both
		// fetched 2026-10-29: "Resource not accessible by integration" / "... by personal access token"
		// for missing permissions). The exact secondary-limit text is not published, so the wording is
		// a fallback signal used ONLY here, when no header decided. Every outcome of this branch is on
		// the safe side: the rate-limit reading waits at least the configured minimum and is bounded by
		// MaxAttempts, and anything else stops. The "not accessible" wording changes the reason shown,
		// never the behaviour: denied and unclassified both stop without a retry.
		body := strings.ToLower(string(r.Body))
		if r.Status == http.StatusForbidden && strings.Contains(body, "rate limit") {
			return Verdict{Kind: KindLimited, RetryAt: now.Add(grow(b.MinSecondaryWait, attempt, b.Max)), Reason: "secondary limit wording; configured minimum", Signal: "wording"}
		}
		if strings.Contains(body, "not accessible") {
			return Verdict{Kind: KindDenied, Reason: "resource not accessible (permission)"}
		}
		return Verdict{Kind: KindUnclassified, Reason: "403 with neither a rate-limit signal nor permission wording"}
	case r.Status >= 500:
		return Verdict{Kind: KindTransient, RetryAt: now.Add(grow(b.Base, attempt, b.Max)), Reason: fmt.Sprintf("status %d", r.Status)}
	}
	return Verdict{Kind: KindOther, Reason: fmt.Sprintf("status %d", r.Status)}
}

// bucket is the budget state of one host resource (GitHub reports the resource in
// x-ratelimit-resource; "" is the bucket for responses and requests that name none).
type bucket struct {
	remaining    int
	known        bool
	resetAt      time.Time
	blockedUntil time.Time
}

// Budget is the call budget shared by every caller using the same credential, kept per host resource
// so an exhausted search quota does not stop ordinary calls. It learns limits from response headers;
// it never assumes one. ReserveCalls is the number of calls kept unspent for writes more urgent than a
// report. Budget implements core.Budget, the hook lane 18's writers call before a report write.
type Budget struct {
	ReserveCalls   int              // from configuration
	ReportResource string           // host resource report writes count against ("" = default bucket)
	Now            func() time.Time // optional clock for Reserve; defaults to time.Now

	mu sync.Mutex
	b  map[string]*bucket
}

func (b *Budget) at(res string) *bucket {
	if b.b == nil {
		b.b = map[string]*bucket{}
	}
	if x, ok := b.b[res]; ok {
		return x
	}
	x := &bucket{}
	b.b[res] = x
	return x
}

// ResourceOf returns the host resource a response was counted against ("" when it names none).
func ResourceOf(r Response) string { return resourceOf(lower(r.Header)) }

func resourceOf(h map[string]string) string {
	return firstOf(h, "x-ratelimit-resource", "ratelimit-name")
}

// Observe reads the rate-limit headers of a response into the bucket the response names.
func (b *Budget) Observe(r Response, now time.Time) {
	h := lower(r.Header)
	b.mu.Lock()
	defer b.mu.Unlock()
	x := b.at(resourceOf(h))
	if v := firstOf(h, "x-ratelimit-remaining", "ratelimit-remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			x.remaining, x.known = n, true
		}
	}
	if t, ok := parseReset(h); ok {
		x.resetAt = t
	}
}

// RemainingFor reports the remaining call count the host last stated for a resource. known is false when
// the host has not stated one - no header was seen, or the window it described has rolled over - and the
// caller must then read it as UNKNOWN, never as "some are left" and never as a number of its own.
func (b *Budget) RemainingFor(res string) (n int, known bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	x := b.at(res)
	return x.remaining, x.known
}

// Block forbids calls on the default bucket until the given time; BlockFor does it for one resource.
func (b *Budget) Block(until time.Time) { b.BlockFor("", until) }

func (b *Budget) BlockFor(res string, until time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if x := b.at(res); until.After(x.blockedUntil) {
		x.blockedUntil = until
	}
}

// Allow says whether an ordinary call may go out now on the default bucket, and otherwise when to look again.
func (b *Budget) Allow(now time.Time) (bool, time.Time) { return b.AllowFor("", now) }

// AllowFor is Allow for one resource. Ordinary calls stop at the reserve.
func (b *Budget) AllowFor(res string, now time.Time) (bool, time.Time) {
	return b.allow(res, now, b.ReserveCalls)
}

// Reserve implements core.Budget. A report write is told to wait (false) when ordinary calls are
// stopped for its resource; anything more urgent than a report is never held, so it reports true.
func (b *Budget) Reserve(p core.Priority) bool {
	if p < core.PriorityReport {
		return true
	}
	now := time.Now()
	if b.Now != nil {
		now = b.Now()
	}
	ok, _ := b.AllowFor(b.ReportResource, now)
	return ok
}

var _ core.Budget = (*Budget)(nil)

// AllowReserved is AllowFor for state and safety writes: they may spend the reserve, down to the
// host's own zero.
func (b *Budget) AllowReserved(res string, now time.Time) (bool, time.Time) {
	return b.allow(res, now, 0)
}

func (b *Budget) allow(res string, now time.Time, floor int) (bool, time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	x := b.at(res)
	if now.Before(x.blockedUntil) {
		return false, x.blockedUntil
	}
	if x.known && x.remaining <= floor && now.Before(x.resetAt) {
		return false, x.resetAt
	}
	if x.known && !x.resetAt.IsZero() && !now.Before(x.resetAt) {
		x.known = false // the window rolled over; forget the stale count
	}
	return true, time.Time{}
}

type entry struct {
	etag, lastMod string
	body          []byte
	redirected    bool // the response that produced this body came through a redirect
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
	learned  map[string]string // path -> host resource, from the last response for that path
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
	Redirected   bool          // the host answered through a redirect
}

// Get fetches a path with the stored validators. A permission denial returns ErrDenied.
func (c *Client) Get(ctx context.Context, path string) (Result, error) {
	now := c.Now()
	c.mu.Lock()
	if c.cache == nil {
		c.cache, c.attempts, c.learned = map[string]entry{}, map[string]int{}, map[string]string{}
	}
	res0 := c.learned[path]
	c.mu.Unlock()
	if ok, until := c.B.AllowFor(res0, now); !ok {
		return Result{Deferred: true, NotBefore: until}, nil
	}
	c.mu.Lock()
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
		if c.bump(path) >= c.Backoff.MaxAttempts {
			return Result{}, fmt.Errorf("%w: transport error", ErrGiveUp)
		}
		return Result{Deferred: true, NotBefore: now.Add(grow(c.Backoff.Base, attempt, c.Backoff.Max))}, nil
	}
	c.B.Observe(resp, now)
	v := Classify(resp, now, c.Backoff, attempt)
	h := lower(resp.Header)
	resName := resourceOf(h)
	c.mu.Lock()
	c.learned[path] = resName
	c.mu.Unlock()
	res := Result{Redirected: resp.Redirected}
	if n, err := strconv.Atoi(h["x-poll-interval"]); err == nil && n > 0 {
		res.PollInterval = time.Duration(n) * time.Second
	}
	switch v.Kind {
	case KindOK:
		c.mu.Lock()
		c.cache[path] = entry{etag: h["etag"], lastMod: h["last-modified"], body: resp.Body, redirected: resp.Redirected}
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
		res.Redirected = res.Redirected || e.redirected // the cached body keeps the redirect it arrived with
		return res, nil
	case KindLimited:
		c.B.BlockFor(resName, v.RetryAt)
		if c.bump(path) >= c.Backoff.MaxAttempts {
			return Result{}, fmt.Errorf("%w: %s", ErrGiveUp, v.Reason)
		}
		res.Deferred, res.NotBefore = true, v.RetryAt
		return res, nil
	case KindTransient:
		if c.bump(path) >= c.Backoff.MaxAttempts {
			return Result{}, fmt.Errorf("%w: %s", ErrGiveUp, v.Reason)
		}
		res.Deferred, res.NotBefore = true, v.RetryAt
		return res, nil
	case KindDenied, KindUnclassified:
		return Result{}, fmt.Errorf("%w: %s", ErrDenied, v.Reason)
	}
	return Result{}, fmt.Errorf("forge: unexpected response: %s", v.Reason)
}

func (c *Client) bump(path string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.attempts == nil {
		c.cache, c.attempts, c.learned = map[string]entry{}, map[string]int{}, map[string]string{}
	}
	c.attempts[path]++
	return c.attempts[path]
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
	// Body is the host's own answer, raw: Redact it before it is recorded anywhere. Error() prints the
	// status only.
	Body []byte
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
