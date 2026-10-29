// Package ghidentity establishes numeric GitHub user IDs for the Issue-author policy using read-only GETs
// through a forge.Client (shared budget, ETag cache). It never sends a write. IDs are the only thing
// compared: a response that carries a login but no numeric id is "unknown", and every failure (transport,
// limit, denial, parse) is reported as an error, which the enforcer reads as identity-unknown.
package ghidentity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// GitHub implements effect.Identity.
type GitHub struct {
	Client *forge.Client
	// MaxAge is how long a looked-up ID may be reused before it is fetched again. It comes from
	// configuration; zero means every call fetches again.
	MaxAge time.Duration
	Now    func() time.Time

	mu    sync.Mutex
	cache map[string]entry
}

type entry struct {
	id int64
	at time.Time
}

var errUnknown = errors.New("ghidentity: identity could not be established")

func (g *GitHub) now() time.Time {
	if g.Now != nil {
		return g.Now()
	}
	return time.Now()
}

func (g *GitHub) fetch(ctx context.Context, key, path string, pick func([]byte) (int64, bool)) (int64, error) {
	now := g.now()
	g.mu.Lock()
	if e, ok := g.cache[key]; ok && g.MaxAge > 0 && now.Sub(e.at) <= g.MaxAge {
		g.mu.Unlock()
		return e.id, nil
	}
	g.mu.Unlock()

	res, err := g.Client.Get(ctx, path)
	switch {
	case err != nil:
		return 0, fmt.Errorf("%w: %v", errUnknown, err)
	case res.Deferred:
		return 0, fmt.Errorf("%w: lookup deferred until %s", errUnknown, res.NotBefore.UTC().Format(time.RFC3339))
	}
	id, ok := pick(res.Body)
	if !ok || id <= 0 {
		return 0, fmt.Errorf("%w: no numeric id in the response", errUnknown)
	}
	g.mu.Lock()
	if g.cache == nil {
		g.cache = map[string]entry{}
	}
	g.cache[key] = entry{id: id, at: now}
	g.mu.Unlock()
	return id, nil
}

func topID(b []byte) (int64, bool) {
	var v struct {
		ID *int64 `json:"id"`
	}
	if json.Unmarshal(b, &v) != nil || v.ID == nil {
		return 0, false
	}
	return *v.ID, true
}

func userID(b []byte) (int64, bool) {
	var v struct {
		User *struct {
			ID *int64 `json:"id"`
		} `json:"user"`
	}
	if json.Unmarshal(b, &v) != nil || v.User == nil || v.User.ID == nil {
		return 0, false
	}
	return *v.User.ID, true
}

// CurrentUserID is the numeric id of the authenticated user (GET /user).
func (g *GitHub) CurrentUserID(ctx context.Context) (int64, error) {
	return g.fetch(ctx, "user", "/user", topID)
}

// IssueAuthorID is the numeric id of the user who opened the Issue.
func (g *GitHub) IssueAuthorID(ctx context.Context, repo string, issue int) (int64, error) {
	return g.fetch(ctx, fmt.Sprintf("issue %s#%d", repo, issue), fmt.Sprintf("/repos/%s/issues/%d", repo, issue), userID)
}

// CommentAuthorID is the numeric id of the user who wrote the comment.
func (g *GitHub) CommentAuthorID(ctx context.Context, repo string, comment int64) (int64, error) {
	return g.fetch(ctx, fmt.Sprintf("comment %s/%d", repo, comment), fmt.Sprintf("/repos/%s/issues/comments/%d", repo, comment), userID)
}
