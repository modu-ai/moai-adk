//go:build windows

// Windows no-op advisory lock stub for observation append.
// Concurrent SubagentStop multi-process race protection is not supported on
// Windows in W3. Single-process write semantics are sufficient for the
// in-session capture pipeline; multi-process concurrency on Windows is a
// follow-up SPEC (LockFileEx via syscall.windows).

package capture

import "os"

func acquireExclusiveLock(_ *os.File) {}

func releaseLock(_ *os.File) {}
