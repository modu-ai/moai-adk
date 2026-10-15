//go:build darwin

package cli

import (
	"encoding/binary"

	"golang.org/x/sys/unix"
)

// loadAverage1 returns the one-minute load average from the vm.loadavg
// sysctl: a struct loadavg of three fixed-point values plus the scale, the
// scale being 1000 on every Darwin release this repository targets.
func loadAverage1() (float64, error) {
	raw, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		return 0, err
	}
	if len(raw) < 4 {
		return 0, errShortLoadAvg
	}
	return float64(binary.NativeEndian.Uint32(raw[0:4])) / 1000.0, nil
}
