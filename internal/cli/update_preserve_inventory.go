// Package cli — update_preserve_inventory.go
//
// PRESERVE inventory snapshot logic for the v2-to-v3 clean reinstall path
// (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 REQ-VVCR-005, REQ-VVCR-006,
// REQ-VVCR-007, REQ-VVCR-008; AC-VVCR-003, AC-VVCR-004).
//
// The PRESERVE inventory enumerates user-owned project paths that MUST
// survive the clean-reinstall cycle byte-identical. This file builds the
// inventory, snapshots files to a backup directory, and detects user-modified
// configuration files via SHA-256 hash diff against the embedded template
// baseline.
//
// Reused infrastructure:
//   - collectUserOwnedFiles (update_namespace_protect.go:100) — REQ-UNP
//     scan over .claude/skills, .claude/agents, .moai/harness.
//   - copyFile (update_archive.go:331) — content-preserving file copy.
//
// PRESERVE inventory composition per REQ-VVCR-005:
//   - .moai/specs/                                      (SPEC documents)
//   - .moai/project/{product,structure,tech}.md         (project docs)
//   - .claude/skills/harness-*                          (user harness skills — canonical + legacy my-harness-*)
//   - .claude/agents/harness/                           (user harness agents)
//   - .claude/agents/local/                             (maintainer agents)
//   - .claude/commands/  (root files + non-moai subdirs) (user commands)

package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PreserveInventory enumerates project-root-relative paths that MUST be
// preserved across the clean-reinstall cycle. Paths use forward-slash
// separators (NFR-UNP-003 normalization) for cross-platform stability.
//
// @MX:ANCHOR: Output contract consumed by runCleanReinstall MERGE-back step
// (REQ-VVCR-021 / REQ-VVCR-022). Modifications MUST update both this struct
// and the acceptance.md AC-VVCR-003 / AC-VVCR-011 entries.
type PreserveInventory struct {
	// Files is the deduplicated, sorted-friendly list of project-root-relative
	// file paths to preserve. The list never contains directories — only
	// individual file entries — so that MERGE-back restores byte-identical
	// content even when the encompassing directory must be recreated.
	Files []string

	// UserModifiedConfigs is the subset of Files representing template-managed
	// configuration files whose current content differs from the embedded
	// template baseline (REQ-VVCR-007). Reported separately so the
	// orchestrator can surface user-modified configs in the dry-run output.
	UserModifiedConfigs []string
}

// preserveInventoryRoots enumerates the scan roots beyond what
// collectUserOwnedFiles already covers. These roots hold paths that are
// NOT under §24 user-owned namespace rules but ARE user-authored content
// (per REQ-VVCR-005).
var preserveInventoryRoots = []string{
	".moai/specs",
	".moai/project",
	".claude/commands",
}

// buildPreserveInventory enumerates every project-root-relative file path
// that the clean-reinstall MUST preserve. It composes the §24 user-owned
// namespace scan (via collectUserOwnedFiles) with three additional roots
// holding project documentation and user-defined commands.
//
// Returns an inventory with PreserveInventory.Files populated. The
// UserModifiedConfigs slice is populated separately by
// detectUserModifiedConfigs and stitched in by the caller.
//
// The function is read-only and tolerates missing scan roots (empty project
// directories contribute zero entries). Errors propagate only when an
// existing root cannot be walked (e.g., permission denied mid-walk).
func buildPreserveInventory(projectRoot string) (PreserveInventory, error) {
	if projectRoot == "" {
		return PreserveInventory{}, errors.New("buildPreserveInventory: empty projectRoot")
	}

	inv := PreserveInventory{}

	// 1) §24 user-owned namespace (reused verbatim from namespace-protect).
	userOwned, err := collectUserOwnedFiles(projectRoot)
	if err != nil {
		return PreserveInventory{}, fmt.Errorf("collect user-owned files: %w", err)
	}
	inv.Files = append(inv.Files, userOwned...)

	// 2) Additional preserve roots: .moai/specs/, .moai/project/, .claude/commands/.
	for _, root := range preserveInventoryRoots {
		absRoot := filepath.Join(projectRoot, filepath.FromSlash(root))
		if _, statErr := os.Stat(absRoot); statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				continue
			}
			return PreserveInventory{}, fmt.Errorf("stat %s: %w", root, statErr)
		}

		walkErr := filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkInnerErr error) error {
			if walkInnerErr != nil {
				return walkInnerErr
			}
			if d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(projectRoot, path)
			if relErr != nil {
				return relErr
			}
			relNorm := filepath.ToSlash(rel)

			// Filter for .claude/commands/: only preserve root-level files and
			// non-moai/ subdirectories (REQ-VVCR-005 explicit clause). Skip
			// .claude/commands/moai/ entirely — those are template-managed.
			if root == ".claude/commands" {
				rest := strings.TrimPrefix(relNorm, ".claude/commands/")
				seg := strings.SplitN(rest, "/", 2)[0]
				if seg == "moai" {
					return nil
				}
			}

			inv.Files = append(inv.Files, relNorm)
			return nil
		})
		if walkErr != nil {
			return PreserveInventory{}, fmt.Errorf("walk %s: %w", root, walkErr)
		}
	}

	// Deduplicate. A path could appear in both user-owned scan (e.g., a
	// nested harness skill under .claude/skills/) and preserve roots only if
	// the roots overlap — which they currently do not, but the dedup is
	// cheap insurance for future extension.
	inv.Files = dedupSorted(inv.Files)

	return inv, nil
}

