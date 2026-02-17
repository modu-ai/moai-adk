package github

import (
	"context"
	"testing"
)

// mockGHClient implements GHClient for testing.
type mockGHClient struct {
	prCreateResult int
	prCreateErr    error
	prViewResult   *PRDetails
	prViewErr      error
	prMergeErr     error
	prChecksResult *CheckStatus
	prChecksErr    error
	pushErr        error
	authErr        error

	// Track calls for assertion.
	prMergeCalled       bool
	prMergeNumber       int
	prMergeMethod       MergeMethod
	prMergeDeleteBranch bool
	pushCalled          bool
	pushDir           string
	prViewCallCount   int
	prChecksCallCount int
}

func (m *mockGHClient) PRCreate(_ context.Context, _ PRCreateOptions) (int, error) {
	return m.prCreateResult, m.prCreateErr
}

func (m *mockGHClient) PRView(_ context.Context, _ int) (*PRDetails, error) {
	m.prViewCallCount++
	return m.prViewResult, m.prViewErr
}

func (m *mockGHClient) PRMerge(_ context.Context, number int, method MergeMethod, deleteBranch bool) error {
	m.prMergeCalled = true
	m.prMergeNumber = number
	m.prMergeMethod = method
	m.prMergeDeleteBranch = deleteBranch
	return m.prMergeErr
}

func (m *mockGHClient) PRChecks(_ context.Context, _ int) (*CheckStatus, error) {
	m.prChecksCallCount++
	return m.prChecksResult, m.prChecksErr
}

func (m *mockGHClient) Push(_ context.Context, dir string) error {
	m.pushCalled = true
	m.pushDir = dir
	return m.pushErr
}

func (m *mockGHClient) IsAuthenticated(_ context.Context) error {
	return m.authErr
}

func TestExtractPRNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "standard github URL",
			input: "https://github.com/modu-ai/moai-adk/pull/456",
			want:  456,
		},
		{
			name:  "URL with trailing newline",
			input: "https://github.com/owner/repo/pull/123\n",
			want:  123,
		},
		{
			name:    "invalid URL no number",
			input:   "https://github.com/owner/repo/pull/abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing pull segment",
			input:   "https://github.com/owner/repo/issues/42",
			wantErr: true,
		},
		{
			name:    "just a number",
			input:   "42",
			wantErr: true,
		},
		{
			name:    "repo URL without pull",
			input:   "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "negative PR number",
			input:   "https://github.com/owner/repo/pull/-1",
			wantErr: true,
		},
		{
			name:    "zero PR number",
			input:   "https://github.com/owner/repo/pull/0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := extractPRNumber(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("extractPRNumber(%q) = %d, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("extractPRNumber(%q) error = %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("extractPRNumber(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestDeriveOverallConclusion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		checks []Check
		want   CheckConclusion
	}{
		{
			name:   "no checks returns pending",
			checks: nil,
			want:   CheckPending,
		},
		{
			name: "all success returns pass",
			checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "success"},
				{Name: "test", Status: "completed", Conclusion: "success"},
			},
			want: CheckPass,
		},
		{
			name: "one failure returns fail",
			checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "success"},
				{Name: "test", Status: "completed", Conclusion: "failure"},
			},
			want: CheckFail,
		},
		{
			name: "one pending returns pending",
			checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "success"},
				{Name: "test", Status: "in_progress", Conclusion: ""},
			},
			want: CheckPending,
		},
		{
			name: "failure takes priority over pending",
			checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "failure"},
				{Name: "test", Status: "in_progress", Conclusion: ""},
			},
			want: CheckFail,
		},
		{
			name: "cancelled is treated as failure",
			checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "cancelled"},
			},
			want: CheckFail,
		},
		{
			name: "timed_out is treated as failure",
			checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "timed_out"},
			},
			want: CheckFail,
		},
		{
			name: "skipped and neutral are pass",
			checks: []Check{
				{Name: "optional", Status: "completed", Conclusion: "skipped"},
				{Name: "info", Status: "completed", Conclusion: "neutral"},
			},
			want: CheckPass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := deriveOverallConclusion(tt.checks)
			if got != tt.want {
				t.Errorf("deriveOverallConclusion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewGHClient(t *testing.T) {
	t.Parallel()

	client := NewGHClient("/tmp/test-repo")
	if client == nil {
		t.Fatal("NewGHClient returned nil")
	}
	if client.root != "/tmp/test-repo" {
		t.Errorf("root = %q, want %q", client.root, "/tmp/test-repo")
	}
}
