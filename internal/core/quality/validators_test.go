package quality

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// --- Mock types (shared by both test files) ---

// mockLSPClient returns configurable diagnostics for testing.
type mockLSPClient struct {
	diagnostics []Diagnostic
	err         error
	delay       time.Duration // optional delay to simulate slow LSP
}

func (m *mockLSPClient) CollectDiagnostics(ctx context.Context) ([]Diagnostic, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.diagnostics, nil
}

// mockGitManager returns configurable git data for testing.
type mockGitManager struct {
	commitMessage string
	commitErr     error
	history       []DiagnosticSnapshot
	historyErr    error
}

func (m *mockGitManager) LastCommitMessage(ctx context.Context) (string, error) {
	if m.commitErr != nil {
		return "", m.commitErr
	}
	return m.commitMessage, nil
}

func (m *mockGitManager) DiagnosticHistory(ctx context.Context) ([]DiagnosticSnapshot, error) {
	if m.historyErr != nil {
		return nil, m.historyErr
	}
	return m.history, nil
}

// --- TestedValidator Tests ---

func TestTestedValidator(t *testing.T) {
	tests := []struct {
		name            string
		diagnostics     []Diagnostic
		coverageTarget  int
		currentCoverage int
		wantPassed      bool
		wantScoreAbove  float64
		wantIssueCount  int
		wantIssueMsg    string
	}{
		{
			name:            "all checks pass",
			diagnostics:     []Diagnostic{},
			coverageTarget:  85,
			currentCoverage: 90,
			wantPassed:      true,
			wantScoreAbove:  0.99,
			wantIssueCount:  0,
		},
		{
			name:            "coverage below threshold",
			diagnostics:     []Diagnostic{},
			coverageTarget:  85,
			currentCoverage: 70,
			wantPassed:      false,
			wantScoreAbove:  0.5,
			wantIssueCount:  1,
			wantIssueMsg:    "test coverage 70% is below target 85%",
		},
		{
			name: "type errors present",
			diagnostics: []Diagnostic{
				{File: "a.go", Line: 1, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 1"},
				{File: "a.go", Line: 5, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 2"},
				{File: "b.go", Line: 3, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 3"},
				{File: "b.go", Line: 7, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 4"},
				{File: "c.go", Line: 2, Severity: SeverityError, Source: "typecheck", Message: "type mismatch 5"},
			},
			coverageTarget:  85,
			currentCoverage: 90,
			wantPassed:      false,
			wantScoreAbove:  0.3,
			wantIssueCount:  5,
		},
		{
			name: "general errors present",
			diagnostics: []Diagnostic{
				{File: "a.go", Line: 10, Severity: SeverityError, Source: "compiler", Message: "undefined variable"},
			},
			coverageTarget:  85,
			currentCoverage: 90,
			wantPassed:      false,
			wantScoreAbove:  0.5,
			wantIssueCount:  1,
		},
		{
			name:            "no coverage target",
			diagnostics:     []Diagnostic{},
			coverageTarget:  0,
			currentCoverage: 0,
			wantPassed:      true,
			wantScoreAbove:  0.99,
			wantIssueCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsp := &mockLSPClient{diagnostics: tt.diagnostics}
			v := NewTestedValidator(lsp, tt.coverageTarget, tt.currentCoverage)

			if v.Name() != PrincipleTested {
				t.Errorf("Name() = %q, want %q", v.Name(), PrincipleTested)
			}

			result, err := v.Validate(context.Background())
			if err != nil {
				t.Fatalf("Validate() returned unexpected error: %v", err)
			}

			if result.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if result.Score < tt.wantScoreAbove {
				t.Errorf("Score = %f, want > %f", result.Score, tt.wantScoreAbove)
			}

			if len(result.Issues) != tt.wantIssueCount {
				t.Errorf("Issues count = %d, want %d", len(result.Issues), tt.wantIssueCount)
			}

			if tt.wantIssueMsg != "" {
				found := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Message, tt.wantIssueMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue message containing %q, got %v", tt.wantIssueMsg, result.Issues)
				}
			}
		})
	}
}

func TestTestedValidator_LSPError(t *testing.T) {
	lsp := &mockLSPClient{err: fmt.Errorf("connection refused")}
	v := NewTestedValidator(lsp, 85, 90)

	_, err := v.Validate(context.Background())
	if err == nil {
		t.Fatal("expected error from LSP failure, got nil")
	}
}

// --- ReadableValidator Tests ---

func TestReadableValidator(t *testing.T) {
	tests := []struct {
		name           string
		diagnostics    []Diagnostic
		wantPassed     bool
		wantScore      float64
		wantIssueCount int
	}{
		{
			name:           "clean - no lint errors",
			diagnostics:    []Diagnostic{},
			wantPassed:     true,
			wantScore:      1.0,
			wantIssueCount: 0,
		},
		{
			name: "lint errors present",
			diagnostics: []Diagnostic{
				{File: "a.go", Line: 1, Severity: SeverityWarning, Source: "lint", Code: "ST1000", Message: "missing package comment"},
				{File: "a.go", Line: 5, Severity: SeverityWarning, Source: "lint", Code: "ST1003", Message: "naming convention"},
				{File: "b.go", Line: 3, Severity: SeverityWarning, Source: "lint", Code: "ST1005", Message: "error string format"},
				{File: "b.go", Line: 7, Severity: SeverityWarning, Source: "lint", Code: "ST1006", Message: "receiver naming"},
			},
			wantPassed:     false,
			wantScore:      0.6,
			wantIssueCount: 4,
		},
		{
			name: "non-lint diagnostics ignored",
			diagnostics: []Diagnostic{
				{File: "a.go", Line: 1, Severity: SeverityError, Source: "typecheck", Message: "type error"},
				{File: "b.go", Line: 2, Severity: SeverityWarning, Source: "security", Message: "sec warning"},
			},
			wantPassed:     true,
			wantScore:      1.0,
			wantIssueCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsp := &mockLSPClient{diagnostics: tt.diagnostics}
			v := NewReadableValidator(lsp)

			if v.Name() != PrincipleReadable {
				t.Errorf("Name() = %q, want %q", v.Name(), PrincipleReadable)
			}

			result, err := v.Validate(context.Background())
			if err != nil {
				t.Fatalf("Validate() returned unexpected error: %v", err)
			}

			if result.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if result.Score != tt.wantScore {
				t.Errorf("Score = %f, want %f", result.Score, tt.wantScore)
			}

			if len(result.Issues) != tt.wantIssueCount {
				t.Errorf("Issues count = %d, want %d", len(result.Issues), tt.wantIssueCount)
			}

			// Verify each issue has a Rule field.
			for i, issue := range result.Issues {
				if issue.Rule == "" {
					t.Errorf("Issue[%d] has empty Rule field", i)
				}
			}
		})
	}
}

// --- UnderstandableValidator Tests ---

func TestUnderstandableValidator(t *testing.T) {
	tests := []struct {
		name             string
		diagnostics      []Diagnostic
		warningThreshold int
		docComplete      bool
		complexityOK     bool
		wantPassed       bool
		wantIssueMsg     string
	}{
		{
			name: "acceptable - all checks pass",
			diagnostics: []Diagnostic{
				{Severity: SeverityWarning, Source: "compiler", Message: "unused variable"},
				{Severity: SeverityWarning, Source: "compiler", Message: "unused import"},
				{Severity: SeverityWarning, Source: "compiler", Message: "shadow variable"},
				{Severity: SeverityWarning, Source: "compiler", Message: "unreachable code"},
				{Severity: SeverityWarning, Source: "compiler", Message: "missing return"},
			},
			warningThreshold: 10,
			docComplete:      true,
			complexityOK:     true,
			wantPassed:       true,
		},
		{
			name: "warnings exceed threshold",
			diagnostics: func() []Diagnostic {
				ds := make([]Diagnostic, 15)
				for i := range ds {
					ds[i] = Diagnostic{Severity: SeverityWarning, Source: "compiler", Message: fmt.Sprintf("warning %d", i+1)}
				}
				return ds
			}(),
			warningThreshold: 10,
			docComplete:      true,
			complexityOK:     true,
			wantPassed:       false,
			wantIssueMsg:     "warning count 15 exceeds threshold 10",
		},
		{
			name:             "doc incomplete",
			diagnostics:      []Diagnostic{},
			warningThreshold: 10,
			docComplete:      false,
			complexityOK:     true,
			wantPassed:       false,
			wantIssueMsg:     "documentation is incomplete",
		},
		{
			name:             "complexity too high",
			diagnostics:      []Diagnostic{},
			warningThreshold: 10,
			docComplete:      true,
			complexityOK:     false,
			wantPassed:       false,
			wantIssueMsg:     "code complexity exceeds",
		},
		{
			name: "security warnings not counted as general warnings",
			diagnostics: []Diagnostic{
				{Severity: SeverityWarning, Source: "security", Message: "sql injection risk"},
			},
			warningThreshold: 0,
			docComplete:      true,
			complexityOK:     true,
			wantPassed:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsp := &mockLSPClient{diagnostics: tt.diagnostics}
			v := NewUnderstandableValidator(lsp, tt.warningThreshold, tt.docComplete, tt.complexityOK)

			if v.Name() != PrincipleUnderstandable {
				t.Errorf("Name() = %q, want %q", v.Name(), PrincipleUnderstandable)
			}

			result, err := v.Validate(context.Background())
			if err != nil {
				t.Fatalf("Validate() returned unexpected error: %v", err)
			}

			if result.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if tt.wantIssueMsg != "" {
				found := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Message, tt.wantIssueMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue containing %q, got %v", tt.wantIssueMsg, result.Issues)
				}
			}
		})
	}
}

