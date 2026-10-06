// remove.go — bundle removal and the manifest-driven removal rule
// (SPEC-USER-ASSET-INSTALL-001 REQ-009, the E3 complement, R-f-② dependency
// maintenance).
//
// THE ONE REMOVAL RULE (iter4 D25): a manifest-tracked file is removed only
// when its current hash equals its manifest hash, OR equals the shipped bytes
// where a shipped source still exists. A file matching neither is REQ-023
// divergence: preserved in place, the shipped replacement backed up under
// ~/.moai/backups/<root-slug>/, reported. A tracked file missing on disk
// drops its manifest entry and counts removed.
package userassets

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/template"
)

// PruneUnselected applies the selection-based criterion (REQ-009): every
// manifest-tracked file whose owning entry is no longer in (L0 ∪ the
// manifest's recorded selection) is removed under the one removal rule —
// manifest-hash match, or shipped-bytes match where a shipped source exists;
// a file matching neither is REQ-023 divergence (preserved + backed up +
// reported); a file missing on disk drops its entry and counts removed.
// `moai update` runs this against the manifest's recorded selection so a
// `moai bundle remove`d bundle's artifact files are pruned (AC-018's D28
// flip-criterion arm).
func (in *Installer) PruneUnselected(manifest *Manifest) (*Result, error) {
	res := &Result{}
	if err := manifest.CanRemove(); err != nil {
		return nil, err
	}
	roots := make(map[RootSlug]resolvedRoot, 4)
	for _, r := range ResolveRoots(in.Home) {
		rr, err := resolveRoot(in.Home, r)
		if err != nil {
			return nil, fmt.Errorf("resolve root %s: %w", r.Slug, err)
		}
		roots[r.Slug] = rr
	}
	allowed := map[string]bool{}
	preserved := []template.Entry{}
	for _, e := range in.preservedEntries(manifest.Bundles) {
		allowed[e.Name] = true
		preserved = append(preserved, e)
	}
	// RF2 (review fix): the prune applies the same R-f-② dependency-
	// deferral rule as RemoveBundle — an entry that is a declared
	// dependency of a PRESERVED entry is kept + reported (the dispatcher's
	// own matrix-class dep list stays excluded, for the D28 reason recorded
	// on RemoveBundle).
	deferral := map[string]bool{}
	for _, e := range preserved {
		if e.Name == "moai" {
			continue
		}
		for _, d := range e.DependsSkills {
			deferral[d] = true
		}
		for _, d := range e.DependsAgents {
			deferral[d] = true
		}
	}

	// F9 (review-fix round 2): the allowed-name judgment keeps a SKILL, but
	// files REMOVED from the new version inside a kept skill must still go:
	// the judgment is per KEY, not per name. For an allowed entry the
	// expected key set comes from the CURRENT catalog entry; tracked keys
	// outside it are retired files and follow the one removal rule.
	allowedKeys := map[string]bool{}
	for _, e := range preserved {
		if allowed[e.Name] {
			for _, k := range in.entryManifestKeys(e) {
				allowedKeys[k] = true
			}
		}
	}

	// Snapshot the keys: removal mutates the map.
	keys := make([]string, 0, len(manifest.Files))
	for k := range manifest.Files {
		keys = append(keys, k)
	}
	deferredNames := map[string]bool{}
	for _, k := range keys {
		slug, rel, ok := splitManifestKey(k)
		if !ok {
			continue
		}
		name, ok := owningEntryName(rel)
		if !ok || allowed[name] {
			// A kept skill's RETIRED file (removed upstream) still follows
			// the one removal rule — the shipped source is gone, so only
			// the manifest-hash alternative applies.
			if allowed[name] && !allowedKeys[k] {
				entry, found := in.lookupEntry(name)
				if !found {
					entry = template.Entry{Name: name}
				}
				if err := in.removeManifestKey(entry, k, slug, rel, roots[slug], manifest, res); err != nil {
					res.Failures = append(res.Failures, FileOutcome{Path: k, Reason: err.Error()})
				}
			}
			continue
		}
		if deferral[name] {
			if !deferredNames[name] {
				deferredNames[name] = true
				res.DeferredDeps = append(res.DeferredDeps, name)
			}
			continue
		}
		entry, found := in.lookupEntry(name)
		if !found {
			// No shipped source anywhere (the entry left the catalog too):
			// only the manifest-hash alternative applies (REQ-009 iter4 D25).
			entry = template.Entry{Name: name}
		}
		if err := in.removeManifestKey(entry, k, slug, rel, roots[slug], manifest, res); err != nil {
			res.Failures = append(res.Failures, FileOutcome{Path: k, Reason: err.Error()})
		}
	}
	return res, nil
}