// dedupSorted returns a deduplicated copy of the input slice. The output
// preserves insertion order modulo deduplication (first occurrence wins).
// The "Sorted" suffix is aspirational — inventory consumers do not require
// strict sort order, only stable enumeration.
func dedupSorted(in []string) []string {
	if len(in) == 0 {
		return in
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// detectUserModifiedConfigs returns the subset of template-managed
// configuration paths whose current content's SHA-256 hash differs from the
// embedded template baseline (REQ-VVCR-007).
//
// The baseline source is the live filesystem reading of the embedded
// template artifact at the same relative path under
// `internal/template/templates/`. Because the embedded fs is consumed at
// `moai update` runtime via a `template.Embedded` value, callers MUST
// inject the baseline reader through `baselineReader` rather than directly
// import the template package here (avoids circular imports — internal/cli
// already imports internal/template; the template package is the consumer
// of cli for some adjacent helpers).
//
// configPaths is the candidate set of template-managed configs to check.
// Typical input is the project-relative list of `.moai/config/sections/*.yaml`
// files (whose embedded baselines are deterministic).
//
// Implementation note: the function is intentionally generic — it does not
// hard-code the list of configs to inspect. The caller in M5 will pass
// the active project's config-paths slice. Tests in this file's companion
// test file exercise the hash-diff logic with synthetic baseline content.
type BaselineReader func(relPath string) ([]byte, error)

func detectUserModifiedConfigs(projectRoot string, configPaths []string, baseline BaselineReader) ([]string, error) {
	if projectRoot == "" {
		return nil, errors.New("detectUserModifiedConfigs: empty projectRoot")
	}
	if baseline == nil {
		return nil, errors.New("detectUserModifiedConfigs: nil BaselineReader")
	}

	var modified []string
	for _, rel := range configPaths {
		current, readErr := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(rel)))
		if readErr != nil {
			if errors.Is(readErr, os.ErrNotExist) {
				// Missing config — not "modified" in the user sense; skip.
				continue
			}
			return nil, fmt.Errorf("read current %s: %w", rel, readErr)
		}
		bl, blErr := baseline(rel)
		if blErr != nil {
			if errors.Is(blErr, os.ErrNotExist) {
				// Baseline missing — config was added by user, not template-managed.
				continue
			}
			return nil, fmt.Errorf("read baseline %s: %w", rel, blErr)
		}
		if sha256Hex(current) != sha256Hex(bl) {
			modified = append(modified, rel)
		}
	}
	return modified, nil
}

