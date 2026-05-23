package migrations_test

import (
	"testing"
)

// TestM001_RewritesHardcodedLiteral verifies that m001 rewrites the hardcoded literal.
// REQ-V3R2-RT-007-022: migration 1 substitutes /Users/goos/go/bin/moai with $HOME/go/bin/moai.
func TestM001_RewritesHardcodedLiteral(t *testing.T) {
	// RED: the m001 package does not yet exist.
	t.Skip("waiting for m001 implementation")
}

// TestM001_NoOpWhenAlreadyClean verifies the no-op path for an already clean project.
// REQ-V3R2-RT-007-023: with no hardcoded literal present, the project is treated as already migrated.
func TestM001_NoOpWhenAlreadyClean(t *testing.T) {
	// RED: the m001 package does not yet exist.
	t.Skip("waiting for m001 implementation")
}

// TestM001_PreservesExecutableBit verifies that execute permissions are preserved.
// REQ-V3R2-RT-007-022: file rewrites preserve the executable bit (0o755).
func TestM001_PreservesExecutableBit(t *testing.T) {
	// RED: the m001 package does not yet exist.
	t.Skip("waiting for m001 implementation")
}

// TestM001_PreservesOtherContent verifies that other content is preserved.
// REQ-V3R2-RT-007-022: only the hardcoded literal is substituted; all other content is preserved.
func TestM001_PreservesOtherContent(t *testing.T) {
	// RED: the m001 package does not yet exist.
	t.Skip("waiting for m001 implementation")
}

// TestM001_RollbackNotImplemented verifies that m001 does not support rollback.
// REQ-V3R2-RT-007-024: m001 declares Rollback: nil and does not support rollback.
func TestM001_RollbackNotImplemented(t *testing.T) {
	// RED: the m001 package does not yet exist.
	t.Skip("waiting for m001 implementation")
}

// TestM001_WindowsGitBash verifies behavior in the Windows Git Bash environment.
// REQ-V3R2-RT-007-060: $HOME is shell-expanded, so the migration is platform-independent.
func TestM001_WindowsGitBash(t *testing.T) {
	// RED: the m001 package does not yet exist.
	t.Skip("waiting for m001 implementation")
}
