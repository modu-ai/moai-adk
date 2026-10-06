package update

// reconcile_classify.go — SPEC-UPDATE-MIGRATION-001 M3 (card t1547): the
// managed-root ownership classifier (REQ-UPM-001..004). It extends the
// SPEC-INIT-SHRINK-001 three-way classifier (migrate_classify.go — same
// package; classifyCarried/manifestEntry are reused verbatim) from the
// dropped component roots to the MANAGED-root set, producing the four
// carriage-gated classes the reconciliation dispositions route on:
//
//	template-owned — template carries the path AND the manifest record is
//	                 healthy AND content equals the render → refresh in place.
//	user-modified  — template carries the path but content diverges, or the
//	                 record is absent/stale → merge-or-conflict.
//	user-owned     — the template does not carry the path (and no prior
//	                 template carried it) → preserve untouched.
//	stale          — a prior template carried the path (manifest record)
//	                 and the current one does not → archive-then-remove.
//
// The class gate is TEMPLATE CARRIAGE (REQ-UPM-002): managed-name matching
// or the clean-target list alone never routes a file into a removal-eligible
// class — the D-15 loss path. IsUserOwnedNamespace paths are forced
// user-owned (REQ-UPM-003) as defense-in-depth.
//
// Symlinks are never classified: Lstat semantics record every link in the
// plan's Symlinks list without dereferencing it (REQ-UPM-004, the
// SPEC-CLI-CLEAN-SYMLINK-001 rule carried forward).
//
// The walk is read-only and deterministic: entries are sorted by slash path
// before classification, so two runs over the same tree and template set
// produce byte-identical output (NFR-UPM-001).

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ReconcileClass is the four-way ownership class of a managed-root file.
type ReconcileClass string

const (
	// ClassTemplateOwned — refresh in place from the render (REQ-UPM-010).
	ClassTemplateOwned ReconcileClass = "template-owned"
	// ClassUserModified — 3-way merge, or conflict → preserve + sidecar.
	ClassUserModified ReconcileClass = "user-modified"
	// ClassUserOwned — preserve untouched, never removed (REQ-UPM-013).
	ClassUserOwned ReconcileClass = "user-owned"
	// ClassStale — archive-then-remove (REQ-UPM-014).
	ClassStale ReconcileClass = "stale"
)

// ReconcileFile is one classified file.
type ReconcileFile struct {
	// RelPath is the file's project-root-relative slash path.
	RelPath string
	// Class is the file's class (redundant with the containing set, carried
	// so a single file travels with its verdict).
	Class ReconcileClass
}

// ReconcilePlan is the classifier's output: the four exclusive classes, the
// symlinks found under the targets (never classified), and the counts the
// outcome summary prints.
type ReconcilePlan struct {
	TemplateOwned []ReconcileFile
	UserModified  []ReconcileFile
	UserOwned     []ReconcileFile
	Stale         []ReconcileFile
	// Symlinks holds the project-root-relative slash paths of every symlink
	// entry found under the targets, live or dangling. They are never
	// classified, merged, archived, or followed (REQ-UPM-004).
	Symlinks []string
}

// Count returns the total classified regular-file count.
func (p ReconcilePlan) Count() int {
	return len(p.TemplateOwned) + len(p.UserModified) + len(p.UserOwned) + len(p.Stale)
}

// ClassOf returns the class of one path, or "" when the path was not seen
// under the targets at all.
func (p ReconcilePlan) ClassOf(relPath string) ReconcileClass {
	for _, set := range [][]ReconcileFile{p.TemplateOwned, p.UserModified, p.UserOwned, p.Stale} {
		for _, f := range set {
			if f.RelPath == relPath {
				return f.Class
			}
		}
	}
	return ""
}

// errClassifyStopped is the one error shape ClassifyManagedRoots returns: a
// walk failure on a target (the classifier is read-only; any error means no
// classification happened and nothing was changed).
var errReconcileClassifyStopped = errors.New("reconcile classify: stopped")

