package cli

// SPEC-MCP-WORKTREE-ROOT-001 — let an MCP caller name the tree it means.
//
// resolveProjectDir() prefers CLAUDE_PROJECT_DIR, and Claude Code sets that to
// the PROJECT root even for a session working inside a worktree. Measured across
// four live moai mcp-server processes in two repositories: every server whose
// working directory was a worktree carried the primary checkout in that
// variable. The env branch wins, so the working directory is never consulted and
// the outcome is deterministic — an MCP audit issued from a worktree audits the
// primary checkout, and a SPEC that exists only on the card's branch is absent
// from the catalogue the auditor reads without anything reporting it missing.
//
// The server cannot infer the right answer: its own working directory is stale
// by construction (a long-lived subprocess cannot follow a worktree switch) and
// its environment names the project. The caller CAN — an agent with a shell runs
// `git rev-parse --show-toplevel`. So the tool takes it as an input.
//
// The repair sits here, at the handler boundary, rather than inside
// resolveProjectDir(): that function has other consumers whose correct answer is
// undecided (goal state, verification snapshots, the convergence state
// directory), and changing it would move them all at once.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// projectRootArg is the optional tool input naming the tree to act on.
const projectRootArg = "project_root"

// The two input descriptions share an opening and differ in exactly the way the
// two resolvers differ — what an ABSENT parameter means. They are separate
// strings because a single shared one was wrong: it promised every tool the
// resolveProjectDir() fallback, which is not what the pass-through tools do, so
// a caller reading the schema for `audit_multi` was told the wrong thing. A
// description that describes four of five tools is not a shared description, it
// is an inaccurate one.
const projectRootDescCommon = "Optional project or worktree root to act on. Supply your own " +
	"`git rev-parse --show-toplevel`. In a worktree session you MUST pass it: the server's own resolution names " +
	"the PRIMARY checkout, so omitting it acts on the wrong tree. An unusable path is rejected, never silently " +
	"replaced by a default, and an accepted path is canonicalized — symlinks resolved — so the call acts on the " +
	"real directory rather than on the spelling that reached it. "

// projectRootDesc describes the parameter on a tool whose absent case falls back
// to resolveProjectDir() — the tools that already resolved a root before this
// parameter existed.
const projectRootDesc = projectRootDescCommon +
	"Absent or empty resolves as before: CLAUDE_PROJECT_DIR, then the server's working directory."

// projectRootPassthroughDesc describes the parameter on a tool that carried NO
// root before this parameter existed, so its absent case supplies none rather
// than substituting a default.
const projectRootPassthroughDesc = projectRootDescCommon +
	"Absent or empty supplies no root at all, exactly as before this parameter existed — the backends then " +
	"resolve their own working directory."

// projectRootOption declares the optional input on a fallback-semantics tool.
func projectRootOption() mcp.ToolOption {
	return mcp.WithString(projectRootArg, mcp.Description(projectRootDesc))
}

// projectRootPassthroughOption declares it on a pass-through-semantics tool.
func projectRootPassthroughOption() mcp.ToolOption {
	return mcp.WithString(projectRootArg, mcp.Description(projectRootPassthroughDesc))
}

// rootSourceParam names the provenance source of an explicit project_root
// argument. The fallback tier names ("env:CLAUDE_PROJECT_DIR", "server-cwd",
// "unresolved") come from resolveProjectDirWithSource (session.go).
const rootSourceParam = "param"

// resolveToolProjectRoot returns the project root a tool call should act on.
//
// An absent or empty project_root resolves exactly as the tool resolved it
// before this parameter existed, so a caller that never learns about it sees no
// change.
//
// A non-empty project_root that cannot be a project root is REJECTED rather than
// replaced by the default. The ergonomic alternative — ignore the bad path, use
// the default — is the one choice that reintroduces the very defect this
// parameter exists to fix: a caller who mistyped its own worktree path would be
// silently returned to acting on the primary checkout, and told it succeeded.
func resolveToolProjectRoot(req mcp.CallToolRequest) (string, error) {
	root, _, err := resolveToolProjectRootWithSource(req)
	return root, err
}

