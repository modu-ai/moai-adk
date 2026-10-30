// Package ghpr reads a pull request through the gh command line and turns it into the notify.PRView the
// monitor watches. It only reads: its call list is fixed to three read verbs and anything else is refused
// before a runner is reached. Tests feed it recorded gh output; no GitHub call is made by this package's
// tests.
package ghpr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/notify"
)

// Runner runs gh with the given arguments and returns its standard output. For `gh pr checks` the tool
// exits non-zero while a check is failing or pending and still prints the JSON; a runner must return
// that output together with the error (the adapter accepts an error alongside valid output for that one
// call, and only for it).
type Runner func(ctx context.Context, args ...string) ([]byte, error)

// Source implements notify.PRSource over gh.
type Source struct {
	Repo string
	Run  Runner
	// Timeout bounds each gh call on its own (zero is refused: no read runs unbounded). The caller names
	// it from its configured read limit; nothing here chooses a number.
	Timeout time.Duration
	// Number returns the pull request number recorded for the Issue (the claim record's receipt).
	Number func(issue int) (pr int, ok bool, err error)
}

var _ notify.PRSource = Source{}

// readOnly is the whole list of calls this package makes.
func readOnly(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch args[0] + " " + args[1] {
	case "pr view", "pr checks", "issue view":
		return true
	}
	return false
}

func (s Source) read(ctx context.Context, args ...string) ([]byte, error) {
	if !readOnly(args) {
		return nil, fmt.Errorf("ghpr: refusing non-read-only call %v", args)
	}
	if s.Timeout <= 0 {
		return nil, errors.New("ghpr: no time limit configured for gh reads")
	}
	cctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.Run(cctx, args...)
}

var sha40 = regexp.MustCompile(`^[0-9a-f]{40}$`)

// PR reads the pull request of an Issue; found is false when the claim record names none.
func (s Source) PR(ctx context.Context, issue int) (notify.PRView, bool, error) {
	if s.Run == nil || s.Number == nil || s.Repo == "" {
		return notify.PRView{}, false, errors.New("ghpr: Repo, Run and Number are required")
	}
	n, ok, err := s.Number(issue)
	if err != nil || !ok {
		return notify.PRView{}, false, err
	}
	num := strconv.Itoa(n)
	prJSON, err := s.read(ctx, "pr", "view", num, "--repo", s.Repo, "--json", "number,headRefOid,state,reviewDecision")
	if err != nil {
		return notify.PRView{}, false, fmt.Errorf("ghpr: pr view: %w", err)
	}
	checksJSON, cerr := s.read(ctx, "pr", "checks", num, "--repo", s.Repo, "--required", "--json", "name,bucket")
	if cerr != nil && len(checksJSON) == 0 {
		return notify.PRView{}, false, fmt.Errorf("ghpr: pr checks: %w", cerr)
	}
	issueJSON, err := s.read(ctx, "issue", "view", strconv.Itoa(issue), "--repo", s.Repo, "--json", "state")
	if err != nil {
		return notify.PRView{}, false, fmt.Errorf("ghpr: issue view: %w", err)
	}
	view, err := Parse(prJSON, checksJSON, issueJSON)
	if err != nil {
		return notify.PRView{}, false, err
	}
	if view.Number != n {
		return notify.PRView{}, false, fmt.Errorf("ghpr: asked for PR #%d and gh answered for #%d", n, view.Number)
	}
	return view, true, nil
}

// Parse builds the view from the three gh outputs. Anything it cannot read is an error: a state that
// is unreadable is never guessed into "success".
func Parse(prJSON, checksJSON, issueJSON []byte) (notify.PRView, error) {
	var pr struct {
		Number         int    `json:"number"`
		HeadRefOid     string `json:"headRefOid"`
		State          string `json:"state"`
		ReviewDecision string `json:"reviewDecision"`
	}
	if err := json.Unmarshal(prJSON, &pr); err != nil {
		return notify.PRView{}, fmt.Errorf("ghpr: pr view output: %w", err)
	}
	if pr.Number <= 0 || !sha40.MatchString(pr.HeadRefOid) {
		return notify.PRView{}, fmt.Errorf("ghpr: pr view output names no pull request head (number %d, head %q)", pr.Number, pr.HeadRefOid)
	}
	var checks []struct {
		Name   string `json:"name"`
		Bucket string `json:"bucket"`
	}
	if err := json.Unmarshal(checksJSON, &checks); err != nil {
		return notify.PRView{}, fmt.Errorf("ghpr: pr checks output: %w", err)
	}
	var is struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(issueJSON, &is); err != nil {
		return notify.PRView{}, fmt.Errorf("ghpr: issue view output: %w", err)
	}
	v := notify.PRView{Number: pr.Number, Head: pr.HeadRefOid, Merged: pr.State == "MERGED", IssueClosed: is.State == "CLOSED"}
	switch pr.ReviewDecision {
	case "APPROVED":
		v.Review = "approved"
	case "CHANGES_REQUESTED":
		v.Review = "changes_requested"
	}
	failed, pending := false, false
	for _, c := range checks {
		switch c.Bucket {
		case "fail", "cancel":
			failed = true
		case "pending":
			pending = true
		case "pass", "skipping":
		default:
			return notify.PRView{}, fmt.Errorf("ghpr: check %q has the unknown bucket %q", c.Name, c.Bucket)
		}
	}
	switch {
	case failed:
		v.CI = "failure"
	case pending:
		v.CI = "pending"
	case len(checks) > 0:
		v.CI = "success"
	}
	return v, nil
}
