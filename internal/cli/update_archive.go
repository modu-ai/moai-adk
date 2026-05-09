// SPEC-V3R3-HARNESS-001 / T-M4-01
// update_archive.go — provides archiveSkill, which recursively copies
// .claude/skills/<id>/ into .moai/archive/skills/v2.16/<id>/.
//
// Idempotency guarantees:
//   - source absent → return nil immediately
//   - archive already exists + content matches → return nil
//   - archive already exists + content differs (drift) → ARCHIVE_DRIFT error
//
// Path traversal protection:
//   - return error if skillID contains ".." or "/"

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// archiveVersion is the version tag used for the archive directory.
const archiveVersion = "v2.16"

// legacySkillIDs lists the 16 skill IDs removed in BC-V3R3-007.
// When `moai update` runs, these skills are moved to .moai/archive/skills/v2.16/.
var legacySkillIDs = []string{
	"moai-domain-backend",
	"moai-domain-frontend",
	"moai-domain-database",
	"moai-domain-db-docs",
	"moai-domain-mobile",
	"moai-framework-electron",
	"moai-library-shadcn",
	"moai-library-mermaid",
	"moai-library-nextra",
	"moai-tool-ast-grep",
	"moai-platform-auth",
	"moai-platform-deployment",
	"moai-platform-chrome-extension",
	"moai-workflow-research",
	"moai-workflow-pencil-integration",
	"moai-formats-data",
}

// archiveSkill copies projectRoot/.claude/skills/<skillID>/
// into projectRoot/.moai/archive/skills/v2.16/<skillID>/.
//
// @MX:ANCHOR: [AUTO] archiveSkill is the entry point for the legacy skill archival contract
// @MX:REASON: [AUTO] called from runUpdate flow, restore-skill, idempotency tests; fan_in >= 3
func archiveSkill(projectRoot, skillID string) error {
	// Guard against path traversal: skillID must not contain ".." or "/" (separator)
	if err := validateSkillID(skillID); err != nil {
		return err
	}

	srcDir := filepath.Join(projectRoot, ".claude", "skills", skillID)

	// If the source is missing, return nil idempotently (handles already-removed case)
	if _, err := os.Stat(srcDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat source skill %s: %w", skillID, err)
	}

	dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, skillID)

	// If the archive already exists, run drift check
	if _, err := os.Stat(dstDir); err == nil {
		if err := checkArchiveDrift(srcDir, dstDir); err != nil {
			return err
		}
		// content matches → idempotent success
		return nil
	}

	// Create the archive destination directory
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("create archive directory for %s: %w", skillID, err)
	}

	// Recursive copy
	if err := copyDirAll(srcDir, dstDir); err != nil {
		// On failure, clean up the partial archive
		_ = os.RemoveAll(dstDir)
		return fmt.Errorf("copy skill %s to archive: %w", skillID, err)
	}

	return nil
}

// validateSkillID checks whether skillID contains any path-traversal characters
// ("..", "/", or absolute paths).
func validateSkillID(skillID string) error {
	// Reject absolute paths
	if filepath.IsAbs(skillID) {
		return &MigrateError{
			Code:    "ARCHIVE_INVALID_ID",
			Message: fmt.Sprintf("skillID must be a simple name, not an absolute path: %q", skillID),
		}
	}
	// Reject ".."
	if strings.Contains(skillID, "..") {
		return &MigrateError{
			Code:    "ARCHIVE_INVALID_ID",
			Message: fmt.Sprintf("skillID must not contain '..': %q", skillID),
		}
	}
	// Reject path separators ("/" or platform-specific)
	if strings.ContainsAny(skillID, "/\\") {
		return &MigrateError{
			Code:    "ARCHIVE_INVALID_ID",
			Message: fmt.Sprintf("skillID must not contain path separators: %q", skillID),
		}
	}
	return nil
}

