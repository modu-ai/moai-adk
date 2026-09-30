//go:build windows

// sweep_cwd_windows.go — the sweep's process-cwd probe on Windows:
// preservation by design, not feature parity (REQ-WS-006, spec.md §E).
// There is no supported cwd API here (research.md §4 —
// NtQueryInformationProcess territory is a separate card), so the probe
// reports UNANSWERABLE and the sweep preserves every tree.

package worktree

import "errors"

// platformProcessCWDs is the Windows stub: an unanswerable probe. The sweep
// renders the error as cwd=undetermined and PRESERVES every tree — never a
// negative (REQ-WS-005).
var platformProcessCWDs = func() ([]string, error) {
	return nil, errors.New("process cwd probe is unsupported on this platform; treat every worktree as occupied")
}
