//go:build windows

// Windows no-op advisory lock stub for tier state mutation.
// Single-process semantics are sufficient for W3; multi-process Windows
// LockFileEx integration is a follow-up SPEC.

package tier

import "os"

func acquireExclusiveLock(_ *os.File) {}

func releaseLock(_ *os.File) {}
