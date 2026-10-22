// Package privatefile stores owner-only files using native filesystem access
// controls: mode 0600 on Unix and a protected owner-only DACL on Windows.
package privatefile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// Check refuses links, nonregular files and any access policy other than the
// native owner-only policy. An unreadable or unknown policy fails closed.
// @MX:ANCHOR: [AUTO] Verify the native owner-only file policy at every receipt and persistence boundary.
// @MX:REASON: Mission receipts, mission state and private replacement share this fail-closed check.
func Check(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("private file: not a regular file")
	}
	return checkNative(path, info)
}

// WriteFile replaces path with an owner-only file. The temporary file is
// private before any data is written. Existing symlinks are never followed.
// @MX:ANCHOR: [AUTO] Replace private payloads only after their native access policy is installed.
// @MX:REASON: Credential stores, integration leases and config backups share this before-bytes privacy boundary.
// @MX:WARN: [AUTO] Keep each private replacement refusal and cleanup path fail-closed.
// @MX:REASON: Creation, writing, closing, policy readback and replacement failures must stop the payload update.
func WriteFile(path string, data []byte) error {
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("private file: unsafe destination")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := CreateTemp(filepath.Dir(path), ".private-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := Check(name); err != nil {
		return err
	}
	if err := atomicfile.Replace(name, path); err != nil {
		return err
	}
	return Check(path)
}
