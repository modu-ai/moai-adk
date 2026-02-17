package github

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/core/quality"
	"github.com/modu-ai/moai-adk/internal/git"
	"github.com/modu-ai/moai-adk/internal/i18n"
)

// TestIntegration_ParseIssue_ToBranchDetection validates the cross-package flow
// from parsing an issue's labels to determining the correct branch prefix.
func TestIntegration_ParseIssue_ToBranchDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		issueJSON  string
		wantBranch string
	}{
		{
			name: "bug label produces fix prefix",
			issueJSON: `{
				"number": 101,
				"title": "Login broken",
				"labels": [{"name": "bug"}]
			}`,
			wantBranch: "fix/issue-101",
		},
		{
			name: "feature label produces feat prefix",
			issueJSON: `{
				"number": 202,
				"title": "Add dark mode",
				"labels": [{"name": "feature"}]
			}`,
			wantBranch: "feat/issue-202",
		},
		{
			name: "enhancement label produces feat prefix",
			issueJSON: `{
				"number": 303,
				"title": "Improve performance",
				"labels": [{"name": "enhancement"}]
			}`,
			wantBranch: "feat/issue-303",
		},
		{
			name: "docs label produces docs prefix",
			issueJSON: `{
				"number": 404,
				"title": "Update README",
				"labels": [{"name": "documentation"}]
			}`,
			wantBranch: "docs/issue-404",
		},
		{
			name: "no matching label defaults to feat",
			issueJSON: `{
				"number": 505,
				"title": "Misc task",
				"labels": [{"name": "chore"}, {"name": "low-priority"}]
			}`,
			wantBranch: "feat/issue-505",
		},
		{
			name: "no labels defaults to feat",
			issueJSON: `{
				"number": 606,
				"title": "Unlabeled task"
			}`,
			wantBranch: "feat/issue-606",
		},
		{
			name: "first matching label wins",
			issueJSON: `{
				"number": 707,
				"title": "Bug with docs",
				"labels": [{"name": "bug"}, {"name": "documentation"}]
			}`,
			wantBranch: "fix/issue-707",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			issue, err := ParseIssueFromJSON([]byte(tt.issueJSON))
			if err != nil {
				t.Fatalf("ParseIssueFromJSON() error: %v", err)
			}

			branch, err := git.FormatIssueBranch(issue.LabelNames(), issue.Number)
			if err != nil {
				t.Fatalf("FormatIssueBranch() error: %v", err)
			}
			if branch != tt.wantBranch {
				t.Errorf("FormatIssueBranch() = %q, want %q", branch, tt.wantBranch)
			}
		})
	}
}

// TestIntegration_LinkIssueToSpec_RoundTrip validates the full lifecycle of
// creating an issue-SPEC link and performing bidirectional lookups.
func TestIntegration_LinkIssueToSpec_RoundTrip(t *testing.T) {
	t.Parallel()

	// Simulate parsing 3 issues and linking them to SPECs.
	issues := []struct {
		json   string
		specID string
	}{
		{
			json:   `{"number": 10, "title": "Auth bug", "labels": [{"name": "bug"}]}`,
			specID: "SPEC-ISSUE-10",
		},
		{
			json:   `{"number": 20, "title": "New API", "labels": [{"name": "feature"}]}`,
			specID: "SPEC-ISSUE-20",
		},
		{
			json:   `{"number": 30, "title": "Fix docs", "labels": [{"name": "documentation"}]}`,
			specID: "SPEC-ISSUE-30",
		},
	}

	dir := t.TempDir()
	linker, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker() error: %v", err)
	}

	// Parse each issue and create a link.
	for _, tc := range issues {
		issue, err := ParseIssueFromJSON([]byte(tc.json))
		if err != nil {
			t.Fatalf("ParseIssueFromJSON() error: %v", err)
		}

		if err := linker.LinkIssueToSpec(issue.Number, tc.specID); err != nil {
			t.Fatalf("LinkIssueToSpec(%d, %q) error: %v", issue.Number, tc.specID, err)
		}
	}

	// Verify all forward lookups.
	for _, tc := range issues {
		issue, _ := ParseIssueFromJSON([]byte(tc.json))

		got, err := linker.GetLinkedSpec(issue.Number)
		if err != nil {
			t.Errorf("GetLinkedSpec(%d) error: %v", issue.Number, err)
			continue
		}
		if got != tc.specID {
			t.Errorf("GetLinkedSpec(%d) = %q, want %q", issue.Number, got, tc.specID)
		}
	}

	// Verify all reverse lookups.
	for _, tc := range issues {
		issue, _ := ParseIssueFromJSON([]byte(tc.json))

		got, err := linker.GetLinkedIssue(tc.specID)
		if err != nil {
			t.Errorf("GetLinkedIssue(%q) error: %v", tc.specID, err)
			continue
		}
		if got != issue.Number {
			t.Errorf("GetLinkedIssue(%q) = %d, want %d", tc.specID, got, issue.Number)
		}
	}

	// Verify total count.
	mappings := linker.ListMappings()
	if len(mappings) != 3 {
		t.Errorf("ListMappings() len = %d, want 3", len(mappings))
	}

	// Verify persistence by creating a new linker from the same directory.
	linker2, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker(reload) error: %v", err)
	}
	reloaded := linker2.ListMappings()
	if len(reloaded) != 3 {
		t.Errorf("reloaded ListMappings() len = %d, want 3", len(reloaded))
	}
}

