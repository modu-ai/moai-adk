//go:build !windows

package harness

// DRAFT mutant of the heal-lock prototype (t1432 amendment 0.4.1): the helper requests a SHARED lock
// (LOCK_SH) instead of an exclusive one. Two healers that both request a shared lock do not exclude
// each other, which is the property REQ-HRH-005 exists for.

import "syscall"

func init() { protoLockMode = syscall.LOCK_SH }
