// Package atomicfile provides a shared atomic, mode-preserving file-write
// helper. It is the single choke point for writes into .moai/config/** and CLI
// persistence paths so that every future write is atomic (crash leaves either
// the complete new or the complete prior content) and preserves the
// destination's pre-existing permission bits.
//
// The helper does NOT hardcode a numeric mode literal: existing-file modes are
// read via os.Stat, new-file modes come from the caller-supplied defaultMode
// parameter. The canonical default for regular config files is defs.FilePerm;
// callers that own a secret file pass 0o600 explicitly.
package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// tempPattern is the glob prefix the helper uses for its in-directory temp
// files. It is exported so tests and the regression guard can detect orphans.
const tempPattern = ".moai-config-*.tmp"

// Write writes data to path atomically using a temp file in the destination's
// directory followed by os.Rename, then applies the correct permission bits
// before the rename so the result file never inherits os.CreateTemp's 0600
// default.
//
// Mode handling:
//   - When path already exists, the result file keeps the destination's
//     pre-existing mode (read via os.Stat), so the write neither narrows nor
//     widens the mode.
//   - When path does not exist, the result file is created at defaultMode (the
//     caller-supplied default; use defs.FilePerm for regular config, 0o600 for
//     caller-declared secret files).
//
// On any error before the rename completes, the temp file is removed so no
// orphan survives the failed call.
//
// @MX:ANCHOR: [AUTO] single atomic, mode-preserving write choke point for config/CLI persistence
// @MX:REASON: REQ-CAW-001..006 — every config/settings writer in internal/config + internal/cli routes through this; reversing its signature forces every call site to change.
func Write(path string, data []byte, defaultMode os.FileMode) error {
	dir := filepath.Dir(path)

	// Resolve the mode the result file must carry. Existing destinations keep
	// their pre-write mode; new destinations take the caller's default.
	mode := defaultMode
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat destination %s: %w", path, err)
	}

	tmp, err := os.CreateTemp(dir, tempPattern)
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	// On every error return below, remove the orphan temp so AC-CAW-006 holds.
	// On the success path os.Rename renames it away, making the remove a no-op.
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Apply the resolved mode before the rename so the live file never briefly
	// carries CreateTemp's 0600. os.Chmod is a no-op (and does not error) on
	// Windows, so this is safe cross-platform.
	if err := os.Chmod(tmpName, mode); err != nil {
		return fmt.Errorf("chmod temp file to %v: %w", mode, err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp to %s: %w", path, err)
	}
	return nil
}
