//go:build !darwin && !linux

package cli

import "errors"

// loadAverage1 has no reader on this platform; codexAuditWaitTimeout keeps
// the base bound (fail-open).
func loadAverage1() (float64, error) {
	return 0, errors.New("load average unavailable on this platform")
}
