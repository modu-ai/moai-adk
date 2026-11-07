// doctor_user_install.go — the `moai doctor` user-install integrity check
// (SPEC-USER-ASSET-INSTALL-001 M5; REQ-014) and the project-vs-lock
// comparison (REQ-015). Read-only, report-only, doctor house style.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// checkUserInstallIntegrity compares every manifest-tracked path against the
// installed user tree: missing, hash-modified, and untracked entries in the
// four roots are reported. Read-only; a corrupt manifest is reported (never
// auto-deleted — AC-021's rebuild-from-scan offer rides this row).
func checkUserInstallIntegrity(homeDir string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "User Install"}

	manifestPath := userassets.ManifestPath(homeDir)
	m, err := userassets.Load(manifestPath)
	if err != nil {
		var ce *userassets.CorruptError
		if userassets.AsCorrupt(err, &ce) {
			check.Status = uikit.CheckWarn
			check.Message = "user-assets.json corrupt — manifest-driven update/removal refused; run 'moai doctor --json' for the path and rebuild from a fresh init"
			return check
		}
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf("cannot read user manifest: %v", err)
		return check
	}
	if len(m.Files) == 0 {
		check.Status = uikit.CheckOK
		check.Message = "no per-user install recorded (run 'moai init')"
		return check
	}

	rootList, err := userassets.ResolveRoots(homeDir)
	if err != nil {
		// An enumeration failure must surface as a warn, never as the "no
		// profiles exist" disguise — with the store unreadable, the manifest
		// comparison below could not see the profile roots at all (card
		// t1702 audit round 4).
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("cannot enumerate the install roots (profile store unreadable) — tracked files not compared: %v", err)
		return check
	}
	rootDirs := make(map[userassets.RootSlug]string, len(rootList))
	for _, r := range rootList {
		rootDirs[r.Slug] = r.Dir
	}

	verified, missing, modified, unverifiable := 0, 0, 0, 0
	for key, fe := range m.Files {
		slug, rel, ok := splitManifestKeyUser(key)
		if !ok {
			continue
		}
		dir, known := rootDirs[slug]
		if !known {
			// F3 (card t1702 audit iter4): a record under a root that is not
			// currently enumerated (a deleted profile, say) was skipped
			// while the summary still counted len(m.Files) as verified.
			// Count it separately — never as verified.
			unverifiable++
			if verbose {
				check.Detail += fmt.Sprintf("\nunverifiable (root not enumerated): %s", key)
			}
			continue
		}
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		data, readErr := os.ReadFile(abs)
		if readErr != nil {
			missing++
			if verbose {
				check.Detail += fmt.Sprintf("\nmissing: %s", key)
			}
			continue
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != fe.SHA256 {
			modified++
			if verbose {
				check.Detail += fmt.Sprintf("\nmodified: %s", key)
			}
			continue
		}
		verified++
	}

	untracked := countUntrackedUserFiles(rootList, m)
	switch {
	case missing > 0 || modified > 0:
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("%d manifest-tracked file(s) missing, %d modified — 'moai update' repairs manifest-hash matches; edits stay preserved", missing, modified)
		if unverifiable > 0 {
			check.Message += fmt.Sprintf("; %d record(s) under roots not currently enumerated — not counted as verified", unverifiable)
		}
	case unverifiable > 0:
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("%d file(s) verified; %d record(s) under roots not currently enumerated (stale profile or unreadable store) — not counted as verified; 'moai update' reconciles", verified, unverifiable)
	case untracked > 0:
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("%d file(s) verified; %d untracked file(s) in the user roots (informational, never removed)", verified, untracked)
	default:
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("%d file(s) verified", verified)
	}
	return check
}

// countUntrackedUserFiles walks the four roots and counts files the manifest
// does not track (informational only — REQ-010 keeps them inviolable).
func countUntrackedUserFiles(roots []userassets.Root, m *userassets.Manifest) int {
	count := 0
	for _, r := range roots {
		_ = filepath.WalkDir(r.Dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(r.Dir, p)
			if relErr != nil {
				return nil
			}
			if _, tracked := m.Files[string(r.Slug)+"/"+filepath.ToSlash(rel)]; !tracked {
				count++
			}
			return nil
		})
	}
	return count
}