// ClassifyManagedRoots classifies every regular file under the given
// managed-root targets into exactly one of the four classes (REQ-UPM-001).
// targets is the same list the removal machinery scopes to (the caller
// passes the run's clean-target list, plus any additional root the pipeline
// must reconcile, e.g. .moai/config); nil selects the default managed
// targets. mf may be nil — a manifest-absent project classifies every
// template-carried file as user-modified (the conservative R-1 route) and
// every template-absent file as user-owned (never removal-eligible).
//
// The walk is read-only. A symlink entry anywhere under a target — and a
// symlinked target root itself — is recorded in Symlinks and skipped:
// classification never dereferences a link (REQ-UPM-004).
func ClassifyManagedRoots(projectRoot string, targets []deploy.CleanTarget, render TemplateRender, mf *manifest.Manifest) (ReconcilePlan, error) {
	planOut := ReconcilePlan{}
	if render == nil {
		return planOut, errors.New("reconcile classify: nil template render")
	}
	if targets == nil {
		targets = deploy.ManagedCleanTargets(projectRoot)
	}

	type found struct {
		rel   string
		bytes []byte
	}
	var files []found

	// resolveTarget expands one clean target (plain, glob, or directory)
	// into the walk roots it covers.
	resolveTarget := func(t deploy.CleanTarget) ([]string, error) {
		if !t.IsGlob {
			return []string{t.FullPath}, nil
		}
		matches, err := filepath.Glob(t.FullPath)
		if err != nil {
			return nil, err
		}
		return matches, nil
	}

	for _, t := range targets {
		roots, err := resolveTarget(t)
		if err != nil {
			return ReconcilePlan{Symlinks: planOut.Symlinks}, errors.Join(errReconcileClassifyStopped, err)
		}
		for _, root := range roots {
			info, err := os.Lstat(root)
			if err != nil {
				continue // target absent: nothing to classify under it
			}
			relRoot, relErr := filepath.Rel(projectRoot, root)
			if relErr != nil {
				return ReconcilePlan{Symlinks: planOut.Symlinks}, errors.Join(errReconcileClassifyStopped, relErr)
			}
			relRoot = filepath.ToSlash(relRoot)
			if info.Mode()&fs.ModeSymlink != 0 {
				// The target ITSELF is a link: record and move on.
				planOut.Symlinks = append(planOut.Symlinks, relRoot)
				continue
			}
			if !info.IsDir() {
				// A plain-file target is classified directly.
				data, readErr := os.ReadFile(root)
				if readErr != nil {
					return ReconcilePlan{Symlinks: planOut.Symlinks}, errors.Join(errReconcileClassifyStopped, readErr)
				}
				files = append(files, found{rel: relRoot, bytes: data})
				continue
			}
			walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
				// REQ-UPM-004: Lstat semantics — the link itself is the
				// entry. Never dereference, never classify, never merge it.
				if d.Type()&fs.ModeSymlink != 0 {
					planOut.Symlinks = append(planOut.Symlinks, rel)
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
				return ReconcilePlan{Symlinks: planOut.Symlinks}, errors.Join(errReconcileClassifyStopped, walkErr)
			}
		}
	}

	// Deterministic output order: sort the whole set once (NFR-UPM-001).
	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })
	sort.Strings(planOut.Symlinks)

	for _, f := range files {
		class := classifyManagedFile(f.rel, f.bytes, render, mf)
		file := ReconcileFile{RelPath: f.rel, Class: class}
		switch class {
		case ClassTemplateOwned:
			planOut.TemplateOwned = append(planOut.TemplateOwned, file)
		case ClassUserModified:
			planOut.UserModified = append(planOut.UserModified, file)
		case ClassUserOwned:
			planOut.UserOwned = append(planOut.UserOwned, file)
		case ClassStale:
			planOut.Stale = append(planOut.Stale, file)
		}
	}
	return planOut, nil
}

// classifyManagedFile decides the class of one regular file (REQ-UPM-001,
// design §2 decision table, ordered):
//
//  1. IsUserOwnedNamespace → user-owned, whatever the manifest says
//     (REQ-UPM-003, defense-in-depth over carriage);
//  2. template carries the path → classifyCarried decides template-owned
//     vs user-modified (the SPEC-INIT-SHRINK-001 logic, reused);
//  3. template does not carry it → stale when a prior template carried it
//     (a manifest record with managed provenance), else user-owned.
func classifyManagedFile(rel string, disk []byte, render TemplateRender, mf *manifest.Manifest) ReconcileClass {
	// REQ-UPM-003: the namespace predicate forces user-owned before any
	// carriage or manifest consideration.
	if plan.IsUserOwnedNamespace(rel) {
		return ClassUserOwned
	}

	rendered, carried := render.Carries(rel)
	if !carried {
		// Not carried by the current template. Removal-eligible ONLY when a
		// prior template carried it — a manifest record with managed
		// provenance is that evidence (REQ-UPM-014). Everything else is
		// user-owned: a moai-custom name the template does not carry is a
		// user file, never removal-eligible (REQ-UPM-002, the D-15 gate).
		entry := manifestEntry(mf, rel)
		if entry != nil && (entry.Provenance == manifest.TemplateManaged || entry.Provenance == manifest.Deprecated) {
			return ClassStale
		}
		return ClassUserOwned
	}

	// Template-carried: reuse the dropped-roots classifier's healthy-record
	// decision (identical → template-owned; absent/stale record or diverging
	// content → user-modified, conservatively).
	if classifyCarried(disk, rendered, manifestEntry(mf, rel)) == ClassIdentical {
		return ClassTemplateOwned
	}
	return ClassUserModified
}
