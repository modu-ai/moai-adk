//go:build windows

package cli

import "os/exec"

// verifyRunPrepare is a no-op on Windows: a timeout terminates only the direct
// child, so grandchildren can outlive it (SPEC-VERIFY-RUN-REUSE-001 §D-5; a job
// object is out of scope).
func verifyRunPrepare(_ *exec.Cmd) {}
