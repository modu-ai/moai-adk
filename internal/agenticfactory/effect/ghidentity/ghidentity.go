// Package ghidentity establishes the facts the Issue-author policy needs — the authenticated user's numeric
// id, an Issue's or comment's author id, and the repository the host really answered for — using read-only
// GETs through a forge.Client (shared budget, ETag cache). It never sends a write. IDs are the only thing
// compared: a response that carries a login but no numeric id is "unknown", and every failure (transport,
// limit, denial, parse) is an error, which the enforcer reads as identity-unknown.
package ghidentity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// GitHub implements effect.Identity.
type GitHub struct {
	Client *forge.Client
	// MaxAge is how long a looked-up fact may be reused before it is fetched again. It comes from
	// configuration; zero means every call fetches again.
	MaxAge time.Duration
	Now    func() time.Time

	mu    sync.Mutex
	cache map[string]entry
}

type entry struct {
	v  any
	at time.Time
}

var errUnknown = errors.New("ghidentity: identity could not be established")

func (g *GitHub) now() time.Time {
	if g.Now != nil {
		return g.Now()
	}
	return time.Now()
}

func (g *GitHub) cached(key string, now time.Time) (any, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if e, ok := g.cache[key]; ok && g.MaxAge > 0 && now.Sub(e.at) <= g.MaxAge {
		return e.v, true
	}
	return nil, false
}

func (g *GitHub) store(key string, v any, now time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.cache == nil {
		g.cache = map[string]entry{}
	}
	g.cache[key] = entry{v: v, at: now}
}

func (g *GitHub) get(ctx context.Context, path string) (forge.Result, error) {
	res, err := g.Client.Get(ctx, path)
	switch {
	case err != nil:
		return forge.Result{}, fmt.Errorf("%w: %v", errUnknown, err)
	case res.Deferred:
		return forge.Result{}, fmt.Errorf("%w: lookup deferred until %s", errUnknown, res.NotBefore.UTC().Format(time.RFC3339))
	}
	return res, nil
}

// repoAndIssue reads "owner/name" and an optional issue number out of a host API URL such as
// https://api.github.com/repos/o/r or .../repos/o/r/issues/7.
func repoAndIssue(u string) (repo string, issue int) {
	i := strings.Index(u, "/repos/")
	if i < 0 {
		return "", 0
	}
	segs := strings.Split(strings.Trim(u[i+len("/repos/"):], "/"), "/")
	if len(segs) < 2 || segs[0] == "" || segs[1] == "" {
		return "", 0
	}
	repo = segs[0] + "/" + segs[1]
	if len(segs) >= 4 && segs[2] == "issues" {
		if n, err := strconv.Atoi(segs[3]); err == nil && n > 0 {
			issue = n
		}
	}
	return repo, issue
}

// CurrentUserID is the numeric id of the authenticated user (GET /user).
func (g *GitHub) CurrentUserID(ctx context.Context) (int64, error) {
	now := g.now()
	if v, ok := g.cached("user", now); ok {
		return v.(int64), nil
	}
	res, err := g.get(ctx, "/user")
	if err != nil {
		return 0, err
	}
	var b struct {
		ID *int64 `json:"id"`
	}
	if json.Unmarshal(res.Body, &b) != nil || b.ID == nil || *b.ID <= 0 {
		return 0, fmt.Errorf("%w: no numeric id in the response", errUnknown)
	}
	g.store("user", *b.ID, now)
	return *b.ID, nil
}

type authored struct {
	User *struct {
		ID *int64 `json:"id"`
	} `json:"user"`
	RepositoryURL string `json:"repository_url"`
	IssueURL      string `json:"issue_url"`
}

func (a authored) authorID() (int64, bool) {
	if a.User == nil || a.User.ID == nil || *a.User.ID <= 0 {
		return 0, false
	}
	return *a.User.ID, true
}

// IssueFacts reads an Issue's author id and the repository the host answered for.
func (g *GitHub) IssueFacts(ctx context.Context, repo string, issue int) (effect.IssueFacts, error) {
	key, now := fmt.Sprintf("issue %s#%d", repo, issue), g.now()
	if v, ok := g.cached(key, now); ok {
		return v.(effect.IssueFacts), nil
	}
	res, err := g.get(ctx, fmt.Sprintf("/repos/%s/issues/%d", repo, issue))
	if err != nil {
		return effect.IssueFacts{}, err
	}
	var b authored
	id, ok := int64(0), false
	if json.Unmarshal(res.Body, &b) == nil {
		id, ok = b.authorID()
	}
	if !ok {
		return effect.IssueFacts{}, fmt.Errorf("%w: no numeric author id in the response", errUnknown)
	}
	actual, _ := repoAndIssue(b.RepositoryURL)
	f := effect.IssueFacts{AuthorID: id, ActualRepo: actual, Redirected: res.Redirected}
	g.store(key, f, now)
	return f, nil
}

// CommentFacts reads a comment's author id, the Issue it belongs to, and the repository the host answered for.
func (g *GitHub) CommentFacts(ctx context.Context, repo string, comment int64) (effect.CommentFacts, error) {
	key, now := fmt.Sprintf("comment %s/%d", repo, comment), g.now()
	if v, ok := g.cached(key, now); ok {
		return v.(effect.CommentFacts), nil
	}
	res, err := g.get(ctx, fmt.Sprintf("/repos/%s/issues/comments/%d", repo, comment))
	if err != nil {
		return effect.CommentFacts{}, err
	}
	var b authored
	id, ok := int64(0), false
	if json.Unmarshal(res.Body, &b) == nil {
		id, ok = b.authorID()
	}
	if !ok {
		return effect.CommentFacts{}, fmt.Errorf("%w: no numeric author id in the response", errUnknown)
	}
	actual, issue := repoAndIssue(b.IssueURL)
	f := effect.CommentFacts{AuthorID: id, Issue: issue, ActualRepo: actual, Redirected: res.Redirected}
	g.store(key, f, now)
	return f, nil
}
