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
			switch provenance {
			case string(manifest.UserCreated):
				untouched++
				return nil
			case string(manifest.UserModified):
				preserved++
				report("  migration: preserved (you modified it): %s", relSlash)
				return nil
			default:
				// template_managed (or unrecorded): removable once the user
				// counterpart is confirmed.
			}
			if !userCounterpartConfirmed(userManifest, relSlash) {
				// Optional-pack (non-L0) asset without an opted-in selection,
				// or a failed counterpart write: stays project-side, reported.
				stayed++
				report("  migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update): %s", relSlash)
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
func userCounterpartConfirmed(userManifest *userassets.Manifest, projectRel string) bool {
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
	if !tracked {
		return false
	}
	return true && fe.SHA256 != ""
}
