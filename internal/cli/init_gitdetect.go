package cli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gitenv"
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

// gitDetectTimeout bounds each git call of the detection (card t2053). A
// variable so a test can shorten it; the production value is the named default
// in internal/config.
var gitDetectTimeout = config.DefaultInitGitProbeTimeout

// errGitDetectTimeout marks a detection call the deadline stopped.
var errGitDetectTimeout = errors.New("timed out")

// gitDetectOutput runs `git -C dir args...` under gitDetectTimeout and returns
// stdout. A deadline hit returns an error wrapping errGitDetectTimeout.
func gitDetectOutput(dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitDetectTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = gitenv.Env() // an inherited GIT_DIR would outrank -C
	cmd.WaitDelay = config.DefaultWorktreeCommandWaitDelay
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("%w after %s", errGitDetectTimeout, gitDetectTimeout)
	}
	return out, err
}

// gitRemoteListFunc lists the configured remote names for a directory.
// Overridable in tests.
var gitRemoteListFunc = func(dir string) (string, error) {
	out, err := gitDetectOutput(dir, "remote")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// gitOriginURLFunc resolves the origin remote URL for a directory.
// Overridable in tests.
var gitOriginURLFunc = func(dir string) (string, error) {
	out, err := gitDetectOutput(dir, "remote", "get-url", "origin")
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
	mode, provider, _ = detectGitConfigReport(dir)
	return mode, provider
}

// detectGitConfigReport is detectGitConfig plus the reason a probe was
// skipped: a git call that timed out (card t2053). Every other outcome — no
// remote, a non-git directory, an unparsable URL — keeps its silent default,
// so skipped is empty. A timeout fails SAFE: the manual mode and the github
// provider, the same fallback as any other failed probe, now with the reason
// reported instead of swallowed.
func detectGitConfigReport(dir string) (mode, provider, skipped string) {
	provider = gitProviderGitHub

	out, err := gitRemoteListFunc(dir)
	if err != nil {
		if errors.Is(err, errGitDetectTimeout) {
			skipped = fmt.Sprintf("git remote %v; using git mode %q", err, gitModeManual)
		}
		return gitModeManual, provider, skipped
	}
	if strings.TrimSpace(out) == "" {
		return gitModeManual, provider, ""
	}
	mode = gitModePersonal

	url, err := gitOriginURLFunc(dir)
	if err != nil {
		if errors.Is(err, errGitDetectTimeout) {
			skipped = fmt.Sprintf("git remote get-url %v; using git provider %q", err, provider)
		}
		return mode, provider, skipped
	}
	if host := hostFromRemoteURL(url); host != "" && !strings.EqualFold(host, gitHubHost) {
		provider = gitProviderGitLab
	}
	return mode, provider, ""
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
