// Package skillresidue repairs the skill-folder leftovers that earlier MoAI
// versions and interrupted deploys leave under .claude/skills: an empty folder of
// a skill the template ships, a retired skill name, and a lowercase skill.md. A
// change is a removal of an empty folder or a rename. A retired skill folder and
// the link to it are reported and left in place. Each action is returned so the
// update can report it.
package skillresidue

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
)

// retiredSkills maps a skill folder name MoAI no longer ships to the name that
// replaced it. The retired folder is reported only when the template ships the
// successor.
var retiredSkills = map[string]string{
	"moai-kanban-foreman": "moai-factory-foreman",
}

// Options are the inputs the cleanup takes from its caller.
type Options struct {
	// Shipped reports whether the embedded template ships a project-relative path
	// such as ".claude/skills/moai-factory-foreman/SKILL.md". A nil Shipped ships
	// nothing, so no folder is removed or renamed on that basis.
	Shipped func(rel string) bool
	// SkipRenames holds back the lowercase skill.md rename. A caller whose update the
	// user may still cancel runs ApplyDeferredRenames after the confirmation instead.
	SkipRenames bool
}

func (o Options) shipped(rel string) bool {
	return o.Shipped != nil && o.Shipped(rel)
}

// Action is one residue the cleanup acted on or reported.
type Action struct {
	// Path is the project-relative, slash-separated path the residue was found at.
	Path string
	// Kind is one of removed-empty-dir, renamed-skill-file,
	// reported-both-skill-files, reported-retired-skill,
	// kept-retired-skill-with-user-files, reported-retired-link, or
	// refused-symlinked-root.
	Kind string
}

// Clean repairs the skill-folder residue under projectRoot/.claude/skills and
// returns what it did. Retired folders, their links, and residue left in place are
// reported and not changed, so a second pass changes nothing on disk and returns
// again only those reports.
func Clean(projectRoot string, opts Options) ([]Action, error) {
	if link, found := symlinkedRoot(projectRoot); found {
		return []Action{{Path: link, Kind: "refused-symlinked-root"}}, nil
	}
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// The one move this pass can make runs through a handle bound to the project
	// root, so a folder swapped for a symbolic link after the checks below cannot
	// lead it outside the project.
	root, err := os.OpenRoot(projectRoot)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	var actions []Action
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		dir := filepath.Join(skillsDir, name)
		rel := ".claude/skills/" + name

		if successor, retired := retiredSkills[name]; retired {
			if !opts.shipped(".claude/skills/" + successor + "/SKILL.md") {
				continue
			}
			if userFile(dir) != "" {
				actions = append(actions, Action{Path: rel, Kind: "kept-retired-skill-with-user-files"})
				continue
			}
			// The retired folder and the link to it stay where they are. The update
			// reports both and moves or removes neither.
			actions = append(actions, Action{Path: rel, Kind: "reported-retired-skill"})
			linked, err := retiredLinkPresent(projectRoot, name)
			if err != nil {
				return actions, err
			}
			if linked {
				actions = append(actions, Action{Path: ".agents/skills/" + name, Kind: "reported-retired-link"})
			}
			continue
		}

		names, err := entryNames(dir)
		if err != nil {
			return actions, err
		}
		switch {
		case len(names) == 0:
			// An empty folder is residue only for a skill the template ships; the
			// deploy refills it. Any other empty folder belongs to the user.
			if opts.shipped(rel + "/SKILL.md") {
				if err := os.Remove(dir); err != nil {
					return actions, err
				}
				actions = append(actions, Action{Path: rel, Kind: "removed-empty-dir"})
			}
		case slices.Contains(names, "skill.md"):
			if slices.Contains(names, "SKILL.md") {
				// Both names present: never rename over or next to the canonical
				// file. Report the pair for a person to resolve.
				actions = append(actions, Action{Path: rel, Kind: "reported-both-skill-files"})
				continue
			}
			// A directory named skill.md belongs to the user, and its contents are not a
			// skill file. Keep it, and report it.
			if !regularFileAt(filepath.Join(dir, "skill.md")) {
				actions = append(actions, Action{Path: rel + "/skill.md", Kind: "kept-non-file-skill-md"})
				continue
			}
			if opts.SkipRenames {
				continue
			}
			if opts.shipped(rel + "/SKILL.md") {
				if err := renameSkillFile(root, rel); err != nil {
					return actions, err
				}
				actions = append(actions, Action{Path: rel + "/skill.md", Kind: "renamed-skill-file"})
			}
		}
	}
	return actions, nil
}