// --- SecuredValidator Tests ---

func TestSecuredValidator(t *testing.T) {
	tests := []struct {
		name           string
		diagnostics    []Diagnostic
		wantPassed     bool
		wantScore      float64
		wantIssueCount int
	}{
		{
			name:           "clean - no security warnings",
			diagnostics:    []Diagnostic{},
			wantPassed:     true,
			wantScore:      1.0,
			wantIssueCount: 0,
		},
		{
			name: "security vulnerabilities found",
			diagnostics: []Diagnostic{
				{File: "auth.go", Line: 42, Severity: SeverityWarning, Source: "security", Code: "G401", Message: "weak crypto"},
				{File: "db.go", Line: 15, Severity: SeverityWarning, Source: "security", Code: "G201", Message: "SQL injection risk"},
			},
			wantPassed:     false,
			wantScore:      0.6,
			wantIssueCount: 2,
		},
		{
			name: "non-security diagnostics ignored",
			diagnostics: []Diagnostic{
				{Severity: SeverityError, Source: "typecheck", Message: "type error"},
				{Severity: SeverityWarning, Source: "lint", Message: "lint warning"},
			},
			wantPassed:     true,
			wantScore:      1.0,
			wantIssueCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsp := &mockLSPClient{diagnostics: tt.diagnostics}
			v := NewSecuredValidator(lsp)

			if v.Name() != PrincipleSecured {
				t.Errorf("Name() = %q, want %q", v.Name(), PrincipleSecured)
			}

			result, err := v.Validate(context.Background())
			if err != nil {
				t.Fatalf("Validate() returned unexpected error: %v", err)
			}

			if result.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if result.Score != tt.wantScore {
				t.Errorf("Score = %f, want %f", result.Score, tt.wantScore)
			}

			if len(result.Issues) != tt.wantIssueCount {
				t.Errorf("Issues count = %d, want %d", len(result.Issues), tt.wantIssueCount)
			}

			// Each security issue should have severity "error".
			for i, issue := range result.Issues {
				if issue.Severity != SeverityError {
					t.Errorf("Issue[%d].Severity = %q, want %q", i, issue.Severity, SeverityError)
				}
			}
		})
	}
}

