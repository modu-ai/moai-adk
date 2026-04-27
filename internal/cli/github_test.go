package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/github"
	"github.com/modu-ai/moai-adk/internal/workflow"
)

// mockIssueParser implements github.IssueParser for testing.
type mockGHIssueParser struct {
	parseFunc func(ctx context.Context, number int) (*github.Issue, error)
}

func (m *mockGHIssueParser) ParseIssue(ctx context.Context, number int) (*github.Issue, error) {
	if m.parseFunc != nil {
		return m.parseFunc(ctx, number)
	}
	return nil, errors.New("not implemented")
}

// mockSpecLinker implements github.SpecLinker for testing.
type mockGHSpecLinker struct {
	linkFunc     func(issueNum int, specID string) error
	getSpecFunc  func(issueNum int) (string, error)
	getIssueFunc func(specID string) (int, error)
	listFunc     func() []github.SpecMapping
}

func (m *mockGHSpecLinker) LinkIssueToSpec(issueNum int, specID string) error {
	if m.linkFunc != nil {
		return m.linkFunc(issueNum, specID)
	}
	return nil
}

func (m *mockGHSpecLinker) GetLinkedSpec(issueNum int) (string, error) {
	if m.getSpecFunc != nil {
		return m.getSpecFunc(issueNum)
	}
	return "", errors.New("not found")
}

func (m *mockGHSpecLinker) GetLinkedIssue(specID string) (int, error) {
	if m.getIssueFunc != nil {
		return m.getIssueFunc(specID)
	}
	return 0, errors.New("not found")
}

func (m *mockGHSpecLinker) ListMappings() []github.SpecMapping {
	if m.listFunc != nil {
		return m.listFunc()
	}
	return nil
}

// --- Tests for parse-issue subcommand ---

func TestRunParseIssue_Success(t *testing.T) {
	origParser := GithubIssueParser
	defer func() { GithubIssueParser = origParser }()

	GithubIssueParser = &mockGHIssueParser{
		parseFunc: func(_ context.Context, number int) (*github.Issue, error) {
			return &github.Issue{
				Number: number,
				Title:  "Fix auth bug",
				Body:   "Users cannot login.",
				Labels: []github.Label{{Name: "bug"}, {Name: "priority:high"}},
				Author: github.Author{Login: "testuser"},
				Comments: []github.Comment{
					{Body: "I can reproduce.", Author: github.Author{Login: "reviewer"}},
				},
			}, nil
		},
	}

	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "parse-issue" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"123"})
			if err != nil {
				t.Fatalf("runParseIssue error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "#123") {
				t.Errorf("output should contain issue number, got %q", output)
			}
			if !strings.Contains(output, "Fix auth bug") {
				t.Errorf("output should contain title, got %q", output)
			}
			if !strings.Contains(output, "testuser") {
				t.Errorf("output should contain author, got %q", output)
			}
			if !strings.Contains(output, "bug") {
				t.Errorf("output should contain labels, got %q", output)
			}
			if !strings.Contains(output, "Comments: 1") {
				t.Errorf("output should contain comment count, got %q", output)
			}
			return
		}
	}
	t.Error("parse-issue subcommand not found")
}

func TestRunParseIssue_InvalidNumber(t *testing.T) {
	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "parse-issue" {
			err := cmd.RunE(cmd, []string{"abc"})
			if err == nil {
				t.Error("runParseIssue should error on non-numeric argument")
			}
			if !strings.Contains(err.Error(), "invalid issue number") {
				t.Errorf("error should mention invalid issue number, got %v", err)
			}
			return
		}
	}
	t.Error("parse-issue subcommand not found")
}

func TestRunParseIssue_ParseError(t *testing.T) {
	origParser := GithubIssueParser
	defer func() { GithubIssueParser = origParser }()

	GithubIssueParser = &mockGHIssueParser{
		parseFunc: func(_ context.Context, _ int) (*github.Issue, error) {
			return nil, errors.New("gh CLI not found")
		},
	}

	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "parse-issue" {
			err := cmd.RunE(cmd, []string{"123"})
			if err == nil {
				t.Error("runParseIssue should error when parser fails")
			}
			if !strings.Contains(err.Error(), "parse issue") {
				t.Errorf("error should mention parse issue, got %v", err)
			}
			return
		}
	}
	t.Error("parse-issue subcommand not found")
}

