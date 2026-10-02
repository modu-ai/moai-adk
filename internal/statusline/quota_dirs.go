package statusline

import (
	"path/filepath"
	"time"
)

// SPEC-QUOTA-RECORD-WORKTREES-001 M1 (REQ-QWR-001, -002): the multi-directory
// quota reading. These are signature-only stubs: the shapes the tests compile
// against, with the behaviour of the single-directory reading. The bodies —
// the file-read worktree enumeration and the cross-directory reading — land at
// M3.

// QuotaStateDirs returns the state directories the quota reading covers for the
// repository rooted at root, the primary checkout's own first, bounded to
// maxDirs linked-worktree entries.
//
// Stub: the primary state directory only.
func QuotaStateDirs(root string, maxDirs int) []string {
	return []string{filepath.Join(root, ".moai", "state")}
}

// AggregateQuotaDirs reads the session telemetry records under every state
// directory in stateDirs and reports each window as AggregateQuota does for one
// directory; a tie on capture time keeps the directory listed first.
//
// Stub: reads the first directory only, through the single-directory reading.
func AggregateQuotaDirs(stateDirs []string, now time.Time, maxAge time.Duration) QuotaAggregate {
	if len(stateDirs) == 0 {
		return QuotaAggregate{
			FiveHour: QuotaReading{State: QuotaUnknown},
			SevenDay: QuotaReading{State: QuotaUnknown},
		}
	}
	return AggregateQuota(stateDirs[0], now, maxAge)
}