// --- TrackableValidator Tests ---

func TestTrackableValidator(t *testing.T) {
	tests := []struct {
		name           string
		commitMessage  string
		structuredLogs bool
		diagTracked    bool
		wantPassed     bool
		wantIssueMsg   string
	}{
		{
			name:           "clean - all checks pass",
			commitMessage:  "feat(quality): add TRUST 5 validation",
			structuredLogs: true,
			diagTracked:    true,
			wantPassed:     true,
		},
		{
			name:           "invalid commit message",
			commitMessage:  "fixed stuff",
			structuredLogs: true,
			diagTracked:    true,
			wantPassed:     false,
			wantIssueMsg:   "commit message does not follow Conventional Commits format",
		},
		{
			name:           "no structured logs",
			commitMessage:  "fix: resolve nil pointer",
			structuredLogs: false,
			diagTracked:    true,
			wantPassed:     false,
			wantIssueMsg:   "structured logging",
		},
		{
			name:           "no diagnostic tracking",
			commitMessage:  "refactor(core): simplify handler",
			structuredLogs: true,
			diagTracked:    false,
			wantPassed:     false,
			wantIssueMsg:   "diagnostic history",
		},
		{
			name:           "all checks fail",
			commitMessage:  "wip",
			structuredLogs: false,
			diagTracked:    false,
			wantPassed:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitMgr := &mockGitManager{commitMessage: tt.commitMessage}
			v := NewTrackableValidator(gitMgr, tt.structuredLogs, tt.diagTracked)

			if v.Name() != PrincipleTrackable {
				t.Errorf("Name() = %q, want %q", v.Name(), PrincipleTrackable)
			}

			result, err := v.Validate(context.Background())
			if err != nil {
				t.Fatalf("Validate() returned unexpected error: %v", err)
			}

			if result.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if tt.wantIssueMsg != "" {
				found := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Message, tt.wantIssueMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue containing %q, got %v", tt.wantIssueMsg, result.Issues)
				}
			}
		})
	}
}

func TestTrackableValidator_GitError(t *testing.T) {
	gitMgr := &mockGitManager{commitErr: fmt.Errorf("git not found")}
	v := NewTrackableValidator(gitMgr, true, true)

	_, err := v.Validate(context.Background())
	if err == nil {
		t.Fatal("expected error from git failure, got nil")
	}
}

// --- Conventional Commit Regex Tests ---

func TestIsConventionalCommit(t *testing.T) {
	tests := []struct {
		message string
		want    bool
	}{
		{"feat: add new feature", true},
		{"fix(auth): resolve login bug", true},
		{"feat(quality): add TRUST 5 validation", true},
		{"refactor(core): simplify handler", true},
		{"docs: update README", true},
		{"test: add unit tests", true},
		{"chore: update dependencies", true},
		{"build: update Makefile", true},
		{"ci: add GitHub Actions", true},
		{"perf: optimize query", true},
		{"style: format code", true},
		{"feat!: breaking change", true},
		{"fix(api)!: breaking fix", true},
		// Invalid formats.
		{"fixed stuff", false},
		{"wip", false},
		{"", false},
		{"FEAT: uppercase type", false},
		{"feat:", false}, // missing description
		{"feat:no space", false},
		{"update: not a valid type", false},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			got := IsConventionalCommit(tt.message)
			if got != tt.want {
				t.Errorf("IsConventionalCommit(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}
