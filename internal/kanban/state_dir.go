// state_dir.go resolves the Todo queue to
// ~/.moai/db/<project-key>/todo/backlog.db. Temporary test projects remain
// project-local so tests cannot touch the operator's home. Legacy project-local
// queues are copied through the logical store and read back before the global
// database becomes canonical; the source remains a rollback snapshot.
package kanban

import (
	"context"
	"os"
	"path/filepath"
	"reflect"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// stateDirName is the project-local state directory `moai todo` owns.
const stateDirName = "todo"

// legacyStateDirName is the name it carried before this SPEC. It appears in
// exactly two places — here and the fallback reader below — which is what
// makes the literal-cleanliness sweep (REQ-TOSQ-018) meaningful: any OTHER
// occurrence in production Go is a consumer that was missed.
const legacyStateDirName = "kanban"

// StateDirForRoot returns the canonical Todo state directory for root.
func StateDirForRoot(root string) string {
	local := projectStateDirForRoot(root)
	override := os.Getenv(paths.EnvHome)
	if _, temporary := TempOriginReason(root); temporary && (override == "" || !filepath.IsAbs(override)) {
		return local
	}
	home, err := HomeDirFn()
	if err != nil || home == "" {
		return local
	}
	moaiHome := filepath.Join(home, ".moai")
	if override != "" && filepath.IsAbs(override) {
		moaiHome = override
	}
	return filepath.Join(moaiHome, "db", homestate.ProjectKey(root), "todo")
}

func legacyHomeStateDirsForRoot(root string) []string {
	moaiHome, err := paths.MoaiHome()
	if err != nil {
		return nil
	}
	keys := []string{homestate.ProjectKey(root), legacyTodoQueueProjectKey(root)}
	seen := map[string]bool{}
	var dirs []string
	for _, key := range keys {
		if seen[key] {
			continue
		}
		seen[key] = true
		base := filepath.Join(moaiHome, "todo", key)
		dirs = append(dirs, base, filepath.Join(base, ".moai", "state", stateDirName), filepath.Join(base, ".moai", "state", legacyStateDirName))
	}
	return dirs
}

func projectStateDirForRoot(root string) string {
	return filepath.Join(root, ".moai", "state", stateDirName)
}

// RuntimeStateDirForRoot is the compatibility home for session registries
// that have not yet moved to the factory database. It must never be used for
// backlog persistence.
func RuntimeStateDirForRoot(root string) string {
	current := projectStateDirForRoot(root)
	if dirExists(current) {
		return current
	}
	legacy := LegacyStateDirForRoot(root)
	if dirExists(legacy) {
		return legacy
	}
	return current
}

// LegacyStateDirForRoot returns the pre-rename directory under root. It is
// the fallback reader's subject and the relocation's source; nothing writes
// through it.
func LegacyStateDirForRoot(root string) string {
	return filepath.Join(root, ".moai", "state", legacyStateDirName)
}

// dirExists reports whether path names an existing directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func queueExists(dir string) bool {
	layout := inspectBacklogLayout(filepath.Join(dir, backlogFileName))
	return layout.dbExists || layout.jsonExists
}

// @MX:ANCHOR: [AUTO] resolveStateDir — the directory-layer resolver every queue path enters through
// @MX:REASON: it is the single place a legacy queue is adopted, so a second resolver would let readers observe different backlogs
//
// resolveStateDir returns the state directory to USE under root, and whether
// the answer is the legacy one.
//
// When adopt is false the resolution is PURE: it observes which directory
// exists and moves nothing. That is the read-only surfaces' path (the console,
// the statusline) — a page render must never perform a one-time irreversible
// relocation.
//
// When adopt is true and only the legacy directory exists, the relocation is
// attempted. A refused relocation is NOT an error: the legacy directory is
// returned and the caller serves from it, which keeps the queue usable on a
// filesystem that cannot rename across whatever boundary sits between the two
// names.
func resolveStateDir(root string, adopt bool) (dir string, legacy bool) {
	current := StateDirForRoot(root)
	// A global Todo directory may already exist because factory/handoff startup
	// materialized the home layout. It wins only when it contains a queue;
	// otherwise a project-local legacy queue still needs adoption.
	if (current == projectStateDirForRoot(root) && dirExists(current)) || queueExists(current) {
		// Stale-copy policy: once the new name exists it wins unconditionally,
		// and the legacy directory is left exactly where it is. Leaving it
		// visible is the point — an operator still writing to the dead path
		// can SEE the divergence, where silently absorbing it would hide the
		// mistake until the cards went missing.
		return current, false
	}

	legacyDir := ""
	for _, candidate := range append([]string{projectStateDirForRoot(root), LegacyStateDirForRoot(root)}, legacyHomeStateDirsForRoot(root)...) {
		if candidate != current && queueExists(candidate) {
			legacyDir = candidate
			break
		}
	}
	if legacyDir == "" {
		// Neither exists: first run. The new name is created on demand by
		// whichever writer gets there first.
		return current, false
	}

	if !adopt {
		return legacyDir, true
	}
	if current == projectStateDirForRoot(root) {
		if err := relocateStateDir(legacyDir, current); err != nil {
			return legacyDir, true
		}
		return current, false
	}
	if err := relocateQueueArtifacts(legacyDir, current); err != nil {
		return legacyDir, true
	}
	return current, false
}