// --- Tests for link-spec subcommand ---

func TestRunLinkSpec_Success(t *testing.T) {
	origFactory := GithubSpecLinkerFactory
	defer func() { GithubSpecLinkerFactory = origFactory }()

	var capturedIssue int
	var capturedSpec string
	GithubSpecLinkerFactory = func(_ string) (github.SpecLinker, error) {
		return &mockGHSpecLinker{
			linkFunc: func(issueNum int, specID string) error {
				capturedIssue = issueNum
				capturedSpec = specID
				return nil
			},
		}, nil
	}

	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "link-spec" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"123", "SPEC-ISSUE-123"})
			if err != nil {
				t.Fatalf("runLinkSpec error: %v", err)
			}

			if capturedIssue != 123 {
				t.Errorf("issue = %d, want 123", capturedIssue)
			}
			if capturedSpec != "SPEC-ISSUE-123" {
				t.Errorf("spec = %q, want %q", capturedSpec, "SPEC-ISSUE-123")
			}

			output := buf.String()
			if !strings.Contains(output, "#123") {
				t.Errorf("output should contain issue number, got %q", output)
			}
			if !strings.Contains(output, "SPEC-ISSUE-123") {
				t.Errorf("output should contain SPEC ID, got %q", output)
			}
			return
		}
	}
	t.Error("link-spec subcommand not found")
}

func TestRunLinkSpec_InvalidNumber(t *testing.T) {
	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "link-spec" {
			err := cmd.RunE(cmd, []string{"abc", "SPEC-ISSUE-123"})
			if err == nil {
				t.Error("runLinkSpec should error on non-numeric issue")
			}
			if !strings.Contains(err.Error(), "invalid issue number") {
				t.Errorf("error should mention invalid issue number, got %v", err)
			}
			return
		}
	}
	t.Error("link-spec subcommand not found")
}

func TestRunLinkSpec_InvalidSpecID(t *testing.T) {
	tests := []struct {
		name   string
		specID string
	}{
		{"empty", ""},
		{"no prefix", "ISSUE-123"},
		{"wrong prefix", "SPEC-123"},
		{"lowercase", "spec-issue-123"},
		{"trailing text", "SPEC-ISSUE-123abc"},
		{"spaces", "SPEC-ISSUE- 123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, cmd := range githubCmd.Commands() {
				if cmd.Name() == "link-spec" {
					err := cmd.RunE(cmd, []string{"1", tt.specID})
					if err == nil {
						t.Fatalf("expected error for SPEC ID %q, got nil", tt.specID)
					}
					if !errors.Is(err, workflow.ErrInvalidSPECID) {
						t.Errorf("error = %v, want ErrInvalidSPECID", err)
					}
					return
				}
			}
			t.Error("link-spec subcommand not found")
		})
	}
}

func TestRunLinkSpec_LinkError(t *testing.T) {
	origFactory := GithubSpecLinkerFactory
	defer func() { GithubSpecLinkerFactory = origFactory }()

	GithubSpecLinkerFactory = func(_ string) (github.SpecLinker, error) {
		return &mockGHSpecLinker{
			linkFunc: func(_ int, _ string) error {
				return errors.New("already linked")
			},
		}, nil
	}

	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "link-spec" {
			err := cmd.RunE(cmd, []string{"123", "SPEC-ISSUE-123"})
			if err == nil {
				t.Error("runLinkSpec should error when linker fails")
			}
			if !strings.Contains(err.Error(), "link spec") {
				t.Errorf("error should mention link spec, got %v", err)
			}
			return
		}
	}
	t.Error("link-spec subcommand not found")
}

