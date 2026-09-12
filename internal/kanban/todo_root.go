// todo_root.go — SPEC-WEB-TODO-QUEUE-001 M1: the backlog queue-root
// resolution, relocated here from internal/cli so the command layer and the
// web console share ONE resolution. A second implementation is a second
// chance to fork the queue, which is the measured failure this resolution
// exists to prevent (30 queued cards on the primary checkout, "queue is
// empty" from a linked worktree — 2026-08-17).
//
// The relocation SPLITS resolving from adopting, because the two callers
// need different things:
//
//   - ResolveTodoQueueRoot is PURE. It performs no MkdirAll, Rename, or
//     WriteFile on ANY branch. It is what internal/web imports: a console
//     that rendered a page must not migrate the operator's backlog as a side
//     effect (REQ-WTQ-001, REQ-WTQ-004).
//   - ResolveTodoQueueRootAdopting is what the `moai todo` command path
//     calls, so that command's behaviour — the adopt-not-shadow migration —
//     is unchanged (REQ-WTQ-004, AC-WTQ-008).
//
// This layer answers ONE question: which project root does the queue hang
// from. Everything ABOUT that root — the redirection into ~/.moai/db/<key>/
// todo, the temporary-origin refusal that keeps a temp launch project-local,
// and the adoption of every legacy location including this file's own former
// ~/.moai/todo/<key> fallback — belongs to resolveStateDir one layer down,
// which is anchored as the single directory-layer resolver precisely so two
// copies of that policy cannot drift apart (state_dir.go @MX:ANCHOR).
//
// A home-based fallback ROOT used to be computed here as well, and that was
// the drift (t621). Since the home-state migration, the layer below re-derives
// the home location from whatever root it is handed — so handing it a home
// root got the project key computed from a home path, landing the queue in
// ~/.moai/db/<key>-<hash>/todo. A session that reached the same project
// through git read ~/.moai/db/<key>/todo and saw an empty queue: the fork this
// resolution exists to prevent, reintroduced one layer up. The launch base is
// therefore the answer on every non-git branch, and it is the ONLY root whose
// project key is the project's own.
//
// The same drift made the read-through predicate (decision D-2, REQ-WTQ-005)
// unserviceable: it asked os.Stat for a `backlog.json` while the engine's
// steady state is a sibling `backlog.db` with no json beside it, so a real
// queue in the normal layout was invisible to it. Both predicates are gone
// with the branch they guarded; resolveStateDir observes layouts through
// queueExists, which reads both artifacts.
package kanban

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// HomeDirFn is this package's home-directory injection seam. The relocated
// fallback used internal/cli's package-level userHomeDirFn seam, which nine
// non-test files there share and which therefore could not travel with the
// code; this is its equivalent here, and without it the fallback root is
// uncontrollable in tests (AC-WTQ-006, AC-WTQ-007). internal/cli rewires it
// to its own seam in init(), so a test overriding either one still reaches
// this resolution.
var HomeDirFn = paths.Home

// ResolveTodoQueueRoot returns the directory the backlog queue hangs from,
// WITHOUT touching the filesystem on any branch: the PRIMARY checkout of the
// repository the launch context sits in, or a home-based fallback when git
// cannot answer.
//
// The queue is the delegation channel between sessions — the lead, the
// foreman loop, and the operator's picks all read one queue. A linked
// worktree holding its own .moai/state would fork that channel, so the
// primary is resolved through the repository itself: git's common directory
// is shared by every checkout and its parent IS the primary checkout's root,
// from any worktree and from the primary alike.
//
// Fail-open direction: an unresolvable git context (no git binary, not a
// repository) keeps the queue usable via the home-based fallback rather than
// erroring — a project without git metadata still gets exactly one queue,
// keyed under ~/.moai/db/ so two such projects cannot collide.
func ResolveTodoQueueRoot(base string) string {
	if root, ok := primaryCheckoutRoot(base); ok {
		return root
	}
	return base
}

// ResolveTodoQueueRootAdopting is the `moai todo` command path's entry point.
// It resolves to the same root as the pure form; the adoption that used to
// live here now happens one layer down, in BacklogPathForRootAdopting, where
// the queue lock is already in play.
//
// The two entry points are kept distinct because their CALLERS' contracts
// differ: the console must reach a resolution that provably writes nothing
// (REQ-WTQ-001, REQ-WTQ-004), and collapsing the names would let a later edit
// give the console an adopting path without any caller changing.
func ResolveTodoQueueRootAdopting(base string) string {
	return ResolveTodoQueueRoot(base)
}

func pathInsideTempDir(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(os.TempDir(), abs)
	return err == nil && rel != ".." && rel != "." && len(rel) > 0 && rel[0] != filepath.Separator && (len(rel) <= 2 || rel[:3] != ".."+string(filepath.Separator))
}

