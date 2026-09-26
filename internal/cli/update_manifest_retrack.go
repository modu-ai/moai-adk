package cli

// update_manifest_retrack.go — card t1275.
//
// `moai update` rewrites deployed files AFTER the Deploy Templates step
// finished tracking them: the Restore Settings step rewrites
// .moai/config/sections/*.yaml via 3-way merge, mergeUserFilesSettlingSnapshot
// rewrites the mergeable set (.claude/settings.json, .moai/status_line.sh,
// .mcp.json, ...), and profile.SyncToProjectConfig rewrites user/language/
// statusline sections post-sync. None of those rewrites used to re-record the
// manifest, and the sync flow never called manifest Save at all — so the
// manifest kept pre-update hashes while the files on disk moved, and the next
// `init --force` (CarryManifestForward) reclassified every drifted file
// user_modified (the four-file freeze of card t1275's origin, plus a
// same-version 24-file config drift measured in isolation).
//
// retrackManifestFiles re-records exactly the paths THIS update just rewrote
// — never a disk-wide sweep: a file the user edited outside any update is not
// in the caller's path list, so its drift is preserved and it still reads as
// user_modified (the two-way invariant the lead set as the acceptance bar).

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// retrackManifestFiles re-tracks the given repo-relative paths in the manifest
// as template_managed with their on-disk hash, preserving each entry's
// recorded TemplateHash, then saves. Paths that do not exist on disk, are not
// regular files, or have no existing manifest entry are skipped silently:
// retracking is bookkeeping for files this update wrote, not a classification
// pass for unknown files. Save runs once after the loop; a Save failure is
// returned (callers report it as a warning, never fail the update for it —
// a stale hash degrades the next init --force, it does not break this one).
func retrackManifestFiles(projectRoot string, mgr manifest.Manager, errOut io.Writer, rels []string) error {
	if mgr == nil || mgr.Manifest() == nil {
		return nil
	}
	dirty := false
	for _, rel := range rels {
		entry, found := mgr.GetEntry(rel)
		if !found || entry.Provenance != manifest.TemplateManaged {
			continue
		}
		path := filepath.Join(projectRoot, filepath.FromSlash(rel))
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		if err := mgr.Track(rel, manifest.TemplateManaged, entry.TemplateHash); err != nil {
			// One unreadable file must not abandon the bookkeeping for the
			// rest; record it and keep going.
			_, _ = fmt.Fprintf(errOut, "  manifest retrack skip %s: %v\n", rel, err)
			continue
		}
		dirty = true
	}
	if !dirty {
		return nil
	}
	return mgr.Save()
}

// retrackPaths is the self-loading form of retrackManifestFiles for call
// sites that do not already hold a manifest manager (the post-sync steps and
// the best-effort repair paths). Same contract: only template_managed entries
// with an existing manifest record are re-recorded.
func retrackPaths(projectRoot string, errOut io.Writer, rels []string) error {
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		// No loadable manifest means nothing to retrack against.
		return nil
	}
	return retrackManifestFiles(projectRoot, mgr, errOut, rels)
}

// retrackSectionFiles retracks every .moai/config/sections/*.yaml the project
// currently carries. The Restore Settings step rewrites that directory, and
// profile.SyncToProjectConfig rewrites a subset of it afterwards — both after
// the deploy already tracked the pre-merge render. Listing the directory (not
// a hardcoded set) keeps new section files covered without touching this
// call site; the template_managed filter inside retrackManifestFiles is what
// keeps user-owned files out.
func retrackSectionFiles(projectRoot string, errOut io.Writer) {
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		// No loadable manifest means nothing to retrack against; the deploy
		// step's own Save is the one that matters for a fresh project.
		return
	}
	matches, err := filepath.Glob(filepath.Join(projectRoot, ".moai", "config", "sections", "*.yaml"))
	if err != nil || len(matches) == 0 {
		return
	}
	rels := make([]string, 0, len(matches))
	for _, m := range matches {
		rel, relErr := filepath.Rel(projectRoot, m)
		if relErr != nil {
			continue
		}
		rels = append(rels, filepath.ToSlash(rel))
	}
	if err := retrackManifestFiles(projectRoot, mgr, errOut, rels); err != nil {
		_, _ = fmt.Fprintf(errOut, "  manifest retrack (config sections): %v\n", err)
	}
}
