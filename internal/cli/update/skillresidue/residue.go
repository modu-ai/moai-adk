// Package skillresidue repairs the skill-folder leftovers that earlier MoAI
// versions and interrupted deploys leave under .claude/skills: an empty folder of
// a skill the template ships, a retired skill name, and a lowercase skill.md. A
// change is a removal of an empty folder, a move into the archive, or a rename,
// and each one is returned as an Action so the update can report it.
package skillresidue

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// retiredSkills maps a skill folder name MoAI no longer ships to the name that
// replaced it. The retired folder is archived only when the template ships the
// successor.
var retiredSkills = map[string]string{
	"moai-kanban-foreman": "moai-factory-foreman",
}

// Options are the inputs the cleanup takes from its caller.
type Options struct {
	// Shipped reports whether the embedded template ships a project-relative path
	// such as ".claude/skills/moai-factory-foreman/SKILL.md". A nil Shipped ships
	// nothing, so no folder is removed, renamed, or archived on that basis.
	Shipped func(rel string) bool
	// Stamp names the archive folder, .moai/archive/skills/retired-<Stamp>/.
	Stamp string
}

func (o Options) shipped(rel string) bool {
	return o.Shipped != nil && o.Shipped(rel)
}

// Action is one residue the cleanup acted on or reported.
type Action struct {
	// Path is the project-relative, slash-separated path the residue was found at.
	Path string
	// Kind is one of removed-empty-dir, renamed-skill-file,
	// reported-both-skill-files, archived-retired-skill,
	// kept-retired-skill-with-user-files, or removed-stale-link.
	Kind string
}

// Clean repairs the skill-folder residue under projectRoot/.claude/skills and
// returns what it did. A second pass over a repaired tree returns nothing.
func Clean(projectRoot string, opts Options) ([]Action, error) {
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
			if err := archiveFolder(projectRoot, opts.Stamp, name, dir); err != nil {
				return actions, err
			}
			actions = append(actions, Action{Path: rel, Kind: "archived-retired-skill"})
			removed, err := removeStaleLink(projectRoot, name)
			if err != nil {
				return actions, err
			}
			if removed {
				actions = append(actions, Action{Path: ".agents/skills/" + name, Kind: "removed-stale-link"})
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

// archiveFolder moves a retired skill folder under the archive, keeping its name.
// It moves the folder rather than deleting it, so nothing is lost.
func archiveFolder(projectRoot, stamp, name, dir string) error {
	dest := filepath.Join(projectRoot, ".moai", "archive", "skills", "retired-"+stamp, name)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.Rename(dir, dest)
}

// removeStaleLink removes the .agents/skills link that pointed at an archived
// folder, so no link dangles. A link that points anywhere else, or a real folder
// in that place, is left alone.
func removeStaleLink(projectRoot, name string) (bool, error) {
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
	if !strings.HasSuffix(filepath.ToSlash(filepath.Clean(target)), ".claude/skills/"+name) {
		return false, nil
	}
	if err := os.Remove(link); err != nil {
		return false, err
	}
	return true, nil
}

// renameSkillFile renames skill.md to SKILL.md through a temporary name, so the
// change also holds on a case-insensitive filesystem.
func renameSkillFile(dir string) error {
	lower := filepath.Join(dir, "skill.md")
	tmp := filepath.Join(dir, "skill.md.tmp-rename")
	if err := os.Rename(lower, tmp); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, "SKILL.md"))
}
