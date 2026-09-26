package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// SnapshotSubdir is the snapshot directory relative to defs.MoAIDir. The
// snapshot is a verbatim byte-copy of the on-disk rendered
// .moai/config/sections/ tree, written right after each install/update
// template deploy (before user values are written over it) so the
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
// @MX:ANCHOR: [AUTO] snapshot write entry — fan_in 3 (init + 2 update deploy sites)
// @MX:REASON: every trigger site funnels through this single write, and each
// MUST call it right after a template deploy and before any user value is
// written over the section files (card t1139); a behavior change here
// propagates to init and both update paths at once. SaveTemplateBase reads the
// result as the next merge BASE.
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
	// Card t1216: attest the tree just written, so SaveTemplateBase can tell
	// this deploy-time snapshot from one written after a restore.
	return AttestSnapshot(projectRoot)
}

// AttestSnapshot records the digest of the snapshot's current sections/ tree,
// declaring it a deploy-time render that SaveTemplateBase may use as BASE.
// WriteSnapshot calls it; nothing that writes the snapshot after a restore may.
func AttestSnapshot(projectRoot string) error {
	sum, err := snapshotSectionsDigest(projectRoot)
	if err != nil {
		return fmt.Errorf("snapshot: digest sections: %w", err)
	}
	if err := os.WriteFile(snapshotAttestPath(projectRoot), []byte(sum+"\n"), defs.FilePerm); err != nil {
		return fmt.Errorf("snapshot: write attestation: %w", err)
	}
	return nil
}

// snapshotAttestPath is the file WriteSnapshot records the sections digest in.
// It sits beside sections/, so neither the BASE copy nor the restore walk sees it.
func snapshotAttestPath(projectRoot string) string {
	return filepath.Join(SnapshotDir(projectRoot), "sections.sha256")
}

// snapshotSectionsDigest hashes every file under the snapshot's sections/
// tree: relative path, length, and bytes, in lexical walk order.
func snapshotSectionsDigest(projectRoot string) (string, error) {
	root := filepath.Join(SnapshotDir(projectRoot), "sections")
	h := sha256.New()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		_, _ = fmt.Fprintf(h, "%s\x00%d\x00", filepath.ToSlash(rel), len(data))
		_, _ = h.Write(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// snapshotUnattested reports whether the snapshot's sections/ tree is NOT the
// one the last WriteSnapshot call attested: no attestation (left by a binary
// that predates it) or a tree rewritten since (an older binary run after a
// newer one). A tree whose digest cannot be computed counts as unattested too:
// copying it anyway would stop at the unreadable file and leave a partial
// BASE, which is the loss this gate exists to prevent (card t1216 sync-audit
// F1).
//
// @MX:NOTE: [AUTO] trust gate for the merge BASE (card t1216) — binaries before card t1139 wrote the snapshot AFTER the restore, so it holds user values; a BASE equal to a user value makes the 3-way merge replace that value with the template default
func snapshotUnattested(projectRoot string) bool {
	want, err := os.ReadFile(snapshotAttestPath(projectRoot))
	if err != nil {
		return true
	}
	got, err := snapshotSectionsDigest(projectRoot)
	return err != nil || strings.TrimSpace(string(want)) != got
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
