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
	"strings"

	"github.com/modu-ai/moai-adk/internal/template"
)

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

	removedBundle, ok := in.Catalog.Catalog.OptionalPacks[name]
	if !ok {
		return nil, fmt.Errorf("unknown bundle %q", name)
	}
	var candidates []template.Entry
	candidates = append(candidates, removedBundle.Skills...)
	candidates = append(candidates, removedBundle.Agents...)

	for _, e := range candidates {
		if preserved[e.Name] {
			res.SharedSurvivors = append(res.SharedSurvivors, e.Name)
			continue
		}
		if deferral[e.Name] {
			res.DeferredDeps = append(res.DeferredDeps, e.Name)
			continue
		}
		if err := in.removeEntry(e, manifest, roots, res); err != nil {
			res.Failures = append(res.Failures, FileOutcome{Path: e.Name, Reason: err.Error()})
		}
	}
	return res, nil
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
