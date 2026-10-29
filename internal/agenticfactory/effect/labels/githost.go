package labels

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// GitHubHost is the GitHub implementation of Host over a forge.Transport. Writes are off by default,
// the only requests it can build are list/create/update on this repository's labels (no delete, no
// other endpoint), and every call goes through the shared Budget: a blocked budget returns a
// LimitedError without sending anything, and failures come back as *forge.StatusError so the effect
// layer decides by status code (limit waits, permission parks).
type GitHubHost struct {
	Repo       string
	T          forge.Transport
	B          *forge.Budget
	AllowWrite bool
	PerPage    int // page size for listing; required, comes from configuration
	Backoff    forge.Backoff
	Now        func() time.Time
}

func (h *GitHubHost) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

func (h *GitHubHost) base() string { return "/repos/" + h.Repo + "/labels" }

// allowed is the request allowlist.
func (h *GitHubHost) allowed(r forge.Request) bool {
	switch r.Method {
	case http.MethodGet:
		return r.Path == h.base() || strings.HasPrefix(r.Path, h.base()+"?")
	case http.MethodPost:
		return r.Path == h.base()
	case http.MethodPatch:
		return strings.HasPrefix(r.Path, h.base()+"/") && !strings.Contains(strings.TrimPrefix(r.Path, h.base()+"/"), "/")
	}
	return false
}

func (h *GitHubHost) do(ctx context.Context, r forge.Request) (forge.Response, error) {
	if !h.allowed(r) {
		return forge.Response{}, fmt.Errorf("labels: refusing request %s %s", r.Method, r.Path)
	}
	now := h.now()
	if ok, until := h.B.Allow(now); !ok {
		return forge.Response{}, &effect.LimitedError{RetryAt: until}
	}
	resp, err := h.T.Do(ctx, r)
	if err != nil {
		return forge.Response{}, err
	}
	h.B.Observe(resp, now)
	switch v := forge.Classify(resp, now, h.Backoff, 0); v.Kind {
	case forge.KindOK:
		return resp, nil
	case forge.KindLimited:
		h.B.Block(v.RetryAt)
	}
	return forge.Response{}, &forge.StatusError{Status: resp.Status, Header: resp.Header, Body: resp.Body}
}

type ghLabel struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

func (h *GitHubHost) List(ctx context.Context) ([]Label, error) {
	if h.PerPage <= 0 {
		return nil, errors.New("labels: GitHubHost.PerPage must come from configuration")
	}
	var out []Label
	for page := 1; ; page++ {
		resp, err := h.do(ctx, forge.Request{Method: http.MethodGet, Path: fmt.Sprintf("%s?per_page=%d&page=%d", h.base(), h.PerPage, page)})
		if err != nil {
			return nil, err
		}
		var got []ghLabel
		if err := json.Unmarshal(resp.Body, &got); err != nil {
			return nil, fmt.Errorf("labels: list response: %w", err)
		}
		for _, l := range got {
			out = append(out, Label{Name: l.Name, Color: strings.ToLower(l.Color), Description: l.Description})
		}
		if len(got) < h.PerPage {
			return out, nil
		}
	}
}

func (h *GitHubHost) Create(ctx context.Context, l Label) error {
	if !h.AllowWrite {
		return ErrWriteDisabled
	}
	body, _ := json.Marshal(ghLabel{Name: l.Name, Color: l.Color, Description: l.Description})
	_, err := h.do(ctx, forge.Request{Method: http.MethodPost, Path: h.base(), Body: body})
	return err
}

func (h *GitHubHost) Update(ctx context.Context, name string, l Label) error {
	if !h.AllowWrite {
		return ErrWriteDisabled
	}
	body, _ := json.Marshal(struct {
		NewName     string `json:"new_name"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}{l.Name, l.Color, l.Description})
	_, err := h.do(ctx, forge.Request{Method: http.MethodPatch, Path: h.base() + "/" + url.PathEscape(name), Body: body})
	return err
}
