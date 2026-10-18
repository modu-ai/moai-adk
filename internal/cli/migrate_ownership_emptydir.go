package cli

// migrate_ownership_emptydir.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M5 (card
// t1865), REQ-OWR-008 and design.md D-6: the separate step that runs after the
// migration walk and removes a skill directory that is empty when the manifest
// records show only template-managed files were ever in it.
//
// The step removes DIRECTORIES, never files: it deletes with os.Remove, which
// can take only an empty directory, so a directory that still holds content
// survives even if the emptiness check were bypassed. It does not depend on the
// removal switch — that switch gates deleting shipped copies, and this step
// deletes none. The one-level prune a removal triggers (pruneEmptiedSkillDir,
// card t1668) stays as it is and is reused here for the skill directory itself.
//
// Candidates are the directories below the two skill roots. The skill roots
// themselves, the agent directories and a symlinked root are never candidates.
// A directory is judged from two record pools — the manifest the run started
// from and the retained prior generations — and is removed only when every file
// ever recorded under it was template-managed and none is on disk. Anything
// less keeps the directory and lists it with `empty-dir-no-provenance`.

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// emptyDirEvidence is the record pool a directory's provenance is judged from:
// the manifest the run started from and the retained prior generations
// (design.md D-6, sources S1 and S2).
type emptyDirEvidence struct {
	// root is the project root, to resolve a recorded path on disk.
	root  string
	start map[string]manifest.Provenance
	gens  []priorGeneration
}

// emptyDirVerdict is the provenance judgment of one directory.
type emptyDirVerdict struct {
	// managedOnly is true when the directory has records, every one of them
	// template-managed, and none of the recorded files is on disk.
	managedOnly bool
	// class describes the directory when managedOnly is false: the strongest
	// non-managed provenance found (user-created, then user-modified), else
	// unregistered.
	class ownershipClass
}

// emptyDirProvenanceFn is the judgment the pruning step applies. It is a
// variable only so a test can inject a deliberately wrong judgment and watch the
// assertions go red; no production path replaces it.
var emptyDirProvenanceFn = dirManagedProvenance

// dirManagedProvenance collects every record under dirRel from both pools and
// answers whether they prove the directory held only template-managed files that
// are now gone. The prefix is dirRel plus a slash, so `moai-a` never claims the
// records of `moai-ab`. A non-managed record in either pool settles the answer:
// it is the record of something that was not the template's to delete.
func dirManagedProvenance(ev emptyDirEvidence, dirRel string) emptyDirVerdict {
	prefix := dirRel + "/"
	paths := map[string]bool{}
	var userCreated, userModified, otherNonManaged bool
	note := func(rel string, p manifest.Provenance) {
		if !strings.HasPrefix(rel, prefix) {
			return
		}
		paths[rel] = true
		switch p {
		case manifest.TemplateManaged:
		case manifest.UserCreated:
			userCreated = true
		case manifest.UserModified:
			userModified = true
		default:
			// deprecated, generated, empty or unknown: not a template-managed
			// record, so not evidence.
			otherNonManaged = true
		}
	}
	for rel, p := range ev.start {
		note(rel, p)
	}
	for _, g := range ev.gens {
		for rel, r := range g.files {
			note(rel, r.provenance)
		}
	}
	switch {
	case len(paths) == 0:
		return emptyDirVerdict{class: ownershipUnregistered}
	case userCreated:
		return emptyDirVerdict{class: ownershipUserCreated}
	case userModified:
		return emptyDirVerdict{class: ownershipUserModified}
	case otherNonManaged:
		return emptyDirVerdict{class: ownershipUnregistered}
	}
	for rel := range paths {
		if _, err := os.Lstat(filepath.Join(ev.root, filepath.FromSlash(rel))); err == nil {
			// A recorded path is still there: the directory was not emptied of
			// what the records describe.
			return emptyDirVerdict{class: ownershipUnregistered}
		}
	}
	return emptyDirVerdict{managedOnly: true}
}

// startOfRunRecords is the provenance of every file the run's manifest holds,
// read before anything in the run drops an entry.
func startOfRunRecords(mgr manifest.Manager) map[string]manifest.Provenance {
	mf := mgr.Manifest()
	if mf == nil {
		return nil
	}
	out := make(map[string]manifest.Provenance, len(mf.Files))
	for rel, f := range mf.Files {
		out[rel] = f.Provenance
	}
	return out
}

