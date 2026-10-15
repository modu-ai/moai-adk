// Package skillresidue repairs the skill-folder leftovers that earlier MoAI
// versions and interrupted deploys leave under .claude/skills: an empty folder of
// a skill the template ships, a retired skill name, and a lowercase skill.md. A
// change is a removal of an empty folder or a rename. A retired skill folder and
// the link to it are reported and left in place. Each action is returned so the
// update can report it.
package skillresidue

import (
	"fmt"
	"os"
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
			if opts.shipped(rel + "/SKILL.md") {
				if err := renameSkillFile(dir); err != nil {
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

// moveOne is the only os.Rename call in this package. Every move goes through it,
// so the registry's destructive-target scan sees one site, not an aliased name.
func moveOne(from, to string) error { return os.Rename(from, to) }

// moveFile is moveOne. A test replaces it to make one move fail on any filesystem.
var moveFile = moveOne

// renameSkillFile renames skill.md to SKILL.md through a temporary name, so the
// change also holds on a case-insensitive filesystem.
func renameSkillFile(dir string) error {
	lower := filepath.Join(dir, "skill.md")
	tmp, err := freeName(dir, "skill.md.tmp-rename")
	if err != nil {
		return err
	}
	if err := moveFile(lower, tmp); err != nil {
		return err
	}
	if err := moveFile(tmp, filepath.Join(dir, "SKILL.md")); err != nil {
		// The content is still under tmp. Put it back under its own name, so no
		// content is left where nothing reads it.
		if restoreErr := moveFile(tmp, lower); restoreErr != nil {
			return fmt.Errorf("%w (restoring %s also failed: %v)", err, lower, restoreErr)
		}
		return err
	}
	return nil
}

// freeName returns dir/base, or dir/base-N for the first N that names nothing in
// dir. A rename onto an existing name replaces it, so a temporary name must never
// be one a user file already holds.
func freeName(dir, base string) (string, error) {
	for n := range 100 {
		name := base
		if n > 0 {
			name = fmt.Sprintf("%s-%d", base, n)
		}
		p := filepath.Join(dir, name)
		_, err := os.Lstat(p)
		if os.IsNotExist(err) {
			return p, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("no free temporary name beside %s", filepath.Join(dir, base))
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
