package github

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// mockIssueParser implements IssueParser for testing.
type mockIssueParser struct {
	parseFunc func(ctx context.Context, number int) (*Issue, error)
}

func (m *mockIssueParser) ParseIssue(ctx context.Context, number int) (*Issue, error) {
	if m.parseFunc != nil {
		return m.parseFunc(ctx, number)
	}
	return nil, errors.New("not implemented")
}

func TestParseIssueFromJSON_ValidIssue(t *testing.T) {
	data := []byte(`{
		"number": 123,
		"title": "Fix authentication bug",
		"body": "Users cannot login with OAuth provider.",
		"labels": [{"name": "bug"}, {"name": "priority:high"}],
		"author": {"login": "testuser"},
		"comments": [
			{"body": "I can reproduce this.", "author": {"login": "commenter1"}}
		]
	}`)

	issue, err := ParseIssueFromJSON(data)
	if err != nil {
		t.Fatalf("ParseIssueFromJSON() error: %v", err)
	}

	if issue.Number != 123 {
		t.Errorf("Number = %d, want 123", issue.Number)
	}
	if issue.Title != "Fix authentication bug" {
		t.Errorf("Title = %q, want %q", issue.Title, "Fix authentication bug")
	}
	if issue.Body != "Users cannot login with OAuth provider." {
		t.Errorf("Body = %q, want %q", issue.Body, "Users cannot login with OAuth provider.")
	}
	if len(issue.Labels) != 2 {
		t.Fatalf("Labels count = %d, want 2", len(issue.Labels))
	}
	if issue.Labels[0].Name != "bug" {
		t.Errorf("Labels[0].Name = %q, want %q", issue.Labels[0].Name, "bug")
	}
	if issue.Author.Login != "testuser" {
		t.Errorf("Author.Login = %q, want %q", issue.Author.Login, "testuser")
	}
	if len(issue.Comments) != 1 {
		t.Fatalf("Comments count = %d, want 1", len(issue.Comments))
	}
	if issue.Comments[0].Body != "I can reproduce this." {
		t.Errorf("Comments[0].Body = %q, want %q", issue.Comments[0].Body, "I can reproduce this.")
	}
}

func TestParseIssueFromJSON_EmptyTitle(t *testing.T) {
	data := []byte(`{"number": 1, "title": "", "body": "some body"}`)

	_, err := ParseIssueFromJSON(data)
	if err == nil {
		t.Fatal("ParseIssueFromJSON() should error on empty title")
	}
	if got := err.Error(); got != "parse issue JSON: empty title" {
		t.Errorf("error = %q, want %q", got, "parse issue JSON: empty title")
	}
}

