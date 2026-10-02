//go:build unix

package cli

// audit_plan_cmd_fifo_unix_test.go — the named-pipe fixture of the audit-plan
// checker's file-error cases (SPEC-AUDIT-MODEL-CONVERGE-001 M4b, AC-ACV-015 (j5)).
// Mkfifo exists on unix only; the sibling file keeps the windows build compiling.
//
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001

import "syscall"

// mkfifoForTest creates a named pipe at path.
func mkfifoForTest(path string) error {
	return syscall.Mkfifo(path, 0o600)
}