// owningEntryName derives the catalog entry name from a manifest key's
// relpath: skills are directory-rooted ("<name>/..."), agents are flat
// files ("<name>.md" / "<name>.toml").
func owningEntryName(rel string) (string, bool) {
	if i := strings.Index(rel, "/"); i > 0 {
		return rel[:i], true
	}
	name := rel
	for _, suffix := range []string{".md", ".toml"} {
		if trimmed, ok := strings.CutSuffix(name, suffix); ok {
			return trimmed, true
		}
	}
	return "", false
}

// lookupEntry finds a catalog entry by name across all sections.
func (in *Installer) lookupEntry(name string) (template.Entry, bool) {
	for _, e := range in.Catalog.AllEntries() {
		if e.Name == name {
			return e, true
		}
	}
	return template.Entry{}, false
}

// removeManifestKey applies the one removal rule to a single manifest key.
func (in *Installer) removeManifestKey(entry template.Entry, k string, slug RootSlug, rel string, root resolvedRoot, manifest *Manifest, res *Result) error {
	record, tracked := manifest.Files[k]
	if !tracked {
		return nil
	}
	// Item 2 (fix round 3 addendum): an unknown root slug resolves to an
	// EMPTY resolvedRoot — os.Remove(abs) then falls back to the CURRENT
	// WORKING DIRECTORY and deletes an external file (the gate's sentinel
	// repro). Refuse the key outright.
	if root.dir == "" {
		return fmt.Errorf("userassets: unknown root slug %q for key %q — deletion refused", slug, k)
	}
	if clean, err := ValidateRelPath(rel); err != nil || clean != rel {
		return fmt.Errorf("userassets: rel path %q escapes its root — deletion refused", rel)
	}
	abs := filepath.Join(root.dir, filepath.FromSlash(rel))
	current, readErr := os.ReadFile(abs)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			delete(manifest.Files, k)
			res.Removed++
			return nil
		}
		return fmt.Errorf("read %s: %w", abs, readErr)
	}
	currentSHA := sha256Hex(current)
	shipped, shippedErr := in.readShippedForKey(entry, k)
	switch {
	case currentSHA == record.SHA256:
		// Manifest-hash match: remove.
	case shippedErr == nil && currentSHA == sha256Hex(shipped):
		// Shipped-bytes alternative (REQ-009 iter4 D25).
	default:
		// REQ-023 divergence: preserve + backup + report.
		if shippedErr == nil {
			if err := in.backupShipped(root, rel, shipped); err != nil {
				return fmt.Errorf("backup shipped bytes: %w", err)
			}
		}
		res.DivergencePreserved++
		res.Divergences = append(res.Divergences, k)
		return nil
	}
	// F1 (review-fix round 2): re-validate the destination parent
	// immediately before the delete — a parent swapped to an
	// outside-pointing symlink between the root resolution and here
	// must not route os.Remove to a file beyond the boundary (the C2
	// posture the write path already carries).
	parentResolved, parentErr := filepath.EvalSymlinks(filepath.Dir(abs))
	if parentErr != nil || !withinRoot(root.dir, parentResolved) {
		return fmt.Errorf("userassets: delete parent re-validation failed — refused (C2 posture)")
	}
	if err := os.Remove(abs); err != nil {
		return err
	}
	delete(manifest.Files, k)
	res.Removed++
	in.pruneEmptyDirs(root, rel)
	return nil
}

