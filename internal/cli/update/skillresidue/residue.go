// Package skillresidue repairs the skill-folder leftovers that earlier MoAI
// versions and interrupted deploys leave under .claude/skills: an empty folder of
// a skill the template ships, a retired skill name, and a lowercase skill.md. A
// change is a removal of an empty folder or a rename. A retired skill folder and
// the link to it are reported and left in place. Each action is returned so the
// update can report it.
package skillresidue

import (
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

// The held-back rename is a few root-bound steps, each behind a named function and a
// package variable. The functions are the only places the package calls the handle's
// Rename or Remove, so the registry's destructive-target scan sees named sites, and the
// pin test counts every reference to them so an alias is visible. A test replaces a
// variable to make one step fail, or to run a second actor between two steps.

// mkdirOne creates a folder inside root and fails when the name exists. That failure is
// what makes a temporary name a claim and not a guess.
func mkdirOne(root *os.Root, name string) error { return root.Mkdir(name, 0o755) }

// moveOne is the package's one root-bound Rename. It replaces an existing destination, so
// the only caller is putBack, which uses it when the restore link has failed and the name
// was seen free just before. The registry's destructive-target scan counts it.
func moveOne(root *os.Root, from, to string) error { return root.Rename(from, to) }

// linkOne makes to a second name for the file at from. It fails when to exists, so it
// can place a name where Rename could only replace one.
func linkOne(root *os.Root, from, to string) error { return root.Link(from, to) }

// dropOne removes one empty folder or one name of a file that has another name. It
// never removes the last name of content (see renameSkillFile).
func dropOne(root *os.Root, name string) error { return root.Remove(name) }

var (
	mkdirFile = mkdirOne
	moveFile  = moveOne
	linkFile  = linkOne
	dropFile  = dropOne
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

// renameSkillFile renames <dirRel>/skill.md to <dirRel>/SKILL.md through root, a handle
// bound to the project root, where dirRel is the skill folder's slash-separated path
// relative to that root. A component swapped for a symbolic link after the caller's
// checks cannot lead the move outside the project: the handle refuses a path that
// leaves it.
//
// The change never replaces a name. The file is linked into a staging folder the step
// creates exclusively (the link fails when the staged name exists) and its own name is
// dropped, which also keeps the change correct on a case-insensitive filesystem;
// SKILL.md is then linked to it, which fails when SKILL.md exists; and only then are the
// staging names dropped. Every drop removes a name of a file that has
// another name at that moment, or an empty folder. A refused link puts the file back
// under its own name; if that is impossible, the content stays in the staging folder
// and the error names it.
//
// Before skill.md is dropped, both names are looked at again and must still identify the
// same file; if skill.md was replaced (an editor's atomic save), it is left in place and
// the step reports an error. A window of microseconds remains between that look and the
// unlink: no rename that refuses to replace a name exists (os.Root has none), so the two
// cannot be made one atomic step, and a replacement landing exactly there is still lost.
//
// @MX:WARN: [AUTO] filesystem moves and removals inside a user's skill folder
// @MX:REASON: a path swapped between a check and a move is the F24/F25 hazard; every
// step goes through root, and no step replaces an existing name.
func renameSkillFile(root *os.Root, dirRel string) error {
	lower := path.Join(dirRel, "skill.md")
	canonical := path.Join(dirRel, "SKILL.md")
	info, err := root.Lstat(lower)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file; nothing renamed", lower)
	}

	stage, err := claimStage(root, dirRel)
	if err != nil {
		return err
	}
	staged := path.Join(stage, "skill.md")
	// Link first: a link fails when staged exists, so a file that appeared inside the
	// stage folder after it was claimed is never replaced. A filesystem without hard
	// links refuses here and nothing changes; the final step below needs them too.
	if err := linkFile(root, lower, staged); err != nil {
		// Nothing linked: skill.md is untouched and the stage folder is ours to remove,
		// unless a foreign file made it non-empty, in which case the removal fails and
		// the error says so.
		return errors.Join(err, dropFile(root, stage))
	}
	// The name skill.md is dropped only while it still identifies the file that was linked.
	// A user's editor that atomic-saves (renames a temporary file over skill.md) in the gap
	// gives the name a new inode, and dropping it would delete the user's new file.
	same, err := sameFileAt(root, lower, staged)
	if err != nil {
		// skill.md cannot be inspected (gone or unreadable): the staged name may be the
		// only one left, so nothing is dropped and the error says where the content is.
		return fmt.Errorf("%w (the content is at %s)", err, staged)
	}
	if !same {
		// skill.md now holds the user's replacement and is left as it is. The staged name
		// is our extra link to the file the user replaced; take it back.
		return errors.Join(
			fmt.Errorf("%s was replaced while it was being renamed; the new file is left in place and nothing was renamed", lower),
			dropFile(root, staged), dropFile(root, stage))
	}
	if err := dropFile(root, lower); err != nil {
		// skill.md still has its own name; undo the extra one (it has another name, so
		// this never drops the last name of the content).
		return errors.Join(err, dropFile(root, staged), dropFile(root, stage))
	}
	if err := linkFile(root, staged, canonical); err != nil {
		if backErr := putBack(root, staged, lower); backErr != nil {
			return fmt.Errorf("%w (restoring %s also failed: %v; the content is at %s)", err, lower, backErr, staged)
		}
		return errors.Join(err, dropFile(root, stage))
	}
	return errors.Join(dropFile(root, staged), dropFile(root, stage))
}

// sameFileAt reports whether the names a and b, looked at without following links,
// identify the same file.
func sameFileAt(root *os.Root, a, b string) (bool, error) {
	ai, err := root.Lstat(a)
	if err != nil {
		return false, err
	}
	bi, err := root.Lstat(b)
	if err != nil {
		return false, err
	}
	return os.SameFile(ai, bi), nil
}

// claimStage creates the staging folder for a rename beside the skill file. Mkdir
// fails when the name exists, so a file or folder that is there, or that appears first,
// is never replaced: the next name is tried.
func claimStage(root *os.Root, dirRel string) (string, error) {
	const base = "skill.md.tmp-rename"
	for n := range 100 {
		name := base
		if n > 0 {
			name = fmt.Sprintf("%s-%d", base, n)
		}
		stage := path.Join(dirRel, name)
		err := mkdirFile(root, stage)
		if err == nil {
			return stage, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("no free temporary name beside %s", path.Join(dirRel, "skill.md"))
}

// putBack returns the staged file to its own name. A hard link cannot replace a name, so
// it is tried first. Where the filesystem has no hard links the file is moved back, and
// only while the name is still free. A taken name leaves the content staged.
func putBack(root *os.Root, staged, lower string) error {
	err := linkFile(root, staged, lower)
	if err == nil {
		return dropFile(root, staged)
	}
	if _, statErr := root.Lstat(lower); statErr == nil {
		return err
	}
	return moveFile(root, staged, lower)
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