// entryNames lists a folder's entry names as stored. The stored name is the only
// reliable one on a case-insensitive filesystem.
func entryNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

// userFile returns the first entry of a retired folder that is not its SKILL.md,
// or "" when the folder holds only SKILL.md. An unreadable folder counts as holding
// user content, so it is kept.
func userFile(dir string) string {
	names, err := entryNames(dir)
	if err != nil {
		return "unreadable"
	}
	for _, n := range names {
		if n != "SKILL.md" {
			return n
		}
	}
	return ""
}

// retiredLinkPresent reports whether .agents/skills/<name> is a symbolic link to the
// retired folder .claude/skills/<name>. It only inspects the link and never removes
// it. A link that points anywhere else, or a real folder in that place, is not
// reported.
func retiredLinkPresent(projectRoot, name string) (bool, error) {
	link := filepath.Join(projectRoot, ".agents", "skills", name)
	info, err := os.Lstat(link)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}
	target, err := os.Readlink(link)
	if err != nil {
		return false, err
	}
	resolved := target
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(link), resolved)
	}
	return filepath.Clean(resolved) == filepath.Join(projectRoot, ".claude", "skills", name), nil
}

// The held-back rename is a few steps, each behind a named function and a package variable.
// The package never calls the handle's Rename; the only place it calls the handle's Remove is
// dropOne, so the registry's destructive-target scan sees one named site, and the pin test
// counts every reference to the seams so an alias is visible. A test replaces a variable to
// make one step fail, or to run a second actor between two steps.

// mkdirOne creates a folder inside root and fails when the name exists. That failure is
// what makes a temporary name a claim and not a guess.
func mkdirOne(root *os.Root, name string) error { return root.Mkdir(name, 0o755) }

// linkOne makes to a second name for the file at from. It fails when to exists, so it
// can place a name where Rename could only replace one.
func linkOne(root *os.Root, from, to string) error { return root.Link(from, to) }

// dropOne removes one empty folder or one name of a file that has another name. It
// never removes the last name of content (see renameByLinks).
func dropOne(root *os.Root, name string) error { return root.Remove(name) }

// errNoReplaceUnsupported is returned by noReplaceRenameAt when the platform or the filesystem
// cannot rename without replacing. Nothing has moved when it is returned.
var errNoReplaceUnsupported = errors.New("no-replace rename is not supported here")

// randomStageSuffix returns 12 hex characters for a staging folder name.
func randomStageSuffix() string {
	var b [6]byte
	_, _ = rand.Read(b[:]) // never fails since Go 1.24 (a failure ends the program)
	return hex.EncodeToString(b[:])
}

var (
	mkdirFile = mkdirOne
	linkFile  = linkOne
	dropFile  = dropOne
	// renameNoReplace renames src in srcDir to dst in dstDir (both open folders) and fails
	// when dst exists, in one atomic step; noReplaceRenameAt is the per-platform
	// implementation, relative to the folders' descriptors where the platform has them.
	renameNoReplace = noReplaceRenameAt
	// stageSuffix returns the random part of a staging folder name; an empty string selects
	// the deterministic names the link-fallback tests address by literal path.
	stageSuffix = randomStageSuffix
)

