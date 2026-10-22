// Package privatefile stores owner-only files using native filesystem access
// controls: mode 0600 on Unix and a protected owner-only DACL on Windows.
package privatefile

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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
	return writeFileWithCheck(path, data, Check)
}

func writeFileWithCheck(path string, data []byte, check func(string) error) (resultErr error) {
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
	if _, err := f.Write(data); err != nil {
		return errors.Join(err, revokePublicationNative(f), f.Close())
	}
	publication, err := CapturePublication(f)
	if err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, publication.Revoke())
		}
		if closeErr := publication.Close(); closeErr != nil {
			if resultErr != nil {
				resultErr = errors.Join(resultErr, closeErr)
			} else {
				slog.Warn("failed to close private publication cleanup handle", "error", closeErr)
			}
		}
	}()
	return publication.install(path, check)
}
