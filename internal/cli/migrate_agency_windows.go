//go:build windows

// migrate_agency_windows.go: Windows no-op permission handling for agency migration.
// @SPEC:SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-012b
package cli

import (
	"io"
	"os"
	"sync"
)

var windowsPermNoticeOnce sync.Once

// applyPermissions is a no-op on Windows. Unix permission bits do not apply.
// Prints a one-time notice to stderr per REQ-MIGRATE-012b.
func applyPermissions(_ string, _ string, _ os.FileInfo, stderr io.Writer) {
	windowsPermNoticeOnce.Do(func() {
		if stderr != nil {
			const msg = "Windows: Unix permission bits not applicable, ACL preserved as-is\n"
			_, _ = io.WriteString(stderr, msg)
		}
	})
}
