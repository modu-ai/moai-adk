package quality

import (
	"os"
	"strings"
)

// gitRepoScopingEnvVars names the git environment variables that decide WHICH
// repository a git command acts on, as opposed to how it behaves or who it
// attributes work to.
//
// A gate step's child inherits the environment of whatever invoked `moai gate`,
// and the invoker that matters is the git pre-commit hook: git exports GIT_DIR
// and GIT_INDEX_FILE into a hook's environment. Each of these outranks the
// working directory, so cmd.Dir alone does not confine a child — a project's
// own test fixtures, creating throwaway commits in a temp directory, wrote them
// into the repository being committed to instead (GH #1691).
//
// Scope is deliberately narrow. Only repository LOCATION is removed:
//
//   - Identity (GIT_AUTHOR_NAME, GIT_COMMITTER_EMAIL, …) stays. A fixture that
//     lost its identity would fail for an unrelated reason, turning a clean
//     isolation failure into a confusing one.
//   - Behaviour (GIT_EDITOR, GIT_PAGER, GIT_SSH_COMMAND, GIT_CONFIG_GLOBAL, …)
//     stays. A project or CI that sets those means them, and dropping them
//     would change what the gate grades rather than where it grades it.
var gitRepoScopingEnvVars = []string{
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

// scrubGitRepoScoping returns env with every gitRepoScopingEnvVars entry
// removed. Everything else is carried through in order, so a step sees the
// environment it would have seen but cannot be pointed at a repository other
// than the one its working directory names.
//
// The returned slice is always non-nil, including when env is empty: os/exec
// reads a nil Env as "inherit the parent's environment", which is exactly the
// behaviour this function exists to prevent.
func scrubGitRepoScoping(env []string) []string {
	drop := make(map[string]struct{}, len(gitRepoScopingEnvVars))
	for _, name := range gitRepoScopingEnvVars {
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

// stepEnv is the environment a quality-gate step's child process runs with.
func stepEnv() []string {
	return scrubGitRepoScoping(os.Environ())
}
