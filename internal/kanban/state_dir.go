// state_dir.go — the project-local state directory and its one-time
// relocation from the legacy name (SPEC-TODO-SQLITE-001 REQ-TOSQ-015, M3;
// absorbs card t309).
//
// The directory is named for the command that owns what it holds. `moai todo`
// owns the queue; there is no `moai kanban` command, and there has not been
// one for as long as the queue has existed. `.moai/state/kanban/` therefore
// described nothing a user could type. The rename rides this SPEC because a
// schema redefinition is the one moment it costs nothing: the storage cutover
// already walks every queue root once, under the lock, so the directory move
// costs no second migration.
//
// The per-session registry files (`<uuid>.json`, plus companions.json and
// leads.json) share the directory and ride along by construction — the
// relocation moves the DIRECTORY, not a file list, so nothing can be
// forgotten from an inventory that does not exist.
//
// Three rules govern it, and the third is the one that matters:
//
//	only legacy exists  → relocate the directory, then proceed
//	both exist          → the new name wins; the legacy directory is left
//	                      STRICTLY untouched (stale-copy policy)
//	relocation refused  → serve the legacy layout READ-ONLY and do not error
//
// The third rule is why this file exists rather than a one-line rename. A
// filesystem that cannot relocate the directory — a cross-device mount, a
// permission the operator did not expect — must leave the queue usable, not
// take it away. Failing open here mirrors ResolveTodoQueueRoot's own posture.
package kanban

import (
	"os"
	"path/filepath"
)

// stateDirName is the project-local state directory `moai todo` owns.
const stateDirName = "todo"

// legacyStateDirName is the name it carried before this SPEC. It appears in
// exactly two places — here and the fallback reader below — which is what
// makes the literal-cleanliness sweep (REQ-TOSQ-018) meaningful: any OTHER
// occurrence in production Go is a consumer that was missed.
const legacyStateDirName = "kanban"

// StateDirForRoot returns the project-local state directory under root.
func StateDirForRoot(root string) string {
	return filepath.Join(root, ".moai", "state", stateDirName)
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

// @MX:ANCHOR: [AUTO] resolveStateDir — the directory-layer resolver every queue and registry path enters through
// @MX:REASON: expected fan_in >= 3 (backlog store open, RecordPath, registry helpers); it is the single place the legacy directory is relocated, so a second resolver would let one surface adopt while another still reads the old name
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
	if dirExists(current) {
		// Stale-copy policy: once the new name exists it wins unconditionally,
		// and the legacy directory is left exactly where it is. Leaving it
		// visible is the point — an operator still writing to the dead path
		// can SEE the divergence, where silently absorbing it would hide the
		// mistake until the cards went missing.
		return current, false
	}

	legacyDir := LegacyStateDirForRoot(root)
	if !dirExists(legacyDir) {
		// Neither exists: first run. The new name is created on demand by
		// whichever writer gets there first.
		return current, false
	}

	if !adopt {
		return legacyDir, true
	}
	if err := relocateStateDir(legacyDir, current); err != nil {
		return legacyDir, true
	}
	return current, false
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
	return filepath.Join(dir, backlogFileName)
}

// backlogFileName is the queue document's base name. It stays `backlog.json`
// after the storage swap: the engine derives its own artifact as the sibling
// `backlog.db`, and keeping this name is what makes the downgrade story
// literally true — an older binary reads only this file and ignores the rest.
const backlogFileName = "backlog.json"
