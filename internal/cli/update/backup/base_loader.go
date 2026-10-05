package backup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// UnattestedBaseMarker is the file SaveTemplateBase leaves in the BASE
// directory when it replaced an unattested snapshot with the embedded
// defaults. RestoreMoaiConfigRetained reads it to report the values the merge
// kept over the current template (card t1216).
const UnattestedBaseMarker = "unattested-snapshot"

// SaveTemplateBase populates destDir/sections/ with the bytes that RestoreMoaiConfig
// will later read as the 3-way merge BASE. It prefers the rendered snapshot
// (Decision D2, REQ-TBS-006) and falls back to SaveTemplateDefaults
// (embedded-raw) when the snapshot is absent (REQ-TBS-007).
//
// Provenance fix (the core of SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001): the
// snapshot carries RENDERED values (version: "3.0.1"), so the merge correctly
// distinguishes "user customized" (old != base) from "template changed"
// (old == base). The embedded-raw fallback carries {{.Version}} placeholders,
// which is today's wrong-base behaviour — but only for the first post-feature
// update cycle (Decision D6 migration cost).
//
// The function is NON-breaking: SaveTemplateDefaults and its existing test
// callers stay intact; this is the new entry point BackupMoaiConfig uses.
func SaveTemplateBase(destDir, projectRoot string) error {
	// Two updates within one second share a backup directory (the timestamp
	// has second granularity), so a marker an earlier run left must not
	// outlive the BASE it described.
	if err := os.Remove(filepath.Join(destDir, UnattestedBaseMarker)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("save template base: clear marker: %w", err)
	}
	if !HasSnapshot(projectRoot) {
		// Fallback: today's embedded-raw BASE. Identical to the pre-SPEC path,
		// so existing SaveTemplateDefaults tests double as fallback tests.
		return SaveTemplateDefaults(destDir)
	}
	if snapshotUnattested(projectRoot) {
		// Card t1216: an unattested snapshot takes the same fallback. It may
		// hold the user's own values (written after a restore), and a BASE
		// equal to a user value reads that value as unchanged and drops it;
		// the embedded BASE can only err toward keeping a value. The marker
		// tells the restore to list the kept values once.
		if err := SaveTemplateDefaults(destDir); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(destDir, UnattestedBaseMarker), nil, defs.FilePerm)
	}

	srcSections := filepath.Join(SnapshotDir(projectRoot), "sections")
	dstSections := filepath.Join(destDir, "sections")
	if err := os.MkdirAll(dstSections, defs.DirPerm); err != nil {
		return fmt.Errorf("save template base: create sections dir: %w", err)
	}

	walkErr := filepath.WalkDir(srcSections, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath, relErr := filepath.Rel(srcSections, path)
		if relErr != nil {
			return relErr
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read snapshot %s: %w", relPath, readErr)
		}
		dst := filepath.Join(dstSections, relPath)
		if mkErr := os.MkdirAll(filepath.Dir(dst), defs.DirPerm); mkErr != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), mkErr)
		}
		if writeErr := os.WriteFile(dst, data, defs.FilePerm); writeErr != nil {
			return fmt.Errorf("write %s: %w", relPath, writeErr)
		}
		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("save template base: walk snapshot: %w", walkErr)
	}
	return nil
}
