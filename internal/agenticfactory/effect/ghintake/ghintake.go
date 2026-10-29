// Package ghintake is the gh/git face of the intake effects: creating an Issue and finding it
// again by the key marker in its body, and reading branch and worktree facts.
//
// Writes are off by default: with AllowWrite false the sender makes no call at all. Even when on,
// the only write it can issue is `gh issue create`; every lookup passes a read-only allowlist.
// Tests use a fake runner only.
package ghintake

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// Runner runs gh (for GH) or git (for Git) with the given arguments and returns stdout.
type Runner func(ctx context.Context, args ...string) ([]byte, error)

// issuePayload is what an IssueCreate request carries in Payload.
type issuePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// IssueRequest builds the effect request for creating one Issue. The same repo, title and body
// always produce the same key.
func IssueRequest(issue int, repo, title, body string) effect.Request {
	raw, _ := json.Marshal(issuePayload{Title: title, Body: body})
	sum := sha256.Sum256(raw)
	return effect.Request{Kind: effect.IssueCreate, Issue: issue, Target: "repo:" + repo, Digest: hex.EncodeToString(sum[:]), Payload: string(raw)}
}

// GH sends and finds Issues.
type GH struct {
	Repo       string
	Run        Runner
	AllowWrite bool // default false: the sender then makes no call
}

// Send creates the Issue, with the idempotency marker appended to its body.
func (g GH) Send(ctx context.Context, in effect.Intent) (string, error) {
	if in.Kind != effect.IssueCreate {
		return "", fmt.Errorf("%w: ghintake sends only issue-create, not %s", effect.ErrNotSent, in.Kind)
	}
	if !g.AllowWrite {
		return "", fmt.Errorf("%w: gh writes disabled (external write needs user approval)", effect.ErrNotSent)
	}
	var p issuePayload
	if err := json.Unmarshal([]byte(in.Payload), &p); err != nil || p.Title == "" {
		return "", fmt.Errorf("%w: issue payload invalid", effect.ErrNotSent)
	}
	body := p.Body + "\n\n" + effect.Marker(in.Key)
	out, err := g.Run(ctx, "issue", "create", "--repo", g.Repo, "--title", p.Title, "--body", body)
	if err != nil {
		return "", err
	}
	return issueNumber(string(out)), nil
}

var numRe = regexp.MustCompile(`/issues/(\d+)\s*$`)

func issueNumber(out string) string {
	if m := numRe.FindStringSubmatch(strings.TrimSpace(out)); m != nil {
		return m[1]
	}
	return strings.TrimSpace(out)
}

// readOnly admits only the fact queries this package makes.
func readOnly(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch args[0] + " " + args[1] {
	case "issue list", "issue view", "worktree list":
		return true
	case "ls-remote --heads":
		for _, a := range args {
			if strings.HasPrefix(a, "--upload-pack") || strings.HasPrefix(a, "--receive-pack") || a == "-o" {
				return false
			}
		}
		return true
	}
	return false
}

func (g GH) read(ctx context.Context, args ...string) ([]byte, error) {
	if !readOnly(args) {
		return nil, fmt.Errorf("ghintake: refusing non-read-only call %v", args)
	}
	return g.Run(ctx, args...)
}

// Git reads branch and worktree facts.
type Git struct{ Run Runner }

func (g Git) read(ctx context.Context, args ...string) ([]byte, error) {
	if !readOnly(args) {
		return nil, fmt.Errorf("ghintake: refusing non-read-only call %v", args)
	}
	return g.Run(ctx, args...)
}

// Lookup answers "is this intent's effect already there?" for the intake kinds.
type Lookup struct {
	GH  GH
	Git Git
}

func (l Lookup) Lookup(ctx context.Context, in effect.Intent) (effect.LookupResult, error) {
	switch in.Kind {
	case effect.IssueCreate:
		return l.issue(ctx, in)
	case effect.BranchPrepare:
		return l.branch(ctx, in)
	case effect.WorktreePrepare:
		return l.worktree(ctx, in)
	}
	return effect.LookupResult{}, fmt.Errorf("ghintake: no lookup for %s", in.Kind)
}

func (l Lookup) issue(ctx context.Context, in effect.Intent) (effect.LookupResult, error) {
	out, err := l.GH.read(ctx, "issue", "list", "--repo", l.GH.Repo, "--state", "all", "--search", in.Key+" in:body", "--json", "number,body", "--limit", "100")
	if err != nil {
		return effect.LookupResult{}, err
	}
	var items []struct {
		Number int    `json:"number"`
		Body   string `json:"body"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return effect.LookupResult{}, fmt.Errorf("ghintake: issue list output: %w", err)
	}
	for _, it := range items {
		if effect.KeyFromMarker(it.Body) == in.Key {
			return effect.LookupResult{Found: true, ProviderID: strconv.Itoa(it.Number), Authoritative: true}, nil
		}
	}
	return effect.LookupResult{Authoritative: true}, nil
}

func (l Lookup) branch(ctx context.Context, in effect.Intent) (effect.LookupResult, error) {
	out, err := l.Git.read(ctx, "ls-remote", "--heads", "origin", in.Target)
	if err != nil {
		return effect.LookupResult{}, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "refs/heads/"+in.Target) {
			return effect.LookupResult{Found: true, ProviderID: in.Target, Authoritative: true}, nil
		}
	}
	return effect.LookupResult{Authoritative: true}, nil
}

func (l Lookup) worktree(ctx context.Context, in effect.Intent) (effect.LookupResult, error) {
	out, err := l.Git.read(ctx, "worktree", "list", "--porcelain")
	if err != nil {
		return effect.LookupResult{}, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "worktree "+in.Target {
			return effect.LookupResult{Found: true, ProviderID: in.Target, Authoritative: true}, nil
		}
	}
	return effect.LookupResult{Authoritative: true}, nil
}
