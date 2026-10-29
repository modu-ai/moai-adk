// Package ctxledger sends and reconciles the per-Issue context ledger comment: exactly one comment
// per Issue, found by a marker, updated in place with only the changed content.
//
// It rides the effect executor: the body is recorded durably first (Submit never waits for the
// upload), the key is the digest of the exact text, same text is never sent again, a lost response
// is resolved by looking the marker up, and a body is sent only after masking has passed.
package ctxledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// ErrNotMasked is returned for a body without the masking-passed flag, or one that still holds a secret.
var ErrNotMasked = errors.New("ctxledger: body has not passed masking; refusing to record or send")

// Marker identifies the one context comment of an Issue. It differs from the state-ledger marker.
func Marker(issue int) string {
	return "<!-- moai-agentic-context:v1 issue=" + strconv.Itoa(issue) + " -->"
}

// Body is the text to publish. MaskPassed is set only by the masking step (lane 18's redaction).
type Body struct {
	Issue      int
	Text       string
	MaskPassed bool
}

// Comment is one Issue comment as the remote reports it.
type Comment struct {
	ID   int64
	Body string
}

// ErrGone is returned by Remote.Update when the comment no longer exists (HTTP 404).
var ErrGone = errors.New("ctxledger: comment no longer exists")

var (
	emailRe = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	homeRe  = regexp.MustCompile(`(?:/Users|/home)/[^/\s]+|[A-Za-z]:\\Users\\[^\\\s]+`)
)

// personalData reports an email address or an absolute home path. The masking step upstream does not
// cover these, so the sender refuses them whatever the flag says.
func personalData(s string) bool { return emailRe.MatchString(s) || homeRe.MatchString(s) }

// ErrNoConditionalRead is returned by a wrapper whose inner Remote cannot read conditionally. The drift
// check then falls back to the marker lookup instead of treating it as a failed check.
var ErrNoConditionalRead = errors.New("ctxledger: the remote has no conditional read")

// Fetcher is the optional cheap read used to notice that a machine-owned comment was edited by hand.
// It is a conditional GET of one comment (If-None-Match with the ETag): notModified=true means the
// remote still holds what the last read saw. See ForgeFetcher for the forge-backed implementation.
type Fetcher interface {
	Fetch(ctx context.Context, issue int, id int64, etag string) (c Comment, newETag string, notModified bool, err error)
}

// Remote is the comment API. The only production implementation is GHRemote.
type Remote interface {
	Find(ctx context.Context, issue int, marker string) (Comment, bool, error)
	Create(ctx context.Context, issue int, body string) (Comment, error)
	Update(ctx context.Context, issue int, id int64, body string) error
}

// Ledger is the sender/reconciler.
type Ledger struct {
	Remote   Remote
	Store    effect.IntentStore
	MaxBytes int
	Timeouts effect.Timeouts
	// Backoff and Now are optional. With a valid Backoff a typed rate-limit error sets the intent's
	// NotBefore; without it a rate limit is treated like any other failed send.
	Backoff forge.Backoff
	Now     func() time.Time

	// cache holds, per Issue, the comment id and body hash of the last write known to have landed.
	// With it a write skips the marker list GET, and an unchanged body skips the write. It is only a
	// hint: it is dropped on any ambiguous outcome or a 404, and the marker lookup is the fallback.
	mu      sync.Mutex
	cache   map[int]cached
	recheck map[int]bool // issues whose same text was submitted again: verify the remote before trusting the hash
}

type cached struct {
	id   int64
	hash string
	etag string // validator of the last conditional read; empty when none
}

func (l *Ledger) cacheGet(issue int) (cached, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	c, ok := l.cache[issue]
	return c, ok
}

func (l *Ledger) cachePut(issue int, c cached) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cache == nil {
		l.cache = map[int]cached{}
	}
	l.cache[issue] = c
}

func (l *Ledger) markRecheck(issue int, on bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.recheck == nil {
		l.recheck = map[int]bool{}
	}
	if on {
		l.recheck[issue] = true
	} else {
		delete(l.recheck, issue)
	}
}

func (l *Ledger) cacheDrop(issue int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.cache, issue)
}

