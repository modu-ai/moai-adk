// Package gitenv scrubs the git environment variables that decide WHICH
// repository a child process acts on.
//
// git exports GIT_DIR (and, on the commit path, GIT_INDEX_FILE) into every hook
// it runs. A pre-commit hook that shells out to `moai` hands those to the
// process, which hands them again to every child it spawns. Those variables
// OUTRANK the working directory, so setting cmd.Dir is not, on its own,
// isolation: a child resolves through the caller's git context instead.
//
// The consequence has two shapes, both reported in GH #1691:
//
//   - A WRITE lands in the caller's repository. A project's own test fixtures,
//     creating throwaway commits in a temp directory, wrote them into the
//     repository being committed to. The reported incident lost an uncommitted
//     reviewed diff and injected a junk commit chain into a live branch.
//   - A READ answers about the caller's repository. Quieter, and it corrupts
//     nothing: a command asked about directory X reports about repository Y,
//     and every decision downstream is taken on another repository's state.
//
// A third shape is specific to a commit made FROM a linked worktree, where the
// exported GIT_DIR names <repo>/.git/worktrees/<name> rather than <repo>/.git.
// A child `git init` under that variable rewrites core.bare=true into
// <repo>/.git/config — shared by every linked worktree — and the whole
// repository then fails with "fatal: this operation must be run in a work tree".
// Only the per-worktree gitdir does this; the shared gitdir and GIT_INDEX_FILE
// alone do not.
//
// This package exists as its own package rather than as a helper beside any one
// caller because the defect it prevents is precisely a fix that could not reach
// its siblings: the first repair lived inside one package, and the identical
// call one screen away in the same file kept leaking. One list, one place.
package gitenv

import (
	"os"
	"strings"
)

// RepoScopingVars names the git environment variables that decide WHICH
// repository a git command acts on, as opposed to how it behaves or who it
// attributes work to.
//
// Scope is deliberately narrow. Only repository LOCATION is removed:
//
//   - Identity (GIT_AUTHOR_NAME, GIT_COMMITTER_EMAIL, …) stays. A child that
//     lost its identity would fail for an unrelated reason, turning a clean
//     isolation failure into a confusing one.
//   - Behaviour (GIT_EDITOR, GIT_PAGER, GIT_SSH_COMMAND, GIT_CONFIG_GLOBAL, …)
//     stays. A project or CI that sets those means them, and dropping them
//     would change what a child does rather than where it does it.
var RepoScopingVars = []string{
	"GIT_DIR",
	"GIT_COMMON_DIR",
	"GIT_WORK_TREE",
	"GIT_INDEX_FILE",
	"GIT_OBJECT_DIRECTORY",
	"GIT_ALTERNATE_OBJECT_DIRECTORIES",
	"GIT_NAMESPACE",
	"GIT_PREFIX",
	"GIT_SUPER_PREFIX",
	"GIT_CEILING_DIRECTORIES",
	"GIT_QUARANTINE_PATH",
}

// Scrub returns env with every RepoScopingVars entry removed. Everything else
// is carried through in order, so a child sees the environment it would have
// seen but cannot be pointed at a repository other than the one its working
// directory names.
//
// The returned slice is always non-nil, including when env is empty: os/exec
// reads a nil Env as "inherit the parent's environment", which is exactly the
// behaviour this function exists to prevent. A caller that assigns the result
// to cmd.Env therefore never re-creates the defect by passing nothing.
func Scrub(env []string) []string {
	drop := make(map[string]struct{}, len(RepoScopingVars))
	for _, name := range RepoScopingVars {
		drop[name] = struct{}{}
	}

	out := make([]string, 0, len(env))
	for _, kv := range env {
		name, _, ok := strings.Cut(kv, "=")
		if !ok {
			// Not a NAME=VALUE pair; pass it through rather than guess.
			out = append(out, kv)
			continue
		}
		if _, dropped := drop[name]; dropped {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// Env is the environment a child process should run with when its repository is
// determined by its working directory rather than by the caller's git context.
//
// Assign it to cmd.Env whenever the child is a git command given an explicit
// directory (via `git -C` or cmd.Dir), or is an arbitrary process that may run
// git of its own.
func Env() []string {
	return Scrub(os.Environ())
}
