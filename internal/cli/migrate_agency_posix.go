//go:build !windows

// migrate_agency_posix.go: POSIX-specific permission preservation for agency migration.
// @SPEC:SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-012a
package cli

import (
	"io"
	"os"
)

// applyPermissions preserves Unix permission bits on POSIX systems.
// On error, logs a warning but does not fail the migration.
func applyPermissions(src, dst string, srcInfo os.FileInfo, stderr io.Writer) {
	perm := srcInfo.Mode().Perm() & 0o7777
	if err := os.Chmod(dst, perm); err != nil {
		if stderr != nil {
			_, _ = io.WriteString(stderr, "warn: could not preserve permissions for "+dst+": "+err.Error()+"\n")
		}
	}
}