// ApplyDeferredRenames performs the lowercase skill.md renames that Clean holds back
// under SkipRenames. It touches only shipped skill folders whose skill.md is a regular
// file, and it skips retired folders, as Clean does. It runs after the update is
// confirmed, so a cancelled update leaves the names as they were.
func ApplyDeferredRenames(projectRoot string, opts Options) ([]Action, error) {
	if link, found := symlinkedRoot(projectRoot); found {
		return []Action{{Path: link, Kind: "refused-symlinked-root"}}, nil
	}
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(projectRoot)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	var actions []Action
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if _, retired := retiredSkills[name]; retired {
			continue
		}
		dir := filepath.Join(skillsDir, name)
		rel := ".claude/skills/" + name
		names, err := entryNames(dir)
		if err != nil {
			return actions, err
		}
		if slices.Contains(names, "SKILL.md") || !slices.Contains(names, "skill.md") {
			continue
		}
		if !regularFileAt(filepath.Join(dir, "skill.md")) || !opts.shipped(rel+"/SKILL.md") {
			continue
		}
		if err := renameSkillFile(root, rel); err != nil {
			return actions, err
		}
		actions = append(actions, Action{Path: rel + "/skill.md", Kind: "renamed-skill-file"})
	}
	return actions, nil
}

// regularFileAt reports whether name is a regular file. A symbolic link, a directory,
// or a FIFO at the path is not.
func regularFileAt(name string) bool {
	info, err := os.Lstat(name)
	return err == nil && info.Mode().IsRegular()
}

// heldDir is a folder opened through the project root and kept open for a whole rename. The
// root refuses a path that leaves the project, and both handles name the folder's inode, so
// what is done through them cannot be redirected by swapping the folder's path afterwards:
// root answers fd-relative looks (Lstat), file supplies the descriptor for the renames.
type heldDir struct {
	root *os.Root
	file *os.File
}

// holdDir opens the folder rel (slash-separated, relative to parent) through parent.
func holdDir(parent *os.Root, rel string) (*heldDir, error) {
	r, err := parent.OpenRoot(rel)
	if err != nil {
		return nil, err
	}
	f, err := r.Open(".")
	if err != nil {
		_ = r.Close()
		return nil, err
	}
	return &heldDir{root: r, file: f}, nil
}

func (h *heldDir) close() {
	_ = h.file.Close()
	_ = h.root.Close()
}

// moveEnd is one end of a move: rel is its slash-separated path relative to the project root
// (for the root-bound look and for messages), name its name inside the held folder dir.
type moveEnd struct {
	rel  string
	dir  *heldDir
	name string
}