// emptyDirTree lists the directories of the tree at abs, deepest first, and
// reports whether the tree holds nothing but directories. A file, a symlink or
// an unreadable entry anywhere in it makes the whole tree not a candidate.
func emptyDirTree(abs string) (dirs []string, onlyDirs bool) {
	onlyDirs = true
	_ = filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			onlyDirs = false
			return filepath.SkipAll
		}
		dirs = append(dirs, p)
		return nil
	})
	// Pre-order lists a parent before its descendants; reversed, every
	// descendant comes before its parent.
	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}
	return dirs, onlyDirs
}

// pruneNestedEmptyDir removes dir when it is empty and stays inside the resolved
// project root. It is the deeper-level sibling of pruneEmptiedSkillDir, which
// deliberately stops at exactly one level below a skill root; the boundary and
// emptiness guards are the same, and os.Remove can take only an empty directory.
func pruneNestedEmptyDir(rootResolved, dir string) (bool, error) {
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil || !withinRootBoundary(rootResolved, resolved) {
		return false, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		return false, nil
	}
	if err := os.Remove(dir); err != nil {
		return false, err
	}
	return true, nil
}

// pruneManagedEmptyDirs is the pruning step (REQ-OWR-008). It looks at each
// skill directory below the two skill roots, and inside it at every directory,
// deepest first: an empty directory the records prove managed-only is removed; an
// empty one they do not is kept and listed (rc may be nil for a direct call, in
// which case only the report line is written). It never removes a file.
func pruneManagedEmptyDirs(projectRoot, rootResolved string, mgr manifest.Manager, src provenanceSource, rc *migrationRunContext, report func(string, ...interface{})) {
	ev := emptyDirEvidence{root: projectRoot, start: startOfRunRecords(mgr), gens: src.generations()}
	for _, rootRel := range projectCommonAssetRels[:2] { // the two skill roots
		absRoot := filepath.Join(projectRoot, filepath.FromSlash(rootRel))
		if info, err := os.Lstat(absRoot); err != nil || !info.IsDir() {
			// Absent, or a symlinked mirror root: the mirror's directories are
			// the target's and are judged there.
			continue
		}
		entries, err := os.ReadDir(absRoot)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			pruneSkillDirTree(projectRoot, rootResolved, filepath.Join(absRoot, e.Name()), ev, rc, report)
		}
	}
}

// pruneSkillDirTree judges the directories of one skill directory's tree,
// deepest first, when the tree holds no file at all.
func pruneSkillDirTree(projectRoot, rootResolved, skillDir string, ev emptyDirEvidence, rc *migrationRunContext, report func(string, ...interface{})) {
	dirs, onlyDirs := emptyDirTree(skillDir)
	if !onlyDirs {
		return
	}
	for _, d := range dirs {
		// A child that was kept makes its parent non-empty: nothing to judge.
		if entries, err := os.ReadDir(d); err != nil || len(entries) != 0 {
			continue
		}
		relOS, err := filepath.Rel(projectRoot, d)
		if err != nil {
			continue
		}
		rel := filepath.ToSlash(relOS)
		verdict := emptyDirProvenanceFn(ev, rel)
		if !verdict.managedOnly {
			class := verdict.class
			if !class.valid() {
				class = ownershipUnregistered
			}
			if rc != nil {
				rc.remaining = append(rc.remaining, remainingRow{
					Rel: rel, Class: class, Reason: keepEmptyDirNoProvenance,
					Counterpart: counterpartStatusNotApplicable, ProfileSide: profileSideNotApplicable,
				})
			}
			report("  migration: empty directory kept (no provenance record covers it) [reason: %s]: %s", keepEmptyDirNoProvenance, rel)
			continue
		}
		var pruned bool
		var pErr error
		if strings.Count(rel, "/") == 2 {
			// Exactly one level below a skill root: the t1668 prune, handed a
			// path whose directory is this one.
			pruned, pErr = pruneEmptiedSkillDir(projectRoot, rootResolved, filepath.Join(d, path.Base(rel)))
		} else {
			pruned, pErr = pruneNestedEmptyDir(rootResolved, d)
		}
		switch {
		case pErr != nil:
			report("  migration: empty directory prune failed: %s: %v", rel, pErr)
		case pruned:
			report("  migration: removed empty managed directory: %s", rel)
		}
	}
}
