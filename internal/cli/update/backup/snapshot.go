package backup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// SnapshotSubdir is the snapshot directory relative to defs.MoAIDir. The
// snapshot is a verbatim byte-copy of the on-disk rendered
// .moai/config/sections/ tree, written at install/update completion so the
// 3-way merge has a RENDERED base (Decision D3, REQ-TBS-003) rather than the
// raw embedded template carrying {{.Version}} placeholders.
//
// @MX:ANCHOR: [AUTO] snapshot location constant — pinned by Decision D1
// @MX:REASON: the snapshot MUST live under .moai/cache/ (gitignored, survives
// the update clean step per research.md §A); moving it re-couples the clean
// step to this SPEC, which Decision D5 explicitly forbids.
const SnapshotSubdir = "cache/template-snapshot"

// SnapshotDir returns the absolute snapshot directory for a project:
// <projectRoot>/.moai/cache/template-snapshot/.
func SnapshotDir(projectRoot string) string {
	return filepath.Join(projectRoot, defs.MoAIDir, SnapshotSubdir)
}

// WriteSnapshot copies every .yaml/.yml file under
// <projectRoot>/.moai/config/sections/ verbatim into
// <projectRoot>/.moai/cache/template-snapshot/sections/<relpath>.
//
// The copy is a byte-for-byte file copy (NOT a re-render), so the snapshot
// carries resolved values (version: "3.0.1") rather than Go-template
// placeholders ({{.Version}}) — REQ-TBS-003, Decision D3.
//
// Scope is .moai/config/sections/ ONLY (REQ-TBS-015); the walk root never
// broadens to .claude/, .moai/project/, or any other template root.
//
// Best-effort non-blocking (REQ-TBS-014): a missing/unreadable config dir
// returns a non-nil error so the caller can log it, but individual copy
// failures are swallowed with a stderr warning so a single bad section file
// does not abort the enclosing init/update. The caller MUST NOT propagate a
// snapshot-write failure into the init/update result.
//
// @MX:ANCHOR: [AUTO] snapshot write entry — fan_in 4 (init + 3 restore sites)
// @MX:REASON: all four Decision D4 trigger sites funnel through this single
// write; a behavior change here propagates to init + both update paths + the
// lockout-escape restore simultaneously. SaveTemplateBase reads the result.
func WriteSnapshot(projectRoot string) error {
	srcDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	info, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("snapshot: stat config sections dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("snapshot: %s is not a directory", srcDir)
	}

	dstDir := filepath.Join(SnapshotDir(projectRoot), "sections")
	if err := os.MkdirAll(dstDir, defs.DirPerm); err != nil {
		return fmt.Errorf("snapshot: create snapshot dir: %w", err)
	}

	walkErr := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Scope guard: only .yaml/.yml files are snapshotted (matches the
		// restore merge surface, which only merges those extensions).
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		relPath, relErr := filepath.Rel(srcDir, path)
		if relErr != nil {
			return relErr
		}

		dst := filepath.Join(dstDir, relPath)
		if mkErr := os.MkdirAll(filepath.Dir(dst), defs.DirPerm); mkErr != nil {
			// Individual copy failure: warn + swallow (REQ-TBS-014).
			_, _ = fmt.Fprintf(os.Stderr, "Warning: snapshot mkdir %s: %v\n", filepath.Dir(dst), mkErr)
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: snapshot read %s: %v\n", relPath, readErr)
			return nil
		}
		if writeErr := os.WriteFile(dst, data, defs.FilePerm); writeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: snapshot write %s: %v\n", relPath, writeErr)
			return nil
		}
		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("snapshot: walk config sections: %w", walkErr)
	}
	return nil
}

// HasSnapshot reports whether a usable snapshot exists: true iff
// .moai/cache/template-snapshot/sections/ exists AND contains at least one
// file. An empty (but present) sections dir reports false so the fallback
// path (SaveTemplateDefaults embedded-raw) is taken on a degenerate snapshot.
func HasSnapshot(projectRoot string) bool {
	sectionsDir := filepath.Join(SnapshotDir(projectRoot), "sections")
	info, err := os.Stat(sectionsDir)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(sectionsDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		// A directory entry counts (nested sections); a file entry counts.
		// Any non-empty presence satisfies the predicate.
		_ = e
		return true
	}
	return false
}