// renameSkillFile renames <dirRel>/skill.md to <dirRel>/SKILL.md, where root is a handle
// bound to the project root and dirRel is the skill folder's slash-separated path relative
// to that root.
//
// The file moves by atomic no-replace renames, never by a link plus an unlink behind a
// check, so no user-visible name is removed and no step can replace one:
//
//	A  claim a staging folder S beside the file (Mkdir fails when the name exists; the name
//	   carries a random part, so a foreign actor has to guess it), then open the skill folder
//	   and S through the root and keep both open (holdDir)
//	B  skill.md -> S/skill.md   (refuses when S/skill.md exists)
//	C  S/skill.md -> SKILL.md   (refuses when SKILL.md exists)
//	D  rmdir S                  (rmdir only: it cannot take a file or a non-empty folder)
//
// B and C are two steps through a private name. On APFS a direct no-replace rename from
// skill.md to SKILL.md was measured to succeed (sync audit iteration 5, F18), so the staging
// hop is not shown to be necessary there; it stays as the designed route, and whether another
// case-folding filesystem (NTFS) needs it was not measured.
//
// Containment. On linux and darwin the renames and the removal of S are relative to the two
// open folders' descriptors (renameat2 RENAME_NOREPLACE, renameatx_np RENAME_EXCL,
// unlinkat AT_REMOVEDIR): swapping the path of the skill folder or of S after it was opened
// cannot redirect them, because the descriptor names the original inode. Opening goes through
// the root, which refuses a path that leaves the project, so a skill folder or S swapped for a
// link out of the project before it was opened is refused. A root-bound Lstat of the source's
// path still precedes every rename (moveNoReplace): with descriptors it no longer contains the
// move, it makes the step refuse when the path no longer leads to the folder it holds. A
// put-back after such a refusal skips the look on these platforms, because the descriptors still
// name both folders and the path it would look at is what was moved: a stage folder renamed out
// of the project takes the first move with it, and the put-back brings the file home. On windows
// the renames are path-based (MoveFileEx has no folder-relative form in x/sys), so that look is
// the containment, the put-back keeps it, and a microsecond window remains between the look and
// the call.
//
// Identity. The step takes the identity of skill.md right before B (fd-relative Lstat) and
// requires the file at S/skill.md to be that file before C and the file at SKILL.md to be that
// file after C. A replacement landing between B and C is refused without moving it; one landing
// inside C itself cannot be prevented, but is reported. After C the requested path must name the
// same file too (root-bound look of SKILL.md): a skill folder renamed away after B with another
// folder at its path leaves SKILL.md in the moved folder, which the step reports as an error.
//
// States and outcomes (L = skill.md, S = staging folder, T = S/skill.md; every row keeps the
// user's bytes unless it says otherwise):
//
//	step  state                                        action                                user data
//	A     name taken (file or folder)                  try another name; 100 taken: error    nothing changed
//	A     Mkdir fails                                  error                                 nothing changed
//	A/B   S swapped for a regular file                 opening S fails (ENOTDIR); rmdir S    foreign file kept
//	                                                   fails; L stays; error names S
//	A/B   S swapped for a link to a folder inside      opened through the link; B lands T    content not lost
//	      or outside the project                       there only when inside; outside:
//	                                                   the root refuses the open, nothing
//	                                                   moves, the link stays, error names S
//	A/B   S swapped for a foreign folder               B lands T in it; C completes; D       content not lost; the
//	                                                   fails if the folder holds anything    foreign folder keeps
//	                                                   else (error names it)                 its own files
//	A/B   skill folder swapped for a link out of       opening the skill folder is refused   outside untouched
//	      the project before it was opened             by the root; nothing moves; error
//	B     S renamed away after it was opened (to a     B lands in the open S, where it went; file back at L (unix); an empty or foreign
//	      link, a folder, out of the project, or       the look before C refuses (the path   folder may sit where S was; windows: the
//	      nothing at its old path)                     leads elsewhere); the put-back goes   put-back is refused by the look, the file
//	                                                   through the held descriptors, with no stays in the moved folder, error says so
//	                                                   look (unix); error
//	B     the skill folder renamed away before B       B lands in S inside the moved folder; file back as skill.md in the moved folder;
//	                                                   look before C refuses; the put-back   error
//	                                                   goes to the held (moved) skill folder
//	B     a foreign file sits at T                     B refused; L untouched; D fails       foreign file and L kept
//	                                                   (S not empty); error names T
//	B     L replaced or deleted before the identity    B moves whatever is at L (an editor's the user's newest file
//	      is taken                                     atomic save: the new file); a missing becomes SKILL.md
//	                                                   L fails the rename; error
//	B     L replaced between the identity and B        the file at T is not the identity:    the moved file stays at T
//	                                                   C is not run; error names T
//	B     B fails for another reason                   L untouched; rmdir S; error           none lost
//	B     B reports errNoReplaceUnsupported            nothing moved; renameByLinks takes    see renameByLinks
//	                                                   over with the same S
//	B-C   T replaced by a foreign file                 the file at T is not the identity:    the foreign file stays at
//	                                                   C is not run, nothing is put back,    T; the content of L was
//	                                                   S is kept; error names T              overwritten by the foreign
//	                                                                                         write, not by this step
//	B-C   a file appears at SKILL.md (or, on a case-   C refused; D puts T back to L        user's file kept; content
//	      folding filesystem, at skill.md)             (refused if L is taken); error names  back at L, or staged at T
//	                                                   SKILL.md; if D is refused the error   when L is taken
//	                                                   names T
//	C     T replaced inside the step itself            the foreign file moves to SKILL.md;   the foreign file is kept
//	                                                   the file at SKILL.md is not the       at SKILL.md; the
//	                                                   identity: error names SKILL.md        content of L was
//	                                                                                         overwritten by the foreign write
//	C     C fails for another reason                   as the row above                      as above
//	C     the skill folder renamed away after B, a     C runs in the held, moved folder; the SKILL.md is in the moved folder; the decoy
//	      decoy folder put at its old path             root-bound look of SKILL.md at the    is kept; nothing is moved back by path
//	                                                   requested path is not that file:
//	                                                   error says the folder was moved;
//	                                                   nothing is removed
//	C-D   S swapped for a file after C                 D fails (ENOTDIR); error names S      foreign file kept
//	C-D   S swapped for a non-empty folder             D fails (ENOTEMPTY); error names S    foreign files kept
//
// A refused put back leaves the content where it is and the error names the place: the staged
// path T while the path of S still leads to the folder that was moved into, otherwise "inside
// the folder this step opened as S, which was moved" (the step cannot name the moved folder's
// new path). No step ever removes a file it did not just move. A microsecond window therefore
// remains only (1) in renameByLinks (the fallback below), (2) between the identity check before
// C and C, which the check after C reports, (3) between the look and the call on platforms whose
// rename is path-based (windows), where a folder renamed away between the look and the call
// also takes the file and the put-back is refused by the look.
//
// @MX:WARN: [AUTO] filesystem moves inside a user's skill folder
// @MX:REASON: a path swapped between a check and a move is the F24/F25 hazard; every move is
// a no-replace rename relative to open folders, no user-visible name is unlinked, and the
// staging folder is removed by rmdir, which cannot take a file.
func renameSkillFile(root *os.Root, dirRel string) error {
	lower := path.Join(dirRel, "skill.md")
	canonical := path.Join(dirRel, "SKILL.md")
	orig, err := root.Lstat(lower)
	if err != nil {
		return err
	}
	if !orig.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file; nothing renamed", lower)
	}

	skill, err := holdDir(root, dirRel)
	if err != nil {
		return err
	}
	defer skill.close()
	stage, stageInfo, err := claimStage(root, dirRel)
	if err != nil {
		return err
	}
	stageDir, err := holdDir(root, stage)
	if err != nil {
		// S was swapped before it could be opened (a file, a link out of the project): nothing
		// moved. rmdir takes S back only while it is still our empty folder.
		return errors.Join(fmt.Errorf("staging folder %s cannot be opened: %w", stage, err), removeStage(skill, stage))
	}
	defer stageDir.close()
	staged := path.Join(stage, "skill.md")
	from := moveEnd{rel: lower, dir: skill, name: "skill.md"}
	mid := moveEnd{rel: staged, dir: stageDir, name: "skill.md"}
	to := moveEnd{rel: canonical, dir: skill, name: "SKILL.md"}

	// The identity of the file about to move, taken from the folder's descriptor right before
	// the move: it is what must be at T before C and at SKILL.md after it.
	moving, err := skill.root.Lstat("skill.md")
	if err != nil {
		return errors.Join(err, removeStage(skill, stage))
	}
	switch err := moveNoReplace(root, from, mid, true); {
	case errors.Is(err, errNoReplaceUnsupported):
		return renameByLinks(root, lower, canonical, stage, stageInfo, orig)
	case err != nil:
		// Nothing moved. The stage folder is ours to take back unless a foreign file made
		// it non-empty or replaced it, in which case rmdir refuses and the error says so.
		return errors.Join(err, removeStage(skill, stage))
	}
	if now, err := stageDir.root.Lstat("skill.md"); err != nil || !os.SameFile(moving, now) {
		return fmt.Errorf("%s no longer holds the file that was moved there (%v); nothing more was moved, and whatever is there was left in place", where(root, stageDir, stage), err)
	}
	if err := moveNoReplace(root, mid, to, true); err != nil {
		// SKILL.md is taken, the move failed, or the path of S no longer leads to the folder
		// that was moved into: put the file back under its own name. On a descriptor-bound
		// platform the put-back does not look at any path (the path may be what changed); a
		// refused put-back leaves the content where it is and the error says where.
		if backErr := moveNoReplace(root, mid, from, !renameIsDescriptorBound); backErr != nil {
			return fmt.Errorf("%w (restoring %s also failed: %v; the content is %s)", err, lower, backErr, contentPlace(root, stageDir, stage, "skill.md"))
		}
		return errors.Join(err, removeStage(skill, stage))
	}
	if now, err := skill.root.Lstat("SKILL.md"); err != nil || !os.SameFile(moving, now) {
		return errors.Join(fmt.Errorf("%s does not hold the file that was staged (%v): the staged name was replaced during the move, and the file now at %s was left in place", canonical, err, canonical), removeStage(skill, stage))
	}
	// The held folder holds the renamed file; the requested path must name it too. A skill
	// folder renamed away after the first rename, with another folder put at its path, passes
	// the look and the held checks and still leaves the requested path without SKILL.md.
	if now, err := root.Lstat(canonical); err != nil || !os.SameFile(moving, now) {
		reason := "the file at that path changed while it was being checked"
		if !leadsTo(root, skill, dirRel) {
			reason = fmt.Sprintf("the skill folder this step opened as %s was moved, so SKILL.md was written inside that moved folder", dirRel)
		}
		return fmt.Errorf("%s does not name the file that was renamed (%v): %s; nothing was removed or moved back, and the staging folder %s inside it was left in place", canonical, err, reason, path.Base(stage))
	}
	return removeStage(skill, stage)
}