// tempOriginSubstituteRoot is the temporary-origin guard
// (SPEC-TODO-HOME-TEMP-GUARD-001, REQ-THG-001): when the launch base is
// classified as a temporary-directory origin, no home queue is resolved to and
// none is created; the substitute root is the LAUNCH BASE itself.
//
// Why the base and not resolveStateDir(base, false): every consumer appends
// BacklogPathForRoot to whatever this resolution returns (internal/cli/todo.go,
// internal/web/todo_queue_read.go), and resolveStateDir already IS the state
// directory — returning it would send readers one level too deep, to the state
// directory nested inside itself, which is a location nothing ever writes.
// The property this substitute must satisfy is therefore stated on the reader's
// side: BacklogPathForRoot(<returned root>) names the project's existing
// project-local backlog file. "" is likewise forbidden — the console puts the
// returned root straight into vm.Root and would read a cwd-relative path.
//
// Placement is load-bearing, not stylistic (REQ-THG-001, plan D12): the
// discriminant is evaluated BEFORE the home-resolution outcome is consulted.
// "Temporary origin" and "home unresolvable" are not mutually exclusive — a
// non-git temp base whose HomeDirFn errors satisfies both — and on that
// intersection this requirement fixes the return at the base. homeTodoQueueRoot's
// no-home return used to be resolveStateDir(base, false), one layer below the
// root, so ordering alone would have decided the value; since t549 it returns
// the base as well, and the placement keeps the guard's guidance path first.
//
// Read-only: TempOriginReason performs Lstat/EvalSymlinks and nothing else.
func tempOriginSubstituteRoot(base string) (root, matchedRoot string, refused bool) {
	if base == "" {
		base = "."
	}
	matchedRoot, isTemp := TempOriginReason(base)
	if !isTemp {
		return "", "", false
	}
	return base, matchedRoot, true
}

// TempOriginRefusal reports whether queue-root resolution for base refuses the
// home queue on temporary-origin grounds, naming both the matched temp root and
// the substitute root the caller will keep using.
//
// It exists so the COMMAND path can surface guidance (REQ-THG-006) without the
// resolvers themselves acquiring an output channel: the pure resolver stays
// silent on every branch, which is what keeps a console page render free of
// side effects and messages (REQ-THG-007, AC-THG-005). Read-only, like the
// resolution it mirrors — it consults the same git branch first, so a git
// repository living under a temp root is reported NOT refused, exactly as the
// resolvers treat it.
func TempOriginRefusal(base string) (substitute, matchedRoot string, refused bool) {
	if _, ok := primaryCheckoutRoot(base); ok {
		return "", "", false
	}
	if explicitMoaiHome() {
		return "", "", false
	}
	return tempOriginSubstituteRoot(base)
}

func explicitMoaiHome() bool {
	value := os.Getenv(paths.EnvHome)
	return value != "" && filepath.IsAbs(value)
}

// primaryCheckoutRoot resolves base to the repository's primary checkout,
// reporting false when git cannot answer. Read-only.
func primaryCheckoutRoot(base string) (string, bool) {
	if dirs, err := gitcore.ResolveGitDirs(base); err == nil && dirs.CommonDir != "" {
		return filepath.Dir(dirs.CommonDir), true
	}
	return "", false
}

// homeTodoQueueRoot names the LEGACY home-based queue root for base —
// ~/.moai/todo/<key> — reporting false when no home is resolvable.
//
// It is no longer a resolution target (t621): a root resolved here is re-keyed
// by the layer below, so returning one forked the queue. What remains is the
// location's identity, which is still load-bearing on the adoption side —
// legacyHomeStateDirsForRoot lists this directory and its nested state dirs
// among the sources resolveStateDir adopts a queue FROM, so an operator whose
// cards were carried into the old fallback still gets them back.
//
// The directory is named for the command that owns the queue (`moai todo` —
// no `moai kanban` command exists). Read-only.
func homeTodoQueueRoot(base string) (string, bool) {
	if base == "" {
		base = "."
	}
	home, err := HomeDirFn()
	if err != nil {
		// No home: fall back to the launch base. It is a ROOT like every other
		// value this resolution returns — consumers extend it with
		// BacklogPathForRoot, which resolves the project-local state directory
		// itself, purely. Returning that state directory here got it extended a
		// second time, and the read landed on a path nothing writes (t549).
		return base, false
	}
	return filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(base)), true
}

// TodoQueueProjectKey derives the fallback queue's directory name from the
// launch directory: a readable base name plus a short digest of the absolute
// path. Two distinct projects sharing a base name still occupy two keys, and
// the mapping is deterministic across runs.
func TodoQueueProjectKey(base string) string {
	return homestate.ProjectKey(base)
}

// legacyTodoQueueProjectKey reproduces the pre-home-state key byte-for-byte
// so on-access migration can find queues created before canonicalization and
// filename sanitization were introduced.
func legacyTodoQueueProjectKey(base string) string {
	abs, err := filepath.Abs(base)
	if err != nil {
		abs = filepath.Clean(base)
	}
	sum := sha256.Sum256([]byte(abs))
	return fmt.Sprintf("%s-%x", filepath.Base(abs), sum[:4])
}
