package cli

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// F3 input-validation constants for git provider identity fields. These mirror
// the documented GitHub / GitLab username rules and are applied before any
// wizard result is persisted so malformed input never lands in a config file.
const (
	// githubUsernameMaxLen is GitHub's documented username length ceiling.
	githubUsernameMaxLen = 39
	// gitlabUsernameMaxLen is GitLab's documented username length ceiling.
	gitlabUsernameMaxLen = 255
)

// githubUsernamePattern accepts a GitHub username: alphanumeric segments
// optionally joined by single hyphens, starting and ending with an
// alphanumeric character (no leading/trailing hyphen, no other punctuation).
var githubUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?$`)

// gitlabUsernamePattern accepts a GitLab username: like GitHub but also
// permitting '.' and '_' in interior positions.
var gitlabUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?$`)

// validateWizardInput checks the externally-supplied identity fields of a
// wizard result before they are persisted (F3). It returns a clear, structured
// error (no interactive prompt — the CLI runs in subagent context and MUST NOT
// call AskUserQuestion) when a value is malformed. Empty values are permitted:
// the field is simply not persisted.
func validateWizardInput(result *wizard.WizardResult) error {
	if result.GitLabInstanceURL != "" {
		if err := validateHTTPSURL(result.GitLabInstanceURL); err != nil {
			return fmt.Errorf("invalid gitlab_instance_url %q: %w", result.GitLabInstanceURL, err)
		}
	}
	if result.GitHubUsername != "" && !isValidGitHubUsername(result.GitHubUsername) {
		return fmt.Errorf("invalid github username %q: must be 1-%d characters, alphanumeric or single hyphens, with no leading or trailing hyphen",
			result.GitHubUsername, githubUsernameMaxLen)
	}
	if result.GitLabUsername != "" && !isValidGitLabUsername(result.GitLabUsername) {
		return fmt.Errorf("invalid gitlab username %q: must be 1-%d characters of letters, digits, '.', '_' or '-', starting and ending with a letter or digit",
			result.GitLabUsername, gitlabUsernameMaxLen)
	}
	return nil
}

// validateHTTPSURL requires a well-formed absolute https:// URL with a host.
// A plaintext http:// endpoint is rejected so a captured token or credential
// is never transmitted over an unencrypted channel.
func validateHTTPSURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("not a well-formed URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("scheme must be https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return errors.New("missing host")
	}
	return nil
}

// isValidGitHubUsername reports whether s is a syntactically valid GitHub
// username.
func isValidGitHubUsername(s string) bool {
	if len(s) > githubUsernameMaxLen {
		return false
	}
	// GitHub disallows consecutive hyphens.
	if strings.Contains(s, "--") {
		return false
	}
	return githubUsernamePattern.MatchString(s)
}

// isValidGitLabUsername reports whether s is a syntactically valid GitLab
// username.
func isValidGitLabUsername(s string) bool {
	if len(s) > gitlabUsernameMaxLen {
		return false
	}
	return gitlabUsernamePattern.MatchString(s)
}
