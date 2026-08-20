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
	// They are the enumerating fallback, used when countArgs is empty.
	issueArgs []string
	prArgs    []string
	// countArgs, when set, is a SINGLE call answering both counts as two lines
	// — issues first, change requests second. A forge that can be asked for a
	// total does not need its items listed, so this path is preferred wherever
	// it exists (see forgeGitHub).
	countArgs []string
	// rateArgs, when set, is a call reporting how much API budget remains, as
	// a bare integer, WITHOUT consuming any of it. Empty means the forge
	// offers no free way to ask and the refresh proceeds unguarded.
	rateArgs []string
}

// forgeGitHub asks for totals rather than for listings.
//
// `gh repo view --json issues,pullRequests` returns each connection's
// totalCount for the OPEN state — the number the two `list` forms were
// reaching by counting items. Measured against `cli/cli` on 2026-08-20:
//
//   - the totals call cost 1 GraphQL point and reported 1012 open issues;
//   - `issue list --limit 1000` cost 10 points — one per page of a hundred —
//     and reported 1000, the cap rather than the count.
//
// The enumeration was therefore both dearer and, past its cap, wrong. The
// totals are exact at any size and cost one point for BOTH numbers, since the
// single call answers issues and pull requests together.
//
// issueArgs/prArgs remain as the shape a forge without a totals endpoint uses;
// on GitHub countArgs takes precedence and they go unused.
var forgeGitHub = forgeSpec{
	name: "github",
	bin:  "gh",
	countArgs: []string{"repo", "view", "--json", "issues,pullRequests",
		"--jq", ".issues.totalCount, .pullRequests.totalCount"},
	rateArgs: []string{"api", "rate_limit", "--jq", ".resources.graphql.remaining"},
	issueArgs: []string{"issue", "list", "--state", "open",
		"--limit", githubListLimit, "--json", "number", "--jq", "length"},
	prArgs: []string{"pr", "list", "--state", "open",
		"--limit", githubListLimit, "--json", "number", "--jq", "length"},
}

// forgeGitLab still enumerates, and the asymmetry with GitHub is deliberate
// rather than an oversight. `glab` exposes no count-only listing: `issue list`
// and `mr list` return items, and the total lives in a REST pagination header
// (`X-Total`) that `glab` does not surface through its own JSON output.
// Reaching it would mean calling the API around `glab`, which buys a second
// authentication path to get right — a poor trade for a status bar. GitLab
// therefore keeps the page-bounded enumeration and the truncation trade
// `gitlabPageSize` documents.
//
// Unverified: `glab` is not installed on the machine this was written on
// (measured absent from PATH, 2026-08-20), so the paragraph above rests on
// glab's documented output rather than on a run. A later confirmation that a
// count-only path exists should move GitLab onto countArgs the same way.
//
// It also differs from GitHub in two ways that are easy to get wrong.
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
