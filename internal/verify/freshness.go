package verify

import "time"

// DefaultTTL is the default wall-clock freshness bound for snapshot reuse
// (settled decision: 10 minutes, configurable by the caller).
const DefaultTTL = 10 * time.Minute

// Fresh is the binary two-leg freshness predicate:
//
//	Fresh ⇔ storedKey == currentKey AND now - recordedAt <= ttl
//
// Both legs are required; either failing means stale and the caller MUST
// re-execute the check instead of reusing — a stale snapshot is never citable
// evidence. There is no partial key match and no grace past the TTL boundary.
// A ttl <= 0 selects DefaultTTL.
func Fresh(storedKey, currentKey string, recordedAt, now time.Time, ttl time.Duration) bool {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if storedKey == "" || storedKey != currentKey {
		return false
	}
	return now.Sub(recordedAt) <= ttl
}