// RemoveBundle removes the entries of the named bundle that are NOT in
// (L0 ∪ the remaining selection) — the E3 complement. Shared entries survive
// with a report note; an entry that is a declared dependency of a PRESERVED
// asset has its deletion deferred (kept + reported, R-f-②). The manifest
// mutation (entry drops) happens here; the caller saves the manifest and the
// updated selection under the user lock.
func (in *Installer) RemoveBundle(manifest *Manifest, name string, remaining []string) (*Result, error) {
	res := &Result{}
	if err := manifest.CanRemove(); err != nil {
		return nil, err // REQ-021 refusal against an unknown schema
	}
	// Resolve the roots once (same posture as Install).
	roots := make(map[RootSlug]resolvedRoot, 4)
	for _, r := range ResolveRoots(in.Home) {
		rr, err := resolveRoot(in.Home, r)
		if err != nil {
			return nil, fmt.Errorf("resolve root %s: %w", r.Slug, err)
		}
		roots[r.Slug] = rr
	}

	// Preserved = L0 ∪ remaining selections. Their declared dependencies are
	// the R-f-② deferral set — EXCLUDING the dispatcher's own dep list (its
	// bundle-row edges are the derivation matrix's conditional class; their
	// absence is handled by the remediation/refusal pattern, and including
	// them would defeat the D28 selection-based prune).
	deferral := map[string]bool{}
	for _, e := range in.preservedEntries(remaining) {
		if e.Name == "moai" {
			continue // the dispatcher — the matrix-class dep holder
		}
		for _, d := range e.DependsSkills {
			deferral[d] = true
		}
		for _, d := range e.DependsAgents {
			deferral[d] = true
		}
	}

	preserved := map[string]bool{}
	for _, e := range in.preservedEntries(remaining) {
		preserved[e.Name] = true
	}

	if _, ok := in.Catalog.Catalog.OptionalPacks[name]; !ok {
		return nil, fmt.Errorf("unknown bundle %q", name)
	}
	// RF3 (review fix): enumerate the removal targets from the MANIFEST —
	// every tracked key whose recorded bundle names the removed bundle.
	// Files installed by an OLDER deployment are tracked but absent from
	// the current source tree; walking the tree misses them and both the
	// file and its record survive the removal.
	bundleKeys := map[string]bool{}
	for k, fe := range manifest.Files {
		if fe.Bundle == name {
			bundleKeys[k] = true
		}
	}

	// Group the keys by owning entry name so the E3/R-f-② classification
	// (shared / deferred) runs per entry as before.
	nameKeys := map[string][]string{}
	for k := range bundleKeys {
		_, rel, ok := splitManifestKey(k)
		if !ok {
			continue
		}
		n, ok := owningEntryName(rel)
		if !ok {
			continue
		}
		nameKeys[n] = append(nameKeys[n], k)
	}
	names := make([]string, 0, len(nameKeys))
	for n := range nameKeys {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, n := range names {
		if preserved[n] {
			res.SharedSurvivors = append(res.SharedSurvivors, n)
			continue
		}
		if deferral[n] {
			res.DeferredDeps = append(res.DeferredDeps, n)
			continue
		}
		entry, found := in.lookupEntry(n)
		if !found {
			entry = template.Entry{Name: n}
		}
		if err := in.removeKeys(entry, nameKeys[n], manifest, roots, res); err != nil {
			res.Failures = append(res.Failures, FileOutcome{Path: n, Reason: err.Error()})
		}
	}
	return res, nil
}

// removeKeys applies the one removal rule to an explicit key set (the RF3
// manifest-driven enumeration).
func (in *Installer) removeKeys(entry template.Entry, keys []string, manifest *Manifest, roots map[RootSlug]resolvedRoot, res *Result) error {
	for _, k := range keys {
		slug, rel, ok := splitManifestKey(k)
		if !ok {
			continue
		}
		if err := in.removeManifestKey(entry, k, slug, rel, roots[slug], manifest, res); err != nil {
			res.Failures = append(res.Failures, FileOutcome{Path: k, Reason: err.Error()})
		}
	}
	return nil
}

// preservedEntries gathers L0 ∪ the named selection's entries.
func (in *Installer) preservedEntries(selection []string) []template.Entry {
	entries := make([]template.Entry, 0, len(in.Catalog.Catalog.Core.Skills)+len(in.Catalog.Catalog.Core.Agents))
	entries = append(entries, in.Catalog.Catalog.Core.Skills...)
	entries = append(entries, in.Catalog.Catalog.Core.Agents...)
	for _, name := range selection {
		if pack, ok := in.Catalog.Catalog.OptionalPacks[name]; ok {
			entries = append(entries, pack.Skills...)
			entries = append(entries, pack.Agents...)
		}
	}
	return entries
}

// entryManifestKeys lists the manifest keys one entry's destinations carry.
func (in *Installer) entryManifestKeys(e template.Entry) []string {
	var keys []string
	switch {
	case strings.HasSuffix(e.Path, "/"):
		srcDir := strings.TrimSuffix(srcPath(e.Path), "/")
		_ = fs.WalkDir(in.Source, srcDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel := strings.TrimPrefix(p, srcDir+"/")
			keys = append(keys,
				string(RootClaudeSkills)+"/"+e.Name+"/"+rel,
				string(RootAgentsSkills)+"/"+e.Name+"/"+rel)
			return nil
		})
	case strings.HasSuffix(e.Path, ".md"):
		keys = append(keys, string(RootClaudeAgents)+"/"+e.Name+".md")
		if _, err := fs.Stat(in.Source, ".codex/agents/moai/"+e.Name+".toml"); err == nil {
			keys = append(keys, string(RootCodexAgents)+"/"+e.Name+".toml")
		}
	}
	return keys
}

