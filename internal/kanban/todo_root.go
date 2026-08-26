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
//     WriteFile on ANY of its three branches. It is what internal/web
//     imports: a console that rendered a page must not migrate the
//     operator's backlog as a side effect (REQ-WTQ-001, REQ-WTQ-004).
//   - ResolveTodoQueueRootAdopting resolves and then adopts. It is what the
//     `moai todo` command path calls, so that command's behaviour — the
//     adopt-not-shadow migration — is unchanged (REQ-WTQ-004, AC-WTQ-008).
//
// Read-through (decision D-2, REQ-WTQ-005): splitting alone would trade a
// write hazard for a read divergence — in a non-git launch context the pure
// resolver would report an empty fallback queue while `moai todo` adopted
// first and reported N. The pure resolver therefore resolves to the
// PROJECT-LOCAL root when the fallback root holds no queue file and a
// project-local one exists, still writing nothing. Its predicate is the
// mirror of adoptLocalTodoQueue's own early returns, so the two agree by
// construction rather than by coincidence.
package kanban

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
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
// keyed under ~/.moai/todo/ so two such projects cannot collide.
func ResolveTodoQueueRoot(base string) string {
	if root, ok := primaryCheckoutRoot(base); ok {
		return root
	}
	return fallbackTodoQueueRoot(base)
}

// ResolveTodoQueueRootAdopting is ResolveTodoQueueRoot plus the queue
// adoption `moai todo` has always performed: on the home-based fallback
// branch it carries a pre-existing project-local queue over before returning
// (adopt-not-shadow). This is the ONLY entry point from which the adoption
// side effect is reachable.
func ResolveTodoQueueRootAdopting(base string) string {
	if root, ok := primaryCheckoutRoot(base); ok {
		return root
	}
	root, ok := homeTodoQueueRoot(base)
	if !ok {
		return root
	}
	adoptLocalTodoQueue(base, root)
	// Adoption is best-effort, so it can leave the fallback root empty while the
	// project-local queue still holds the operator's cards. Resolving through the
	// pure resolver afterwards is what keeps this path and the console's agreeing
	// in that case too (REQ-WTQ-005): a successful adoption makes the fallback
	// root the populated one and it wins, and a failed adoption reads through to
	// the local queue rather than reporting it empty.
	return fallbackTodoQueueRoot(base)
}

// primaryCheckoutRoot resolves base to the repository's primary checkout,
// reporting false when git cannot answer. Read-only.
func primaryCheckoutRoot(base string) (string, bool) {
	if dirs, err := gitcore.ResolveGitDirs(base); err == nil && dirs.CommonDir != "" {
		return filepath.Dir(dirs.CommonDir), true
	}
	return "", false
}

// homeTodoQueueRoot returns the home-based queue root for base, reporting
// false when no home is resolvable — in which case it returns the in-project
// root instead, keeping the queue usable rather than failing the caller
// outright. Read-only.
//
// The directory is named for the command that owns the queue (`moai todo` —
// no `moai kanban` command exists), deliberately NOT "kanban": the
// .moai/state/kanban/ directory also holds per-session kanban records
// (<uuid>.json), and a "kanban" fallback name would read as moving those too
// — a scope this queue never touches.
func homeTodoQueueRoot(base string) (string, bool) {
	if base == "" {
		base = "."
	}
	home, err := HomeDirFn()
	if err != nil {
		return filepath.Join(base, ".moai", "state", "kanban"), false
	}
	return filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(base)), true
}

// fallbackTodoQueueRoot is the PURE fallback: it computes the home-based
// root and, per decision D-2, reads through to the project-local root when
// that home root holds no queue file while a project-local one exists. It
// writes nothing on any path.
func fallbackTodoQueueRoot(base string) string {
	if base == "" {
		base = "."
	}
	root, ok := homeTodoQueueRoot(base)
	if !ok {
		return root
	}
	// The mirror of adoptLocalTodoQueue's early returns: a populated fallback
	// wins (it was adopted on an earlier run), and with no local file there is
	// nothing to read through to.
	if _, err := os.Stat(BacklogPathForRoot(root)); err == nil {
		return root
	}
	if _, err := os.Stat(BacklogPathForRoot(base)); err != nil {
		return root
	}
	return base
}

// adoptLocalTodoQueue moves a pre-existing project-local queue into a fresh
// fallback root, so the fallback's first run adopts the project's cards
// instead of shadowing them behind an empty queue (the lossless-migration
// requirement: item count and states must survive the cutover).
//
// Best-effort throughout — a failure leaves the local queue exactly where it
// was and the fallback simply starts empty THIS run; the data is never
// destroyed, so a later run can adopt it again once the obstruction clears.
// When the fallback already has a queue file the local one is left untouched
// (adopted on an earlier run; a local file reappearing after that is a
// downgrade-era snapshot the populated fallback deliberately ignores).
func adoptLocalTodoQueue(base, fallbackRoot string) {
	local := BacklogPathForRoot(base)
	// BacklogPathForRoot, not a bare join: every consumer resolves the store
	// through it (internal/cli/todo.go, internal/web/todo_queue_read.go), so a
	// target built any other way is a path nothing reads — the queue would be
	// moved out of the operator's sight rather than migrated.
	target := BacklogPathForRoot(fallbackRoot)
	if _, err := os.Stat(target); err == nil {
		return
	}
	if _, err := os.Stat(local); err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return
	}
	// Same-volume rename is atomic and leaves no duplicate behind.
	if err := os.Rename(local, target); err == nil {
		return
	}
	// Cross-volume (EXDEV) or rename-refusing filesystem: copy the bytes and
	// KEEP the original — deletion is the one outcome the lossless
	// requirement forbids, and a leftover original is inert (the populated
	// fallback wins on every later run).
	data, err := os.ReadFile(local)
	if err != nil {
		return
	}
	_ = os.WriteFile(target, data, 0o600)
}

// TodoQueueProjectKey derives the fallback queue's directory name from the
// launch directory: a readable base name plus a short digest of the absolute
// path. Two distinct projects sharing a base name still occupy two keys, and
// the mapping is deterministic across runs.
func TodoQueueProjectKey(base string) string {
	abs, err := filepath.Abs(base)
	if err != nil {
		abs = filepath.Clean(base)
	}
	sum := sha256.Sum256([]byte(abs))
	return fmt.Sprintf("%s-%x", filepath.Base(abs), sum[:4])
}
