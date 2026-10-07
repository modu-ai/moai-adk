// doctor_user_install.go — the `moai doctor` user-install integrity check
// (SPEC-USER-ASSET-INSTALL-001 M5; REQ-014) and the project-vs-lock
// comparison (REQ-015). Read-only, report-only, doctor house style.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/manifest"
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

	roots := userassets.ResolveRoots(homeDir)
	rootDirs := make(map[userassets.RootSlug]string, len(roots))
	for _, r := range roots {
		rootDirs[r.Slug] = r.Dir
	}

	missing, modified := 0, 0
	for key, fe := range m.Files {
		slug, rel, ok := splitManifestKeyUser(key)
		if !ok {
			continue
		}
		dir, known := rootDirs[slug]
		if !known {
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
		}
	}

	untracked := countUntrackedUserFiles(roots, m)
	switch {
	case missing > 0 || modified > 0:
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("%d manifest-tracked file(s) missing, %d modified — 'moai update' repairs manifest-hash matches; edits stay preserved", missing, modified)
	case untracked > 0:
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("%d file(s) verified; %d untracked file(s) in the user roots (informational, never removed)", len(m.Files)-missing, untracked)
	default:
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("%d file(s) verified", len(m.Files))
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

// checkProjectVsLock compares the project tree against the project lock file
// in BOTH directions (REQ-015): a project file absent from the lock, and a
// lock entry absent from the project. Read-only; the lock schema is not
// touched (C3).
func checkProjectVsLock(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Project Lock"}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		check.Status = uikit.CheckOK
		check.Message = "no project manifest (nothing to compare)"
		return check
	}

	notInLock, notOnDisk := 0, 0
	for p := range mgr.Manifest().Files {
		abs := filepath.Join(projectRoot, filepath.FromSlash(p))
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			notOnDisk++
			if verbose {
				check.Detail += fmt.Sprintf("\nlock entry absent from project: %s", p)
			}
		}
	}
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
		if _, tracked := mgr.Manifest().Files[relSlash]; !tracked {
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
		return nil
	})

	switch {
	case notInLock > 0 || notOnDisk > 0:
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("lock drift: %d project file(s) absent from lock, %d lock entr(ies) absent from project", notInLock, notOnDisk)
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
