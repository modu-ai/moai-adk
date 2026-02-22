package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/github"
	"github.com/modu-ai/moai-adk/internal/workflow"
)

// GithubIssueParser is the issue parser used by github subcommands.
// Set during dependency injection; tests replace it with a mock.
var GithubIssueParser github.IssueParser

// GithubSpecLinkerFactory creates a SpecLinker for the given project root.
// Tests replace this with a factory that returns a mock.
var GithubSpecLinkerFactory = func(projectRoot string) (github.SpecLinker, error) {
	return github.NewSpecLinker(projectRoot)
}

// GitHub CLI styles.
var (
	ghPrimary = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C45A3C", Dark: "#DA7756"})
	ghBorder  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"})
	ghSuccess = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"})
	ghMuted   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"})
)

func ghCardStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ghBorder.GetForeground()).
		Padding(0, 2)
}

func ghSuccessCard(title string, details ...string) string {
	titleLine := ghSuccess.Render("\u2713") + " " + title
	var body strings.Builder
	body.WriteString(titleLine)
	if len(details) > 0 {
		body.WriteString("\n\n")
		for i, d := range details {
			if i > 0 {
				body.WriteString("\n")
			}
			body.WriteString(d)
		}
	}
	return ghCardStyle().Render(body.String())
}

func ghInfoCard(title, content string) string {
	titleLine := ghPrimary.Bold(true).Render(title)
	body := titleLine + "\n\n" + content
	return ghCardStyle().Render(body)
}

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub integration commands",
	Long:  "Commands for GitHub issue parsing, SPEC linking, and workflow automation.",
}

func init() {
	rootCmd.AddCommand(githubCmd)
	githubCmd.AddCommand(newParseIssueCmd())
	githubCmd.AddCommand(newLinkSpecCmd())
}

func newParseIssueCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "parse-issue <number>",
		Short: "Parse a GitHub issue",
		Long: `Parse a GitHub issue and display its contents.
Uses the gh CLI to fetch issue data from the current repository.

Example:
  moai github parse-issue 123`,
		Args: cobra.ExactArgs(1),
		RunE: runParseIssue,
	}
}

func runParseIssue(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	number, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid issue number %q: %w", args[0], err)
	}

	parser := GithubIssueParser
	if parser == nil {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return fmt.Errorf("get working directory: %w", cwdErr)
		}
		parser = github.NewIssueParser(cwd)
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	issue, err := parser.ParseIssue(ctx, number)
	if err != nil {
		return fmt.Errorf("parse issue: %w", err)
	}

	// Format issue details.
	var details []string
	details = append(details, fmt.Sprintf("Number:  #%d", issue.Number))
	details = append(details, fmt.Sprintf("Title:   %s", issue.Title))
	details = append(details, fmt.Sprintf("Author:  %s", issue.Author.Login))

	if len(issue.Labels) > 0 {
		names := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			names[i] = l.Name
		}
		details = append(details, fmt.Sprintf("Labels:  %s", strings.Join(names, ", ")))
	}

	if issue.Body != "" {
		// Truncate body for display.
		body := issue.Body
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		details = append(details, "")
		details = append(details, ghMuted.Render(body))
	}

	if len(issue.Comments) > 0 {
		details = append(details, "")
		details = append(details, fmt.Sprintf("Comments: %d", len(issue.Comments)))
	}

	_, _ = fmt.Fprintln(out, ghInfoCard(
		fmt.Sprintf("GitHub Issue #%d", issue.Number),
		strings.Join(details, "\n"),
	))
	return nil
}

func newLinkSpecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "link-spec <issue-number> <spec-id>",
		Short: "Link a GitHub issue to a SPEC document",
		Long: `Create a bidirectional link between a GitHub issue and a SPEC document.
The mapping is stored in .moai/github-spec-registry.json.

Example:
  moai github link-spec 123 SPEC-ISSUE-123`,
		Args: cobra.ExactArgs(2),
		RunE: runLinkSpec,
	}
}

func runLinkSpec(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	issueNum, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid issue number %q: %w", args[0], err)
	}

	specID := args[1]
	if err := workflow.ValidateSpecID(specID); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	linker, err := GithubSpecLinkerFactory(cwd)
	if err != nil {
		return fmt.Errorf("create spec linker: %w", err)
	}

	if err := linker.LinkIssueToSpec(issueNum, specID); err != nil {
		return fmt.Errorf("link spec: %w", err)
	}

	_, _ = fmt.Fprintln(out, ghSuccessCard(
		fmt.Sprintf("Linked Issue #%d to %s", issueNum, specID),
		fmt.Sprintf("Issue:     #%d", issueNum),
		fmt.Sprintf("SPEC:      %s", specID),
		fmt.Sprintf("Registry:  .moai/%s", github.RegistryFileName),
		fmt.Sprintf("Linked at: %s", time.Now().Format("2006-01-02 15:04:05")),
	))
	return nil
}
