//go:build windows

// Package cli — update_cleanup_windows.go
//
// Windows-specific process-liveness probe for SPEC-V3R3-UPDATE-CLEANUP-001
// stale-lock detection.
//
// SPEC-V2.20.0RC1 hotfix extracted this file to fix `syscall.Kill undefined`
// on the windows/amd64 target; the placeholder here reported every PID as
// alive, which the mcp_server_runtime liveness split and the doctor's
// dead-record pruning then inherited (t426 windows census axis 2a). The probe
// now delegates to the shared session implementation (OpenProcess +
// GetExitCodeProcess).

package cli

import (
	"github.com/modu-ai/moai-adk/internal/session"
)

// isProcessAlive delegates to the shared per-platform probe.
func isProcessAlive(pid int) bool {
	return session.IsProcessAlive(pid)
}
