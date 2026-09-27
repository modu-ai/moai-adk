package homestate

import (
	"fmt"
	"path/filepath"
)

// RefuseMutationFromNonCanonicalTree stops a state mutation whose caller tree
// is not the tree the mutation would actually reach.
//
// CanonicalProjectRoot resolves a linked worktree to the root every worktree of
// the repository shares on purpose — one repository keeps one queue, and the
// todo queue depends on that. That root is the primary checkout in an ordinary
// repository and the git directory itself in a --separate-git-dir, bare, or
// submodule one. MOAI_HOME redirects only the TARGET. A caller working from a
// worktree with an isolated home therefore mutates the SHARED root's live state
// while believing both ends are isolated, and nothing in the result contradicts
// that belief.
//
// t952 closed this for the one caller that goes through the CLI runner. This is
// the same gate stated where the hazard lives, so a caller that never passes
// through that runner is refused too. Refusing is the whole repair: the
// operator who wants the canonical project's mutation runs it from the
// canonical root, and that one change of directory is what makes the intent
// explicit.
//
// The gate is deliberately narrow. It binds mutation entry points only; reads,
// dry-runs, and recovery paths are how this situation is discovered and
// escaped. Reads and dry-runs mutate nothing. Recovery paths DO mutate — the
// recover path quarantines the live backlog DB by renaming it, restores it from
// a backup, and clears the migration marker, none of which its "recover a
// dead-owner migration marker" surface announces. Their exemption therefore
// rests on the escape-hatch argument alone, never on a no-mutation claim.
// Outside a repository — an ordinary temp directory, say —
// CanonicalProjectRoot returns the caller's own path, so the gate is inert.
func RefuseMutationFromNonCanonicalTree(callerRoot string) error {
	caller, err := filepath.Abs(callerRoot)
	if err != nil {
		return nil // Unresolvable caller path: leave the later gates to judge.
	}
	if resolved, err := filepath.EvalSymlinks(caller); err == nil {
		caller = resolved
	}
	canonical := CanonicalProjectRoot(callerRoot)
	if filepath.Clean(caller) == filepath.Clean(canonical) {
		return nil
	}
	return fmt.Errorf(
		"refused: this tree (%s) is not the canonical root the mutation reaches (%s). "+
			"MOAI_HOME isolates the target only, so mutating here would move the canonical root's live state. "+
			"Run it from the canonical root",
		caller, canonical)
}
