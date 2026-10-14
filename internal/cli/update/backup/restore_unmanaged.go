package backup

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// unmanagedConfigDirs are the config directories a backup carries in full but
// that the 3-way section merge never touches.
var unmanagedConfigDirs = []string{"evaluator-profiles", "astgrep-rules"}

// UnmanagedRestoreRef records one file the unmanaged-directory restore acted on.
type UnmanagedRestoreRef struct {
	// Path is the project-relative, slash-separated path of the working file.
	Path string
	// Action is "restored" (the file was missing and was written from the backup),
	// "side-copy" (the working file was kept and the backup copy was written beside
	// it), or "side-copy-blocked" (a different side copy already exists, so nothing
	// was written).
	Action string
}

// RestoreUnmanagedConfigDirs restores the backup's evaluator-profiles/ and
// astgrep-rules/ without overwriting what the new template wrote. A backup keeps
// no per-file template base for these directories, so a difference cannot be
// split into "user edit" and "older template". The new template therefore stays
// in place, and the backup copy is written beside it as <name>.from-<backup
// stamp>. A missing file is restored outright, identical content is a no-op, and
// a repeated run changes nothing.
func RestoreUnmanagedConfigDirs(projectRoot, backupDir string) ([]UnmanagedRestoreRef, error) {
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	stamp := filepath.Base(filepath.Clean(backupDir))
	var refs []UnmanagedRestoreRef
	for _, dir := range unmanagedConfigDirs {
		srcRoot := filepath.Join(backupDir, dir)
		if info, err := os.Stat(srcRoot); err != nil || !info.IsDir() {
			continue
		}
		walkErr := filepath.WalkDir(srcRoot, func(src string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Symlinks and other non-regular entries are skipped: following a link
			// out of the backup is the containment hazard the section restore guards.
			if d.IsDir() || !d.Type().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(srcRoot, src)
			if err != nil {
				return err
			}
			target := filepath.Join(configDir, dir, rel)
			if !RestoreTargetContained(configDir, target) {
				return nil
			}
			backupData, err := os.ReadFile(src)
			if err != nil {
				return err
			}
			ref, err := restoreUnmanagedFile(target, backupData, stamp)
			if err != nil || ref == nil {
				return err
			}
			rootRel, err := filepath.Rel(projectRoot, target)
			if err != nil {
				return err
			}
			ref.Path = filepath.ToSlash(rootRel)
			refs = append(refs, *ref)
			return nil
		})
		if walkErr != nil {
			return refs, walkErr
		}
	}
	return refs, nil
}

// restoreUnmanagedFile applies the restore rule to one target file. It returns
// nil when there is nothing to report.
func restoreUnmanagedFile(target string, backupData []byte, stamp string) (*UnmanagedRestoreRef, error) {
	working, err := os.ReadFile(target)
	switch {
	case os.IsNotExist(err):
		if mkErr := os.MkdirAll(filepath.Dir(target), defs.DirPerm); mkErr != nil {
			return nil, mkErr
		}
		if wErr := os.WriteFile(target, backupData, defs.FilePerm); wErr != nil {
			return nil, wErr
		}
		return &UnmanagedRestoreRef{Action: "restored"}, nil
	case err != nil:
		return nil, err
	case bytes.Equal(working, backupData):
		return nil, nil
	}

	side := target + ".from-" + stamp
	existing, err := os.ReadFile(side)
	switch {
	case err == nil && bytes.Equal(existing, backupData):
		return nil, nil
	case err == nil:
		return &UnmanagedRestoreRef{Action: "side-copy-blocked"}, nil
	case !os.IsNotExist(err):
		return nil, err
	}
	if wErr := os.WriteFile(side, backupData, defs.FilePerm); wErr != nil {
		return nil, wErr
	}
	return &UnmanagedRestoreRef{Action: "side-copy"}, nil
}