// --- Priority 2: Review → Merge Integration (Milestones 3-4) ---

// TestIntegration_ReviewThenMerge_ApprovedFlow validates the full pipeline
// where a PR passes review and CI, then merges successfully.
func TestIntegration_ReviewThenMerge_ApprovedFlow(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    500,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{
			Overall: CheckPass,
			Checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "success"},
				{Name: "test", Status: "completed", Conclusion: "success"},
			},
		},
	}
	gate := &mockQualityGate{
		report: &quality.Report{Passed: true, Score: 0.95},
	}

	// Create real reviewer and merger with shared GH client.
	reviewer := mustNewPRReviewer(t, gh, gate, nil)
	merger := mustNewPRMerger(t, gh, reviewer, nil)

	// Step 1: Review the PR.
	report, err := reviewer.Review(context.Background(), 500, "SPEC-ISSUE-500", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if report.Decision != ReviewApprove {
		t.Fatalf("Review Decision = %q, want APPROVE", report.Decision)
	}
	if report.QualityReport == nil || !report.QualityReport.Passed {
		t.Error("quality report should be passed")
	}

	// Step 2: Merge the PR using the same reviewer instance.
	result, err := merger.Merge(context.Background(), 500, MergeOptions{
		AutoMerge:     true,
		Method:        MergeMethodSquash,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-500",
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if !result.Merged {
		t.Error("Merged = false, want true")
	}
	if result.Method != MergeMethodSquash {
		t.Errorf("Method = %q, want %q", result.Method, MergeMethodSquash)
	}
	if !gh.prMergeCalled {
		t.Error("gh.PRMerge was not called")
	}
	if gh.prMergeMethod != MergeMethodSquash {
		t.Errorf("gh merge method = %q, want %q", gh.prMergeMethod, MergeMethodSquash)
	}
}

// TestIntegration_ReviewThenMerge_BlockedFlow validates that a failing review
// and CI failure both block the merge operation.
func TestIntegration_ReviewThenMerge_BlockedFlow(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    501,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{
			Overall: CheckFail,
			Checks: []Check{
				{Name: "test", Status: "completed", Conclusion: "failure"},
			},
		},
	}
	gate := &mockQualityGate{
		report: &quality.Report{Passed: false, Score: 0.4},
	}

	reviewer := mustNewPRReviewer(t, gh, gate, nil)
	merger := mustNewPRMerger(t, gh, reviewer, nil)

	// Step 1: Review should return REQUEST_CHANGES.
	report, err := reviewer.Review(context.Background(), 501, "SPEC-ISSUE-501", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if report.Decision != ReviewRequestChanges {
		t.Errorf("Review Decision = %q, want REQUEST_CHANGES", report.Decision)
	}
	if len(report.Issues) == 0 {
		t.Error("Review Issues should not be empty")
	}

	// Step 2: Merge should fail with ErrMergeBlocked.
	_, err = merger.Merge(context.Background(), 501, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-501",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMergeBlocked) {
		t.Errorf("error = %v, want ErrMergeBlocked", err)
	}
	if gh.prMergeCalled {
		t.Error("gh.PRMerge should not have been called when merge is blocked")
	}
}

// TestIntegration_MergeCallCount verifies that CheckPrerequisites passes
// pre-fetched data to the reviewer, avoiding redundant PRView and PRChecks calls.
// Before the optimization, RequireReview+RequireChecks caused 4 gh calls (2 PRView + 2 PRChecks).
// After: only 2 gh calls (1 PRView + 1 PRChecks) because data is fetched once and shared.
func TestIntegration_MergeCallCount(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    510,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{
			Overall: CheckPass,
			Checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "success"},
			},
		},
	}
	gate := &mockQualityGate{
		report: &quality.Report{Passed: true, Score: 1.0},
	}

	reviewer := mustNewPRReviewer(t, gh, gate, nil)
	merger := mustNewPRMerger(t, gh, reviewer, nil)

	_, err := merger.Merge(context.Background(), 510, MergeOptions{
		AutoMerge:     true,
		Method:        MergeMethodSquash,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-510",
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// With pre-fetched data, PRView and PRChecks should each be called exactly once.
	if gh.prViewCallCount != 1 {
		t.Errorf("PRView called %d times, want 1", gh.prViewCallCount)
	}
	if gh.prChecksCallCount != 1 {
		t.Errorf("PRChecks called %d times, want 1", gh.prChecksCallCount)
	}
}

// --- Milestone 5-6: Full Pipeline & i18n Integration ---

// TestIntegration_FullPipeline_IssueToClose validates the complete lifecycle:
// parse issue → detect branch → link SPEC → generate i18n comment → close issue.
func TestIntegration_FullPipeline_IssueToClose(t *testing.T) {
	t.Parallel()

	// Phase 1: Parse an issue.
	issueJSON := `{
		"number": 600,
		"title": "Fix login timeout",
		"body": "Sessions expire too quickly.",
		"labels": [{"name": "bug"}],
		"author": {"login": "reporter"},
		"comments": [{"body": "Confirmed.", "author": {"login": "ops"}}]
	}`
	issue, err := ParseIssueFromJSON([]byte(issueJSON))
	if err != nil {
		t.Fatalf("ParseIssueFromJSON() error: %v", err)
	}
	if issue.Number != 600 {
		t.Fatalf("issue.Number = %d, want 600", issue.Number)
	}

	// Phase 2: Detect branch prefix.
	branch, err := git.FormatIssueBranch(issue.LabelNames(), issue.Number)
	if err != nil {
		t.Fatalf("FormatIssueBranch() error: %v", err)
	}
	if branch != "fix/issue-600" {
		t.Errorf("FormatIssueBranch() = %q, want %q", branch, "fix/issue-600")
	}

	// Phase 3: Link to SPEC.
	dir := t.TempDir()
	linker, err := NewSpecLinker(dir)
	if err != nil {
		t.Fatalf("NewSpecLinker() error: %v", err)
	}
	specID := "SPEC-ISSUE-600"
	if err := linker.LinkIssueToSpec(issue.Number, specID); err != nil {
		t.Fatalf("LinkIssueToSpec() error: %v", err)
	}

	// Phase 4: Generate multilingual comment (Korean).
	gen := i18n.NewCommentGenerator()
	comment, err := gen.Generate("ko", &i18n.CommentData{
		Summary:         "Fixed login timeout by extending session duration",
		PRNumber:        450,
		IssueNumber:     600,
		MergedAt:        time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
		TimeZone:        "KST",
		CoveragePercent: 92,
	})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !strings.Contains(comment, "이슈가 성공적으로 해결되었습니다") {
		t.Error("comment missing Korean success message")
	}
	if !strings.Contains(comment, "#450") {
		t.Error("comment missing PR number")
	}

	// Phase 5: Close the issue with mock exec.
	var capturedArgs [][]string
	mockExec := func(_ context.Context, _ string, args ...string) (string, error) {
		argsCopy := make([]string, len(args))
		copy(argsCopy, args)
		capturedArgs = append(capturedArgs, argsCopy)
		return "", nil
	}

	closer := NewIssueCloser(dir,
		WithExecFunc(mockExec),
		WithMaxRetries(1),
		WithRetryDelay(0),
	)
	result, err := closer.Close(context.Background(), issue.Number, comment)
	if err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	if !result.CommentPosted {
		t.Error("CommentPosted = false, want true")
	}
	if !result.IssueClosed {
		t.Error("IssueClosed = false, want true")
	}

	// Verify the mock received a comment call with issue 600.
	if len(capturedArgs) < 1 {
		t.Fatal("no exec calls captured")
	}
	commentCall := capturedArgs[0]
	if commentCall[0] != "issue" || commentCall[1] != "comment" || commentCall[2] != "600" {
		t.Errorf("first exec call = %v, want [issue comment 600 ...]", commentCall[:3])
	}

	// Phase 6: Verify SPEC link is still intact after closure.
	gotSpec, err := linker.GetLinkedSpec(600)
	if err != nil {
		t.Fatalf("GetLinkedSpec(600) error: %v", err)
	}
	if gotSpec != specID {
		t.Errorf("GetLinkedSpec(600) = %q, want %q", gotSpec, specID)
	}
}

// TestIntegration_I18nCommentToIssueCloser validates that i18n.CommentGenerator
// output feeds correctly into IssueCloser and the 3-step closure process works.
func TestIntegration_I18nCommentToIssueCloser(t *testing.T) {
	t.Parallel()

	// Generate a comment with zero coverage (should omit coverage line).
	gen := i18n.NewCommentGenerator()
	comment, err := gen.Generate("en", &i18n.CommentData{
		Summary:         "Fixed login bug",
		PRNumber:        100,
		IssueNumber:     50,
		MergedAt:        time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		TimeZone:        "UTC",
		CoveragePercent: 0,
	})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Verify comment omits coverage when zero.
	if strings.Contains(comment, "Coverage") {
		t.Error("comment should not contain coverage line when CoveragePercent is 0")
	}

	// Feed the comment to IssueCloser.
	var capturedArgs [][]string
	mockExec := func(_ context.Context, _ string, args ...string) (string, error) {
		argsCopy := make([]string, len(args))
		copy(argsCopy, args)
		capturedArgs = append(capturedArgs, argsCopy)
		return "", nil
	}

	closer := NewIssueCloser(t.TempDir(),
		WithExecFunc(mockExec),
		WithMaxRetries(1),
		WithRetryDelay(0),
	)
	result, err := closer.Close(context.Background(), 50, comment)
	if err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	// Verify all 3 steps succeeded.
	if !result.CommentPosted {
		t.Error("CommentPosted = false")
	}
	if !result.LabelAdded {
		t.Error("LabelAdded = false")
	}
	if !result.IssueClosed {
		t.Error("IssueClosed = false")
	}

	// Verify 3 exec calls: comment, edit (label), close.
	if len(capturedArgs) != 3 {
		t.Fatalf("exec call count = %d, want 3", len(capturedArgs))
	}

	// Call 1: issue comment 50 --body <comment>.
	if capturedArgs[0][0] != "issue" || capturedArgs[0][1] != "comment" {
		t.Errorf("call 1 = %v, want [issue comment ...]", capturedArgs[0][:2])
	}

	// Verify the comment body was passed through.
	bodyIdx := -1
	for i, arg := range capturedArgs[0] {
		if arg == "--body" {
			bodyIdx = i + 1
			break
		}
	}
	if bodyIdx < 0 || bodyIdx >= len(capturedArgs[0]) {
		t.Fatal("--body flag not found in comment call")
	}
	if !strings.Contains(capturedArgs[0][bodyIdx], "Fixed login bug") {
		t.Errorf("comment body missing summary, got %q", capturedArgs[0][bodyIdx])
	}

	// Call 2: issue edit 50 --add-label resolved.
	if capturedArgs[1][0] != "issue" || capturedArgs[1][1] != "edit" {
		t.Errorf("call 2 = %v, want [issue edit ...]", capturedArgs[1][:2])
	}

	// Call 3: issue close 50.
	if capturedArgs[2][0] != "issue" || capturedArgs[2][1] != "close" {
		t.Errorf("call 3 = %v, want [issue close ...]", capturedArgs[2][:2])
	}
}

// TestIntegration_MultilingualCommentGeneration validates that all 4 supported
// languages produce valid comments and unsupported languages fall back to English.
func TestIntegration_MultilingualCommentGeneration(t *testing.T) {
	t.Parallel()

	gen := i18n.NewCommentGenerator()
	data := &i18n.CommentData{
		Summary:         "Added user authentication",
		PRNumber:        456,
		IssueNumber:     123,
		MergedAt:        time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
		TimeZone:        "UTC",
		CoveragePercent: 88,
	}

	tests := []struct {
		lang         string
		wantContains []string
	}{
		{
			lang:         "en",
			wantContains: []string{"resolved successfully", "#456", "88%", "Added user authentication"},
		},
		{
			lang:         "ko",
			wantContains: []string{"성공적으로 해결", "#456", "88%", "Added user authentication"},
		},
		{
			lang:         "ja",
			wantContains: []string{"正常に解決されました", "#456", "88%", "Added user authentication"},
		},
		{
			lang:         "zh",
			wantContains: []string{"成功解决", "#456", "88%", "Added user authentication"},
		},
		{
			lang:         "de",
			wantContains: []string{"resolved successfully", "#456", "88%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			t.Parallel()

			comment, err := gen.Generate(tt.lang, data)
			if err != nil {
				t.Fatalf("Generate(%q) error: %v", tt.lang, err)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(comment, want) {
					t.Errorf("Generate(%q) missing %q in:\n%s", tt.lang, want, comment)
				}
			}
		})
	}
}
