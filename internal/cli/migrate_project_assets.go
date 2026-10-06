// migrate_project_assets.go — the REQ-020 project migration (SPEC-USER-
// ASSET-INSTALL-001 M4): classify existing project-side common skills and
// agents, and remove each template-managed file ONLY after its user
// counterpart is confirmed (manifest-tracked with a matching hash — REQ-024's
// upgrade arm per-asset gate, iter4 D24). user_modified files are preserved
// with a report (C6 — never silently absorbed); user_created files are
// untouched. A file whose counterpart is unconfirmed stays project-side and
// is reported.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// projectCommonAssetRels enumerates the project-relative common-asset roots
// M4 stops emitting (REQ-005) and migrates (REQ-020).
var projectCommonAssetRels = []string{
	".claude/skills/",
	".agents/skills/",
	".claude/agents/moai/",
	".codex/agents/moai/",
}

// migrateProjectCommonAssets removes template-managed common skills and
// agents from the project after confirming each user counterpart (REQ-020).
// The user manifest is read-only here (the user-asset phase already ran and
// holds the confirmed install); the project manifest is updated to drop the
// removed entries.
func migrateProjectCommonAssets(projectRoot, homeDir string, out fmt.Stringer, report func(string, ...interface{})) error {
	migrationPreservedProjectFiles = map[string]bool{}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		// No project manifest: nothing provenance-classified to migrate.
		return nil
	}
	userManifest, err := userassets.Load(userassets.ManifestPath(homeDir))
	if err != nil {
		return fmt.Errorf("load user manifest for migration: %w", err)
	}

	var removed, preserved, stayed, untouched int
	// Item 1 (fix round 3 addendum): anchor the deletion boundary to the
	// project root — a .claude (or sibling root) swapped to an
	// outside-pointing symlink must not route os.Remove to an external
	// sentinel even when the recorded hash matches.
	rootResolved, rootErr := filepath.EvalSymlinks(projectRoot)
	if rootErr != nil {
		return fmt.Errorf("resolve project root: %w", rootErr)
	}
	for _, rootRel := range projectCommonAssetRels {
		absRoot := filepath.Join(projectRoot, filepath.FromSlash(rootRel))
		err := filepath.WalkDir(absRoot, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return fs.SkipDir
				}
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(projectRoot, p)
			if relErr != nil {
				return nil
			}
			relSlash := filepath.ToSlash(rel)

			entry, hasEntry := mgr.GetEntry(relSlash)
			provenance := ""
			if hasEntry && entry != nil {
				provenance = string(entry.Provenance)
			}
			// F3 (review-fix round 2): judge the CURRENT bytes, not the
			// recorded provenance alone — a file the user modified AFTER
			// install still reads template_managed. When the on-disk hash
			// differs from the recorded template hash, the file is
			// user-modified in fact and is preserved (C6). An untracked
			// file (no manifest entry) is user-owned — never deleted
			// (AGENTS.local.md user-settings protection).
			if !hasEntry || entry == nil {
				untouched++
				migrationPreservedProjectFiles[p] = true
				return nil
			}
			if provenance == string(manifest.UserCreated) {
				untouched++
				return nil
			}
			// Item 1 (fix round 3): the deletion preconditions are
			// MANDATORY — a failed read or an empty recorded template hash
			// PRESERVES the file (never falls through to deletion). Deletion
			// requires: bytes readable AND template hash recorded AND
			// current bytes matching it AND the user counterpart confirmed.
			currentBytes, readErr := os.ReadFile(p)
			if readErr != nil || entry.TemplateHash == "" {
				preserved++
				migrationPreservedProjectFiles[p] = true
				report("  migration: preserved (cannot verify against the template — read failed or no template hash): %s", relSlash)
				return nil
			}
			if manifest.HashBytes(currentBytes) != entry.TemplateHash {
				preserved++
				migrationPreservedProjectFiles[p] = true
				report("  migration: preserved (bytes differ from the template — treated as user-modified): %s", relSlash)
				return nil
			}
			if provenance == string(manifest.UserModified) {
				preserved++
				migrationPreservedProjectFiles[p] = true
				report("  migration: preserved (you modified it): %s", relSlash)
				return nil
			}
			// template_managed with template-matching bytes: removable once
			// the user counterpart is confirmed.
			if !userCounterpartConfirmed(userManifest, homeDir, relSlash) {
				// Optional-pack (non-L0) asset without an opted-in selection,
				// or a failed counterpart write: stays project-side, reported
				// — and gated out of the cleanup list (F2).
				stayed++
				migrationPreservedProjectFiles[p] = true
				report("  migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update): %s", relSlash)
				return nil
			}
			resolvedP, resErr := filepath.EvalSymlinks(p)
			if resErr != nil || !withinRootBoundary(rootResolved, resolvedP) {
				report("  migration: removal refused (path escapes the project root through a symlink): %s", relSlash)
				return nil
			}
			if err := os.Remove(p); err != nil {
				report("  migration: removal failed: %s: %v", relSlash, err)
				return nil
			}
			removed++
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("walk %s: %w", rootRel, err)
		}
	}

	// Drop manifest entries whose file is gone (removed by the loop above or
	// a prior run), then save once.
	dropped := 0
	for p := range mgr.Manifest().Files {
		if _, err := os.Stat(filepath.Join(projectRoot, p)); os.IsNotExist(err) {
			_ = mgr.Remove(p)
			dropped++
		}
	}
	if removed > 0 || dropped > 0 || preserved > 0 || stayed > 0 {
		if err := mgr.Save(); err != nil {
			return fmt.Errorf("save project manifest after migration: %w", err)
		}
		report("migration: %d removed, %d preserved (user-modified), %d kept project-side, %d user-created untouched", removed, preserved, stayed, untouched)
	}
	return nil
}

