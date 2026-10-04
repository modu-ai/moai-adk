package update

// migrate_classify.go — SPEC-INIT-SHRINK-001 REQ-010/REQ-013 (plan M1): the
// migration classifier. Every deployed file under a dropped component root
// (.claude/skills/**, .claude/commands/**, the Codex mirror) is classified
// into exactly one of three classes BEFORE anything is removed:
//
//	identical — the template render for this project's context carries the
//	            path, the on-disk content equals that render, and the
//	            manifest record is healthy (managed, not stale, not
//	            user-modified). Removed without archive (OD-3 settled (a)):
//	            the plugin and the template render are the recovery source.
//	modified  — template-carried, and the content differs from the render, or
//	            the manifest record is absent or stale for it (RK-9, OD-3
//	            settled condition: absent/stale routes here conservatively —
//	            archive-then-remove).
//	foreign   — NOT carried by the template render for this project's
//	            context. Every user-created skill or command is foreign on
//	            this ground alone, a moai-custom name included, whatever the
//	            managed-name rules (plan.go's IsMoaiManaged / namespace
//	            rules) match and whatever the manifest says. Foreign files
//	            are preserved byte-for-byte and never archived (REQ-013).
//
// The class gate is TEMPLATE CARRIAGE. Managed-name matching alone never
// routes a file into a removal class — that is the D-15 loss path this
// classifier exists to close (leader condition 5a).
//
// Symlinks are never classified: they are reported in the plan's Symlinks
// list and the removal/archive machinery refuses them (REQ-013, the
// REQ-SEC-003 rule of the archive contract).

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// MigrateClass is the three-way classification of a dropped-root file.
type MigrateClass string

const (
	// ClassIdentical — removable without archive (OD-3 settled (a)).
	ClassIdentical MigrateClass = "identical"
	// ClassModified — archive-then-remove (REQ-012); absent/stale manifest
	// records land here conservatively.
	ClassModified MigrateClass = "modified"
	// ClassForeign — preserved byte-for-byte, never archived (REQ-013).
	ClassForeign MigrateClass = "foreign"
)

// MigrationArchiveTag is the archive directory tag the migration's
// modified-class backup uses (REQ-012; design §3 step 3): distinct from the
// legacy-skill "v2.16" tag so the two preservation flows never share a root.
// The archive unit is the classified FILE — see ArchiveSkillFileRoot and
// ArchiveFilesRoot for the two layouts.
const MigrationArchiveTag = "init-shrink-migration"

// ArchiveSkillFileRoot is the archive root for a classified file under
// .claude/skills: .moai/archive/skills/<MigrationArchiveTag>/<skill>/<rest>.
// A per-file write preserves the skill-directory layout while the
// directory's identical and foreign members are NOT copied (design §3
// step 3; the archive unit is the classified file, never a whole directory).
func ArchiveSkillFileRoot() string {
	return filepath.Join(".moai", "archive", "skills", MigrationArchiveTag)
}

// ArchiveFilesRoot is the standalone-file variant (commands, mirror files):
// .moai/archive/files/<MigrationArchiveTag>/<original relative path>.
func ArchiveFilesRoot() string {
	return filepath.Join(".moai", "archive", "files", MigrationArchiveTag)
}

// droppedComponentRoots are the classified roots, project-root-relative
// slash paths (REQ-010): the two Claude roots the thin deploy stops
// carrying, and the Codex mirror of REQ-006.
var droppedComponentRoots = []string{
	".claude/skills",
	".claude/commands",
	".agents/skills",
}

// MigrationRoots returns the dropped component roots this migration
// classifies. A copy: callers cannot mutate the classifier's scope.
func MigrationRoots() []string {
	out := make([]string, len(droppedComponentRoots))
	copy(out, droppedComponentRoots)
	return out
}

// TemplateRender is the deploy-mode template render the classifier compares
// against (REQ-010). Production wires it to the embedded template tree and
// the project's render context; tests substitute an in-memory FS.
type TemplateRender interface {
	// Carries reports whether the render has a source at the
	// deploy-relative slash path; content holds the rendered bytes when
	// carried.
	Carries(relPath string) (content []byte, carried bool)
}

// MigrationFile is one classified file.
type MigrationFile struct {
	// RelPath is the file's slash path relative to the project root.
	RelPath string
	// Class is the file's class (redundant with the containing set, carried
	// so a single file travels with its verdict).
	Class MigrateClass
}

// MigrationPlan is the classifier's output: the three exclusive classes,
// the symlinks found under the roots (never classified), and the counts
// REQ-010 prints.
type MigrationPlan struct {
	Identical []MigrationFile
	Modified  []MigrationFile
	Foreign   []MigrationFile
	// Symlinks holds the project-root-relative slash paths of every symlink
	// entry found under the dropped roots. They are never classified,
	// archived, or followed (REQ-013); mirror-entry handling is REQ-006's
	// own clause.
	Symlinks []string
}

// CountIdentical returns the identical-set size.
func (p MigrationPlan) CountIdentical() int { return len(p.Identical) }

