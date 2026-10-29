package ghintake

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/intake"
)

// Publisher is the gh-backed intake.Publisher. Creating an Issue is the only write, it is off unless the
// embedded GH has AllowWrite, and it goes through the protection guard first. A request is correlated with
// its Issue by RequestMarker(key) in the body; a lost answer is therefore found again with FindIssue and
// never created twice. Every other call is a read-only GET/list.
type Publisher struct {
	GH GH
	// Identity answers AuthenticatedUser and the author ids of duplicate candidates (numeric ids only).
	Identity effect.Identity
	// Content supplies the title and body of the Issue for a request: the claim record keeps only a hash of
	// the payload, so the text comes from wherever the caller keeps it.
	Content func(ctx context.Context, req claim.Request) (title, body string, err error)
	// Limit is the most duplicate candidates a search returns. It comes from configuration; zero is refused.
	Limit int
}

var _ intake.Publisher = (*Publisher)(nil)

// RequestMarker is the correlation marker a request leaves in its Issue body.
func RequestMarker(key string) string { return "<!-- agentic-factory:request=" + key + " -->" }

type listed struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	URL    string `json:"url"`
}

// notSent wraps err with intake.ErrNotSent when the failure proves nothing was created: a refusal by the
// guard, writes being off, or an HTTP client error (the host rejected the request). A server error, a
// transport error and a lost answer prove nothing, so they are returned unwrapped (an unknown effect).
func notSent(err error) error {
	if errors.Is(err, effect.ErrRefused) {
		return fmt.Errorf("%w: %w", intake.ErrNotSent, err)
	}
	var se *forge.StatusError
	if errors.As(err, &se) && se.Status >= 400 && se.Status < 500 {
		return fmt.Errorf("%w: %w", intake.ErrNotSent, err)
	}
	return err
}

func (p *Publisher) PublishIssue(ctx context.Context, req claim.Request) (int, string, error) {
	if !p.GH.AllowWrite {
		return 0, "", fmt.Errorf("%w: gh writes disabled (external write needs user approval)", intake.ErrNotSent)
	}
	if err := effect.CheckRequest("POST", "/repos/"+p.GH.Repo+"/issues", nil); err != nil {
		return 0, "", notSent(err)
	}
	title, body, err := p.Content(ctx, req)
	if err != nil || strings.TrimSpace(title) == "" {
		return 0, "", fmt.Errorf("%w: no title or body for the request", intake.ErrNotSent)
	}
	out, err := p.GH.Run(ctx, "issue", "create", "--repo", p.GH.Repo, "--title", title, "--body", body+"\n\n"+RequestMarker(req.Key))
	if err != nil {
		return 0, "", notSent(err)
	}
	ref := strings.TrimSpace(string(out))
	n, convErr := strconv.Atoi(issueNumber(ref))
	if convErr != nil || n <= 0 {
		return 0, ref, fmt.Errorf("ghintake: the create answer did not name an Issue: %q", ref) // answered but unreadable: unknown effect
	}
	return n, ref, nil
}

// search lists Issues matching term. truncated is true when the list reached the configured limit, so
// that an Issue absent from it proves nothing.
func (p *Publisher) search(ctx context.Context, term string) (items []listed, truncated bool, err error) {
	limit, err := p.GH.listLimit()
	if err != nil {
		return nil, false, err
	}
	out, err := p.GH.read(ctx, "issue", "list", "--repo", p.GH.Repo, "--state", "all", "--search", term, "--json", "number,title,body,url", "--limit", limit)
	if err != nil {
		return nil, false, err
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, false, fmt.Errorf("ghintake: issue list output: %w", err)
	}
	return items, len(items) >= p.GH.ListLimit, nil
}

func (p *Publisher) FindIssue(ctx context.Context, req claim.Request) (int, string, bool, error) {
	items, truncated, err := p.search(ctx, req.Key+" in:body")
	if err != nil {
		return 0, "", false, err
	}
	for _, it := range items {
		if strings.Contains(it.Body, RequestMarker(req.Key)) {
			return it.Number, it.URL, true, nil
		}
	}
	if truncated {
		return 0, "", false, errCutShort
	}
	return 0, "", false, nil
}

// AuthenticatedUser is the numeric id of the authenticated user, or 0 and an error when it cannot be established.
func (p *Publisher) AuthenticatedUser(ctx context.Context) (int64, error) {
	if p.Identity == nil {
		return 0, errors.New("ghintake: no identity lookup wired")
	}
	id, err := p.Identity.CurrentUserID(ctx)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("ghintake: authenticated user unknown: %v", err)
	}
	return id, nil
}

// FindDuplicates lists Issues with the same title. It only reads. A candidate's author id comes from the
// identity lookup (a numeric id); when that cannot be established the candidate is still reported, with
// AuthorID 0 meaning "unknown".
func (p *Publisher) FindDuplicates(ctx context.Context, req claim.Request) ([]intake.Candidate, error) {
	if p.Limit <= 0 {
		return nil, errors.New("ghintake: the duplicate candidate limit must come from configuration")
	}
	title, _, err := p.Content(ctx, req)
	if err != nil || strings.TrimSpace(title) == "" {
		return nil, errors.New("ghintake: no title to search duplicates for")
	}
	items, truncated, err := p.search(ctx, fmt.Sprintf("%q in:title", title))
	if err != nil {
		return nil, err
	}
	if truncated { // a duplicate may be missing from a cut-short list, and a missed duplicate is a second Issue
		return nil, errCutShort
	}
	var out []intake.Candidate
	for _, it := range items {
		if it.Title != title {
			continue
		}
		c := intake.Candidate{Issue: it.Number, Ref: it.URL}
		if p.Identity != nil {
			if f, err := p.Identity.IssueFacts(ctx, p.GH.Repo, it.Number); err == nil && f.AuthorID > 0 {
				c.AuthorID = f.AuthorID
			}
		}
		out = append(out, c)
		if len(out) >= p.Limit {
			break
		}
	}
	return out, nil
}