// leadsTo reports whether the path rel, looked at through root without following links, still
// names the folder h holds open.
func leadsTo(root *os.Root, h *heldDir, rel string) bool {
	byPath, err := root.Lstat(rel)
	if err != nil {
		return false
	}
	held, err := h.root.Lstat(".")
	return err == nil && os.SameFile(byPath, held)
}

// where names the held folder h, opened as rel, for an error message: rel while the path still
// leads to it, otherwise rel with the fact that the folder was moved away from it.
func where(root *os.Root, h *heldDir, rel string) string {
	if leadsTo(root, h, rel) {
		return rel
	}
	return fmt.Sprintf("the folder this step opened as %s, which was moved", rel)
}

// contentPlace says where the file name inside the held folder h (opened as rel) is: at its
// path while the path still leads to the folder, otherwise inside the folder that was moved.
func contentPlace(root *os.Root, h *heldDir, rel, name string) string {
	if leadsTo(root, h, rel) {
		return "at " + path.Join(rel, name)
	}
	return fmt.Sprintf("in %s inside the folder this step opened as %s, which was moved", name, rel)
}

// moveNoReplace renames from to to without replacing to. With look set, a root-bound look at
// from precedes the rename: where the rename is relative to descriptors it makes the step
// refuse when the path of from no longer leads to the folder that is held, where it is
// path-based (windows) it is the containment against a folder swapped for a link out of the
// project. A put-back passes look false on a descriptor-bound platform: the path it would look
// at is exactly what may have been moved away, and the descriptors still name the folders.
func moveNoReplace(root *os.Root, from, to moveEnd, look bool) error {
	if look {
		if _, err := root.Lstat(from.rel); err != nil {
			return fmt.Errorf("move %s to %s: %w", from.rel, to.rel, err)
		}
	}
	if err := renameNoReplace(from.dir.file, from.name, to.dir.file, to.name); err != nil {
		return fmt.Errorf("move %s to %s: %w", from.rel, to.rel, err)
	}
	return nil
}

