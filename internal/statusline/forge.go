package statusline

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// A forge is whatever hosts the issues and the change requests for this
// checkout. Everything around the count — the cache, the TTL, the detached
// refresh, the binary-name guard — behaves the same whichever one answers, so
// the forges are a table of a binary name and two argument lists rather than an
// interface whose implementations would each duplicate the rest to vary three
// strings.

// gitlabPageSize bounds how many items `glab` enumerates in one call. GitLab
// paginates and this asks for a single page, so a project with more open items
// than this reports the page rather than the true count — the same trade the
// GitHub path already makes with its own limit, and the same reason: a status
// bar wants a number now, not an accurate number later.
const gitlabPageSize = "100"

// forgeSpec names one hosting service and how to ask it for open counts.
type forgeSpec struct {
	// name is the token that selects this forge explicitly in config.
	name string
	// bin is the CLI that answers. Absent from PATH means no counts, which
	// renders as nothing rather than as an error.
	bin string
	// issueArgs and prArgs each end in a jq filter that reduces the listing to
	// a bare integer, so the caller parses one number whichever forge answered.
	issueArgs []string
	prArgs    []string
}

var forgeGitHub = forgeSpec{
	name: "github",
	bin:  "gh",
	issueArgs: []string{"issue", "list", "--state", "open",
		"--limit", githubListLimit, "--json", "number", "--jq", "length"},
	prArgs: []string{"pr", "list", "--state", "open",
		"--limit", githubListLimit, "--json", "number", "--jq", "length"},
}

// forgeGitLab differs from GitHub in two ways that are easy to get wrong.
//
// Open is glab's default for both listings, so no state flag is passed: `gh`
// needs `--state open` to widen from its own default, while giving glab a state
// flag would narrow rather than widen.
//
// The long-form `--output` is used deliberately. Its short form is not stable
// across glab's own subcommands — on `mr list` `-F` abbreviates `--output`,
// while on `issue list` `-F` abbreviates `--output-format`, which selects
// details/ids/urls and would not produce JSON at all.
var forgeGitLab = forgeSpec{
	name: "gitlab",
	bin:  "glab",
	issueArgs: []string{"issue", "list",
		"--per-page", gitlabPageSize, "--output", "json", "--jq", "length"},
	prArgs: []string{"mr", "list",
		"--per-page", gitlabPageSize, "--output", "json", "--jq", "length"},
}

// argsFor returns the argument list for one kind of item. Anything that is not
// a change request is treated as an issue, so a caller cannot silently ask for
// a listing that does not exist.
func (f forgeSpec) argsFor(kind string) []string {
	if kind == "pr" {
		return f.prArgs
	}
	return f.issueArgs
}

// resolveForge picks the forge for a checkout: an explicit config value wins,
// otherwise the remote's host decides. The second return is false when no forge
// applies, which renders nothing rather than guessing.
//
// An unrecognised override yields no forge rather than falling back to
// detection. A typo should show an absent segment — a visible symptom the
// operator can trace to the value they just typed — rather than quietly
// counting against whatever the hostname happened to suggest.
func resolveForge(remoteURL, override string) (forgeSpec, bool) {
	switch strings.ToLower(strings.TrimSpace(override)) {
	case forgeGitHub.name:
		return forgeGitHub, true
	case forgeGitLab.name:
		return forgeGitLab, true
	case "none", "off":
		return forgeSpec{}, false
	case "":
		return forgeFromRemote(remoteURL)
	default:
		return forgeSpec{}, false
	}
}

// forgeFromRemote maps a remote URL's host onto a forge.
//
// Only the public hosts are recognised. A self-hosted instance answers on a
// name that carries no signal — a company GitLab at git.example.com is
// indistinguishable from a company GitHub Enterprise at the same shape of
// address — so those resolve to no forge and wait for the config override
// rather than being guessed at.
func forgeFromRemote(remoteURL string) (forgeSpec, bool) {
	switch remoteHost(remoteURL) {
	case "github.com", "www.github.com":
		return forgeGitHub, true
	case "gitlab.com", "www.gitlab.com":
		return forgeGitLab, true
	}
	return forgeSpec{}, false
}

// remoteHost extracts the host from a git remote URL.
//
// Two forms reach here and only one is a URL. `https://host/o/r.git` and
// `ssh://git@host/o/r.git` parse; `git@host:o/r.git` is scp-like syntax that
// url.Parse reads as a path with no host, so it is split by hand. The
// discriminator is the scheme separator rather than the colon, because the
// URL form carries a colon too.
func remoteHost(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}

	if !strings.Contains(s, "://") {
		colon := strings.Index(s, ":")
		if colon <= 0 {
			return ""
		}
		hostPart := s[:colon]
		if at := strings.LastIndex(hostPart, "@"); at >= 0 {
			hostPart = hostPart[at+1:]
		}
		return strings.ToLower(hostPart)
	}

	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

// forgeOverride reads statusline.forge from the project's statusline.yaml.
//
// A local shape rather than the full config loader, matching how this package
// already reads llm.yaml: an absent, unreadable, or malformed file yields an
// empty override and detection decides instead.
func forgeOverride(boardRoot string) string {
	if boardRoot == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(boardRoot, ".moai", "config", "sections", "statusline.yaml"))
	if err != nil {
		return ""
	}
	var parsed struct {
		Statusline struct {
			Forge string `yaml:"forge"`
		} `yaml:"statusline"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Statusline.Forge)
}

// originRemoteURL reports the origin remote of a checkout, or "" when there is
// no repository, no origin, or no git.
//
// This shells out rather than reusing the package's git.Repository adapter
// because it runs only in the detached refresh child, where one more process is
// already the cost of the operation. The render path never calls it.
func originRemoteURL(ctx context.Context, dir string) string {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