// readShippedForKey reads the shipped bytes for one destination key.
func (in *Installer) readShippedForKey(e template.Entry, key string) ([]byte, error) {
	slug, rel, ok := splitManifestKey(key)
	if !ok {
		return nil, fmt.Errorf("bad key %q", key)
	}
	sourcePath, ok := in.sourcePathFor(slug, e, rel)
	if !ok {
		return nil, fmt.Errorf("no shipped source for %s", key)
	}
	return fs.ReadFile(in.Source, sourcePath)
}

// sourcePathFor maps a (root slug, entry, rel) triple back to the source
// tree path (stripped FS form).
func (in *Installer) sourcePathFor(slug RootSlug, e template.Entry, rel string) (string, bool) {
	switch slug {
	case RootClaudeSkills, RootAgentsSkills:
		// F10 (review-fix round 2): use the entry's ACTUAL catalog path —
		// published command skills live under .agents/skills, not .claude/
		// skills, and their divergence backups were silently skipped.
		// (path.Join also collapses the entry path's trailing slash — a
		// double slash makes fstest.MapFS reads fail in tests.)
		return path.Join(srcPath(e.Path), strings.TrimPrefix(rel, e.Name+"/")), true
	case RootClaudeAgents:
		return srcPath(e.Path), true
	case RootCodexAgents:
		return ".codex/agents/moai/" + e.Name + ".toml", true
	}
	return "", false
}

// pruneEmptyDirs removes now-empty directories below a root after a file
// removal (the bundle's skill directory should not linger as an empty shell).
func (in *Installer) pruneEmptyDirs(root resolvedRoot, rel string) {
	dir := filepath.Dir(filepath.Join(root.dir, filepath.FromSlash(rel)))
	for i := 0; i < 8; i++ {
		if !withinRoot(root.dir, dir) || dir == root.dir {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}
