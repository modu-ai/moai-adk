package cli

import (
	"fmt"
	"os/exec"
	"strings"

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
	// --untracked-files=all: default porcelain collapses a wholly-untracked
	// NEW directory to "?? dir/", which the exact-path intersection below
	// would never match (audit round-1 P2, card t1729) — enumerate files.
	out, err := exec.Command("git", "-C", projectRoot, "status", "--porcelain", "--untracked-files=all").Output()
	if err != nil {
		return ""
	}
	changed := make(map[string]bool)
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 4 {
			continue
		}
		// Porcelain v1: "XY " then the path (byte 3 onward). Rename entries
		// carry "old -> new"; the exact-path intersection below matches the
		// whole tail only, so a renamed template path is not counted — the
		// notice undercounts rather than crying wolf.
		changed[strings.TrimSpace(line[3:])] = true
	}
	n := 0
	for _, p := range templatePaths {
		p = strings.TrimSpace(p)
		if p != "" && changed[p] {
			n++
		}
	}
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("update left %d template file(s) uncommitted — worktrees branched from the committed base will not see them until committed: commit on the base branch, then refresh a worktree with `moai worktree sync`", n)
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
