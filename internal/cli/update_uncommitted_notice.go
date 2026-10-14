package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/merge"
)

// t1729 — the end-of-update uncommitted-template notice.
//
// `moai update` refreshes `.claude/rules` / `.moai/policies` (and the rest
// of the managed template set) only in the tree where it ran. Until the
// operator commits those changes on the base branch, every worktree branched
// from the committed base keeps the pre-update files — permanently, from the
// worktree's point of view — and a factory session in such a worktree logs
// "role rule file carries no moai:role-core markers" plus the version-skew
// line. The update therefore ends with a one-line notice naming the
// commit-then-sync path whenever the sync's change set is still uncommitted.

// uncommittedTemplateNotice renders the notice when the working tree carries
// UNCOMMITTED changes to any of templatePaths (the sync's analyzed change
// set); "" otherwise. Any git failure — the project is not a repository, git
// is absent — returns "": the notice is advisory and must never fail an
// update that otherwise succeeded.
//
// Precision note: the working-tree status is intersected with the change
// set, so a user's own unrelated edits never trigger the notice, and files
// the sync rewrote byte-identically (a no-op re-render) never appear in
// porcelain and are naturally excluded. Rename entries ("R  old -> new") and
// core.quotePath-escaped non-ASCII paths fall outside the exact-path
// intersection — the notice then undercounts rather than crying wolf.
func uncommittedTemplateNotice(projectRoot string, templatePaths []string) string {
	if len(templatePaths) == 0 {
		return ""
	}
	// gitenv.Env (audit round-2 P2, card t1729): an inherited GIT_DIR /
	// GIT_WORK_TREE would redirect the query at another repository — the
	// child's repo is determined by -C projectRoot alone.
	runGit := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", projectRoot}, args...)...)
		cmd.Env = gitenv.Env()
		out, err := cmd.Output()
		return string(out), err
	}
	// Path-basis normalization (audit round-2 P2): porcelain prints
	// REPO-root-relative paths while templatePaths are project-relative; a
	// project living in a subdirectory of its repository would never match
	// without prefixing. `rev-parse --show-prefix` is git's own answer for
	// the cwd→repo-root prefix (slash-terminated, empty at the toplevel) —
	// it also fails outside a repository, keeping the fail-open shape, and
	// avoids the resolved-vs-symlinked path mismatch a toplevel Rel() would
	// hit on macOS (/private/var vs /var).
	prefOut, err := runGit("rev-parse", "--show-prefix")
	if err != nil {
		return ""
	}
	prefix := strings.TrimSpace(prefOut)
	wanted := make(map[string]bool, len(templatePaths))
	pathspecs := make([]string, 0, len(templatePaths))
	for _, p := range templatePaths {
		if p = strings.TrimSpace(p); p != "" {
			wanted[prefix+p] = true
			pathspecs = append(pathspecs, p)
		}
	}
	if len(wanted) == 0 {
		return ""
	}
	// --untracked-files=all: default porcelain collapses a wholly-untracked
	// NEW directory to "?? dir/", which the exact-path intersection below
	// would never match (audit round-1 P2). The pathspecs scope the query
	// to this sync's change set (git resolves them cwd-relative), so a
	// user's unrelated edits still never trigger the notice.
	args := append([]string{"status", "--porcelain", "--untracked-files=all", "--"}, pathspecs...)
	out, err := runGit(args...)
	if err != nil {
		return ""
	}
	n := 0
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 4 {
			continue
		}
		// Porcelain v1: "XY " then the path (byte 3 onward). Rename entries
		// carry "old -> new"; the exact-path intersection below matches the
		// whole tail only, so a renamed template path is not counted — the
		// notice undercounts rather than crying wolf.
		if wanted[strings.TrimSpace(line[3:])] {
			n++
		}
	}
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("update left %d template file(s) uncommitted — worktrees branched from the committed base will not see them until committed: commit on the base branch, then merge the base branch into each worktree (`git merge main` inside the worktree)", n)
}

// mergeAnalysisPaths extracts the deployment-relative paths one template
// sync analyzed — the change set the uncommitted notice queries git status
// against.
func mergeAnalysisPaths(files []merge.FileAnalysis) []string {
	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	return paths
}