// removeStage removes the staging folder with rmdir, relative to the held skill folder, which
// fails on a file, a link and a folder that holds anything, so a foreign entry that took the
// name stays and the error names it.
func removeStage(skill *heldDir, stage string) error {
	if err := removeDirAt(skill.file, path.Base(stage)); err != nil {
		return fmt.Errorf("staging folder %s not removed: %w", stage, err)
	}
	return nil
}

// renameByLinks is the fallback of renameSkillFile for a filesystem or platform without a
// no-replace rename (noReplaceRename reported errNoReplaceUnsupported before anything moved).
// It runs the round-3 design unchanged, with the staging folder renameSkillFile claimed:
// links and identity-checked removals through root. Only this fallback keeps a microsecond
// window, between each identity look and the unlink it guards.
//
// The change never replaces a name. The file is linked into a staging folder the step
// creates exclusively (the link fails when the staged name exists) and its own name is
// dropped, which also keeps the change correct on a case-insensitive filesystem;
// SKILL.md is then linked to it, which fails when SKILL.md exists; and only then are the
// staging names dropped. A refused SKILL.md link puts the file back under its own name by
// another link; if that is refused too, the content stays in the staging folder and the
// error names it. No step moves a file with a rename that could replace a name.
//
// Identity rule. The step captures the identity (Lstat) of skill.md before it links, and of
// the staging folder when it claims it. A name is removed only while Lstat still reports
// that identity for it (dropOurs): a name that is not proven to be our own link or our own
// folder is left in place and reported. skill.md itself is dropped only after both names
// were looked at again and both still match.
//
// States after the first link (L = skill.md, T = the staged name, S = the staging folder;
// "ours" = same file as the original captured before the link):
//
//	state                                 action                                  user data
//	L ours, T ours                        continue (drop L, link SKILL.md,        none lost
//	                                      drop T, drop S)
//	L replaced, T ours                    keep L; drop T (our link) and S; error L is the user's new file; the old
//	                                                                              inode only loses the extra name the
//	                                                                              editor's save already unlinked
//	T replaced, L ours                    keep L, T, S; error names both         foreign T and original L kept
//	L replaced, T replaced                keep L, T, S; error names both         both foreign files kept
//	L gone                                keep T, S; error names T               content kept at T
//	T gone                                keep L, S; error names both            content kept at L
//	L or T cannot be inspected            keep L, T, S; error names both         nothing removed. The two are not
//	                                                                              separable by permissions (T lies under
//	                                                                              S, and L's folder is searched for both),
//	                                                                              so "L unreadable, T readable" cannot
//	                                                                              occur; "T unreadable" (S loses search
//	                                                                              permission) is tested
//	drop of L fails                       keep L; drop T only if ours, then S    none lost
//	undo of T fails (drop of L failed     keep L, T, S; error                    content kept at L and T
//	and the drop of T fails or T is
//	no longer ours)
//	link of SKILL.md refused, restore     drop T only if ours, then S; error     skill.md back, nothing lost
//	link works
//	link of SKILL.md refused, restore     keep T, S; error names T               content kept at T; a file the user
//	link refused                                                                  created at skill.md is never replaced
//	SKILL.md linked                       drop T only if ours, then S            none lost; a foreign T is kept and
//	                                                                              reported
//
// A window of microseconds remains on this path, and only on this path: this platform or
// filesystem has no rename that refuses to replace a name (os.Root has none either), so a
// look and the unlink it guards cannot be made one atomic step. A replacement landing
// between the second look and the unlink of skill.md is still lost, and so is one landing
// between the look inside dropOurs and the removal.
//
// @MX:WARN: [AUTO] filesystem moves and removals inside a user's skill folder
// @MX:REASON: a path swapped between a check and a move is the F24/F25 hazard; every
// step goes through root, no step replaces an existing name, and every drop checks that
// the name is still ours.
func renameByLinks(root *os.Root, lower, canonical, stage string, stageInfo, orig os.FileInfo) error {
	staged := path.Join(stage, "skill.md")
	// Link first: a link fails when staged exists, so a file that appeared inside the
	// stage folder after it was claimed is never replaced. A filesystem without hard
	// links refuses here and nothing changes; the final step below needs them too.
	if err := linkFile(root, lower, staged); err != nil {
		// Nothing linked: skill.md is untouched and the stage folder is ours to remove,
		// unless a foreign file made it non-empty, in which case the removal fails and
		// the error says so.
		return errors.Join(err, dropOurs(root, stage, stageInfo))
	}
	// Each name is compared with the original identity; see the state table above.
	lowerOurs, lowerErr := ours(root, lower, orig)
	stagedOurs, stagedErr := ours(root, staged, orig)
	switch {
	case lowerErr != nil || stagedErr != nil:
		// A name cannot be inspected (gone or unreadable): the other may be the only one
		// left, so nothing is dropped and the error says where the content is.
		return errors.Join(fmt.Errorf("cannot confirm %s and %s; nothing was removed (the content is at %s)", lower, staged, staged), lowerErr, stagedErr)
	case !stagedOurs:
		// The staged name was replaced by someone else (or both were): nothing here is
		// proven ours, so both files and the folder stay.
		return fmt.Errorf("%s and %s no longer both identify the file being renamed; both are left in place with the staging folder, and nothing was renamed", lower, staged)
	case !lowerOurs:
		// skill.md now holds the user's replacement and is left as it is. The staged name
		// is our extra link to the file the user replaced; take it back.
		return errors.Join(
			fmt.Errorf("%s was replaced while it was being renamed; the new file is left in place and nothing was renamed", lower),
			dropOurs(root, staged, orig), dropOurs(root, stage, stageInfo))
	}
	if err := dropFile(root, lower); err != nil {
		// skill.md still has its own name; undo the extra one (it has another name, so
		// this never drops the last name of the content).
		return errors.Join(err, dropOurs(root, staged, orig), dropOurs(root, stage, stageInfo))
	}
	if err := linkFile(root, staged, canonical); err != nil {
		// Put the file back under its own name by a link, which cannot replace a name: a
		// refused restore leaves the content staged, and the error says where.
		if backErr := linkFile(root, staged, lower); backErr != nil {
			return fmt.Errorf("%w (restoring %s also failed: %v; the content is at %s)", err, lower, backErr, staged)
		}
		return errors.Join(err, dropOurs(root, staged, orig), dropOurs(root, stage, stageInfo))
	}
	return errors.Join(dropOurs(root, staged, orig), dropOurs(root, stage, stageInfo))
}