func TestParseIssueFromJSON_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json`)

	_, err := ParseIssueFromJSON(data)
	if err == nil {
		t.Fatal("ParseIssueFromJSON() should error on invalid JSON")
	}
}

func TestParseIssueFromJSON_MinimalIssue(t *testing.T) {
	data := []byte(`{"number": 42, "title": "Simple task"}`)

	issue, err := ParseIssueFromJSON(data)
	if err != nil {
		t.Fatalf("ParseIssueFromJSON() error: %v", err)
	}
	if issue.Number != 42 {
		t.Errorf("Number = %d, want 42", issue.Number)
	}
	if issue.Body != "" {
		t.Errorf("Body = %q, want empty", issue.Body)
	}
	if len(issue.Labels) != 0 {
		t.Errorf("Labels count = %d, want 0", len(issue.Labels))
	}
	if len(issue.Comments) != 0 {
		t.Errorf("Comments count = %d, want 0", len(issue.Comments))
	}
}

func TestParseIssueFromJSON_EmptyBody(t *testing.T) {
	data := []byte(`{"number": 10, "title": "No description", "body": ""}`)

	issue, err := ParseIssueFromJSON(data)
	if err != nil {
		t.Fatalf("ParseIssueFromJSON() error: %v", err)
	}
	if issue.Body != "" {
		t.Errorf("Body = %q, want empty", issue.Body)
	}
}

func TestIssue_LabelNames(t *testing.T) {
	tests := []struct {
		name   string
		labels []Label
		want   []string
	}{
		{
			name:   "multiple labels",
			labels: []Label{{Name: "bug"}, {Name: "priority:high"}},
			want:   []string{"bug", "priority:high"},
		},
		{
			name:   "no labels",
			labels: []Label{},
			want:   []string{},
		},
		{
			name:   "nil labels",
			labels: nil,
			want:   []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &Issue{Labels: tt.labels}
			got := issue.LabelNames()
			if len(got) != len(tt.want) {
				t.Fatalf("LabelNames() len = %d, want %d", len(got), len(tt.want))
			}
			for i, name := range got {
				if name != tt.want[i] {
					t.Errorf("LabelNames()[%d] = %q, want %q", i, name, tt.want[i])
				}
			}
		})
	}
}

func TestIssueJSON_RoundTrip(t *testing.T) {
	original := &Issue{
		Number: 99,
		Title:  "Round trip test",
		Body:   "Testing JSON serialization",
		Labels: []Label{{Name: "test"}},
		Author: Author{Login: "roundtripper"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseIssueFromJSON(data)
	if err != nil {
		t.Fatalf("ParseIssueFromJSON error: %v", err)
	}

	if parsed.Number != original.Number {
		t.Errorf("Number = %d, want %d", parsed.Number, original.Number)
	}
	if parsed.Title != original.Title {
		t.Errorf("Title = %q, want %q", parsed.Title, original.Title)
	}
	if parsed.Author.Login != original.Author.Login {
		t.Errorf("Author.Login = %q, want %q", parsed.Author.Login, original.Author.Login)
	}
}

func TestMockIssueParser_Interface(t *testing.T) {
	mock := &mockIssueParser{
		parseFunc: func(_ context.Context, number int) (*Issue, error) {
			if number == 123 {
				return &Issue{Number: 123, Title: "Mock issue"}, nil
			}
			return nil, errors.New("not found")
		},
	}

	// Verify mock satisfies interface.
	var _ IssueParser = mock

	ctx := context.Background()

	issue, err := mock.ParseIssue(ctx, 123)
	if err != nil {
		t.Fatalf("ParseIssue(123) error: %v", err)
	}
	if issue.Title != "Mock issue" {
		t.Errorf("Title = %q, want %q", issue.Title, "Mock issue")
	}

	_, err = mock.ParseIssue(ctx, 999)
	if err == nil {
		t.Error("ParseIssue(999) should error")
	}
}

func TestNewIssueParser(t *testing.T) {
	t.Parallel()

	parser := NewIssueParser("/tmp/test-repo")
	if parser == nil {
		t.Fatal("NewIssueParser returned nil")
	}

	// Verify it implements the interface
	var _ IssueParser = parser //nolint:staticcheck // explicit interface check
}

func TestMockIssueParser_NilFunc(t *testing.T) {
	t.Parallel()

	mock := &mockIssueParser{}
	_, err := mock.ParseIssue(context.Background(), 1)
	if err == nil {
		t.Error("expected error from nil parseFunc")
	}
}

func TestIssue_LabelNames_SingleLabel(t *testing.T) {
	t.Parallel()

	issue := &Issue{
		Labels: []Label{{Name: "enhancement"}},
	}
	names := issue.LabelNames()
	if len(names) != 1 {
		t.Fatalf("LabelNames() len = %d, want 1", len(names))
	}
	if names[0] != "enhancement" {
		t.Errorf("LabelNames()[0] = %q, want %q", names[0], "enhancement")
	}
}

func TestParseIssueFromJSON_WithComments(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"number": 200,
		"title": "Multi-comment issue",
		"body": "body text",
		"labels": [],
		"author": {"login": "user1"},
		"comments": [
			{"body": "First comment", "author": {"login": "user2"}},
			{"body": "Second comment", "author": {"login": "user3"}},
			{"body": "Third comment", "author": {"login": "user1"}}
		]
	}`)

	issue, err := ParseIssueFromJSON(data)
	if err != nil {
		t.Fatalf("ParseIssueFromJSON() error: %v", err)
	}
	if len(issue.Comments) != 3 {
		t.Errorf("Comments count = %d, want 3", len(issue.Comments))
	}
}