// checkArchiveDrift compares the contents of the source directory and
// an existing archive directory using SHA-256 hashes.
// Returns nil when contents match, ARCHIVE_DRIFT error otherwise.
func checkArchiveDrift(srcDir, dstDir string) error {
	srcHashes, err := computeDirHashes(srcDir)
	if err != nil {
		return fmt.Errorf("compute source hashes: %w", err)
	}
	dstHashes, err := computeDirHashes(dstDir)
	if err != nil {
		return fmt.Errorf("compute archive hashes: %w", err)
	}

	if len(srcHashes) != len(dstHashes) {
		return &MigrateError{
			Code: "ARCHIVE_DRIFT",
			Message: fmt.Sprintf(
				"archive already exists but file count differs (src=%d, dst=%d). "+
					"Use --force to overwrite.",
				len(srcHashes), len(dstHashes),
			),
		}
	}

	for rel, srcHash := range srcHashes {
		dstHash, ok := dstHashes[rel]
		if !ok || srcHash != dstHash {
			return &MigrateError{
				Code: "ARCHIVE_DRIFT",
				Message: fmt.Sprintf(
					"archive already exists but content differs for %s. "+
						"Use --force to overwrite.",
					rel,
				),
			}
		}
	}

	return nil
}

// computeDirHashes returns SHA-256 hashes for every file in dir as a
// path → hash map. Reuses hashFile (declared in design_folder.go).
func computeDirHashes(dir string) (map[string]string, error) {
	hashes := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		// hashFile is declared in design_folder.go and returns []byte
		rawHash, err := hashFile(path)
		if err != nil {
			return err
		}
		hashes[rel] = fmt.Sprintf("%x", rawHash)
		return nil
	})
	return hashes, err
}

// copyDirAll recursively copies every entry in srcDir into dstDir,
// preserving each file's Unix permission bits.
func copyDirAll(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, rel)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(dstPath, info.Mode().Perm())
		}

		return copyFile(path, dstPath)
	})
}

// archiveLegacySkills walks legacySkillIDs from projectRoot and calls
// archiveSkill for each. Returns the number of newly archived skills.
//
// Idempotency: if the archive already exists and content matches, the skill
// is not counted.
//
// Output format:
//
//	archive: <id> → .moai/archive/skills/v2.16/<id>
//	total: N skills archived, 0 user customizations modified
//
// @MX:ANCHOR: [AUTO] archiveLegacySkills is the entry point for legacy skill archival in the update flow
// @MX:REASON: [AUTO] called from runUpdate, dry-run, and idempotency tests; fan_in >= 3
// @MX:NOTE: [AUTO] M4-S4d-3 DDD migration — archive progress uses tui.CheckLine "ok";
// summary uses tui.Pill PillOk; dry-run uses PillInfo + CheckLine "info"; the
// "archive:" / "total:" keywords are preserved (test contains assertion).
func archiveLegacySkills(projectRoot string, out io.Writer) (int, error) {
	th := resolveTheme()
	archived := 0
	for _, id := range legacySkillIDs {
		srcDir := filepath.Join(projectRoot, ".claude", "skills", id)
		if _, err := os.Stat(srcDir); err != nil {
			// source missing → skip
			continue
		}

		// Check whether the archive already exists first (idempotency check)
		dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, id)
		alreadyArchived := false
		if _, err := os.Stat(dstDir); err == nil {
			alreadyArchived = true
		}

		if err := archiveSkill(projectRoot, id); err != nil {
			return archived, fmt.Errorf("archive %s: %w", id, err)
		}

		// Do not count if already existed (idempotent run)
		if alreadyArchived {
			continue
		}

		archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
		_, _ = fmt.Fprintln(out, tui.CheckLine("ok", "archive: "+id, "→ "+archiveDst, "", &th))
		archived++
	}

	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
	return archived, nil
}

// dryRunArchiveLegacySkills runs in --dry-run mode and prints the planned
// work without making any filesystem changes.
func dryRunArchiveLegacySkills(projectRoot string, out io.Writer) error {
	th := resolveTheme()
	planned := 0
	for _, id := range legacySkillIDs {
		srcDir := filepath.Join(projectRoot, ".claude", "skills", id)
		if _, err := os.Stat(srcDir); err != nil {
			continue
		}
		archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "[dry-run] archive: "+id, "→ "+archiveDst, "", &th))
		planned++
	}
	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Solid: false, Label: fmt.Sprintf("[dry-run] total: %d skills archived, 0 user customizations modified", planned), Theme: &th}))
	return nil
}

// copyFile copies a single file from src to dst, preserving the source's
// permission bits.
func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s → %s: %w", src, dst, err)
	}

	return nil
}