// CountModified returns the modified-set size.
func (p MigrationPlan) CountModified() int { return len(p.Modified) }

// CountForeign returns the foreign-set size.
func (p MigrationPlan) CountForeign() int { return len(p.Foreign) }

// ClassOf returns the class of one path, or "" when the path was not seen
// under the dropped roots at all.
func (p MigrationPlan) ClassOf(relPath string) MigrateClass {
	for _, f := range p.Identical {
		if f.RelPath == relPath {
			return ClassIdentical
		}
	}
	for _, f := range p.Modified {
		if f.RelPath == relPath {
			return ClassModified
		}
	}
	for _, f := range p.Foreign {
		if f.RelPath == relPath {
			return ClassForeign
		}
	}
	return ""
}

// errClassifyStopped is the one error shape ClassifyMigration returns: a
// walk failure on a dropped root (the classifier is read-only; any error
// means no classification happened).
var errClassifyStopped = errors.New("classify: stopped")

// ClassifyMigration walks the dropped component roots under projectRoot and
// classifies every regular file into exactly one of the three classes
// (REQ-010). mf may be nil — a project with no manifest classifies every
// template-carried file as modified (the conservative RK-9 route) and
// everything else as foreign.
//
// The walk is read-only. A symlink entry anywhere under a root is recorded
// in Symlinks and skipped: classification never dereferences a link
// (REQ-013).
func ClassifyMigration(projectRoot string, render TemplateRender, mf *manifest.Manifest) (MigrationPlan, error) {
	plan := MigrationPlan{}
	if render == nil {
		return plan, errors.New("classify: nil template render")
	}

	type found struct {
		rel   string
		bytes []byte
	}
	var files []found

	for _, root := range droppedComponentRoots {
		absRoot := filepath.Join(projectRoot, filepath.FromSlash(root))
		info, err := os.Lstat(absRoot)
		if err != nil {
			continue // root absent: nothing to classify under it
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			// The root ITSELF is a link: record and move on.
			plan.Symlinks = append(plan.Symlinks, root)
			continue
		}
		if !info.IsDir() {
			continue
		}
		walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, relErr := filepath.Rel(projectRoot, path)
			if relErr != nil {
				return relErr
			}
			rel = filepath.ToSlash(rel)
			if d.IsDir() {
				return nil
			}
			// REQ-013: Lstat semantics — the link itself is the entry. Never
			// dereference, never classify, never archive it.
			if d.Type()&fs.ModeSymlink != 0 {
				plan.Symlinks = append(plan.Symlinks, rel)
				return nil
			}
			if !d.Type().IsRegular() {
				return nil // sockets, fifos: outside the classified universe
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			files = append(files, found{rel: rel, bytes: data})
			return nil
		})
		if walkErr != nil {
			return MigrationPlan{Symlinks: plan.Symlinks}, errors.Join(errClassifyStopped, walkErr)
		}
	}

	// Deterministic output order: walk order is already sorted per root, but
	// the root list order governs; sort the whole slice once.
	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })
	sort.Strings(plan.Symlinks)

	for _, f := range files {
		rendered, carried := render.Carries(f.rel)
		if !carried {
			// The template render does not carry the path: foreign whatever
			// its manifest state or managed-name match (REQ-010/REQ-013).
			plan.Foreign = append(plan.Foreign, MigrationFile{RelPath: f.rel, Class: ClassForeign})
			continue
		}
		class := classifyCarried(f.bytes, rendered, manifestEntry(mf, f.rel))
		file := MigrationFile{RelPath: f.rel, Class: class}
		if class == ClassIdentical {
			plan.Identical = append(plan.Identical, file)
		} else {
			plan.Modified = append(plan.Modified, file)
		}
	}
	return plan, nil
}

// manifestEntry returns the file's manifest record, or nil.
func manifestEntry(mf *manifest.Manifest, relPath string) *manifest.FileEntry {
	if mf == nil {
		return nil
	}
	if entry, ok := mf.Files[relPath]; ok {
		return &entry
	}
	return nil
}

// classifyCarried decides between identical and modified for a
// template-carried file (REQ-010):
//
//	identical requires ALL of: a healthy record (provenance
//	template_managed, current hash matching the on-disk bytes), and content
//	equal to the render.
//	everything else is modified — content differing from the render, an
//	absent record, a stale record, or a user-owned provenance.
func classifyCarried(disk, rendered []byte, entry *manifest.FileEntry) MigrateClass {
	if entry == nil {
		return ClassModified // absent record → archive-then-remove (OD-3 settled condition)
	}
	if entry.Provenance != manifest.TemplateManaged {
		return ClassModified // user_modified / user_created / deprecated: never silently removable
	}
	if entry.CurrentHash != "" && entry.CurrentHash != manifest.HashBytes(disk) {
		return ClassModified // stale record: the file moved since the last track
	}
	if !bytes.Equal(disk, rendered) {
		return ClassModified
	}
	return ClassIdentical
}
