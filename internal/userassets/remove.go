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

// removeEntry removes one entry's destination files per the one removal rule.
func (in *Installer) removeEntry(e template.Entry, manifest *Manifest, roots map[RootSlug]resolvedRoot, res *Result) error {
	keys := in.entryManifestKeys(e)
	for _, k := range keys {
		slug, rel, ok := splitManifestKey(k)
		if !ok {
			continue
		}
		root := roots[slug]
		abs := filepath.Join(root.dir, filepath.FromSlash(rel))
		record, tracked := manifest.Files[k]
		if !tracked {
			continue // nothing we installed at this path — never delete
		}
		current, readErr := os.ReadFile(abs)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				// Missing on disk: drop the entry, count removed (the truth
				// table's missing arm).
				delete(manifest.Files, k)
				res.Removed++
				continue
			}
			return fmt.Errorf("read %s: %w", abs, readErr)
		}
		currentSHA := sha256Hex(current)
		shipped, shippedErr := in.readShippedForKey(e, k)
		switch {
		case currentSHA == record.SHA256:
			// Manifest-hash match: remove.
		case shippedErr == nil && currentSHA == sha256Hex(shipped):
			// Shipped-bytes alternative (the manifest-stale arm): the bytes
			// are moai's own — remove under REQ-009's extended rule.
		default:
			// REQ-023 divergence: preserve + backup + report. A shipped
			// source exists here (the bundle definition persists in the
			// catalog), so the backup arm applies.
			if shippedErr == nil {
				if err := in.backupShipped(root, rel, shipped); err != nil {
					return fmt.Errorf("backup shipped bytes: %w", err)
				}
			}
			res.DivergencePreserved++
			res.Divergences = append(res.Divergences, k)
			continue
		}
		// RF1 (review fix): re-validate the destination parent immediately
		// before the delete — a parent swapped to an outside-pointing
		// symlink between the root resolution and here must not let
		// os.Remove reach a file beyond the root boundary.
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
	}
	return nil
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
	case RootClaudeSkills:
		return strings.TrimSuffix(srcPath(e.Path), "/") + "/" + strings.TrimPrefix(rel, e.Name+"/"), true
	case RootAgentsSkills:
		// The codex-root skill tree mirrors the .claude skill tree.
		claudeRel := strings.TrimPrefix(rel, e.Name+"/")
		return ".claude/skills/" + e.Name + "/" + claudeRel, true
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
