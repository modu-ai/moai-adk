package cli

import (
	"errors"
	"time"
)

// codexAuditWaitLoadThreshold is the one-minute load average at or above which
// a codex audit's wait bound doubles (card t1781). At load 200 the machine is
// running several compile/test suites at once — the observed stall that
// motivated the card happened at load 357/364/352 — so a bound tuned on a
// quiet machine no longer fits, and the audit reads as hung when it is only
// waiting in line.
const codexAuditWaitLoadThreshold = 200.0

// errShortLoadAvg reports a load-average payload too short to decode; the
// callers treat it as "no reader" and keep the base bound.
var errShortLoadAvg = errors.New("load-average payload too short")

// codexAuditWaitTimeoutForLoad scales base by the wait multiplier for the
// given one-minute load average. It is the testable core; the live path is
// codexAuditWaitTimeout.
func codexAuditWaitTimeoutForLoad(base time.Duration, load1 float64) time.Duration {
	if load1 >= codexAuditWaitLoadThreshold {
		return 2 * base
	}
	return base
}

// codexAuditWaitTimeout scales base by the machine's current one-minute load
// average. An unreadable load (a platform without a reader, a read error) and
// a negative reading leave base untouched — fail-open, so this can never
// become a new way for an audit to be cut short.
func codexAuditWaitTimeout(base time.Duration) time.Duration {
	load1, err := loadAverage1()
	if err != nil || load1 < 0 {
		return base
	}
	return codexAuditWaitTimeoutForLoad(base, load1)
}