// sha256Hex returns the SHA-256 hex digest of input.
//
// Distinct from hashBytes (design_folder.go:191) which returns the raw
// SHA-256 bytes; this helper returns the hex-encoded string used by
// SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 hash diff (REQ-VVCR-007).
func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// snapshotPreserveInventory copies every file in inv.Files into backupDir
// preserving the project-root-relative directory hierarchy. The backupDir
// is expected to be an absolute path created by the caller via
// resolveNamespaceBackupDir + os.MkdirAll.
//
// On any copy error, the partially-populated backup is left in place — the
// caller MUST emit a clear error and is responsible for whether to delete
// the partial backup. (Unlike backupUserOwnedNamespace which does defensive
// cleanup, here the caller orchestrates the higher-level 7-step canonical
// order and may want to retain a partial backup for inspection.)
//
// A `.complete` marker is written at backupDir root after all copies succeed
// (mirrors backupUserOwnedNamespace REQ-UNP-007 atomicity convention).
func snapshotPreserveInventory(projectRoot string, inv PreserveInventory, backupDir string) error {
	if backupDir == "" {
		return errors.New("snapshotPreserveInventory: empty backupDir")
	}

	for _, rel := range inv.Files {
		srcPath := filepath.Join(projectRoot, filepath.FromSlash(rel))
		dstPath := filepath.Join(backupDir, filepath.FromSlash(rel))

		// Ensure parent directory exists.
		if mkErr := os.MkdirAll(filepath.Dir(dstPath), 0o755); mkErr != nil {
			return fmt.Errorf("create backup parent for %s: %w", rel, mkErr)
		}

		if copyErr := copyFile(srcPath, dstPath); copyErr != nil {
			return fmt.Errorf("copy %s → backup: %w", rel, copyErr)
		}
	}

	// Atomicity marker (REQ-VVCR-003 echo).
	markerPath := filepath.Join(backupDir, ".complete")
	markerContent := fmt.Sprintf("inventory=%d\n", len(inv.Files))
	if markerErr := os.WriteFile(markerPath, []byte(markerContent), 0o644); markerErr != nil {
		return fmt.Errorf("write .complete marker: %w", markerErr)
	}

	return nil
}

// mergeBackPreserveInventory restores files from backupDir back into
// projectRoot at their original project-root-relative paths. Used by Step 6
// of the canonical clean-reinstall order (REQ-VVCR-021, REQ-VVCR-022).
//
// Existing destination files are overwritten — the reinstall step (Step 5)
// is expected to not have written anything to a PRESERVE path (because
// PRESERVE paths are by definition outside the template-managed surface),
// but we overwrite unconditionally to make the restore idempotent.
func mergeBackPreserveInventory(projectRoot string, inv PreserveInventory, backupDir string) error {
	if backupDir == "" {
		return errors.New("mergeBackPreserveInventory: empty backupDir")
	}

	for _, rel := range inv.Files {
		srcPath := filepath.Join(backupDir, filepath.FromSlash(rel))
		dstPath := filepath.Join(projectRoot, filepath.FromSlash(rel))

		// Skip files missing from the backup (e.g., were not actually
		// present at snapshot time but listed in inv.Files due to a race —
		// shouldn't happen but be defensive).
		if _, statErr := os.Stat(srcPath); statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("stat backup %s: %w", rel, statErr)
		}

		if mkErr := os.MkdirAll(filepath.Dir(dstPath), 0o755); mkErr != nil {
			return fmt.Errorf("create restore parent for %s: %w", rel, mkErr)
		}

		if copyErr := copyFile(srcPath, dstPath); copyErr != nil {
			return fmt.Errorf("restore %s ← backup: %w", rel, copyErr)
		}
	}

	return nil
}

// computeInventoryHashes returns a SHA-256 map keyed by project-root-relative
// path for each file in inv.Files. Used by integrity verification in Step 7
// (REQ-VVCR-023) to assert that PRESERVE files are byte-identical before/after.
func computeInventoryHashes(projectRoot string, inv PreserveInventory) (map[string]string, error) {
	hashes := make(map[string]string, len(inv.Files))
	for _, rel := range inv.Files {
		abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
		f, openErr := os.Open(abs)
		if openErr != nil {
			if errors.Is(openErr, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("open %s: %w", rel, openErr)
		}
		h := sha256.New()
		if _, copyErr := io.Copy(h, f); copyErr != nil {
			_ = f.Close()
			return nil, fmt.Errorf("hash %s: %w", rel, copyErr)
		}
		_ = f.Close()
		hashes[rel] = hex.EncodeToString(h.Sum(nil))
	}
	return hashes, nil
}
