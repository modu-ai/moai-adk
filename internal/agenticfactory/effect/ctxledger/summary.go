package ctxledger

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// The checkpoint-recovery summary link: the fourth and last step. After the commit, the remote ACK and
// the manifest are remote-durable, one comment on the Issue points a reader at the manifest, pinned to the
// SHA the remote confirmed. The step posts through the same ledger comment as everything else (one comment
// per Issue, updated in place, marker lookup on a lost response, masking and personal-data checks), so a
// retry with the same checkpoint is idempotent and makes no second comment. It pushes nothing and holds no
// new permission.
//
// A failed comment fails only this step. PublishSummaryLink reports it by value (StepFailed) and returns no
// error, so a caller cannot mistake it for a failed recovery; whether the recovery succeeded is decided
// by the earlier steps alone.

// Durable is one earlier step as the remote confirmed it.
type Durable struct {
	SHA            string // full 40-hex SHA the remote reported for the step
	RemoteVerified bool   // a lookup on the remote found it; a local record alone is not enough
}

// Manifest is the recovery manifest: where it lives, its content hash, and its own remote confirmation.
type Manifest struct {
	Path    string // repository-relative path of the manifest
	Hash    string // content hash recorded in the checkpoint
	Durable Durable
}

// Checkpoint is what the recovery has made durable so far.
type Checkpoint struct {
	Issue    int
	Commit   Durable
	ACK      Durable // the remote's acknowledgement of the control-ref write
	Manifest Manifest
	// LinkTemplate turns a pinned SHA and the manifest path into the link that is posted. It must be https,
	// carry no credentials or query, and contain the {sha} placeholder (and {path}). The host layout is the
	// caller's: GitHub and GitLab differ.
	LinkTemplate string
}

// StepStatus is the outcome of the summary step.
type StepStatus string

const (
	StepPublished StepStatus = "published" // the comment is on the Issue (new, or unchanged and verified)
	StepSkipped   StepStatus = "skipped"   // an earlier step is not remote-durable yet, or the input is unusable; nothing was sent
	StepFailed    StepStatus = "failed"    // the comment did not land; Err says why; this step only
)

// StepResult is the step's outcome. Err is set only with StepFailed or an unusable input.
type StepResult struct {
	Status StepStatus
	Reason string
	Link   string
	Err    error
}

var fullSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

func notReady(reason string) StepResult { return StepResult{Status: StepSkipped, Reason: reason} }

// pinned returns the one SHA every step agrees on, or why the checkpoint is not ready.
func (cp Checkpoint) pinned() (string, string) {
	steps := []struct {
		name string
		d    Durable
	}{{"commit", cp.Commit}, {"remote ACK", cp.ACK}, {"manifest", cp.Manifest.Durable}}
	for _, s := range steps {
		switch {
		case !s.d.RemoteVerified:
			return "", s.name + " is not confirmed on the remote"
		case !fullSHA.MatchString(s.d.SHA):
			return "", s.name + " has no full lower-case SHA"
		case s.d.SHA != cp.Commit.SHA:
			return "", s.name + " is confirmed at a different SHA than the commit"
		}
	}
	if strings.TrimSpace(cp.Manifest.Hash) == "" {
		return "", "the manifest has no recorded hash"
	}
	return cp.Commit.SHA, ""
}

// link builds the pinned link, refusing a template or path that could carry a credential, a person's
// details, or a path outside the repository.
func (cp Checkpoint) link(sha string) (string, error) {
	p := cp.Manifest.Path
	if p == "" || strings.HasPrefix(p, "/") || strings.ContainsAny(p, "\x00\r\n\t\\?#") {
		return "", errors.New("the manifest path is empty, absolute, or has unsafe characters")
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", errors.New("the manifest path has dot or empty segments")
		}
	}
	if personalData(p) {
		return "", errors.New("the manifest path holds an email address or a home path")
	}
	u, err := url.Parse(cp.LinkTemplate)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", errors.New("the link template must be a plain https URL without credentials, query or fragment")
	}
	if !strings.Contains(cp.LinkTemplate, "{sha}") {
		return "", errors.New("the link template has no {sha} placeholder: a link must be pinned")
	}
	out := strings.NewReplacer("{sha}", sha, "{path}", p).Replace(cp.LinkTemplate)
	if _, changed := monitor.Mask(out); changed {
		return "", errors.New("the link would carry a secret")
	}
	return out, nil
}

