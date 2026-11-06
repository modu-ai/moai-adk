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
	// it), "side-copy-blocked" (a different side copy already exists, so nothing
	// was written), "side-copy-refused" (the side-copy name is a symbolic link or
	// sits under one, or is not a regular file, so nothing was written or read), or
	// "target-refused" (the target is a symbolic link or another non-regular file, or
	// its parent chain leads outside the config directory or cannot be resolved, so
	// nothing was written or read).
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
	return RestoreUnmanagedConfigDirsSkipping(projectRoot, backupDir, nil)
}

// RestoreUnmanagedConfigDirsSkipping is RestoreUnmanagedConfigDirs with a skip
// filter. skip receives the project-relative slash path of each backup file and
// reports true for a file this run already archived, removed, or preserved; such a
// file is not written back, so the restore cannot undo the reconciliation report
// (the same exclusion the section restore applies). A nil skip restores everything.
func RestoreUnmanagedConfigDirsSkipping(projectRoot, backupDir string, skip func(refPath string) bool) ([]UnmanagedRestoreRef, error) {
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
			rootRel, err := filepath.Rel(projectRoot, target)
			if err != nil {
				return err
			}
			refPath := filepath.ToSlash(rootRel)
			if skip != nil && skip(refPath) {
				return nil
			}
			// A refused target is recorded, not skipped: the report must name the
			// file the restore did not write.
			if !RestoreTargetContained(configDir, target) {
				refs = append(refs, UnmanagedRestoreRef{Path: refPath, Action: "target-refused"})
				return nil
			}
			backupData, err := os.ReadFile(src)
			if err != nil {
				return err
			}
			ref, err := restoreUnmanagedFile(configDir, target, backupData, stamp)
			if err != nil || ref == nil {
				return err
			}
			ref.Path = refPath
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
func restoreUnmanagedFile(configDir, target string, backupData []byte, stamp string) (*UnmanagedRestoreRef, error) {
	// A non-regular target (FIFO, device, socket) is never read: reading a FIFO blocks
	// until a writer appears. It is refused and reported, as a symbolic link is.
	if info, lerr := os.Lstat(target); lerr == nil && !info.Mode().IsRegular() {
		return &UnmanagedRestoreRef{Action: "target-refused"}, nil
	}
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
	// The side copy is a write target like the working file: a symbolic link at its
	// name, or a parent chain that leads out of configDir, refuses the write.
	if !RestoreTargetContained(configDir, side) {
		return &UnmanagedRestoreRef{Action: "side-copy-refused"}, nil
	}
	// A non-regular side copy is never read, for the same blocking reason as the target.
	if info, lerr := os.Lstat(side); lerr == nil && !info.Mode().IsRegular() {
		return &UnmanagedRestoreRef{Action: "side-copy-refused"}, nil
	}
	existing, err := os.ReadFile(side)
	switch {
	case err == nil && bytes.Equal(existing, backupData):
		return nil, nil
	case err == nil:
		return &UnmanagedRestoreRef{Action: "side-copy-blocked"}, nil
	case !os.IsNotExist(err):
		return nil, err
	}
	if wErr := writeNewFile(side, backupData); wErr != nil {
		if os.IsExist(wErr) {
			return &UnmanagedRestoreRef{Action: "side-copy-blocked"}, nil
		}
		return nil, wErr
	}
	return &UnmanagedRestoreRef{Action: "side-copy"}, nil
}

// writeNewFile creates path and writes data into it. O_EXCL fails the create on any
// existing name, a dangling symbolic link included, so the write never follows a link.
func writeNewFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defs.FilePerm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
