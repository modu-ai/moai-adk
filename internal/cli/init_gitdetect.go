package cli

import (
	"os/exec"
	"strings"
)

// Git mode / provider values resolved by detection. These mirror the closed
// value sets validated by validateInitFlags (--git-mode / --git-provider).
const (
	gitModeManual     = "manual"
	gitModePersonal   = "personal"
	gitProviderGitHub = "github"
	gitProviderGitLab = "gitlab"

	// gitHubHost is the canonical GitHub host. Any other remote host is
	// treated as GitLab (self-hosted instances use arbitrary hostnames).
	gitHubHost = "github.com"
)

// gitRemoteListFunc lists the configured remote names for a directory.
// Overridable in tests.
var gitRemoteListFunc = func(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "remote").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// gitOriginURLFunc resolves the origin remote URL for a directory.
// Overridable in tests.
var gitOriginURLFunc = func(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// detectGitConfig resolves the Git automation mode and hosting provider from
// the repository's configured remotes, replacing the git_mode / git_provider
// wizard questions on the init path.
//
// Mode: a directory with no configured remote — including a non-git directory —
// yields "manual"; at least one remote yields "personal".
//
// Provider: derived from the origin remote's host. A github.com host yields
// "github"; any other host yields "gitlab" (self-hosted GitLab instances use
// arbitrary hostnames). When origin is absent or its host cannot be parsed,
// the provider falls back to the "github" default.
func detectGitConfig(dir string) (mode, provider string) {
	provider = gitProviderGitHub

	out, err := gitRemoteListFunc(dir)
	if err != nil || strings.TrimSpace(out) == "" {
		return gitModeManual, provider
	}
	mode = gitModePersonal

	url, err := gitOriginURLFunc(dir)
	if err != nil {
		return mode, provider
	}
	if host := hostFromRemoteURL(url); host != "" && !strings.EqualFold(host, gitHubHost) {
		provider = gitProviderGitLab
	}
	return mode, provider
}

// hostFromRemoteURL extracts the host from a git remote URL. It handles the
// scp-like SSH form (git@github.com:user/repo.git), explicit schemes
// (https://github.com/user/repo.git, ssh://git@gitlab.example.com/u/r.git),
// and returns "" when no host can be determined (e.g. a local filesystem path).
func hostFromRemoteURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}

	// Strip an explicit scheme (https://, ssh://, git://, ...).
	if i := strings.Index(url, "://"); i >= 0 {
		url = url[i+3:]
	} else if strings.Contains(url, ":") && !strings.HasPrefix(url, "/") {
		// scp-like form: [user@]host:path — normalize the colon to a slash so
		// the shared parsing below applies. A leading "/" means a local path.
		url = strings.Replace(url, ":", "/", 1)
	} else {
		// Local filesystem path — no host.
		return ""
	}

	// Drop userinfo (user@host).
	if i := strings.Index(url, "@"); i >= 0 {
		url = url[i+1:]
	}
	// Keep only the host segment.
	if i := strings.Index(url, "/"); i >= 0 {
		url = url[:i]
	}
	// Drop an explicit port.
	if i := strings.Index(url, ":"); i >= 0 {
		url = url[:i]
	}
	return url
}
