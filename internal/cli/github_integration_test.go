package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/github"
	"github.com/spf13/cobra"
)

// TestGithubCLI_ParseThenLink_Workflow validates the full CLI workflow of
// parsing an issue and then linking it to a SPEC document.
func TestGithubCLI_ParseThenLink_Workflow(t *testing.T) {
	// Save and restore originals.
	origParser := GithubIssueParser
	origFactory := GithubSpecLinkerFactory
	defer func() {
		GithubIssueParser = origParser
		GithubSpecLinkerFactory = origFactory
	}()

	// Set up mock parser.
	GithubIssueParser = &mockGHIssueParser{
		parseFunc: func(_ context.Context, number int) (*github.Issue, error) {
			return &github.Issue{
				Number: number,
				Title:  "Auth session timeout",
				Body:   "Sessions expire too quickly.",
				Labels: []github.Label{{Name: "bug"}, {Name: "auth"}},
				Author: github.Author{Login: "reporter"},
				Comments: []github.Comment{
					{Body: "Confirmed on production.", Author: github.Author{Login: "ops"}},
				},
			}, nil
		},
	}

	// Track link-spec calls.
	var linkedIssue int
	var linkedSpec string
	GithubSpecLinkerFactory = func(_ string) (github.SpecLinker, error) {
		return &mockGHSpecLinker{
			linkFunc: func(issueNum int, specID string) error {
				linkedIssue = issueNum
				linkedSpec = specID
				return nil
			},
		}, nil
	}

	// Step 1: Parse the issue.
	var parseCmd, linkCmd *cobra.Command
	for _, cmd := range githubCmd.Commands() {
		switch cmd.Name() {
		case "parse-issue":
			parseCmd = cmd
		case "link-spec":
			linkCmd = cmd
		}
	}
	if parseCmd == nil || linkCmd == nil {
		t.Fatal("parse-issue or link-spec subcommand not found")
	}

	parseBuf := new(bytes.Buffer)
	parseCmd.SetOut(parseBuf)
	parseCmd.SetErr(parseBuf)

	if err := parseCmd.RunE(parseCmd, []string{"456"}); err != nil {
		t.Fatalf("parse-issue error: %v", err)
	}

	parseOutput := parseBuf.String()
	if !strings.Contains(parseOutput, "#456") {
		t.Errorf("parse output should contain issue number, got %q", parseOutput)
	}
	if !strings.Contains(parseOutput, "Auth session timeout") {
		t.Errorf("parse output should contain title, got %q", parseOutput)
	}

	// Step 2: Link the issue to a SPEC.
	linkBuf := new(bytes.Buffer)
	linkCmd.SetOut(linkBuf)
	linkCmd.SetErr(linkBuf)

	if err := linkCmd.RunE(linkCmd, []string{"456", "SPEC-ISSUE-456"}); err != nil {
		t.Fatalf("link-spec error: %v", err)
	}

	linkOutput := linkBuf.String()
	if !strings.Contains(linkOutput, "#456") {
		t.Errorf("link output should contain issue number, got %q", linkOutput)
	}
	if !strings.Contains(linkOutput, "SPEC-ISSUE-456") {
		t.Errorf("link output should contain SPEC ID, got %q", linkOutput)
	}

	// Verify the mock captured the correct arguments.
	if linkedIssue != 456 {
		t.Errorf("linked issue = %d, want 456", linkedIssue)
	}
	if linkedSpec != "SPEC-ISSUE-456" {
		t.Errorf("linked spec = %q, want %q", linkedSpec, "SPEC-ISSUE-456")
	}
}

// TestGithubCLI_ErrorPropagation validates that errors from underlying
// components propagate correctly through CLI commands.
func TestGithubCLI_ErrorPropagation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		command     string
		args        []string
		setupParser func()
		setupLinker func()
		wantErrMsg  string
	}{
		{
			name:    "parse-issue with gh failure",
			command: "parse-issue",
			args:    []string{"789"},
			setupParser: func() {
				GithubIssueParser = &mockGHIssueParser{
					parseFunc: func(_ context.Context, _ int) (*github.Issue, error) {
						return nil, errors.New("gh: Could not resolve to a Repository")
					},
				}
			},
			wantErrMsg: "parse issue",
		},
		{
			name:    "link-spec with duplicate mapping",
			command: "link-spec",
			args:    []string{"100", "SPEC-ISSUE-100"},
			setupLinker: func() {
				GithubSpecLinkerFactory = func(_ string) (github.SpecLinker, error) {
					return &mockGHSpecLinker{
						linkFunc: func(_ int, _ string) error {
							return github.ErrMappingExists
						},
					}, nil
				}
			},
			wantErrMsg: "link spec",
		},
		{
			name:    "link-spec with factory error",
			command: "link-spec",
			args:    []string{"200", "SPEC-ISSUE-200"},
			setupLinker: func() {
				GithubSpecLinkerFactory = func(_ string) (github.SpecLinker, error) {
					return nil, errors.New("permission denied")
				}
			},
			wantErrMsg: "create spec linker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore.
			origParser := GithubIssueParser
			origFactory := GithubSpecLinkerFactory
			defer func() {
				GithubIssueParser = origParser
				GithubSpecLinkerFactory = origFactory
			}()

			if tt.setupParser != nil {
				tt.setupParser()
			}
			if tt.setupLinker != nil {
				tt.setupLinker()
			}

			var targetCmd *cobra.Command
			for _, cmd := range githubCmd.Commands() {
				if cmd.Name() == tt.command {
					targetCmd = cmd
					break
				}
			}
			if targetCmd == nil {
				t.Fatalf("subcommand %q not found", tt.command)
			}

			err := targetCmd.RunE(targetCmd, tt.args)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
			}
		})
	}
}