// userCounterpartConfirmed reports whether the project file's user-side
// counterpart is manifest-tracked with a hash matching the installed bytes
// (REQ-020's per-asset gate). The mapping is mechanical: a project path
// .claude/skills/<name>/X ↔ user key claude-skills/<name>/X;
// .agents/skills/<name>/X ↔ agents-skills/<name>/X;
// .claude/agents/moai/<n>.md ↔ claude-agents/<n>.md;
// .codex/agents/moai/<n>.toml ↔ codex-agents/<n>.toml.
func userCounterpartConfirmed(userManifest *userassets.Manifest, homeDir, projectRel string) bool {
	var userKey string
	switch {
	case strings.HasPrefix(projectRel, ".claude/skills/"):
		userKey = "claude-skills/" + strings.TrimPrefix(projectRel, ".claude/skills/")
	case strings.HasPrefix(projectRel, ".agents/skills/"):
		userKey = "agents-skills/" + strings.TrimPrefix(projectRel, ".agents/skills/")
	case strings.HasPrefix(projectRel, ".claude/agents/moai/"):
		userKey = "claude-agents/" + strings.TrimPrefix(projectRel, ".claude/agents/moai/")
	case strings.HasPrefix(projectRel, ".codex/agents/moai/"):
		userKey = "codex-agents/" + strings.TrimPrefix(projectRel, ".codex/agents/moai/")
	default:
		return false
	}
	fe, tracked := userManifest.Files[userKey]
	if !tracked || fe.SHA256 == "" {
		return false
	}
	// F4 (review-fix round 2): the recorded hash alone is NOT confirmation —
	// the counterpart FILE must exist on disk and its CURRENT bytes must
	// hash to the recorded value. A stale record from a failed install
	// must never authorize the project-side deletion.
	slug, rel, ok := strings.Cut(userKey, "/")
	if !ok {
		return false
	}
	dir := userassets.RootBySlugDir(homeDir, userassets.RootSlug(slug))
	if dir == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		return false
	}
	// The USER manifest stores bare hex (its own convention); the PROJECT
	// manifest stores the 'sha256:<hex>' prefix form — the two formats are
	// intentionally distinct (B4).
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]) == fe.SHA256
}

// withinRootBoundary reports whether path is inside (or equal to) the
// resolved root (the same containment predicate the userassets package
// uses; duplicated here to avoid importing the package for one helper).
func withinRootBoundary(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && rel != "")
}
