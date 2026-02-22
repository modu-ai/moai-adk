package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Issue represents a parsed GitHub issue.
type Issue struct {
	Number   int       `json:"number"`
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Labels   []Label   `json:"labels"`
	Author   Author    `json:"author"`
	Comments []Comment `json:"comments"`
}

// Label represents a GitHub issue label.
type Label struct {
	Name string `json:"name"`
}

// Author represents a GitHub user.
type Author struct {
	Login string `json:"login"`
}

// Comment represents a GitHub issue comment.
type Comment struct {
	Body      string    `json:"body"`
	Author    Author    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}

// LabelNames returns a string slice of label names for this issue.
func (i *Issue) LabelNames() []string {
	names := make([]string, len(i.Labels))
	for idx, l := range i.Labels {
		names[idx] = l.Name
	}
	return names
}

// IssueParser parses GitHub issues.
type IssueParser interface {
	ParseIssue(ctx context.Context, number int) (*Issue, error)
}

// ghIssueParser implements IssueParser using the gh CLI.
type ghIssueParser struct {
	root string
	// execFn is the function used to execute gh commands.
	// If nil, the package-level execGH function is used.
	execFn execFunc
}

// NewIssueParser returns an IssueParser that uses the gh CLI.
// The root parameter sets the working directory for gh commands,
// which must be inside the target GitHub repository.
func NewIssueParser(root string) IssueParser {
	return &ghIssueParser{root: root}
}

// newIssueParserWithExec creates a ghIssueParser with a custom exec function for testing.
func newIssueParserWithExec(root string, fn execFunc) IssueParser {
	return &ghIssueParser{root: root, execFn: fn}
}

// issueFields is the comma-separated list of fields to request from gh.
const issueFields = "number,title,body,labels,author,comments"

// ParseIssue fetches and parses a GitHub issue by number using the gh CLI.
// Reuses the package-level execGH helper for consistent command execution.
func (p *ghIssueParser) ParseIssue(ctx context.Context, number int) (*Issue, error) {
	if number <= 0 {
		return nil, fmt.Errorf("parse issue: invalid issue number %d", number)
	}

	exec := execGH
	if p.execFn != nil {
		exec = p.execFn
	}

	output, err := exec(ctx, p.root, "issue", "view",
		strconv.Itoa(number), "--json", issueFields)
	if err != nil {
		return nil, fmt.Errorf("parse issue #%d: %w", number, err)
	}

	var issue Issue
	if err := json.Unmarshal([]byte(output), &issue); err != nil {
		return nil, fmt.Errorf("parse issue #%d: unmarshal: %w", number, err)
	}

	if issue.Title == "" {
		return nil, fmt.Errorf("parse issue #%d: empty title", number)
	}

	return &issue, nil
}

// ParseIssueFromJSON parses an Issue from raw JSON bytes.
// Useful for testing and offline processing.
func ParseIssueFromJSON(data []byte) (*Issue, error) {
	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parse issue JSON: %w", err)
	}
	if issue.Title == "" {
		return nil, fmt.Errorf("parse issue JSON: empty title")
	}
	return &issue, nil
}