func digest(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func (l *Ledger) compose(b Body) string {
	full := Marker(b.Issue) + "\n" + b.Text
	if l.MaxBytes <= 0 || len(full) <= l.MaxBytes {
		return full
	}
	note := fmt.Sprintf("\n…[summary: %d bytes omitted]", len(full))
	room := l.MaxBytes - len(Marker(b.Issue)) - 1 - len(note)
	if room < 0 {
		room = 0
	}
	cut := b.Text
	if len(cut) > room {
		cut = cut[:room]
		for !utf8.ValidString(cut) && len(cut) > 0 {
			cut = cut[:len(cut)-1]
		}
	}
	return Marker(b.Issue) + "\n" + cut + note
}

func target(issue int) string { return "issue-" + strconv.Itoa(issue) }

func (l *Ledger) intents(issue int) ([]effect.Intent, error) {
	all, err := l.Store.List()
	if err != nil {
		return nil, err
	}
	var out []effect.Intent
	for _, in := range all {
		if in.Kind == effect.LedgerComment && in.Issue == issue {
			out = append(out, in)
		}
	}
	return out, nil
}

// Submit records the body as an intent and returns at once. It never touches the remote.
func (l *Ledger) Submit(_ context.Context, b Body) (effect.Intent, error) {
	if !b.MaskPassed {
		return effect.Intent{}, ErrNotMasked
	}
	if _, changed := monitor.Mask(b.Text); changed || personalData(b.Text) {
		return effect.Intent{}, ErrNotMasked
	}
	full := l.compose(b)
	dg := digest(full)
	key := effect.Key(effect.LedgerComment, target(b.Issue), dg)
	mine, err := l.intents(b.Issue)
	if err != nil {
		return effect.Intent{}, err
	}
	latest := 0
	for _, in := range mine {
		if in.Seq > latest {
			latest = in.Seq
		}
	}
	for _, in := range mine {
		if in.Key != key {
			continue
		}
		if in.Seq == latest {
			l.markRecheck(b.Issue, true) // the comment is machine-owned: confirm nobody edited it, then send nothing
			return in, nil
		}
		in.Status, in.Attempts, in.ProviderID, in.LastError, in.Seq = effect.StatusIntended, 0, "", "", latest+1
		return in, l.Store.Put(in) // older text submitted again: re-arm so the remote is brought back to it
	}
	in := effect.Intent{Key: key, Kind: effect.LedgerComment, Issue: b.Issue, Target: target(b.Issue), Digest: dg, Payload: full, Status: effect.StatusIntended, Seq: latest + 1}
	return in, l.Store.Put(in)
}

// Flush sends the newest pending intent of every Issue, one at a time. Older pending intents are
// superseded; a failure of one Issue does not stop the others.
func (l *Ledger) Flush(ctx context.Context) error {
	all, err := l.Store.List()
	if err != nil {
		return err
	}
	newest := map[int]effect.Intent{}
	for _, in := range all {
		if in.Kind == effect.LedgerComment && in.Seq >= newest[in.Issue].Seq {
			newest[in.Issue] = in
		}
	}
	ex := &effect.Executor{Store: l.Store, Send: sender{l.Remote, l}, Lookup: lookup{l.Remote}, Timeouts: l.Timeouts, Now: l.Now}
	var first error
	for _, in := range all {
		if in.Kind != effect.LedgerComment {
			continue
		}
		if newest[in.Issue].Key != in.Key {
			if in.Status != effect.StatusConfirmed && in.Status != effect.StatusSuperseded {
				in.Status = effect.StatusSuperseded
				if err := l.Store.Put(in); err != nil && first == nil {
					first = err
				}
			}
			continue
		}
		got, err := ex.Execute(ctx, effect.Request{Kind: in.Kind, Issue: in.Issue, Target: in.Target, Digest: in.Digest, Payload: in.Payload})
		if first == nil {
			switch {
			case err != nil:
				first = err
			case got.Status != effect.StatusConfirmed:
				first = fmt.Errorf("ctxledger: issue %d comment did not land (%s): %s", got.Issue, got.Status, got.LastError)
			}
		}
	}
	for _, in := range newest {
		l.mu.Lock()
		want := l.recheck[in.Issue]
		l.mu.Unlock()
		if !want || in.Status != effect.StatusConfirmed {
			continue
		}
		drift, derr := l.drifted(ctx, in)
		if derr != nil {
			if first == nil {
				first = fmt.Errorf("ctxledger: issue %d: could not check the comment for edits: %w", in.Issue, derr)
			}
			continue // stays marked; the next Flush checks again
		}
		l.markRecheck(in.Issue, false)
		if !drift {
			continue
		}
		in.Status, in.Attempts, in.ProviderID = effect.StatusIntended, 0, ""
		if err := l.Store.Put(in); err != nil && first == nil {
			first = err
			continue
		}
		got, err := ex.Execute(ctx, effect.Request{Kind: in.Kind, Issue: in.Issue, Target: in.Target, Digest: in.Digest, Payload: in.Payload})
		if first == nil {
			switch {
			case err != nil:
				first = err
			case got.Status != effect.StatusConfirmed:
				first = fmt.Errorf("ctxledger: issue %d comment did not land (%s): %s", got.Issue, got.Status, got.LastError)
			}
		}
	}
	return first
}

// drifted reports whether the remote comment no longer holds the intent's text. With a Fetcher and a
// remembered id it is one conditional read (a 304 means unchanged); otherwise it is the marker lookup.
// When drift is found the remembered hash is cleared so the sender writes to the known id at once.
func (l *Ledger) drifted(ctx context.Context, in effect.Intent) (bool, error) {
	c, have := l.cacheGet(in.Issue)
	if f, ok := l.Remote.(Fetcher); ok && have {
		cm, tag, notModified, err := f.Fetch(ctx, in.Issue, c.id, c.etag)
		if errors.Is(err, ErrNoConditionalRead) {
			goto lookup
		}
		switch {
		case err != nil && isGone(err):
			l.cacheDrop(in.Issue) // deleted: the sender falls back to the marker lookup and recreates it
			return true, nil
		case err != nil:
			return false, l.classify(err)
		case notModified:
			return false, nil
		}
		c.etag = tag
		if cm.Body == in.Payload {
			l.cachePut(in.Issue, c)
			return false, nil
		}
		c.hash = ""
		l.cachePut(in.Issue, c)
		return true, nil
	}
lookup:
	cm, found, err := l.Remote.Find(ctx, in.Issue, Marker(in.Issue))
	if err != nil {
		return false, l.classify(err)
	}
	if !found {
		l.cacheDrop(in.Issue)
		return true, nil
	}
	if cm.Body == in.Payload {
		return false, nil
	}
	l.cachePut(in.Issue, cached{id: cm.ID})
	return true, nil
}

type sender struct {
	r Remote
	l *Ledger
}

func (s sender) Send(ctx context.Context, in effect.Intent) (string, error) {
	if personalData(in.Payload) {
		return "", fmt.Errorf("%w: body holds an email address or a home path", effect.ErrNotSent)
	}
	if _, changed := monitor.Mask(in.Payload); changed {
		return "", fmt.Errorf("%w: body still holds a secret", effect.ErrNotSent)
	}
	id, err := s.send(ctx, in)
	if err != nil {
		s.l.cacheDrop(in.Issue) // outcome unknown: do not trust a remembered id or hash
		return "", err
	}
	s.l.cachePut(in.Issue, cached{id: id, hash: digest(in.Payload)}) // a new write: the old validator no longer describes the body
	return strconv.FormatInt(id, 10), nil
}

func (s sender) send(ctx context.Context, in effect.Intent) (int64, error) {
	if c, ok := s.l.cacheGet(in.Issue); ok {
		if c.hash == digest(in.Payload) {
			return c.id, nil // already on the remote: send nothing
		}
		err := s.r.Update(ctx, in.Issue, c.id, in.Payload)
		if err == nil {
			return c.id, nil
		}
		if !isGone(err) {
			return 0, s.l.classify(err)
		}
		s.l.cacheDrop(in.Issue) // the comment is gone: fall back to the marker lookup
	}
	c, found, err := s.r.Find(ctx, in.Issue, Marker(in.Issue))
	if err != nil {
		return 0, s.l.classify(err)
	}
	switch {
	case found && c.Body == in.Payload:
		return c.ID, nil
	case found:
		if err := s.r.Update(ctx, in.Issue, c.ID, in.Payload); err != nil {
			return 0, s.l.classify(err) // the lookup just found the comment: a refusal now is permission, not "gone"
		}
		return c.ID, nil
	}
	nc, err := s.r.Create(ctx, in.Issue, in.Payload)
	if err != nil {
		return 0, s.l.classify(err)
	}
	return nc.ID, nil
}

// isGone reports a missing comment from a typed signal only: the documented ErrGone or an HTTP 404
// status error. Error wording never decides it ("repository not found" is a permission problem, and a
// 404 on a private repository is indistinguishable from a deleted comment — which is why a 404 is only
// ever followed by the marker lookup, and a lookup that is refused parks the intent).
func isGone(err error) bool {
	if errors.Is(err, ErrGone) {
		return true
	}
	var se *forge.StatusError
	return errors.As(err, &se) && se.Status == 404
}

// classify maps a typed failure of Find or Create to what the effect executor understands. A
// permission or missing-repository answer parks the intent; a rate limit waits for NotBefore.
func (l *Ledger) classify(err error) error {
	now := time.Now()
	if l.Now != nil {
		now = l.Now()
	}
	var se *forge.StatusError
	if !errors.As(err, &se) {
		return err
	}
	if l.Backoff.Validate() == nil {
		if v, ok := forge.ClassifyError(err, now, l.Backoff, 0); ok {
			if mapped := forge.EffectError(v); mapped != nil {
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

type lookup struct{ r Remote }

func (l lookup) Lookup(ctx context.Context, in effect.Intent) (effect.LookupResult, error) {
	c, found, err := l.r.Find(ctx, in.Issue, Marker(in.Issue))
	if err != nil {
		return effect.LookupResult{}, err
	}
	if found && c.Body == in.Payload {
		return effect.LookupResult{Found: true, ProviderID: strconv.FormatInt(c.ID, 10), Authoritative: true}, nil
	}
	return effect.LookupResult{Authoritative: true}, nil // the query worked; this text is not there
}

// Runner runs gh and returns stdout.
type Runner func(ctx context.Context, args ...string) ([]byte, error)

// GHRemote implements Remote with gh. It issues only comment list/create/update calls.
type GHRemote struct {
	Repo string
	Run  Runner
	// Enforcer decides whether a comment may be created or edited. Both need the authenticated user's id
	// and the Issue author's id (and, for an edit, the comment author's id: only a comment we wrote may be
	// edited). Without one, the call is refused (identity-unknown) before gh runs: the remote fails closed
	// rather than writing where it cannot attribute the write.
	Enforcer *effect.Enforcer
}

func (g GHRemote) Find(ctx context.Context, issue int, marker string) (Comment, bool, error) {
	if err := effect.CheckRequest("GET", fmt.Sprintf("/repos/%s/issues/%d/comments", g.Repo, issue), nil); err != nil {
		return Comment{}, false, err
	}
	out, err := g.Run(ctx, "api", fmt.Sprintf("repos/%s/issues/%d/comments", g.Repo, issue), "--paginate", "--jq", ".[] | {id: .id, body: .body}")
	if err != nil {
		return Comment{}, false, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		var c struct {
			ID   int64  `json:"id"`
			Body string `json:"body"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(line)), &c) == nil && strings.Contains(c.Body, marker) {
			return Comment{ID: c.ID, Body: c.Body}, true, nil
		}
	}
	return Comment{}, false, nil
}

func (g GHRemote) Create(ctx context.Context, issue int, body string) (Comment, error) {
	if err := g.Enforcer.Check(ctx, "POST", fmt.Sprintf("/repos/%s/issues/%d/comments", g.Repo, issue), nil, nil); err != nil {
		return Comment{}, err
	}
	out, err := g.Run(ctx, "api", "-X", "POST", fmt.Sprintf("repos/%s/issues/%d/comments", g.Repo, issue), "-f", "body="+body)
	if err != nil {
		return Comment{}, err
	}
	var c struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(out, &c); err != nil {
		return Comment{}, fmt.Errorf("ctxledger: create response: %w", err)
	}
	return Comment{ID: c.ID, Body: body}, nil
}

func (g GHRemote) Update(ctx context.Context, issue int, id int64, body string) error {
	if id <= 0 {
		return errors.New("ctxledger: update needs a comment id")
	}
	if err := g.Enforcer.Check(ctx, "PATCH", fmt.Sprintf("/repos/%s/issues/comments/%d", g.Repo, id), nil, nil); err != nil {
		return err
	}
	_, err := g.Run(ctx, "api", "-X", "PATCH", fmt.Sprintf("repos/%s/issues/comments/%d", g.Repo, id), "-f", "body="+body)
	return err // a Runner reports HTTP failures as *forge.StatusError; isGone reads the status
}

// ForgeFetcher implements Fetcher on a forge.Client, which keeps the ETag cache and the shared
// budget. A call the budget defers is reported as a LimitedError so the caller waits instead of
// writing. Whether a 304 counts against the primary limit follows the host's documentation
// (https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api states it
// does not count for authorized requests); that statement is taken from the doc summary and was not
// measured here.
type ForgeFetcher struct {
	Client *forge.Client
	Repo   string
}

func (f ForgeFetcher) Fetch(ctx context.Context, issue int, id int64, _ string) (Comment, string, bool, error) {
	res, err := f.Client.Get(ctx, fmt.Sprintf("/repos/%s/issues/comments/%d", f.Repo, id))
	switch {
	case err != nil:
		return Comment{}, "", false, err
	case res.Deferred:
		return Comment{}, "", false, &effect.LimitedError{RetryAt: res.NotBefore}
	case res.FromCache:
		return Comment{}, "", true, nil
	}
	var c struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal(res.Body, &c); err != nil {
		return Comment{}, "", false, fmt.Errorf("ctxledger: comment response: %w", err)
	}
	return Comment{ID: c.ID, Body: c.Body}, "", false, nil
}
