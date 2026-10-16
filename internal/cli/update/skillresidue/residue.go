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
// package variable. The package never calls the handle's Rename; the only place it calls
// the handle's Remove is dropOne, so the registry's destructive-target scan sees one named
// site, and the pin test counts every reference to the seams so an alias is visible. A test
// replaces a variable to make one step fail, or to run a second actor between two steps.

// mkdirOne creates a folder inside root and fails when the name exists. That failure is
// what makes a temporary name a claim and not a guess.
func mkdirOne(root *os.Root, name string) error { return root.Mkdir(name, 0o755) }

// linkOne makes to a second name for the file at from. It fails when to exists, so it
// can place a name where Rename could only replace one.
func linkOne(root *os.Root, from, to string) error { return root.Link(from, to) }

// dropOne removes one empty folder or one name of a file that has another name. It
// never removes the last name of content (see renameSkillFile).
func dropOne(root *os.Root, name string) error { return root.Remove(name) }

var (
	mkdirFile = mkdirOne
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
// A window of microseconds remains between the second look and the unlink of skill.md:
// no rename that refuses to replace a name exists (os.Root has none), so the two cannot
// be made one atomic step, and a replacement landing exactly there is still lost. The same
// holds for the look before each other drop (dropOurs): the identity is read, then the
// name is removed, and a replacement landing between the two is removed with it.
//
// @MX:WARN: [AUTO] filesystem moves and removals inside a user's skill folder
// @MX:REASON: a path swapped between a check and a move is the F24/F25 hazard; every
// step goes through root, no step replaces an existing name, and every drop checks that
// the name is still ours.
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

	stage, stageInfo, err := claimStage(root, dirRel)
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
// its identity. Mkdir fails when the name exists, so a file or folder that is there, or
// that appears first, is never replaced: the next name is tried.
func claimStage(root *os.Root, dirRel string) (string, os.FileInfo, error) {
	const base = "skill.md.tmp-rename"
	for n := range 100 {
		name := base
		if n > 0 {
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