func TestRunLinkSpec_FactoryError(t *testing.T) {
	origFactory := GithubSpecLinkerFactory
	defer func() { GithubSpecLinkerFactory = origFactory }()

	GithubSpecLinkerFactory = func(_ string) (github.SpecLinker, error) {
		return nil, errors.New("cannot create linker")
	}

	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "link-spec" {
			err := cmd.RunE(cmd, []string{"123", "SPEC-ISSUE-123"})
			if err == nil {
				t.Error("runLinkSpec should error when factory fails")
			}
			if !strings.Contains(err.Error(), "create spec linker") {
				t.Errorf("error should mention create spec linker, got %v", err)
			}
			return
		}
	}
	t.Error("link-spec subcommand not found")
}

// --- Tests for command structure ---

func TestGithubCmd_HasSubcommands(t *testing.T) {
	expected := map[string]bool{
		"parse-issue": false,
		"link-spec":   false,
		"init":        false, // T-01: 새로운 서브커맨드
		"runner":      false, // T-01: 새로운 서브커맨드
		"auth":        false, // T-01: 새로운 서브커맨드
		"workflow":    false, // T-01: 새로운 서브커맨드
		"status":      false, // T-01: 새로운 서브커맨드
	}

	for _, cmd := range githubCmd.Commands() {
		if _, ok := expected[cmd.Name()]; ok {
			expected[cmd.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("github subcommand %q not found", name)
		}
	}
}

func TestGithubCmd_SubcommandsHaveLongDesc(t *testing.T) {
	for _, cmd := range githubCmd.Commands() {
		if cmd.Long == "" {
			t.Errorf("github subcommand %q should have a Long description", cmd.Name())
		}
	}
}

func TestGithubCmd_ParseIssueRequiresArg(t *testing.T) {
	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "parse-issue" {
			err := cmd.Args(cmd, []string{})
			if err == nil {
				t.Error("parse-issue should require an argument")
			}
			return
		}
	}
	t.Error("parse-issue subcommand not found")
}

func TestGithubCmd_LinkSpecRequiresTwoArgs(t *testing.T) {
	for _, cmd := range githubCmd.Commands() {
		if cmd.Name() == "link-spec" {
			err := cmd.Args(cmd, []string{"123"})
			if err == nil {
				t.Error("link-spec should require two arguments")
			}
			return
		}
	}
	t.Error("link-spec subcommand not found")
}

// --- T-01: 테스트 for 새로운 기능 ---

func TestGithubCmd_HasDryRunFlag(t *testing.T) {
	// githubCmd는 --dry-run 지속적 플래그를 가져야 함
	flag := githubCmd.PersistentFlags().Lookup("dry-run")
	if flag == nil {
		t.Error("github command should have --dry-run persistent flag")
		return
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("--dry-run flag should be bool, got %s", flag.Value.Type())
	}
}

func TestGithubCmd_HasRunnerGroup(t *testing.T) {
	// "runner" 그룹이 존재해야 함
	found := false
	for _, group := range githubCmd.Groups() {
		if group.ID == "runner" {
			found = true
			break
		}
	}
	if !found {
		t.Error("github command should have 'runner' group")
	}
}

func TestGithubCmd_HasAuthGroup(t *testing.T) {
	// "auth" 그룹이 존재해야 함
	found := false
	for _, group := range githubCmd.Groups() {
		if group.ID == "auth" {
			found = true
			break
		}
	}
	if !found {
		t.Error("github command should have 'auth' group")
	}
}

func TestGithubCmd_NewSubcommandsRegistered(t *testing.T) {
	// 새로운 서브커맨드들이 등록되어야 함
	subcommands := []string{"init", "runner", "auth", "workflow", "status"}

	for _, name := range subcommands {
		found := false
		for _, cmd := range githubCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("github subcommand %q not found", name)
		}
	}
}

func TestGithubCmd_ExistingCommandsPreserved(t *testing.T) {
	// 기존 명령들이 보존되어야 함
	existing := []string{"parse-issue", "link-spec"}

	for _, name := range existing {
		found := false
		for _, cmd := range githubCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("existing github subcommand %q not found (should be preserved)", name)
		}
	}
}