// resolveToolProjectRootWithSource is resolveToolProjectRoot plus the
// provenance tier that produced the root (t236 / issue #1640): "param" for an
// explicit argument, otherwise the fallback tier from
// resolveProjectDirWithSource. Validation and reject-not-fallback semantics
// are identical to resolveToolProjectRoot — this is the same resolver with
// the source made observable, so catalog responses can carry it as "_root"
// and warn when resolution did NOT come from the caller.
func resolveToolProjectRootWithSource(req mcp.CallToolRequest) (string, string, error) {
	raw := strings.TrimSpace(req.GetString(projectRootArg, ""))
	if raw == "" {
		dir, source := resolveProjectDirWithSource()
		return dir, source, nil
	}
	root, err := validateProjectRoot(raw)
	if err != nil {
		return "", "", err
	}
	return root, rootSourceParam, nil
}

// rootProvenanceMap builds the "_root" block a catalog response carries:
// which tree was read and where the resolution came from. A warning is
// attached ONLY when the source is not "param" — a fallback resolution froze
// at server spawn, and the caller must be told rather than left reading
// another tree in silence (live gap L2 of the t236 reproduction).
func rootProvenanceMap(root, source string) map[string]any {
	prov := map[string]any{
		"source": source,
		"dir":    root,
	}
	if source != rootSourceParam {
		prov["warning"] = "project_root not passed — resolved from " + source +
			", which froze at server spawn; a session that moved worktrees is reading another tree; " +
			"pass project_root = git rev-parse --show-toplevel"
	}
	return prov
}

// resolveOptionalToolProjectRoot is the pass-through variant, for a surface that
// carries NO root today rather than carrying the wrong one.
//
// The distinction is what keeps REQ-2 literally true on both codex paths.
// `codex_audit` already hands codex a `cwd` (resolveProjectDir()), so an absent
// parameter there must keep handing it that same value — resolveToolProjectRoot's
// fallback. `audit_multi`'s fan-out hands codex no `cwd` at all, so an absent
// parameter there must keep handing it none: substituting a default would change
// what an existing caller's backend receives, which is the one thing REQ-2
// forbids. Validation and rejection are identical; only the absent case differs.
func resolveOptionalToolProjectRoot(req mcp.CallToolRequest) (string, error) {
	raw := strings.TrimSpace(req.GetString(projectRootArg, ""))
	if raw == "" {
		return "", nil
	}
	return validateProjectRoot(raw)
}

// validateProjectRoot turns a caller-supplied path into an absolute,
// symlink-free project root, or rejects it. Shared by both resolvers so the two
// cannot drift into accepting different things.
func validateProjectRoot(raw string) (string, error) {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("project_root %q cannot be resolved to an absolute path: %w", raw, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("project_root %q does not exist", raw)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("project_root %q is not a directory", raw)
	}

	// filepath.Abs does not resolve symlinks, so canonicalize before returning.
	// Nothing compares this result against another path TODAY, which is why the
	// omission was harmless; it stops being harmless the moment a containment
	// constraint is added (the audit's suggested "the root must sit under this
	// repository's git common dir"), because a symlinked spelling and its target
	// are two different strings for one directory and the comparison would then
	// turn on the spelling. Canonicalizing here is what keeps such a boundary
	// load-bearing rather than decorative.
	//
	// A canonicalization failure is a REJECTION, not a fallback to the
	// uncanonicalized path — same reasoning as the rejections above: returning a
	// path the caller cannot be told is non-canonical reintroduces "it
	// succeeded, on something other than what you named".
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("project_root %q cannot be canonicalized: %w", raw, err)
	}

	moaiDir := filepath.Join(canonical, ".moai")
	moaiInfo, err := os.Stat(moaiDir)
	if err != nil || !moaiInfo.IsDir() {
		return "", fmt.Errorf("project_root %q has no .moai directory, so it is not a MoAI project root", raw)
	}

	return canonical, nil
}