// splitManifestKeyUser splits a user-manifest key into (root slug, relpath).
func splitManifestKeyUser(key string) (userassets.RootSlug, string, bool) {
	slug, rest, ok := strings.Cut(key, "/")
	if !ok {
		return "", "", false
	}
	return userassets.RootSlug(slug), rest, true
}

// checkProfileSkillReach verifies that ONE active Claude config root can
// resolve the core skills a standard install lands: a SKILL.md per
// embedded-catalog core entry under <root>/skills/<name>/ (card t1702). The
// caller resolves the active root (CLAUDE_CONFIG_DIR when set, else
// ~/.claude) plus its DISPLAY form — the env value verbatim, or the literal
// "~/.claude" for the fallback — so the rendered message stays free of
// machine-specific absolute paths (the doctor golden snapshots require that
// hermeticity). Repair 3 (card t1702 audit P2): when the active root is
// OUTSIDE moai's install roots (a user's own config dir), the warning names
// what would actually fix it instead of advising 'moai update', which
// cannot reach that root. Report-only: it never installs.
func checkProfileSkillReach(activeRoot, displayRoot, homeDir string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Skill Reach"}

	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("cannot load embedded catalog: %v", err)
		return check
	}
	core := cat.Catalog.Core.Skills
	if len(core) == 0 {
		check.Status = uikit.CheckOK
		check.Message = "no core skills in the embedded catalog"
		return check
	}

	missing := 0
	for _, e := range core {
		if _, err := os.Stat(filepath.Join(activeRoot, "skills", e.Name, "SKILL.md")); err == nil {
			continue
		}
		missing++
		if verbose {
			check.Detail += fmt.Sprintf("\nmissing: %s", e.Name)
		}
	}
	if missing > 0 {
		check.Status = uikit.CheckWarn
		installable, enumErr := activeRootIsInstallable(homeDir, activeRoot)
		switch {
		case enumErr != nil:
			// The store cannot be inspected, so "outside moai's install
			// roots" would be an unobserved claim — name the enumeration
			// failure instead (card t1702 audit round 4).
			check.Message = fmt.Sprintf("%d of %d core skill(s) NOT resolvable from %s — cannot confirm this config root is moai-managed (profile store unreadable): %v", missing, len(core), displayRoot, enumErr)
		case installable:
			check.Message = fmt.Sprintf("%d of %d core skill(s) NOT resolvable from %s — run 'moai update' to install into the active config root", missing, len(core), displayRoot)
		default:
			check.Message = fmt.Sprintf("%d of %d core skill(s) NOT resolvable from %s — this config dir is outside moai's install roots (~/.claude and the profile dirs under ~/.moai/claude-profiles); point CLAUDE_CONFIG_DIR at a managed root or copy the core skills in manually", missing, len(core), displayRoot)
		}
		return check
	}
	check.Status = uikit.CheckOK
	check.Message = fmt.Sprintf("%d core skill(s) resolvable from %s", len(core), displayRoot)
	return check
}

// activeRootIsInstallable reports whether 'moai update' can install into
// the active root: the root must be ~/.claude or exactly one of the real
// profile dirs under the profile base — the two families ResolveRoots
// enumerates (repair 3, card t1702 audit P2). A profile store that exists
// but cannot be read yields an error so the caller can distinguish "outside
// the families" from "cannot judge" (card t1702 audit round 4).
func activeRootIsInstallable(homeDir, activeRoot string) (bool, error) {
	if activeRoot == filepath.Join(homeDir, ".claude") {
		return true, nil
	}
	base := userassets.ProfileRootsBase(homeDir)
	names, err := userassets.RealProfileDirNames(homeDir)
	if err != nil {
		return false, err
	}
	for _, name := range names {
		if activeRoot == filepath.Join(base, name) {
			return true, nil
		}
	}
	return false, nil
}