// ours reports whether name, looked at without following links, still identifies the file
// captured in orig.
func ours(root *os.Root, name string, orig os.FileInfo) (bool, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return false, err
	}
	return os.SameFile(orig, info), nil
}

// dropOurs removes name only while it still identifies the file captured in orig; a name
// that cannot be looked at or has changed is left in place and the error says so.
func dropOurs(root *os.Root, name string, orig os.FileInfo) error {
	ok, err := ours(root, name, orig)
	if err != nil {
		return fmt.Errorf("%s not removed: %w", name, err)
	}
	if !ok {
		return fmt.Errorf("%s is no longer the file this step created; left in place", name)
	}
	return dropFile(root, name)
}

// claimStage creates the staging folder for a rename beside the skill file and returns
// its identity (which only the link fallback uses). Mkdir fails when the name exists, so a
// file or folder that is there, or that appears first, is never replaced: the next name is
// tried. The name carries a random part (stageSuffix), so a foreign actor that wants to
// swap the folder has to guess it.
func claimStage(root *os.Root, dirRel string) (string, os.FileInfo, error) {
	const base = "skill.md.tmp-rename"
	for n := range 100 {
		name := base
		if suffix := stageSuffix(); suffix != "" {
			name = base + "-" + suffix
		} else if n > 0 {
			name = fmt.Sprintf("%s-%d", base, n)
		}
		stage := path.Join(dirRel, name)
		err := mkdirFile(root, stage)
		if err == nil {
			info, statErr := root.Lstat(stage)
			if statErr != nil {
				return "", nil, fmt.Errorf("%w (the empty folder %s is left in place)", statErr, stage)
			}
			return stage, info, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", nil, err
		}
	}
	return "", nil, fmt.Errorf("no free temporary name beside %s", path.Join(dirRel, "skill.md"))
}

// symlinkedRoot reports the first of .moai, .claude, or .claude/skills under
// projectRoot that is a symbolic link. Cleanup reads, moves, and removes entries
// under those paths, so a link there would reach a directory outside the project.
func symlinkedRoot(projectRoot string) (string, bool) {
	for _, rel := range []string{".moai", ".claude", ".claude/skills"} {
		info, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(rel)))
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return rel, true
		}
	}
	return "", false
}
