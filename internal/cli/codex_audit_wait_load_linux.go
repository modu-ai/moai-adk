//go:build linux

package cli

import (
	"os"
	"strconv"
	"strings"
)

// loadAverage1 returns the one-minute load average from /proc/loadavg's first
// field ("0.52 0.58 0.59 1/444 12345").
func loadAverage1() (float64, error) {
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return 0, errShortLoadAvg
	}
	return strconv.ParseFloat(fields[0], 64)
}
