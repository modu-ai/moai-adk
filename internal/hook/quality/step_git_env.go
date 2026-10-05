package quality

import (
	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// The scrub itself lives in internal/gitenv (card t560). It was defined here
// first (card t516, GH #1691) and closing that defect one package at a time is
// what let the sibling call in this very file keep leaking — so the list moved
// somewhere every caller can reach, and this file kept only its local names.
//
// Behaviour is unchanged; the tests written against these names in t516 still
// bind, which is what makes the move safe to read as a move.

// scrubGitRepoScoping returns env with every gitenv.RepoScopingVars entry
// removed, always as a non-nil slice. See that variable for the scope argument
// — identity and behaviour variables are deliberately kept.
func scrubGitRepoScoping(env []string) []string {
	return gitenv.Scrub(env)
}

// stepEnv is the environment a quality-gate step's child process runs with.
func stepEnv() []string {
	return gitenv.Env()
}