// PublishSummaryLink posts (or confirms) the summary comment for a checkpoint. It makes no remote call
// unless every earlier step is remote-durable and the link and text pass the checks.
func (l *Ledger) PublishSummaryLink(ctx context.Context, cp Checkpoint) StepResult {
	if cp.Issue <= 0 {
		return StepResult{Status: StepSkipped, Reason: "no Issue number", Err: errors.New("ctxledger: checkpoint without an Issue")}
	}
	sha, why := cp.pinned()
	if why != "" {
		return notReady(why)
	}
	link, err := cp.link(sha)
	if err != nil {
		return StepResult{Status: StepSkipped, Reason: "unusable link input", Err: err}
	}
	text := fmt.Sprintf("Checkpoint recovered (remote-confirmed %s)\nmanifest: %s\nmanifest hash: %s\n", sha, link, strings.TrimSpace(cp.Manifest.Hash))
	if _, changed := monitor.Mask(text); changed || personalData(text) {
		return StepResult{Status: StepSkipped, Reason: "the summary would carry a secret or personal data", Err: ErrNotMasked}
	}
	if _, err := l.Submit(ctx, Body{Issue: cp.Issue, Text: text, MaskPassed: true}); err != nil {
		return StepResult{Status: StepFailed, Reason: "the summary could not be recorded", Err: err}
	}
	if err := l.Flush(ctx); err != nil {
		return StepResult{Status: StepFailed, Reason: "the comment did not land", Err: err}
	}
	return StepResult{Status: StepPublished, Link: link}
}

// Calls is a tally of remote calls by kind.
type Calls struct{ Find, Create, Update, Fetch int }

// Total is the sum of all calls.
func (c Calls) Total() int { return c.Find + c.Create + c.Update + c.Fetch }

// CountingRemote wraps a Remote and counts every call it forwards, so a run can report how many API calls
// it made. It adds no behaviour. If the inner Remote also reads conditionally (Fetcher), those reads are
// counted as well.
type CountingRemote struct {
	Inner Remote

	mu sync.Mutex
	n  Calls
}

func (c *CountingRemote) bump(f func(*Calls)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	f(&c.n)
}

func (c *CountingRemote) Find(ctx context.Context, issue int, marker string) (Comment, bool, error) {
	c.bump(func(n *Calls) { n.Find++ })
	return c.Inner.Find(ctx, issue, marker)
}

func (c *CountingRemote) Create(ctx context.Context, issue int, body string) (Comment, error) {
	c.bump(func(n *Calls) { n.Create++ })
	return c.Inner.Create(ctx, issue, body)
}

func (c *CountingRemote) Update(ctx context.Context, issue int, id int64, body string) error {
	c.bump(func(n *Calls) { n.Update++ })
	return c.Inner.Update(ctx, issue, id, body)
}

// Fetch forwards a conditional read when the inner Remote has one, and says so by returning an error when it does not.
func (c *CountingRemote) Fetch(ctx context.Context, issue int, id int64, etag string) (Comment, string, bool, error) {
	f, ok := c.Inner.(Fetcher)
	if !ok {
		return Comment{}, "", false, ErrNoConditionalRead
	}
	c.bump(func(n *Calls) { n.Fetch++ })
	return f.Fetch(ctx, issue, id, etag)
}

// Snapshot returns the counts so far; Reset zeroes them.
func (c *CountingRemote) Snapshot() Calls {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func (c *CountingRemote) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n = Calls{}
}
