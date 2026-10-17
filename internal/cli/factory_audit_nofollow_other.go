//go:build !unix

package cli

// auditLogOpenNoFollow is zero off Unix. Windows has no O_NOFOLLOW, and an
// O_NONBLOCK value has no effect there because Windows has no FIFOs for it to
// guard. The Lstat check in openLeaderAnchorAuditLog is therefore the only
// guard there (SPEC-LAUNCHER-ARGV-001 REQ-LA-009; card t1604, O28).
const auditLogOpenNoFollow = 0