// checkProjectVsLock compares the project tree against the project lock file
// in BOTH directions (REQ-015): a project file absent from the lock, and a
// lock entry absent from the project. Read-only; the lock schema is not
// touched (C3).
//
// M7 (REQ-DOC-001/002, SPEC-USERASSET-DEPLOY-GUARD-001):
//   - the manifest is READ FIRST and judged read-only: a corrupt manifest
//     is reported at a failure grade (never the CheckOK disguise) and the
//     manager's quarantining Load is never reached, so the corrupt
//     original stays at its path and a pre-existing .corrupt recovery copy
//     is never overwritten (the audited double assertion);
//   - a tree that offers NO lock-comparable surface reports the vacuous
//     condition instead of the "matches the lock file" claim (AC-020).
func checkProjectVsLock(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Project Lock"}

	manifestPath := filepath.Join(projectRoot, ".moai", "manifest.json")
	manifestData, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		if !os.IsNotExist(readErr) {
			// present but unreadable: surfaced, never disguised
			check.Status = uikit.CheckWarn
			check.Message = fmt.Sprintf("project manifest unreadable: %v", readErr)
			return check
		}
		// genuinely absent: nothing to compare — the honest shape
		check.Status = uikit.CheckOK
		check.Message = "no project manifest (nothing to compare)"
		return check
	}
	// Finding A (card t1591): decode into the type manifest.Load decodes into, but
	// read-only. Valid JSON of the wrong shape (for example {"files":123}) is as
	// corrupt as text that is not JSON, and it must never reach Load, which moves
	// a manifest it cannot decode aside.
	var mf manifest.Manifest
	if json.Unmarshal(manifestData, &mf) != nil {
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf("project manifest at %s is corrupt — run 'moai doctor --json' for recovery guidance; the file is preserved untouched", manifestPath)
		if verbose {
			check.Detail = "the doctor reads the manifest read-only: the corrupt bytes stay at their path and any prior .corrupt recovery copy is left alone."
		}
		return check
	}

	notInLock, notOnDisk := 0, 0
	for p := range mf.Files {
		abs := filepath.Join(projectRoot, filepath.FromSlash(p))
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			notOnDisk++
			if verbose {
				check.Detail += fmt.Sprintf("\nlock entry absent from project: %s", p)
			}
		}
	}
	comparable := 0
	_ = filepath.WalkDir(projectRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(projectRoot, p)
		if relErr != nil {
			return nil
		}
		relSlash := filepath.ToSlash(rel)
		if strings.HasPrefix(relSlash, ".git/") || strings.HasPrefix(relSlash, ".moai/state/") ||
			strings.HasPrefix(relSlash, "node_modules/") {
			return nil
		}
		if _, tracked := mf.Files[relSlash]; !tracked {
			// Only template-recorded ROOTS are lock-comparable surface.
			for _, root := range projectCommonAssetRels {
				if strings.HasPrefix(relSlash, root) {
					notInLock++
					if verbose {
						check.Detail += fmt.Sprintf("\nproject file absent from lock: %s", relSlash)
					}
					break
				}
			}
		}
		for _, root := range projectCommonAssetRels {
			if strings.HasPrefix(relSlash, root) {
				comparable++
				break
			}
		}
		return nil
	})

	switch {
	case notInLock > 0 || notOnDisk > 0:
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("lock drift: %d project file(s) absent from lock, %d lock entr(ies) absent from project", notInLock, notOnDisk)
	case comparable == 0:
		// M7 (REQ-DOC-002): the vacuous-comparison honesty — the tree
		// offers no lock-comparable surface, so a "matches the lock file"
		// claim here would have compared nothing.
		check.Status = uikit.CheckWarn
		check.Message = "the project tree offers no lock-comparable surface — the lock comparison is vacuous (nothing was judged)"
	default:
		check.Status = uikit.CheckOK
		check.Message = "project tree matches the lock file"
	}
	return check
}

// checkPluginMigrationAdvisory is the REQ-019 informational row (iter4 D32):
// prior plugin installs need the manual `claude plugin uninstall` step —
// doctor reports it, never executes it.
func checkPluginMigrationAdvisory(verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Plugin Migration"}
	check.Status = uikit.CheckOK
	check.Message = "plugin carrier retired — if a prior release installed the moai plugin, remove it manually: claude plugin uninstall moai"
	if verbose {
		check.Detail = "The moai plugin (SPEC-PLUGIN-MARKETPLACE-001) is retired; common skills and agents install into your user folders now. Doctor never removes the plugin itself."
	}
	return check
}
