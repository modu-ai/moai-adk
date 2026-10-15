package cli

import (
	"errors"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// errShortLoadAvg reports a load-average payload too short to decode; the
// callers treat it as "no reader" and keep the base bound.
var errShortLoadAvg = errors.New("load-average payload too short")

// codexAuditWaitTimeoutForLoad scales base by the wait multiplier for the
// given one-minute load average. It is the testable core; the live path is
// codexAuditWaitTimeout.
func codexAuditWaitTimeoutForLoad(base time.Duration, load1 float64) time.Duration {
	if load1 >= config.DefaultCodexAuditWaitLoadThreshold {
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