// relocateQueueArtifacts moves only Todo-owned files into the global project
// database directory. Session registries that happened to share the old
// project directory are deliberately left for the Factory migration.
func relocateQueueArtifacts(from, to string) (err error) {
	sourceQueue := filepath.Join(from, backlogFileName)
	sourceStore := NewBacklogStore(sourceQueue)
	lock, err := sourceStore.acquireLock()
	if err != nil {
		return err
	}
	defer func() {
		err = joinBacklogReleaseErr(err, lock.Release(), sourceQueue)
	}()

	if err := os.MkdirAll(to, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(to, 0o700); err != nil {
		return err
	}
	targetQueue := filepath.Join(to, backlogFileName)
	targetStore := NewBacklogStore(targetQueue)
	targetLock, err := targetStore.acquireLock()
	if err != nil {
		return err
	}
	defer func() {
		err = joinBacklogReleaseErr(err, targetLock.Release(), targetQueue)
	}()

	// Another adopting process may have completed the copy while this caller
	// waited on the legacy lock. The first verified home database wins; a late
	// migrator must not overwrite cards already added there.
	if queueExists(to) {
		return nil
	}
	sourceLayout := inspectBacklogLayout(sourceQueue)
	if !sourceLayout.dbExists && !sourceLayout.jsonExists {
		return os.ErrNotExist
	}
	record, err := sourceStore.LoadPure()
	if err != nil {
		return err
	}
	targetEngine, err := openBacklogEngine(backlogSQLitePath(targetQueue))
	if err != nil {
		return err
	}
	migrationComplete := false
	defer func() {
		if !migrationComplete {
			removeBacklogDBArtifacts(backlogSQLitePath(targetQueue))
		}
	}()
	if err := targetEngine.writeRecord(context.Background(), record); err != nil {
		_ = targetEngine.close()
		return err
	}
	readback, err := targetEngine.readRecord(context.Background())
	closeErr := targetEngine.close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if !reflect.DeepEqual(record, readback) {
		return os.ErrInvalid
	}
	migrationComplete = true
	// The source remains as a rollback snapshot. Once the explicit migration
	// command archives it, the canonical directory already wins every read.
	return nil
}

// relocateStateDir renames the legacy directory to the current one, creating
// the parent as needed. Same-volume renames are atomic, so the registry files
// and the queue arrive together or not at all — there is no window in which a
// session record exists under one name and the queue under the other.
//
// Failure is returned rather than swallowed so the caller can fall back; no
// path here copies, merges, or deletes. A relocation that cannot be a rename
// is refused, because a copy-then-delete would introduce exactly the partial
// state the atomic rename exists to prevent.
func relocateStateDir(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	return os.Rename(from, to)
}

// BacklogPathForRoot returns the backlog queue file's canonical location under
// a project root — the one path shape every root-relative consumer builds its
// store from, so no two surfaces hand-roll their own join.
//
// PURE: it computes a path and observes which directory exists, but relocates
// nothing. The relocation belongs to the adopting open path, where the queue
// lock is already in play.
func BacklogPathForRoot(root string) string {
	dir, _ := resolveStateDir(root, false)
	return filepath.Join(dir, backlogFileName)
}

// BacklogPathForRootAdopting is BacklogPathForRoot plus the one-time directory
// relocation — the `moai todo` command path's form, mirroring the existing
// ResolveTodoQueueRoot / ResolveTodoQueueRootAdopting split one layer down.
func BacklogPathForRootAdopting(root string) string {
	dir, _ := resolveStateDir(root, true)
	if StateDirForRoot(root) != projectStateDirForRoot(root) {
		_ = homestate.EnsureProjectLayout(root)
	}
	return filepath.Join(dir, backlogFileName)
}

// backlogFileName is the queue document's base name. It stays `backlog.json`
// after the storage swap: the engine derives its own artifact as the sibling
// `backlog.db`, and keeping this name is what makes the downgrade story
// literally true — an older binary reads only this file and ignores the rest.
const backlogFileName = "backlog.json"
